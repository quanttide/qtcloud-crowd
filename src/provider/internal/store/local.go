package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// defaultPublicDir 是本地公开数据层根目录：对应 OSS 公开桶。
// 私有数据默认在 data/crowd.json，公开层落在 data/public/ 下。
const defaultPublicDir = "data/public"

// Local 是本地文件系统实现：key 即文件路径（如 "data/crowd.json"）。
// 公开桶操作把 public/tasks/{id}.json 映射为 {公开根目录}/tasks/{id}.json。
type Local struct {
	publicDir string // 公开数据层根目录；空 = data/public
}

// NewLocal 创建本地文件存储。
func NewLocal() *Local {
	return &Local{}
}

// NewLocalPublicDir 创建本地文件存储并指定公开数据层根目录（默认 data/public，
// 测试可注入临时目录避免污染仓库）。
func NewLocalPublicDir(dir string) *Local {
	return &Local{publicDir: dir}
}

func (s *Local) publicRoot() string {
	if s.publicDir != "" {
		return s.publicDir
	}
	return defaultPublicDir
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

// PutPublic 写公开对象：public/tasks/{id}.json 映射为 {公开根目录}/tasks/{id}.json。
func (s *Local) PutPublic(ctx context.Context, key string, data []byte) error {
	return putFile(s.publicPath(key), data)
}

// DeletePublic 删除公开对象；不存在时幂等返回（认领/关闭时对象可能已不存在）。
func (s *Local) DeletePublic(_ context.Context, key string) error {
	if err := os.Remove(s.publicPath(key)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// publicPath 把公开 key（public/...）映射为公开根目录下的相对路径：
// "public/tasks/t1.json" → "{publicRoot}/tasks/t1.json"。
func (s *Local) publicPath(key string) string {
	return filepath.Join(s.publicRoot(), filepath.FromSlash(strings.TrimPrefix(key, "public/")))
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
