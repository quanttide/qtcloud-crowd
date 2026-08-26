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

// mockOSSServer 模拟阿里云 OSS：校验签名头并保存对象（后台桶 bucket / 公开桶 public-bucket）。
func mockOSSServer(t *testing.T) (*httptest.Server, *sync.Map, *sync.Map) {
	t.Helper()
	private := &sync.Map{}
	public := &sync.Map{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "OSS AKID:") {
			t.Errorf("missing/invalid Authorization header: %q", auth)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var objects *sync.Map
		var key string
		switch {
		case strings.HasPrefix(r.URL.Path, "/bucket/"):
			objects, key = private, strings.TrimPrefix(r.URL.Path, "/bucket/")
		case strings.HasPrefix(r.URL.Path, "/public-bucket/"):
			objects, key = public, strings.TrimPrefix(r.URL.Path, "/public-bucket/")
		default:
			http.Error(w, "unknown bucket", http.StatusNotFound)
			return
		}
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
		case http.MethodDelete:
			objects.Delete(key)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, private, public
}

func TestOSSPutGetRoundTrip(t *testing.T) {
	srv, objects, _ := mockOSSServer(t)
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
	srv, _, _ := mockOSSServer(t)
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
	srv, _, _ := mockOSSServer(t)
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

func TestOSSPutPublicWritesPublicBucket(t *testing.T) {
	srv, private, public := mockOSSServer(t)
	st := NewOSS(OSSConfig{
		Endpoint:        srv.URL,
		Bucket:          "bucket",
		PublicBucket:    "public-bucket",
		AccessKeyID:     "AKID",
		AccessKeySecret: "SECRET",
	})
	ctx := context.Background()
	key := "public/tasks/t1.json"
	payload := []byte(`{"status":"published"}`)

	if err := st.PutPublic(ctx, key, payload); err != nil {
		t.Fatalf("PutPublic: %v", err)
	}
	if v, ok := public.Load(key); !ok {
		t.Fatal("公开对象应写入公开桶")
	} else if string(v.([]byte)) != string(payload) {
		t.Fatalf("got %q, want %q", v, payload)
	}
	if _, ok := private.Load(key); ok {
		t.Fatal("公开对象不应写入后台桶")
	}

	// 删除：公开对象移除，重复删除幂等。
	if err := st.DeletePublic(ctx, key); err != nil {
		t.Fatalf("DeletePublic: %v", err)
	}
	if _, ok := public.Load(key); ok {
		t.Fatal("公开对象应已删除")
	}
	if err := st.DeletePublic(ctx, key); err != nil {
		t.Fatalf("重复删除应幂等: %v", err)
	}
}

func TestOSSPublicBucketNotConfigured(t *testing.T) {
	srv, _, _ := mockOSSServer(t)
	st := NewOSS(OSSConfig{
		Endpoint:        srv.URL,
		Bucket:          "bucket",
		AccessKeyID:     "AKID",
		AccessKeySecret: "SECRET",
	})
	ctx := context.Background()
	if err := st.PutPublic(ctx, "public/tasks/t1.json", []byte(`{}`)); err == nil {
		t.Fatal("公开桶未配置时 PutPublic 应报错")
	}
	if err := st.DeletePublic(ctx, "public/tasks/t1.json"); err == nil {
		t.Fatal("公开桶未配置时 DeletePublic 应报错")
	}
}
