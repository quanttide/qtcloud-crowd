package store

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OSSConfig 是阿里云 OSS 连接配置（环境变量注入）。
type OSSConfig struct {
	Endpoint        string // 如 https://oss-cn-hangzhou.aliyuncs.com
	Bucket          string // 后台桶（私有：审核/认证/结算完整数据）
	PublicBucket    string // 公开桶（公共读 + CDN：黄页快照/任务池）
	AccessKeyID     string
	AccessKeySecret string
}

// OSS 是阿里云 OSS 实现：key 即对象名（Path 风格签名）。
type OSS struct {
	cfg  OSSConfig
	http *http.Client
}

// NewOSS 创建 OSS 存储。
func NewOSS(cfg OSSConfig) *OSS {
	return &OSS{cfg: cfg, http: &http.Client{Timeout: 10 * time.Second}}
}

// Get 读取对象；对象不存在（404）时返回 ErrNotFound。
func (s *OSS) Get(ctx context.Context, key string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.objectURL(s.cfg.Bucket, key), nil)
	if err != nil {
		return nil, err
	}
	s.sign(req, s.cfg.Bucket, key, "")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oss get %s: %s", key, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// Put 写入后台桶对象（覆盖语义）。
func (s *OSS) Put(ctx context.Context, key string, data []byte) error {
	return s.putObject(ctx, s.cfg.Bucket, key, data)
}

// PutPublic 写公开桶对象（黄页快照，覆盖语义）。
func (s *OSS) PutPublic(ctx context.Context, key string, data []byte) error {
	if s.cfg.PublicBucket == "" {
		return fmt.Errorf("oss put public %s: public bucket not configured", key)
	}
	return s.putObject(ctx, s.cfg.PublicBucket, key, data)
}

// DeletePublic 删除公开桶对象；404 视为已删除（幂等——认领/关闭时对象可能已不存在）。
func (s *OSS) DeletePublic(ctx context.Context, key string) error {
	if s.cfg.PublicBucket == "" {
		return fmt.Errorf("oss delete public %s: public bucket not configured", key)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.objectURL(s.cfg.PublicBucket, key), nil)
	if err != nil {
		return err
	}
	s.sign(req, s.cfg.PublicBucket, key, "")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oss delete public %s: %s", key, resp.Status)
	}
	return nil
}

func (s *OSS) putObject(ctx context.Context, bucket, key string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.objectURL(bucket, key), bytes.NewReader(data))
	if err != nil {
		return err
	}
	s.sign(req, bucket, key, "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oss put %s: %s", key, resp.Status)
	}
	return nil
}

func (s *OSS) objectURL(bucket, key string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cfg.Endpoint, "/"), bucket, key)
}

// sign 按阿里云 OSS 规范签名（Path 风格，仅 Date/Content-Type 参与）。
func (s *OSS) sign(req *http.Request, bucket, key, contentType string) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	stringToSign := fmt.Sprintf(
		"%s\n\n%s\n%s\n/%s/%s",
		req.Method, contentType, date, bucket, key,
	)
	h := hmac.New(sha1.New, []byte(s.cfg.AccessKeySecret))
	h.Write([]byte(stringToSign))
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil))
	req.Header.Set("Authorization", "OSS "+s.cfg.AccessKeyID+":"+sig)
}
