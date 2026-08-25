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
