/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\adapter\r2\r2.go
 * @Description: Cloudflare R2 Worker 对象存储适配器
 *
 * 通过 Cloudflare Worker 代理访问 R2 存储
 * 使用 AES-CBC-PKCS7 加密签名保证安全性
 * 仅支持：预签名上传地址、上传对象（其他操作通过 Worker API 扩展）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package r2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/ossx"
	"github.com/HitGameAI/apex-toolbox/pkg/sign"
)

// Config R2 Worker 适配器配置
type Config struct {
	Endpoint string // Worker 代理端点地址
	Key      string // AES 加密密钥（32 字节）
	IV       string // AES 初始化向量（16 字节）
}

// Adapter R2 Worker 存储适配器
type Adapter struct {
	config     *Config
	httpClient *http.Client
}

// NewAdapter 创建 R2 Worker 适配器
func NewAdapter(cfg *Config) (*Adapter, error) {
	if cfg == nil {
		return nil, ossx.ErrNilClient
	}
	if cfg.Endpoint == "" {
		return nil, ossx.ErrInvalidBucketName
	}

	return &Adapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Provider 返回存储提供商类型
func (a *Adapter) Provider() ossx.ProviderType {
	return ossx.ProviderR2
}

// generateSign 生成 AES 加密签名
func (a *Adapter) generateSign(key string) (string, error) {
	signData := map[string]string{
		"key": key,
	}

	jsonData, err := json.Marshal(signData)
	if err != nil {
		return "", fmt.Errorf("JSON marshal failed: %w", err)
	}

	encrypted, err := sign.AesEncryptWithIV(string(jsonData), []byte(a.config.Key), []byte(a.config.IV))
	if err != nil {
		return "", fmt.Errorf("AES encrypt failed: %w", err)
	}

	return encrypted, nil
}

// PresignedDownloadURL 生成预签名下载地址（R2 Worker 不支持）
func (a *Adapter) PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	return nil, fmt.Errorf("%w: R2 Worker 不支持预签名下载", ossx.ErrUnsupportedProvider)
}

// PresignedUploadURL 生成预签名上传地址
func (a *Adapter) PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	if err := ossx.ValidatePresignParams(bucketName, objectKey, 1*time.Second); err != nil {
		return nil, err
	}

	s, err := a.generateSign(objectKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrPresignFailed, err)
	}

	uploadURL := fmt.Sprintf("%s/%s/%s?sign=%s", a.config.Endpoint, bucketName, objectKey, s)

	return &ossx.PresignedResult{
		URL:    uploadURL,
		Method: http.MethodPut,
	}, nil
}

// PutObject 上传对象到 R2
func (a *Adapter) PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ossx.ObjectInfo, error) {
	if contentType == "" {
		contentType = ossx.DefaultContentType()
	}

	s, err := a.generateSign(objectKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}

	uploadURL := fmt.Sprintf("%s/%s/%s?sign=%s", a.config.Endpoint, bucketName, objectKey, s)

	body, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: R2 Worker 返回状态码 %d", ossx.ErrUploadFailed, resp.StatusCode)
	}

	return &ossx.ObjectInfo{
		Key:         objectKey,
		Size:        size,
		ContentType: contentType,
	}, nil
}

// GetObject 获取对象元信息（R2 Worker 有限支持）
func (a *Adapter) GetObject(ctx context.Context, bucketName, objectKey string) (*ossx.ObjectInfo, error) {
	return nil, fmt.Errorf("%w: R2 Worker 不支持获取对象元信息", ossx.ErrUnsupportedProvider)
}

// GetObjectBlob 获取对象二进制内容（R2 Worker 不支持）
func (a *Adapter) GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	return nil, fmt.Errorf("%w: R2 Worker 不支持下载对象", ossx.ErrUnsupportedProvider)
}

// DeleteObject 删除对象（R2 Worker 不支持）
func (a *Adapter) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	return fmt.Errorf("%w: R2 Worker 不支持删除对象", ossx.ErrUnsupportedProvider)
}

// ListObjects 列举对象（R2 Worker 不支持）
func (a *Adapter) ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ossx.ObjectInfo, string, error) {
	return nil, "", fmt.Errorf("%w: R2 Worker 不支持列举对象", ossx.ErrUnsupportedProvider)
}

// BucketExists 判断存储桶是否存在（R2 Worker 不支持）
func (a *Adapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return false, fmt.Errorf("%w: R2 Worker 不支持桶存在判断", ossx.ErrUnsupportedProvider)
}

// IsConnected 检查 R2 Worker 是否可用
func (a *Adapter) IsConnected() bool {
	if a.config == nil || a.config.Endpoint == "" {
		return false
	}
	resp, err := a.httpClient.Get(a.config.Endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
