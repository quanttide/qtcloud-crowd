// Package partner 实现"人：执行方管理"——执行方模型、仓储与 HTTP API。
package partner

// Type 是执行方类型。
type Type string

const (
	TypeChannel  Type = "channel"  // 渠道
	TypeAgent    Type = "agent"    // 代理
	TypeTraining Type = "training" // 实训基地成员
)

// Partner 是执行方模型。
type Partner struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      Type   `json:"type"`
	Certified bool   `json:"certified"` // 认证状态：接单前必须先知道对方是谁
}
