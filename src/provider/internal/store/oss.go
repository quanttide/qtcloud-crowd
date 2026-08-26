package store

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig 是阿里云 OSS 连接配置（环境变量注入）。
type OSSConfig struct {
	Endpoint        string // 如 oss-cn-hangzhou.aliyuncs.com（不带 https://）
	Bucket          string // 后台私有桶（审核/认证/结算完整数据）
	AccessKeyID     string
	AccessKeySecret string
}

// OSS 是阿里云 OSS 实现，使用官方 SDK。
type OSS struct {
	client *oss.Client
	bucket *oss.Bucket
}

// NewOSS 创建 OSS 存储。
func NewOSS(cfg OSSConfig) (*OSS, error) {
	// 标准化 endpoint（去掉 https:// 前缀）
	endpoint := cfg.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := oss.New(endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, err
	}

	return &OSS{
		client: client,
		bucket: bucket,
	}, nil
}

// Get 读取对象；对象不存在（404）时返回 ErrNotFound。
func (s *OSS) Get(ctx context.Context, key string) ([]byte, error) {
	body, err := s.bucket.GetObject(key)
	if err != nil {
		if ossErr, ok := err.(oss.ServiceError); ok && ossErr.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

// Put 写入后台桶对象（覆盖语义）。
func (s *OSS) Put(ctx context.Context, key string, data []byte) error {
	return s.bucket.PutObject(key, bytes.NewReader(data))
}
