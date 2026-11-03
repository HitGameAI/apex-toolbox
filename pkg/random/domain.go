/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-05 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-01-05 15:16:27
 * @FilePath: \go-toolbox\pkg\random\domain.go
 * @Description: 域名相关随机生成工具
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package random

import (
	"sort"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// DomainKeywordBuilder 域名关键词生成器（支持链式调用）
type DomainKeywordBuilder struct {
	baseKeyword       string // 基础关键词
	count             int    // 生成数量
	minPrefixLen      int    // 最小前缀长度
	maxPrefixLen      int    // 最大前缀长度
	minSuffixLen      int    // 最小后缀长度
	maxSuffixLen      int    // 最大后缀长度
	maxBaseKeywordLen int    // baseKeyword 的最大长度，超过则随机截断
}

// NewDomainKeywordBuilder 创建域名关键词生成器
// 参数:
//   - baseKeyword: 基础关键词（如用户搜索的词）
//
// 示例:
//   - random.NewDomainKeywordBuilder("game").Generate()
//   - random.NewDomainKeywordBuilder("shop").WithCount(5).WithPrefixLength(2, 4).Generate()
func NewDomainKeywordBuilder(baseKeyword string) *DomainKeywordBuilder {
	return &DomainKeywordBuilder{
		baseKeyword:       baseKeyword,
		count:             3,
		minPrefixLen:      1,
		maxPrefixLen:      3,
		minSuffixLen:      1,
		maxSuffixLen:      3,
		maxBaseKeywordLen: 30,
	}
}

// WithCount 设置生成的关键词数量
func (b *DomainKeywordBuilder) WithCount(count int) *DomainKeywordBuilder {
	b.count = mathx.IfLeZero(count, b.count)
	return b
}

// WithPrefixLength 设置前缀长度范围
func (b *DomainKeywordBuilder) WithPrefixLength(min, max int) *DomainKeywordBuilder {
	b.maxPrefixLen = mathx.IfGt(max, 0, max, b.maxPrefixLen)
	b.minPrefixLen = mathx.IfGt(min, 0, min, b.minPrefixLen)
	// 确保最小值不大于最大值
	if b.minPrefixLen > b.maxPrefixLen {
		b.minPrefixLen, b.maxPrefixLen = b.maxPrefixLen, b.minPrefixLen
	}
	return b
}

// WithSuffixLength 设置后缀长度范围
func (b *DomainKeywordBuilder) WithSuffixLength(min, max int) *DomainKeywordBuilder {
	b.maxSuffixLen = mathx.IfGt(max, 0, max, b.maxSuffixLen)
	b.minSuffixLen = mathx.IfGt(min, 0, min, b.minSuffixLen)
	// 确保最小值不大于最大值
	if b.minSuffixLen > b.maxSuffixLen {
		b.minSuffixLen, b.maxSuffixLen = b.maxSuffixLen, b.minSuffixLen
	}
	return b
}

// WithMaxBaseKeywordLength 设置 baseKeyword 的最大长度
// 如果 baseKeyword 超过此长度，将从随机位置截取
func (b *DomainKeywordBuilder) WithMaxBaseKeywordLength(maxLen int) *DomainKeywordBuilder {
	b.maxBaseKeywordLen = mathx.IfGt(maxLen, 0, maxLen, b.maxBaseKeywordLen)
	return b
}

// Generate 生成随机关键词组合列表
func (b *DomainKeywordBuilder) Generate() []string {
	keywords := make([]string, b.count)
	for i := 0; i < b.count; i++ {
		// 处理 baseKeyword：如果超过最大长度，随机截取
		keyword := b.baseKeyword
		if len(keyword) > b.maxBaseKeywordLen {
			// 随机选择起始位置进行截取
			maxStart := len(keyword) - b.maxBaseKeywordLen
			start := RandInt(0, maxStart+1)
			keyword = keyword[start : start+b.maxBaseKeywordLen]
		}

		// 生成随机长度的前缀和后缀
		prefixLen := RandInt(b.minPrefixLen, b.maxPrefixLen+1)
		suffixLen := RandInt(b.minSuffixLen, b.maxSuffixLen+1)

		prefix := RandString(prefixLen, LOWERCASE|NUMBER)
		suffix := RandString(suffixLen, LOWERCASE|NUMBER)

		keywords[i] = prefix + keyword + suffix
	}

	return keywords
}

// GenerateAndJoinWithTLDs 生成关键词、拼接TLD并用分隔符连接
// 参数:
//   - tlds: TLD列表，如 ["com", "net", "org"]
//   - separator: 连接符，如 "," 或 ";"
//
// 返回:
//   - string: 连接后的字符串
//
// 示例:
//   - NewDomainKeywordBuilder("game").WithCount(2).GenerateAndJoinWithTLDs([]string{"com", "net"}, ",")
//     -> "ab1game23.com,ab1game23.net,xy5game67.com,xy5game67.net"
func (b *DomainKeywordBuilder) GenerateAndJoinWithTLDs(tlds []string, separator string) string {
	keywords := b.Generate()
	fullDomains := JoinDomainsWithTLDs(keywords, tlds)
	return strings.Join(fullDomains, separator)
}

// JoinDomainsWithTLDs 将域名关键词与多个TLD（顶级域名）拼接
// 参数:
//   - domains: 域名关键词列表（不包含后缀）
//   - tlds: TLD列表，如 ["com", "net", "org"]
//   - priorityTLDs: 优先级TLD列表（可选），指定哪些TLD应该排在前面，默认为 ["com"]
//
// 返回:
//   - []string: 拼接后的完整域名列表（按优先级TLD、其他TLD字母顺序分组）
//
// 示例:
//   - JoinDomainsWithTLDs([]string{"game", "shop"}, []string{"com", "net", "org"})
//     -> ["game.com", "shop.com", "game.net", "shop.net", "game.org", "shop.org"]
//   - JoinDomainsWithTLDs([]string{"game", "shop"}, []string{"com", "net", "org"}, "net", "com")
//     -> ["game.net", "shop.net", "game.com", "shop.com", "game.org", "shop.org"]
func JoinDomainsWithTLDs(domains []string, tlds []string, priorityTLDs ...string) []string {
	// 如果没有指定优先级TLD，默认使用 com
	if len(priorityTLDs) == 0 {
		priorityTLDs = []string{"com"}
	}

	// 创建优先级TLD的map，用于快速查找
	priorityMap := make(map[string]int)
	for i, tld := range priorityTLDs {
		priorityMap[tld] = i
	}

	// 复制并排序 TLD 列表
	sortedTLDs := make([]string, len(tlds))
	copy(sortedTLDs, tlds)
	sort.Slice(sortedTLDs, func(i, j int) bool {
		iPriority, iHasPriority := priorityMap[sortedTLDs[i]]
		jPriority, jHasPriority := priorityMap[sortedTLDs[j]]

		// 如果两个都有优先级，按优先级顺序排列
		if iHasPriority && jHasPriority {
			return iPriority < jPriority
		}
		// 有优先级的排在前面
		if iHasPriority {
			return true
		}
		if jHasPriority {
			return false
		}
		// 都没有优先级，按字母顺序排列
		return sortedTLDs[i] < sortedTLDs[j]
	})

	result := make([]string, 0, len(domains)*len(sortedTLDs))
	for _, tld := range sortedTLDs {
		for _, domain := range domains {
			result = append(result, domain+"."+tld)
		}
	}
	return result
}
