package task

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/quanttide/qtcloud-crowd-provider/internal/publish"
	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// newTestHandler 用本地文件存储 + 临时文件构建任务 handler（公开层落在 {dir}/data/public/）。
func newTestHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	key := filepath.Join(dir, "data", "crowd.json")
	st := store.NewLocalPublicDir(filepath.Join(dir, "data", "public"))
	return NewHandler(NewRepository(st, key), publish.NewPublisher(st)), key
}

func doJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, r)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeList[T any](t *testing.T, rec *httptest.ResponseRecorder) []T {
	t.Helper()
	var out []T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func TestTaskListEmpty(t *testing.T) {
	h, _ := newTestHandler(t)
	rec := doJSON(t, h, http.MethodGet, "/api/tasks", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if got := decodeList[Task](t, rec); len(got) != 0 {
		t.Fatalf("want empty, got %+v", got)
	}
}

func TestTaskUpsertAndOverwrite(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"id":"t1","title":"撰写财报摘要","content":"读财报","acceptance_criteria":"覆盖收入利润现金流","status":"pending"}`
	if rec := doJSON(t, h, http.MethodPut, "/api/tasks", body); rec.Code != http.StatusOK {
		t.Fatalf("PUT 新增 want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 同 id 覆盖：status 改为 reviewing
	body2 := `{"id":"t1","title":"撰写财报摘要","content":"读财报","acceptance_criteria":"覆盖收入利润现金流","status":"reviewing"}`
	if rec := doJSON(t, h, http.MethodPut, "/api/tasks", body2); rec.Code != http.StatusOK {
		t.Fatalf("PUT 覆盖 want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	got := decodeList[Task](t, doJSON(t, h, http.MethodGet, "/api/tasks", ""))
	if len(got) != 1 {
		t.Fatalf("同 id 覆盖后应只有 1 条，got %d", len(got))
	}
	if got[0].Status != StatusReviewing {
		t.Fatalf("status want reviewing, got %s", got[0].Status)
	}
}

func TestTaskPublishWithoutCriteriaRejected(t *testing.T) {
	// 失败路径：验收准则兜底——说不清验收不能发布（进入 reviewing）。
	h, _ := newTestHandler(t)
	body := `{"id":"t1","title":"说不清的任务","content":"做事","acceptance_criteria":"","status":"reviewing"}`
	rec := doJSON(t, h, http.MethodPut, "/api/tasks", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestTaskMissingIDRejected(t *testing.T) {
	h, _ := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPut, "/api/tasks", `{"title":"no id"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestTaskUpsertPreservesOtherResources(t *testing.T) {
	// 数据隔离：任务写回不能弄丢执行方/结算（共享数据集文档）。
	h, key := newTestHandler(t)
	seed := `{"tasks":[{"id":"t1","title":"A","content":"","acceptance_criteria":"","status":"pending"}],"partners":[{"id":"p1","name":"张三","type":"agent","certified":true}],"settlements":[]}`
	if err := store.NewLocal().Put(t.Context(), key, []byte(seed)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := doJSON(t, h, http.MethodPut, "/api/tasks", `{"id":"t2","title":"B","content":"","acceptance_criteria":"","status":"pending"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	data, err := store.NewLocal().Get(t.Context(), key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	var ds store.Dataset
	if err := json.Unmarshal(data, &ds); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(ds.Tasks) != 2 {
		t.Fatalf("tasks want 2, got %d", len(ds.Tasks))
	}
	if len(ds.Partners) != 1 {
		t.Fatalf("partners want 1, got %d", len(ds.Partners))
	}
}

// TestTaskStateMachineFlow 状态机流转：pending→reviewing→published→accepted（认领）→reviewing（交付）→done。
func TestTaskStateMachineFlow(t *testing.T) {
	h, _ := newTestHandler(t)
	taskBody := func(status string) string {
		return `{"id":"t1","title":"渠道推广任务","content":"按量潮标准推广","acceptance_criteria":"推广数据完整","reward":"按量潮标准结算","status":"` + status + `"}`
	}

	steps := []struct {
		method string
		path   string
		body   string
		want   Status
	}{
		{http.MethodPut, "/api/tasks", taskBody("pending"), StatusPending},
		{http.MethodPut, "/api/tasks", taskBody("reviewing"), StatusReviewing},
		{http.MethodPut, "/api/tasks", taskBody("published"), StatusPublished},
		{http.MethodPost, "/api/tasks/t1/claim", `{"partner_id":"p1"}`, StatusAccepted},
		{http.MethodPost, "/api/tasks/t1/deliver", ``, StatusReviewing},
		{http.MethodPut, "/api/tasks", taskBody("done"), StatusDone},
	}
	for _, s := range steps {
		rec := doJSON(t, h, s.method, s.path, s.body)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s want 200, got %d: %s", s.method, s.path, rec.Code, rec.Body.String())
		}
		var got Task
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Status != s.want {
			t.Fatalf("%s %s want status %s, got %s", s.method, s.path, s.want, got.Status)
		}
	}

	list := decodeList[Task](t, doJSON(t, h, http.MethodGet, "/api/tasks", ""))
	if len(list) != 1 || list[0].Status != StatusDone {
		t.Fatalf("最终状态 want done, got %+v", list)
	}
}

// TestTaskPublishWritesPublicLayer 发布写公开层（local 模式）：
// 审核通过（status=published）生成 data/public/tasks/{id}.json；认领后删除。
func TestTaskPublishWritesPublicLayer(t *testing.T) {
	h, key := newTestHandler(t)
	// key = {dir}/data/crowd.json → 公开对象 = {dir}/data/public/tasks/t3.json
	publicObj := filepath.Join(filepath.Dir(key), "public", "tasks", "t3.json")

	body := `{"id":"t3","title":"渠道推广任务","content":"按量潮标准推广","acceptance_criteria":"推广数据完整","reward":"按量潮标准结算","apply_guide":"联系量潮运营报名","status":"published"}`
	if rec := doJSON(t, h, http.MethodPut, "/api/tasks", body); rec.Code != http.StatusOK {
		t.Fatalf("发布 want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 生成：public/ 目录下出现 tasks/t3.json，内容是黄页快照（不含验收准则等内部数据）。
	data, err := os.ReadFile(publicObj)
	if err != nil {
		t.Fatalf("公开对象未生成: %v", err)
	}
	var snap struct {
		Title      string `json:"title"`
		Reward     string `json:"reward"`
		ApplyGuide string `json:"apply_guide"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("解析公开对象: %v", err)
	}
	if snap.Title != "渠道推广任务" || snap.Reward != "按量潮标准结算" || snap.ApplyGuide != "联系量潮运营报名" || snap.Status != "published" {
		t.Fatalf("黄页快照内容不符: %+v", snap)
	}

	// 认领 → 删除公开对象（任务不再可接）。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/t3/claim", `{"partner_id":"p1"}`); rec.Code != http.StatusOK {
		t.Fatalf("认领 want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(publicObj); !os.IsNotExist(err) {
		t.Fatalf("认领后公开对象应删除, got %v", err)
	}
}

// TestTaskClaimWriteBack 写回 API——认领：成功（published→accepted）/ 未知任务 404 / 状态非法 400。
func TestTaskClaimWriteBack(t *testing.T) {
	h, _ := newTestHandler(t)
	// 种子：t1 已发布（published），t2 待审核（reviewing）。
	putOK(t, h, `{"id":"t1","title":"A","content":"","acceptance_criteria":"达标","status":"published"}`)
	putOK(t, h, `{"id":"t2","title":"B","content":"","acceptance_criteria":"达标","status":"reviewing"}`)

	// 成功：published→accepted，partner_id 落库。
	rec := doJSON(t, h, http.MethodPost, "/api/tasks/t1/claim", `{"partner_id":"p1"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("claim 成功 want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusAccepted || got.PartnerID != "p1" {
		t.Fatalf("认领结果不符: %+v", got)
	}

	// 未知任务 → 404。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/nope/claim", `{"partner_id":"p1"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("未知任务 want 404, got %d", rec.Code)
	}
	// 状态非法（reviewing 不可认领）→ 400。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/t2/claim", `{"partner_id":"p1"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("非 published 认领 want 400, got %d", rec.Code)
	}
	// 缺 partner_id → 400。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/t1/claim", `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("缺 partner_id want 400, got %d", rec.Code)
	}
}

// TestTaskDeliverWriteBack 写回 API——交付：成功（accepted→reviewing）/ 未知任务 404 / 状态非法 400。
func TestTaskDeliverWriteBack(t *testing.T) {
	h, _ := newTestHandler(t)
	putOK(t, h, `{"id":"t1","title":"A","content":"","acceptance_criteria":"达标","status":"published"}`)
	putOK(t, h, `{"id":"t2","title":"B","content":"","acceptance_criteria":"达标","status":"reviewing"}`)
	// 认领 t1 → accepted。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/t1/claim", `{"partner_id":"p1"}`); rec.Code != http.StatusOK {
		t.Fatalf("预置认领 want 200, got %d", rec.Code)
	}

	// 成功：accepted→reviewing。
	rec := doJSON(t, h, http.MethodPost, "/api/tasks/t1/deliver", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("deliver 成功 want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusReviewing {
		t.Fatalf("交付后 want reviewing, got %s", got.Status)
	}

	// 未知任务 → 404。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/nope/deliver", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("未知任务 want 404, got %d", rec.Code)
	}
	// 状态非法（published 不可交付）→ 400。
	if rec := doJSON(t, h, http.MethodPost, "/api/tasks/t2/deliver", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("非 accepted 交付 want 400, got %d", rec.Code)
	}
}

func putOK(t *testing.T, h http.Handler, body string) {
	t.Helper()
	if rec := doJSON(t, h, http.MethodPut, "/api/tasks", body); rec.Code != http.StatusOK {
		t.Fatalf("PUT want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
