// Package task 实现"事：任务审核"——任务模型、仓储与 HTTP API。
package task

import "strings"

// Status 是任务状态（pending → reviewing → done）。
type Status string

const (
	StatusPending   Status = "pending"
	StatusReviewing Status = "reviewing"
	StatusDone      Status = "done"
)

// Task 是任务模型。
type Task struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	AcceptanceCriteria string `json:"acceptance_criteria"`
	Status             Status `json:"status"`
}

// CanPublish 是验收准则兜底（模型层约束）：任务说不清验收 = 不能发布。
func (t Task) CanPublish() bool {
	return strings.TrimSpace(t.AcceptanceCriteria) != ""
}
