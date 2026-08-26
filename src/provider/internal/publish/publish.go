// Package publish 实现 OSS 共享数据层的"发布"：
// 审核通过的任务写公开桶（黄页快照——当前可接任务），认领/关闭时撤回。
//
// 依赖方向：后台（qtcloud-crowd provider）只依赖 OSS——发布是"推"的动作（带审计）；
// 前台（qtcrowd site/studio）直接读公开桶（CDN 静态分发），不经过本包。
package publish

import (
	"context"
	"encoding/json"

	"github.com/quanttide/qtcloud-crowd-provider/internal/store"
)

// Prefix 是公开桶内黄页快照对象的路径前缀约定：public/tasks/{id}.json。
const Prefix = "public/tasks/"

// Snapshot 是公开桶中的黄页快照：黄页模型视图（title/reward/报名引导）。
// 内部数据（验收准则等）留后台桶——模型不同构由桶边界天然解决。
type Snapshot struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Reward      string `json:"reward"`
	ApplyGuide  string `json:"apply_guide"`
	Status      string `json:"status"`
}

// Key 返回任务 id 对应的公开对象 key（public/tasks/{id}.json）。
func Key(id string) string {
	return Prefix + id + ".json"
}

// Publisher 负责向公开数据层发布/撤回任务黄页快照。
type Publisher struct {
	st store.Store
}

// NewPublisher 创建发布器。
func NewPublisher(st store.Store) *Publisher {
	return &Publisher{st: st}
}

// Publish 写黄页快照对象；status 恒为 published——快照即发布态。
func (p *Publisher) Publish(ctx context.Context, snap Snapshot) error {
	snap.Status = "published"
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return p.st.PutPublic(ctx, Key(snap.ID), data)
}

// Remove 删除公开对象（任务被认领/关闭时）；对象不存在时幂等返回。
func (p *Publisher) Remove(ctx context.Context, id string) error {
	return p.st.DeletePublic(ctx, Key(id))
}
