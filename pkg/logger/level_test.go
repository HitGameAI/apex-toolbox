/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-08 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-08 00:00:00
 * @FilePath: \go-toolbox\pkg\logger\level_test.go
 * @Description: 日志级别测试套件
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package logger

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// LevelTestSuite 日志级别测试套件
type LevelTestSuite struct {
	suite.Suite
}

// TestLogLevelConstants 测试日志级别常量
func (suite *LevelTestSuite) TestLogLevelConstants() {
	assert.Equal(suite.T(), LogLevel(0), DEBUG)
	assert.Equal(suite.T(), LogLevel(1), INFO)
	assert.Equal(suite.T(), LogLevel(2), WARN)
	assert.Equal(suite.T(), LogLevel(3), ERROR)
	assert.Equal(suite.T(), LogLevel(4), FATAL)

	// 测试级别顺序
	assert.True(suite.T(), DEBUG < INFO)
	assert.True(suite.T(), INFO < WARN)
	assert.True(suite.T(), WARN < ERROR)
	assert.True(suite.T(), ERROR < FATAL)
}

// TestLogLevelString 测试级别字符串表示
func (suite *LevelTestSuite) TestLogLevelString() {
	assert.Equal(suite.T(), "DEBUG", DEBUG.String())
	assert.Equal(suite.T(), "INFO", INFO.String())
	assert.Equal(suite.T(), "WARN", WARN.String())
	assert.Equal(suite.T(), "ERROR", ERROR.String())
	assert.Equal(suite.T(), "FATAL", FATAL.String())

	// 测试无效级别
	invalidLevel := LogLevel(999)
	expected := "UNKNOWN(999)"
	assert.Equal(suite.T(), expected, invalidLevel.String())

	// 测试 TRACE 级别 (-1)
	traceLevel := LogLevel(-1)
	expected = "TRACE"
	assert.Equal(suite.T(), expected, traceLevel.String())
}

// TestLogLevelEmoji 测试级别表情符号
func (suite *LevelTestSuite) TestLogLevelEmoji() {
	assert.Equal(suite.T(), "🐛", DEBUG.Emoji())
	assert.Equal(suite.T(), "ℹ️", INFO.Emoji())
	assert.Equal(suite.T(), "⚠️", WARN.Emoji())
	assert.Equal(suite.T(), "❌", ERROR.Emoji())
	assert.Equal(suite.T(), "💀", FATAL.Emoji())

	// 测试无效级别
	invalidLevel := LogLevel(999)
	assert.Equal(suite.T(), "❓", invalidLevel.Emoji())
}

// TestLogLevelColor 测试级别颜色代码
func (suite *LevelTestSuite) TestLogLevelColor() {
	assert.Equal(suite.T(), "\033[36m", DEBUG.Color()) // 青色
	assert.Equal(suite.T(), "\033[32m", INFO.Color())  // 绿色
	assert.Equal(suite.T(), "\033[33m", WARN.Color())  // 黄色
	assert.Equal(suite.T(), "\033[31m", ERROR.Color()) // 红色
	assert.Equal(suite.T(), "\033[35m", FATAL.Color()) // 紫色

	// 测试无效级别
	invalidLevel := LogLevel(999)
	assert.Equal(suite.T(), "\033[0m", invalidLevel.Color()) // 重置颜色
}

// TestParseLevel 测试从字符串解析级别
func (suite *LevelTestSuite) TestParseLevel() {
	// 测试有效的级别字符串
	tests := []struct {
		input    string
		expected LogLevel
		hasError bool
	}{
		{"DEBUG", DEBUG, false},
		{"debug", DEBUG, false},
		{"Debug", DEBUG, false},
		{"INFO", INFO, false},
		{"info", INFO, false},
		{"Info", INFO, false},
		{"WARN", WARN, false},
		{"warn", WARN, false},
		{"WARNING", WARN, false},
		{"warning", WARN, false},
		{"ERROR", ERROR, false},
		{"error", ERROR, false},
		{"Error", ERROR, false},
		{"FATAL", FATAL, false},
		{"fatal", FATAL, false},
		{"Fatal", FATAL, false},
		// 带空格的测试
		{"  DEBUG  ", DEBUG, false},
		{"  INFO  ", INFO, false},
		// 无效输入
		{"INVALID", DEBUG, true}, // 默认返回DEBUG
		{"", DEBUG, true},
		{"123", DEBUG, true},
		{"NULL", DEBUG, true},
	}

	for _, test := range tests {
		suite.T().Run(test.input, func(t *testing.T) {
			level, err := ParseLevel(test.input)
			assert.Equal(t, test.expected, level)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogLevelIsEnabled 测试级别启用检查
func (suite *LevelTestSuite) TestLogLevelIsEnabled() {
	// 测试DEBUG级别
	assert.True(suite.T(), DEBUG.IsEnabled(DEBUG))
	assert.True(suite.T(), DEBUG.IsEnabled(INFO))
	assert.True(suite.T(), DEBUG.IsEnabled(WARN))
	assert.True(suite.T(), DEBUG.IsEnabled(ERROR))
	assert.True(suite.T(), DEBUG.IsEnabled(FATAL))

	// 测试INFO级别
	assert.False(suite.T(), INFO.IsEnabled(DEBUG))
	assert.True(suite.T(), INFO.IsEnabled(INFO))
	assert.True(suite.T(), INFO.IsEnabled(WARN))
	assert.True(suite.T(), INFO.IsEnabled(ERROR))
	assert.True(suite.T(), INFO.IsEnabled(FATAL))

	// 测试WARN级别
	assert.False(suite.T(), WARN.IsEnabled(DEBUG))
	assert.False(suite.T(), WARN.IsEnabled(INFO))
	assert.True(suite.T(), WARN.IsEnabled(WARN))
	assert.True(suite.T(), WARN.IsEnabled(ERROR))
	assert.True(suite.T(), WARN.IsEnabled(FATAL))

	// 测试ERROR级别
	assert.False(suite.T(), ERROR.IsEnabled(DEBUG))
	assert.False(suite.T(), ERROR.IsEnabled(INFO))
	assert.False(suite.T(), ERROR.IsEnabled(WARN))
	assert.True(suite.T(), ERROR.IsEnabled(ERROR))
	assert.True(suite.T(), ERROR.IsEnabled(FATAL))

	// 测试FATAL级别
	assert.False(suite.T(), FATAL.IsEnabled(DEBUG))
	assert.False(suite.T(), FATAL.IsEnabled(INFO))
	assert.False(suite.T(), FATAL.IsEnabled(WARN))
	assert.False(suite.T(), FATAL.IsEnabled(ERROR))
	assert.True(suite.T(), FATAL.IsEnabled(FATAL))
}

// TestGetAllLevels 测试获取所有级别
func (suite *LevelTestSuite) TestGetAllLevels() {
	levels := GetAllLevels()
	expected := []LogLevel{DEBUG, INFO, WARN, ERROR, FATAL}

	assert.Equal(suite.T(), expected, levels)
	assert.Len(suite.T(), levels, 5)

	// 验证顺序
	for i := 0; i < len(levels)-1; i++ {
		assert.True(suite.T(), levels[i] < levels[i+1],
			"Levels should be in ascending order: %v should be less than %v",
			levels[i], levels[i+1])
	}
}

// TestGetLevelNames 测试获取所有级别名称
func (suite *LevelTestSuite) TestGetLevelNames() {
	names := GetLevelNames()
	expected := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	assert.Equal(suite.T(), expected, names)
	assert.Len(suite.T(), names, 5)

	// 验证每个名称都对应正确的级别
	levels := GetAllLevels()
	for i, name := range names {
		assert.Equal(suite.T(), name, levels[i].String())
	}
}

// TestLevelComparison 测试级别比较
func (suite *LevelTestSuite) TestLevelComparison() {
	// 测试相等比较
	assert.Equal(suite.T(), DEBUG, DEBUG)
	assert.Equal(suite.T(), INFO, INFO)
	assert.False(suite.T(), DEBUG == INFO)

	// 测试大小比较
	assert.True(suite.T(), DEBUG < INFO)
	assert.True(suite.T(), INFO < WARN)
	assert.True(suite.T(), WARN < ERROR)
	assert.True(suite.T(), ERROR < FATAL)

	assert.True(suite.T(), FATAL > ERROR)
	assert.True(suite.T(), ERROR > WARN)
	assert.True(suite.T(), WARN > INFO)
	assert.True(suite.T(), INFO > DEBUG)

	// 测试大于等于和小于等于
	assert.True(suite.T(), INFO >= DEBUG)
	assert.True(suite.T(), FATAL >= ERROR)
	assert.True(suite.T(), DEBUG <= INFO)
	assert.True(suite.T(), ERROR <= FATAL)
}

// TestLevelRange 测试级别范围
func (suite *LevelTestSuite) TestLevelRange() {
	// 测试级别是否在有效范围内
	validLevels := []LogLevel{DEBUG, INFO, WARN, ERROR, FATAL}
	for _, level := range validLevels {
		assert.True(suite.T(), level >= DEBUG && level <= FATAL,
			"Level %v should be in valid range", level)
	}

	// 测试超出范围的级别
	invalidLevels := []LogLevel{LogLevel(-1), LogLevel(5), LogLevel(100)}
	for _, level := range invalidLevels {
		assert.False(suite.T(), level >= DEBUG && level <= FATAL,
			"Level %v should be outside valid range", level)
	}
}

// TestLevelConsistency 测试级别一致性
func (suite *LevelTestSuite) TestLevelConsistency() {
	levels := GetAllLevels()

	for _, level := range levels {
		// 字符串表示应该是可解析的
		parsed, err := ParseLevel(level.String())
		assert.NoError(suite.T(), err,
			"Level %v string representation should be parseable", level)
		assert.Equal(suite.T(), level, parsed,
			"Parsed level should equal original level")

		// 表情符号应该不为空
		assert.NotEmpty(suite.T(), level.Emoji(),
			"Level %v should have an emoji", level)

		// 颜色代码应该不为空
		assert.NotEmpty(suite.T(), level.Color(),
			"Level %v should have a color code", level)

		// 字符串表示应该不为空
		assert.NotEmpty(suite.T(), level.String(),
			"Level %v should have a string representation", level)
	}
}

// TestLevelEdgeCases 测试边界情况
func (suite *LevelTestSuite) TestLevelEdgeCases() {
	// 测试极大值
	largeLevel := LogLevel(1000000)
	assert.Contains(suite.T(), largeLevel.String(), "UNKNOWN")
	assert.Equal(suite.T(), "❓", largeLevel.Emoji())
	assert.Equal(suite.T(), "\033[0m", largeLevel.Color())

	// 测试极小值
	smallLevel := LogLevel(-1000000)
	assert.Contains(suite.T(), smallLevel.String(), "UNKNOWN")
	assert.Equal(suite.T(), "❓", smallLevel.Emoji())
	assert.Equal(suite.T(), "\033[0m", smallLevel.Color())

	// 测试零值
	zeroLevel := LogLevel(0)
	assert.Equal(suite.T(), DEBUG, zeroLevel) // 应该等于DEBUG
}

// TestParseLevelCaseInsensitive 测试大小写不敏感解析
func (suite *LevelTestSuite) TestParseLevelCaseInsensitive() {
	testCases := []string{
		"debug", "DEBUG", "Debug", "dEbUg",
		"info", "INFO", "Info", "iNfO",
		"warn", "WARN", "Warn", "wArN",
		"warning", "WARNING", "Warning", "wArNiNg",
		"error", "ERROR", "Error", "eRrOr",
		"fatal", "FATAL", "Fatal", "fAtAl",
	}

	for _, testCase := range testCases {
		suite.T().Run(testCase, func(t *testing.T) {
			level, err := ParseLevel(testCase)
			assert.NoError(t, err, "Should parse %s without error", testCase)

			// 验证解析结果是正确的级别
			upper := strings.ToUpper(testCase)
			if upper == "WARNING" {
				upper = "WARN" // WARNING 映射到 WARN
			}

			expected, parseErr := ParseLevel(upper)
			assert.NoError(t, parseErr)
			assert.Equal(t, expected, level,
				"Case insensitive parsing should work for %s", testCase)
		})
	}
}

// TestLevelConfig 测试级别配置
func (suite *LevelTestSuite) TestLevelConfig() {
	// 验证级别配置映射是完整的
	expectedLevels := []LogLevel{DEBUG, INFO, WARN, ERROR, FATAL}

	for _, level := range expectedLevels {
		// 每个级别都应该有配置
		info := level.Info()
		assert.NotEmpty(suite.T(), info.Name,
			"Level %v should have name", level)

		// 验证配置字段不为空
		assert.NotEmpty(suite.T(), info.Emoji,
			"Level %v should have emoji", level)
		assert.NotEmpty(suite.T(), info.Color,
			"Level %v should have color", level)

		// 验证名称与String()方法一致
		assert.Equal(suite.T(), info.Name, level.String(),
			"Config name should match String() for level %v", level)
	}
}

// 运行测试套件
func TestLevelSuite(t *testing.T) {
	suite.Run(t, new(LevelTestSuite))
}

// TestLevelBenchmark 级别操作的基准测试（模拟）
func TestLevelBenchmark(t *testing.T) {
	levels := GetAllLevels()

	t.Run("StringConversion", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			for _, level := range levels {
				_ = level.String()
			}
		}
	})

	t.Run("EmojiRetrieval", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			for _, level := range levels {
				_ = level.Emoji()
			}
		}
	})

	t.Run("ColorRetrieval", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			for _, level := range levels {
				_ = level.Color()
			}
		}
	})

	t.Run("LevelParsing", func(t *testing.T) {
		levelNames := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
		for i := 0; i < 1000; i++ {
			for _, name := range levelNames {
				_, _ = ParseLevel(name)
			}
		}
	})

	t.Run("LevelComparison", func(t *testing.T) {
		for i := 0; i < 10000; i++ {
			for j, level1 := range levels {
				for k, level2 := range levels {
					_ = level1 < level2
					_ = level1 == level2
					_ = level1 > level2
					_ = level1.IsEnabled(level2)
					// 避免未使用变量警告
					_, _ = j, k
				}
			}
		}
	})
}

// TestLevelMemoryUsage 测试级别操作的内存使用
func TestLevelMemoryUsage(t *testing.T) {
	// 创建大量级别实例
	levels := make([]LogLevel, 10000)
	for i := 0; i < 10000; i++ {
		levels[i] = LogLevel(i % 5) // 循环使用有效级别
	}

	// 验证所有级别都正确创建
	assert.Len(t, levels, 10000)

	// 对每个级别执行操作
	for i := 0; i < 1000; i++ { // 减少迭代次数以避免测试超时
		level := levels[i]
		_ = level.String()
		_ = level.Emoji()
		_ = level.Color()
		_ = level.IsEnabled(INFO)
	}
}

// TestLogLevelUnmarshalYAML 测试 YAML 反序列化
func TestLogLevelUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected LogLevel
		hasError bool
	}{
		{"string_debug", "level: debug", DEBUG, false},
		{"string_info", "level: info", INFO, false},
		{"string_warn", "level: warn", WARN, false},
		{"string_error", "level: error", ERROR, false},
		{"string_uppercase", "level: DEBUG", DEBUG, false},
		{"integer_value", "level: 1", DEBUG, true},
		{"invalid_string", "level: invalid", DEBUG, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Level LogLevel `yaml:"level"`
			}
			err := yaml.Unmarshal([]byte(tt.yaml), &config)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config.Level)
			}
		})
	}
}

// TestLogLevelUnmarshalJSON 测试 JSON 反序列化
func TestLogLevelUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected LogLevel
		hasError bool
	}{
		{"string_debug", `{"level":"debug"}`, DEBUG, false},
		{"string_info", `{"level":"info"}`, INFO, false},
		{"string_warn", `{"level":"warn"}`, WARN, false},
		{"string_error", `{"level":"error"}`, ERROR, false},
		{"string_uppercase", `{"level":"DEBUG"}`, DEBUG, false},
		{"invalid_string", `{"level":"invalid"}`, INFO, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config struct {
				Level LogLevel `json:"level"`
			}
			err := json.Unmarshal([]byte(tt.json), &config)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config.Level)
			}
		})
	}
}

// TestLogLevelMarshalText 测试文本序列化
func TestLogLevelMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"debug", DEBUG, "DEBUG"},
		{"info", INFO, "INFO"},
		{"warn", WARN, "WARN"},
		{"error", ERROR, "ERROR"},
		{"fatal", FATAL, "FATAL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.level.MarshalText()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(text))
		})
	}
}

// TestLogLevelRoundTrip 测试序列化和反序列化往返
func TestLogLevelRoundTrip(t *testing.T) {
	levels := []LogLevel{DEBUG, INFO, WARN, ERROR, FATAL}
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			// JSON 往返
			jsonData, err := json.Marshal(struct {
				Level LogLevel `json:"level"`
			}{Level: level})
			assert.NoError(t, err)
			var jsonResult struct {
				Level LogLevel `json:"level"`
			}
			err = json.Unmarshal(jsonData, &jsonResult)
			assert.NoError(t, err)
			assert.Equal(t, level, jsonResult.Level)

			// YAML 往返
			yamlData, err := yaml.Marshal(struct {
				Level LogLevel `yaml:"level"`
			}{Level: level})
			assert.NoError(t, err)
			var yamlResult struct {
				Level LogLevel `yaml:"level"`
			}
			err = yaml.Unmarshal(yamlData, &yamlResult)
			assert.NoError(t, err)
			assert.Equal(t, level, yamlResult.Level)
		})
	}
}
