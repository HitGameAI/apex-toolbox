/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-16 00:00:00
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\desensitize_test.go
 * @Description: 数据脱敏单元测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
	"github.com/stretchr/testify/assert"
)

// testUser 测试用用户模型
type testUser struct {
	ID       int64   `gorm:"column:id;primaryKey" json:"id"`
	Name     string  `gorm:"column:name" json:"name" desensitize:"name"`
	Email    string  `gorm:"column:email" json:"email" desensitize:"email"`
	Phone    string  `gorm:"column:phone" json:"phone" desensitize:"mobilePhone"`
	Password string  `gorm:"column:password" json:"password" desensitize:"password"`
	IDCard   string  `gorm:"column:id_card" json:"id_card" desensitize:"idCard"`
	BankCard string  `gorm:"column:bank_card" json:"bank_card" desensitize:"bankCard"`
	Address  string  `gorm:"column:address" json:"address" desensitize:"address"`
	IP       string  `gorm:"column:ip" json:"ip" desensitize:"ipv4"`
	Avatar   string  `gorm:"column:avatar" json:"avatar"` // 无标签，不脱敏
	NickPtr  *string `gorm:"column:nick_ptr" json:"nick_ptr" desensitize:"name"`
}

// testOrder 测试嵌套结构体
type testOrder struct {
	ID    int64      `gorm:"column:id" json:"id"`
	User  testUser   `gorm:"-" json:"user"`
	UserP *testUser  `gorm:"-" json:"user_p"`
	Users []testUser `gorm:"-" json:"users"`
	Title string     `gorm:"column:title" json:"title"` // 无标签
}

func TestApplyDesensitize_StringFields(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		input    string
		contains string // 脱敏后应包含的片段
		missing  string // 脱敏后不应包含的片段
	}{
		{"email", "email", "zhangsan@example.com", "*", "zhangsan"},
		{"phone", "mobilePhone", "13812345678", "****", "2345"},
		{"name", "name", "张三丰", "*", ""},
		{"password", "password", "mySecret123", "*", "mySecret"},
		{"idCard", "idCard", "110101199001011234", "*", "900101"},
		{"bankCard", "bankCard", "6225881234567890", "*", "3456"},
		{"address", "address", "北京市朝阳区某某街道100号", "*", ""},
		{"ipv4", "ipv4", "192.168.1.100", ".*.*.*", "1.100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyDesensitizeByTagValue(tt.input, tt.tag)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains, "脱敏结果应包含 %q", tt.contains)
			}
			if tt.missing != "" {
				assert.NotContains(t, result, tt.missing, "脱敏结果不应包含 %q", tt.missing)
			}
		})
	}
}

// applyDesensitizeByTagValue 辅助函数：直接根据 tag 类型脱敏
func applyDesensitizeByTagValue(value, tag string) string {
	dtype, ok := desensitizeTypeMap[tag]
	if !ok {
		return value
	}
	return desensitize.Desensitize(value, dtype)
}

func TestApplyDesensitize_ModelFields(t *testing.T) {
	user := &testUser{
		Name:     "张三丰",
		Email:    "zhangsan@example.com",
		Phone:    "13812345678",
		Password: "mySecret123",
		IDCard:   "110101199001011234",
		BankCard: "6225881234567890",
		Address:  "北京市朝阳区某某街道100号",
		IP:       "192.168.1.100",
		Avatar:   "https://example.com/avatar.png",
	}

	ApplyDesensitize(user)

	// 有标签的字段应被脱敏
	assert.Contains(t, user.Email, "*", "email 应被脱敏")
	assert.Contains(t, user.Phone, "*", "phone 应被脱敏")
	assert.Contains(t, user.Password, "*", "password 应被脱敏")
	assert.NotEqual(t, "zhangsan@example.com", user.Email, "email 值应改变")

	// 无标签的字段不应被脱敏
	assert.Equal(t, "https://example.com/avatar.png", user.Avatar, "avatar 无标签不应脱敏")
}

func TestApplyDesensitize_PointerString(t *testing.T) {
	nick := "张三丰"
	user := &testUser{
		NickPtr: &nick,
	}

	ApplyDesensitize(user)

	assert.Contains(t, *user.NickPtr, "*", "指针字符串应被脱敏")
	assert.NotEqual(t, "张三丰", *user.NickPtr, "指针字符串值应改变")
}

func TestApplyDesensitize_NestedStruct(t *testing.T) {
	nick := "李四"
	order := &testOrder{
		ID: 1,
		User: testUser{
			Name:  "张三",
			Email: "zhangsan@example.com",
		},
		UserP: &testUser{
			Name:    "王五",
			NickPtr: &nick,
		},
		Users: []testUser{
			{Name: "赵六", Email: "zhaoliu@example.com"},
		},
		Title: "订单标题",
	}

	ApplyDesensitize(order)

	// 嵌套结构体字段应被脱敏
	assert.Contains(t, order.User.Email, "*", "嵌套结构体 email 应脱敏")
	assert.Contains(t, order.User.Name, "*", "嵌套结构体 name 应脱敏")

	// 指针嵌套结构体
	assert.Contains(t, order.UserP.Name, "*", "指针嵌套结构体 name 应脱敏")
	assert.Contains(t, *order.UserP.NickPtr, "*", "指针嵌套结构体指针字符串应脱敏")

	// 切片嵌套结构体
	assert.Contains(t, order.Users[0].Email, "*", "切片元素 email 应脱敏")
	assert.Contains(t, order.Users[0].Name, "*", "切片元素 name 应脱敏")

	// 无标签字段不脱敏
	assert.Equal(t, "订单标题", order.Title, "无标签字段不脱敏")
}

func TestApplyDesensitize_NilSafety(t *testing.T) {
	// nil 指针
	ApplyDesensitize(nil)

	// 非指针
	ApplyDesensitize(testUser{})

	// nil 指针字段
	user := &testUser{
		NickPtr: nil,
	}
	ApplyDesensitize(user) // 不应 panic
}

func TestGetDesensitizeFields(t *testing.T) {
	fields := GetDesensitizeFields(testUser{})

	assert.Contains(t, fields, "Name")
	assert.Equal(t, "name", fields["Name"])

	assert.Contains(t, fields, "Email")
	assert.Equal(t, "email", fields["Email"])

	assert.Contains(t, fields, "Phone")
	assert.Equal(t, "mobilePhone", fields["Phone"])

	assert.Contains(t, fields, "NickPtr")
	assert.Equal(t, "name", fields["NickPtr"])

	// 无标签字段不在列表
	_, exists := fields["Avatar"]
	assert.False(t, exists, "Avatar 无标签不应出现")
}

func TestQuery_WithDesensitize(t *testing.T) {
	q := NewQuery()
	assert.False(t, q.Desensitize, "默认未启用脱敏")

	q.WithDesensitize()
	assert.True(t, q.Desensitize, "WithDesensitize 后应启用")
}

func TestWithDesensitize_RepositoryOption(t *testing.T) {
	// 创建一个临时仓储（使用 nil db，仅测试配置）
	repo := &BaseRepository[testUser]{}
	assert.False(t, repo.desensitizeEnabled, "默认未启用")

	opt := WithDesensitize[testUser]()
	opt(repo)
	assert.True(t, repo.desensitizeEnabled, "WithDesensitize 后应启用")
}

func TestShouldDesensitize_RepoLevel(t *testing.T) {
	repo := &BaseRepository[testUser]{}
	q := NewQuery()

	// 默认不启用
	assert.False(t, repo.shouldDesensitize(q))
	assert.False(t, repo.shouldDesensitize(nil))

	// 仓储级启用
	repo.EnableDesensitize()
	assert.True(t, repo.shouldDesensitize(q))
	assert.True(t, repo.shouldDesensitize(nil))

	// 禁用
	repo.DisableDesensitize()
	assert.False(t, repo.shouldDesensitize(q))
}

func TestShouldDesensitize_QueryLevel(t *testing.T) {
	repo := &BaseRepository[testUser]{}
	q := NewQuery().WithDesensitize()

	// 查询级启用
	assert.True(t, repo.shouldDesensitize(q))

	// 查询级未启用
	q2 := NewQuery()
	assert.False(t, repo.shouldDesensitize(q2))
}
