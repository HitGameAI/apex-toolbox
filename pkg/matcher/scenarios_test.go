/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-15 10:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-15 10:00:00
 * @FilePath: \go-toolbox\pkg\matcher\scenarios_test.go
 * @Description: 匹配器场景测试 - 模拟真实业务场景
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package matcher

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/contextx"
)

// ========== 场景 1: HTTP 路由匹配 ==========

type RouteAction struct {
	Handler     string
	Middleware  []string
	RateLimit   int
	Timeout     time.Duration
	CacheEnable bool
}

func TestScenario_HTTPRouting(t *testing.T) {
	m := NewMatcher[*RouteAction]().EnableCache(5 * time.Minute)

	// 添加路由规则
	routes := []struct {
		path      string
		method    string
		action    *RouteAction
		priority  int
		matchFunc func(*contextx.Context) bool
	}{
		{
			path:     "/api/users/:id",
			method:   "GET",
			action:   &RouteAction{Handler: "GetUser", RateLimit: 100},
			priority: 10,
			matchFunc: func(ctx *contextx.Context) bool {
				path := contextx.Get[string](ctx, "path")
				// 匹配 /api/users/:id 格式
				return strings.HasPrefix(path, "/api/users/") && len(path) > len("/api/users/")
			},
		},
		{
			path:     "/api/users",
			method:   "POST",
			action:   &RouteAction{Handler: "CreateUser", RateLimit: 50},
			priority: 10,
			matchFunc: func(ctx *contextx.Context) bool {
				return contextx.Get[string](ctx, "path") == "/api/users"
			},
		},
		{
			path:     "/api/users",
			method:   "GET",
			action:   &RouteAction{Handler: "ListUsers", RateLimit: 200},
			priority: 9, // 稍低优先级，避免与 /api/users/:id 冲突
			matchFunc: func(ctx *contextx.Context) bool {
				return contextx.Get[string](ctx, "path") == "/api/users"
			},
		},
		{
			path:     "/api/orders/*",
			method:   "GET",
			action:   &RouteAction{Handler: "GetOrder", RateLimit: 150},
			priority: 5,
			matchFunc: func(ctx *contextx.Context) bool {
				return strings.HasPrefix(contextx.Get[string](ctx, "path"), "/api/orders/")
			},
		},
		{
			path:     "/api/products/:id",
			method:   "GET",
			action:   &RouteAction{Handler: "GetProduct", RateLimit: 300, CacheEnable: true},
			priority: 10,
			matchFunc: func(ctx *contextx.Context) bool {
				path := contextx.Get[string](ctx, "path")
				return strings.HasPrefix(path, "/api/products/") && len(path) > len("/api/products/")
			},
		},
		{
			path:     "/admin/*",
			method:   "ANY",
			action:   &RouteAction{Handler: "AdminPanel", Middleware: []string{"auth", "admin"}},
			priority: 100,
			matchFunc: func(ctx *contextx.Context) bool {
				return strings.HasPrefix(contextx.Get[string](ctx, "path"), "/admin/")
			},
		},
		{
			path:     "/public/*",
			method:   "ANY",
			action:   &RouteAction{Handler: "PublicResource"},
			priority: 1,
			matchFunc: func(ctx *contextx.Context) bool {
				return strings.HasPrefix(contextx.Get[string](ctx, "path"), "/public/")
			},
		},
	}

	for _, route := range routes {
		method := route.method
		matchFunc := route.matchFunc
		m.AddRule(
			NewChainRule(route.action).
				When(func(ctx *contextx.Context) bool {
					// 先检查 method
					if method != "ANY" && contextx.Get[string](ctx, "method") != method {
						return false
					}
					// 再检查路径
					return matchFunc(ctx)
				}).
				WithPriority(route.priority).
				WithID(fmt.Sprintf("%s:%s", route.method, route.path)),
		)
	}

	// 测试用例
	testCases := []struct {
		path            string
		method          string
		expectedHandler string
		shouldMatch     bool
	}{
		{"/api/users/123", "GET", "GetUser", true},
		{"/api/users", "POST", "CreateUser", true},
		{"/api/users", "GET", "ListUsers", true},
		{"/api/orders/456", "GET", "GetOrder", true},
		{"/api/products/789", "GET", "GetProduct", true},
		{"/admin/dashboard", "GET", "AdminPanel", true},
		{"/public/assets/logo.png", "GET", "PublicResource", true},
		{"/unknown/path", "GET", "", false},
	}

	for _, tc := range testCases {
		ctx := contextx.NewContext().WithValue("path", tc.path)
		ctx = ctx.WithValue("method", tc.method)

		result, matched := m.Match(ctx)

		if matched != tc.shouldMatch {
			t.Errorf("Path %s %s: expected match=%v, got=%v", tc.method, tc.path, tc.shouldMatch, matched)
			continue
		}

		if matched && result.Handler != tc.expectedHandler {
			t.Errorf("Path %s %s: expected handler=%s, got=%s", tc.method, tc.path, tc.expectedHandler, result.Handler)
		}
	}

	// 重复请求以测试缓存
	for i := 0; i < 5; i++ {
		ctx := contextx.NewContext().
			WithValue("path", "/api/users/123")
		ctx = ctx.WithValue("method", "GET")
		m.Match(ctx)
	}

	// 验证缓存效果
	stats := m.Stats()
	t.Logf("HTTP Routing Stats: %+v", stats)
	if stats["cache_hits"] < 1 {
		t.Error("Expected cache hits > 0")
	}
}

// ========== 场景 2: 限流规则匹配 ==========

type RateLimitRule struct {
	Strategy string // token-bucket, leaky-bucket, sliding-window
	Rate     int    // 每秒请求数
	Burst    int    // 突发容量
	Scope    string // global, per-ip, per-user
}

func TestScenario_RateLimiting(t *testing.T) {
	m := NewMatcher[*RateLimitRule]()

	// 定义限流规则（优先级从高到低）
	rules := []struct {
		name     string
		rule     *RateLimitRule
		matcher  func(*contextx.Context) bool
		priority int
	}{
		{
			name:     "Admin No Limit",
			rule:     &RateLimitRule{Strategy: "none", Rate: -1},
			matcher:  MatchBool("is_admin", true),
			priority: 100,
		},
		{
			name:     "VIP User High Limit",
			rule:     &RateLimitRule{Strategy: "token-bucket", Rate: 1000, Burst: 2000, Scope: "per-user"},
			matcher:  MatchString("user_level", "vip"),
			priority: 90,
		},
		{
			name:     "API Route Strict Limit",
			rule:     &RateLimitRule{Strategy: "sliding-window", Rate: 10, Burst: 20, Scope: "per-ip"},
			matcher:  MatchPrefix("path", "/api/sensitive"),
			priority: 80,
		},
		{
			name:     "Regular User Normal Limit",
			rule:     &RateLimitRule{Strategy: "token-bucket", Rate: 100, Burst: 200, Scope: "per-user"},
			matcher:  MatchBool("is_authenticated", true),
			priority: 50,
		},
		{
			name:     "Anonymous User Low Limit",
			rule:     &RateLimitRule{Strategy: "fixed-window", Rate: 10, Burst: 10, Scope: "per-ip"},
			matcher:  MatchBool("is_authenticated", false),
			priority: 10,
		},
		{
			name:     "Global Default",
			rule:     &RateLimitRule{Strategy: "leaky-bucket", Rate: 50, Burst: 100, Scope: "global"},
			matcher:  func(ctx *contextx.Context) bool { return true },
			priority: 1,
		},
	}

	for _, r := range rules {
		m.AddRule(
			NewChainRule(r.rule).
				When(r.matcher).
				WithPriority(r.priority).
				WithID(r.name),
		)
	}

	// 测试用例
	testCases := []struct {
		name         string
		ctx          map[string]interface{}
		expectedRule string
		expectedRate int
	}{
		{
			name: "Admin User",
			ctx: map[string]interface{}{
				"is_admin": true,
				"user_id":  "admin_001",
			},
			expectedRule: "none",
			expectedRate: -1,
		},
		{
			name: "VIP User",
			ctx: map[string]interface{}{
				"is_admin":         false,
				"user_level":       "vip",
				"is_authenticated": true,
			},
			expectedRule: "token-bucket",
			expectedRate: 1000,
		},
		{
			name: "Sensitive API",
			ctx: map[string]interface{}{
				"path":             "/api/sensitive/data",
				"is_authenticated": true,
			},
			expectedRule: "sliding-window",
			expectedRate: 10,
		},
		{
			name: "Regular User",
			ctx: map[string]interface{}{
				"is_authenticated": true,
				"user_level":       "regular",
			},
			expectedRule: "token-bucket",
			expectedRate: 100,
		},
		{
			name: "Anonymous User",
			ctx: map[string]interface{}{
				"is_authenticated": false,
				"ip":               "192.168.1.1",
			},
			expectedRule: "fixed-window",
			expectedRate: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := contextx.NewContext()
			for k, v := range tc.ctx {
				ctx.WithValue(k, v)
			}

			result, matched := m.Match(ctx)
			if !matched {
				t.Fatalf("Expected to match a rule")
			}

			if result.Strategy != tc.expectedRule {
				t.Errorf("Expected strategy=%s, got=%s", tc.expectedRule, result.Strategy)
			}

			if result.Rate != tc.expectedRate {
				t.Errorf("Expected rate=%d, got=%d", tc.expectedRate, result.Rate)
			}
		})
	}
}

// ========== 场景 3: IP 黑白名单 ==========

type IPAction struct {
	Action   string // allow, deny, throttle, captcha
	Reason   string
	LogLevel string
}

func TestScenario_IPBlacklist(t *testing.T) {
	m := NewMatcher[*IPAction]()

	// 黑名单 IP 段
	blacklistRanges := []string{
		"10.0.0.0/8",
		"192.168.1.100",
		"172.16.0.0/12",
	}

	// 白名单 IP
	whitelistIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"192.168.1.1",
	}

	// 可疑 IP（需要验证码）
	suspiciousIPs := []string{
		"123.45.67.89",
		"98.76.54.32",
	}

	// 添加白名单规则（最高优先级）
	for _, ip := range whitelistIPs {
		m.AddRule(
			NewChainRule(&IPAction{Action: "allow", Reason: "whitelist"}).
				When(MatchString("ip", ip)).
				WithPriority(100),
		)
	}

	// 添加黑名单规则
	for _, ipRange := range blacklistRanges {
		ipPrefix := ipRange
		m.AddRule(
			NewChainRule(&IPAction{Action: "deny", Reason: "blacklist", LogLevel: "warn"}).
				When(func(ctx *contextx.Context) bool {
					ip := contextx.Get[string](ctx, "ip")
					// 简化判断：检查 IP 前缀
					if strings.Contains(ipPrefix, "/") {
						// CIDR 格式，取前缀
						prefix := strings.Split(ipPrefix, "/")[0]
						// 取前两段或三段
						parts := strings.Split(prefix, ".")
						if len(parts) >= 2 {
							checkPrefix := strings.Join(parts[:2], ".")
							return strings.HasPrefix(ip, checkPrefix)
						}
					} else {
						// 精确匹配
						return ip == ipPrefix
					}
					return false
				}).
				WithPriority(90),
		)
	}

	// 添加可疑 IP 规则
	for _, ip := range suspiciousIPs {
		m.AddRule(
			NewChainRule(&IPAction{Action: "captcha", Reason: "suspicious"}).
				When(MatchString("ip", ip)).
				WithPriority(50),
		)
	}

	// 默认允许
	m.AddRule(
		NewChainRule(&IPAction{Action: "allow", Reason: "default"}).
			When(func(ctx *contextx.Context) bool { return true }).
			WithPriority(1),
	)

	// 测试用例
	testCases := []struct {
		ip             string
		expectedAction string
		expectedReason string
	}{
		{"8.8.8.8", "allow", "whitelist"},
		{"192.168.1.1", "allow", "whitelist"},
		{"10.0.0.1", "deny", "blacklist"},
		{"192.168.1.100", "deny", "blacklist"},
		{"123.45.67.89", "captcha", "suspicious"},
		{"114.114.114.114", "allow", "default"},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			ctx := contextx.NewContext().WithValue("ip", tc.ip)
			result, matched := m.Match(ctx)

			if !matched {
				t.Fatal("Expected to match")
			}

			if result.Action != tc.expectedAction {
				t.Errorf("Expected action=%s, got=%s", tc.expectedAction, result.Action)
			}

			if result.Reason != tc.expectedReason {
				t.Errorf("Expected reason=%s, got=%s", tc.expectedReason, result.Reason)
			}
		})
	}
}

// ========== 场景 4: 用户权限控制 ==========

type Permission struct {
	CanRead   bool
	CanWrite  bool
	CanDelete bool
	CanAdmin  bool
	Resources []string
}

func TestScenario_PermissionControl(t *testing.T) {
	m := NewMatcher[*Permission]()

	// 超级管理员
	m.AddRule(
		NewChainRule(&Permission{
			CanRead: true, CanWrite: true, CanDelete: true, CanAdmin: true,
			Resources: []string{"*"},
		}).
			When(MatchString("role", "super_admin")).
			WithPriority(100),
	)

	// 管理员
	m.AddRule(
		NewChainRule(&Permission{
			CanRead: true, CanWrite: true, CanDelete: true, CanAdmin: false,
			Resources: []string{"users", "orders", "products"},
		}).
			When(MatchString("role", "admin")).
			WithPriority(90),
	)

	// 编辑者
	m.AddRule(
		NewChainRule(&Permission{
			CanRead: true, CanWrite: true, CanDelete: false, CanAdmin: false,
			Resources: []string{"articles", "comments"},
		}).
			When(MatchString("role", "editor")).
			WithPriority(50),
	)

	// 普通用户
	m.AddRule(
		NewChainRule(&Permission{
			CanRead: true, CanWrite: false, CanDelete: false, CanAdmin: false,
			Resources: []string{"public"},
		}).
			When(MatchString("role", "user")).
			WithPriority(10),
	)

	// 访客
	m.AddRule(
		NewChainRule(&Permission{
			CanRead: true, CanWrite: false, CanDelete: false, CanAdmin: false,
			Resources: []string{"public"},
		}).
			When(func(ctx *contextx.Context) bool { return true }).
			WithPriority(1),
	)

	testCases := []struct {
		role      string
		canWrite  bool
		canDelete bool
		canAdmin  bool
	}{
		{"super_admin", true, true, true},
		{"admin", true, true, false},
		{"editor", true, false, false},
		{"user", false, false, false},
		{"guest", false, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.role, func(t *testing.T) {
			ctx := contextx.NewContext().WithValue("role", tc.role)
			result, matched := m.Match(ctx)
			if !matched {
				t.Fatal("Expected to match")
			}

			if result.CanWrite != tc.canWrite {
				t.Errorf("Expected CanWrite=%v, got=%v", tc.canWrite, result.CanWrite)
			}

			if result.CanDelete != tc.canDelete {
				t.Errorf("Expected CanDelete=%v, got=%v", tc.canDelete, result.CanDelete)
			}

			if result.CanAdmin != tc.canAdmin {
				t.Errorf("Expected CanAdmin=%v, got=%v", tc.canAdmin, result.CanAdmin)
			}
		})
	}
}

// ========== 场景 5: 动态配置路由 ==========

type DynamicConfig struct {
	Upstream       string
	Timeout        time.Duration
	RetryCount     int
	CircuitBreaker bool
}

func TestScenario_DynamicRouting(t *testing.T) {
	m := NewMatcher[*DynamicConfig]()

	// A/B 测试路由
	m.AddRule(
		NewChainRule(&DynamicConfig{
			Upstream: "service-v2",
			Timeout:  3 * time.Second,
		}).
			When(func(ctx *contextx.Context) bool {
				// 10% 流量到 v2
				userId := contextx.Get[string](ctx, "user_id")
				if userId == "" {
					return false
				}
				hash := 0
				for _, c := range userId {
					hash += int(c)
				}
				return hash%10 == 0
			}).
			WithPriority(100),
	)

	// VIP 用户路由到高性能服务
	m.AddRule(
		NewChainRule(&DynamicConfig{
			Upstream:   "service-premium",
			Timeout:    5 * time.Second,
			RetryCount: 3,
		}).
			When(MatchString("user_level", "vip")).
			WithPriority(90),
	)

	// 默认路由
	m.AddRule(
		NewChainRule(&DynamicConfig{
			Upstream:       "service-v1",
			Timeout:        2 * time.Second,
			RetryCount:     2,
			CircuitBreaker: true,
		}).
			When(func(ctx *contextx.Context) bool { return true }).
			WithPriority(1),
	)

	// 模拟 1000 个用户请求
	userDistribution := make(map[string]int)

	for i := 0; i < 1000; i++ {
		ctx := contextx.NewContext().WithValue("user_id", fmt.Sprintf("user_%d", i))
		ctx = ctx.WithValue("user_level", func() string {
			if i%20 == 0 {
				return "vip"
			}
			return "regular"
		}())

		result, matched := m.Match(ctx)
		if !matched {
			t.Fatal("Expected to match")
		}

		userDistribution[result.Upstream]++
	}

	t.Logf("User distribution: %+v", userDistribution)

	// 验证分布
	if userDistribution["service-premium"] < 40 || userDistribution["service-premium"] > 60 {
		t.Errorf("Expected ~50 VIP users, got %d", userDistribution["service-premium"])
	}

	if userDistribution["service-v2"] < 80 || userDistribution["service-v2"] > 120 {
		t.Errorf("Expected ~100 A/B test users, got %d", userDistribution["service-v2"])
	}
}

// ========== 场景 6: 高并发订单路由 ==========

type OrderRoute struct {
	ShardID  int
	Database string
	CacheKey string
	Priority string
}

func TestScenario_OrderSharding(t *testing.T) {
	m := NewMatcher[*OrderRoute]().EnableCache(1 * time.Minute)

	// 按订单 ID 分片到不同数据库
	shardCount := 10
	for i := 0; i < shardCount; i++ {
		shardID := i
		m.AddRule(
			NewChainRule(&OrderRoute{
				ShardID:  shardID,
				Database: fmt.Sprintf("order_db_%d", shardID),
				CacheKey: fmt.Sprintf("order_cache_%d", shardID),
			}).
				When(func(ctx *contextx.Context) bool {
					orderID := contextx.Get[string](ctx, "order_id")
					if orderID == "" {
						return false
					}
					// 使用 FNV-1a 哈希计算分片，保证分布均匀
					h := fnv.New32a()
					h.Write([]byte(orderID))
					return int(h.Sum32()%uint32(shardCount)) == shardID
				}).
				WithPriority(10),
		)
	}

	// 并发测试
	concurrency := 100
	ordersPerGoroutine := 100
	var wg sync.WaitGroup
	var totalMatches atomic.Int64
	shardDistribution := make([]atomic.Int64, shardCount)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < ordersPerGoroutine; j++ {
				orderID := fmt.Sprintf("ORD_%d_%d_%d", goroutineID, j, time.Now().UnixNano())
				ctx := contextx.NewContext().WithValue("order_id", orderID)

				result, matched := m.Match(ctx)
				if matched {
					totalMatches.Add(1)
					shardDistribution[result.ShardID].Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证结果
	expectedTotal := int64(concurrency * ordersPerGoroutine)
	if totalMatches.Load() != expectedTotal {
		t.Errorf("Expected %d matches, got %d", expectedTotal, totalMatches.Load())
	}

	// 验证分片分布均匀性
	t.Log("Shard distribution:")
	for i := 0; i < shardCount; i++ {
		count := shardDistribution[i].Load()
		t.Logf("  Shard %d: %d orders", i, count)

		// 每个分片应该接收大约 10% 的订单
		expectedPerShard := expectedTotal / int64(shardCount)
		deviation := float64(count-expectedPerShard) / float64(expectedPerShard) * 100
		if deviation > 20 || deviation < -20 {
			t.Errorf("Shard %d distribution deviation: %.2f%% (count=%d, expected~%d)",
				i, deviation, count, expectedPerShard)
		}
	}

	stats := m.Stats()
	t.Logf("Order sharding stats: %+v", stats)
}

// ========== 场景 7: 多条件组合匹配 ==========

type PromoRule struct {
	DiscountPercent int
	MaxDiscount     float64
	PromoCode       string
	Description     string
}

func TestScenario_PromoEngine(t *testing.T) {
	m := NewMatcher[*PromoRule]()

	// 新用户首单优惠
	m.AddRule(
		NewChainRule(&PromoRule{
			DiscountPercent: 20,
			MaxDiscount:     50.0,
			PromoCode:       "NEW_USER_20",
			Description:     "新用户首单8折",
		}).
			When(MatchBool("is_new_user", true)).
			When(MatchString("order_count", "0")).
			WithPriority(100),
	)

	// VIP 会员优惠
	m.AddRule(
		NewChainRule(&PromoRule{
			DiscountPercent: 15,
			MaxDiscount:     100.0,
			PromoCode:       "VIP_15",
			Description:     "VIP会员85折",
		}).
			When(MatchString("user_level", "vip")).
			WithPriority(90),
	)

	// 满减优惠
	m.AddRule(
		NewChainRule(&PromoRule{
			DiscountPercent: 10,
			MaxDiscount:     30.0,
			PromoCode:       "FULL_100_10",
			Description:     "满100减10元",
		}).
			When(func(ctx *contextx.Context) bool {
				amount := contextx.Get[string](ctx, "order_amount")
				if amount == "" {
					return false
				}
				amt, _ := strconv.ParseFloat(amount, 64)
				return amt >= 100.0
			}).
			WithPriority(50),
	)

	// 节日优惠
	m.AddRule(
		NewChainRule(&PromoRule{
			DiscountPercent: 5,
			MaxDiscount:     20.0,
			PromoCode:       "HOLIDAY_5",
			Description:     "节日特惠95折",
		}).
			When(func(ctx *contextx.Context) bool {
				date := contextx.Get[string](ctx, "date")
				// 简化判断：检查是否包含 "holiday"
				return strings.Contains(date, "holiday")
			}).
			WithPriority(30),
	)

	// 默认无优惠
	m.AddRule(
		NewChainRule(&PromoRule{
			DiscountPercent: 0,
			MaxDiscount:     0,
			PromoCode:       "NONE",
			Description:     "无优惠",
		}).
			When(func(ctx *contextx.Context) bool { return true }).
			WithPriority(1),
	)

	testCases := []struct {
		name             string
		ctx              map[string]interface{}
		expectedDiscount int
		expectedPromo    string
	}{
		{
			name: "New User First Order",
			ctx: map[string]interface{}{
				"is_new_user": true,
				"order_count": "0",
			},
			expectedDiscount: 20,
			expectedPromo:    "NEW_USER_20",
		},
		{
			name: "VIP Member",
			ctx: map[string]interface{}{
				"user_level":  "vip",
				"is_new_user": false,
			},
			expectedDiscount: 15,
			expectedPromo:    "VIP_15",
		},
		{
			name: "Full Reduction",
			ctx: map[string]interface{}{
				"order_amount": "150.00",
				"user_level":   "regular",
			},
			expectedDiscount: 10,
			expectedPromo:    "FULL_100_10",
		},
		{
			name: "Holiday Promotion",
			ctx: map[string]interface{}{
				"date":         "2025-01-01-holiday",
				"order_amount": "50.00",
			},
			expectedDiscount: 5,
			expectedPromo:    "HOLIDAY_5",
		},
		{
			name: "No Promotion",
			ctx: map[string]interface{}{
				"order_amount": "50.00",
			},
			expectedDiscount: 0,
			expectedPromo:    "NONE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := contextx.NewContext()
			for k, v := range tc.ctx {
				ctx.WithValue(k, v)
			}

			result, matched := m.Match(ctx)
			if !matched {
				t.Fatal("Expected to match")
			}

			if result.DiscountPercent != tc.expectedDiscount {
				t.Errorf("Expected discount=%d%%, got=%d%%", tc.expectedDiscount, result.DiscountPercent)
			}

			if result.PromoCode != tc.expectedPromo {
				t.Errorf("Expected promo=%s, got=%s", tc.expectedPromo, result.PromoCode)
			}

			t.Logf("Matched: %s (discount=%d%%, max=%.2f)",
				result.Description, result.DiscountPercent, result.MaxDiscount)
		})
	}
}

// ========== 场景 8: 中间件链式处理 ==========

func TestScenario_MiddlewareChain(t *testing.T) {
	type Result struct {
		Value string
	}

	m := NewMatcher[*Result]()

	// 日志中间件
	var logEntries []string
	m.Use(func(ctx *contextx.Context, next func() (*Result, bool)) (*Result, bool) {
		start := time.Now()
		result, matched := next()
		duration := time.Since(start)
		logEntries = append(logEntries, fmt.Sprintf("Match took %v, matched=%v", duration, matched))
		return result, matched
	})

	// 鉴权中间件
	var authChecks atomic.Int64
	m.Use(func(ctx *contextx.Context, next func() (*Result, bool)) (*Result, bool) {
		if !ctx.GetBool("authenticated") {
			authChecks.Add(1)
			return nil, false
		}
		return next()
	})

	// 限流中间件
	var rateLimitHits atomic.Int64
	m.Use(func(ctx *contextx.Context, next func() (*Result, bool)) (*Result, bool) {
		if ctx.GetBool("rate_limited") {
			rateLimitHits.Add(1)
			return nil, false
		}
		return next()
	})

	m.AddRule(
		NewChainRule(&Result{Value: "success"}).
			When(func(ctx *contextx.Context) bool { return true }).
			WithPriority(10),
	)

	// 测试未认证
	ctx1 := contextx.NewContext().WithValue("authenticated", false)
	_, matched := m.Match(ctx1)
	if matched {
		t.Error("Should not match when not authenticated")
	}
	if authChecks.Load() != 1 {
		t.Errorf("Expected 1 auth check, got %d", authChecks.Load())
	}

	// 测试限流
	ctx2 := contextx.NewContext().WithValue("authenticated", true)
	ctx2 = ctx2.WithValue("rate_limited", true)
	_, matched = m.Match(ctx2)
	if matched {
		t.Error("Should not match when rate limited")
	}
	if rateLimitHits.Load() != 1 {
		t.Errorf("Expected 1 rate limit hit, got %d", rateLimitHits.Load())
	}

	// 测试成功
	ctx3 := contextx.NewContext().WithValue("authenticated", true)
	result, matched := m.Match(ctx3)
	if !matched {
		t.Error("Should match when authenticated")
	}
	if result.Value != "success" {
		t.Errorf("Expected value=success, got=%s", result.Value)
	}

	if len(logEntries) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logEntries))
	}
	t.Logf("Log entries: %v", logEntries)
}

// ========== 场景 9: 缓存失效和重建 ==========

func TestScenario_CacheInvalidation(t *testing.T) {
	type Config struct {
		Version int
		Value   string
	}

	m := NewMatcher[*Config]().EnableCache(100 * time.Millisecond)

	version := 1
	m.AddRule(
		NewChainRule(&Config{Version: version, Value: "v1"}).
			When(MatchString("key", "config")).
			WithPriority(10),
	)

	ctx := contextx.NewContext().WithValue("key", "config")

	// 第一次匹配 - 缓存未命中
	result1, _ := m.Match(ctx)
	if result1.Version != 1 {
		t.Errorf("Expected version=1, got=%d", result1.Version)
	}

	stats1 := m.Stats()
	if stats1["cache_misses"] != 1 {
		t.Error("Expected 1 cache miss")
	}

	// 第二次匹配 - 缓存命中
	result2, _ := m.Match(ctx)
	if result2.Version != 1 {
		t.Errorf("Expected version=1, got=%d", result2.Version)
	}

	stats2 := m.Stats()
	if stats2["cache_hits"] != 1 {
		t.Error("Expected 1 cache hit")
	}

	// 等待缓存过期
	time.Sleep(150 * time.Millisecond)

	// 更新规则
	m.ClearRules()
	version = 2
	m.AddRule(
		NewChainRule(&Config{Version: version, Value: "v2"}).
			When(MatchString("key", "config")).
			WithPriority(10),
	)

	// 缓存过期后重新匹配
	result3, _ := m.Match(ctx)
	if result3.Version != 2 {
		t.Errorf("Expected version=2, got=%d", result3.Version)
	}

	stats3 := m.Stats()
	t.Logf("Final stats: %+v", stats3)
}

// ========== 场景 10: 极限压力测试 ==========

func TestScenario_ExtremePressure(t *testing.T) {
	type Action struct {
		ID int
	}

	m := NewMatcher[*Action]().EnableCache(5 * time.Minute)

	// 添加大量规则（race 检测下降低规模避免超时）
	ruleCount := 1000
	if testing.Short() {
		ruleCount = 100
	}
	for i := 0; i < ruleCount; i++ {
		id := i
		m.AddRule(
			NewChainRule(&Action{ID: id}).
				When(MatchString("id", fmt.Sprintf("id_%d", i))).
				WithPriority(i),
		)
	}

	// 极限并发测试（race 检测下降低并发度避免超时）
	concurrency := 100
	iterationsPerGoroutine := 50
	if testing.Short() {
		concurrency = 20
		iterationsPerGoroutine = 10
	}
	var wg sync.WaitGroup
	var totalOps atomic.Int64
	var successOps atomic.Int64

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < iterationsPerGoroutine; j++ {
				// 随机选择一个 ID
				targetID := (goroutineID*iterationsPerGoroutine + j) % ruleCount
				ctx := contextx.NewContext().WithValue("id", fmt.Sprintf("id_%d", targetID))

				result, matched := m.Match(ctx)
				totalOps.Add(1)
				if matched && result.ID == targetID {
					successOps.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// 验证结果
	total := totalOps.Load()
	success := successOps.Load()

	t.Logf("Extreme pressure test:")
	t.Logf("  Rules: %d", ruleCount)
	t.Logf("  Concurrency: %d", concurrency)
	t.Logf("  Total ops: %d", total)
	t.Logf("  Success ops: %d", success)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Ops/sec: %.2f", float64(total)/duration.Seconds())

	stats := m.Stats()
	t.Logf("  Stats: %+v", stats)

	if success != total {
		t.Errorf("Expected all operations to succeed, got %d/%d", success, total)
	}

	// 性能要求：至少 1000 ops/sec（CI环境友好）
	opsPerSec := float64(total) / duration.Seconds()
	if opsPerSec < 1000 {
		t.Logf("警告：性能较低 %.2f ops/sec", opsPerSec)
	} else if opsPerSec < 5000 {
		t.Logf("性能可接受: %.2f ops/sec", opsPerSec)
	}
}
