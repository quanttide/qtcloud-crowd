package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalGetPutRoundTrip(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "data", "crowd.json")
	st := NewLocal()

	ctx := context.Background()
	if err := st.Put(ctx, key, []byte(`{"hello":"world"}`)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := st.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != `{"hello":"world"}` {
		t.Fatalf("got %q, want %q", got, `{"hello":"world"}`)
	}
}

func TestLocalGetNotFound(t *testing.T) {
	st := NewLocal()
	_, err := st.Get(context.Background(), filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestLocalPutAtomicNoTmpLeftover(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "crowd.json")
	st := NewLocal()
	if err := st.Put(context.Background(), key, []byte("[]")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, err := os.Stat(key + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("tmp 文件应已清理，但存在：%v", err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Fatalf("正式文件应存在：%v", err)
	}
}

func TestLocalPutOverwrite(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "crowd.json")
	st := NewLocal()
	ctx := context.Background()
	if err := st.Put(ctx, key, []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := st.Put(ctx, key, []byte("v2")); err != nil {
		t.Fatal(err)
	}
	got, _ := st.Get(ctx, key)
	if string(got) != "v2" {
		t.Fatalf("got %q, want v2", got)
	}
}

func TestLocalPutPublicWritesPublicDir(t *testing.T) {
	// 发布写公开层：key（public/tasks/{id}.json）映射到公开根目录下的 tasks/{id}.json。
	publicDir := t.TempDir()
	st := NewLocalPublicDir(publicDir)
	ctx := context.Background()

	if err := st.PutPublic(ctx, "public/tasks/t1.json", []byte(`{"status":"published"}`)); err != nil {
		t.Fatalf("PutPublic: %v", err)
	}
	got, err := st.Get(ctx, filepath.Join(publicDir, "tasks", "t1.json"))
	if err != nil {
		t.Fatalf("Get 公开对象: %v", err)
	}
	if string(got) != `{"status":"published"}` {
		t.Fatalf("got %q", got)
	}
}

func TestLocalDeletePublicRemovesAndIdempotent(t *testing.T) {
	publicDir := t.TempDir()
	st := NewLocalPublicDir(publicDir)
	ctx := context.Background()
	key := "public/tasks/t1.json"
	objPath := filepath.Join(publicDir, "tasks", "t1.json")

	if err := st.PutPublic(ctx, key, []byte(`{}`)); err != nil {
		t.Fatalf("PutPublic: %v", err)
	}
	if err := st.DeletePublic(ctx, key); err != nil {
		t.Fatalf("DeletePublic: %v", err)
	}
	if _, err := st.Get(ctx, objPath); !errors.Is(err, ErrNotFound) {
		t.Fatalf("删除后应不存在，got %v", err)
	}
	// 幂等：对象不存在时重复删除不报错（认领/关闭时对象可能已不存在）。
	if err := st.DeletePublic(ctx, key); err != nil {
		t.Fatalf("重复删除应幂等，got %v", err)
	}
}

func TestLocalDefaultPublicDir(t *testing.T) {
	// 约定：local 模式公开层默认落在 data/public/ 目录。
	if got := NewLocal().publicRoot(); got != "data/public" {
		t.Fatalf("默认公开目录 want data/public, got %s", got)
	}
	if got := NewLocalPublicDir("/tmp/x").publicRoot(); got != "/tmp/x" {
		t.Fatalf("注入公开目录 want /tmp/x, got %s", got)
	}
}
