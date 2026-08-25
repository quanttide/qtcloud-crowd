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
	Bucket          string
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.objectURL(key), nil)
	if err != nil {
		return nil, err
	}
	s.sign(req, key, "")
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

// Put 写入对象（覆盖语义）。
func (s *OSS) Put(ctx context.Context, key string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.objectURL(key), bytes.NewReader(data))
	if err != nil {
		return err
	}
	s.sign(req, key, "application/json")
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

func (s *OSS) objectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cfg.Endpoint, "/"), s.cfg.Bucket, key)
}

// sign 按阿里云 OSS 规范签名（Path 风格，仅 Date/Content-Type 参与）。
func (s *OSS) sign(req *http.Request, key, contentType string) {
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	stringToSign := fmt.Sprintf(
		"%s\n\n%s\n%s\n/%s/%s",
		req.Method, contentType, date, s.cfg.Bucket, key,
	)
	h := hmac.New(sha1.New, []byte(s.cfg.AccessKeySecret))
	h.Write([]byte(stringToSign))
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil))
	req.Header.Set("Authorization", "OSS "+s.cfg.AccessKeyID+":"+sig)
}
