package task

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// newTestHandler 用本地文件存储 + 临时文件构建任务 handler。
func newTestHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	key := filepath.Join(t.TempDir(), "crowd.json")
	st := store.NewLocal()
	return NewHandler(NewRepository(st, key)), key
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
