package publish

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

func TestKey(t *testing.T) {
	if got := Key("t1"); got != "public/tasks/t1.json" {
		t.Fatalf("Key want public/tasks/t1.json, got %s", got)
	}
}

func TestPublisherPublishAndRemove(t *testing.T) {
	dir := t.TempDir()
	st := store.NewLocalPublicDir(filepath.Join(dir, "public"))
	pub := NewPublisher(st)
	ctx := context.Background()

	snap := Snapshot{
		ID:          "t1",
		Title:       "渠道推广任务",
		Description: "按量潮标准完成渠道推广",
		Reward:      "按量潮标准结算",
		ApplyGuide:  "联系量潮运营报名",
	}
	if err := pub.Publish(ctx, snap); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// 发布投递前台层：local 模式生成 public/tasks/t1.json（黄页快照字段 + status published）。
	data, err := st.Get(ctx, filepath.Join(dir, "public", "tasks", "t1.json"))
	if err != nil {
		t.Fatalf("公开对象未生成: %v", err)
	}
	var got Snapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("解析快照: %v", err)
	}
	if got.Reward != "按量潮标准结算" || got.Description != "按量潮标准完成渠道推广" || got.Status != "published" {
		t.Fatalf("快照内容不符: %+v", got)
	}

	// 撤回：认领/关闭时删除公开对象。
	if err := pub.Remove(ctx, "t1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := st.Get(ctx, filepath.Join(dir, "public", "tasks", "t1.json")); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("撤回后公开对象应不存在, got %v", err)
	}
}
