/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-03 16:36:21
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\query_filter_test.go
 * @Description: Query 泛型过滤方法测试 - 确保100%测试覆盖率
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/logger"
	"github.com/HitGameAI/apex-toolbox/pkg/sqlbuilder/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	gormlogger "gorm.io/gorm/logger"
)

type TestStatus int32

const (
	TestStatusPending TestStatus = 1
	TestStatusActive  TestStatus = 2
	TestStatusClosed  TestStatus = 3
)

type TestPriority int32

const (
	TestPriorityLow    TestPriority = 1
	TestPriorityMedium TestPriority = 2
	TestPriorityHigh   TestPriority = 3
)

// ==============================================================================
// Filter 构造函数
// ==============================================================================

func TestFilterConstructors(t *testing.T) {
	t.Run("NewEqFilter", func(t *testing.T) {
		f := NewEqFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_EQ, f.Operator)
		assert.Equal(t, "test", f.Value)
	})

	t.Run("NewNeqFilter", func(t *testing.T) {
		f := NewNeqFilter("status", "deleted")
		assert.Equal(t, "status", f.Field)
		assert.Equal(t, constants.OP_NEQ, f.Operator)
		assert.Equal(t, "deleted", f.Value)
	})

	t.Run("NewGtFilter", func(t *testing.T) {
		f := NewGtFilter("age", 18)
		assert.Equal(t, "age", f.Field)
		assert.Equal(t, constants.OP_GT, f.Operator)
		assert.Equal(t, 18, f.Value)
	})

	t.Run("NewGteFilter", func(t *testing.T) {
		f := NewGteFilter("score", 60)
		assert.Equal(t, "score", f.Field)
		assert.Equal(t, constants.OP_GTE, f.Operator)
		assert.Equal(t, 60, f.Value)
	})

	t.Run("NewLtFilter", func(t *testing.T) {
		f := NewLtFilter("age", 65)
		assert.Equal(t, "age", f.Field)
		assert.Equal(t, constants.OP_LT, f.Operator)
		assert.Equal(t, 65, f.Value)
	})

	t.Run("NewLteFilter", func(t *testing.T) {
		f := NewLteFilter("price", 100.0)
		assert.Equal(t, "price", f.Field)
		assert.Equal(t, constants.OP_LTE, f.Operator)
		assert.Equal(t, 100.0, f.Value)
	})

	t.Run("NewInFilter", func(t *testing.T) {
		f := NewInFilter("status", 1, 2, 3)
		assert.Equal(t, "status", f.Field)
		assert.Equal(t, constants.OP_IN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 3, len(values))
	})

	t.Run("NewInFilter_nil", func(t *testing.T) {
		f := NewInFilter("status", nil)
		assert.Equal(t, constants.OP_IN, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewInFilterSlice", func(t *testing.T) {
		f := NewInFilterSlice("status", []interface{}{1, 2})
		assert.Equal(t, constants.OP_IN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 2, len(values))
	})

	t.Run("NewInFilterSlice_nil", func(t *testing.T) {
		f := NewInFilterSlice("status", nil)
		assert.Equal(t, constants.OP_IN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 0, len(values))
	})

	t.Run("NewNotInFilter", func(t *testing.T) {
		f := NewNotInFilter("status", "deleted", "banned")
		assert.Equal(t, "status", f.Field)
		assert.Equal(t, constants.OP_NOT_IN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 2, len(values))
	})

	t.Run("NewNotInFilter_nil", func(t *testing.T) {
		f := NewNotInFilter("status", nil)
		assert.Equal(t, constants.OP_NOT_IN, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewNotInFilterSlice", func(t *testing.T) {
		f := NewNotInFilterSlice("status", []interface{}{"a", "b"})
		assert.Equal(t, constants.OP_NOT_IN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 2, len(values))
	})

	t.Run("NewNotInFilterSlice_nil", func(t *testing.T) {
		f := NewNotInFilterSlice("status", nil)
		values := f.Value.([]interface{})
		assert.Equal(t, 0, len(values))
	})

	t.Run("NewLikeFilter", func(t *testing.T) {
		f := NewLikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_LIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})

	t.Run("NewBetweenFilter", func(t *testing.T) {
		f := NewBetweenFilter("age", 18, 65)
		assert.Equal(t, "age", f.Field)
		assert.Equal(t, constants.OP_BETWEEN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 2, len(values))
		assert.Equal(t, 18, values[0])
		assert.Equal(t, 65, values[1])
	})

	t.Run("NewIsNullFilter", func(t *testing.T) {
		f := NewIsNullFilter("deleted_at")
		assert.Equal(t, "deleted_at", f.Field)
		assert.Equal(t, constants.OP_IS_NULL, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewIsNotNullFilter", func(t *testing.T) {
		f := NewIsNotNullFilter("email")
		assert.Equal(t, "email", f.Field)
		assert.Equal(t, constants.OP_IS_NOT_NULL, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewStartsWithFilter", func(t *testing.T) {
		f := NewStartsWithFilter("username", "admin")
		assert.Equal(t, "username", f.Field)
		assert.Equal(t, constants.OP_STARTS_WITH, f.Operator)
		assert.Equal(t, "admin", f.Value)
	})

	t.Run("NewEndsWithFilter", func(t *testing.T) {
		f := NewEndsWithFilter("email", "@example.com")
		assert.Equal(t, "email", f.Field)
		assert.Equal(t, constants.OP_ENDS_WITH, f.Operator)
		assert.Equal(t, "@example.com", f.Value)
	})

	t.Run("NewNotLikeFilter", func(t *testing.T) {
		f := NewNotLikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_NOT_LIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})

	t.Run("NewRegexpFilter", func(t *testing.T) {
		f := NewRegexpFilter("name", "^admin")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_REGEX, f.Operator)
		assert.Equal(t, "^admin", f.Value)
	})

	t.Run("NewFindInSetFilter", func(t *testing.T) {
		f := NewFindInSetFilter("tags", "important")
		assert.Equal(t, "tags", f.Field)
		assert.Equal(t, constants.OP_FIND_IN_SET, f.Operator)
		assert.Equal(t, "important", f.Value)
	})

	t.Run("NewJsonbLikeFilter", func(t *testing.T) {
		f := NewJsonbLikeFilter("translations", "hello")
		assert.Equal(t, "translations", f.Field)
		assert.Equal(t, constants.OP_JSONB_LIKE, f.Operator)
		assert.Equal(t, "%hello%", f.Value)
	})

	t.Run("NewFilter", func(t *testing.T) {
		f := NewFilter("field", constants.OP_EQ, "value")
		assert.Equal(t, "field", f.Field)
		assert.Equal(t, constants.OP_EQ, f.Operator)
		assert.Equal(t, "value", f.Value)
	})

	t.Run("protobuf wrappers", func(t *testing.T) {
		f := NewEqFilter("group_id", wrapperspb.String("g1"))
		assert.Equal(t, "g1", f.Value)

		stringValue := wrapperspb.String("g2")
		f = NewEqFilter("group_id", stringValue)
		assert.Equal(t, "g2", f.Value)

		f = NewEqFilter("enabled", wrapperspb.Bool(false))
		assert.Equal(t, false, f.Value)

		f = NewInFilter("group_id", wrapperspb.String("g1"), wrapperspb.String("g2"))
		values := f.Value.([]interface{})
		assert.Equal(t, []interface{}{"g1", "g2"}, values)

		f = NewBetweenFilter("age", wrapperspb.Int32(1), wrapperspb.Int32(2))
		values = f.Value.([]interface{})
		assert.Equal(t, int32(1), values[0])
		assert.Equal(t, int32(2), values[1])
	})

	t.Run("NewSubQuery", func(t *testing.T) {
		sq := NewSubQuery("SELECT id FROM users WHERE active = ?", true)
		assert.Equal(t, "SELECT id FROM users WHERE active = ?", sq.SQL)
		assert.Equal(t, []interface{}{true}, sq.Args)
	})
}

// ==============================================================================
// FilterGroup 方法
// ==============================================================================

func TestFilterGroup(t *testing.T) {
	t.Run("NewFilterGroup_AND", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		assert.Equal(t, constants.LOGIC_AND, fg.LogicOp)
		assert.True(t, fg.IsEmpty())
	})

	t.Run("NewFilterGroup_OR", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR)
		assert.Equal(t, constants.LOGIC_OR, fg.LogicOp)
	})

	t.Run("NewFilterGroup_默认AND", func(t *testing.T) {
		fg := NewFilterGroup(constants.Operator("UNKNOWN"))
		assert.Equal(t, constants.LOGIC_AND, fg.LogicOp)
	})

	t.Run("AddFilter", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddFilter(NewEqFilter("name", "test"))
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, len(fg.Filters))
		assert.False(t, fg.IsEmpty())
	})

	t.Run("AddFilter_nil", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilter(nil)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddFilters", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilters(NewEqFilter("a", 1), nil, NewEqFilter("b", 2))
		assert.Equal(t, 2, len(fg.Filters))
	})

	t.Run("AddGroup", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		sub := NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active"))
		result := main.AddGroup(sub)
		assert.Equal(t, main, result)
		assert.Equal(t, 1, len(main.Groups))
	})

	t.Run("AddGroup_nil", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGroup(nil)
		assert.Equal(t, 0, len(fg.Groups))
	})

	t.Run("IsEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		assert.True(t, fg.IsEmpty())
		fg.AddFilter(NewEqFilter("name", "test"))
		assert.False(t, fg.IsEmpty())
	})

	t.Run("Count", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		assert.Equal(t, 0, fg.Count())
		fg.AddFilter(NewEqFilter("a", 1))
		fg.AddFilter(NewEqFilter("b", 2))
		assert.Equal(t, 2, fg.Count())
		sub := NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("c", 3))
		fg.AddGroup(sub)
		assert.Equal(t, 3, fg.Count())
	})

	t.Run("Count_nil", func(t *testing.T) {
		var fg *FilterGroup
		assert.Equal(t, 0, fg.Count())
	})

	t.Run("AddFilterIf_true", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIf(true, NewEqFilter("name", "test"))
		assert.Equal(t, 1, len(fg.Filters))
	})

	t.Run("AddFilterIf_false", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIf(false, NewEqFilter("name", "test"))
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddFilterIf_nil", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIf(true, nil)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddFilterIfValueNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIfValueNotEmpty(NewEqFilter("name", "test"))
		assert.Equal(t, 1, len(fg.Filters))
	})

	t.Run("AddFilterIfValueNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIfValueNotEmpty(NewEqFilter("name", ""))
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddFilterIfValueNotEmpty_nil", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIfValueNotEmpty(nil)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIfNotEmpty("name", constants.OP_EQ, "test")
		assert.Equal(t, 1, len(fg.Filters))
	})

	t.Run("AddFilterIfNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFilterIfNotEmpty("name", constants.OP_EQ, "")
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddEqFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddEqFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_EQ, fg.Filters[0].Operator)
	})

	t.Run("AddNeqFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNeqFilterIfNotEmpty("status", "deleted")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_NEQ, fg.Filters[0].Operator)
	})

	t.Run("AddGtFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGtFilterIfNotEmpty("age", 18)
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_GT, fg.Filters[0].Operator)
	})

	t.Run("AddGteFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGteFilterIfNotEmpty("score", 60)
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_GTE, fg.Filters[0].Operator)
	})

	t.Run("AddLtFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLtFilterIfNotEmpty("age", 65)
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_LT, fg.Filters[0].Operator)
	})

	t.Run("AddLteFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLteFilterIfNotEmpty("price", 100)
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_LTE, fg.Filters[0].Operator)
	})

	t.Run("AddLikeFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_LIKE, fg.Filters[0].Operator)
	})

	t.Run("AddLikeFilterIfNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddInFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddInFilterIfNotEmpty("status", []string{"active", "pending"})
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_IN, fg.Filters[0].Operator)
	})

	t.Run("AddInFilterIfNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddInFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddNotInFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotInFilterIfNotEmpty("status", []string{"deleted", "banned"})
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_NOT_IN, fg.Filters[0].Operator)
	})

	t.Run("AddNotInFilterIfNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotInFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddBetweenFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddBetweenFilterIfNotEmpty("age", 18, 65)
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_BETWEEN, fg.Filters[0].Operator)
	})

	t.Run("AddBetweenFilterIfNotEmpty_minEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddBetweenFilterIfNotEmpty("age", nil, 65)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddBetweenFilterIfNotEmpty_maxEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddBetweenFilterIfNotEmpty("age", 18, nil)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddStartsWithFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddStartsWithFilterIfNotEmpty("name", "John")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_STARTS_WITH, fg.Filters[0].Operator)
	})

	t.Run("AddEndsWithFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddEndsWithFilterIfNotEmpty("email", "@example.com")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_ENDS_WITH, fg.Filters[0].Operator)
	})

	t.Run("AddRegexpFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddRegexpFilterIfNotEmpty("name", "^admin")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_REGEX, fg.Filters[0].Operator)
	})

	t.Run("AddRegexpFilterIfNotEmpty_empty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddRegexpFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddNotLikeFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotLikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_NOT_LIKE, fg.Filters[0].Operator)
	})

	t.Run("AddFindInSetFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFindInSetFilterIfNotEmpty("tags", "important")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_FIND_IN_SET, fg.Filters[0].Operator)
	})

	t.Run("AddGroupIf_true", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		sub := NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active"))
		main.AddGroupIf(true, sub)
		assert.Equal(t, 1, len(main.Groups))
	})

	t.Run("AddGroupIf_false", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		sub := NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active"))
		main.AddGroupIf(false, sub)
		assert.Equal(t, 0, len(main.Groups))
	})

	t.Run("AddGroupIf_empty", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		empty := NewFilterGroup(constants.LOGIC_OR)
		main.AddGroupIf(true, empty)
		assert.Equal(t, 0, len(main.Groups))
	})

	t.Run("AddGroupIfNotEmpty", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		sub := NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active"))
		main.AddGroupIfNotEmpty(sub)
		assert.Equal(t, 1, len(main.Groups))
	})

	t.Run("AddGroupIfNotEmpty_nil", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		main.AddGroupIfNotEmpty(nil)
		assert.Equal(t, 0, len(main.Groups))
	})

	t.Run("AddGroupIfNotEmpty_empty", func(t *testing.T) {
		main := NewFilterGroup(constants.LOGIC_AND)
		empty := NewFilterGroup(constants.LOGIC_OR)
		main.AddGroupIfNotEmpty(empty)
		assert.Equal(t, 0, len(main.Groups))
	})

	t.Run("Clear", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active")))
		assert.False(t, fg.IsEmpty())
		result := fg.Clear()
		assert.Equal(t, fg, result)
		assert.True(t, fg.IsEmpty())
	})

	t.Run("Clone", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(NewFilterGroup(constants.LOGIC_OR).AddFilter(NewEqFilter("status", "active")))
		cloned := fg.Clone()
		assert.Equal(t, fg.LogicOp, cloned.LogicOp)
		assert.Equal(t, len(fg.Filters), len(cloned.Filters))
		assert.Equal(t, len(fg.Groups), len(cloned.Groups))
		cloned.AddFilter(NewEqFilter("extra", "value"))
		assert.NotEqual(t, len(fg.Filters), len(cloned.Filters))
	})
}

// ==============================================================================
// Query.AddFilter / AddFilters
// ==============================================================================

func TestQueryAndFilters(t *testing.T) {
	t.Run("AddFilter", func(t *testing.T) {
		q := NewQuery()
		result := q.AddFilter(NewEqFilter("name", "test"))
		assert.Equal(t, q, result)
		assert.Equal(t, 1, len(q.Filters))
	})

	t.Run("AddFilter_nil", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddFilters", func(t *testing.T) {
		q := NewQuery()
		q.AddFilters(NewEqFilter("a", 1), nil, NewEqFilter("b", 2))
		assert.Equal(t, 2, len(q.Filters))
	})
}

// ==============================================================================
// Query.AddFilterIfNotEmpty
// ==============================================================================

func TestAddFilterIfNotEmpty(t *testing.T) {
	t.Run("字符串", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "name", q.Filters[0].Field)
		assert.Equal(t, constants.OP_EQ, q.Filters[0].Operator)
		assert.Equal(t, "test", q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery()
		q.AddFilterIfNotEmpty("name", nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("字符串指针", func(t *testing.T) {
		str := "test"
		q := NewQuery()
		q.AddFilterIfNotEmpty("name", &str)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "test", q.Filters[0].Value)

		emptyStr := ""
		q = NewQuery()
		q.AddFilterIfNotEmpty("name", &emptyStr)
		assert.Equal(t, 0, len(q.Filters))

		var nilStr *string
		q = NewQuery()
		q.AddFilterIfNotEmpty("name", nilStr)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("字符串切片", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("status", []string{"active", "pending"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))

		q = NewQuery()
		q.AddFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 0, len(q.Filters))

		var nilSlice []string
		q = NewQuery()
		q.AddFilterIfNotEmpty("status", nilSlice)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("int切片", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("age", []int{20, 30, 40})
		assert.Equal(t, 1, len(q.Filters))
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 3, len(values))

		q = NewQuery()
		q.AddFilterIfNotEmpty("age", []int{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("int32切片", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("count", []int32{100, 200, 300})
		assert.Equal(t, 1, len(q.Filters))
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, int32(100), values[0])

		q = NewQuery()
		q.AddFilterIfNotEmpty("count", []int32{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("int64切片", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("id", []int64{1000, 2000, 3000})
		assert.Equal(t, 1, len(q.Filters))
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, int64(1000), values[0])

		q = NewQuery()
		q.AddFilterIfNotEmpty("id", []int64{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("单个整数", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("age", 25)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, 25, q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("count", int32(100))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, int32(100), q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("id", int64(1000))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, int64(1000), q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("count", uint(100))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, uint(100), q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("count", uint32(100))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, uint32(100), q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("id", uint64(1000))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, uint64(1000), q.Filters[0].Value)
	})

	t.Run("布尔值", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("is_active", true)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, true, q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("is_deleted", false)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, false, q.Filters[0].Value)
	})

	t.Run("布尔指针", func(t *testing.T) {
		trueVal := true
		q := NewQuery()
		q.AddFilterIfNotEmpty("is_active", &trueVal)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, true, q.Filters[0].Value)

		falseVal := false
		q = NewQuery()
		q.AddFilterIfNotEmpty("is_active", &falseVal)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, false, q.Filters[0].Value)

		var nilBool *bool
		q = NewQuery()
		q.AddFilterIfNotEmpty("is_active", nilBool)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("枚举切片", func(t *testing.T) {
		q := NewQuery()
		statuses := []TestStatus{TestStatusPending, TestStatusActive}
		q.AddFilterIfNotEmpty("status", statuses)
		assert.Equal(t, 1, len(q.Filters))
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))

		q = NewQuery()
		q.AddFilterIfNotEmpty("status", []TestStatus{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("单个枚举值", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("status", TestStatusActive)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, TestStatusActive, q.Filters[0].Value)
	})

	t.Run("枚举指针", func(t *testing.T) {
		status := TestStatusActive
		q := NewQuery()
		q.AddFilterIfNotEmpty("status", &status)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, TestStatusActive, q.Filters[0].Value)

		var nilStatus *TestStatus
		q = NewQuery()
		q.AddFilterIfNotEmpty("status", nilStatus)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("整数指针", func(t *testing.T) {
		age := 25
		q := NewQuery()
		q.AddFilterIfNotEmpty("age", &age)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, 25, q.Filters[0].Value)

		zero := 0
		q = NewQuery()
		q.AddFilterIfNotEmpty("count", &zero)
		assert.Equal(t, 0, len(q.Filters))

		var nilInt *int
		q = NewQuery()
		q.AddFilterIfNotEmpty("age", nilInt)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("int64指针", func(t *testing.T) {
		id := int64(12345)
		q := NewQuery()
		q.AddFilterIfNotEmpty("id", &id)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, int64(12345), q.Filters[0].Value)

		var nilInt64 *int64
		q = NewQuery()
		q.AddFilterIfNotEmpty("id", nilInt64)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.
			AddFilterIfNotEmpty("name", "test").
			AddFilterIfNotEmpty("age", 25).
			AddFilterIfNotEmpty("status", []string{"active", "pending"}).
			AddFilterIfNotEmpty("empty", "").
			AddFilterIfNotEmpty("nil", nil).
			AddFilterIfNotEmpty("empty_slice", []int{})

		assert.Equal(t, q, result)
		assert.Equal(t, 3, len(q.Filters))
	})
}

// ==============================================================================
// Query.AddLikeFilterIfNotEmpty
// ==============================================================================

func TestAddLikeFilterIfNotEmpty(t *testing.T) {
	t.Run("非空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddLikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LIKE, q.Filters[0].Operator)
		assert.Equal(t, "%test%", q.Filters[0].Value)
	})

	t.Run("空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddLikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddLikeFilterIfNotEmpty("title", "golang")
		assert.Equal(t, q, result)
		assert.Equal(t, "%golang%", q.Filters[0].Value)
	})
}

// ==============================================================================
// Query.AddJsonbLikeFilterIfNotEmpty
// ==============================================================================

func TestAddJsonbLikeFilterIfNotEmpty(t *testing.T) {
	t.Run("非空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonbLikeFilterIfNotEmpty("translations", "hello")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_JSONB_LIKE, q.Filters[0].Operator)
		assert.Equal(t, "%hello%", q.Filters[0].Value)
	})
	t.Run("空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonbLikeFilterIfNotEmpty("translations", "")
		assert.Equal(t, 0, len(q.Filters))
	})
	t.Run("nil关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonbLikeFilterIfNotEmpty("translations", nil)
		assert.Equal(t, 0, len(q.Filters))
	})
	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddJsonbLikeFilterIfNotEmpty("data", "keyword")
		assert.Equal(t, q, result)
		assert.Equal(t, "%keyword%", q.Filters[0].Value)
	})
	t.Run("protobuf StringValue", func(t *testing.T) {
		q := NewQuery()
		q.AddJsonbLikeFilterIfNotEmpty("translations", wrapperspb.String("test"))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "%test%", q.Filters[0].Value)
	})
}

// ==============================================================================
// Query.AddTimeRangeFilter
// ==============================================================================

func TestAddTimeRangeFilter(t *testing.T) {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	t.Run("完整时间范围", func(t *testing.T) {
		q := NewQuery()
		q.AddTimeRangeFilter("created_at", startTime, endTime)
		assert.Equal(t, 2, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
		assert.Equal(t, constants.OP_LTE, q.Filters[1].Operator)
	})

	t.Run("只有开始时间", func(t *testing.T) {
		q := NewQuery()
		q.AddTimeRangeFilter("created_at", startTime, nil)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
	})

	t.Run("只有结束时间", func(t *testing.T) {
		q := NewQuery()
		q.AddTimeRangeFilter("created_at", nil, endTime)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LTE, q.Filters[0].Operator)
	})

	t.Run("都为nil", func(t *testing.T) {
		q := NewQuery()
		q.AddTimeRangeFilter("created_at", nil, nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddTimeRangeFilter("updated_at", startTime, endTime)
		assert.Equal(t, q, result)
	})

	t.Run("时间指针", func(t *testing.T) {
		q := NewQuery()
		q.AddTimeRangeFilter("created_at", &startTime, &endTime)
		assert.Equal(t, 2, len(q.Filters))
		assert.Equal(t, &startTime, q.Filters[0].Value)
		assert.Equal(t, &endTime, q.Filters[1].Value)
	})

	t.Run("零值时间", func(t *testing.T) {
		q := NewQuery()
		zeroTime := time.Time{}
		q.AddTimeRangeFilter("created_at", zeroTime, zeroTime)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("零值时间指针", func(t *testing.T) {
		q := NewQuery()
		zeroTimePtr := &time.Time{}
		q.AddTimeRangeFilter("created_at", zeroTimePtr, zeroTimePtr)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("有效时间指针", func(t *testing.T) {
		q := NewQuery()
		now := time.Now()
		q.AddTimeRangeFilter("created_at", &now, &now)
		assert.Equal(t, 2, len(q.Filters))
	})

	t.Run("混合-只有开始时间指针", func(t *testing.T) {
		q := NewQuery()
		now := time.Now()
		q.AddTimeRangeFilter("created_at", &now, nil)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
	})

	t.Run("混合-只有结束时间指针", func(t *testing.T) {
		q := NewQuery()
		now := time.Now()
		q.AddTimeRangeFilter("created_at", nil, &now)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LTE, q.Filters[0].Operator)
	})

	t.Run("验证操作符和字段", func(t *testing.T) {
		q := NewQuery()
		now := time.Now()
		q.AddTimeRangeFilter("created_at", &now, &now)
		assert.Equal(t, 2, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
		assert.Equal(t, constants.OP_LTE, q.Filters[1].Operator)
		assert.Equal(t, "created_at", q.Filters[0].Field)
		assert.Equal(t, "created_at", q.Filters[1].Field)
	})
}

// ==============================================================================
// Query.AddRawFilter
// ==============================================================================

func TestAddRawFilter(t *testing.T) {
	t.Run("基本原始SQL", func(t *testing.T) {
		q := NewQuery()
		q.AddRawFilter("age > 18 AND status = 'active'")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "age > 18 AND status = 'active'", q.Filters[0].Field)
		assert.Equal(t, constants.OP_RAW, q.Filters[0].Operator)
		assert.Nil(t, q.Filters[0].Value)
	})

	t.Run("IS NOT NULL条件", func(t *testing.T) {
		q := NewQuery()
		q.AddRawFilter("to_agent_id IS NOT NULL AND to_agent_id != ''")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_RAW, q.Filters[0].Operator)
	})

	t.Run("子查询条件", func(t *testing.T) {
		q := NewQuery()
		q.AddRawFilter("user_id IN (SELECT id FROM premium_users)")
		assert.Equal(t, 1, len(q.Filters))
	})

	t.Run("空字符串", func(t *testing.T) {
		q := NewQuery()
		q.AddRawFilter("")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("组合使用", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewEqFilter("status", "active")).
			AddRawFilter("age > 18").
			AddFilter(NewLikeFilter("name", "张三"))
		assert.Equal(t, 3, len(q.Filters))
		assert.Equal(t, constants.OP_EQ, q.Filters[0].Operator)
		assert.Equal(t, constants.OP_RAW, q.Filters[1].Operator)
		assert.Equal(t, constants.OP_LIKE, q.Filters[2].Operator)
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddRawFilter("created_at > NOW() - INTERVAL 7 DAY")
		assert.Equal(t, q, result)
	})
}

// ==============================================================================
// Query IfNotEmpty 系列方法
// ==============================================================================

func TestIfNotEmptyMethods(t *testing.T) {
	t.Run("AddInFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery()
		q.AddInFilterIfNotEmpty("status", []string{"active", "pending"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IN, q.Filters[0].Operator)

		q = NewQuery()
		q.AddInFilterIfNotEmpty("id", []int{1, 2, 3})
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 3, len(values))

		q = NewQuery()
		priorities := []TestPriority{TestPriorityLow, TestPriorityHigh}
		q.AddInFilterIfNotEmpty("priority", priorities)
		assert.Equal(t, 1, len(q.Filters))

		q = NewQuery()
		q.AddInFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery()
		q.AddInFilterIfNotEmpty("status", nil)
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery()
		q.AddInFilterIfNotEmpty("status", "active")
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery()
		result := q.AddInFilterIfNotEmpty("tags", []string{"go", "rust"})
		assert.Equal(t, q, result)
	})

	t.Run("AddNeqFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddNeqFilterIfNotEmpty("status", "deleted")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NEQ, q.Filters[0].Operator)

		q = NewQuery().AddNeqFilterIfNotEmpty("status", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddGtFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddGtFilterIfNotEmpty("age", 18)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GT, q.Filters[0].Operator)

		q = NewQuery().AddGtFilterIfNotEmpty("age", 0)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddGteFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddGteFilterIfNotEmpty("score", 60)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)

		q = NewQuery().AddGteFilterIfNotEmpty("score", nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddLtFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddLtFilterIfNotEmpty("age", 65)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LT, q.Filters[0].Operator)

		q = NewQuery().AddLtFilterIfNotEmpty("age", nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddLteFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddLteFilterIfNotEmpty("price", 100.0)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LTE, q.Filters[0].Operator)

		q = NewQuery().AddLteFilterIfNotEmpty("price", nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddNotInFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddNotInFilterIfNotEmpty("status", []string{"deleted", "banned"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NOT_IN, q.Filters[0].Operator)

		q = NewQuery().AddNotInFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 0, len(q.Filters))

		var nilSlice []string
		q = NewQuery().AddNotInFilterIfNotEmpty("status", nilSlice)
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery().AddNotInFilterIfNotEmpty("status", nil)
		assert.Equal(t, 0, len(q.Filters))

		values := [3]string{"a", "b", "c"}
		q = NewQuery().AddNotInFilterIfNotEmpty("status", values)
		assert.Equal(t, 1, len(q.Filters))

		var emptyArray [0]string
		q = NewQuery().AddNotInFilterIfNotEmpty("status", emptyArray)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddBetweenFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddBetweenFilterIfNotEmpty("age", 18, 65)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)

		q = NewQuery().AddBetweenFilterIfNotEmpty("age", nil, 65)
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery().AddBetweenFilterIfNotEmpty("age", 18, nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddStartsWithFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddStartsWithFilterIfNotEmpty("name", "John")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_STARTS_WITH, q.Filters[0].Operator)

		q = NewQuery().AddStartsWithFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddEndsWithFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddEndsWithFilterIfNotEmpty("email", "@example.com")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_ENDS_WITH, q.Filters[0].Operator)

		q = NewQuery().AddEndsWithFilterIfNotEmpty("email", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddNotLikeFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddNotLikeFilterIfNotEmpty("name", "%test%")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NOT_LIKE, q.Filters[0].Operator)

		q = NewQuery().AddNotLikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddFindInSetFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery().AddFindInSetFilterIfNotEmpty("tags", "important")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_FIND_IN_SET, q.Filters[0].Operator)

		q = NewQuery().AddFindInSetFilterIfNotEmpty("tags", "")
		assert.Equal(t, 0, len(q.Filters))
	})
}

// ==============================================================================
// Query.AddCursorFilter
// ==============================================================================

func TestAddCursorFilter(t *testing.T) {
	t.Run("空字符串游标", func(t *testing.T) {
		q := NewQuery()
		result := q.AddCursorFilter("message_id", "", false)
		assert.Equal(t, q, result)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("nil游标", func(t *testing.T) {
		q := NewQuery()
		result := q.AddCursorFilter("message_id", nil, false)
		assert.Equal(t, q, result)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("零值int游标", func(t *testing.T) {
		q := NewQuery()
		result := q.AddCursorFilter("id", 0, false)
		assert.Equal(t, q, result)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("向后翻页", func(t *testing.T) {
		q := NewQuery()
		result := q.AddCursorFilter("message_id", "msg_100", false)
		assert.Equal(t, q, result)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GT, q.Filters[0].Operator)
		assert.Equal(t, "msg_100", q.Filters[0].Value)
	})

	t.Run("向前翻页", func(t *testing.T) {
		q := NewQuery()
		result := q.AddCursorFilter("message_id", "msg_100", true)
		_ = result
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LT, q.Filters[0].Operator)
	})

	t.Run("int64游标", func(t *testing.T) {
		q := NewQuery()
		q.AddCursorFilter("id", int64(9999), false)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GT, q.Filters[0].Operator)
		assert.Equal(t, int64(9999), q.Filters[0].Value)
	})

	t.Run("与其他过滤条件链式组合", func(t *testing.T) {
		tests := []struct {
			name        string
			cursor      any
			isPrev      bool
			wantFilters int
			wantOp      constants.Operator
		}{
			{"向后翻页-字符串游标", "msg_50", false, 2, constants.OP_GT},
			{"向前翻页-字符串游标", "msg_50", true, 2, constants.OP_LT},
			{"空游标-不添加", "", false, 1, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := NewQuery().AddEqual("session_id", "sess_abc")
				q.AddCursorFilter("message_id", tt.cursor, tt.isPrev)
				assert.Equal(t, tt.wantFilters, len(q.Filters))
				if tt.wantFilters == 2 {
					assert.Equal(t, tt.wantOp, q.Filters[1].Operator)
					assert.Equal(t, tt.cursor, q.Filters[1].Value)
				}
			})
		}
	})
}

// ==============================================================================
// Query.AddEqOrInFilter
// ==============================================================================

func TestAddEqOrInFilter(t *testing.T) {
	t.Run("nil值", func(t *testing.T) {
		q := NewQuery()
		result := q.AddEqOrInFilter("session_id", nil)
		assert.Equal(t, q, result)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("空切片", func(t *testing.T) {
		q := NewQuery()
		q.AddEqOrInFilter("session_id", []string{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("单值使用EQ", func(t *testing.T) {
		q := NewQuery()
		q.AddEqOrInFilter("session_id", []string{"session_abc"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_EQ, q.Filters[0].Operator)
		assert.Equal(t, "session_abc", q.Filters[0].Value)
	})

	t.Run("多值使用IN", func(t *testing.T) {
		q := NewQuery()
		q.AddEqOrInFilter("session_id", []string{"s1", "s2", "s3"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]any)
		assert.Equal(t, 3, len(values))
	})

	t.Run("int64切片", func(t *testing.T) {
		q := NewQuery()
		q.AddEqOrInFilter("id", []int64{100, 200})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]any)
		assert.Equal(t, int64(100), values[0])
	})

	t.Run("非切片值", func(t *testing.T) {
		q := NewQuery()
		q.AddEqOrInFilter("status", "active")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式场景", func(t *testing.T) {
		tests := []struct {
			name        string
			values      any
			wantFilters int
			wantOp      constants.Operator
		}{
			{"多值-IN", []string{"s1", "s2"}, 1, constants.OP_IN},
			{"单值-EQ", []string{"s1"}, 1, constants.OP_EQ},
			{"空切片-跳过", []string{}, 0, ""},
			{"nil-跳过", nil, 0, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := NewQuery()
				q.AddEqOrInFilter("session_id", tt.values)
				assert.Equal(t, tt.wantFilters, len(q.Filters))
				if tt.wantFilters == 1 {
					assert.Equal(t, tt.wantOp, q.Filters[0].Operator)
				}
			})
		}
	})
}

// ==============================================================================
// Query 基础条件构建方法
// ==============================================================================

func TestQueryBasicConditions(t *testing.T) {
	t.Run("AddEqual", func(t *testing.T) {
		q := NewQuery()
		result := q.AddEqual("status", 1)
		assert.Equal(t, q, result)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "status", q.Filters[0].Field)
		assert.Equal(t, constants.OP_EQ, q.Filters[0].Operator)
		assert.Equal(t, 1, q.Filters[0].Value)
	})

	t.Run("AddNotEqual", func(t *testing.T) {
		q := NewQuery()
		q.AddNotEqual("status", 0)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NEQ, q.Filters[0].Operator)
		assert.Equal(t, 0, q.Filters[0].Value)
	})

	t.Run("AddLike", func(t *testing.T) {
		q := NewQuery()
		q.AddLike("name", "test")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LIKE, q.Filters[0].Operator)
		assert.Equal(t, "%test%", q.Filters[0].Value)

		q = NewQuery()
		q.AddLike("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddStartsWith", func(t *testing.T) {
		q := NewQuery()
		q.AddStartsWith("username", "admin")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_STARTS_WITH, q.Filters[0].Operator)
		assert.Equal(t, "admin", q.Filters[0].Value)

		q = NewQuery()
		q.AddStartsWith("username", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddEndsWith", func(t *testing.T) {
		q := NewQuery()
		q.AddEndsWith("email", "@example.com")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_ENDS_WITH, q.Filters[0].Operator)

		q = NewQuery()
		q.AddEndsWith("email", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddIn", func(t *testing.T) {
		q := NewQuery()
		q.AddIn("status", 1, 2, 3)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 3, len(values))

		q = NewQuery()
		q.AddIn("status")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddNotIn", func(t *testing.T) {
		q := NewQuery()
		q.AddNotIn("status", "deleted", "disabled")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NOT_IN, q.Filters[0].Operator)

		q = NewQuery()
		q.AddNotIn("status")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddGreaterThan", func(t *testing.T) {
		q := NewQuery()
		q.AddGreaterThan("age", 18)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GT, q.Filters[0].Operator)
		assert.Equal(t, 18, q.Filters[0].Value)
	})

	t.Run("AddGreaterEqual", func(t *testing.T) {
		q := NewQuery()
		q.AddGreaterEqual("score", 60)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
		assert.Equal(t, 60, q.Filters[0].Value)
	})

	t.Run("AddLessThan", func(t *testing.T) {
		q := NewQuery()
		q.AddLessThan("price", 100.0)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LT, q.Filters[0].Operator)
		assert.Equal(t, 100.0, q.Filters[0].Value)
	})

	t.Run("AddLessEqual", func(t *testing.T) {
		q := NewQuery()
		q.AddLessEqual("count", 50)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LTE, q.Filters[0].Operator)
		assert.Equal(t, 50, q.Filters[0].Value)
	})

	t.Run("AddBetween", func(t *testing.T) {
		q := NewQuery()
		q.AddBetween("age", 18, 65)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))
	})

	t.Run("AddIsNull", func(t *testing.T) {
		q := NewQuery()
		q.AddIsNull("deleted_at")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IS_NULL, q.Filters[0].Operator)
		assert.Nil(t, q.Filters[0].Value)
	})

	t.Run("AddIsNotNull", func(t *testing.T) {
		q := NewQuery()
		q.AddIsNotNull("email")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_IS_NOT_NULL, q.Filters[0].Operator)
		assert.Nil(t, q.Filters[0].Value)
	})
}

// ==============================================================================
// Query 排序方法
// ==============================================================================

func TestQueryOrderMethods(t *testing.T) {
	t.Run("AddOrder", func(t *testing.T) {
		q := NewQuery()
		q.AddOrder("created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "created_at", q.Orders[0].Field)
		assert.Equal(t, "DESC", q.Orders[0].Direction)
	})

	t.Run("AddOrderAsc", func(t *testing.T) {
		q := NewQuery()
		result := q.AddOrderAsc("name")
		assert.Equal(t, q, result)
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "name", q.Orders[0].Field)
		assert.Equal(t, constants.Asc, q.Orders[0].Direction)
	})

	t.Run("AddOrderDesc", func(t *testing.T) {
		q := NewQuery()
		q.AddOrderDesc("created_at")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, constants.Desc, q.Orders[0].Direction)
	})

	t.Run("AddRawOrder", func(t *testing.T) {
		q := NewQuery()
		result := q.AddRawOrder("created_at DESC, updated_at ASC")
		assert.Len(t, result.Orders, 1)
		assert.Equal(t, "created_at DESC, updated_at ASC", result.Orders[0].Field)
		assert.Equal(t, "", result.Orders[0].Direction)
		assert.Equal(t, q, result)

		q = NewQuery()
		q.AddRawOrder("")
		assert.Len(t, q.Orders, 0)

		q = NewQuery()
		q.AddRawOrder("priority DESC").
			AddRawOrder("CASE WHEN status = 'urgent' THEN 1 ELSE 2 END").
			AddOrder("id", "ASC")
		assert.Len(t, q.Orders, 3)
		assert.Equal(t, "priority DESC", q.Orders[0].Field)
		assert.Equal(t, "CASE WHEN status = 'urgent' THEN 1 ELSE 2 END", q.Orders[1].Field)
		assert.Equal(t, "id", q.Orders[2].Field)
		assert.Equal(t, "ASC", q.Orders[2].Direction)
	})

	t.Run("AddSafeOrder", func(t *testing.T) {
		q := NewQuery()
		q.AddSafeOrder("", "", "created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "created_at", q.Orders[0].Field)
		assert.Equal(t, "DESC", q.Orders[0].Direction)

		q = NewQuery()
		q.AddSafeOrder("updated_at", "ASC", "created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "updated_at", q.Orders[0].Field)
		assert.Equal(t, "ASC", q.Orders[0].Direction)

		allowedFields := []string{"id", "created_at", "updated_at", "name"}
		q = NewQuery()
		q.AddSafeOrder("name", "ASC", "created_at", "DESC", allowedFields)
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "name", q.Orders[0].Field)

		allowedFields = []string{"id", "created_at", "updated_at"}
		q = NewQuery()
		q.AddSafeOrder("malicious_field", "ASC", "created_at", "DESC", allowedFields)
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "created_at", q.Orders[0].Field)

		q = NewQuery()
		q.AddSafeOrder("id; DROP TABLE users--", "DESC", "created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "created_at", q.Orders[0].Field)

		q = NewQuery()
		q.AddSafeOrder("id", "INVALID", "created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "id", q.Orders[0].Field)
		assert.Equal(t, "DESC", q.Orders[0].Direction)

		q = NewQuery()
		q.AddSafeOrder("id", "asc", "created_at", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "ASC", q.Orders[0].Direction)

		q = NewQuery()
		q.AddSafeOrder("id", "DeSc", "created_at", "ASC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "DESC", q.Orders[0].Direction)

		q = NewQuery()
		q.AddSafeOrder("valid_field_123", "ASC", "created_at", "DESC", []string{})
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "valid_field_123", q.Orders[0].Field)

		q = NewQuery()
		q.AddSafeOrder("users.created_at", "ASC", "id", "DESC")
		assert.Equal(t, 1, len(q.Orders))
		assert.Equal(t, "users.created_at", q.Orders[0].Field)

		q = NewQuery()
		q.AddSafeOrder("name", "ASC", "id", "DESC").
			AddSafeOrder("created_at", "DESC", "updated_at", "ASC")
		assert.Equal(t, 2, len(q.Orders))
	})

	t.Run("AddSafeOrder_特殊字符被阻止", func(t *testing.T) {
		testCases := []struct {
			name   string
			sortBy string
		}{
			{"空格", "created at"},
			{"单引号", "id'OR'1'='1"},
			{"双引号", "id\"OR\"1\"=\"1"},
			{"反引号", "id`"},
			{"括号", "id()"},
			{"星号", "id*"},
			{"逗号", "id,name"},
			{"分号", "id;DROP TABLE"},
			{"减号", "id-name"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				q := NewQuery()
				q.AddSafeOrder(tc.sortBy, "ASC", "created_at", "DESC")
				assert.Equal(t, 1, len(q.Orders))
				assert.Equal(t, "created_at", q.Orders[0].Field)
			})
		}
	})
}

// ==============================================================================
// Query 时间相关方法
// ==============================================================================

func TestQueryTimeMethods(t *testing.T) {
	t.Run("AddTimeAfter", func(t *testing.T) {
		testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		q := NewQuery()
		q.AddTimeAfter("created_at", testTime)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GT, q.Filters[0].Operator)
		assert.Equal(t, testTime, q.Filters[0].Value)

		q = NewQuery()
		q.AddTimeAfter("created_at", time.Time{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddTimeBefore", func(t *testing.T) {
		testTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
		q := NewQuery()
		q.AddTimeBefore("updated_at", testTime)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LT, q.Filters[0].Operator)

		q = NewQuery()
		q.AddTimeBefore("updated_at", time.Time{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddTimeBetween", func(t *testing.T) {
		start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

		q := NewQuery()
		q.AddTimeBetween("created_at", start, end)
		assert.Equal(t, 2, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)
		assert.Equal(t, constants.OP_LTE, q.Filters[1].Operator)

		q = NewQuery()
		q.AddTimeBetween("created_at", start, time.Time{})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_GTE, q.Filters[0].Operator)

		q = NewQuery()
		q.AddTimeBetween("created_at", time.Time{}, end)
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_LTE, q.Filters[0].Operator)

		q = NewQuery()
		q.AddTimeBetween("created_at", time.Time{}, time.Time{})
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("AddToday", func(t *testing.T) {
		q := NewQuery()
		q.AddToday("created_at")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))
		startTime := values[0].(time.Time)
		endTime := values[1].(time.Time)
		now := time.Now()
		assert.Equal(t, now.Year(), startTime.Year())
		assert.Equal(t, now.Month(), startTime.Month())
		assert.Equal(t, now.Day(), startTime.Day())
		assert.True(t, endTime.After(startTime))
	})

	t.Run("AddThisWeek", func(t *testing.T) {
		q := NewQuery()
		q.AddThisWeek("created_at")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))

		start := values[0].(time.Time)
		end := values[1].(time.Time)

		// 按墙上时钟断言区间边界，而不是按时长
		//
		// 原先断言 duration < 7*24h+1s，把「一周恒等于 168 小时」写进了测试
		// 夏令时切换那一周并不是：2026-10-26 ~ 11-01 这周美国时区在 11-01 回拨
		// 一小时，区间实际跨 168h59m59.999999999s，于是每年固定有一周会挂
		// 而「本周」要的本来就是墙上时钟语义——周一零点到周日最后一刻，
		// 这也正是 SQL BETWEEN 需要的边界，实现没有问题
		//
		// [EN] The bounds are asserted on the wall clock rather than on the
		// elapsed duration. The old assertion baked in "a week is 168 hours",
		// which the DST-transition week is not: in US zones the week of
		// 2026-10-26 falls back an hour on 11-01 and spans 168h59m59.999999999s,
		// so the test failed for one week every year. "This week" means wall
		// clock anyway — Monday midnight through the last instant of Sunday,
		// which is exactly what SQL BETWEEN needs
		assert.Equal(t, time.Monday, start.Weekday(), "起点应为本周一")
		assert.Equal(t, 0, start.Hour())
		assert.Equal(t, 0, start.Minute())
		assert.Equal(t, 0, start.Second())
		assert.Equal(t, 0, start.Nanosecond())

		assert.Equal(t, time.Sunday, end.Weekday(), "终点应落在本周日")
		assert.Equal(t, start.AddDate(0, 0, 7).Add(-time.Nanosecond), end,
			"终点应为下周一零点减一纳秒")

		// 当下必须落在区间内，否则周一的计算偏了
		now := time.Now()
		assert.False(t, now.Before(start), "now=%v 早于区间起点 %v", now, start)
		assert.False(t, now.After(end), "now=%v 晚于区间终点 %v", now, end)
	})

	t.Run("AddThisMonth", func(t *testing.T) {
		q := NewQuery()
		q.AddThisMonth("created_at")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))
		now := time.Now()
		startTime := values[0].(time.Time)
		assert.Equal(t, now.Year(), startTime.Year())
		assert.Equal(t, now.Month(), startTime.Month())
		assert.Equal(t, 1, startTime.Day())
	})

	t.Run("AddThisYear", func(t *testing.T) {
		q := NewQuery()
		q.AddThisYear("created_at")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_BETWEEN, q.Filters[0].Operator)
		values := q.Filters[0].Value.([]interface{})
		assert.Equal(t, 2, len(values))
		now := time.Now()
		startTime := values[0].(time.Time)
		assert.Equal(t, now.Year(), startTime.Year())
		assert.Equal(t, time.January, startTime.Month())
		assert.Equal(t, 1, startTime.Day())
	})
}

// ==============================================================================
// Query 分页与限制方法
// ==============================================================================

func TestQueryPagingMethods(t *testing.T) {
	t.Run("WithDistinct", func(t *testing.T) {
		q := NewQuery()
		result := q.WithDistinct()
		assert.Equal(t, q, result)
		assert.True(t, q.Distinct)
	})

	t.Run("WithPaging", func(t *testing.T) {
		q := NewQuery()
		result := q.WithPaging(2, 20)
		assert.Equal(t, q, result)
		assert.NotNil(t, q.Pagination)
		assert.Equal(t, 2, q.Pagination.Page)
		assert.Equal(t, 20, q.Pagination.PageSize)

		q = NewQuery()
		q.WithPaging(0, -5)
		assert.Equal(t, 1, q.Pagination.Page)
		assert.Equal(t, constants.DefaultPageSize, q.Pagination.PageSize)

		q = NewQuery()
		q.WithPaging(-5, -10)
		assert.Equal(t, constants.DefaultPage, q.Pagination.Page)
		assert.Equal(t, constants.DefaultPageSize, q.Pagination.PageSize)
	})

	t.Run("SetPage", func(t *testing.T) {
		q := NewQuery().WithPaging(1, 10)
		q.SetPage(5)
		assert.Equal(t, 5, q.Pagination.Page)

		q = NewQuery()
		q.SetPage(3)
		assert.NotNil(t, q.Pagination)
		assert.Equal(t, 3, q.Pagination.Page)
	})

	t.Run("SetPageSize", func(t *testing.T) {
		q := NewQuery().WithPaging(1, 10)
		q.SetPageSize(50)
		assert.Equal(t, 50, q.Pagination.PageSize)

		q = NewQuery()
		q.SetPageSize(20)
		assert.NotNil(t, q.Pagination)
		assert.Equal(t, 20, q.Pagination.PageSize)
	})

	t.Run("SetPagination", func(t *testing.T) {
		q := NewQuery()
		q.SetPagination(3, 25)
		assert.NotNil(t, q.Pagination)
		assert.Equal(t, 3, q.Pagination.Page)
		assert.Equal(t, 25, q.Pagination.PageSize)
	})

	t.Run("GetPagination", func(t *testing.T) {
		q := NewQuery().WithPaging(2, 20)
		p := q.GetPagination()
		assert.Equal(t, 2, p.Page)
		assert.Equal(t, 20, p.PageSize)

		q = NewQuery()
		p = q.GetPagination()
		assert.NotNil(t, p)
		assert.Equal(t, constants.DefaultPage, p.Page)
		assert.Equal(t, constants.DefaultPageSize, p.PageSize)
	})

	t.Run("Limit", func(t *testing.T) {
		q := NewQuery()
		result := q.Limit(10)
		assert.Equal(t, q, result)
		assert.NotNil(t, q.LimitValue)
		assert.Equal(t, 10, *q.LimitValue)
	})

	t.Run("Offset", func(t *testing.T) {
		q := NewQuery()
		result := q.Offset(20)
		assert.Equal(t, q, result)
		assert.NotNil(t, q.OffsetValue)
		assert.Equal(t, 20, *q.OffsetValue)
	})
}

// ==============================================================================
// Query GroupBy / Having / Select / Omit
// ==============================================================================

func TestQueryGroupSelectOmit(t *testing.T) {
	t.Run("AddGroupBy", func(t *testing.T) {
		q := NewQuery()
		q.AddGroupBy("status")
		assert.True(t, q.HasGroupBy())

		q2 := &Query{}
		q2.AddGroupBy("status")
		assert.NotNil(t, q2.GroupBy)
		assert.Equal(t, 1, len(q2.GroupBy))
		assert.Equal(t, "status", q2.GroupBy[0])
	})

	t.Run("AddHaving", func(t *testing.T) {
		q := NewQuery()
		q.AddHaving(NewGtFilter("count", 5))
		assert.True(t, q.HasHaving())

		q2 := &Query{}
		q2.AddHaving(nil)
		assert.Equal(t, 0, len(q2.Having))

		q3 := &Query{}
		q3.AddHaving(NewEqFilter("count", 10))
		assert.NotNil(t, q3.Having)
		assert.Equal(t, 1, len(q3.Having))
	})

	t.Run("Select", func(t *testing.T) {
		q := NewQuery()
		q.Select("id", "name")
		assert.True(t, q.HasSelectFields())
	})

	t.Run("Omit", func(t *testing.T) {
		q := NewQuery()
		q.Omit("password")
		assert.True(t, q.HasOmitFields())
	})

	t.Run("OmitSensitive", func(t *testing.T) {
		q := NewQuery()
		q.OmitSensitive()
		assert.True(t, q.HasOmitFields())
		assert.Contains(t, q.OmitFields, "password")
		assert.Contains(t, q.OmitFields, "token")
	})

	t.Run("OmitLargeFields", func(t *testing.T) {
		q := NewQuery()
		q.OmitLargeFields()
		assert.True(t, q.HasOmitFields())
		assert.Contains(t, q.OmitFields, "content")
		assert.Contains(t, q.OmitFields, "description")
	})
}

// ==============================================================================
// Query 状态检查方法
// ==============================================================================

func TestQueryHasMethods(t *testing.T) {
	t.Run("HasPagination", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasPagination())
		q.WithPaging(1, 10)
		assert.True(t, q.HasPagination())
	})

	t.Run("HasGroupBy", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasGroupBy())
		q.AddGroupBy("status")
		assert.True(t, q.HasGroupBy())
	})

	t.Run("HasHaving", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasHaving())
		q.AddHaving(NewGtFilter("count", 5))
		assert.True(t, q.HasHaving())
	})

	t.Run("HasOrders", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasOrders())
		q.AddOrder("created_at", constants.Desc)
		assert.True(t, q.HasOrders())
	})

	t.Run("HasSelectFields", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasSelectFields())
		q.Select("id", "name")
		assert.True(t, q.HasSelectFields())
	})

	t.Run("HasOmitFields", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasOmitFields())
		q.Omit("password")
		assert.True(t, q.HasOmitFields())
	})

	t.Run("IsLimited", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.IsLimited())
		q.Limit(50)
		assert.True(t, q.IsLimited())
	})

	t.Run("IsOffset", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.IsOffset())
		q.Offset(100)
		assert.True(t, q.IsOffset())
	})

	t.Run("HasFilters", func(t *testing.T) {
		q := NewQuery()
		assert.False(t, q.HasFilters())
		q.AddEqual("name", "test")
		assert.True(t, q.HasFilters())
	})

	t.Run("GetAllFilters", func(t *testing.T) {
		q := NewQuery()
		q.AddEqual("name", "test")
		allFilters := q.GetAllFilters()
		assert.Equal(t, 1, len(allFilters))

		q.WithFilterGroup(NewFilterGroup(constants.LOGIC_AND).AddFilter(NewEqFilter("age", 18)))
		allFilters = q.GetAllFilters()
		assert.Equal(t, 2, len(allFilters))
	})
}

// ==============================================================================
// Query 重置方法
// ==============================================================================

func TestQueryResetMethods(t *testing.T) {
	t.Run("ResetFilters", func(t *testing.T) {
		q := NewQuery().
			AddEqual("status", "active").
			WithFilterGroup(NewFilterGroup(constants.LOGIC_AND).AddFilter(NewEqFilter("age", 18)))
		assert.True(t, q.HasFilters())
		q.ResetFilters()
		assert.False(t, q.HasFilters())
		assert.Len(t, q.Filters, 0)
		assert.Nil(t, q.FilterGroup)
	})

	t.Run("ResetOrders", func(t *testing.T) {
		q := NewQuery().AddOrderAsc("name").AddOrderDesc("created_at")
		assert.True(t, q.HasOrders())
		q.ResetOrders()
		assert.False(t, q.HasOrders())
		assert.Len(t, q.Orders, 0)
	})

	t.Run("ResetPagination", func(t *testing.T) {
		q := NewQuery().WithPaging(2, 20)
		assert.True(t, q.HasPagination())
		q.ResetPagination()
		assert.False(t, q.HasPagination())
		assert.Nil(t, q.Pagination)
	})
}

// ==============================================================================
// Query.Clone
// ==============================================================================

func TestQueryClone(t *testing.T) {
	t.Run("正常克隆", func(t *testing.T) {
		q := NewQuery().
			AddEqual("name", "test").
			AddOrderAsc("created_at").
			WithPaging(1, 10).
			Limit(100).
			WithDistinct()
		cloned := q.Clone()
		assert.Equal(t, len(q.Filters), len(cloned.Filters))
		assert.Equal(t, len(q.Orders), len(cloned.Orders))
		assert.Equal(t, q.Pagination.Page, cloned.Pagination.Page)
		assert.Equal(t, q.Distinct, cloned.Distinct)
		assert.Equal(t, *q.LimitValue, *cloned.LimitValue)
	})

	t.Run("克隆独立性", func(t *testing.T) {
		q := NewQuery().AddEqual("name", "test")
		cloned := q.Clone()
		cloned.AddEqual("age", 25)
		assert.NotEqual(t, len(q.Filters), len(cloned.Filters))
	})

	t.Run("nil查询克隆", func(t *testing.T) {
		var q *Query
		cloned := q.Clone()
		assert.NotNil(t, cloned)
	})
}

// ==============================================================================
// Query.WithFilterGroup
// ==============================================================================

func TestWithFilterGroup(t *testing.T) {
	t.Run("设置FilterGroup", func(t *testing.T) {
		q := NewQuery()
		group := NewFilterGroup(constants.LOGIC_AND).AddFilter(NewEqFilter("name", "test"))
		result := q.WithFilterGroup(group)
		assert.Equal(t, q, result)
		assert.NotNil(t, q.FilterGroup)
	})
}

// ==============================================================================
// Query.BuildWhereClause
// ==============================================================================

func TestBuildWhereClause(t *testing.T) {
	t.Run("nil查询", func(t *testing.T) {
		var q *Query
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("空查询", func(t *testing.T) {
		q := NewQuery()
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("单个EQ条件", func(t *testing.T) {
		q := NewQuery().AddEqual("name", "test")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "name")
		assert.Equal(t, 1, len(args))
		assert.Equal(t, "test", args[0])
	})

	t.Run("多个条件AND连接", func(t *testing.T) {
		q := NewQuery().
			AddEqual("name", "test").
			AddGreaterThan("age", 18)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "AND")
		assert.Equal(t, 2, len(args))
	})

	t.Run("nil过滤器被跳过", func(t *testing.T) {
		q := NewQuery()
		q.Filters = []*Filter{nil, NewEqFilter("name", "test"), nil}
		_, args := q.BuildWhereClause()
		assert.Equal(t, 1, len(args))
	})

	t.Run("RAW条件", func(t *testing.T) {
		q := NewQuery().AddRawFilter("status = 1")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "status = 1")
		assert.Equal(t, 0, len(args))
	})

	t.Run("IS NULL条件", func(t *testing.T) {
		q := NewQuery().AddIsNull("deleted_at")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "IS NULL")
		assert.Equal(t, 0, len(args))
	})

	t.Run("IS NOT NULL条件", func(t *testing.T) {
		q := NewQuery().AddIsNotNull("email")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "IS NOT NULL")
		assert.Equal(t, 0, len(args))
	})

	t.Run("BETWEEN条件", func(t *testing.T) {
		q := NewQuery().AddBetween("age", 18, 65)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "BETWEEN")
		assert.Equal(t, 2, len(args))
		assert.Equal(t, 18, args[0])
		assert.Equal(t, 65, args[1])
	})

	t.Run("STARTS_WITH条件", func(t *testing.T) {
		q := NewQuery().AddStartsWith("name", "admin")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "LIKE")
		assert.Equal(t, 1, len(args))
		assert.Equal(t, "admin%", args[0])
	})

	t.Run("ENDS_WITH条件", func(t *testing.T) {
		q := NewQuery().AddEndsWith("email", "@example.com")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "LIKE")
		assert.Equal(t, 1, len(args))
		assert.Equal(t, "%@example.com", args[0])
	})

	t.Run("FIND_IN_SET条件", func(t *testing.T) {
		q := NewQuery().AddFindInSetFilterIfNotEmpty("tags", "go")
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "FIND_IN_SET")
		assert.Contains(t, clause, "> 0")
		assert.Equal(t, 1, len(args))
	})

	t.Run("FilterGroup-AND", func(t *testing.T) {
		q := NewQuery().WithFilterGroup(
			NewFilterGroup(constants.LOGIC_AND).
				AddFilter(NewEqFilter("name", "test")).
				AddFilter(NewGtFilter("age", 18)),
		)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "AND")
		assert.Equal(t, 2, len(args))
	})

	t.Run("FilterGroup-OR", func(t *testing.T) {
		q := NewQuery().WithFilterGroup(
			NewFilterGroup(constants.LOGIC_OR).
				AddFilter(NewEqFilter("status", "active")).
				AddFilter(NewEqFilter("status", "pending")),
		)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "OR")
		assert.Equal(t, 2, len(args))
	})

	t.Run("FilterGroup单条件不加括号", func(t *testing.T) {
		q := NewQuery().WithFilterGroup(
			NewFilterGroup(constants.LOGIC_AND).
				AddFilter(NewEqFilter("name", "test")),
		)
		clause, _ := q.BuildWhereClause()
		assert.NotContains(t, clause, "(")
	})

	t.Run("FilterGroup多条件加括号", func(t *testing.T) {
		q := NewQuery().WithFilterGroup(
			NewFilterGroup(constants.LOGIC_AND).
				AddFilter(NewEqFilter("name", "test")).
				AddFilter(NewGtFilter("age", 18)),
		)
		clause, _ := q.BuildWhereClause()
		assert.Contains(t, clause, "(")
		assert.Contains(t, clause, ")")
	})

	t.Run("嵌套FilterGroup", func(t *testing.T) {
		innerGroup := NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("status", "active")).
			AddFilter(NewEqFilter("status", "pending"))
		outerGroup := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(innerGroup)
		q := NewQuery().WithFilterGroup(outerGroup)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "OR")
		assert.Contains(t, clause, "AND")
		assert.Equal(t, 3, len(args))
	})

	t.Run("空FilterGroup被跳过", func(t *testing.T) {
		q := NewQuery().WithFilterGroup(NewFilterGroup(constants.LOGIC_AND))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("简单过滤器和FilterGroup混合", func(t *testing.T) {
		q := NewQuery().
			AddEqual("type", "user").
			WithFilterGroup(
				NewFilterGroup(constants.LOGIC_OR).
					AddFilter(NewEqFilter("status", "active")).
					AddFilter(NewEqFilter("status", "pending")),
			)
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "AND")
		assert.Equal(t, 3, len(args))
	})

	t.Run("未知操作符返回空", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.Operator("UNKNOWN_OP"), "value"))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("BETWEEN值不是双元素切片", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.OP_BETWEEN, "not_a_slice"))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("STARTS_WITH值非字符串", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.OP_STARTS_WITH, 123))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("ENDS_WITH值非字符串", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.OP_ENDS_WITH, 123))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("CONTAINS操作符", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.OP_CONTAINS, "test"))
		clause, args := q.BuildWhereClause()
		assert.Contains(t, clause, "LIKE")
		assert.Equal(t, 1, len(args))
		assert.Equal(t, "%test%", args[0])
	})

	t.Run("CONTAINS值非字符串", func(t *testing.T) {
		q := NewQuery()
		q.AddFilter(NewFilter("field", constants.OP_CONTAINS, 123))
		clause, args := q.BuildWhereClause()
		assert.Equal(t, "", clause)
		assert.Nil(t, args)
	})

	t.Run("nil过滤器在buildFilterCondition", func(t *testing.T) {
		q := NewQuery()
		condition, arg := q.buildFilterCondition(nil)
		assert.Equal(t, "", condition)
		assert.Nil(t, arg)
	})

	t.Run("嵌套子组单条件不加括号", func(t *testing.T) {
		innerGroup := NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("status", "active"))
		outerGroup := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(innerGroup)
		q := NewQuery().WithFilterGroup(outerGroup)
		clause, _ := q.BuildWhereClause()
		assert.NotContains(t, clause, "OR")
		assert.Contains(t, clause, "AND")
	})

	t.Run("nil子组被跳过", func(t *testing.T) {
		outerGroup := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test"))
		outerGroup.AddGroup(nil)
		q := NewQuery().WithFilterGroup(outerGroup)
		_, args := q.BuildWhereClause()
		assert.Equal(t, 1, len(args))
	})

	t.Run("空子组被跳过", func(t *testing.T) {
		outerGroup := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test"))
		outerGroup.AddGroup(NewFilterGroup(constants.LOGIC_OR))
		q := NewQuery().WithFilterGroup(outerGroup)
		_, args := q.BuildWhereClause()
		assert.Equal(t, 1, len(args))
	})
}

// ==============================================================================
// FilterGroup 剩余方法
// ==============================================================================

func TestFilterGroupRemainingMethods(t *testing.T) {
	t.Run("AddFilterIf", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddFilterIf(true, NewEqFilter("name", "test"))
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, fg.Count())

		fg.AddFilterIf(false, NewEqFilter("age", 18))
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddFilterIfValueNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddFilterIfValueNotEmpty(NewEqFilter("name", "test"))
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, fg.Count())

		fg.AddFilterIfValueNotEmpty(NewEqFilter("age", ""))
		assert.Equal(t, 1, fg.Count())

		fg.AddFilterIfValueNotEmpty(NewEqFilter("age", nil))
		assert.Equal(t, 1, fg.Count())

		fg.AddFilterIfValueNotEmpty(nil)
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddRegexpFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddRegexpFilterIfNotEmpty("name", "^test.*")
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, fg.Count())

		fg.AddRegexpFilterIfNotEmpty("email", "")
		assert.Equal(t, 1, fg.Count())

		fg.AddRegexpFilterIfNotEmpty("phone", nil)
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddGroupIf", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		subGroup := NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("status", "active"))
		result := fg.AddGroupIf(true, subGroup)
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, len(fg.Groups))

		fg.AddGroupIf(false, NewFilterGroup(constants.LOGIC_AND))
		assert.Equal(t, 1, len(fg.Groups))
	})

	t.Run("AddGroupIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		nonEmptyGroup := NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("status", "active"))
		result := fg.AddGroupIfNotEmpty(nonEmptyGroup)
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, len(fg.Groups))

		emptyGroup := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGroupIfNotEmpty(emptyGroup)
		assert.Equal(t, 1, len(fg.Groups))

		fg.AddGroupIfNotEmpty(nil)
		assert.Equal(t, 1, len(fg.Groups))
	})

	t.Run("Clear", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddGroup(NewFilterGroup(constants.LOGIC_OR))
		result := fg.Clear()
		assert.Equal(t, fg, result)
		assert.True(t, fg.IsEmpty())
		assert.Equal(t, 0, len(fg.Groups))
	})

	t.Run("Clone", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND).
			AddFilter(NewEqFilter("name", "test")).
			AddFilter(NewGtFilter("age", 18))
		cloned := fg.Clone()
		assert.Equal(t, fg.LogicOp, cloned.LogicOp)
		assert.Equal(t, len(fg.Filters), len(cloned.Filters))

		cloned.AddFilter(NewEqFilter("extra", "value"))
		assert.NotEqual(t, len(fg.Filters), len(cloned.Filters))
	})

	t.Run("getDepth", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		assert.Equal(t, 1, fg.getDepth())

		innerGroup := NewFilterGroup(constants.LOGIC_OR)
		fg.AddGroup(innerGroup)
		assert.Equal(t, 2, fg.getDepth())

		deepGroup := NewFilterGroup(constants.LOGIC_AND)
		innerGroup.AddGroup(deepGroup)
		assert.Equal(t, 3, fg.getDepth())
	})

	t.Run("Count", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		assert.Equal(t, 0, fg.Count())

		fg.AddFilter(NewEqFilter("name", "test"))
		assert.Equal(t, 1, fg.Count())

		fg.AddFilter(NewGtFilter("age", 18))
		assert.Equal(t, 2, fg.Count())

		innerGroup := NewFilterGroup(constants.LOGIC_OR).
			AddFilter(NewEqFilter("status", "active"))
		fg.AddGroup(innerGroup)
		assert.Equal(t, 3, fg.Count())
	})

	t.Run("AddEqFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddEqFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_EQ, fg.Filters[0].Operator)

		fg.AddEqFilterIfNotEmpty("email", "")
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddNeqFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNeqFilterIfNotEmpty("status", "deleted")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_NEQ, fg.Filters[0].Operator)

		fg.AddNeqFilterIfNotEmpty("status", "")
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddGtFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGtFilterIfNotEmpty("age", 18)
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_GT, fg.Filters[0].Operator)
	})

	t.Run("AddGteFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddGteFilterIfNotEmpty("score", 60)
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_GTE, fg.Filters[0].Operator)
	})

	t.Run("AddLtFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLtFilterIfNotEmpty("price", 100)
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_LT, fg.Filters[0].Operator)
	})

	t.Run("AddLteFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLteFilterIfNotEmpty("count", 50)
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_LTE, fg.Filters[0].Operator)
	})

	t.Run("AddLikeFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddLikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_LIKE, fg.Filters[0].Operator)

		fg.AddLikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddInFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddInFilterIfNotEmpty("status", []string{"active", "pending"})
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_IN, fg.Filters[0].Operator)

		fg.AddInFilterIfNotEmpty("status", []string{})
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddNotInFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotInFilterIfNotEmpty("status", []string{"deleted", "disabled"})
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_NOT_IN, fg.Filters[0].Operator)
	})

	t.Run("AddBetweenFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddBetweenFilterIfNotEmpty("age", 18, 65)
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_BETWEEN, fg.Filters[0].Operator)

		fg.AddBetweenFilterIfNotEmpty("price", nil, 100)
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddStartsWithFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddStartsWithFilterIfNotEmpty("name", "admin")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_STARTS_WITH, fg.Filters[0].Operator)
	})

	t.Run("AddEndsWithFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddEndsWithFilterIfNotEmpty("email", "@example.com")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_ENDS_WITH, fg.Filters[0].Operator)
	})

	t.Run("AddNotLikeFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotLikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_NOT_LIKE, fg.Filters[0].Operator)

		fg.AddNotLikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 1, fg.Count())
	})

	t.Run("AddFindInSetFilterIfNotEmpty", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddFindInSetFilterIfNotEmpty("tags", "go")
		assert.Equal(t, 1, fg.Count())
		assert.Equal(t, constants.OP_FIND_IN_SET, fg.Filters[0].Operator)

		fg.AddFindInSetFilterIfNotEmpty("tags", "")
		assert.Equal(t, 1, fg.Count())
	})
}

// ==============================================================================
// Filter 构造函数补充
// ==============================================================================

func TestFilterConstructorsExtended(t *testing.T) {
	t.Run("NewInFilterSlice", func(t *testing.T) {
		values := []interface{}{1, 2, 3}
		f := NewInFilterSlice("id", values)
		assert.Equal(t, "id", f.Field)
		assert.Equal(t, constants.OP_IN, f.Operator)
		resultValues := f.Value.([]interface{})
		assert.Equal(t, 3, len(resultValues))
	})

	t.Run("NewInFilterWithTypedSlice", func(t *testing.T) {
		// 传入 []int64 单个切片参数应自动展开为独立元素，避免生成 IN ((...)) 双重括号
		ids := []int64{1, 2, 3}
		f := NewInFilter("id", ids)
		assert.Equal(t, "id", f.Field)
		assert.Equal(t, constants.OP_IN, f.Operator)
		resultValues := f.Value.([]interface{})
		assert.Equal(t, 3, len(resultValues))
		assert.Equal(t, int64(1), resultValues[0])
		assert.Equal(t, int64(3), resultValues[2])
	})

	t.Run("NewNotInFilterSlice", func(t *testing.T) {
		values := []interface{}{"a", "b"}
		f := NewNotInFilterSlice("status", values)
		assert.Equal(t, "status", f.Field)
		assert.Equal(t, constants.OP_NOT_IN, f.Operator)
		resultValues := f.Value.([]interface{})
		assert.Equal(t, 2, len(resultValues))
	})

	t.Run("NewNotInFilterWithTypedSlice", func(t *testing.T) {
		// 传入 []string 单个切片参数应自动展开为独立元素，避免生成 NOT IN ((...)) 双重括号
		codes := []string{"a", "b", "c"}
		f := NewNotInFilter("code", codes)
		assert.Equal(t, "code", f.Field)
		assert.Equal(t, constants.OP_NOT_IN, f.Operator)
		resultValues := f.Value.([]interface{})
		assert.Equal(t, 3, len(resultValues))
		assert.Equal(t, "a", resultValues[0])
		assert.Equal(t, "c", resultValues[2])
	})

	t.Run("NewRegexpFilter", func(t *testing.T) {
		f := NewRegexpFilter("name", "^test.*")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_REGEXP, f.Operator)
		assert.Equal(t, "^test.*", f.Value)
	})

	t.Run("NewFindInSetFilter", func(t *testing.T) {
		f := NewFindInSetFilter("tags", "go")
		assert.Equal(t, "tags", f.Field)
		assert.Equal(t, constants.OP_FIND_IN_SET, f.Operator)
		assert.Equal(t, "go", f.Value)
	})

	t.Run("NewNotLikeFilter", func(t *testing.T) {
		f := NewNotLikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_NOT_LIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})

	t.Run("NewStartsWithFilter", func(t *testing.T) {
		f := NewStartsWithFilter("name", "admin")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_STARTS_WITH, f.Operator)
		assert.Equal(t, "admin", f.Value)
	})

	t.Run("NewEndsWithFilter", func(t *testing.T) {
		f := NewEndsWithFilter("email", "@example.com")
		assert.Equal(t, "email", f.Field)
		assert.Equal(t, constants.OP_ENDS_WITH, f.Operator)
		assert.Equal(t, "@example.com", f.Value)
	})

	t.Run("NewIsNullFilter", func(t *testing.T) {
		f := NewIsNullFilter("deleted_at")
		assert.Equal(t, "deleted_at", f.Field)
		assert.Equal(t, constants.OP_IS_NULL, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewIsNotNullFilter", func(t *testing.T) {
		f := NewIsNotNullFilter("email")
		assert.Equal(t, "email", f.Field)
		assert.Equal(t, constants.OP_IS_NOT_NULL, f.Operator)
		assert.Nil(t, f.Value)
	})

	t.Run("NewBetweenFilter", func(t *testing.T) {
		f := NewBetweenFilter("age", 18, 65)
		assert.Equal(t, "age", f.Field)
		assert.Equal(t, constants.OP_BETWEEN, f.Operator)
		values := f.Value.([]interface{})
		assert.Equal(t, 2, len(values))
		assert.Equal(t, 18, values[0])
		assert.Equal(t, 65, values[1])
	})

	t.Run("NewNeqFilter", func(t *testing.T) {
		f := NewNeqFilter("status", "deleted")
		assert.Equal(t, "status", f.Field)
		assert.Equal(t, constants.OP_NEQ, f.Operator)
		assert.Equal(t, "deleted", f.Value)
	})

	t.Run("NewLikeFilter", func(t *testing.T) {
		f := NewLikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_LIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})
}

// ==============================================================================
// SubQuery
// ==============================================================================

func TestSubQueryInFilterFile(t *testing.T) {
	t.Run("NewSubQuery", func(t *testing.T) {
		sq := NewSubQuery("SELECT id FROM users WHERE active = ?", true)
		assert.Equal(t, "SELECT id FROM users WHERE active = ?", sq.SQL)
		assert.Equal(t, 1, len(sq.Args))
		assert.Equal(t, true, sq.Args[0])
	})

	t.Run("NewSubQuery无参数", func(t *testing.T) {
		sq := NewSubQuery("SELECT id FROM users")
		assert.Equal(t, "SELECT id FROM users", sq.SQL)
		assert.Equal(t, 0, len(sq.Args))
	})

	t.Run("NewSubQuery多参数", func(t *testing.T) {
		sq := NewSubQuery("SELECT id FROM users WHERE age > ? AND status = ?", 18, "active")
		assert.Equal(t, 2, len(sq.Args))
	})
}

// TestNewOperatorsIntegration 测试新增操作符的集成
func TestNewOperatorsIntegration(t *testing.T) {
	// 测试 NewStartsWithFilter
	filter := NewStartsWithFilter("name", "user")
	assert.Equal(t, "name", filter.Field)
	assert.Equal(t, constants.OP_STARTS_WITH, filter.Operator)
	assert.Equal(t, "user", filter.Value)

	// 测试 NewEndsWithFilter
	filter = NewEndsWithFilter("email", "@example.com")
	assert.Equal(t, "email", filter.Field)
	assert.Equal(t, constants.OP_ENDS_WITH, filter.Operator)
	assert.Equal(t, "@example.com", filter.Value)
}

// TestBuildFilterConditionWithNewOperators 测试构建过滤条件的新操作符
func TestBuildFilterConditionWithNewOperators(t *testing.T) {
	testCases := []struct {
		name              string
		filter            *Filter
		expectedCondition string
		expectedArg       interface{}
		shouldHaveArg     bool
	}{
		{
			name:              "STARTS_WITH filter",
			filter:            &Filter{Field: "name", Operator: constants.OP_STARTS_WITH, Value: "user"},
			expectedCondition: "name LIKE ?",
			expectedArg:       "user" + constants.SQL_WILDCARD_ANY,
			shouldHaveArg:     true,
		},
		{
			name:              "ENDS_WITH filter",
			filter:            &Filter{Field: "email", Operator: constants.OP_ENDS_WITH, Value: "@example.com"},
			expectedCondition: "email LIKE ?",
			expectedArg:       constants.SQL_WILDCARD_ANY + "@example.com",
			shouldHaveArg:     true,
		},
		{
			name:              "CONTAINS filter",
			filter:            &Filter{Field: "description", Operator: constants.OP_CONTAINS, Value: "keyword"},
			expectedCondition: "description LIKE ?",
			expectedArg:       constants.SQL_WILDCARD_ANY + "keyword" + constants.SQL_WILDCARD_ANY,
			shouldHaveArg:     true,
		},
		{
			name:              "STARTS_WITH with non-string value",
			filter:            &Filter{Field: "name", Operator: constants.OP_STARTS_WITH, Value: 123},
			expectedCondition: "",
			shouldHaveArg:     false,
		},
		{
			name:              "ENDS_WITH with empty string",
			filter:            &Filter{Field: "name", Operator: constants.OP_ENDS_WITH, Value: ""},
			expectedCondition: "name LIKE ?",
			expectedArg:       constants.SQL_WILDCARD_ANY,
			shouldHaveArg:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			condition, arg := buildFilterCondition(tc.filter, nil)
			assert.Equal(t, tc.expectedCondition, condition)

			if tc.shouldHaveArg {
				assert.Equal(t, tc.expectedArg, arg)
			} else {
				assert.Nil(t, arg)
			}
		})
	}
}

func TestProtobufWrapperFilterValues(t *testing.T) {
	t.Run("AddFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery()
		q.AddFilterIfNotEmpty("group_id", wrapperspb.String("g1"))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "g1", q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("group_id", wrapperspb.String(""))
		assert.Equal(t, 0, len(q.Filters))

		q = NewQuery()
		q.AddFilterIfNotEmpty("enabled", wrapperspb.Bool(false))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, false, q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("sort_order", wrapperspb.Int32(0))
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, int32(0), q.Filters[0].Value)

		q = NewQuery()
		q.AddFilterIfNotEmpty("code", wrapperspb.StringValue{Value: "TENANT"})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, "TENANT", q.Filters[0].Value)
	})

	t.Run("AddInFilterIfNotEmpty", func(t *testing.T) {
		q := NewQuery()
		q.AddInFilterIfNotEmpty("group_id", []*wrapperspb.StringValue{wrapperspb.String("g1"), wrapperspb.String("g2")})
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, []interface{}{"g1", "g2"}, q.Filters[0].Value)
	})

	t.Run("direct filter condition", func(t *testing.T) {
		condition, arg := buildFilterCondition(&Filter{
			Field:    "group_id",
			Operator: constants.OP_EQ,
			Value:    wrapperspb.String("g1"),
		}, nil)

		assert.Equal(t, "group_id = ?", condition)
		assert.Equal(t, "g1", arg)

		condition, arg = buildFilterCondition(&Filter{
			Field:    "code",
			Operator: constants.OP_LIKE,
			Value:    wrapperspb.StringValue{Value: "%tenant%"},
		}, nil)

		assert.Equal(t, "code LIKE ?", condition)
		assert.Equal(t, "%tenant%", arg)
	})
}

func TestQueryWithFiltersSkipsNilBuilders(t *testing.T) {
	query := NewQuery().WithFilters(nil, func(q *Query) {
		q.AddFilter(NewFilter("status", constants.OP_EQ, "active"))
	})

	assert.Len(t, query.Filters, 1)
	assert.Equal(t, "status", query.Filters[0].Field)
}

// ============================================================
// NewInSubQueryFilter / NewNotInSubQueryFilter 构造函数
// ============================================================

func TestNewInSubQueryFilter(t *testing.T) {
	f := NewInSubQueryFilter("group_id", "SELECT id FROM dict_groups WHERE type IN (?)", []string{"a", "b"})

	assert.Equal(t, "group_id", f.Field)
	assert.Equal(t, constants.OP_IN, f.Operator)
	sq, ok := f.Value.(*SubQuery)
	require.True(t, ok, "Value 应为 *SubQuery 类型")
	assert.Equal(t, "SELECT id FROM dict_groups WHERE type IN (?)", sq.SQL)
}

func TestNewNotInSubQueryFilter(t *testing.T) {
	f := NewNotInSubQueryFilter("group_id", "SELECT id FROM dict_groups WHERE type IN (?)", []string{"a"})

	assert.Equal(t, "group_id", f.Field)
	assert.Equal(t, constants.OP_NOT_IN, f.Operator)
	sq, ok := f.Value.(*SubQuery)
	require.True(t, ok, "Value 应为 *SubQuery 类型")
	assert.Equal(t, "SELECT id FROM dict_groups WHERE type IN (?)", sq.SQL)
}

func TestNewInSubQueryFilter_NoArgs(t *testing.T) {
	// 无参数子查询（如 SELECT id FROM x，无占位符）
	f := NewInSubQueryFilter("id", "SELECT id FROM dict_groups")
	assert.Equal(t, constants.OP_IN, f.Operator)
	sq, ok := f.Value.(*SubQuery)
	require.True(t, ok)
	assert.Equal(t, 0, len(sq.Args))
}

// ============================================================
// AddInSubQueryFilterIfNotEmpty
// ============================================================

func TestAddInSubQueryFilterIfNotEmpty_WithSlice(t *testing.T) {
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)",
		[]string{"typeA", "typeB"})

	assert.Len(t, q.Filters, 1)
	f := q.Filters[0]
	assert.Equal(t, "group_id", f.Field)
	assert.Equal(t, constants.OP_IN, f.Operator)
	_, ok := f.Value.(*SubQuery)
	assert.True(t, ok, "Value 应为 *SubQuery")
}

func TestAddInSubQueryFilterIfNotEmpty_EmptySlice_Skipped(t *testing.T) {
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)",
		[]string{})

	assert.Empty(t, q.Filters, "空切片应跳过")
}

func TestAddInSubQueryFilterIfNotEmpty_Nil_Skipped(t *testing.T) {
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)", nil)

	assert.Empty(t, q.Filters, "nil 应跳过")
}

func TestAddInSubQueryFilterIfNotEmpty_NoArgs_Skipped(t *testing.T) {
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)")

	assert.Empty(t, q.Filters, "无参数应跳过")
}

func TestAddInSubQueryFilterIfNotEmpty_IntSlice(t *testing.T) {
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("user_id",
		"SELECT user_id FROM orders WHERE amount > ?",
		[]int64{100, 200, 300})

	assert.Len(t, q.Filters, 1)
	f := q.Filters[0]
	assert.Equal(t, "user_id", f.Field)
	sq, ok := f.Value.(*SubQuery)
	require.True(t, ok)
	assert.Equal(t, "SELECT user_id FROM orders WHERE amount > ?", sq.SQL)
}

func TestAddInSubQueryFilterIfNotEmpty_MultiScalarArgs(t *testing.T) {
	// 多占位符子查询：多参数按顺序逐个绑定，参数数量须与占位符一致
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("user_id",
		"SELECT user_id FROM scope_bindings WHERE domain = ? AND tenant_id = ? AND (scope_type = ? OR (scope_type IN (?, ?) AND region_code = ?))",
		1, "tenant-a", 2, 3, 4, "CN")

	assert.Len(t, q.Filters, 1)
	sq, ok := q.Filters[0].Value.(*SubQuery)
	require.True(t, ok, "Value 应为 *SubQuery")
	assert.Len(t, sq.Args, 6, "多参数应全部保留，不得丢弃导致占位符悬空")
	assert.Equal(t, []interface{}{1, "tenant-a", 2, 3, 4, "CN"}, sq.Args)
}

func TestAddInSubQueryFilterIfNotEmpty_MultiArgs_WithPointerDeref(t *testing.T) {
	// 多参数场景下指针参数应解引用
	region := "CN"
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("user_id",
		"SELECT user_id FROM scope_bindings WHERE domain = ? AND region_code = ?",
		1, &region)

	sq, ok := q.Filters[0].Value.(*SubQuery)
	require.True(t, ok)
	assert.Equal(t, []interface{}{1, "CN"}, sq.Args)
}

func TestAddInSubQueryFilterIfNotEmpty_PointerToSlice(t *testing.T) {
	types := []string{"a", "b"}
	q := NewQuery().AddInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)", &types)

	assert.Len(t, q.Filters, 1, "指向切片的指针应解引用并添加")
}

// ============================================================
// AddNotInSubQueryFilterIfNotEmpty
// ============================================================

func TestAddNotInSubQueryFilterIfNotEmpty_WithSlice(t *testing.T) {
	q := NewQuery().AddNotInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)",
		[]string{"typeA"})

	assert.Len(t, q.Filters, 1)
	f := q.Filters[0]
	assert.Equal(t, constants.OP_NOT_IN, f.Operator)
}

func TestAddNotInSubQueryFilterIfNotEmpty_EmptySlice_Skipped(t *testing.T) {
	q := NewQuery().AddNotInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)", []string{})

	assert.Empty(t, q.Filters)
}

func TestAddNotInSubQueryFilterIfNotEmpty_Nil_Skipped(t *testing.T) {
	q := NewQuery().AddNotInSubQueryFilterIfNotEmpty("group_id",
		"SELECT id FROM dict_groups WHERE type IN (?)", nil)

	assert.Empty(t, q.Filters)
}

// ============================================================
// subQueryArgsIfNotEmpty 内部辅助函数
// ============================================================

func TestSubQueryArgsIfNotEmpty_NilArgs(t *testing.T) {
	deref, empty := subQueryArgsIfNotEmpty(nil)
	assert.True(t, empty)
	assert.Nil(t, deref)
}

func TestSubQueryArgsIfNotEmpty_EmptyArgs(t *testing.T) {
	deref, empty := subQueryArgsIfNotEmpty([]interface{}{})
	assert.True(t, empty)
	assert.Nil(t, deref)
}

func TestSubQueryArgsIfNotEmpty_NilFirstArg(t *testing.T) {
	deref, empty := subQueryArgsIfNotEmpty([]interface{}{nil})
	assert.True(t, empty)
	assert.Nil(t, deref)
}

func TestSubQueryArgsIfNotEmpty_EmptySliceFirstArg(t *testing.T) {
	deref, empty := subQueryArgsIfNotEmpty([]interface{}{[]string{}})
	assert.True(t, empty)
	assert.Nil(t, deref)
}

func TestSubQueryArgsIfNotEmpty_NonEmptySliceFirstArg(t *testing.T) {
	deref, empty := subQueryArgsIfNotEmpty([]interface{}{[]string{"a", "b"}})
	assert.False(t, empty)
	assert.NotNil(t, deref)
}

func TestSubQueryArgsIfNotEmpty_ScalarFirstArg(t *testing.T) {
	// 非空标量值（如字符串）应判定为非空
	args, empty := subQueryArgsIfNotEmpty([]interface{}{"active"})
	assert.False(t, empty)
	assert.Equal(t, []interface{}{"active"}, args)
}

// ============================================================
// 链式调用与 BuildWhereClause 集成
// ============================================================

func TestAddInSubQueryFilterIfNotEmpty_ChainedWithOtherFilters(t *testing.T) {
	// 模拟 dict_repository 场景：链式拼接 LIKE + 子查询
	q := NewQuery().
		AddLikeFilterIfNotEmpty("code", "abc").
		AddInSubQueryFilterIfNotEmpty("group_id",
			"SELECT id FROM dict_groups WHERE type IN (?)",
			[]string{"typeA", "typeB"})

	assert.Len(t, q.Filters, 2)
	// 验证子查询过滤器存在
	var sub *Filter
	for _, f := range q.Filters {
		if _, ok := f.Value.(*SubQuery); ok {
			sub = f
			break
		}
	}
	require.NotNil(t, sub, "应包含子查询过滤器")
	assert.Equal(t, "group_id", sub.Field)
}

// ComputedTestPost 主表测试模型，LinkedCount 为物理列（将被计算字段覆盖）
type ComputedTestPost struct {
	ID          int    `gorm:"column:id;primaryKey"`
	Title       string `gorm:"column:title"`
	UserID      int    `gorm:"column:user_id"`
	LinkedCount int    `gorm:"column:linked_count"`
}

func (ComputedTestPost) TableName() string { return "computed_test_posts" }

// ============================================================
// NewKeywordFilterGroup / WithKeywordFilter
// 关键字多字段 OR 模糊匹配
// ============================================================

func TestNewKeywordFilterGroup_EmptyKeyword(t *testing.T) {
	// keyword 为空时返回 nil
	group := NewKeywordFilterGroup("", []string{"id"}, []string{"name"})
	assert.Nil(t, group)
}

func TestNewKeywordFilterGroup_EmptyFields(t *testing.T) {
	// 字段列表均为空时返回 nil
	group := NewKeywordFilterGroup("test", nil, nil)
	assert.Nil(t, group)
}

func TestNewKeywordFilterGroup_OnlyLikeFields(t *testing.T) {
	group := NewKeywordFilterGroup("test", []string{"id", "code"}, nil)
	require.NotNil(t, group)
	assert.Equal(t, constants.LOGIC_OR, group.LogicOp)
	assert.Len(t, group.Filters, 2)
	// 验证都是 LIKE 操作符，值均为 %keyword%
	for _, f := range group.Filters {
		assert.Equal(t, constants.OP_LIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	}
}

func TestNewKeywordFilterGroup_OnlyILikeFields(t *testing.T) {
	group := NewKeywordFilterGroup("test", nil, []string{"name", "code"})
	require.NotNil(t, group)
	assert.Equal(t, constants.LOGIC_OR, group.LogicOp)
	assert.Len(t, group.Filters, 2)
	// 验证都是 ILIKE 操作符，值均为 %keyword%
	for _, f := range group.Filters {
		assert.Equal(t, constants.OP_ILIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	}
}

func TestNewKeywordFilterGroup_MixedFields(t *testing.T) {
	// 同时有 likeFields 和 iLikeFields
	group := NewKeywordFilterGroup("test", []string{"id"}, []string{"name", "code"})
	require.NotNil(t, group)
	assert.Equal(t, constants.LOGIC_OR, group.LogicOp)
	assert.Len(t, group.Filters, 3)

	// 统计操作符数量
	opCount := map[constants.Operator]int{}
	for _, f := range group.Filters {
		opCount[f.Operator]++
	}
	assert.Equal(t, 1, opCount[constants.OP_LIKE])  // id 用 LIKE
	assert.Equal(t, 2, opCount[constants.OP_ILIKE]) // name, code 用 ILIKE

	// 验证字段集合与值
	fields := map[string]bool{}
	for _, f := range group.Filters {
		fields[f.Field] = true
		assert.Equal(t, "%test%", f.Value)
	}
	assert.True(t, fields["id"])
	assert.True(t, fields["name"])
	assert.True(t, fields["code"])
}

func TestWithKeywordFilter_EmptyKeyword(t *testing.T) {
	q := NewQuery().AddLikeFilterIfNotEmpty("name", "existing")
	originalFilterCount := len(q.Filters)
	// keyword 为空时不修改 query
	result := q.WithKeywordFilter("", []string{"id"}, []string{"name"})
	assert.Equal(t, q, result)                         // 返回原 query
	assert.Nil(t, result.FilterGroup)                  // 未设置 FilterGroup
	assert.Len(t, result.Filters, originalFilterCount) // 原有 filter 不变
}

func TestWithKeywordFilter_EmptyFields(t *testing.T) {
	q := NewQuery()
	// 字段列表均为空时不修改 query
	result := q.WithKeywordFilter("test", nil, nil)
	assert.Equal(t, q, result)
	assert.Nil(t, result.FilterGroup)
}

func TestWithKeywordFilter_SetsFilterGroup(t *testing.T) {
	q := NewQuery().AddSafeOrder("", "", "created_at", "DESC")
	result := q.WithKeywordFilter("test", []string{"tenant_id"}, []string{"name", "code"})

	// 返回的是同一个 query（链式）
	assert.Equal(t, q, result)

	// FilterGroup 被设置
	require.NotNil(t, result.FilterGroup)
	assert.Equal(t, constants.LOGIC_OR, result.FilterGroup.LogicOp)
	assert.Len(t, result.FilterGroup.Filters, 3)

	// 验证字段与值
	fields := map[string]bool{}
	for _, f := range result.FilterGroup.Filters {
		fields[f.Field] = true
		assert.Equal(t, "%test%", f.Value)
	}
	assert.True(t, fields["tenant_id"])
	assert.True(t, fields["name"])
	assert.True(t, fields["code"])
}

func TestWithKeywordFilter_ChainedWithOtherFilters(t *testing.T) {
	// 模拟实际场景：status 等值过滤 + keyword 多字段 OR（完整链式调用）
	q := NewQuery().
		AddSafeOrder("", "", "created_at", "DESC").
		AddFilterIfNotEmpty("status", 1).
		WithKeywordFilter("abc", []string{"tenant_id"}, []string{"name", "code"})

	// 原有等值过滤保留
	assert.Len(t, q.Filters, 1)
	assert.Equal(t, "status", q.Filters[0].Field)

	// FilterGroup 被设置
	require.NotNil(t, q.FilterGroup)
	assert.Equal(t, constants.LOGIC_OR, q.FilterGroup.LogicOp)
	assert.Len(t, q.FilterGroup.Filters, 3)
}

// ComputedTestComment 关联表（评论）
type ComputedTestComment struct {
	ID     int `gorm:"column:id;primaryKey"`
	PostID int `gorm:"column:post_id"`
}

func (ComputedTestComment) TableName() string { return "computed_test_comments" }

func setupComputedTestDB(t *testing.T) *gorm.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&ComputedTestPost{}, &ComputedTestComment{}))
	require.NoError(t, gormDB.Create(&ComputedTestPost{ID: 1, Title: "post1", UserID: 1, LinkedCount: 0}).Error)
	require.NoError(t, gormDB.Create(&ComputedTestPost{ID: 2, Title: "post2", UserID: 1, LinkedCount: 0}).Error)
	require.NoError(t, gormDB.Create(&ComputedTestPost{ID: 3, Title: "post3", UserID: 2, LinkedCount: 0}).Error)
	require.NoError(t, gormDB.Create(&ComputedTestComment{ID: 1, PostID: 1}).Error)
	require.NoError(t, gormDB.Create(&ComputedTestComment{ID: 2, PostID: 1}).Error)
	require.NoError(t, gormDB.Create(&ComputedTestComment{ID: 3, PostID: 2}).Error)
	return gormDB
}

// === AddComputedField / AddComputedFields 方法测试 ===

func TestAddComputedField(t *testing.T) {
	q := NewQuery()
	q.AddComputedField("(SELECT COUNT(*) FROM t WHERE id = 1)", "cnt")
	assert.Len(t, q.ComputedFields, 1)
	assert.Equal(t, "(SELECT COUNT(*) FROM t WHERE id = 1)", q.ComputedFields[0].Expr)
	assert.Equal(t, "cnt", q.ComputedFields[0].Alias)
}

func TestAddComputedField_NoAlias(t *testing.T) {
	q := NewQuery()
	q.AddComputedField("COUNT(*)", "")
	assert.Len(t, q.ComputedFields, 1)
	assert.Empty(t, q.ComputedFields[0].Alias)
}

func TestAddComputedFields(t *testing.T) {
	q := NewQuery()
	q.AddComputedFields(
		ComputedField{Expr: "(SELECT 1)", Alias: "a"},
		ComputedField{Expr: "(SELECT 2)", Alias: "b"},
	)
	assert.Len(t, q.ComputedFields, 2)
}

func TestAddComputedField_OnClone(t *testing.T) {
	q := NewQuery()
	q.AddComputedField("(SELECT 1)", "a")
	cloned := q.Clone()
	cloned.AddComputedField("(SELECT 2)", "b")
	assert.Len(t, q.ComputedFields, 1, "原 query 不受 clone 影响")
	assert.Len(t, cloned.ComputedFields, 2)
}

// === buildComputedSelect 单元测试 ===

func TestBuildComputedSelect_WithAlias(t *testing.T) {
	q := NewQuery().AddComputedField("(SELECT COUNT(*) FROM t)", "cnt")
	selects := buildComputedSelect(q)
	assert.Equal(t, []string{"(SELECT COUNT(*) FROM t) AS cnt"}, selects)
}

func TestBuildComputedSelect_NoAlias(t *testing.T) {
	q := NewQuery().AddComputedField("COUNT(*)", "")
	selects := buildComputedSelect(q)
	assert.Equal(t, []string{"COUNT(*)"}, selects)
}

func TestBuildComputedSelect_Empty(t *testing.T) {
	q := NewQuery()
	selects := buildComputedSelect(q)
	assert.Empty(t, selects)
}

// === ApplyJoins + ComputedFields 集成测试 ===

func TestApplyJoins_ComputedFieldOnly(t *testing.T) {
	gormDB := setupComputedTestDB(t)
	dryDB := gormDB.Session(&gorm.Session{DryRun: true}).Table("computed_test_posts")
	q := NewQuery().AddComputedField("(SELECT COUNT(*) FROM computed_test_comments WHERE post_id = computed_test_posts.id)", "linked_count")
	result := ApplyJoins(dryDB, q, "computed_test_posts")
	result = result.Find(&[]ComputedTestPost{})
	sql := result.Statement.SQL.String()
	assert.Contains(t, sql, "computed_test_posts.*")
	assert.Contains(t, sql, "AS linked_count")
}

func TestApplyJoins_NoJoinsNoComputed(t *testing.T) {
	gormDB := setupComputedTestDB(t)
	db := gormDB.Table("computed_test_posts")
	result := ApplyJoins(db, NewQuery(), "computed_test_posts")
	assert.Same(t, db, result)
}

// === 实际查询覆盖测试 ===

func TestComputedField_QueryOverridesMainColumn(t *testing.T) {
	gormDB := setupComputedTestDB(t)
	ctx := context.Background()

	q := NewQuery().AddComputedField(
		"(SELECT COUNT(*) FROM computed_test_comments WHERE post_id = computed_test_posts.id)",
		"linked_count",
	)
	db := ApplyJoins(gormDB.WithContext(ctx), q, "computed_test_posts")
	var posts []ComputedTestPost
	require.NoError(t, db.Order("id ASC").Find(&posts).Error)
	require.Len(t, posts, 3)
	assert.Equal(t, 2, posts[0].LinkedCount, "post1 的 LinkedCount 应被计算字段覆盖为 2")
	assert.Equal(t, 1, posts[1].LinkedCount, "post2 的 LinkedCount 应被计算字段覆盖为 1")
	assert.Equal(t, 0, posts[2].LinkedCount, "post3 的 LinkedCount 应为 0")
}

func TestComputedField_QuerySingle(t *testing.T) {
	gormDB := setupComputedTestDB(t)
	ctx := context.Background()

	q := NewQuery().AddComputedField(
		"(SELECT COUNT(*) FROM computed_test_comments WHERE post_id = computed_test_posts.id)",
		"linked_count",
	)
	db := ApplyJoins(gormDB.WithContext(ctx), q, "computed_test_posts")
	var post ComputedTestPost
	require.NoError(t, db.Where("id = ?", 1).First(&post).Error)
	assert.Equal(t, 2, post.LinkedCount, "post1 的 LinkedCount 应为 2")
	assert.Equal(t, "post1", post.Title)
}

func TestComputedField_CountNotAffected(t *testing.T) {
	gormDB := setupComputedTestDB(t)
	ctx := context.Background()

	q := NewQuery().AddComputedField(
		"(SELECT COUNT(*) FROM computed_test_comments WHERE post_id = computed_test_posts.id)",
		"linked_count",
	)
	db := ApplyJoins(gormDB.WithContext(ctx).Model(&ComputedTestPost{}), q, "computed_test_posts")
	var count int64
	require.NoError(t, db.Count(&count).Error)
	assert.Equal(t, int64(3), count, "Count 查询应返回总行数，不受 ComputedFields 影响")
}

func TestNewILikeFilter(t *testing.T) {
	t.Run("非空值", func(t *testing.T) {
		f := NewILikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_ILIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})

	t.Run("空值仍构造", func(t *testing.T) {
		f := NewILikeFilter("name", "")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_ILIKE, f.Operator)
		assert.Equal(t, "%%", f.Value)
	})
}

func TestNewNotILikeFilter(t *testing.T) {
	t.Run("非空值", func(t *testing.T) {
		f := NewNotILikeFilter("name", "test")
		assert.Equal(t, "name", f.Field)
		assert.Equal(t, constants.OP_NOT_ILIKE, f.Operator)
		assert.Equal(t, "%test%", f.Value)
	})

	t.Run("空值仍构造", func(t *testing.T) {
		f := NewNotILikeFilter("name", "")
		assert.Equal(t, constants.OP_NOT_ILIKE, f.Operator)
		assert.Equal(t, "%%", f.Value)
	})
}

// ==============================================================================
// FilterGroup.AddILikeFilterIfNotEmpty / AddNotILikeFilterIfNotEmpty
// ==============================================================================

func TestFilterGroupAddILikeFiltersIfNotEmpty(t *testing.T) {
	t.Run("AddILikeFilterIfNotEmpty 非空", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR)
		fg.AddILikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_ILIKE, fg.Filters[0].Operator)
		assert.Equal(t, "%test%", fg.Filters[0].Value)
	})

	t.Run("AddILikeFilterIfNotEmpty 空值跳过", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR)
		fg.AddILikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddILikeFilterIfNotEmpty nil跳过", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR)
		fg.AddILikeFilterIfNotEmpty("name", nil)
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddILikeFilterIfNotEmpty 链式调用", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR)
		result := fg.AddILikeFilterIfNotEmpty("code", "abc")
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, len(fg.Filters))
	})

	t.Run("AddNotILikeFilterIfNotEmpty 非空", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotILikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(fg.Filters))
		assert.Equal(t, constants.OP_NOT_ILIKE, fg.Filters[0].Operator)
		assert.Equal(t, "%test%", fg.Filters[0].Value)
	})

	t.Run("AddNotILikeFilterIfNotEmpty 空值跳过", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		fg.AddNotILikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(fg.Filters))
	})

	t.Run("AddNotILikeFilterIfNotEmpty 链式调用", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_AND)
		result := fg.AddNotILikeFilterIfNotEmpty("name", "demo")
		assert.Equal(t, fg, result)
		assert.Equal(t, 1, len(fg.Filters))
	})

	t.Run("多字段 OR ILIKE 组合", func(t *testing.T) {
		fg := NewFilterGroup(constants.LOGIC_OR).
			AddILikeFilterIfNotEmpty("code", "abc").
			AddILikeFilterIfNotEmpty("name", "abc")
		assert.Equal(t, 2, len(fg.Filters))
		assert.Equal(t, constants.OP_ILIKE, fg.Filters[0].Operator)
		assert.Equal(t, constants.OP_ILIKE, fg.Filters[1].Operator)
	})
}

// ==============================================================================
// Query.AddILikeFilterIfNotEmpty / AddNotILikeFilterIfNotEmpty
// ==============================================================================

func TestQueryAddILikeFilterIfNotEmpty(t *testing.T) {
	t.Run("非空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddILikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_ILIKE, q.Filters[0].Operator)
		assert.Equal(t, "%test%", q.Filters[0].Value)
	})

	t.Run("空关键词跳过", func(t *testing.T) {
		q := NewQuery()
		q.AddILikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("nil关键词跳过", func(t *testing.T) {
		q := NewQuery()
		q.AddILikeFilterIfNotEmpty("name", nil)
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddILikeFilterIfNotEmpty("title", "golang")
		assert.Equal(t, q, result)
		assert.Equal(t, "%golang%", q.Filters[0].Value)
	})
}

func TestQueryAddNotILikeFilterIfNotEmpty(t *testing.T) {
	t.Run("非空关键词", func(t *testing.T) {
		q := NewQuery()
		q.AddNotILikeFilterIfNotEmpty("name", "test")
		assert.Equal(t, 1, len(q.Filters))
		assert.Equal(t, constants.OP_NOT_ILIKE, q.Filters[0].Operator)
		assert.Equal(t, "%test%", q.Filters[0].Value)
	})

	t.Run("空关键词跳过", func(t *testing.T) {
		q := NewQuery()
		q.AddNotILikeFilterIfNotEmpty("name", "")
		assert.Equal(t, 0, len(q.Filters))
	})

	t.Run("链式调用", func(t *testing.T) {
		q := NewQuery()
		result := q.AddNotILikeFilterIfNotEmpty("name", "demo")
		assert.Equal(t, q, result)
		assert.Equal(t, 1, len(q.Filters))
	})
}

// ==============================================================================
// SQL 模板生成测试 - buildFilterCondition（query.go 纯函数路径）
// ==============================================================================

func TestBuildFilterConditionILike(t *testing.T) {
	testCases := []struct {
		name              string
		filter            *Filter
		expectedCondition string
		expectedArg       interface{}
	}{
		{
			name:              "OP_ILIKE 生成 LOWER(field) LIKE LOWER(?)",
			filter:            NewILikeFilter("name", "test"),
			expectedCondition: "LOWER(name) LIKE LOWER(?)",
			expectedArg:       "%test%",
		},
		{
			name:              "OP_NOT_ILIKE 生成 LOWER(field) NOT LIKE LOWER(?)",
			filter:            NewNotILikeFilter("email", "spam"),
			expectedCondition: "LOWER(email) NOT LIKE LOWER(?)",
			expectedArg:       "%spam%",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			condition, arg := buildFilterCondition(tc.filter, nil)
			assert.Equal(t, tc.expectedCondition, condition)
			assert.Equal(t, tc.expectedArg, arg)
		})
	}
}

// TestBuildConditionILike 测试 base.go 的 buildCondition（通用模板路径）
func TestBuildConditionILike(t *testing.T) {
	t.Run("OP_ILIKE", func(t *testing.T) {
		condition, arg := buildFilterCondition(NewILikeFilter("title", "demo"), nil)
		assert.Equal(t, "LOWER(title) LIKE LOWER(?)", condition)
		assert.Equal(t, "%demo%", arg)
	})

	t.Run("OP_NOT_ILIKE", func(t *testing.T) {
		condition, arg := buildFilterCondition(NewNotILikeFilter("title", "demo"), nil)
		assert.Equal(t, "LOWER(title) NOT LIKE LOWER(?)", condition)
		assert.Equal(t, "%demo%", arg)
	})
}

// ==============================================================================
// applyFilter 不 panic 测试（base.go db 查询路径）
// ==============================================================================

func TestApplyFilterILike(t *testing.T) {
	gormDB, err := setupTestDB()
	require.NoError(t, err)

	dbQuery := gormDB.Table("test_users")

	filters := []*Filter{
		NewILikeFilter("name", "test"),
		NewNotILikeFilter("email", "spam"),
	}

	for _, f := range filters {
		assert.NotPanics(t, func() {
			ApplyFilter(dbQuery, f)
		}, "应用 ILIKE 过滤器不应 panic")
	}
}

// ==============================================================================
// 端到端集成测试：验证大小写不敏感搜索真实生效
// ==============================================================================

// ilikeTestUser ILIKE 集成测试专用模型（独立表，避免与其他测试数据冲突）
type ilikeTestUser struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"column:name"`
	Code  string `json:"code" gorm:"column:code"`
	Email string `json:"email" gorm:"column:email"`
}

func (ilikeTestUser) TableName() string {
	return "ilike_test_users"
}

// setupILikeTestDB 设置 ILIKE 专用测试库
func setupILikeTestDB() (*gorm.DB, error) {
	gormDB, err := gorm.Open(sqlite.Open("file:ilike_memdb?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	gormDB.Exec("DROP TABLE IF EXISTS ilike_test_users")
	if err := gormDB.AutoMigrate(&ilikeTestUser{}); err != nil {
		return nil, err
	}
	return gormDB, nil
}

// TestILikeCaseInsensitiveSearchE2E 端到端验证：数据 "ABcDef" 可被任意大小写关键词匹配
// 这是用户的核心诉求："ABcDef 我随便输入大小写都可以支持 bc Bc bC 都能搜"
func TestILikeCaseInsensitiveSearchE2E(t *testing.T) {
	gormDB, err := setupILikeTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[ilikeTestUser](dbHandler, logger.NewLogger(), "ilike_test_users")
	ctx := context.Background()

	// 插入混合大小写数据
	seed := []*ilikeTestUser{
		{Name: "ABcDef", Code: "CdEfGh", Email: "User@Example.COM"},
		{Name: "HelloWorld", Code: "hElLo", Email: "test@demo.com"},
		{Name: "GoLang", Code: "GOLANG", Email: "foo@bar.org"},
	}
	for _, u := range seed {
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)
	}

	// 用户原始诉求：数据 ABcDef，输入 bc/Bc/bC/BC 都应能搜到
	t.Run("ABcDef 任意大小写关键词均匹配", func(t *testing.T) {
		keywords := []string{"bc", "Bc", "bC", "BC", "abcd", "ABCD", "AbCdEf", "def", "DEF", "DeF"}
		for _, kw := range keywords {
			q := NewQuery().AddFilter(NewILikeFilter("name", kw))
			results, err := repo.List(ctx, q)
			require.NoError(t, err, "关键词 %q 查询失败", kw)
			require.Len(t, results, 1, "关键词 %q 应匹配到 ABcDef", kw)
			assert.Equal(t, "ABcDef", results[0].Name)
		}
	})

	// 对比说明：SQLite 的 LIKE 默认对 ASCII 大小写不敏感（方言特性），
	// 而 PostgreSQL 的 LIKE 默认大小写敏感，ILIKE 通过显式 LOWER() 保证
	// 跨数据库（MySQL/PostgreSQL/SQLite）一致的大小写不敏感行为，
	// 不依赖任何方言的默认 collation，这是 ILIKE 的核心价值
	t.Run("ILIKE 与 LIKE 在 SQLite 下均匹配（方言差异说明）", func(t *testing.T) {
		qLike := NewQuery().AddFilter(NewLikeFilter("name", "bc"))
		resultsLike, err := repo.List(ctx, qLike)
		require.NoError(t, err)

		qILike := NewQuery().AddFilter(NewILikeFilter("name", "bc"))
		resultsILike, err := repo.List(ctx, qILike)
		require.NoError(t, err)

		// SQLite 下两者都匹配（ASCII 大小写不敏感特性）
		assert.Len(t, resultsLike, 1, "SQLite LIKE 默认对 ASCII 大小写不敏感")
		assert.Len(t, resultsILike, 1, "ILIKE 通过 LOWER() 保证大小写不敏感")
		// 关键区别：ILIKE 在 PostgreSQL/MySQL 下同样保证大小写不敏感，而 LIKE 不保证
	})

	// NOT ILIKE 验证：排除包含某子串（任意大小写）的记录
	t.Run("NOT ILIKE 排除任意大小写子串", func(t *testing.T) {
		q := NewQuery().AddFilter(NewNotILikeFilter("name", "ello"))
		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		// HelloWorld 含 "ello"（任意大小写）被排除，剩余 ABcDef + GoLang
		assert.Len(t, results, 2)
	})

	// 多字段 OR ILIKE：在 name 或 code 上搜索同一关键词
	t.Run("多字段 OR ILIKE 匹配", func(t *testing.T) {
		// "cdef" 在 name=ABcDef；"cd" 同时在 name=ABcDef 和 code=CdEfGh
		keywordGroup := NewFilterGroup(constants.LOGIC_OR).
			AddILikeFilterIfNotEmpty("name", "cdef").
			AddILikeFilterIfNotEmpty("code", "cdef")
		q := NewQuery().WithFilterGroup(keywordGroup)

		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "ABcDef", results[0].Name)
	})

	// 大小写混合的 Email 字段搜索
	t.Run("Email 字段大小写不敏感搜索", func(t *testing.T) {
		q := NewQuery().AddFilter(NewILikeFilter("email", "USER@example.com"))
		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "User@Example.COM", results[0].Email)
	})

	// Query.AddILikeFilterIfNotEmpty 空值不参与过滤
	t.Run("空关键词不参与过滤", func(t *testing.T) {
		q := NewQuery().AddILikeFilterIfNotEmpty("name", "")
		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		assert.Len(t, results, len(seed), "空关键词应返回全部记录")
	})
}

// TestILikeFilterGroupORE2E 端到端验证：FilterGroup OR + ILIKE 多字段搜索
// 模拟 game_brand_repository 的 code/name OR 关键字搜索场景
func TestILikeFilterGroupORE2E(t *testing.T) {
	gormDB, err := setupILikeTestDB()
	require.NoError(t, err)

	dbHandler := newTestDBHandler(gormDB)
	repo := NewBaseRepository[ilikeTestUser](dbHandler, logger.NewLogger(), "ilike_test_users")
	ctx := context.Background()

	seed := []*ilikeTestUser{
		{Name: "NetEase", Code: "ne001", Email: "a@x.com"},
		{Name: "Tencent", Code: "TC002", Email: "b@x.com"},
		{Name: "Blizzard", Code: "BZ003", Email: "c@x.com"},
	}
	for _, u := range seed {
		_, err := repo.Create(ctx, u)
		require.NoError(t, err)
	}

	// "tC" 应匹配 Tencent 的 code（TC002）和 name（Tencent）
	t.Run("tC 匹配 name 或 code（OR ILIKE）", func(t *testing.T) {
		keywordGroup := NewFilterGroup(constants.LOGIC_OR).
			AddILikeFilterIfNotEmpty("code", "tC").
			AddILikeFilterIfNotEmpty("name", "tC")
		q := NewQuery().WithFilterGroup(keywordGroup)

		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Tencent", results[0].Name)
	})

	// "BLIZZ" 只匹配 Blizzard
	t.Run("BLIZZ 匹配 Blizzard", func(t *testing.T) {
		keywordGroup := NewFilterGroup(constants.LOGIC_OR).
			AddILikeFilterIfNotEmpty("code", "BLIZZ").
			AddILikeFilterIfNotEmpty("name", "BLIZZ")
		q := NewQuery().WithFilterGroup(keywordGroup)

		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Blizzard", results[0].Name)
	})

	// "ne" 同时匹配 NetEase(name) 和 ne001(code)
	t.Run("ne 匹配多条", func(t *testing.T) {
		keywordGroup := NewFilterGroup(constants.LOGIC_OR).
			AddILikeFilterIfNotEmpty("code", "ne").
			AddILikeFilterIfNotEmpty("name", "ne")
		q := NewQuery().WithFilterGroup(keywordGroup)

		results, err := repo.List(ctx, q)
		require.NoError(t, err)
		assert.Len(t, results, 1, "ne 应匹配 NetEase（name 含 ne，code 含 ne）")
		assert.Equal(t, "NetEase", results[0].Name)
	})
}
