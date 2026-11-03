/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-21 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-04 19:31:55
 * @FilePath: \go-toolbox\pkg\idgen\factory.go
 * @Description: ID 生成器工厂
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package idgen

import (
	"github.com/HitGameAI/apex-toolbox/pkg/osx"
)

// NewIDGenerator 创建 ID 生成器
func NewIDGenerator(generatorType interface{}) IDGenerator {
	var typeStr string

	switch v := generatorType.(type) {
	case GeneratorType:
		typeStr = v.String()
	case string:
		typeStr = v
	default:
		return NewDefaultIDGenerator()
	}

	switch typeStr {
	case "uuid":
		return NewUUIDGenerator()
	case "nanoid":
		return NewNanoIDGenerator()
	case "snowflake":
		return NewSnowflakeGenerator(
			boundNodeID(osx.GetWorkerIdForSnowflake(), SnowflakeMaxWorkerID),
			boundNodeID(osx.GetDatacenterId(), SnowflakeMaxDatacenterID),
		)
	case "shortflake", "short":
		return NewShortFlakeGenerator(boundNodeID(osx.GetWorkerId(), ShortFlakeMaxNodeID))
	case "shortid":
		return NewShortIDGenerator()
	case "numeric":
		return NewNumericIDGenerator()
	case "ulid":
		return NewULIDGenerator()
	case "default", "hex", "logger", "":
		return NewDefaultIDGenerator()
	default:
		return NewDefaultIDGenerator()
	}
}

// NewIDGeneratorFromString 从字符串创建 ID 生成器（已废弃，使用 NewIDGenerator 代替）
// Deprecated: 使用 NewIDGenerator 代替
func NewIDGeneratorFromString(generatorType string) IDGenerator {
	return NewIDGenerator(generatorType)
}

// boundNodeID 把宿主推导出的 ID 收进生成器的容量内
//
// osx 推导出的 worker ID 上限由 osx.SetMaxWorkerID 决定（默认 1024），与各生成器
// 的位宽并不相同，所以这里必须收口，否则构造函数会按越界 panic
//
// 要注意这一步本身就是折回：部署规模超过 maxID+1 台时，第 maxID+1 台会和第 1 台
// 落到同一个 ID，从此发出重复的 ID工厂拿不到部署规模，没法替调用方判断，只能
// 保证不崩需要严格保证不重号时，用 NewSnowflakeGeneratorWithConfig 或
// NewShortFlakeGeneratorWithConfig 显式指定，让越界在启动时报出来
// ShortFlake 的节点位只有 6 位，上限 64 台，尤其容易踩到
//
// [EN] Folds a host-derived ID into a generator's capacity. The worker ID range
// osx derives is set by osx.SetMaxWorkerID (1024 by default) and does not match
// each generator's field width, so it has to be bounded here or the constructor
// would panic. Note this step is itself a fold: beyond maxID+1 machines, machine
// maxID+2 collides with the first and both emit duplicate IDs. The factory
// cannot know the deployment size, so it only guarantees it will not crash.
// Where uniqueness must be guaranteed, pass an explicit config so an
// out-of-range ID is reported at startup instead. ShortFlake's 6-bit node field,
// capping at 64 machines, is the easiest to hit
func boundNodeID(id, maxID int64) int64 {
	if id < 0 {
		id = -id
	}
	if id < 0 {
		// -id 溢出：仅当 id 为 MinInt64 时发生，取 0
		return 0
	}
	return id % (maxID + 1)
}
