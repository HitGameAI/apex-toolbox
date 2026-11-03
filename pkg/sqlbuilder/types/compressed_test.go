/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-07-30 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-07-30 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\types\compressed_test.go
 * @Description: CompressedText / CompressedTextMap 单元测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package types

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestCompressedText_Value_Scan(t *testing.T) {
	t.Run("空值", func(t *testing.T) {
		c := CompressedText("")
		v, err := c.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("普通文本压缩存储", func(t *testing.T) {
		original := CompressedText("hello world 你好世界")
		v, err := original.Value()
		require.NoError(t, err)
		require.NotNil(t, v)

		// 压缩后数据应带有 GZIP: 前缀
		b, ok := v.([]byte)
		require.True(t, ok)
		require.Greater(t, len(b), 5)
		assert.Equal(t, "GZIP:", string(b[:5]))

		// 扫描回来应还原原文
		var decoded CompressedText
		err = decoded.Scan(b)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("兼容未压缩历史数据", func(t *testing.T) {
		plain := []byte("plain text without gzip prefix")
		var decoded CompressedText
		err := decoded.Scan(plain)
		require.NoError(t, err)
		assert.Equal(t, "plain text without gzip prefix", decoded.String())
	})

	t.Run("nil scan", func(t *testing.T) {
		var decoded CompressedText
		err := decoded.Scan(nil)
		require.NoError(t, err)
		assert.Equal(t, "", decoded.String())
	})
}

func TestCompressedTextMap_Value_Scan(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var c CompressedTextMap
		v, err := c.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("正常 map 压缩存储", func(t *testing.T) {
		original := CompressedTextMap{
			"Content-Type":  {"application/json"},
			"Authorization": {"Bearer token"},
		}
		v, err := original.Value()
		require.NoError(t, err)
		require.NotNil(t, v)

		b, ok := v.([]byte)
		require.True(t, ok)
		require.Greater(t, len(b), 5)
		assert.Equal(t, "GZIP:", string(b[:5]))

		var decoded CompressedTextMap
		err = decoded.Scan(b)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("兼容未压缩历史数据", func(t *testing.T) {
		plain := []byte(`{"X-Custom":["val1","val2"]}`)
		var decoded CompressedTextMap
		err := decoded.Scan(plain)
		require.NoError(t, err)
		assert.Equal(t, []string{"val1", "val2"}, decoded["X-Custom"])
	})

	t.Run("empty map", func(t *testing.T) {
		c := CompressedTextMap{}
		v, err := c.Value()
		require.NoError(t, err)
		require.NotNil(t, v)

		var decoded CompressedTextMap
		err = decoded.Scan(v)
		require.NoError(t, err)
		assert.Empty(t, decoded)
	})

	t.Run("nil scan", func(t *testing.T) {
		var decoded CompressedTextMap
		err := decoded.Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, decoded)
	})
}

func TestCompressedTextMap_ToMap_FromMap(t *testing.T) {
	m := map[string][]string{"k": {"v1", "v2"}}
	c := CompressedTextMapFromMap(m)
	assert.Equal(t, m, c.ToMap())
}

// TestCompressedTypes_GormSchemaParse 回归测试：gorm schema 解析不得将压缩类型字段误判为关联关系
// 背景：CompressedTextMap 底层为 Map，未实现 GormDataTypeInterface 时 DataType 为空，
// gorm 会走 parseRelation → getOrParse 报 "unsupported data type"
func TestCompressedTypes_GormSchemaParse(t *testing.T) {
	type accessRecord struct {
		ID       string            `gorm:"column:id"`
		Headers  CompressedTextMap `gorm:"column:headers"`
		Body     CompressedText    `gorm:"column:body"`
		Response CompressedText    `gorm:"column:response"`
	}

	s, err := schema.Parse(&accessRecord{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	for _, name := range []string{"Headers", "Body", "Response"} {
		f := s.FieldsByName[name]
		require.NotNil(t, f)
		assert.Equal(t, schema.DataType("string"), f.DataType)
	}
	assert.Empty(t, s.Relationships.HasOne)
	assert.Empty(t, s.Relationships.HasMany)
}
