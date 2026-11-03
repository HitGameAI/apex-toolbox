/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-02 09:57:31
 * @FilePath: \go-toolbox\pkg\logger\sanitize_benchmark_test.go
 * @Description: 脱敏判定与值扫描的性能基准
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
)

// sanitizeBenchLogger 构造指定脱敏强度的基准 Logger（上限保持默认开启）
func sanitizeBenchLogger(mode SanitizeMode) *Logger {
	l := benchLogger(FormatJSON, false)
	l.sanitize = mode
	l.maxLineBytes = DefaultMaxLineBytes
	return l
}

// ==================== 字段名判定 ====================
//
// 判定路径的成本全在"未命中"的字段名上——生产日志里绝大多数字段名都不敏感，
// 每条日志的每个字段都要走一遍。这几个基准守住两级索引的收益不被后续改动吃掉

// BenchmarkSanitize_KeyMiss 普通业务字段名：形状位图直接判负
func BenchmarkSanitize_KeyMiss(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = desensitize.ActionForKey("region")
	}
}

// BenchmarkSanitize_KeyExactHit 规范 snake_case 的敏感字段名：一次查表命中
func BenchmarkSanitize_KeyExactHit(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = desensitize.ActionForKey("access_token")
	}
}

// BenchmarkSanitize_KeyCamelHit 非规范形态：需归一化后重查
func BenchmarkSanitize_KeyCamelHit(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = desensitize.ActionForKey("refreshToken")
	}
}

// BenchmarkSanitize_KeySuffixHit 后缀规则命中：反向字典树回溯
func BenchmarkSanitize_KeySuffixHit(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = desensitize.ActionForKey("wx_app_secret")
	}
}

// BenchmarkSanitize_KeyOpsVisible 运维字段：命中敏感后缀后再查豁免表，判定链最长
func BenchmarkSanitize_KeyOpsVisible(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = desensitize.ActionForKey("idempotency_key")
	}
}

// ==================== 整条日志 ====================

// BenchmarkSanitize_Off 关闭脱敏的基线
func BenchmarkSanitize_Off(b *testing.B) {
	l := sanitizeBenchLogger(SanitizeOff)
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("req done", "uid", 10086, "app", "gameai", "ns", "business",
			"conn", 9527, "region", "cn")
	}
}

// BenchmarkSanitize_Keys 按字段名脱敏（默认强度）：与上一项之差即净代价
func BenchmarkSanitize_Keys(b *testing.B) {
	l := sanitizeBenchLogger(SanitizeKeys)
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("req done", "uid", 10086, "app", "gameai", "ns", "business",
			"conn", 9527, "region", "cn")
	}
}

// BenchmarkSanitize_OpsFields 全是运维字段：字段名判定的最坏情况
func BenchmarkSanitize_OpsFields(b *testing.B) {
	l := sanitizeBenchLogger(SanitizeKeys)
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("ops", "ip_address", "10.0.0.1", "cache_key", "user:1001",
			"error_code", "E4001", "idempotency_key", "idem-1", "status_code", 200)
	}
}

// BenchmarkSanitize_ChildSteady 子 Logger 的固定字段：脱敏在预编码阶段完成，
// 稳态应与关闭脱敏持平
func BenchmarkSanitize_ChildSteady(b *testing.B) {
	child := sanitizeBenchLogger(SanitizeKeys).WithField("request_id", "rq_0123456789abcdef")
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		child.InfoMsg("req done")
	}
}

// ==================== 值扫描 ====================

// BenchmarkSanitize_PrefilterMiss 预过滤判负：值扫描模式下绝大多数文本应走这条路
func BenchmarkSanitize_PrefilterMiss(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = mayContainSensitiveText("req done path=/api/v1/users status ok")
	}
}

// BenchmarkSanitize_PrefilterOpsIdentifiers 运维标识不得误命中预过滤
// （判据若放宽到"含若干数字"，trace_id 与时间戳会让预过滤形同虚设）
func BenchmarkSanitize_PrefilterOpsIdentifiers(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_ = mayContainSensitiveText("trace_id tr_abc0123def4567890 ts 1756141206 elapsed 12ms")
	}
}

// BenchmarkSanitize_ScanHit 命中扫描规则的完整路径
func BenchmarkSanitize_ScanHit(b *testing.B) {
	l := sanitizeBenchLogger(SanitizeKeysAndValues)
	b.ReportAllocs()
	for range b.N {
		l.InfoKV("cb", "note", "callback https://api.x.com/v1/cb?signature=abc&id=7 to alice@example.com")
	}
}
