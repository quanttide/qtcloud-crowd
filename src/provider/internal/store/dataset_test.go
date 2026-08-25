package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestDatasetLoadMissingReturnsEmpty(t *testing.T) {
	st := NewLocal()
	ds, err := Load(context.Background(), st, filepath.Join(t.TempDir(), "none.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ds.Tasks) != 0 || len(ds.Partners) != 0 || len(ds.Settlements) != 0 {
		t.Fatal("缺失数据应返回空数据集")
	}
}

func TestDatasetSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "crowd.json")
	st := NewLocal()
	ctx := context.Background()

	ds := &Dataset{
		Tasks:       []json.RawMessage{json.RawMessage(`{"id":"t1"}`)},
		Partners:    []json.RawMessage{},
		Settlements: []json.RawMessage{json.RawMessage(`{"id":"s1"}`)},
	}
	if err := ds.Save(ctx, st, key); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(ctx, st, key)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Tasks) != 1 || len(loaded.Partners) != 0 || len(loaded.Settlements) != 1 {
		t.Fatalf("往返不一致: %+v", loaded)
	}
}

func TestDatasetLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "crowd.json")
	st := NewLocal()
	if err := st.Put(context.Background(), key, []byte(`{not json`)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, err := Load(context.Background(), st, key); err == nil {
		t.Fatal("损坏 JSON 应报错")
	}
}
