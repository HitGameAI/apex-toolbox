/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\client.go
 * @Description: go-toolbox/pkg/ossx 对象存储易用性封装客户端
 *
 * 纯易用性封装，不管理连接生命周期
 * 连接由调用方（如 go-rpc-gateway）负责创建和关闭
 * 通过 Adapter 适配器模式支持 MinIO / AWS S3 / 阿里云 OSS / 腾讯云 COS / Cloudflare R2 Worker
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package ossx

import (
	"context"
	"io"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
)

// ProviderType 存储提供商标识
type ProviderType string

const (
	ProviderMinIO      ProviderType = "minio"       // MinIO 对象存储
	ProviderS3         ProviderType = "s3"          // AWS S3 及兼容协议存储
	ProviderAliyunOSS  ProviderType = "aliyun_oss"  // 阿里云 OSS
	ProviderTencentCOS ProviderType = "tencent_cos" // 腾讯云 COS
	ProviderR2         ProviderType = "r2"          // Cloudflare R2 Worker
)

// PresignedResult 预签名URL结果
type PresignedResult struct {
	URL          string            `json:"url"`           // 预签名访问地址
	Method       string            `json:"method"`        // HTTP 请求方法（GET/PUT）
	SignedHeader map[string]string `json:"signed_header"` // 需携带的签名头
}

// ObjectInfo 对象元信息
type ObjectInfo struct {
	Key          string            `json:"key"`           // 对象键
	Size         int64             `json:"size"`          // 对象大小（字节）
	ETag         string            `json:"etag"`          // 对象 ETag
	ContentType  string            `json:"content_type"`  // 内容类型
	LastModified time.Time         `json:"last_modified"` // 最后修改时间
	Metadata     map[string]string `json:"metadata"`      // 自定义元数据
}

// BucketInfo 存储桶信息
type BucketInfo struct {
	Name         string    `json:"name"`          // 存储桶名称
	CreationDate time.Time `json:"creation_date"` // 创建时间
}

// Adapter 对象存储适配器接口
// 各存储提供商实现此接口，通过适配器模式统一操作
type Adapter interface {
	// Provider 返回存储提供商类型
	Provider() ProviderType
	// PresignedDownloadURL 生成预签名下载地址
	PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error)
	// PresignedUploadURL 生成预签名上传地址
	PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error)
	// PutObject 上传对象
	PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ObjectInfo, error)
	// GetObject 获取对象元信息
	GetObject(ctx context.Context, bucketName, objectKey string) (*ObjectInfo, error)
	// GetObjectBlob 获取对象二进制内容
	GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error)
	// DeleteObject 删除对象
	DeleteObject(ctx context.Context, bucketName, objectKey string) error
	// ListObjects 列举对象，返回对象列表和下一页标记
	ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ObjectInfo, string, error)
	// BucketExists 判断存储桶是否存在
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	// IsConnected 检查底层客户端是否可用
	IsConnected() bool
}

// Client go-toolbox/pkg/ossx 客户端，持有 Adapter 实例并委托调用
type Client struct {
	adapter Adapter        // 存储适配器
	logger  logger.ILogger // 日志记录器
}

// New 创建 go-toolbox/pkg/ossx 客户端
// adapter: 存储适配器实例（如 minio.Adapter、s3adapter.Adapter、r2.Adapter）
// log: 可选日志记录器，不传则使用默认日志器
func New(adapter Adapter, log ...logger.ILogger) (*Client, error) {
	if adapter == nil {
		return nil, ErrNilAdapter
	}

	var l logger.ILogger
	if len(log) > 0 && log[0] != nil {
		l = log[0]
	} else {
		l = NewDefaultLogger()
	}

	return &Client{
		adapter: adapter,
		logger:  l,
	}, nil
}

// Adapter 返回底层适配器实例
func (c *Client) Adapter() Adapter {
	return c.adapter
}

// Provider 返回当前存储提供商类型
func (c *Client) Provider() ProviderType {
	return c.adapter.Provider()
}

// IsConnected 检查底层客户端连接状态
func (c *Client) IsConnected() bool {
	return c.adapter.IsConnected()
}

// PresignedDownloadURL 生成预签名下载地址
func (c *Client) PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error) {
	result, err := c.adapter.PresignedDownloadURL(ctx, bucketName, objectKey, expiry)
	if err != nil {
		c.logger.ErrorContextKV(ctx, "预签名下载地址生成失败", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey, "error", err)
		return nil, err
	}
	c.logger.DebugContextKV(ctx, "预签名下载地址已生成", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey)
	return result, nil
}

// PresignedUploadURL 生成预签名上传地址
func (c *Client) PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*PresignedResult, error) {
	result, err := c.adapter.PresignedUploadURL(ctx, bucketName, objectKey, expiry)
	if err != nil {
		c.logger.ErrorContextKV(ctx, "预签名上传地址生成失败", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey, "error", err)
		return nil, err
	}
	c.logger.DebugContextKV(ctx, "预签名上传地址已生成", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey)
	return result, nil
}

// PutObject 上传对象到指定存储桶
func (c *Client) PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ObjectInfo, error) {
	info, err := c.adapter.PutObject(ctx, bucketName, objectKey, data, size, contentType)
	if err != nil {
		c.logger.ErrorContextKV(ctx, "对象上传失败", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey, "error", err)
		return nil, err
	}
	c.logger.DebugContextKV(ctx, "对象已上传", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey, "size", info.Size)
	return info, nil
}

// GetObject 获取对象元信息
func (c *Client) GetObject(ctx context.Context, bucketName, objectKey string) (*ObjectInfo, error) {
	return c.adapter.GetObject(ctx, bucketName, objectKey)
}

// GetObjectBlob 获取对象二进制内容
func (c *Client) GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	return c.adapter.GetObjectBlob(ctx, bucketName, objectKey)
}

// DeleteObject 删除指定对象
func (c *Client) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	err := c.adapter.DeleteObject(ctx, bucketName, objectKey)
	if err != nil {
		c.logger.ErrorContextKV(ctx, "对象删除失败", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey, "error", err)
		return err
	}
	c.logger.DebugContextKV(ctx, "对象已删除", "provider", c.adapter.Provider(), "bucket", bucketName, "key", objectKey)
	return nil
}

// ListObjects 列举存储桶中的对象
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ObjectInfo, string, error) {
	return c.adapter.ListObjects(ctx, bucketName, prefix, limit)
}

// BucketExists 判断存储桶是否存在
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return c.adapter.BucketExists(ctx, bucketName)
}

// Stats 返回客户端状态信息
func (c *Client) Stats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["provider"] = string(c.adapter.Provider())
	stats["connected"] = c.adapter.IsConnected()
	return stats
}
