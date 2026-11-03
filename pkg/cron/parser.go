/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-25 10:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 12:07:16
 * @FilePath: \go-toolbox\pkg\cron\parser.go
 * @Description: Cron 表达式解析器实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// NewCronParser 创建一个自定义配置的解析器
//
// 示例：
//
//	标准解析器(不包含秒)
//	parser := NewCronParser(CronMinute | CronHour | CronDom | CronMonth | CronDow)
//	schedule, err := parser.Parse("0 0 15 */3 *")
//
//	包含秒的解析器
//	parser := NewCronParser(CronSecond | CronMinute | CronHour | CronDom | CronMonth | CronDow)
//	schedule, err := parser.Parse("0 0 0 15 */3 *")
func NewCronParser(options CronParseOption) *CronParser {
	optionals := 0
	if options&CronDowOptional > 0 {
		optionals++
	}
	if options&CronSecondOptional > 0 {
		optionals++
	}
	if optionals > 1 {
		panic("不能配置多个可选字段")
	}
	return &CronParser{options: options}
}

// Parse 解析 cron 表达式并返回调度规范
// 支持时区、描述符和标准 cron 表达式
func (p *CronParser) Parse(spec string) (CronSchedule, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("cron 表达式不能为空")
	}

	// 提取时区(如果存在)
	loc := time.Local
	if strings.HasPrefix(spec, "TZ=") || strings.HasPrefix(spec, "CRON_TZ=") {
		i, eq := strings.Index(spec, " "), strings.Index(spec, "=")
		if i == -1 || eq == -1 {
			return nil, fmt.Errorf("时区格式错误: %s", spec)
		}
		var err error
		if loc, err = time.LoadLocation(spec[eq+1 : i]); err != nil {
			return nil, fmt.Errorf("无效的时区 %s: %v", spec[eq+1:i], err)
		}
		spec = strings.TrimSpace(spec[i:])
	}

	// 处理命名调度(描述符)
	if strings.HasPrefix(spec, "@") {
		if p.options&CronDescriptor == 0 {
			return nil, fmt.Errorf("描述符不受支持: %v", spec)
		}
		// 优先使用别名映射
		if expr, ok := GetExpression(spec); ok {
			spec = expr
		} else {
			// 处理 @every 等特殊描述符，以及别名未覆盖的标准描述符
			return parseCronDescriptor(spec, loc)
		}
	}

	// 按空白字符分割
	fields := strings.Fields(spec)

	// 验证并填充省略或可选的字段
	var err error
	fields, err = p.normalizeFields(fields)
	if err != nil {
		return nil, err
	}

	// 创建 schedule 对象
	schedule := &CronSpecSchedule{
		Location: loc,
	}

	// 解析秒字段
	schedule.Second, _, _, _, _, _, err = ParseFieldWithSpecialChars(fields[0], cronSeconds, 0, false, false)
	if err != nil {
		return nil, fmt.Errorf("解析秒字段失败: %v", err)
	}

	// 解析分钟字段
	schedule.Minute, _, _, _, _, _, err = ParseFieldWithSpecialChars(fields[1], cronMinutes, 0, false, false)
	if err != nil {
		return nil, fmt.Errorf("解析分钟字段失败: %v", err)
	}

	// 解析小时字段
	schedule.Hour, _, _, _, _, _, err = ParseFieldWithSpecialChars(fields[2], cronHours, 0, false, false)
	if err != nil {
		return nil, fmt.Errorf("解析小时字段失败: %v", err)
	}

	// 解析日期字段 (支持 L, W, LW)
	schedule.Dom, schedule.LastDay, schedule.LastWeekday, schedule.NearestWeekday, _, _, err =
		ParseFieldWithSpecialChars(fields[3], cronDom, cronStarBit, true, false)
	if err != nil {
		return nil, fmt.Errorf("解析日期字段失败: %v", err)
	}

	// 解析月份字段
	schedule.Month, _, _, _, _, _, err = ParseFieldWithSpecialChars(fields[4], cronMonths, 0, false, false)
	if err != nil {
		return nil, fmt.Errorf("解析月份字段失败: %v", err)
	}

	// 解析星期字段 (支持 L, #)
	schedule.Dow, _, _, _, schedule.LastDow, schedule.NthDow, err =
		ParseFieldWithSpecialChars(fields[5], cronDow, cronStarBit, false, true)
	if err != nil {
		return nil, fmt.Errorf("解析星期字段失败: %v", err)
	}

	return schedule, nil
}

// normalizeFields 规范化字段，填充默认值
func (p *CronParser) normalizeFields(fields []string) ([]string, error) {
	// 验证可选字段并添加到选项中
	options := p.options
	if options&CronSecondOptional > 0 {
		options |= CronSecond
	}
	if options&CronDowOptional > 0 {
		options |= CronDow
	}
	optionals := mathx.IF(options&CronSecondOptional > 0, 1, 0) + mathx.IF(options&CronDowOptional > 0, 1, 0)
	if optionals > 1 {
		return nil, fmt.Errorf("不能配置多个可选字段")
	}

	// 计算需要的字段数量
	max := 0
	for _, place := range cronPlaces {
		if options&place > 0 {
			max++
		}
	}
	min := max - optionals

	// 验证字段数量
	if count := len(fields); count < min || count > max {
		msg := mathx.IF(min == max,
			fmt.Sprintf("期望 %d 个字段，实际 %d 个: %s", min, count, strings.Join(fields, " ")),
			fmt.Sprintf("期望 %d 到 %d 个字段，实际 %d 个: %s", min, max, count, strings.Join(fields, " ")),
		)
		return nil, fmt.Errorf("%s", msg)
	}

	// 如果未提供可选字段，则填充默认值
	if min < max && len(fields) == min {
		if options&CronDowOptional > 0 {
			fields = append(fields, cronDefaults[5])
		} else if options&CronSecondOptional > 0 {
			fields = append([]string{cronDefaults[0]}, fields...)
		} else {
			return nil, fmt.Errorf("未知的可选字段")
		}
	}

	// 填充不在选项中的字段为默认值
	n := 0
	expandedFields := make([]string, len(cronPlaces))
	copy(expandedFields, cronDefaults)
	for i, place := range cronPlaces {
		if options&place > 0 {
			expandedFields[i] = fields[n]
			n++
		}
	}
	return expandedFields, nil
}

// CronStandardParser 标准解析器(分 时 日 月 周，5个字段)
var CronStandardParser = NewCronParser(
	CronMinute | CronHour | CronDom | CronMonth | CronDow | CronDescriptor,
)

// CronSecondParser 带秒的解析器(秒 分 时 日 月 周，6个字段)
var CronSecondParser = NewCronParser(
	CronSecond | CronMinute | CronHour | CronDom | CronMonth | CronDow | CronDescriptor,
)

// ParseCronStandard 使用标准解析器解析 cron 表达式(5个字段)
func ParseCronStandard(spec string) (CronSchedule, error) {
	return CronStandardParser.Parse(spec)
}

// ParseCronWithSeconds 使用带秒的解析器解析 cron 表达式(6个字段)
func ParseCronWithSeconds(spec string) (CronSchedule, error) {
	return CronSecondParser.Parse(spec)
}
