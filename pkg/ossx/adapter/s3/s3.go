/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-ossx\adapter\s3\s3.go
 * @Description: AWS S3 对象存储适配器
 *
 * 基于 aws-sdk-go-v2/service/s3 实现 ossx.Adapter 接口
 * 支持：预签名上传/下载、对象 CRUD、桶存在判断、桶名前缀
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/ossx"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Option S3 适配器可选配置
type Option func(*Adapter)

// WithPrefix 为桶名添加前缀（多租户场景）
func WithPrefix(prefix string) Option {
	return func(a *Adapter) {
		a.prefix = prefix
	}
}

// Adapter S3 存储适配器
type Adapter struct {
	client *s3.Client
	prefix string
}

// NewAdapter 创建 S3 适配器
// client: 由调用方创建并管理生命周期的 s3.Client 实例
// opts: 可选配置（如 WithPrefix）
func NewAdapter(client *s3.Client, opts ...Option) (*Adapter, error) {
	if client == nil {
		return nil, ossx.ErrNilClient
	}

	a := &Adapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

// resolveBucket 解析带前缀的桶名
func (a *Adapter) resolveBucket(bucketName string) string {
	if a.prefix != "" {
		return a.prefix + bucketName
	}
	return bucketName
}

// Provider 返回存储提供商类型
func (a *Adapter) Provider() ossx.ProviderType {
	return ossx.ProviderS3
}

// PresignedDownloadURL 生成预签名下载地址
func (a *Adapter) PresignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	if err := ossx.ValidatePresignParams(bucketName, objectKey, expiry); err != nil {
		return nil, err
	}

	bucketName = a.resolveBucket(bucketName)
	presignClient := s3.NewPresignClient(a.client)

	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrPresignFailed, err)
	}

	signedHeader := make(map[string]string)
	for k, v := range result.SignedHeader {
		if len(v) > 0 {
			signedHeader[k] = v[0]
		}
	}

	return &ossx.PresignedResult{
		URL:          result.URL,
		Method:       result.Method,
		SignedHeader: signedHeader,
	}, nil
}

// PresignedUploadURL 生成预签名上传地址
func (a *Adapter) PresignedUploadURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*ossx.PresignedResult, error) {
	if err := ossx.ValidatePresignParams(bucketName, objectKey, expiry); err != nil {
		return nil, err
	}

	bucketName = a.resolveBucket(bucketName)
	presignClient := s3.NewPresignClient(a.client)

	result, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrPresignFailed, err)
	}

	signedHeader := make(map[string]string)
	for k, v := range result.SignedHeader {
		if len(v) > 0 {
			signedHeader[k] = v[0]
		}
	}

	return &ossx.PresignedResult{
		URL:          result.URL,
		Method:       result.Method,
		SignedHeader: signedHeader,
	}, nil
}

// PutObject 上传对象到指定存储桶
func (a *Adapter) PutObject(ctx context.Context, bucketName, objectKey string, data io.Reader, size int64, contentType string) (*ossx.ObjectInfo, error) {
	bucketName = a.resolveBucket(bucketName)
	if contentType == "" {
		contentType = ossx.DefaultContentType()
	}

	result, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          data,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrUploadFailed, err)
	}

	return &ossx.ObjectInfo{
		Key:         objectKey,
		Size:        size,
		ETag:        aws.ToString(result.ETag),
		ContentType: contentType,
	}, nil
}

// GetObject 获取对象元信息
func (a *Adapter) GetObject(ctx context.Context, bucketName, objectKey string) (*ossx.ObjectInfo, error) {
	bucketName = a.resolveBucket(bucketName)

	result, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrObjectNotFound, err)
	}

	metadata := make(map[string]string)
	for k, v := range result.Metadata {
		metadata[k] = v
	}

	return &ossx.ObjectInfo{
		Key:          objectKey,
		Size:         aws.ToInt64(result.ContentLength),
		ETag:         aws.ToString(result.ETag),
		ContentType:  aws.ToString(result.ContentType),
		LastModified: aws.ToTime(result.LastModified),
		Metadata:     metadata,
	}, nil
}

// GetObjectBlob 获取对象二进制内容
func (a *Adapter) GetObjectBlob(ctx context.Context, bucketName, objectKey string) ([]byte, error) {
	bucketName = a.resolveBucket(bucketName)

	result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrDownloadFailed, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ossx.ErrDownloadFailed, err)
	}

	return data, nil
}

// DeleteObject 删除指定对象
func (a *Adapter) DeleteObject(ctx context.Context, bucketName, objectKey string) error {
	bucketName = a.resolveBucket(bucketName)

	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ossx.ErrDeleteFailed, err)
	}
	return nil
}

// ListObjects 列举存储桶中的对象
func (a *Adapter) ListObjects(ctx context.Context, bucketName, prefix string, limit int) ([]ossx.ObjectInfo, string, error) {
	bucketName = a.resolveBucket(bucketName)
	maxKeys := int32(limit)

	result, err := a.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: &maxKeys,
	})
	if err != nil {
		return nil, "", err
	}

	objects := make([]ossx.ObjectInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		objects = append(objects, ossx.ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			ETag:         aws.ToString(obj.ETag),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}

	nextToken := ""
	if result.NextContinuationToken != nil {
		nextToken = aws.ToString(result.NextContinuationToken)
	}

	return objects, nextToken, nil
}

// BucketExists 判断存储桶是否存在
func (a *Adapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	bucketName = a.resolveBucket(bucketName)

	_, err := a.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// IsConnected 检查底层 S3 客户端是否可用
func (a *Adapter) IsConnected() bool {
	if a.client == nil {
		return false
	}
	_, err := a.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String("health-check"),
	})
	return err == nil
}

// PutObjectFromBytes 从字节切片上传对象（便捷方法）
func (a *Adapter) PutObjectFromBytes(ctx context.Context, bucketName, objectKey string, data []byte, contentType string) (*ossx.ObjectInfo, error) {
	return a.PutObject(ctx, bucketName, objectKey, bytes.NewReader(data), int64(len(data)), contentType)
}
