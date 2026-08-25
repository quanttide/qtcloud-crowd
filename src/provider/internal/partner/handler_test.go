package partner

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

func TestPartnerCertifyFlow(t *testing.T) {
	h := newTestHandler(t)

	// 新增待认证执行方
	body := `{"id":"p1","name":"张三","type":"agent","certified":false}`
	if rec := doJSON(t, h, http.MethodPut, "/api/partners", body); rec.Code != http.StatusOK {
		t.Fatalf("PUT want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 认证动作：certified 置 true
	body2 := `{"id":"p1","name":"张三","type":"agent","certified":true}`
	if rec := doJSON(t, h, http.MethodPut, "/api/partners", body2); rec.Code != http.StatusOK {
		t.Fatalf("PUT 认证 want 200, got %d", rec.Code)
	}

	got := decodeList[Partner](t, doJSON(t, h, http.MethodGet, "/api/partners", ""))
	if len(got) != 1 || !got[0].Certified {
		t.Fatalf("want 1 条已认证执行方，got %+v", got)
	}
}

func TestPartnerInvalidTypeRejected(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPut, "/api/partners", `{"id":"p1","name":"张三","type":"mystery","certified":false}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("非法 type want 400, got %d", rec.Code)
	}
}

func TestPartnerMissingNameRejected(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodPut, "/api/partners", `{"id":"p1","type":"agent","certified":false}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("缺 name want 400, got %d", rec.Code)
	}
}

func TestPartnerListEmpty(t *testing.T) {
	h := newTestHandler(t)
	rec := doJSON(t, h, http.MethodGet, "/api/partners", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if got := decodeList[Partner](t, rec); len(got) != 0 {
		t.Fatalf("want empty, got %+v", got)
	}
}
