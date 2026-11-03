/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-11-08 03:59:11
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-12 23:05:08
 * @FilePath: \go-toolbox\pkg\random\rand.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package random

import (
	"fmt"
	"math"
	"math/rand"
	randv2 "math/rand/v2"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"encoding/json"

	"github.com/HitGameAI/apex-toolbox/pkg/convert"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// NewRand goroutine-safe rand.Rand, optional seed value
func NewRand(seed ...int64) *rand.Rand {
	n := time.Now().UnixMicro()
	n += atomic.AddInt64(&randCounter, 1)
	if len(seed) > 0 {
		n = seed[0]
	}
	return rand.New(rand.NewSource(n))
}

// fastIntn / fastInt63n / fastFloat64 内部快速随机数
//
// 为什么不走 NewRand
//
// NewRand 每次调用都要构造一整套生成器：rand.NewSource 会分配 607 项的 int64
// 表（约 4.9KB）并跑完播种预热实测一次 7.4 微秒、7KB原先 RandInt、
// RandFloat、RandString 以及 duration 里一共十六处都是「构造一套生成器只为取
// 一个数」，RandString 更要在循环里取 n 个UUID 调五次 RandString，于是一个
// UUID 要三十多微秒
//
// 这里改用 math/rand/v2 的顶层函数：它按 M 维护 ChaCha8 状态，goroutine 安全、
// 无锁、零分配，单次取数在纳秒级
//
// NewRand 本身保留且语义不变——给了显式种子时必须可复现，
// TestNewRand 就钉着 NewRand(1).Int63() 的精确值顺带去掉了原先那个每次调用
// 新建的 sync.Pool：池的寿命只有这一次调用，第二次 Get 永远不会发生，是纯开销
// 它还有个随机性缺陷：池内 New 闭包捕获同一个种子，池中每个 source 的序列完全
// 相同，一旦返回的 Rand 被并发共享，各 goroutine 会取到一样的随机数
//
// [EN] Internal fast random numbers, bypassing NewRand. NewRand constructs a
// whole generator per call: rand.NewSource allocates a 607-entry int64 table of
// about 4.9KB and runs the seeding warm-up, measured at 7.4 microseconds and 7KB.
// Sixteen call sites used to build a generator just to draw one number, and
// RandString drew n of them in a loop, so a UUID — five RandString calls — cost
// over thirty microseconds. These use math/rand/v2's top-level functions, which
// keep ChaCha8 state per M: goroutine-safe, lock-free, allocation-free and
// nanoseconds per draw. NewRand itself keeps its semantics, since an explicit
// seed must stay reproducible and TestNewRand pins NewRand(1).Int63() exactly.
// Its per-call sync.Pool is gone: the pool lived for one call so a second Get
// could never happen, and its New closure captured one seed, giving every source
// in the pool an identical sequence
func fastIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return randv2.IntN(n)
}

func fastInt63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return randv2.Int64N(n)
}

func fastFloat64() float64 {
	return randv2.Float64()
}

var (
	// 大写字母
	matchCapital *[]int
	// 小写字母
	matchLowercase *[]int
	// 特殊符号
	matchSpecial *[]int
	// 数字
	matchNumber *[]int

	once sync.Once

	randCounter int64
)

// RandInt
/**
 *  @Description: 随机整数
 *  @param start
 *  @param end
 *  @return v
 */
func RandInt(min, max int) (v int) {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return fastIntn(max-min) + min
}

// RandFloat
/**
 *  @Description: 随机小数
 *  @param min
 *  @param max
 *  @return v
 */
func RandFloat(min, max float64) (v float64) {
	return min + fastFloat64()*(max-min)
}

// initASCII 初始化ASCII码列表
func initASCII() {
	once.Do(func() {
		matchCapital = createASCIIList(65, 90)                                 // 大写字母
		matchLowercase = createASCIIList(97, 122)                              // 小写字母
		matchNumber = createASCIIList(48, 57)                                  // 数字
		matchSpecial = &[]int{46, 64, 36, 33, 37, 42, 35, 95, 126, 63, 38, 94} // 特殊字符 (. @$!%*#_~?&^)
		buildCharSets()
	})
}

// randCharSets 按 mode 预先算好的字符集
//
// mode 是四位掩码（CAPITAL=1 LOWERCASE=2 SPECIAL=4 NUMBER=8），一共十六种组合，
// 启动时一次算完原先 RandString 每次调用都要从 nil 开始 append 四个字符集，
// 反复扩容，而这四个集合是常量
//
// 存 []byte 而不是 []int：这些字符全在 ASCII 范围内，直接按字节写出即可，
// 不必走 WriteRune 的 UTF-8 编码路径
//
// [EN] Character sets precomputed per mode. The mode is a four-bit mask, so
// there are sixteen combinations, all built once at startup; RandString used to
// append the four sets onto a nil slice on every call, regrowing it each time,
// though the sets are constant. Stored as []byte rather than []int because every
// character is ASCII and can be written as a byte, with no need for WriteRune's
// UTF-8 encoding path
var randCharSets [16][]byte

// buildCharSets 由 ASCII 码表派生出十六种 mode 的字节集合
// 从 matchXxx 派生而不是另写一份字面量，保证两者不会走偏
func buildCharSets() {
	appendAll := func(dst []byte, src *[]int) []byte {
		for _, c := range *src {
			dst = append(dst, byte(c))
		}
		return dst
	}
	for mode := 0; mode < len(randCharSets); mode++ {
		var set []byte
		// 顺序与原实现一致：大写、小写、特殊、数字
		if mode&int(CAPITAL) != 0 {
			set = appendAll(set, matchCapital)
		}
		if mode&int(LOWERCASE) != 0 {
			set = appendAll(set, matchLowercase)
		}
		if mode&int(SPECIAL) != 0 {
			set = appendAll(set, matchSpecial)
		}
		if mode&int(NUMBER) != 0 {
			set = appendAll(set, matchNumber)
		}
		randCharSets[mode] = set
	}
}

// createASCIIList 创建指定范围的ASCII码列表
func createASCIIList(start, end int) *[]int {
	list := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		list[i-start] = i
	}
	return &list
}

// RandString
/**
 *  @Description: 随机生成字符串
 *  @param n 字符串长度
 *  @param mode 字符串模式 random.NUMBER|random.LOWERCASE|random.SPECIAL|random.CAPITAL)
 *  @return str 生成的字符串
 */
func RandString(n int, mode RandType) string {
	if n <= 0 {
		return ""
	}
	// mode 只有低四位有意义，高位按原实现一样忽略
	set := randCharSets[mode&0x0F]
	if len(set) == 0 {
		return ""
	}
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = set[fastIntn(len(set))]
	}
	return string(buf)
}

// UUID 生成 UUID 字符串
func UUID() string {
	return RandString(8, LOWERCASE|NUMBER) + "-" +
		RandString(4, LOWERCASE|NUMBER) + "-" +
		RandString(4, LOWERCASE|NUMBER) + "-" +
		RandString(4, LOWERCASE|NUMBER) + "-" +
		RandString(12, LOWERCASE|NUMBER)
}

// RandStringSlice 指定长度的随机字符串
func RandStringSlice(count, len int, mode RandType) (result []string) {
	for i := 0; i < count; i++ {
		result = append(result, RandString(len, mode))
	}
	return result
}

var defaultSliceSize = 1000

// RandNumericalLargeSlice 随机生成大数据整数切片
func RandNumericalLargeSlice[T types.Numerical](largeSize ...int) []T {
	if len(largeSize) > 0 {
		defaultSliceSize = largeSize[0]
	}
	slice := make([]T, defaultSliceSize)
	for i := 0; i < defaultSliceSize; i++ {
		slice[i] = T(i % 100) // 重复一些值以测试去重和重复检查
	}
	return slice
}

// RandNumber 随机一个数字字符串
func RandNumber(length int, customBytes ...string) string {
	var sb strings.Builder
	randBytes := DEC_BYTES
	if len(customBytes) > 0 {
		randBytes = customBytes[0]
	}
	if length > 0 {
		for i := 0; i < length; i++ {
			sb.WriteString(string(randBytes[NewRand().Intn(len(randBytes))]))
		}
	}
	return sb.String()
}

// RandHex 随机一个hex字符串
func RandHex(bytesLen int, customBytes ...string) string {
	var sb strings.Builder
	randBytes := HEX_BYTES
	if len(customBytes) > 0 {
		randBytes = customBytes[0]
	}
	for i := 0; i < bytesLen<<1; i++ {
		sb.WriteString(string(randBytes[NewRand().Intn(len(randBytes))]))
	}
	return sb.String()
}

// FRandInt (>=)min - (<)max
func FRandInt(min, max int) int {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return FastIntn(max-min) + min
}

// FRandUint32 (>=)min - (<)max
func FRandUint32(min, max uint32) uint32 {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return FastRandn(max-min) + min
}

// FastIntn this is similar to rand.Intn, but faster.
// A non-negative pseudo-random number in the half-open interval [0,n).
// Return 0 if n <= 0.
func FastIntn(n int) int {
	if n <= 0 {
		return 0
	}
	// 使用uint32(math.MaxUint32)避免32位架构上的常量溢出
	if uint64(n) <= uint64(uint32(math.MaxUint32)) {
		return int(FastRandn(uint32(n)))
	}
	return int(NewRand().Int63n(int64(n)))
}

// FRandString a random string, which may contain uppercase letters, lowercase letters and numbers.
// Ref: stackoverflow.icza
func FRandString(n int) string {
	return convert.B2S(FRandBytes(n))
}

// FRandHexString 指定长度的随机 hex 字符串
func FRandHexString(n int) string {
	return convert.B2S(FRandHexBytes(n))
}

// FRandAlphaString 指定长度的随机字母字符串
func FRandAlphaString(n int) string {
	return convert.B2S(FRandAlphaBytes(n))
}

// FRandDecString 指定长度的随机数字字符串
func FRandDecString(n int) string {
	return convert.B2S(FRandDecBytes(n))
}

// FRandBytes random bytes, but faster.
func FRandBytes(n int) []byte {
	return FRandBytesLetters(n, LETTER_BYTES)
}

// FRandAlphaBytes generates random alpha bytes.
func FRandAlphaBytes(n int) []byte {
	return FRandBytesLetters(n, ALPHA_BYTES)
}

// FRandHexBytes generates random hexadecimal bytes.
func FRandHexBytes(n int) []byte {
	return FRandBytesLetters(n, HEX_BYTES)
}

// RandDecBytes 指定长度的随机数字切片
func FRandDecBytes(n int) []byte {
	return FRandBytesLetters(n, DEC_BYTES)
}

// FRandBytesLetters 生成指定长度的字符切片
func FRandBytesLetters(n int, letters string) []byte {
	if n < 1 || len(letters) < 2 {
		return nil
	}
	b := make([]byte, n)
	for i, cache, remain := n-1, FastRand(), LETTER_IDX_MAX; i >= 0; {
		if remain == 0 {
			cache, remain = FastRand(), LETTER_IDX_MAX
		}
		if idx := int(cache & LETTER_IDX_MASK); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= LETTER_IDX_BITS
		remain--
	}
	return b
}

// FRandBytesJSON 生成指定长度的随机字节字符串，并返回 JSON 格式
func FRandBytesJSON(length int) (string, error) {
	// 生成随机字节
	randomBytes := FRandBytes(length)

	// 将字节转换为 JSON 格式
	jsonData, err := json.Marshal(randomBytes)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// 生成随机布尔值
func FRandBool() bool {
	return FRandInt(1, 2) == 1
}

// 生成随机时间
func FRandTime() time.Time {
	return time.Now().Add(time.Duration(FRandInt(1, 1000)) * time.Hour)
}

// GenerateAvailablePort 返回一个随机的可用端口号
// 如果未提供ports参数或提供的参数长度不为2，则使用默认端口范围1024到65535
func GenerateAvailablePort(ports ...int) (int, error) {
	// 设置默认端口范围
	minPort, maxPort := 1024, 65535

	// 检查是否提供了有效的ports参数
	if len(ports) == 2 {
		minPort = ports[0]
		maxPort = ports[1]
		// 验证提供的端口范围是否有效
		if minPort > maxPort {
			return 0, fmt.Errorf("minimum port cannot be greater than maximum port")
		}
	}

	for {
		port := RandInt(minPort, maxPort)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			// 成功监听端口，关闭监听器并返回端口号
			listener.Close()
			return port, nil
		}
		// 如果端口已被使用或发生其他错误，则尝试下一个端口
	}
}

// RandNumerical 泛型函数，生成从 start 到 end（包含）的切片
func RandNumerical[T types.Numerical](start, end T, step ...T) []T {
	// 如果结束值小于开始值，返回 nil 表示无效输入
	if end < start {
		return nil
	}

	// 默认步长为 1，如果传入了步长参数，则使用传入的步长
	var stepVal T = 1
	if len(step) > 0 {
		stepVal = step[0]
	}

	// 步长必须大于 0，否则返回 nil 表示无效步长
	if stepVal <= 0 {
		return nil
	}

	// 根据 start 的具体类型区分处理浮点数和整数
	switch any(start).(type) {
	// 如果是浮点数类型（float32 或 float64）
	case float32, float64:
		// 计算序列长度 size，向下取整后加 1，确保包含起点
		size := int((end-start)/stepVal) + 1
		if size <= 0 {
			return nil
		}

		// 创建结果切片，容量为 size
		result := make([]T, size)

		// 将步长和起点转换为 float64 方便计算
		stepF := float64(stepVal)
		startF := float64(start)

		// 依次计算每个元素的值：start + i * step
		for i := 0; i < size; i++ {
			result[i] = T(startF + float64(i)*stepF) // 直接转换为泛型浮点数类型 T
		}

		return result

	// 其他类型视为整数类型处理
	default:
		// 将整数类型参数转换为 float64，方便做除法和乘法计算
		startF := float64(start)
		endF := float64(end)
		stepF := float64(stepVal)

		// 计算序列长度 size，向下取整后加 1，确保包含起点
		size := int((endF-startF)/stepF) + 1
		if size <= 0 {
			return nil
		}

		// 创建结果切片，容量为 size
		result := make([]T, size)

		// 依次计算每个元素的值（浮点数计算）
		for i := 0; i < size; i++ {
			valF := startF + float64(i)*stepF

			// 利用 Float64ToInt 函数将浮点数安全转换回整数泛型 T
			// RoundNone 表示不做额外取整，直接转换
			val, err := convert.Float64ToInt[T](valF, convert.RoundNone)
			if err != nil {
				// 转换失败时 panic，提示错误信息
				panic(fmt.Sprintf("RandNumerical: 转换失败 %v", err))
			}
			result[i] = val
		}

		return result
	}
}

// RandNumericalWithRandomStep 泛型函数，生成从 start 到 end（包含或不超过）
// 每一步随机步长在 [minStep, maxStep] 范围内（浮点数或整数）
func RandNumericalWithRandomStep[T types.Numerical](start, end, minStep, maxStep T) []T {
	if end < start {
		return nil
	}
	if minStep <= 0 || maxStep < minStep {
		return nil
	}

	var result []T
	current := start

	switch any(start).(type) {
	case float32, float64:
		for current <= end {
			result = append(result, current)
			// 生成随机浮点步长
			step := RandFloat(float64(minStep), float64(maxStep))
			current += T(step)
		}
	default:
		for current <= end {
			result = append(result, current)
			// 生成随机整数步长
			step := RandInt(int(minStep), int(maxStep))
			current += T(step)
		}
	}

	return result
}

func init() {
	// 初始化ASCII码列表，只执行一次
	initASCII()
}
