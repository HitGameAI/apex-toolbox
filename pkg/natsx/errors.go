/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-20 00:00:00
 * @FilePath: \go-toolbox\pkg\natsx\errors.go
 * @Description: go-toolbox/pkg/natsx 错误定义
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package natsx

import "errors"

var (
	ErrNotConnected             = errors.New("nats: not connected")
	ErrAlreadyClosed            = errors.New("nats: already closed")
	ErrInvalidSubject           = errors.New("nats: invalid subject")
	ErrInvalidMessage           = errors.New("nats: invalid message")
	ErrPublishFailed            = errors.New("nats: publish failed")
	ErrSubscribeFailed          = errors.New("nats: subscribe failed")
	ErrJetStreamFailed          = errors.New("nats: jetstream operation failed")
	ErrBucketNotFound           = errors.New("nats: kv bucket not found")
	ErrKeyNotFound              = errors.New("nats: kv key not found")
	ErrTimeout                  = errors.New("nats: operation timed out")
	ErrUnavailable              = errors.New("nats: service unavailable")
	ErrGlobalPoolNotInitialized = errors.New("nats: global consumer pool not initialized, call InitWorkerPool first")

	// ErrPermanent 永久性失败哨兵错误
	// handler 返回 errors.Is 命中此哨兵的错误时，库直接 Term 终止消息（不再重投），
	// 与 Nak 重试路径区分：格式损坏、业务上无法匹配等重试不可修复的场景
	// 用法：return fmt.Errorf("%w: order not found", natsx.ErrPermanent)
	ErrPermanent = errors.New("nats: permanent failure, message will be terminated")
)
