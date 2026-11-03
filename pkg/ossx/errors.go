/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\errors.go
 * @Description: go-toolbox/pkg/ossx 核心错误定义
 *
 * 定义所有适配器共用的错误变量，适配器可使用 fmt.Errorf("%w: ...") 包装
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package ossx

import "errors"

var (
	ErrNilClient           = errors.New("ossx: 客户端为空")     // 客户端实例为空
	ErrNilAdapter          = errors.New("ossx: 适配器为空")     // 适配器实例为空
	ErrInvalidBucketName   = errors.New("ossx: 无效的存储桶名称")  // 存储桶名称为空
	ErrInvalidObjectKey    = errors.New("ossx: 无效的对象键")    // 对象键为空
	ErrInvalidExpiry       = errors.New("ossx: 无效的过期时间")   // 预签名过期时间必须大于零
	ErrPresignFailed       = errors.New("ossx: 预签名操作失败")   // 生成预签名URL失败
	ErrUploadFailed        = errors.New("ossx: 上传操作失败")    // 上传对象失败
	ErrDownloadFailed      = errors.New("ossx: 下载操作失败")    // 下载对象失败
	ErrDeleteFailed        = errors.New("ossx: 删除操作失败")    // 删除对象失败
	ErrBucketNotFound      = errors.New("ossx: 存储桶不存在")    // 指定存储桶不存在
	ErrObjectNotFound      = errors.New("ossx: 对象不存在")     // 指定对象不存在
	ErrUnsupportedProvider = errors.New("ossx: 不支持的存储提供商") // 当前适配器不支持此操作
)
