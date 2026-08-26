package store

import (
	"context"
	"os"
	"path/filepath"
)

// Local 是本地文件系统实现：key 即文件路径（如 "data/crowd.json"）。
type Local struct{}

// NewLocal 创建本地文件存储。
func NewLocal() *Local {
	return &Local{}
}

// Get 读取文件；不存在时返回 ErrNotFound。
func (s *Local) Get(_ context.Context, key string) ([]byte, error) {
	data, err := os.ReadFile(key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

// Put 原子写入文件：先写临时文件再 rename，避免写一半损坏数据。
func (s *Local) Put(_ context.Context, key string, data []byte) error {
	return putFile(key, data)
}

// putFile 原子写文件：先写临时文件再 rename。
func putFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
