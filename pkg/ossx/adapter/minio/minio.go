/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\adapter\minio\minio.go
 * @Description: MinIO 对象存储适配器
 *
 * 基于 minio-go/v7 实现 ossx.Adapter 接口
 * 支持：预签名上传/下载、对象 CRUD、桶存在判断
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package minio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/ossx"
	"github.com/minio/minio-go/v7"
)

// Adapter MinIO 存储适配器
type Adapter struct {
	client *minio.Client
}

// NewAdapter 创建 MinIO 适配器
// client: 由调用方创建并管理生命周期的 minio.Client 实例
func NewAdapter(client *minio.Client) (*Adapter, error) {
	if client == nil {
		return nil, ossx.ErrNilClient
	}
	return &Adapter{client: client}, nil
}

// Provider 返回存储提供商类型
func (a *Adapter) Provider() ossx.ProviderType {
	return ossx.ProviderMinIO
}

// PresignedDownloadURL 生成预签名下载地址
func (a *Adapter) PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	if err := ossx.ValidatePresignParams(bucketName, objectKey, expiry); err != nil {
		return nil, err
	}

	u, err := a.client.PresignedGetObject(ctx, bucketName, objectKey, expiry, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrPresignFailed, err)
	}

	return &ossx.PresignedResult{
		URL:          u.String(),
		Method:       http.MethodGet,
		SignedHeader: ossx.HeaderToMap(u.Query()),
	}, nil
}

// PresignedUploadURL 生成预签名上传地址
func (a *Adapter) PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	if err := ossx.ValidatePresignParams(bucketName, objectKey, expiry); err != nil {
		return nil, err
	}

	u, err := a.client.PresignedPutObject(ctx, bucketName, objectKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrPresignFailed, err)
	}

	return &ossx.PresignedResult{
		URL:          u.String(),
		Method:       http.MethodPut,
		SignedHeader: ossx.HeaderToMap(u.Query()),
	}, nil
}

// PutObject 上传对象到指定存储桶
func (a *Adapter) PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ossx.ObjectInfo, error) {
	if contentType == "" {
		contentType = ossx.DefaultContentType()
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := a.client.PutObject(ctx, bucketName, objectKey, data, size, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}

	return &ossx.ObjectInfo{
		Key:         info.Key,
		Size:        info.Size,
		ETag:        info.ETag,
		ContentType: contentType,
	}, nil
}

// GetObject 获取对象元信息
func (a *Adapter) GetObject(ctx context.Context, bucketName, objectKey string) (*ossx.ObjectInfo, error) {
	stat, err := a.client.StatObject(ctx, bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrObjectNotFound, err)
	}

	return &ossx.ObjectInfo{
		Key:          stat.Key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		Metadata:     stat.UserMetadata,
	}, nil
}

// GetObjectBlob 获取对象二进制内容
func (a *Adapter) GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	obj, err := a.client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrDownloadFailed, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrDownloadFailed, err)
	}

	return data, nil
}

// DeleteObject 删除指定对象
func (a *Adapter) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	err := a.client.RemoveObject(ctx, bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("%w: %v", ossx.ErrDeleteFailed, err)
	}
	return nil
}

// ListObjects 列举存储桶中的对象
func (a *Adapter) ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ossx.ObjectInfo, string, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		MaxKeys:   limit,
	}

	var objects []ossx.ObjectInfo
	var nextMarker string

	for object := range a.client.ListObjects(ctx, bucketName, opts) {
		if object.Err != nil {
			return nil, "", object.Err
		}

		objects = append(objects, ossx.ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			ETag:         object.ETag,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
		})

		if len(objects) >= limit {
			nextMarker = object.Key
			break
		}
	}

	return objects, nextMarker, nil
}

// BucketExists 判断存储桶是否存在
func (a *Adapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return a.client.BucketExists(ctx, bucketName)
}

// IsConnected 检查底层 MinIO 客户端是否可用
func (a *Adapter) IsConnected() bool {
	if a.client == nil {
		return false
	}
	_, err := a.client.HealthCheck(3 * time.Second)
	return err == nil
}
