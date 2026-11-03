/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-30 08:51:03
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-30 09:50:16
 * @FilePath: \go-toolbox\pkg\safe\bench_test.go
 * @Description: 常驻基准 - 安全访问、字段名归一化、时间窗哈希与数值兜底
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package safe

import (
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
)

var (
	benchTemporalHashResult string
)

func BenchmarkTemporalHasher_Hash(b *testing.B) {
	hasher := NewTemporalHasher(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchTemporalHashResult = hasher.Hash("user123", "device456", "iOS")
	}
}

func BenchmarkTemporalHasher_HashMap(b *testing.B) {
	hasher := NewTemporalHasher(nil)
	kvMap := map[string]string{
		"userId":   "user123",
		"deviceId": "device456",
		"platform": "iOS",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchTemporalHashResult = hasher.HashMap(kvMap)
	}
}

func BenchmarkTemporalHasher_HashAt(b *testing.B) {
	hasher := NewTemporalHasher(nil)
	t := time.Now()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchTemporalHashResult = hasher.HashAt(t, "user123", "device456", "iOS")
	}
}

func BenchmarkTemporalHasher_DifferentPartsCount(b *testing.B) {
	hasher := NewTemporalHasher(nil)

	tests := []struct {
		name  string
		parts []string
	}{
		{"2个参数", []string{"user123", "device456"}},
		{"3个参数", []string{"user123", "device456", "iOS"}},
		{"5个参数", []string{"user123", "device456", "iOS", "Mozilla/5.0", "1.0.0"}},
		{"10个参数", []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchTemporalHashResult = hasher.Hash(tt.parts...)
			}
		})
	}
}

func BenchmarkTemporalHasher_DifferentHashLength(b *testing.B) {
	tests := []struct {
		name   string
		length int
	}{
		{"8字符", 8},
		{"12字符", 12},
		{"16字符", 16},
		{"32字符", 32},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			hasher := NewTemporalHasher(
				WithWindow(5*time.Minute),
				WithLength(tt.length),
			)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchTemporalHashResult = hasher.Hash("user123", "device456", "iOS")
			}
		})
	}
}

func BenchmarkTemporalHasher_DifferentMapSize(b *testing.B) {
	tests := []struct {
		name  string
		kvMap map[string]string
	}{
		{
			"2个键值对",
			map[string]string{
				"userId":   "user123",
				"deviceId": "device456",
			},
		},
		{
			"5个键值对",
			map[string]string{
				"userId":     "user123",
				"deviceId":   "device456",
				"platform":   "iOS",
				"User-Agent": "Mozilla/5.0",
				"version":    "1.0.0",
			},
		},
		{
			"10个键值对",
			map[string]string{
				"userId":     "user123",
				"deviceId":   "device456",
				"platform":   "iOS",
				"User-Agent": "Mozilla/5.0",
				"version":    "1.0.0",
				"language":   "zh-CN",
				"timezone":   "Asia/Shanghai",
				"network":    "WiFi",
				"carrier":    "China Mobile",
				"model":      "iPhone 13",
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			hasher := NewTemporalHasher(nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchTemporalHashResult = hasher.HashMap(tt.kvMap)
			}
		})
	}
}

func BenchmarkTemporalHasher_IsExpired(b *testing.B) {
	hasher := NewTemporalHasher(nil)
	hash := hasher.Hash("user123", "device456", "iOS")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = hasher.IsExpired(hash, "user123", "device456", "iOS")
	}
}

func BenchmarkTemporalHasher_Parallel(b *testing.B) {
	hasher := NewTemporalHasher(nil)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchTemporalHashResult = hasher.Hash("user123", "device456", "iOS")
		}
	})
}

func BenchmarkTemporalHasher_ParallelWithMap(b *testing.B) {
	hasher := NewTemporalHasher(nil)
	kvMap := map[string]string{
		"userId":   "user123",
		"deviceId": "device456",
		"platform": "iOS",
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchTemporalHashResult = hasher.HashMap(kvMap)
		}
	})
}

// ——— Field 热路径：原名命中、变体命中、未命中与 nil 链路 ———

type benchProfile struct{ Avatar string }
type benchUser struct {
	Name    string
	Profile *benchProfile
}

var benchFieldMap = map[string]interface{}{"user_name": "kamalyes", "age": 33}
var benchFieldStruct = &benchUser{Name: "kamalyes", Profile: &benchProfile{Avatar: "a.png"}}

func BenchmarkField_MapExactHit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(benchFieldMap).Field("user_name").String()
	}
}
func BenchmarkField_MapVariantHit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(benchFieldMap).Field("userName").String()
	}
}
func BenchmarkField_MapMiss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(benchFieldMap).Field("nope").String()
	}
}
func BenchmarkField_StructExactHit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(benchFieldStruct).Field("Name").String()
	}
}
func BenchmarkField_StructChain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(benchFieldStruct).Field("Profile").Field("Avatar").String()
	}
}
func BenchmarkField_NilChain(b *testing.B) {
	var u *benchUser
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New(u).Field("Profile").Field("Avatar").String()
	}
}
func BenchmarkField_NormalizeFieldName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = stringx.NormalizeFieldName("UserName")
	}
}
