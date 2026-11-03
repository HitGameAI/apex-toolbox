/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-03-29 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-03-29 00:00:00
 * @FilePath: \go-toolbox\pkg\i18n\manager.go
 * @Description: 国际化管理器 - 核心翻译引擎（配置为本包自有 Config，零第三方依赖）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package i18n

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// 确保 Manager 实现 context.Translator 接口
var _ Translator = (*Manager)(nil)

// Manager 国际化管理器
// 核心翻译引擎，负责消息加载、语言解析和翻译
type Manager struct {
	config *Config

	// messages 以原子指针持有一份不可变的 语言 -> 键 -> 文案 表
	//
	// 翻译是读多写极少的路径：构造期装载一次，之后只有 ReloadLanguage 会改
	// 原先用 sync.RWMutex，每次 RLock/RUnlock 都要原子改同一个读者计数，
	// 多核并发下所有读者争抢同一条缓存行——实测并发反而比串行慢数倍
	// 改为读侧只做一次原子载入、写侧复制整表再替换（copy-on-write），
	// 读路径彻底无锁且不再有跨核写争用
	//
	// [EN] Translation is read-mostly: the table is filled at construction and
	// only ReloadLanguage mutates it. An RWMutex made every reader atomically
	// bump a shared reader count, so concurrent readers contended on one cache
	// line and parallel throughput was several times worse than sequential.
	// Readers now do a single atomic load; writers copy the whole table and swap
	// it in, which leaves the read path lock-free.
	messages atomic.Pointer[messageTable]

	// writeMu 只串行化写者（loadLanguage / ReloadLanguage），读者完全不参与
	// [EN] Serializes writers only; readers never touch it.
	writeMu sync.Mutex

	logger    Logger
	formatter Formatter
}

// messageTable 不可变的消息表：语言 -> 消息键 -> 文案
// [EN] Immutable message table; replaced wholesale on write.
type messageTable map[string]map[string]string

// NewManager 创建国际化管理器
// config: i18n 配置，为 nil 时使用 DefaultConfig()
func NewManager(config *Config) (*Manager, error) {
	return newManager(config, nil, nil)
}

// NewManagerWithLogger 创建带自定义日志的国际化管理器
func NewManagerWithLogger(config *Config, log Logger) (*Manager, error) {
	return newManager(config, log, nil)
}

// NewManagerWithFormatter 创建带自定义格式化器的国际化管理器
func NewManagerWithFormatter(config *Config, formatter Formatter) (*Manager, error) {
	return newManager(config, nil, formatter)
}

// NewManagerWith 创建同时指定日志与格式化器的国际化管理器
// log / formatter 为 nil 时各自使用默认值
func NewManagerWith(config *Config, log Logger, formatter Formatter) (*Manager, error) {
	return newManager(config, log, formatter)
}

// newManager 唯一的构造入口
//
// 日志器必须在预加载之前装好：映射目标语言预加载失败只记日志不返错，
// 若像先前那样构造完再塞 logger，这条告警会被默认空实现吞掉，
// 调用方传了 logger 也看不到任何东西
// [EN] The logger must be installed before preloading: a failure to preload a
// mapping target is only logged, not returned, so assigning the logger after
// construction (as before) silently dropped that warning even when the caller
// supplied one.
func newManager(config *Config, log Logger, formatter Formatter) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.MessageLoader == nil && config.MessagesPath != "" {
		config.MessageLoader = NewFileLoader(config.MessagesPath)
	}

	if config.MessageLoader == nil {
		return nil, NewError(ErrTypeMessageLoaderRequired)
	}

	if log == nil {
		log = nopLogger{}
	}
	if formatter == nil {
		formatter = &DefaultFormatter{}
	}

	mgr := &Manager{
		config:    config,
		logger:    log,
		formatter: formatter,
	}
	initial := make(messageTable, len(config.SupportedLanguages))
	mgr.messages.Store(&initial)

	for _, lang := range config.SupportedLanguages {
		if err := mgr.loadLanguage(lang); err != nil {
			return nil, NewError(ErrTypeLanguageLoadFailed, fmt.Sprintf("语言 %s: %v", lang, err))
		}
	}

	// 预加载映射目标语言：缺失不致命，仅告警，避免运行期首次命中时才去读盘
	// [EN] Preload mapping targets; a miss is non-fatal and only warned about.
	for _, targetLang := range config.LanguageMapping {
		if _, exists := mgr.load()[targetLang]; exists {
			continue
		}
		if err := mgr.loadLanguage(targetLang); err != nil {
			mgr.logger.Warn("预加载目标语言 %s 失败: %v", targetLang, err)
		}
	}

	return mgr, nil
}

// loadLanguage 加载指定语言的消息并原子替换整表
// [EN] Loads one language and atomically swaps in a new table.
func (m *Manager) loadLanguage(language string) error {
	messages, err := m.config.MessageLoader.LoadMessages(language)
	if err != nil {
		return err
	}

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	current := m.load()
	next := make(messageTable, len(current)+1)
	for lang, msgs := range current {
		next[lang] = msgs
	}
	next[language] = messages
	m.messages.Store(&next)

	return nil
}

// load 读取当前消息表，未初始化时返回 nil（对 map 读取与 len/range 均安全）
// [EN] Loads the current table; a nil map is safe to read, len and range.
func (m *Manager) load() messageTable {
	if t := m.messages.Load(); t != nil {
		return *t
	}
	return nil
}

// GetMessage 获取翻译消息（实现 Translator 接口）
func (m *Manager) GetMessage(language, key string, args ...any) string {
	return m.getMessageInternal(language, key, args, nil)
}

// GetMessageWithMap 使用 map 模板数据获取翻译消息（实现 Translator 接口）
func (m *Manager) GetMessageWithMap(language, key string, templateData map[string]any) string {
	return m.getMessageInternal(language, key, nil, templateData)
}

// getMessageInternal 内部获取消息的实现
func (m *Manager) getMessageInternal(language, key string, args []any, templateData map[string]any) string {
	// 整个翻译过程只在此处做一次原子载入，后续查表与格式化都不持有任何锁
	// [EN] One atomic load for the whole operation; no lock is held while formatting.
	table := m.load()

	if message := m.getMessageFromLanguage(table, language, key); message != "" {
		return m.formatter.Format(message, args, templateData)
	}

	if m.config.EnableFallback && language != m.config.DefaultLanguage {
		if message := m.getMessageFromLanguage(table, m.config.DefaultLanguage, key); message != "" {
			return m.formatter.Format(message, args, templateData)
		}
	}

	return key
}

// getMessageFromLanguage 从指定语言获取消息
func (m *Manager) getMessageFromLanguage(table messageTable, language, key string) string {
	return table[m.config.ResolveLanguage(language)][key]
}

// IsLanguageSupported 检查语言是否被支持（实现 Translator 接口）
func (m *Manager) IsLanguageSupported(language string) bool {
	return m.config.IsSupportedLanguage(language)
}

// GetConfig 获取当前配置
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetSupportedLanguages 获取支持的语言列表
func (m *Manager) GetSupportedLanguages() []string {
	return m.config.SupportedLanguages
}

// GetDefaultLanguage 获取默认语言
func (m *Manager) GetDefaultLanguage() string {
	return m.config.DefaultLanguage
}

// ReloadLanguage 重新加载指定语言的消息
func (m *Manager) ReloadLanguage(language string) error {
	return m.loadLanguage(language)
}

// ResolveLanguage 解析语言代码（委托给 Config.ResolveLanguage）
func (m *Manager) ResolveLanguage(lang string) string {
	return m.config.ResolveLanguage(lang)
}

// HasLanguage 检查指定语言的消息是否已加载
func (m *Manager) HasLanguage(language string) bool {
	_, exists := m.load()[language]
	return exists
}

// GetLoadedLanguages 获取已加载的语言列表
func (m *Manager) GetLoadedLanguages() []string {
	table := m.load()

	languages := make([]string, 0, len(table))
	for lang := range table {
		languages = append(languages, lang)
	}
	return languages
}

// GetMessageKeys 获取指定语言的所有消息键
func (m *Manager) GetMessageKeys(language string) []string {
	if langMessages, exists := m.load()[language]; exists {
		keys := make([]string, 0, len(langMessages))
		for key := range langMessages {
			keys = append(keys, key)
		}
		return keys
	}
	return nil
}

// IsEnabled 检查国际化是否启用
func (m *Manager) IsEnabled() bool {
	return m.config.IsEnabled()
}
