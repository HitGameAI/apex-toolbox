/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-23 00:00:00
 * @FilePath: \go-toolbox\pkg\ossx\logger.go
 * @Description: go-toolbox/pkg/ossx 默认日志器，直接复用 go-toolbox/pkg/logger
 *
 * 当调用方未传入日志器时，使用此默认日志器
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package ossx

import (
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
)

// NewDefaultLogger 创建默认日志器
// 日志级别 INFO，前缀 [OSSX]，不显示调用者，启用彩色输出
func NewDefaultLogger() logger.ILogger {
	return logger.NewLogger().
		WithLevel(logger.INFO).
		WithPrefix("[OSSX] ").
		WithShowCaller(false).
		WithColorful(true).
		WithTimeFormat(time.RFC3339)
}
