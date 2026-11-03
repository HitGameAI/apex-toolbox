/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\option.go
 * @Description: 选项模式 - 使用 Functional Options 配置转换器
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import "time"

// Option 转换器配置选项函数
type Option func(*Options)

// Options 转换器配置项
type Options struct {
	AutoTimeConversion bool              // 自动时间转换开关
	FieldMapping       map[string]string // 字段名映射: Model字段名 -> PB字段名
	TagMappingEnabled  bool              // 是否启用 struct tag 自动映射
	TagName            string            // struct tag 名称（默认 pbmo）
	DesensitizeEnabled bool              // 是否启用脱敏（默认关闭，开启后按字段 desensitize tag 生效）
	SafeMode           bool              // 是否启用安全模式
	Concurrency        int               // 并发数（批量转换时使用）
	BatchSize          int               // 批处理大小
	Timeout            time.Duration     // 超时时间
}

// DefaultOptions 默认配置
// FieldMapping 惰性初始化：无映射配置时不分配 map
func DefaultOptions() *Options {
	return &Options{
		AutoTimeConversion: true,
		TagMappingEnabled:  true,
		TagName:            "pbmo",
		DesensitizeEnabled: false,
		SafeMode:           false,
		Concurrency:        0,
		BatchSize:          100,
		Timeout:            30 * time.Second,
	}
}

// ensureFieldMapping 惰性初始化字段映射 map
func ensureFieldMapping(o *Options) map[string]string {
	if o.FieldMapping == nil {
		o.FieldMapping = make(map[string]string)
	}
	return o.FieldMapping
}

// ApplyOptions 应用选项到默认配置
func ApplyOptions(opts ...Option) *Options {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithAutoTimeConversion 设置自动时间转换
func WithAutoTimeConversion(enabled bool) Option {
	return func(o *Options) {
		o.AutoTimeConversion = enabled
	}
}

// WithFieldMapping 设置字段映射
func WithFieldMapping(modelField, pbField string) Option {
	return func(o *Options) {
		ensureFieldMapping(o)[modelField] = pbField
	}
}

// WithFieldMappings 批量设置字段映射
func WithFieldMappings(mappings map[string]string) Option {
	return func(o *Options) {
		fieldMapping := ensureFieldMapping(o)
		for k, v := range mappings {
			fieldMapping[k] = v
		}
	}
}

// WithTagMapping 启用/禁用 struct tag 映射
func WithTagMapping(enabled bool) Option {
	return func(o *Options) {
		o.TagMappingEnabled = enabled
	}
}

// WithTagName 设置 struct tag 名称
func WithTagName(name string) Option {
	return func(o *Options) {
		o.TagName = name
	}
}

// WithDesensitize 启用/禁用脱敏，默认关闭
//
// 开启后也不是所有字段都脱敏：只处理带 desensitize tag 的字符串字段，
// 无 tag 的类型拷贝计划与关闭时逐字节相同，不付任何代价
//
// 作用范围是本转换器自己的顶层字段，嵌套结构体不参与
// 嵌套结构体的拷贝计划按「源类型+目标类型」全进程共享，承载不了 per-converter
// 的开关——这与 AutoTimeConversion 对嵌套同样不生效是一致的
// 嵌套字段需要脱敏时在输出层调 desensitize.Desensitization
//
// [EN] Off by default. Enabling it does not mask everything: only string fields
// carrying a desensitize tag are touched, and a type with no tags gets
// byte-for-byte the same copy plan as with masking off, at no cost. The scope is
// this converter's own top-level fields; nested structs are excluded because
// their copy plans are shared process-wide by source and destination type and
// cannot carry a per-converter switch, exactly as AutoTimeConversion does not
// reach nested plans either. Mask nested fields at the output layer with
// desensitize.Desensitization
func WithDesensitize(enabled bool) Option {
	return func(o *Options) {
		o.DesensitizeEnabled = enabled
	}
}

// WithSafeMode 启用/禁用安全模式
func WithSafeMode(enabled bool) Option {
	return func(o *Options) {
		o.SafeMode = enabled
	}
}

// WithConcurrency 设置并发数
func WithConcurrency(n int) Option {
	return func(o *Options) {
		o.Concurrency = n
	}
}

// WithBatchSize 设置批处理大小
func WithBatchSize(size int) Option {
	return func(o *Options) {
		o.BatchSize = size
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}
