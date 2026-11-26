/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-11-09 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-09 15:30:22
 * @FilePath: \go-toolbox\pkg\mathx\bit.go
 * @Description: 位操作相关工具函数集，包含64位和大整数位掩码的生成与解析
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package mathx

import (
	"math/big"
)

// GetBit64 返回范围 [min, max] 内，步长为 step 的所有位掩码
func GetBit64(min, max, step uint) uint64 {
	if step == 0 || min > max || max >= 64 {
		return 0
	}
	if step == 1 {
		// 生成从min到max连续的1位掩码
		length := max - min + 1
		return ((uint64(1) << length) - 1) << min
	}
	var bits uint64 = 0
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// Bit64ToArray 将掩码中所有被置1的位索引提取为数组
func Bit64ToArray(bit uint64) []uint {
	positions := make([]uint, 0) // 显式初始化为空切片
	for i := uint(0); i < 64; i++ {
		if bit&(1<<i) != 0 {
			positions = append(positions, i)
		}
	}
	return positions
}

// GetBitBig 返回范围 [min, max] 内，步长为 step 的所有位掩码，使用big.Int支持大位数
func GetBitBig(min, max, step uint) *big.Int {
	if step == 0 || min > max {
		return big.NewInt(0)
	}
	bits := big.NewInt(0)
	for i := min; i <= max; i += step {
		bits.SetBit(bits, int(i), 1)
	}
	return bits
}

// BitToArrayBig 将big.Int掩码中所有被置1的位索引提取为数组
func BitToArrayBig(bits *big.Int) []uint {
	positions := make([]uint, 0) // 显式初始化为空切片
	for i := 0; i < bits.BitLen(); i++ {
		if bits.Bit(i) == 1 {
			positions = append(positions, uint(i))
		}
	}
	return positions
}
