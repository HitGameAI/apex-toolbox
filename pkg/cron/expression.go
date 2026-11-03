/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-24 19:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 11:59:31
 * @FilePath: \go-toolbox\pkg\cron\expression.go
 * @Description: Cron 表达式字段解析器
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package cron

import (
	"fmt"
	"strings"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/errorx"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// 基础周期表达式(6字段：秒 分 时 日 月 周)
const (
	EverySecond    = "* * * * * *"    // 每秒
	EveryMinute    = "0 * * * * *"    // 每分钟
	Every5Minutes  = "0 */5 * * * *"  // 每5分钟
	Every15Minutes = "0 */15 * * * *" // 每15分钟
	Every30Minutes = "0 */30 * * * *" // 每30分钟
	EveryHour      = "0 0 * * * *"    // 每小时
	Every2Hours    = "0 0 */2 * * *"  // 每2小时
	Every3Hours    = "0 0 */3 * * *"  // 每3小时
	Every6Hours    = "0 0 */6 * * *"  // 每6小时
	Every12Hours   = "0 0 */12 * * *" // 每12小时
	EveryDay       = "0 0 0 * * *"    // 每天午夜
	EveryWeek      = "0 0 0 * * 0"    // 每周日午夜
	EveryMonth     = "0 0 0 1 * *"    // 每月1号午夜
	EveryYear      = "0 0 0 1 1 *"    // 每年1月1日午夜
)

// 工作日与周末表达式
const (
	WorkdaysOnly  = "0 0 0 * * 1-5"  // 工作日(周一到周五)午夜
	WeekendsOnly  = "0 0 0 * * 0,6"  // 周末(周六和周日)午夜
	WorkdaysAt9AM = "0 0 9 * * 1-5"  // 工作日早9点
	WorkdaysAt6PM = "0 0 18 * * 1-5" // 工作日晚6点
	WeekendAt10AM = "0 0 10 * * 0,6" // 周末早10点
)

// 具体星期表达式
const (
	Monday    = "0 0 0 * * 1" // 每周一午夜
	Tuesday   = "0 0 0 * * 2" // 每周二午夜
	Wednesday = "0 0 0 * * 3" // 每周三午夜
	Thursday  = "0 0 0 * * 4" // 每周四午夜
	Friday    = "0 0 0 * * 5" // 每周五午夜
	Saturday  = "0 0 0 * * 6" // 每周六午夜
	Sunday    = "0 0 0 * * 0" // 每周日午夜
)

// 特定时段表达式
const (
	DailyAt9AM     = "0 0 9 * * *"         // 每天早9点
	DailyAt12PM    = "0 0 12 * * *"        // 每天中午12点
	DailyAt6PM     = "0 0 18 * * *"        // 每天晚6点
	DailyAt11PM    = "0 0 23 * * *"        // 每天晚11点
	MorningHours   = "0 0 6-12 * * *"      // 早晨时段(6-12点)
	AfternoonHours = "0 0 12-18 * * *"     // 下午时段(12-18点)
	EveningHours   = "0 0 18-23 * * *"     // 晚上时段(18-23点)
	BusinessHours  = "0 0 9-17 * * *"      // 营业时间(9-17点)
	PeakHours      = "0 0 8-9,17-18 * * *" // 高峰期(8-9点和17-18点)
	OffPeakHours   = "0 0 0-6,22-23 * * *" // 低谷期(0-6点和22-23点)
)

// 月度表达式
const (
	FirstDayOfMonth = "0 0 0 1 * *"  // 每月第1天午夜
	LastDayOfMonth  = "0 0 0 L * *"  // 每月最后一天午夜(需要特殊处理)
	MidMonth        = "0 0 0 15 * *" // 每月15号午夜
	MonthlyReport   = "0 0 2 1 * *"  // 每月1号凌晨2点(报表生成)
)

// 季度和年度表达式
const (
	QuarterlyReport = "0 0 2 1 */3 *" // 每季度首日凌晨2点
	YearlyReport    = "0 0 2 1 1 *"   // 每年1月1日凌晨2点
)

// 调试和测试表达式
const (
	Every10Seconds = "*/10 * * * * *" // 每10秒(测试用)
	Every30Seconds = "*/30 * * * * *" // 每30秒(测试用)
)

// ExpressionAliases 表达式别名映射
var ExpressionAliases = map[string]string{
	// 基础周期别名
	"@secondly": EverySecond,
	"@minutely": EveryMinute,
	"@hourly":   EveryHour,
	"@daily":    EveryDay,
	"@midnight": EveryDay,
	"@weekly":   EveryWeek,
	"@monthly":  EveryMonth,
	"@yearly":   EveryYear,
	"@annually": EveryYear,

	// 工作日和周末
	"@workdays": WorkdaysOnly,
	"@weekends": WeekendsOnly,

	// 每 N 分钟
	"@every_5min":  Every5Minutes,
	"@every_15min": Every15Minutes,
	"@every_30min": Every30Minutes,

	// 每 N 小时
	"@every_2h":  Every2Hours,
	"@every_3h":  Every3Hours,
	"@every_6h":  Every6Hours,
	"@every_12h": Every12Hours,

	// 具体星期别名
	"@monday":    Monday,
	"@tuesday":   Tuesday,
	"@wednesday": Wednesday,
	"@thursday":  Thursday,
	"@friday":    Friday,
	"@saturday":  Saturday,
	"@sunday":    Sunday,

	// 特定时段
	"@daily_9am":  DailyAt9AM,
	"@daily_12pm": DailyAt12PM,
	"@daily_6pm":  DailyAt6PM,
	"@daily_11pm": DailyAt11PM,

	// 工作时间段
	"@business_hours": BusinessHours,
	"@morning_hours":  MorningHours,
	"@afternoon":      AfternoonHours,
	"@evening":        EveningHours,
	"@peak_hours":     PeakHours,
	"@off_peak":       OffPeakHours,

	// 工作日特定时间
	"@workdays_9am": WorkdaysAt9AM,
	"@workdays_6pm": WorkdaysAt6PM,

	// 周末特定时间
	"@weekend_10am": WeekendAt10AM,

	// 月度相关
	"@first_of_month": FirstDayOfMonth,
	"@mid_month":      MidMonth,
	"@monthly_report": MonthlyReport,

	// 季度和年度
	"@quarterly":     QuarterlyReport,
	"@yearly_report": YearlyReport,

	// 测试用
	"@every_10s": Every10Seconds,
	"@every_30s": Every30Seconds,
}

// GetExpression 通过别名获取 Cron 表达式
func GetExpression(alias string) (string, bool) {
	expr, ok := ExpressionAliases[alias]
	return expr, ok
}

// ParseFieldToBits 解析逗号分隔的字段表达式，返回位集合
// 示例: "1-5,10,15-20/2" -> bits
// starBit: 当表达式为纯 * 或 ? 时，设置此位(用于 Dom/Dow 特殊逻辑)
func ParseFieldToBits[T types.Numerical](field string, bounds types.Bounds[T], starBit uint64) (uint64, error) {
	var bits uint64
	for _, expr := range strings.Split(field, ",") {
		bit, err := ParseExprToBits(expr, bounds, starBit)
		if err != nil {
			return 0, err
		}
		bits |= bit
	}
	return bits, nil
}

// ParseExprToBits 解析单个 Cron 表达式，支持: number | start-end[/step] | * | ?
// 示例: "1-10/2", "jan-mar", "*/5", "*"
// 注意：不处理 L, W, #, C 等特殊字符，这些由 ParseFieldWithSpecialChars 处理
func ParseExprToBits[T types.Numerical](expr string, bounds types.Bounds[T], starBit uint64) (uint64, error) {
	parts := strings.Split(expr, "/")
	if len(parts) > 2 {
		return 0, errorx.NewInvalidFormatError(fmt.Sprintf("表达式 '%s' 包含过多斜杠", expr))
	}

	// 解析范围
	start, end, isStar, err := parseFieldExpr(parts[0], bounds)
	if err != nil {
		return 0, err
	}

	// 解析步长
	step := T(1)
	if len(parts) == 2 {
		if step, err = ParseIntOrName[T](parts[1], nil); err != nil {
			return 0, errorx.WrapError("解析步长失败", err)
		}
		end = mathx.IF(start == end, bounds.Max, end) // "N/step" -> "N-max/step"
		isStar = false                                // 有步长不算纯通配符
	}

	// 验证
	if err := validateFieldExpr(start, end, step, bounds); err != nil {
		return 0, err
	}

	// 纯通配符时应用 starBit
	return mathx.GetBit64(uint(start), uint(end), uint(step)) | mathx.IF(isStar, starBit, 0), nil
}

// parseFieldExpr 解析字段表达式
func parseFieldExpr[T types.Numerical](expr string, bounds types.Bounds[T]) (start, end T, isStar bool, err error) {
	if expr == "*" || expr == "?" {
		return bounds.Min, bounds.Max, true, nil
	}

	parts := strings.Split(expr, "-")
	if len(parts) > 2 {
		return 0, 0, false, errorx.NewInvalidFormatError(fmt.Sprintf("表达式 '%s' 包含过多连字符", expr))
	}

	if start, err = ParseIntOrName(parts[0], bounds.Names); err != nil {
		return 0, 0, false, err
	}

	if len(parts) == 2 {
		if end, err = ParseIntOrName(parts[1], bounds.Names); err != nil {
			return 0, 0, false, err
		}
	} else {
		end = start
	}
	return start, end, false, nil
}

// validateFieldExpr 验证字段表达式有效性
func validateFieldExpr[T types.Numerical](start, end, step T, bounds types.Bounds[T]) error {
	if step <= 0 {
		return errorx.NewInvalidParamError("步长必须大于零")
	}
	if start < bounds.Min || end > bounds.Max {
		return errorx.NewInvalidParamError(fmt.Sprintf("值超出范围 [%v, %v]", bounds.Min, bounds.Max))
	}
	if start > end {
		return errorx.NewInvalidParamError(fmt.Sprintf("起始值 %v 大于结束值 %v", start, end))
	}
	return nil
}

// ParseIntOrName 解析整数或命名值
func ParseIntOrName[T types.Numerical](expr string, names map[string]T) (T, error) {
	var zero T
	// 命名值查找
	if names != nil {
		if val, ok := names[strings.ToLower(expr)]; ok {
			return val, nil
		}
	}

	// 整数解析
	val, err := convert.MustIntT[T](expr, nil)
	if err != nil {
		return zero, errorx.WrapError(fmt.Sprintf("无法解析 '%s'", expr), err)
	}
	if val < 0 {
		return zero, errorx.NewInvalidParamError("数值不能为负")
	}

	// 特殊处理：DOW字段兼容 7 作为周日 (Quartz兼容性)
	if _, hasSun := names["sun"]; names != nil && hasSun && val > 6 {
		if val == 7 {
			return 0, nil // 7 转换为周日(0)
		}
		return zero, errorx.NewInvalidParamError(fmt.Sprintf("星期值超出范围，有效值为 0-7，实际值为 %v", val))
	}

	return val, nil
}

// ParseFieldWithSpecialChars 解析可能包含特殊字符的字段
// 返回值：bits, lastDay, lastWeekday, nearestWeekday, lastDow, nthDow, error
func ParseFieldWithSpecialChars(fieldStr string, bounds cronBounds, starBit uint64, isDOM bool, isDOW bool) (
	bits uint64,
	lastDay bool,
	lastWeekday bool,
	nearestWeekday int,
	lastDow int,
	nthDow int,
	err error,
) {
	// 初始化返回值
	nearestWeekday = -1
	lastDow = -1
	nthDow = -1

	// 检查 LW (最后一个工作日)
	if isDOM && fieldStr == "LW" {
		lastWeekday = true
		return 0, false, true, -1, -1, -1, nil
	}

	// 检查 L (最后一天)
	if isDOM && fieldStr == "L" {
		lastDay = true
		return 0, true, false, -1, -1, -1, nil
	}

	// 检查星期字段的单独 L (在 Quartz 中，单独的 L 在星期字段表示周六，即最后一天)
	if isDOW && fieldStr == "L" {
		// 在 Quartz 中，星期字段的 L 通常表示周六(6)或周日(0)，这取决于实现
		// 这里我们将其解释为周日(0)
		lastDow = 0
		return 0, false, false, -1, 0, -1, nil
	}

	// 检查 nW (最近的工作日，如 15W)
	if isDOM && len(fieldStr) > 1 && fieldStr[len(fieldStr)-1] == cronWeekdayChar {
		dayStr := fieldStr[:len(fieldStr)-1]
		day, parseErr := convert.MustIntT[int](dayStr, nil)
		if parseErr != nil {
			err = fmt.Errorf("无效的工作日表达式 '%s': %v", fieldStr, parseErr)
			return
		}
		if day < 1 || day > 31 {
			err = fmt.Errorf("工作日的日期必须在 1-31 之间: %s", fieldStr)
			return
		}
		nearestWeekday = day
		return 0, false, false, day, -1, -1, nil
	}

	// 检查星期字段的 nL (最后一个星期X，如 6L 表示最后一个星期五)
	if isDOW && len(fieldStr) > 1 && fieldStr[len(fieldStr)-1] == cronLastDayChar {
		dowStr := fieldStr[:len(fieldStr)-1]
		var dow uint
		if dowVal, ok := cronDow.Names[strings.ToLower(dowStr)]; ok {
			dow = dowVal
		} else {
			dowInt, parseErr := convert.MustIntT[int](dowStr, nil)
			if parseErr != nil {
				err = fmt.Errorf("无效的星期L表达式 '%s': %v", fieldStr, parseErr)
				return
			}
			if dowInt < 0 || dowInt > 6 {
				err = fmt.Errorf("星期必须在 0-6 之间: %s", fieldStr)
				return
			}
			dow = uint(dowInt)
		}
		lastDow = int(dow)
		return 0, false, false, -1, lastDow, -1, nil
	}

	// 检查星期字段的 n#m (第m个星期n，如 6#3 表示第3个星期五)
	if isDOW && strings.Contains(fieldStr, string(cronNthChar)) {
		parts := strings.Split(fieldStr, string(cronNthChar))
		if len(parts) != 2 {
			err = fmt.Errorf("无效的#表达式 '%s'", fieldStr)
			return
		}

		// 解析星期
		var dow uint
		if dowVal, ok := cronDow.Names[strings.ToLower(parts[0])]; ok {
			dow = dowVal
		} else {
			dowInt, parseErr := convert.MustIntT[int](parts[0], nil)
			if parseErr != nil {
				err = fmt.Errorf("无效的星期#表达式 '%s': %v", fieldStr, parseErr)
				return
			}
			if dowInt < 0 || dowInt > 6 {
				err = fmt.Errorf("星期必须在 0-6 之间: %s", fieldStr)
				return
			}
			dow = uint(dowInt)
		}

		// 解析第几个
		nth, parseErr := convert.MustIntT[int](parts[1], nil)
		if parseErr != nil {
			err = fmt.Errorf("无效的#表达式 '%s': %v", fieldStr, parseErr)
			return
		}
		if nth < 1 || nth > 5 {
			err = fmt.Errorf("第几个星期必须在 1-5 之间: %s", fieldStr)
			return
		}

		// 编码为单个整数：高位是星期，低位是第几个
		nthDow = int(dow)*10 + nth
		return 0, false, false, -1, -1, nthDow, nil
	}

	// 处理包含逗号的列表，特别处理其中可能的特殊字符
	if strings.Contains(fieldStr, ",") {
		parts := strings.Split(fieldStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)

			// 如果这部分包含特殊字符，则不能作为列表的一部分
			if isDOM && (part == "L" || part == "LW" || (len(part) > 1 && part[len(part)-1] == cronWeekdayChar)) {
				err = fmt.Errorf("特殊字符表达式 '%s' 不能出现在列表中: %s", part, fieldStr)
				return
			}
			if isDOW && (len(part) > 1 && (part[len(part)-1] == cronLastDayChar || strings.Contains(part, string(cronNthChar)))) {
				err = fmt.Errorf("特殊字符表达式 '%s' 不能出现在列表中: %s", part, fieldStr)
				return
			}
		}
	}
	// ��ͨ�ֶν���
	bits, err = ParseFieldToBits(fieldStr, bounds, starBit)
	return bits, false, false, -1, -1, -1, err
}
