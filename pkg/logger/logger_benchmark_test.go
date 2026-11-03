/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-07 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 11:06:28
 * @FilePath: \go-toolbox\pkg\logger\logger_benchmark_test.go
 * @Description: 日志性能基准测试
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"context"
	"io"
	"runtime"
	"testing"

	"google.golang.org/grpc/metadata"
)

// 创建输出到 io.Discard 的 logger，消除 I/O 开销，专注测量 logger 自身开销
func benchLogger(format FormatType, showCaller bool) *Logger {
	l := &Logger{
		colorful:     false,
		format:       format,
		timeFormat:   "2006/01/02 15:04:05",
		output:       io.Discard,
		timestampKey: defaultTimestampKey,
		levelKey:     defaultLevelKey,
		messageKey:   defaultMessageKey,
		callerKey:    defaultCallerKey,
		contextKeys:  append([]compiledContextKey(nil), defaultCompiledContextKeys...),
	}
	l.level.Store(int32(DEBUG))
	l.showCaller.Store(showCaller)
	return l
}

// ==================== JSON 模式（生产默认） ====================

func BenchmarkJSON_Info_NoCaller(b *testing.B) {
	l := benchLogger(FormatJSON, false)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("用户登录成功 user_id=%d ip=%s", 12345, "192.168.1.1")
	}
}

func BenchmarkJSON_Info_WithCaller(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("用户登录成功 user_id=%d ip=%s", 12345, "192.168.1.1")
	}
}

func BenchmarkJSON_InfoMsg_WithCaller(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoMsg("服务启动完成")
	}
}

func BenchmarkJSON_InfoKV_WithCaller(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoKV("数据库查询", "table", "users", "rows", 42, "cost_ms", 3)
	}
}

func BenchmarkJSON_InfoContext_WithCaller(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoContext(ctx, "处理请求 method=%s path=%s", "GET", "/api/v1/users")
	}
}

func BenchmarkJSON_Debug_Disabled(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	l.SetLevel(INFO) // DEBUG 被禁用，测试快速路径
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Debug("这条日志不会输出 %d", 42)
	}
}

// ==================== Text 模式 ====================

func BenchmarkText_Info_WithCaller(b *testing.B) {
	l := benchLogger(FormatText, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("用户登录成功 user_id=%d ip=%s", 12345, "192.168.1.1")
	}
}

func BenchmarkText_InfoKV_WithCaller(b *testing.B) {
	l := benchLogger(FormatText, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoKV("数据库查询", "table", "users", "rows", 42, "cost_ms", 3)
	}
}

// text + ctx 是先前唯一没有常驻基准的 caller 入口组合，
// 栈窗口在这条路径上落空时输出仍然正确，只是退化成完整栈回溯，
// 没有基准就没人看得见这 4 倍开销——留住它，下次回归立刻现形
func BenchmarkText_InfoContextKV_WithCaller(b *testing.B) {
	l := benchLogger(FormatText, true)
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoContextKV(ctx, "数据库查询", "table", "users", "rows", 42, "cost_ms", 3)
	}
}

// ==================== gRPC metadata fallback 场景 ====================
// 线上（apex-notify-service 5w 连接压测）heap profile 铁证：context 提取路径的
// metadata.FromIncomingContext 全量 MD 深拷贝是 go-toolbox/pkg/logger 热路径的主要分配来源。
// 本基准模拟 gRPC 服务端请求：ctx 携带 incoming metadata（真实请求体量），
// contextKeys 中的 key 需 fallback 到 metadata 提取，测量单条日志的完整开销

func benchMetadataCtx() context.Context {
	md := metadata.Pairs(
		"app_id", "10001",
		"namespace", "live",
		"uid", "123456",
		"authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.test-token",
		"user-agent", "grpc-go/1.66.0",
		"content-type", "application/grpc",
	)
	return metadata.NewIncomingContext(context.Background(), md)
}

func BenchmarkJSON_InfoContext_MetadataFallback(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	l.contextKeys = compileContextKeys([]string{"app_id", "namespace", "uid"})
	ctx := benchMetadataCtx()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoContext(ctx, "处理请求 method=%s path=%s", "GET", "/api/v1/users")
	}
}

// ==================== 并发场景 ====================

func BenchmarkJSON_Info_Parallel(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("并发日志 user_id=%d", 12345)
		}
	})
}

// ==================== Caller 开销隔离 ====================

func BenchmarkFindExternalCaller(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.findExternalCaller(2, 3)
	}
}

// ==================== WithFields 场景 ====================

func BenchmarkJSON_WithFields(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.WithField("request_id", "abc-123").WithField("user_id", 42).Info("处理完成")
	}
}

// ==================== caller 定位成本 ====================
//
// runtime.Callers 的成本随请求帧数线性增长，而库内的 PC 缓存查询只占几纳秒。
// 这组基准守住"各入口邻接层浅层定位"的收益：一旦某条路径的帧布局变化导致
// 窗口装不下用户帧，就会退化到 12 帧全量回退，成本翻倍且无任何报错

// benchRuntimeCallers 测量 runtime.Callers 在给定帧数下的裸成本
func benchRuntimeCallers(b *testing.B, frames int) {
	var pcs [16]uintptr
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = runtime.Callers(2, pcs[:frames])
	}
}

func BenchmarkCaller_RuntimeCallers1(b *testing.B)  { benchRuntimeCallers(b, 1) }
func BenchmarkCaller_RuntimeCallers3(b *testing.B)  { benchRuntimeCallers(b, 3) }
func BenchmarkCaller_RuntimeCallers6(b *testing.B)  { benchRuntimeCallers(b, 6) }
func BenchmarkCaller_RuntimeCallers12(b *testing.B) { benchRuntimeCallers(b, 12) }

// BenchmarkCaller_CacheLoad PC 缓存命中的成本（相对栈回溯可忽略）
func BenchmarkCaller_CacheLoad(b *testing.B) {
	var pcs [4]uintptr
	runtime.Callers(2, pcs[:1])
	callerCache.Store(pcs[0], &callerInfo{file: "bench.go"})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = callerCache.Load(pcs[0])
	}
}

// BenchmarkCaller_Shallow 入口邻接层的浅层定位（3 帧窗口）
func BenchmarkCaller_Shallow(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.findExternalCaller(2, 3)
	}
}

// BenchmarkCaller_Deep 头部层的兜底定位（6 帧窗口）
func BenchmarkCaller_Deep(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.findExternalCaller(3, 6)
	}
}

// BenchmarkCaller_MsgEntry 纯消息入口：帧布局 [ultraLog, InfoMsg, 用户]
func BenchmarkCaller_MsgEntry(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoMsg("req done")
	}
}

// BenchmarkCaller_FormatEntry 格式化入口：多一层 ultraLogf
func BenchmarkCaller_FormatEntry(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("req done")
	}
}

// BenchmarkCaller_KVEntry KV 入口：帧布局 [logWithKV, InfoKV, 用户]
func BenchmarkCaller_KVEntry(b *testing.B) {
	l := benchLogger(FormatJSON, true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.InfoKV("req done", "uid", 10086, "region", "cn")
	}
}
