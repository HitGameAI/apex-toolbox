/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-18 15:38:52
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 15:38:52
 * @FilePath: \go-toolbox\pkg\i18n\bench_test.go
 * @Description: 热路径基准：占位符替换、嵌套扁平化、语言协商与端到端翻译
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"testing"
)

// 预先构造的固定输入，避免把输入构造成本计入基准
// [EN] Fixtures built once so the benchmark measures the function, not setup.
var (
	benchTemplateData = map[string]any{"name": "kamalyes", "count": 42, "ok": true}

	benchMsgNoPlaceholder = "Resource not found"
	benchMsgOne           = "Hello {name}"
	benchMsgThree         = "Hello {name}, you have {count} items, active={ok}"
	benchMsgGoTemplate    = "Hello {{.name}}, you have {{count}} items"

	benchNested = map[string]any{
		"error": map[string]any{
			"internal":   "内部错误",
			"notfound":   "资源未找到",
			"validation": map[string]any{"failed": "校验失败", "required": "必填"},
		},
		"ok":   "成功",
		"user": map[string]any{"created": "已创建", "deleted": "已删除"},
	}

	benchAcceptLang = "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"

	benchSink string
	benchMap  map[string]string
)

func BenchmarkFormat_NoPlaceholder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = FormatWithTemplateData(benchMsgNoPlaceholder, benchTemplateData)
	}
}

func BenchmarkFormat_OnePlaceholder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = FormatWithTemplateData(benchMsgOne, benchTemplateData)
	}
}

func BenchmarkFormat_ThreePlaceholders(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = FormatWithTemplateData(benchMsgThree, benchTemplateData)
	}
}

func BenchmarkFormat_GoTemplateStyle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = FormatWithTemplateData(benchMsgGoTemplate, benchTemplateData)
	}
}

func BenchmarkFlattenToMessages(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchMap = FlattenToMessages(benchNested)
	}
}

func BenchmarkParseAcceptLanguage(b *testing.B) {
	cfg := DefaultConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = cfg.ParseAcceptLanguage(benchAcceptLang)
	}
}

func BenchmarkNormalizeLanguage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSink = NormalizeLanguage("zh-hans-cn")
	}
}

// BenchmarkGetMessage 端到端：命中语言 + 命中键，无模板
func BenchmarkGetMessage(b *testing.B) {
	mgr := benchManager(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = mgr.GetMessage("zh", "hello")
	}
}

// BenchmarkGetMessageWithMap 端到端：命中 + 三个命名占位符
func BenchmarkGetMessageWithMap(b *testing.B) {
	mgr := benchManager(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = mgr.GetMessageWithMap("zh", "greeting", benchTemplateData)
	}
}

// BenchmarkGetMessage_Fallback 端到端：指定语言缺键，回退默认语言
func BenchmarkGetMessage_Fallback(b *testing.B) {
	mgr := benchManager(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = mgr.GetMessage("zh", "only_en")
	}
}

// BenchmarkGetMessage_Parallel 并发读：验证 RWMutex 在只读路径上的表现
func BenchmarkGetMessage_Parallel(b *testing.B) {
	mgr := benchManager(b)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mgr.GetMessage("zh", "hello")
		}
	})
}

func benchManager(tb testing.TB) *Manager {
	tb.Helper()
	mgr, err := NewManager(&Config{
		Enabled:            true,
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "zh"},
		EnableFallback:     true,
		MessageLoader: NewMapLoader(map[string]map[string]string{
			"en": {"hello": "Hello", "greeting": "Hello {name}, you have {count} items, active={ok}", "only_en": "English only"},
			"zh": {"hello": "你好", "greeting": "你好 {name}，共 {count} 项，启用={ok}"},
		}),
	})
	if err != nil {
		tb.Fatalf("构造 Manager 失败: %v", err)
	}
	return mgr
}
