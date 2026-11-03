/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-18 13:12:33
 * @FilePath: \go-toolbox\pkg\i18n\yamlloader\yamlloader.go
 * @Description: YAML 消息加载器（独立子包：唯一需要 YAML 解析器的地方）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

// Package yamlloader 提供从 YAML 文件或字符串加载翻译消息的加载器
//
// 它单独成包，是为了让 go-toolbox/pkg/i18n 主包保持零第三方依赖：只有真的用 YAML 存放
// 翻译文件时才需要引入 YAML 解析器
// 用 JSON 或内存 map 的使用方完全不会碰到它
//
// [EN] A separate package so the go-toolbox/pkg/i18n root stays dependency-free: a YAML
// parser is only pulled in by callers that actually keep locales in YAML.
// Callers using JSON or in-memory maps never touch it.
package yamlloader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	i18n "github.com/HitGameAI/apex-toolbox/pkg/i18n"
)

// 确保两个加载器都实现主包的 Loader 接口
var (
	_ i18n.Loader = (*FileLoader)(nil)
	_ i18n.Loader = (*StringLoader)(nil)
)

// FileLoader YAML 文件消息加载器
// 文件命名规则: {localesPath}/{language}.yaml 或 {localesPath}/{language}.yml
type FileLoader struct {
	localesPath string
}

// New 创建 YAML 文件消息加载器
// localesPath: 翻译文件所在目录，如 "./locales"
func New(localesPath string) *FileLoader {
	return &FileLoader{localesPath: localesPath}
}

// LoadMessages 加载指定语言的消息
// 依次尝试 {language}.yaml 与 {language}.yml，支持嵌套结构自动扁平化为点号格式
//
// 直接把 YAML 解析成 map 后扁平化；原实现走的是 YAML -> JSON 字节 -> map，
// 等于解析两遍还多一次序列化
// [EN] Parses YAML straight into a map; the previous implementation went
// YAML -> JSON bytes -> map, parsing twice with a serialization in between.
func (f *FileLoader) LoadMessages(language string) (map[string]string, error) {
	yamlPath := filepath.Join(f.localesPath, language+".yaml")
	ymlPath := filepath.Join(f.localesPath, language+".yml")

	filePath := ""
	if _, err := os.Stat(yamlPath); err == nil {
		filePath = yamlPath
	} else if _, err := os.Stat(ymlPath); err == nil {
		filePath = ymlPath
	}

	if filePath == "" {
		return nil, i18n.NewError(i18n.ErrTypeLoaderFileNotFound, fmt.Sprintf("路径: %s 或 %s", yamlPath, ymlPath))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, i18n.NewError(i18n.ErrTypeLoaderReadFailed, fmt.Sprintf("文件: %s, 错误: %v", filePath, err))
	}

	var nested map[string]any
	if err := yaml.Unmarshal(data, &nested); err != nil {
		return nil, i18n.NewError(i18n.ErrTypeLoaderParseFailed, fmt.Sprintf("文件: %s, YAML 解析失败: %v", filePath, err))
	}

	return i18n.FlattenToMessages(nested), nil
}

// StringLoader YAML 字符串消息加载器
// 适用于配置内嵌或动态生成的场景，格式为 语言 -> 键 -> 文案
type StringLoader struct {
	messages map[string]map[string]string
}

// NewString 创建 YAML 字符串消息加载器
//
// 格式示例:
//
//	zh:
//	  hello: 你好
//	en:
//	  hello: Hello
func NewString(yamlStr string) (*StringLoader, error) {
	var messages map[string]map[string]string
	if err := yaml.Unmarshal([]byte(yamlStr), &messages); err != nil {
		return nil, i18n.NewError(i18n.ErrTypeLoaderParseFailed, fmt.Sprintf("YAML 解析失败: %v", err))
	}

	return &StringLoader{messages: messages}, nil
}

// LoadMessages 加载指定语言的消息
func (s *StringLoader) LoadMessages(language string) (map[string]string, error) {
	if messages, exists := s.messages[language]; exists {
		return messages, nil
	}
	return nil, i18n.NewError(i18n.ErrTypeLoaderLanguageNotFound, fmt.Sprintf("语言: %s", language))
}
