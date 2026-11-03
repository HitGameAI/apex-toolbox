/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\manager_test.go
 * @Description: 国际化管理器测试
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testLocalesPath = "locales"
	testFiles       = []string{
		"locales/en.json",
		"locales/zh.json",
	}
)

func TestMain(m *testing.M) {
	setupTestFiles()
	code := m.Run()
	cleanupTestFiles()
	os.Exit(code)
}

func setupTestFiles() {
	os.MkdirAll(testLocalesPath, 0755)
	os.WriteFile(testFiles[0], []byte(`{"hello": "Hello", "greeting": "Hello %s"}`), 0644)
	os.WriteFile(testFiles[1], []byte(`{"hello": "你好", "greeting": "你好 %s"}`), 0644)
}

func cleanupTestFiles() {
	os.RemoveAll(testLocalesPath)
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{name: "使用默认配置", config: nil, expectError: false},
		{name: "使用自定义配置", config: &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}, expectError: false},
		{name: "无消息加载器", config: &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: nil, MessagesPath: ""}, expectError: true},
		{name: "使用 MessagesPath", config: &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: nil, MessagesPath: testLocalesPath}, expectError: false},
		{name: "带语言映射", config: &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true, LanguageMapping: map[string]string{"en-US": "en"}}, expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

func TestNewManagerWithLogger(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}

	tests := []struct {
		name     string
		logger   Logger
		expected bool
	}{
		{name: "带自定义日志", logger: &recordingLogger{}, expected: true},
		{name: "带 nil 日志", logger: nil, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManagerWithLogger(config, tt.logger)
			assert.NoError(t, err)
			assert.NotNil(t, manager)
		})
	}
}

func TestNewManagerWithFormatter(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}

	tests := []struct {
		name      string
		formatter Formatter
		expected  bool
	}{
		{name: "带自定义格式化器", formatter: &DefaultFormatter{}, expected: true},
		{name: "带 nil 格式化器", formatter: nil, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManagerWithFormatter(config, tt.formatter)
			assert.NoError(t, err)
			assert.NotNil(t, manager)
		})
	}
}

func TestManager_loadLanguage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	err = manager.loadLanguage("en")
	assert.NoError(t, err)
}

func TestManager_GetMessage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello", "greeting": "Hello %s"}, "zh": {"hello": "你好", "greeting": "你好 %s"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		language string
		key      string
		args     []any
		expected string
	}{
		{name: "英文消息", language: "en", key: "hello", args: nil, expected: "Hello"},
		{name: "中文消息", language: "zh", key: "hello", args: nil, expected: "你好"},
		{name: "带参数的消息", language: "en", key: "greeting", args: []any{"John"}, expected: "Hello John"},
		{name: "不存在的键", language: "en", key: "non_existent", args: nil, expected: "non_existent"},
		{name: "语言回退", language: "fr", key: "hello", args: nil, expected: "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetMessage(tt.language, tt.key, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_GetMessageWithMap(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"greeting": "Hello {name}"}, "zh": {"greeting": "你好 {name}"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)

	tests := []struct {
		name         string
		language     string
		key          string
		templateData map[string]any
		expected     string
	}{
		{name: "英文消息带命名参数", language: "en", key: "greeting", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
		{name: "中文消息带命名参数", language: "zh", key: "greeting", templateData: map[string]any{"name": "John"}, expected: "你好 John"},
		{name: "不存在的键", language: "en", key: "non_existent", templateData: map[string]any{"name": "John"}, expected: "non_existent"},
		{name: "语言回退", language: "fr", key: "greeting", templateData: map[string]any{"name": "John"}, expected: "Hello John"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetMessageWithMap(tt.language, tt.key, tt.templateData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_IsLanguageSupported(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)

	tests := []struct {
		language string
		expected bool
	}{
		{language: "en", expected: true},
		{language: "zh", expected: true},
		{language: "fr", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := manager.IsLanguageSupported(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_GetDefaultLanguage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.Equal(t, "en", manager.GetDefaultLanguage())
}

func TestManager_GetSupportedLanguages(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.Len(t, manager.GetSupportedLanguages(), 2)
}

func TestManager_GetConfig(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager.GetConfig())
	assert.Equal(t, "en", manager.GetConfig().DefaultLanguage)
}

func TestManager_ReloadLanguage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.NoError(t, manager.ReloadLanguage("en"))
}

func TestManager_ResolveLanguage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.Equal(t, "en", manager.ResolveLanguage("en-US"))
}

func TestManager_HasLanguage(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)

	tests := []struct {
		language string
		expected bool
	}{
		{language: "en", expected: true},
		{language: "zh", expected: true},
		{language: "fr", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			assert.Equal(t, tt.expected, manager.HasLanguage(tt.language))
		})
	}
}

func TestManager_GetLoadedLanguages(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)
	assert.Len(t, manager.GetLoadedLanguages(), 2)
}

func TestManager_GetMessageKeys(t *testing.T) {
	config := &Config{Enabled: true, DefaultLanguage: "en", SupportedLanguages: []string{"en", "zh"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello", "greeting": "Hello %s"}, "zh": {"hello": "你好"}}), EnableFallback: true}
	manager, err := NewManager(config)
	assert.NoError(t, err)

	tests := []struct {
		language    string
		expectedLen int
		expectedNil bool
	}{
		{language: "en", expectedLen: 2, expectedNil: false},
		{language: "zh", expectedLen: 1, expectedNil: false},
		{language: "fr", expectedLen: 0, expectedNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			keys := manager.GetMessageKeys(tt.language)
			if tt.expectedNil {
				assert.Nil(t, keys)
			} else {
				assert.Len(t, keys, tt.expectedLen)
			}
		})
	}
}

func TestManager_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{name: "启用国际化", enabled: true, expected: true},
		{name: "禁用国际化", enabled: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Enabled: tt.enabled, DefaultLanguage: "en", SupportedLanguages: []string{"en"}, MessageLoader: NewMapLoader(map[string]map[string]string{"en": {"hello": "Hello"}}), EnableFallback: true}
			manager, err := NewManager(config)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, manager.IsEnabled())
		})
	}
}

// recordingLogger 记录 Warn 调用的假 logger，用于验证日志确实被接上
// [EN] A fake logger that records Warn calls, so the tests can assert the
// injected logger is actually used.
type recordingLogger struct {
	warns []string
}

func (r *recordingLogger) Warn(format string, args ...any) {
	r.warns = append(r.warns, fmt.Sprintf(format, args...))
}

// failingLoader 只对 en 成功，其余语言一律失败，用来触发预加载告警路径
type failingLoader struct{}

func (failingLoader) LoadMessages(language string) (map[string]string, error) {
	if language == "en" {
		return map[string]string{"hello": "Hello"}, nil
	}
	return nil, NewError(ErrTypeLanguageNotFound, language)
}

// TestNewManagerWithLogger_WarnsOnMappingPreloadFailure 验证映射目标语言预加载失败时
// 会走注入的 logger，而不是静默吞掉
func TestNewManagerWithLogger_WarnsOnMappingPreloadFailure(t *testing.T) {
	rec := &recordingLogger{}
	config := &Config{
		Enabled:            true,
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en"},
		MessageLoader:      failingLoader{},
		EnableFallback:     true,
		// 映射目标 ja 不在 SupportedLanguages 里，预加载时会失败
		LanguageMapping: map[string]string{"ja-jp": "ja"},
	}

	manager, err := NewManagerWithLogger(config, rec)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotEmpty(t, rec.warns, "预加载 ja 失败应当产生一条告警")
	assert.Contains(t, rec.warns[0], "ja")
}

// TestManager_ConcurrentReadWhileReload 并发读 + 并发重载：
// 验证 copy-on-write 下读者永远看到一份完整快照，不会读到半更新的表
// [EN] Concurrent reads against concurrent reloads: every reader must observe a
// complete snapshot, never a half-updated table.
func TestManager_ConcurrentReadWhileReload(t *testing.T) {
	mgr, err := NewManager(&Config{
		Enabled:            true,
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "zh"},
		EnableFallback:     true,
		MessageLoader: NewMapLoader(map[string]map[string]string{
			"en": {"hello": "Hello"},
			"zh": {"hello": "你好"},
		}),
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// 8 个读者持续翻译
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if got := mgr.GetMessage("zh", "hello"); got != "你好" {
					t.Errorf("读到非预期文案: %q", got)
					return
				}
				if got := mgr.GetMessage("en", "hello"); got != "Hello" {
					t.Errorf("读到非预期文案: %q", got)
					return
				}
			}
		}()
	}

	// 2 个写者不断重载
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				require.NoError(t, mgr.ReloadLanguage("zh"))
				require.NoError(t, mgr.ReloadLanguage("en"))
			}
		}()
	}

	// 等写者跑完再停读者
	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()

	// 重载过后两种语言都必须仍然在表里
	assert.True(t, mgr.HasLanguage("zh"))
	assert.True(t, mgr.HasLanguage("en"))
	assert.ElementsMatch(t, []string{"en", "zh"}, mgr.GetLoadedLanguages())
}
