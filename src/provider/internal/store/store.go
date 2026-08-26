// Package store 提供对象存储抽象：以 key（逻辑路径/对象名）读写 JSON 数据。
//
// 本地实现（local）把 key 视为文件路径；OSS 实现（oss）把 key 视为对象名。
// 量潮众包管理云只用对象存储、不引数据库。
package store

import (
	"context"
	"errors"
)

// ErrNotFound 表示 key 对应的数据不存在。
var ErrNotFound = errors.New("store: key not found")

// Store 是对象存储抽象（后台桶 + 公开桶两层）。
//
// OSS 共享数据层把存储一分为二：后台桶（私有，完整数据）与公开桶（公共读 + CDN，
// 黄页快照——当前可接任务）。后台桶用 Get/Put，公开桶用 PutPublic/DeletePublic。
// 公开对象 key 遵循路径前缀约定 public/tasks/{id}.json。
type Store interface {
	// Get 读取后台桶 key 对应的数据；不存在时返回 ErrNotFound。
	Get(ctx context.Context, key string) ([]byte, error)
	// Put 写入后台桶 key 对应的数据（原子写，不存在则创建）。
	Put(ctx context.Context, key string, data []byte) error
	// PutPublic 写入公开桶对象（黄页快照，覆盖语义）。
	PutPublic(ctx context.Context, key string, data []byte) error
	// DeletePublic 删除公开桶对象；对象不存在视为已删除（幂等）。
	DeletePublic(ctx context.Context, key string) error
}
