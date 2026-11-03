/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-25 10:30:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-25 11:26:37
 * @FilePath: \go-toolbox\pkg\cron\descriptor.go
 * @Description: Cron 描述符解析(如 @yearly, @monthly 等)
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

// parseCronDescriptor 解析特殊描述符(支持 @every 和 ExpressionAliases 未覆盖的特殊描述符)
// 注意：大部分标准描述符已在 parser.Parse 中通过 ExpressionAliases 处理
func parseCronDescriptor(descriptor string, loc *time.Location) (CronSchedule, error) {
	// 特殊时段描述符(ExpressionAliases 中没有的)
	switch descriptor {
	case "@night":
		// 深夜时段(凌晨2点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 2), nil

	case "@dawn":
		// 黎明时段(早晨6点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 6), nil

	case "@noon":
		// 中午12点
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 12), nil

	case "@dusk":
		// 黄昏时段(傍晚6点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 18), nil

	case "@late_night":
		// 深夜时段(23点-凌晨1点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1<<23 | 1<<0 | 1<<1), nil

	case "@early_morning":
		// 清晨时段(早晨5-7点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(mathx.GetBit64(5, 7, 1)), nil

	case "@lunch_time":
		// 午餐时间(11点-13点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(mathx.GetBit64(11, 13, 1)), nil

	case "@dinner_time":
		// 晚餐时间(18点-20点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(mathx.GetBit64(18, 20, 1)), nil

	case "@workday_start":
		// 工作日开始(周一到周五早上9点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 9).
			WithDow(mathx.GetBit64(1, 5, 1)), nil

	case "@workday_end":
		// 工作日结束(周一到周五晚上6点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 18).
			WithDow(mathx.GetBit64(1, 5, 1)), nil

	case "@weekend_morning":
		// 周末早晨(周六日早上10点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 10).
			WithDow(cronAtZero | 1<<6), nil

	case "@weekend_evening":
		// 周末晚上(周六日晚上8点)
		return NewZeroCronSpecSchedule(loc).
			WithHour(1 << 20).
			WithDow(cronAtZero | 1<<6), nil

	case "@month_end":
		// 每月最后一天(需要特殊处理，这里用28-31日近似)
		return NewZeroCronSpecSchedule(loc).
			WithDom(mathx.GetBit64(28, 31, 1)), nil

	case "@quarter_start":
		// 每季度开始(1月、4月、7月、10月的1号)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(1<<1 | 1<<4 | 1<<7 | 1<<10), nil

	case "@quarter_end":
		// 每季度结束(3月、6月、9月、12月的最后一天)
		return NewZeroCronSpecSchedule(loc).
			WithDom(mathx.GetBit64(28, 31, 1)).
			WithMonth(1<<3 | 1<<6 | 1<<9 | 1<<12), nil

	case "@year_start":
		// 每年开始(1月1日午夜)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(cronAtFirst), nil

	case "@year_end":
		// 每年结束(12月31日午夜)
		return NewZeroCronSpecSchedule(loc).
			WithDom(1 << 31).
			WithMonth(1 << 12), nil

	case "@spring_start":
		// 春季开始(3月1日)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(1 << 3), nil

	case "@summer_start":
		// 夏季开始(6月1日)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(1 << 6), nil

	case "@autumn_start":
		// 秋季开始(9月1日)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(1 << 9), nil

	case "@winter_start":
		// 冬季开始(12月1日)
		return NewZeroCronSpecSchedule(loc).
			WithDom(cronAtFirst).
			WithMonth(1 << 12), nil
	}

	// @every 间隔执行
	const every = "@every "
	if strings.HasPrefix(descriptor, every) {
		duration, err := time.ParseDuration(descriptor[len(every):])
		if err != nil {
			return nil, fmt.Errorf("解析间隔失败 '%s': %v", descriptor, err)
		}
		if duration < time.Second {
			return nil, fmt.Errorf("间隔太短: %v", duration)
		}
		return &CronEverySchedule{Duration: duration}, nil
	}

	return nil, fmt.Errorf("无法识别的描述符: %v", descriptor)
}
