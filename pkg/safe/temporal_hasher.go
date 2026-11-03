/**
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-03-09 17:12:35
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-03-09 17:32:18
 * @FilePath: \go-toolbox\pkg\safe\temporal_hasher.go
 * @Description: 临时哈希生成器，基于时间窗口生成短期一致性哈希
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package safe

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// temporalHasherConfig 临时哈希生成器配置（私有）
type temporalHasherConfig struct {
	window    time.Duration
	length    int
	separator string
}

// TemporalHasherOption 配置选项函数
type TemporalHasherOption func(*temporalHasherConfig)

// WithWindow 设置时间窗口
func WithWindow(window time.Duration) TemporalHasherOption {
	return func(c *temporalHasherConfig) {
		c.window = window
	}
}

// WithLength 设置哈希长度
func WithLength(length int) TemporalHasherOption {
	return func(c *temporalHasherConfig) {
		c.length = length
	}
}

// WithSeparator 设置分隔符
func WithSeparator(separator string) TemporalHasherOption {
	return func(c *temporalHasherConfig) {
		c.separator = separator
	}
}

// TemporalHasher 临时哈希生成器
// 在时间窗口内，相同输入生成相同哈希；超过窗口后生成新哈希
//
// 使用场景：
// - WebSocket ClientID：同一用户+设备短期内复用连接标识
// - 会话标识：短期内相同条件复用会话
// - 防重放：基于时间窗口的请求去重
// - 缓存键：短期内相同条件复用缓存
type TemporalHasher struct {
	config *temporalHasherConfig
}

// NewTemporalHasher 创建临时哈希生成器
func NewTemporalHasher(opts ...TemporalHasherOption) *TemporalHasher {
	// 空配置，等待 options 填充
	config := &temporalHasherConfig{}

	// 应用选项
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}

	// 若为空则使用默认值
	config.window = mathx.IfLeZero(config.window, 5*time.Minute)
	config.length = mathx.IfLeZero(config.length, 12)
	config.separator = mathx.IfEmpty(config.separator, "|")

	return &TemporalHasher{config: config}
}

// Hash 生成临时哈希
// parts: 参与哈希计算的各个部分（会自动排序以保证一致性）
func (h *TemporalHasher) Hash(parts ...string) string {
	return h.HashAt(time.Now(), parts...)
}

// HashAt 使用指定时间生成哈希
func (h *TemporalHasher) HashAt(t time.Time, parts ...string) string {
	slot := h.timeSlot(t)
	data := h.buildData(slot, parts...)
	return h.computeHash(data)
}

// HashMap 使用 map 生成哈希（会自动排序 key）
func (h *TemporalHasher) HashMap(kvMap map[string]string) string {
	return h.HashMapAt(time.Now(), kvMap)
}

// HashMapAt 使用指定时间和 map 生成哈希
func (h *TemporalHasher) HashMapAt(t time.Time, kvMap map[string]string) string {
	slot := h.timeSlot(t)

	// 提取并排序 keys
	keys := make([]string, 0, len(kvMap))
	for k := range kvMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建 parts
	parts := make([]string, 0, len(kvMap)+1)
	parts = append(parts, fmt.Sprintf("%d", slot))

	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, kvMap[k]))
	}

	data := strings.Join(parts, h.config.separator)
	return h.computeHash(data)
}

// IsExpired 检查哈希是否已过期（超过时间窗口）
func (h *TemporalHasher) IsExpired(hash string, parts ...string) bool {
	current := h.Hash(parts...)
	return hash != current
}

// IsExpiredMap 检查哈希是否已过期（使用 map 参数）
func (h *TemporalHasher) IsExpiredMap(hash string, kvMap map[string]string) bool {
	current := h.HashMap(kvMap)
	return hash != current
}

// Window 获取配置的时间窗口
func (h *TemporalHasher) Window() time.Duration {
	return h.config.window
}

// Length 获取配置的哈希长度
func (h *TemporalHasher) Length() int {
	return h.config.length
}

// timeSlot 获取时间槽（将时间按窗口分段）
// 使用纳秒精度避免大窗口时的精度损失
func (h *TemporalHasher) timeSlot(t time.Time) int64 {
	return t.UnixNano() / int64(h.config.window)
}

// buildData 构建待哈希数据
func (h *TemporalHasher) buildData(slot int64, parts ...string) string {
	// 排序 parts 以保证一致性
	sorted := make([]string, len(parts))
	copy(sorted, parts)
	sort.Strings(sorted)

	// 添加时间槽
	all := make([]string, 0, len(sorted)+1)
	all = append(all, fmt.Sprintf("%d", slot))
	all = append(all, sorted...)

	return strings.Join(all, h.config.separator)
}

// computeHash 计算哈希（使用 ShortHashWithLength）
func (h *TemporalHasher) computeHash(data string) string {
	// 使用 safe.ShortHashWithLength 生成短哈希
	// 优点：
	// 1. 天然小写（0-9a-z）
	// 2. 固定长度（自动补齐或截取）
	// 3. 基于 FNV-1a 算法，性能优异
	// 4. 碰撞概率可控
	return ShortHashWithLength(data, h.config.length)
}
