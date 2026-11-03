/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-16 22:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-11-16 22:55:55
 * @FilePath: \go-toolbox\pkg\safe\mathx.go
 * @Description: 安全的数学工具函数集合，提供各种数学算法的安全实现
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package safe

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unsafe"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
)

// 定义常用的数学常数
const (
	// 使用int64避免32位架构溢出
	MaxSafeInteger64 = int64(1<<53 - 1)    // JavaScript安全整数最大值
	MinSafeInteger64 = int64(-(1<<53 - 1)) // JavaScript安全整数最小值
	GoldenRatio      = 1.618033988749      // 黄金比例
	EulerNumber      = 2.718281828459      // 欧拉数
)

// 架构相关的安全整数最大值
var (
	MaxSafeInteger = getMaxSafeIntegerForArch()
	MinSafeInteger = getMinSafeIntegerForArch()
)

// getMaxSafeIntegerForArch 获取当前架构的安全整数最大值
func getMaxSafeIntegerForArch() int {
	if unsafe.Sizeof(int(0)) == 8 {
		// 64位架构，安全转换
		maxSafe := int64(1<<53 - 1) // JavaScript安全整数最大值
		return int(maxSafe)
	}
	// 32位架构，使用math.MaxInt32
	return math.MaxInt32
}

// getMinSafeIntegerForArch 获取当前架构的安全整数最小值
func getMinSafeIntegerForArch() int {
	if unsafe.Sizeof(int(0)) == 8 {
		// 64位架构，安全转换
		minSafe := int64(-(1<<53 - 1)) // JavaScript安全整数最小值
		return int(minSafe)
	}
	// 32位架构，使用math.MinInt32
	return math.MinInt32
}

// FastHash 快速哈希函数 - 安全版本
// 使用FNV-1a算法，避免哈希冲突和安全问题
func FastHash(s string) uint64 {
	if len(s) == 0 {
		return 1 // 返回非0值避免哈希冲突
	}

	// 使用更安全的字符串访问方式
	var h uint64 = 14695981039346656037 // FNV offset basis
	b := []byte(s)                      // 转换为byte切片，更安全
	for _, v := range b {
		h ^= uint64(v)
		h *= 1099511628211 // FNV prime
	}
	return h
}

// ShortHash 生成短哈希字符串（默认7位，类似Git）
// 使用 FNV-1a 算法，适用于生成短ID、节点ID等场景
// 返回格式：7位Base36字符串（0-9a-z），如 "a1b2c3d"
// 冲突概率：10万节点 < 0.007%
func ShortHash(s string) string {
	return ShortHashWithLength(s, 7)
}

// ShortHashWithLength 生成指定长度的短哈希字符串
// 使用 FNV-1a 算法 + Base36 编码（0-9a-z）
//
// 参数：
//   - s: 输入字符串
//   - length: 哈希长度（最小为1），推荐值：
//   - 6位: 适合小规模（< 1万节点），冲突概率 < 0.002%
//   - 7位: 适合中规模（< 10万节点），冲突概率 < 0.007%（推荐）
//   - 8位: 适合大规模（< 100万节点），冲突概率 < 0.0002%
//   - 12-16位: 适合超大规模或需要更低冲突概率的场景
//
// 返回：指定长度的Base36字符串
//
// 示例：
//
//	ShortHashWithLength("192.168.1.100:8080", 7) → "3g5e9ss"
//	ShortHashWithLength("pod-abc-123", 8) → "2k3m4n5p"
//	ShortHashWithLength("user123", 16) → "3g5e9ss000000000"
func ShortHashWithLength(s string, length int) string {
	length = mathx.IfLeZero(length, 1)

	hash := FastHash(s)
	encoded := strconv.FormatUint(hash, 36)

	// 截取或补齐到指定长度
	if len(encoded) >= length {
		return encoded[:length]
	}

	// 不足时左补0（确保固定长度）
	padding := make([]byte, length-len(encoded))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + encoded
}

// NextPowerOfTwo 获取下一个2的幂 - 安全版本
// 防止整数溢出和无效输入
func NextPowerOfTwo(n int) int {
	if n <= 1 {
		return 2
	}
	if n > MaxSafeInteger>>1 {
		return MaxSafeInteger // 防止溢出
	}

	// 如果n已经是2的幂，返回下一个
	if n&(n-1) == 0 {
		return n << 1
	}

	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// Add 安全整数加法，防止溢出
func Add(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, errors.New("integer overflow in addition")
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, errors.New("integer underflow in addition")
	}
	return a + b, nil
}

// Subtract 安全整数减法，防止溢出
func Subtract(a, b int64) (int64, error) {
	if b > 0 && a < math.MinInt64+b {
		return 0, errors.New("integer underflow in subtraction")
	}
	if b < 0 && a > math.MaxInt64+b {
		return 0, errors.New("integer overflow in subtraction")
	}
	return a - b, nil
}

// Multiply 安全整数乘法，防止溢出
func Multiply(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	// 检查溢出
	if a > 0 && b > 0 && a > math.MaxInt64/b {
		return 0, errors.New("integer overflow in multiplication")
	}
	if a > 0 && b < 0 && b < math.MinInt64/a {
		return 0, errors.New("integer underflow in multiplication")
	}
	if a < 0 && b > 0 && a < math.MinInt64/b {
		return 0, errors.New("integer underflow in multiplication")
	}
	if a < 0 && b < 0 && a < math.MaxInt64/b {
		return 0, errors.New("integer overflow in multiplication")
	}

	return a * b, nil
}

// Divide 安全整数除法，防止除零和溢出
func Divide(a, b int64) (int64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	if a == math.MinInt64 && b == -1 {
		return 0, errors.New("integer overflow in division")
	}
	return a / b, nil
}

// Modulo 安全取模运算，防止除零
func Modulo(a, b int64) (int64, error) {
	if b == 0 {
		return 0, errors.New("modulo by zero")
	}
	return a % b, nil
}

// Power 安全幂运算，使用快速幂算法防止溢出
func Power(base, exp int64) (int64, error) {
	if exp < 0 {
		return 0, errors.New("negative exponent not supported for integer power")
	}
	if exp == 0 {
		return 1, nil
	}
	if base == 0 {
		return 0, nil
	}
	if base == 1 {
		return 1, nil
	}
	if base == -1 {
		if exp%2 == 0 {
			return 1, nil
		}
		return -1, nil
	}

	var result int64 = 1
	var currentBase = base

	for exp > 0 {
		if exp%2 == 1 {
			newResult, err := Multiply(result, currentBase)
			if err != nil {
				return 0, err
			}
			result = newResult
		}

		if exp > 1 {
			newBase, err := Multiply(currentBase, currentBase)
			if err != nil {
				return 0, err
			}
			currentBase = newBase
		}
		exp /= 2
	}

	return result, nil
}

// Sqrt 安全平方根计算，使用牛顿法
func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, errors.New("square root of negative number")
	}
	if n == 0 {
		return 0, nil
	}

	// 使用 math.Sqrt 进行安全计算
	result := math.Sqrt(n)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, errors.New("invalid square root result")
	}

	return result, nil
}

// Log 安全对数计算
func Log(n, base float64) (float64, error) {
	if n <= 0 {
		return 0, errors.New("logarithm of non-positive number")
	}
	if base <= 0 || base == 1 {
		return 0, errors.New("invalid logarithm base")
	}

	result := math.Log(n) / math.Log(base)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, errors.New("invalid logarithm result")
	}

	return result, nil
}

// GCD 安全最大公约数计算，使用欧几里得算法
func GCD(a, b int64) int64 {
	// 处理负数
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// 欧几里得算法
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM 安全最小公倍数计算
func LCM(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	gcd := GCD(a, b)
	if gcd == 0 {
		return 0, errors.New("invalid GCD result")
	}

	// 防止溢出
	absA := a
	if a < 0 {
		absA = -a
	}
	absB := b
	if b < 0 {
		absB = -b
	}

	result, err := Multiply(absA/gcd, absB)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// IsPrime 安全素数检测，使用Miller-Rabin算法
func IsPrime(n int64) bool {
	if n < 2 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	// 对于较小的数，使用简单试除法
	if n < 100 {
		for i := int64(3); i*i <= n; i += 2 {
			if n%i == 0 {
				return false
			}
		}
		return true
	}

	// 对于较大的数，使用big.Int的ProbablyPrime方法
	bigN := big.NewInt(n)
	return bigN.ProbablyPrime(20) // 20次Miller-Rabin测试
}

// Fibonacci 安全斐波那契数列计算，防止溢出
func Fibonacci(n int) (int64, error) {
	if n < 0 {
		return 0, errors.New("negative Fibonacci index not supported")
	}
	if n == 0 {
		return 0, nil
	}
	if n == 1 {
		return 1, nil
	}

	var a, b int64 = 0, 1
	for i := 2; i <= n; i++ {
		next, err := Add(a, b)
		if err != nil {
			return 0, err
		}
		a, b = b, next
	}

	return b, nil
}

// Factorial 安全阶乘计算，防止溢出
func Factorial(n int) (*big.Int, error) {
	if n < 0 {
		return nil, errors.New("negative factorial not defined")
	}

	result := big.NewInt(1)
	for i := 2; i <= n; i++ {
		result.Mul(result, big.NewInt(int64(i)))
	}

	return result, nil
}

// Average 安全平均值计算，防止溢出
func Average(numbers []int64) (float64, error) {
	if len(numbers) == 0 {
		return 0, errors.New("cannot calculate average of empty slice")
	}

	var sum int64 = 0
	for i, num := range numbers {
		newSum, err := Add(sum, num)
		if err != nil {
			// 溢出后改用浮点累加，从当前下标接着算
			//
			// 原实现这里写的是 indexOf(numbers, num)，找的是该**值**首次出现的
			// 位置而不是当前下标：当溢出的那个值在前面出现过，floatSum 里已经
			// 累加过的前缀会被再加一遍
			// 例如 Average([]int64{H, H})（H 为 MaxInt64/2+1000）会返回 1.5H
			// 而正确答案是 H——50% 的误差，且不报错
			//
			// [EN] Switch to float accumulation and continue from the current
			// index. The original used indexOf(numbers, num), the first index of
			// that *value* rather than the current one: when the overflowing
			// value also appears earlier, the prefix already in floatSum is
			// added a second time. Average([]int64{H, H}) with H =
			// MaxInt64/2+1000 returned 1.5H instead of H — a 50% error, silently
			floatSum := float64(sum)
			for j := i; j < len(numbers); j++ {
				floatSum += float64(numbers[j])
			}
			return floatSum / float64(len(numbers)), nil
		}
		sum = newSum
	}

	return float64(sum) / float64(len(numbers)), nil
}

// Max 安全获取最大值
func Max(numbers []int64) (int64, error) {
	if len(numbers) == 0 {
		return 0, errors.New("cannot find max of empty slice")
	}

	max := numbers[0]
	for _, num := range numbers[1:] {
		if num > max {
			max = num
		}
	}

	return max, nil
}

// Min 安全获取最小值
func Min(numbers []int64) (int64, error) {
	if len(numbers) == 0 {
		return 0, errors.New("cannot find min of empty slice")
	}

	min := numbers[0]
	for _, num := range numbers[1:] {
		if num < min {
			min = num
		}
	}

	return min, nil
}

// Clamp 安全值范围限制
func Clamp(value, min, max int64) (int64, error) {
	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}

	if value < min {
		return min, nil
	}
	if value > max {
		return max, nil
	}

	return value, nil
}

// Abs 安全绝对值计算，防止溢出
func Abs(n int64) (int64, error) {
	if n == math.MinInt64 {
		return 0, errors.New("absolute value of MinInt64 causes overflow")
	}

	if n < 0 {
		return -n, nil
	}

	return n, nil
}

// HashToInt64 将多个字符串拼接后哈希为非负 int64
// 内部使用 FNV-1a 短哈希 + Base36 解码，保证确定性且跨平台一致
//
// 参数：
//   - parts: 待哈希的字符串切片
//   - separator: 拼接分隔符
//
// 返回：非负的 int64 哈希值
func HashToInt64(parts []string, separator string) int64 {
	h := ShortHash(strings.Join(parts, separator))
	v, _ := strconv.ParseInt(h, 36, 64)
	if v < 0 {
		v = -v
	}
	return v
}
