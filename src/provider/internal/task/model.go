// Package task 实现"事：任务审核"——任务模型、仓储与 HTTP API。
package task

import "strings"

// Status 是任务状态（pending → reviewing → published → done）。
//
// 审核通过 = published（后台写公开桶黄页快照）；认领写回 published → accepted，
// 交付写回 accepted → reviewing（待验收）。
type Status string

const (
	StatusPending   Status = "pending"
	StatusReviewing Status = "reviewing"
	StatusPublished Status = "published" // 审核通过，已发布到公开桶（当前可接任务）
	StatusAccepted  Status = "accepted"  // 已被执行方认领（公开对象已撤回）
	StatusDone      Status = "done"
)

// Task 是任务模型。
// Reward/ApplyGuide 是黄页快照字段（title/description/reward/报名引导），
// 审核通过时随公开对象发布；PartnerID 记录认领方（写回 API 写入）。
type Task struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	AcceptanceCriteria string `json:"acceptance_criteria"`
	Reward             string `json:"reward"`
	ApplyGuide         string `json:"apply_guide"`
	PartnerID          string `json:"partner_id,omitempty"` // 认领方
	Status             Status `json:"status"`
}

// CanPublish 是验收准则兜底（模型层约束）：任务说不清验收 = 不能发布。
func (t Task) CanPublish() bool {
	return strings.TrimSpace(t.AcceptanceCriteria) != ""
}
