/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-11 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-16 11:15:27
 * @FilePath: \go-toolbox\pkg\sqlbuilder\repository\cursor_test.go
 * @Description: 游标工具测试（用于复合排序分页）
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

// ==============================================================================
// NewTimeEntityCursor
// ==============================================================================

func TestNewTimeEntityCursor(t *testing.T) {
	t.Run("正常创建", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursor(&now, "123")
		assert.NotNil(t, cursor)
		assert.Equal(t, now.Format(TimeFormat), cursor.Timestamp)
		assert.Equal(t, "123", cursor.EntityID)
	})

	t.Run("nil时间戳", func(t *testing.T) {
		cursor := NewTimeEntityCursor(nil, "123")
		assert.NotNil(t, cursor)
		assert.True(t, cursor.IsEmpty())
	})

	t.Run("空实体ID", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursor(&now, "")
		assert.NotNil(t, cursor)
		assert.True(t, cursor.IsEmpty())
	})
}

// ==============================================================================
// TimeEntityCursor - String
// ==============================================================================

func TestTimeEntityCursor_String(t *testing.T) {
	t.Run("正常字符串表示", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursor(&now, "123")
		assert.Equal(t, cursor.Timestamp+CursorSeparator+cursor.EntityID, cursor.String())
	})

	t.Run("空游标", func(t *testing.T) {
		cursor := &TimeEntityCursor{}
		assert.Equal(t, "", cursor.String())
	})
}

// ==============================================================================
// TimeEntityCursor - IsValid
// ==============================================================================

func TestTimeEntityCursor_IsValid(t *testing.T) {
	t.Run("空游标无效", func(t *testing.T) {
		cursor := NewTimeEntityCursor(nil, "")
		assert.Error(t, cursor.IsValid())
	})

	t.Run("有效游标", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		assert.NoError(t, cursor.IsValid())
	})

	t.Run("缺少时间戳", func(t *testing.T) {
		cursor := &TimeEntityCursor{EntityID: "123"}
		assert.Error(t, cursor.IsValid())
	})

	t.Run("缺少实体ID", func(t *testing.T) {
		cursor := &TimeEntityCursor{Timestamp: time.Now().Format(TimeFormat)}
		assert.Error(t, cursor.IsValid())
	})
}

// ==============================================================================
// TimeEntityCursor - Clone
// ==============================================================================

func TestTimeEntityCursor_Clone(t *testing.T) {
	t.Run("正常克隆", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		clone := cursor.Clone()
		assert.Equal(t, cursor, clone)
		assert.NotSame(t, cursor, clone)
	})

	t.Run("nil游标克隆", func(t *testing.T) {
		var cursor *TimeEntityCursor
		clone := cursor.Clone()
		assert.NotNil(t, clone)
		assert.True(t, clone.IsEmpty())
	})
}

// ==============================================================================
// TimeEntityCursor - Equal
// ==============================================================================

func TestTimeEntityCursor_Equal(t *testing.T) {
	t.Run("两个nil", func(t *testing.T) {
		var c1, c2 *TimeEntityCursor
		assert.True(t, c1.Equal(c2))
	})

	t.Run("一个nil一个非nil", func(t *testing.T) {
		c1 := NewTimeEntityCursor(timePtr(time.Now()), "123")
		assert.False(t, c1.Equal(nil))
	})

	t.Run("相同游标", func(t *testing.T) {
		now := time.Now()
		c1 := NewTimeEntityCursor(&now, "123")
		c2 := NewTimeEntityCursor(&now, "123")
		assert.True(t, c1.Equal(c2))
	})

	t.Run("不同游标", func(t *testing.T) {
		c1 := NewTimeEntityCursor(timePtr(time.Now()), "123")
		c2 := NewTimeEntityCursor(timePtr(time.Now()), "456")
		assert.False(t, c1.Equal(c2))
	})
}

// ==============================================================================
// TimeEntityCursor - Compare
// ==============================================================================

func TestTimeEntityCursor_Compare(t *testing.T) {
	t.Run("时间不同-小于", func(t *testing.T) {
		cursor1 := NewTimeEntityCursor(timePtr(time.Now().Add(-time.Minute)), "123")
		cursor2 := NewTimeEntityCursor(timePtr(time.Now()), "123")
		result, err := cursor1.Compare(cursor2)
		assert.NoError(t, err)
		assert.Equal(t, -1, result)
	})

	t.Run("时间相同-ID不同", func(t *testing.T) {
		cursor2 := NewTimeEntityCursor(timePtr(time.Now()), "123")
		cursor3 := NewTimeEntityCursor(timePtr(time.Now()), "124")
		result, err := cursor2.Compare(cursor3)
		assert.NoError(t, err)
		assert.Equal(t, -1, result)
	})

	t.Run("完全相同", func(t *testing.T) {
		now := time.Now()
		cursor2 := NewTimeEntityCursor(&now, "123")
		cursor4 := NewTimeEntityCursor(&now, "123")
		result, err := cursor2.Compare(cursor4)
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("nil游标比较", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		_, err := cursor.Compare(nil)
		assert.Error(t, err)
	})
}

// ==============================================================================
// TimeEntityCursor - ToJSON / ParseTimeEntityCursorFromJSON
// ==============================================================================

func TestTimeEntityCursor_JSON(t *testing.T) {
	t.Run("ToJSON", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		jsonStr, err := cursor.ToJSON()
		assert.NoError(t, err)
		assert.Contains(t, jsonStr, cursor.EntityID)
		assert.Contains(t, jsonStr, cursor.Timestamp)
	})

	t.Run("空游标ToJSON", func(t *testing.T) {
		cursor := &TimeEntityCursor{}
		_, err := cursor.ToJSON()
		assert.Error(t, err)
	})

	t.Run("ParseTimeEntityCursorFromJSON", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		jsonStr, _ := cursor.ToJSON()
		parsedCursor, err := ParseTimeEntityCursorFromJSON(jsonStr)
		assert.NoError(t, err)
		assert.Equal(t, cursor, parsedCursor)
	})

	t.Run("空字符串ParseFromJSON", func(t *testing.T) {
		result, err := ParseTimeEntityCursorFromJSON("")
		assert.NoError(t, err)
		assert.True(t, result.IsEmpty())
	})

	t.Run("无效JSON ParseFromJSON", func(t *testing.T) {
		_, err := ParseTimeEntityCursorFromJSON(`{invalid}`)
		assert.Error(t, err)
	})
}

// ==============================================================================
// TimeEntityCursor - ToBase64 / ParseTimeEntityCursor
// ==============================================================================

func TestTimeEntityCursor_Base64(t *testing.T) {
	t.Run("ToBase64", func(t *testing.T) {
		cursor := NewTimeEntityCursor(timePtr(time.Now()), "123")
		base64Str := cursor.ToBase64()
		assert.NotEmpty(t, base64Str)
		decodedCursor := ParseTimeEntityCursor(base64Str)
		assert.Equal(t, cursor, decodedCursor)
	})

	t.Run("空游标ToBase64", func(t *testing.T) {
		cursor := &TimeEntityCursor{}
		assert.Equal(t, "", cursor.ToBase64())
	})

	t.Run("ParseTimeEntityCursor-空字符串", func(t *testing.T) {
		result := ParseTimeEntityCursor("")
		assert.True(t, result.IsEmpty())
	})

	t.Run("ParseTimeEntityCursor-无分隔符", func(t *testing.T) {
		result := ParseTimeEntityCursor("no_separator")
		assert.True(t, result.IsEmpty())
	})
}

// ==============================================================================
// TimeEntityCursor - GetTime
// ==============================================================================

func TestTimeEntityCursor_GetTime(t *testing.T) {
	t.Run("正常获取时间", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursor(&now, "123")
		parsedTime, err := cursor.GetTime()
		assert.NoError(t, err)
		assert.NotNil(t, parsedTime)
	})

	t.Run("空时间戳", func(t *testing.T) {
		cursor := &TimeEntityCursor{EntityID: "123"}
		_, err := cursor.GetTime()
		assert.Error(t, err)
	})

	t.Run("无效时间戳格式", func(t *testing.T) {
		cursor := &TimeEntityCursor{Timestamp: "invalid", EntityID: "123"}
		_, err := cursor.GetTime()
		assert.Error(t, err)
	})
}

// ==============================================================================
// CursorBuilder
// ==============================================================================

func TestCursorBuilder(t *testing.T) {
	t.Run("正常构建", func(t *testing.T) {
		builder := NewCursorBuilder().
			WithTime(timePtr(time.Now())).
			WithEntityID("123")
		cursor, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "123", cursor.EntityID)
	})

	t.Run("WithTimestamp", func(t *testing.T) {
		now := time.Now()
		builder := NewCursorBuilder().
			WithTimestamp(now.Format(TimeFormat)).
			WithEntityID("456")
		cursor, err := builder.Build()
		assert.NoError(t, err)
		assert.Equal(t, "456", cursor.EntityID)
	})

	t.Run("缺少必填字段", func(t *testing.T) {
		builder := NewCursorBuilder().WithEntityID("123")
		_, err := builder.Build()
		assert.Error(t, err)
	})

	t.Run("MustBuild成功", func(t *testing.T) {
		builder := NewCursorBuilder().
			WithTime(timePtr(time.Now())).
			WithEntityID("123")
		cursor := builder.MustBuild()
		assert.NotNil(t, cursor)
	})

	t.Run("MustBuild失败panic", func(t *testing.T) {
		builder := NewCursorBuilder()
		assert.Panics(t, func() {
			builder.MustBuild()
		})
	})
}

// ==============================================================================
// NewTimeEntityCursorFromTime
// ==============================================================================

func TestNewTimeEntityCursorFromTime(t *testing.T) {
	t.Run("正常创建", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursorFromTime(now, "789")
		assert.NotNil(t, cursor)
		assert.Equal(t, "789", cursor.EntityID)
		assert.Equal(t, now.Format(TimeFormat), cursor.Timestamp)
	})

	t.Run("自定义时间格式", func(t *testing.T) {
		now := time.Now()
		cursor := NewTimeEntityCursorFromTime(now, "789", time.RFC3339)
		assert.NotNil(t, cursor)
		assert.Equal(t, now.Format(time.RFC3339), cursor.Timestamp)
	})

	t.Run("空实体ID", func(t *testing.T) {
		cursor := NewTimeEntityCursorFromTime(time.Now(), "")
		assert.True(t, cursor.IsEmpty())
	})
}
