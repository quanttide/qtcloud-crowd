package store

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// mockOSSServer 模拟阿里云 OSS：校验签名头并保存对象。
func mockOSSServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	objects := &sync.Map{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "OSS AKID:") {
			t.Errorf("missing/invalid Authorization header: %q", auth)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/bucket/")
		switch r.Method {
		case http.MethodGet:
			v, ok := objects.Load(key)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			_, _ = w.Write(v.([]byte))
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			objects.Store(key, data)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, objects
}

func TestOSSPutGetRoundTrip(t *testing.T) {
	srv, objects := mockOSSServer(t)
	st := NewOSS(OSSConfig{
		Endpoint:        srv.URL,
		Bucket:          "bucket",
		AccessKeyID:     "AKID",
		AccessKeySecret: "SECRET",
	})
	ctx := context.Background()

	payload := []byte(`{"tasks":[]}`)
	if err := st.Put(ctx, "data/crowd.json", payload); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, ok := objects.Load("data/crowd.json"); !ok {
		t.Fatal("对象未写入 mock OSS")
	}

	got, err := st.Get(ctx, "data/crowd.json")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("got %q, want %q", got, payload)
	}
}

func TestOSSGetNotFound(t *testing.T) {
	srv, _ := mockOSSServer(t)
	st := NewOSS(OSSConfig{
		Endpoint:        srv.URL,
		Bucket:          "bucket",
		AccessKeyID:     "AKID",
		AccessKeySecret: "SECRET",
	})
	_, err := st.Get(context.Background(), "missing.json")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestOSSPutWithContentType(t *testing.T) {
	// PUT 带 Content-Type: application/json，请求带签名可被 mock server 接受。
	srv, _ := mockOSSServer(t)
	st := NewOSS(OSSConfig{
		Endpoint:        srv.URL,
		Bucket:          "bucket",
		AccessKeyID:     "AKID",
		AccessKeySecret: "SECRET",
	})
	if err := st.Put(context.Background(), "k.json", []byte(`{}`)); err != nil {
		t.Fatalf("Put: %v", err)
	}
}
