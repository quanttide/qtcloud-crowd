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

// Store 是对象存储抽象。
type Store interface {
	// Get 读取 key 对应的数据；不存在时返回 ErrNotFound。
	Get(ctx context.Context, key string) ([]byte, error)
	// Put 写入 key 对应的数据（原子写，不存在则创建）。
	Put(ctx context.Context, key string, data []byte) error
}
