package store

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig 是阿里云 OSS 连接配置（环境变量注入）。
type OSSConfig struct {
	Endpoint        string // 如 oss-cn-hangzhou.aliyuncs.com（不带 https://）
	Bucket          string // 后台桶（私有：审核/认证/结算完整数据）
	PublicBucket    string // 前台桶（qtcrowd-site——投递目标，前台自有：公共读 + CDN：黄页快照/任务池）
	AccessKeyID     string
	AccessKeySecret string
}

// OSS 是阿里云 OSS 实现，使用官方 SDK。
type OSS struct {
	client       *oss.Client
	bucket       *oss.Bucket
	publicBucket *oss.Bucket
	cfg          OSSConfig
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

	var publicBucket *oss.Bucket
	if cfg.PublicBucket != "" {
		publicBucket, err = client.Bucket(cfg.PublicBucket)
		if err != nil {
			return nil, err
		}
	}

	return &OSS{
		client:       client,
		bucket:       bucket,
		publicBucket: publicBucket,
		cfg:          cfg,
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

// PutPublic 写前台桶对象（黄页快照，覆盖语义——投递者角色）。
func (s *OSS) PutPublic(ctx context.Context, key string, data []byte) error {
	if s.publicBucket == nil {
		return errors.New("oss put public: public bucket not configured")
	}
	return s.publicBucket.PutObject(key, bytes.NewReader(data))
}

// DeletePublic 删除前台桶对象；404 视为已删除（幂等——认领/关闭时对象可能已不存在）。
func (s *OSS) DeletePublic(ctx context.Context, key string) error {
	if s.publicBucket == nil {
		return errors.New("oss delete public: public bucket not configured")
	}
	err := s.publicBucket.DeleteObject(key)
	if err != nil {
		if ossErr, ok := err.(oss.ServiceError); ok && ossErr.StatusCode == 404 {
			return nil // 幂等
		}
		return err
	}
	return nil
}