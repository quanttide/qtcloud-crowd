package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/quanttide/qtcloud-crowd-provider/internal/partner"
	"github.com/quanttide/qtcloud-crowd-provider/internal/settlement"
	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
	"github.com/quanttide/qtcloud-crowd-provider/internal/task"
)

// TestSmoke 冒烟测试：用仓库内真实种子 data/crowd.json（复制到临时文件），
// 走完整 REST API：读三视角数据 → 各写一条 → 再次读取验证。
func TestSmoke(t *testing.T) {
	seed, err := os.ReadFile(filepath.Join("..", "..", "data", "crowd.json"))
	if err != nil {
		t.Fatalf("读取种子数据: %v", err)
	}

	key := filepath.Join(t.TempDir(), "crowd.json")
	if err := os.WriteFile(key, seed, 0o644); err != nil {
		t.Fatalf("写入种子: %v", err)
	}

	srv := httptest.NewServer(newMux(store.NewLocal(), key))
	t.Cleanup(srv.Close)

	// 1) 种子数据可读：3 任务 / 1 执行方 / 1 结算
	tasks := getList[task.Task](t, srv.URL+"/api/tasks")
	if len(tasks) != 3 {
		t.Fatalf("种子任务 want 3, got %d", len(tasks))
	}
	partners := getList[partner.Partner](t, srv.URL+"/api/partners")
	if len(partners) != 1 {
		t.Fatalf("种子执行方 want 1, got %d", len(partners))
	}
	settlements := getList[settlement.Settlement](t, srv.URL+"/api/settlements")
	if len(settlements) != 1 || settlements[0].Amount != 888.5 {
		t.Fatalf("种子结算 want 1 笔 888.5, got %+v", settlements)
	}
	// 种子含 published 任务（渠道推广任务）：公开层语义的样例数据。
	if tasks[2].Status != task.StatusPublished || tasks[2].Reward != "按量潮标准结算" {
		t.Fatalf("种子 t3 应为 published 渠道推广任务, got %+v", tasks[2])
	}

	// 2) 各写一条：审核发布新任务 / 认证新执行方 / 记一笔结算
	putJSON(t, srv.URL+"/api/tasks", `{"id":"t4","title":"整理会议纪要","content":"输出行动项","acceptance_criteria":"行动项含负责人与截止日","status":"reviewing"}`)
	putJSON(t, srv.URL+"/api/partners", `{"id":"p2","name":"李四","type":"training","certified":true}`)
	postJSON(t, srv.URL+"/api/settlements", `{"id":"s2","task_id":"t1","partner_id":"p2","amount":200,"settled_at":"2026-08-25T12:00:00+08:00"}`)

	// 3) 再次读取验证写入生效
	if got := getList[task.Task](t, srv.URL+"/api/tasks"); len(got) != 4 {
		t.Fatalf("写入后任务 want 4, got %d", len(got))
	}
	if got := getList[partner.Partner](t, srv.URL+"/api/partners"); len(got) != 2 {
		t.Fatalf("写入后执行方 want 2, got %d", len(got))
	}
	if got := getList[settlement.Settlement](t, srv.URL+"/api/settlements"); len(got) != 2 {
		t.Fatalf("写入后结算 want 2, got %d", len(got))
	}

	// 4) 数据已持久化到文件（原子写落盘）
	raw, err := os.ReadFile(key)
	if err != nil {
		t.Fatalf("读取落盘文件: %v", err)
	}
	var ds store.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatalf("落盘 JSON 解析: %v", err)
	}
	if len(ds.Tasks) != 4 || len(ds.Partners) != 2 || len(ds.Settlements) != 2 {
		t.Fatalf("落盘内容不一致: tasks=%d partners=%d settlements=%d", len(ds.Tasks), len(ds.Partners), len(ds.Settlements))
	}
	if _, err := os.Stat(key + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("原子写后不应残留 .tmp 文件")
	}
}

func getList[T any](t *testing.T, url string) []T {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s want 200, got %d: %s", url, resp.StatusCode, body)
	}
	var out []T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	return out
}

func putJSON(t *testing.T, url, body string) {
	t.Helper()
	req(t, http.MethodPut, url, body, http.StatusOK)
}

func postJSON(t *testing.T, url, body string) {
	t.Helper()
	req(t, http.MethodPost, url, body, http.StatusOK)
}

func req(t *testing.T, method, url, body string, wantStatus int) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s want %d, got %d: %s", method, url, wantStatus, resp.StatusCode, body)
	}
}

// TestGetenvFallback 验证默认环境变量取值。
func TestGetenvFallback(t *testing.T) {
	if got := getenv("QTCLOUD_CROWD_ADDR_NOT_SET_12345", ":8080"); got != ":8080" {
		t.Fatalf("want :8080, got %s", got)
	}
}
