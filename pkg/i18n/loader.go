/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\loader.go
 * @Description: 消息加载器：接口、嵌套扁平化与三个内置实现
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loader 消息加载器接口
// 定义从不同数据源加载国际化翻译消息的标准契约
// 与 go-config/pkg/i18n.MessageLoader 接口兼容，所有实现均同时满足两个接口
type Loader interface {
	// LoadMessages 加载指定语言的翻译消息
	// language: 语言代码，如 "zh", "en", "ja" 等
	// 返回该语言的消息映射（键为消息标识，值为翻译文本）
	LoadMessages(language string) (map[string]string, error)
}

// FlattenToMessages 将嵌套 map 扁平化为点号分隔的消息映射
// 例如: {"error": {"internal": "错误"}} -> {"error.internal": "错误"}
//
// 单趟递归直接写入结果 map，并复用一个 []byte 拼接前缀，避免原先
// mathx.FlattenMap + TransformMapValues 两趟遍历与中间 map 的开销
// [EN] One recursive pass writing straight into the result map, reusing a byte
// buffer for the key prefix; the previous version walked twice through an
// intermediate map.
func FlattenToMessages(data map[string]any) map[string]string {
	out := make(map[string]string, len(data))
	flattenInto(out, make([]byte, 0, 64), data)
	return out
}

// flattenInto 递归展开嵌套 map，prefix 为已拼接的键前缀
func flattenInto(out map[string]string, prefix []byte, data map[string]any) {
	for k, v := range data {
		key := prefix
		if len(key) > 0 {
			key = append(key, '.')
		}
		key = append(key, k...)

		switch val := v.(type) {
		case string:
			out[string(key)] = val
		case map[string]any:
			flattenInto(out, key, val)
		case map[any]any:
			// YAML 解析可能产出 map[any]any
			nested := make(map[string]any, len(val))
			for nk, nv := range val {
				nested[fmt.Sprintf("%v", nk)] = nv
			}
			flattenInto(out, key, nested)
		default:
			out[string(key)] = fmt.Sprintf("%v", val)
		}
	}
}

// AsMessageLoader 将 Loader 视作 MessageLoader
// MessageLoader 现在是 Loader 的类型别名，故直接返回即可；保留此函数仅为兼容既有调用方
// [EN] MessageLoader is now an alias of Loader, so this is the identity function;
// kept only for API compatibility.
func AsMessageLoader(l Loader) MessageLoader {
	return l
}

// 确保 FileLoader 同时实现 Loader 和 MessageLoader 接口
var (
	_ Loader = (*FileLoader)(nil)
)

// FileLoader 文件系统消息加载器
// 从指定目录下的 JSON 文件加载翻译消息，支持嵌套 JSON 自动扁平化
// 文件命名规则: {localesPath}/{language}.json，如 ./locales/zh.json
type FileLoader struct {
	// localesPath 翻译文件所在目录路径
	localesPath string
}

// NewFileLoader 创建文件系统消息加载器
// localesPath: 翻译文件所在目录路径，如 "./locales" 或 "resources/locales"
func NewFileLoader(localesPath string) *FileLoader {
	return &FileLoader{
		localesPath: localesPath,
	}
}

// LoadMessages 加载指定语言的消息
// 读取 {localesPath}/{language}.json 文件，支持嵌套 JSON 自动扁平化为点号格式
// language: 语言代码，如 "zh", "en", "ja" 等
func (f *FileLoader) LoadMessages(language string) (map[string]string, error) {
	filePath := filepath.Join(f.localesPath, language+".json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, NewError(ErrTypeLoaderFileNotFound, fmt.Sprintf("路径: %s", filePath))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewError(ErrTypeLoaderReadFailed, fmt.Sprintf("文件: %s, 错误: %v", filePath, err))
	}

	var nested map[string]any
	if err := json.Unmarshal(data, &nested); err != nil {
		return nil, NewError(ErrTypeLoaderParseFailed, fmt.Sprintf("文件: %s, 错误: %v", filePath, err))
	}

	return FlattenToMessages(nested), nil
}

// 确保 JSONLoader 同时实现 Loader 和 MessageLoader 接口
var (
	_ Loader = (*JSONLoader)(nil)
)

// JSONLoader JSON 消息加载器
// 从 JSON 字符串直接解析翻译消息，适用于配置文件内嵌或动态生成的场景
// JSON 格式为: {"zh": {"key": "值"}, "en": {"key": "value"}}
type JSONLoader struct {
	// messages 已解析的语言消息映射
	messages map[string]map[string]string
}

// NewJSONLoader 创建 JSON 消息加载器
// messagesJSON: JSON 格式的翻译消息字符串
// 格式示例: {"zh": {"hello": "你好"}, "en": {"hello": "Hello"}}
func NewJSONLoader(messagesJSON string) (*JSONLoader, error) {
	var messages map[string]map[string]string
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return nil, NewError(ErrTypeLoaderParseFailed, fmt.Sprintf("JSON 解析失败: %v", err))
	}

	return &JSONLoader{messages: messages}, nil
}

// LoadMessages 加载指定语言的消息
func (j *JSONLoader) LoadMessages(language string) (map[string]string, error) {
	if messages, exists := j.messages[language]; exists {
		return messages, nil
	}
	return nil, NewError(ErrTypeLoaderLanguageNotFound, fmt.Sprintf("语言: %s", language))
}

// 确保 MapLoader 同时实现 Loader 和 MessageLoader 接口
var (
	_ Loader = (*MapLoader)(nil)
)

// MapLoader 内存 Map 消息加载器
// 直接从内存中的 map 数据结构加载翻译消息
// 适用于程序启动时硬编码或动态构建的翻译数据
type MapLoader struct {
	// messages 语言消息映射
	messages map[string]map[string]string
}

// NewMapLoader 创建内存 Map 消息加载器
// messages: 语言消息映射，格式为 map[语言代码]map[消息键]翻译文本
// 示例: map[string]map[string]string{"zh": {"hello": "你好"}, "en": {"hello": "Hello"}}
func NewMapLoader(messages map[string]map[string]string) *MapLoader {
	return &MapLoader{messages: messages}
}

// LoadMessages 加载指定语言的消息
func (m *MapLoader) LoadMessages(language string) (map[string]string, error) {
	if messages, exists := m.messages[language]; exists {
		return messages, nil
	}
	return nil, NewError(ErrTypeLoaderLanguageNotFound, fmt.Sprintf("语言: %s", language))
}
