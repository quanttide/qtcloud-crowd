package settlement

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

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	key := filepath.Join(t.TempDir(), "crowd.json")
	return NewHandler(NewRepository(store.NewLocal(), key))
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

func TestSettlementCreateAndList(t *testing.T) {
	h := newTestHandler(t)

	body := `{"id":"s1","task_id":"t1","partner_id":"p1","amount":888.5,"settled_at":"2026-08-25T10:00:00+08:00"}`
	if rec := doJSON(t, h, http.MethodPost, "/api/settlements", body); rec.Code != http.StatusOK {
		t.Fatalf("POST 记一笔 want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	got := decodeList[Settlement](t, doJSON(t, h, http.MethodGet, "/api/settlements", ""))
	if len(got) != 1 || got[0].Amount != 888.5 {
		t.Fatalf("want 1 笔 888.5，got %+v", got)
	}
}

func TestSettlementNonPositiveAmountRejected(t *testing.T) {
	// 失败路径：金额必须大于 0。
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPost, "/api/settlements", `{"id":"s1","task_id":"t1","partner_id":"p1","amount":0,"settled_at":"2026-08-25T10:00:00+08:00"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("amount<=0 want 400, got %d", rec.Code)
	}
}

func TestSettlementMissingFieldsRejected(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPost, "/api/settlements", `{"id":"s1","amount":100}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("缺 task_id/partner_id want 400, got %d", rec.Code)
	}
}

func TestSettlementInvalidBodyRejected(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPost, "/api/settlements", `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("坏 JSON want 400, got %d", rec.Code)
	}
}

func TestSettlementMethodNotAllowed(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPut, "/api/settlements", `{}`)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT want 405, got %d", rec.Code)
	}
}
