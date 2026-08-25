// Package settlement 实现"财：结算管理"——验收通过后记一笔。
package settlement

// Settlement 是结算模型：钱从哪来（任务）、付给谁（执行方）、付了多少。
type Settlement struct {
	ID        string  `json:"id"`
	TaskID    string  `json:"task_id"`
	PartnerID string  `json:"partner_id"`
	Amount    float64 `json:"amount"`
	SettledAt string  `json:"settled_at"` // ISO 8601
}
