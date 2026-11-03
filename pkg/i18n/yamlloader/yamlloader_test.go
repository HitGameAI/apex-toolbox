/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 13:12:33
 * @FilePath: \go-toolbox\pkg\i18n\yamlloader\yamlloader_test.go
 * @Description: YAML 消息加载器测试（随加载器从主包迁入）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package yamlloader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupYAMLTestFiles(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "en.yaml"), []byte("hello: Hello\ngreeting: Hello {name}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "zh.yml"), []byte("hello: 你好\ngreeting: 你好 {name}"), 0o644))
	// 嵌套结构：验证扁平化为点号格式
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ja.yaml"), []byte("error:\n  internal: 内部エラー\n  notfound: 見つかりません"), 0o644))
	return dir
}

func TestNew(t *testing.T) {
	loader := New("./locales")
	assert.NotNil(t, loader)
	assert.Equal(t, "./locales", loader.localesPath)
}

func TestFileLoader_LoadMessages(t *testing.T) {
	dir := setupYAMLTestFiles(t)
	loader := New(dir)

	tests := []struct {
		name        string
		language    string
		expectError bool
		expectedLen int
	}{
		{name: "加载 .yaml 文件", language: "en", expectError: false, expectedLen: 2},
		{name: "加载 .yml 文件", language: "zh", expectError: false, expectedLen: 2},
		{name: "文件不存在", language: "fr", expectError: true, expectedLen: 0},
		{name: "嵌套结构扁平化", language: "ja", expectError: false, expectedLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := loader.LoadMessages(tt.language)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, messages)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, messages, tt.expectedLen)
		})
	}
}

// TestFileLoader_Flatten 验证嵌套 YAML 确实被展平成点号键
func TestFileLoader_Flatten(t *testing.T) {
	dir := setupYAMLTestFiles(t)

	messages, err := New(dir).LoadMessages("ja")
	require.NoError(t, err)
	assert.Equal(t, "内部エラー", messages["error.internal"])
	assert.Equal(t, "見つかりません", messages["error.notfound"])
}

func TestStringLoader(t *testing.T) {
	loader, err := NewString("zh:\n  hello: 你好\nen:\n  hello: Hello")
	require.NoError(t, err)

	zh, err := loader.LoadMessages("zh")
	require.NoError(t, err)
	assert.Equal(t, "你好", zh["hello"])

	en, err := loader.LoadMessages("en")
	require.NoError(t, err)
	assert.Equal(t, "Hello", en["hello"])

	_, err = loader.LoadMessages("fr")
	assert.Error(t, err)
}

func TestStringLoader_InvalidYAML(t *testing.T) {
	_, err := NewString("zh:\n\thello: 你好")
	assert.Error(t, err, "YAML 不允许用制表符缩进，应当报错")
}
