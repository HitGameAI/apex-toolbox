/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\option_test.go
 * @Description: 选项模式测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.True(t, opts.AutoTimeConversion, "默认应启用自动时间转换")
	assert.Nil(t, opts.FieldMapping, "字段映射应惰性初始化，默认不分配")
	assert.Empty(t, opts.FieldMapping, "默认字段映射应为空")
	assert.True(t, opts.TagMappingEnabled, "默认应启用tag映射")
	assert.Equal(t, "pbmo", opts.TagName, "默认tag名应为pbmo")
	assert.False(t, opts.DesensitizeEnabled, "默认不启用脱敏")
	assert.False(t, opts.SafeMode, "默认不启用安全模式")
	assert.Equal(t, 100, opts.BatchSize, "默认批处理大小应为100")
	assert.Equal(t, 30*time.Second, opts.Timeout, "默认超时应为30秒")
}

func TestApplyOptions(t *testing.T) {
	opts := ApplyOptions(
		WithAutoTimeConversion(false),
		WithDesensitize(true),
		WithSafeMode(true),
		WithBatchSize(50),
		WithTimeout(10*time.Second),
	)

	assert.False(t, opts.AutoTimeConversion)
	assert.True(t, opts.DesensitizeEnabled)
	assert.True(t, opts.SafeMode)
	assert.Equal(t, 50, opts.BatchSize)
	assert.Equal(t, 10*time.Second, opts.Timeout)
}

func TestWithFieldMapping(t *testing.T) {
	opts := ApplyOptions(
		WithFieldMapping("ID", "Id"),
		WithFieldMapping("Name", "UserName"),
	)

	assert.Equal(t, "Id", opts.FieldMapping["ID"])
	assert.Equal(t, "UserName", opts.FieldMapping["Name"])
}

func TestWithFieldMappings(t *testing.T) {
	opts := ApplyOptions(
		WithFieldMappings(map[string]string{
			"ID":    "Id",
			"Email": "UserEmail",
		}),
	)

	assert.Equal(t, "Id", opts.FieldMapping["ID"])
	assert.Equal(t, "UserEmail", opts.FieldMapping["Email"])
}

func TestWithTagMapping(t *testing.T) {
	opts := ApplyOptions(WithTagMapping(false))
	assert.False(t, opts.TagMappingEnabled)
}

func TestWithTagName(t *testing.T) {
	opts := ApplyOptions(WithTagName("custom"))
	assert.Equal(t, "custom", opts.TagName)
}

func TestWithConcurrency(t *testing.T) {
	opts := ApplyOptions(WithConcurrency(4))
	assert.Equal(t, 4, opts.Concurrency)
}

func TestApplyOptionsEmpty(t *testing.T) {
	opts := ApplyOptions()
	defaults := DefaultOptions()

	assert.Equal(t, defaults.AutoTimeConversion, opts.AutoTimeConversion)
	assert.Equal(t, defaults.TagMappingEnabled, opts.TagMappingEnabled)
	assert.Equal(t, defaults.TagName, opts.TagName)
	assert.Equal(t, defaults.BatchSize, opts.BatchSize)
}
