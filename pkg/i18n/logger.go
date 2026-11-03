/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-18 13:12:33
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 13:12:33
 * @FilePath: \go-toolbox\pkg\i18n\logger.go
 * @Description: 最小日志接口（取代 go-toolbox/pkg/logger 依赖：go-toolbox/pkg/logger 会带入 grpc 与 otel）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

// Logger 最小日志接口
//
// 本包只在"预加载映射目标语言失败"这一处记录日志，因此接口只要一个方法
// 任何实现了 Warn(string, ...any) 的 logger（含 go-toolbox/pkg/logger）都能直接传入，
// 不传则静默——库不应该替使用方决定日志去向
// [EN] The package logs in exactly one place, so one method is enough. Any
// logger with Warn(string, ...any) satisfies it; nil means silence, because a
// library should not pick a log destination for its caller.
type Logger interface {
	Warn(format string, args ...any)
}

// nopLogger 默认的空实现
type nopLogger struct{}

func (nopLogger) Warn(string, ...any) {}
