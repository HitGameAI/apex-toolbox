/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\loader_file_test.go
 * @Description: 文件系统消息加载器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader_LoadMessages(t *testing.T) {
	testDir := "test_files"
	testFiles := []string{
		"test_files/en.json",
		"test_files/zh.json",
	}

	// 准备测试文件
	os.MkdirAll(testDir, 0755)
	os.WriteFile(testFiles[0], []byte(`{"hello": "Hello", "greeting": "Hello {name}"}`), 0644)
	os.WriteFile(testFiles[1], []byte(`{"hello": "你好", "greeting": "你好 {name}"}`), 0644)
	defer os.RemoveAll(testDir)

	loader := NewFileLoader(testDir)

	tests := []struct {
		name        string
		language    string
		expectError bool
		expectedLen int
	}{
		{name: "加载英文文件", language: "en", expectError: false, expectedLen: 2},
		{name: "加载中文文件", language: "zh", expectError: false, expectedLen: 2},
		{name: "文件不存在", language: "fr", expectError: true, expectedLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := loader.LoadMessages(tt.language)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				assert.Len(t, messages, tt.expectedLen)
			}
		})
	}
}
