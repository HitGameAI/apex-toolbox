/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-17 13:55:21
 * @FilePath: \go-toolbox\pkg\mathx\slice.go
 * @Description: 包含与切片相关的通用函数，例如计算最小值和最大值、差集、并集等
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package mathx

import (
	"cmp"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"sync"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// MinSlice 返回切片中的最小值
// Deprecated: 使用 SliceMinOrdered 替代，支持所有 Ordered 类型（int/float/string 等）
func MinSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return SliceMinOrdered(values)
}

// MaxSlice 返回切片中的最大值
//
// Deprecated: 使用 SliceMaxOrdered 替代，支持所有 Ordered 类型（int/float/string 等）
// 迁移示例：
//
//	// 旧：
//	max := MaxSlice(values)
//	// 新：
//	max := SliceMaxOrdered(values)
func MaxSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return SliceMaxOrdered(values)
}

// SliceMinMax 计算列表中元素的最小值或最大值
// 接收一个切片和一个 MinMaxFunc 类型的函数，
// 根据提供的函数决定是计算最小值还是最大值
// 如果列表为空，则返回错误
func SliceMinMax[T types.Numerical](list []T, f types.MinMaxFunc[T]) (T, error) {
	if len(list) == 0 {
		var zero T
		return zero, errors.New("列表为空") // 返回错误信息
	}

	result := list[0] // 初始化结果为列表的第一个元素
	for _, v := range list[1:] {
		result = f(result, v) // 使用提供的函数更新结果
	}
	return result, nil // 返回最终结果和 nil 错误
}

// SliceFisherYates 洗牌算法打乱数组
func SliceFisherYates[T comparable](slice []T, maxRetries int) error {
	original := make([]T, len(slice))
	copy(original, slice)

	for retries := 0; retries < maxRetries; retries++ {
		for i := len(slice) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)                   // 生成 0 到 i 之间的随机数
			slice[i], slice[j] = slice[j], slice[i] // 交换
		}

		// 校验洗牌后的切片是否与原始切片相同
		if !SliceEqual(original, slice) {
			return nil // 如果不同，返回 nil 表示成功
		}
	}

	// 如果达到最大重试次数，返回错误
	return errors.New("failed to shuffle slice after max retries")
}

// SliceDiffSetSorted 计算两个已排序数组的对称差集（symmetric difference）
// 返回仅在其中一个数组中出现的元素（保持有序，自动去重）
//
// 算法：双指针一次遍历，O(n+m) 时间复杂度
//   - arr1[i] < arr2[j]: arr1[i] 仅在 arr1 中，加入结果，i++
//   - arr1[i] > arr2[j]: arr2[j] 仅在 arr2 中，加入结果，j++
//   - 相等: 同时出现在两边，跳过，i++, j++
//   - 自动跳过各自数组内的重复元素（set 语义）
//
// 注意：输入必须已排序，否则结果不正确
func SliceDiffSetSorted[T types.Ordered](arr1, arr2 []T) []T {
	diff := make([]T, 0, len(arr1)+len(arr2))
	i, j := 0, 0

	for i < len(arr1) && j < len(arr2) {
		switch {
		case arr1[i] < arr2[j]:
			diff = append(diff, arr1[i])
			// 跳过 arr1 中的重复
			for i+1 < len(arr1) && arr1[i+1] == arr1[i] {
				i++
			}
			i++
		case arr1[i] > arr2[j]:
			diff = append(diff, arr2[j])
			// 跳过 arr2 中的重复
			for j+1 < len(arr2) && arr2[j+1] == arr2[j] {
				j++
			}
			j++
		default:
			// 相等，跳过（同时出现在两边，不属于对称差集）
			// 跳过 arr1 和 arr2 中的重复
			cur := arr1[i]
			for i < len(arr1) && arr1[i] == cur {
				i++
			}
			for j < len(arr2) && arr2[j] == cur {
				j++
			}
		}
	}

	// 添加 arr1 剩余元素（跳过重复）
	for ; i < len(arr1); i++ {
		if i == 0 || arr1[i] != arr1[i-1] {
			diff = append(diff, arr1[i])
		}
	}

	// 添加 arr2 剩余元素（跳过重复）
	for ; j < len(arr2); j++ {
		if j == 0 || arr2[j] != arr2[j-1] {
			diff = append(diff, arr2[j])
		}
	}

	return diff
}

// SliceUnion 计算两个数组的并集
//
// Deprecated: 使用 SliceUnionMulti 替代，支持变参多个切片且保持首次出现顺序
// 迁移示例：
//
//	// 旧：
//	union := SliceUnion(arr1, arr2)
//	// 新：
//	union := SliceUnionMulti(arr1, arr2)
//
// 注意：SliceUnionMulti 保持首次出现顺序，SliceUnion 不保证顺序（map 遍历）
func SliceUnion[T comparable](arr1, arr2 []T) []T {
	return SliceUnionMulti(arr1, arr2)
}

// SliceUniq 集合去重
func SliceUniq[T ~[]E, E comparable](list T) T {
	if len(list) == 0 {
		return list
	}

	ret := make(T, 0, len(list))
	m := make(map[E]struct{}, len(list))
	for _, v := range list {
		if _, exists := m[v]; !exists {
			ret = append(ret, v)
			m[v] = struct{}{}
		}
	}
	return ret
}

// SliceDiff 返回两个集合之间的差异
//
// Deprecated: 使用 SliceDifference 替代，语义更清晰
// 迁移示例：
//
//	// 旧：
//	left, right := SliceDiff(list1, list2)
//	// 新：
//	left, right := SliceDifference(list1, list2)
func SliceDiff[T ~[]E, E comparable](list1 T, list2 T) (ret1 T, ret2 T) {
	return SliceDifference(list1, list2)
}

// SliceWithout 返回不包括所有给定值的切片
func SliceWithout[T ~[]E, E comparable](list T, exclude ...E) T {
	if len(list) == 0 {
		return list
	}

	m := make(map[E]struct{}, len(exclude))
	for _, v := range exclude {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(list))
	for _, v := range list {
		if _, exists := m[v]; !exists {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceIntersect 返回两个集合的交集
func SliceIntersect[T ~[]E, E comparable](list1 T, list2 T) T {
	m := make(map[E]struct{}, len(list1))
	for _, v := range list1 {
		m[v] = struct{}{}
	}

	ret := make(T, 0, len(list1)) // 预分配内存
	for _, v := range list2 {
		if _, exists := m[v]; exists {
			ret = append(ret, v)
		}
	}
	return ret
}

// SliceContains 检查切片中是否包含某个元素
//
// Deprecated: 使用 SliceContainsComparable 替代，支持所有 comparable 类型
// （types.Ordered 是 comparable 的子集，SliceContainsComparable 涵盖其全部用例）
// 迁移示例：
//
//	// 旧：
//	ok := SliceContains(slice, element)
//	// 新：
//	ok := SliceContainsComparable(slice, element)
func SliceContains[T types.Ordered](slice []T, element T) bool {
	return SliceContainsComparable(slice, element)
}

// SliceContainsComparable 检查切片中是否包含某个元素（支持 comparable 类型）
// 适用于枚举、字符串、结构体等任意 comparable 类型
//
// 一律线性查找，零分配
//
// 原先按长度分流：超过 1000 条就先把整个切片灌进 map 再查一次
// 那条分支在任何情况下都更慢，包括它本想优化的最坏情况（元素不存在或在末尾）——
// 线性扫描是 n 次比较，建 map 是 n 次插入，后者单次成本高出一个量级，
// 而「O(1) 查找」只省下最后那一次比较
// 实测（Apple M4，元素不存在）：
//
//	规模      线性              哈希
//	1000      232.7 ns / 0 B    7141 ns / 36 KB / 5 allocs
//	20000     4546 ns / 0 B     182821 ns / 591 KB / 65 allocs
//	200000    45520 ns / 0 B    2096300 ns / 4.7 MB / 513 allocs
//
// map 只在同一切片上反复查找时才划算，而本函数每次调用都要重建它，
// 调用方若要多次查找，应当自己建一次集合复用
//
// [EN] Always a linear scan, with no allocation. The length-based split that
// loaded slices over 1000 elements into a map first was slower in every case,
// including the worst case it meant to improve: a linear scan is n comparisons
// while building the map is n insertions, each an order of magnitude dearer,
// and the O(1) lookup only saves the final comparison. A map pays off only
// across repeated lookups on the same slice, which this function cannot do
// since it rebuilds one per call; callers needing many lookups should build a
// set once themselves
func SliceContainsComparable[T comparable](slice []T, element T) bool {
	for i := range slice {
		if slice[i] == element {
			return true
		}
	}
	return false
}

// SliceHasDuplicates 检查切片中是否存在重复元素
//
// 算法：单次遍历 + map 记录已见元素，O(n) 时间 O(n) 空间
//   - map 查找/插入平均 O(1)，比 goroutine + mutex 方案更快（无锁竞争、无 goroutine 启动开销）
//   - 发现第一个重复立即返回，避免完整遍历
//
// 修复说明：原实现用 goroutine 分块处理，存在三个问题：
//  1. 跨 chunk 的重复检测不到（每个 goroutine 只检查自己 chunk 内的重复）
//  2. 共享 map 写入有 race（len(m) 读取未加锁）
//  3. 小切片 goroutine 启动开销远大于直接计算
//
// 预分配到整个切片长度是刻意的，不要「顺手改掉」
//
// 看上去它抵消了提前退出的收益——重复出现在第 101 个元素时，仍为两万个槽位
// 付了钱实测过三种写法（Apple M4，两万个 int）：
//
//	形状            预分配              不给提示            上限 1024
//	第101个重复      44669 ns / 591 KB   2301 ns / 4.4 KB    3217 ns / 37 KB
//	最后一个重复     244265 ns / 591 KB  601343 ns / 1183 KB 576321 ns / 1145 KB
//	无重复          243984 ns / 591 KB  598450 ns / 1183 KB 578225 ns / 1145 KB
//
// 没有普遍更优的选择：提前退出的场景不给提示快 19 倍，而要扫完全程的场景
// 预分配快 2.5 倍且内存只要一半（后者的代价是 map 从小长到两万要反复 rehash）
// 本函数的典型用途是校验输入「有重复就拒绝」，合法数据里没有重复，
// 因此扫完全程才是常态，按它来取舍
// 调用方若明知多半有重复且只关心有无，自己用不给提示的 map 更划算
//
// [EN] Pre-sizing to the full length is deliberate. It looks as though it
// cancels the early exit, and three variants were measured: with no size hint
// an early duplicate is 19x faster, but a full scan is 2.5x slower and uses
// twice the memory, since the map has to rehash its way up from nothing. This
// function is typically used to validate input and reject duplicates, so valid
// data has none and the full scan is the common case; the trade follows that
func SliceHasDuplicates[T comparable](slice []T) bool {
	seen := make(map[T]struct{}, len(slice))
	for i := range slice {
		if _, ok := seen[slice[i]]; ok {
			return true // 发现重复，立即返回
		}
		seen[slice[i]] = struct{}{}
	}
	return false
}

// SliceRemoveEmpty 移除切片中的"空"元素（零值或 nil）
//
// 适用于元素类型不可比较（如含 slice/map 字段的结构体）的场景
// 对于 comparable 类型，优先使用 CompactSlice（用 == 比较零值，无反射开销）
//
// 性能：使用 reflect.ValueOf(v).IsZero() 直接判断，比 validator.IsEmptyValue 少一层封装
func SliceRemoveEmpty[T any](slice []T) []T {
	result := make([]T, 0, len(slice))
	for i := range slice {
		v := reflect.ValueOf(slice[i])
		// IsValid() 为 false 表示 nil interface，视为空值跳过
		if v.IsValid() && !v.IsZero() {
			result = append(result, slice[i])
		}
	}
	return result
}

// SliceRemoveDuplicates 移除切片中的重复值
//
// Deprecated: 使用 SliceUniq 替代，命名更简洁，语义更清晰
// 迁移示例：
//
//	// 旧：
//	uniq := SliceRemoveDuplicates(numbers)
//	// 新：
//	uniq := SliceUniq(numbers)
func SliceRemoveDuplicates[T comparable](numbers []T) []T {
	return SliceUniq(numbers)
}

// SliceRemove 根据给定的条件函数保留切片中的元素
//
// Deprecated: 使用 FilterSlice 替代，命名更符合通用约定（filter = 保留满足条件的）
// 迁移示例：
//
//	// 旧：
//	result := SliceRemove(arr, func(v T) bool { return v > 0 })
//	// 新：
//	result := FilterSlice(arr, func(v T) bool { return v > 0 })
//
// 注意：原 SliceRemove 命名有歧义（实际是保留满足条件的，而非移除），FilterSlice 命名更清晰
func SliceRemove[T any](arr []T, condition func(T) bool) []T {
	return FilterSlice(arr, condition)
}

// SliceRemoveZero 移除切片中的零值
//
// Deprecated: 使用 CompactSlice 替代，命名更简洁（compact = 移除零值）
// 迁移示例：
//
//	// 旧：
//	result := SliceRemoveZero(arr)
//	// 新：
//	result := CompactSlice(arr)
func SliceRemoveZero[T comparable](arr []T) []T {
	return CompactSlice(arr)
}

// SliceRemoveValue 移除切片中的指定值
//
// Deprecated: 使用 SliceWithout 替代，支持多值排除
// 迁移示例：
//
//	// 旧：
//	result := SliceRemoveValue(arr, value)
//	// 新：
//	result := SliceWithout(arr, value)
//
// 注意：SliceWithout 支持变参排除多个值：SliceWithout(arr, v1, v2, v3)
func SliceRemoveValue[T comparable, Slice ~[]T](arr Slice, value T) Slice {
	return SliceWithout(arr, value)
}

// SliceChunk 将一个切片分割成多个子切片
// size 参数指定每个子切片的大小
// 返回一个包含子切片的切片
func SliceChunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil // 如果 size <= 0，则返回 nil
	}

	var batches [][]T // 创建子切片切片
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice) // 确保不超出边界
		}
		batches = append(batches, slice[i:end]) // 切片而不复制
	}
	return batches
}

// SliceEqual 比较两个切片是否相等，支持任何类型
func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// InsertionSort 对小规模数组使用插入排序
func InsertionSort(arr []int) {
	if len(arr) < 2 {
		return
	}
	low := 0
	high := len(arr) - 1

	for currentIndex := low + 1; currentIndex <= high; currentIndex++ {
		key := arr[currentIndex]
		sortedIndex := currentIndex - 1
		for sortedIndex >= low && arr[sortedIndex] > key {
			arr[sortedIndex+1] = arr[sortedIndex]
			sortedIndex--
		}
		arr[sortedIndex+1] = key
	}
}

// QuickSort 实现快速排序算法
func QuickSort(arr []int, low, high int) {
	if low < high {
		// 获取分区索引
		pi := partition(arr, low, high)

		// 递归排序分区
		QuickSort(arr, low, pi-1)  // 排序基准左侧
		QuickSort(arr, pi+1, high) // 排序基准右侧
	}
}

// partition 进行分区操作
func partition(arr []int, low, high int) int {
	pivot := arr[high]     // 选择最后一个元素作为基准
	sortedIndex := low - 1 // 小于基准的元素的索引

	for currentIndex := low; currentIndex < high; currentIndex++ {
		if arr[currentIndex] < pivot { // 如果当前元素小于基准
			sortedIndex++                                                             // 增加小于基准的元素索引
			arr[sortedIndex], arr[currentIndex] = arr[currentIndex], arr[sortedIndex] // 交换
		}
	}
	// 将基准放到正确的位置
	arr[sortedIndex+1], arr[high] = arr[high], arr[sortedIndex+1]
	return sortedIndex + 1 // 返回基准的索引
}

// BubbleSort 实现冒泡排序算法
func BubbleSort(arr []int) {
	n := len(arr)

	for currentIndex := 0; currentIndex < n-1; currentIndex++ {
		for sortedIndex := 0; sortedIndex < n-currentIndex-1; sortedIndex++ {
			if arr[sortedIndex] > arr[sortedIndex+1] { // 如果当前元素大于下一个元素，则交换
				arr[sortedIndex], arr[sortedIndex+1] = arr[sortedIndex+1], arr[sortedIndex]
			}
		}
	}
}

// RepeatField 返回一个长度为 count 的切片，
// 切片中的每个元素都是传入的 field 值的副本
func RepeatField[T any](field T, count int) []T {
	s := make([]T, count) // 创建一个长度为 count 的切片，元素类型为 T
	for i := range s {
		s[i] = field // 将 field 赋值给切片的每个元素
	}
	return s // 返回填充好的切片
}

// SliceChain 是一个支持链式调用的泛型切片操作结构体，带并发安全锁保护
// 通过链式调用可以方便地对切片进行追加、去重、排序、过滤等操作
type SliceChain[T comparable] struct {
	mu   sync.RWMutex // 读写锁
	data []T
}

// FromSlice 根据普通切片创建一个新的 SliceChain 实例
// Params:
//   - slice: 需要转换成链式操作结构的普通切片，类型为 []T
//
// Returns:
//   - 返回一个指向新创建的 SliceChain 实例的指针，内部包含了传入的切片数据副本
func FromSlice[T comparable](slice []T) *SliceChain[T] {
	sc := &SliceChain[T]{}
	if len(slice) > 0 {
		sc.data = append(sc.data, slice...)
	}
	return sc
}

// Append 追加元素到当前切片，支持链式调用
// Params:
//   - elements: 可变参数，表示需要追加到当前切片中的元素，类型为 T
//
// Returns:
//   - 返回当前 SliceChain 实例指针，方便链式调用
//
// Examples:
//
//	sc.Append(1, 2, 3)
func (sc *SliceChain[T]) Append(elements ...T) *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data = append(sc.data, elements...)
	return sc
}

// Uniq 去重，移除切片中重复的元素，保持元素顺序，支持链式调用
// Returns:
//   - 返回当前 SliceChain 实例指针，去重后的数据保存在内部切片中
func (sc *SliceChain[T]) Uniq() *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data = SliceRemoveDuplicates(sc.data)
	return sc
}

// RemoveValue 移除切片中所有等于指定值的元素，支持链式调用
// Params：
//   - value: 需要从切片中移除的元素值，类型为 T
//
// Returns:
//   - 返回当前 SliceChain 实例指针，移除指定值后的数据保存在内部切片中
func (sc *SliceChain[T]) RemoveValue(value T) *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	n := 0
	for _, v := range sc.data {
		if v != value {
			sc.data[n] = v
			n++
		}
	}
	sc.data = sc.data[:n]
	return sc
}

// RemoveEmpty 移除“空值”元素，空值定义为元素等于类型零值，支持链式调用
// 适用于数字、字符串、指针等类型的零值判断
// Returns:
//   - 返回当前 SliceChain 实例指针，移除零值元素后的数据保存在内部切片中
func (sc *SliceChain[T]) RemoveEmpty() *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	result := sc.data[:0]
	for _, v := range sc.data {
		if !validator.IsCEmpty(v) {
			result = append(result, v)
		}
	}
	sc.data = result
	return sc
}

// Filter 根据传入的过滤函数 f，保留满足条件的元素，支持链式调用
// Params：
//   - f: 过滤函数，接收一个元素 T，返回 bool，返回 true 表示保留该元素，false 表示过滤掉
//
// Returns:
//   - 返回当前 SliceChain 实例指针，过滤后的数据保存在内部切片中
//
// Examples:
//
//	sc.Filter(func(x int) bool { return x%2 == 0 })
func (sc *SliceChain[T]) Filter(f func(T) bool) *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	result := sc.data[:0] // 利用切片复用内存，避免额外分配
	for _, v := range sc.data {
		if f(v) {
			result = append(result, v)
		}
	}
	sc.data = result
	return sc
}

// Sort 使用传入的 less 函数对切片进行排序，支持链式调用
// Params:
//   - less: 比较函数，接收两个元素 a, b，返回 bool，返回 true 表示 a < b
//
// Returns:
//   - 返回当前 SliceChain 实例指针，排序后的数据保存在内部切片中
//
// Examples:
//
//	sc.Sort(func(a, b int) bool { return a < b })
func (sc *SliceChain[T]) Sort(less func(a, b T) bool) *SliceChain[T] {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sort.Slice(sc.data, func(i, j int) bool {
		return less(sc.data[i], sc.data[j])
	})
	return sc
}

// Data 返回当前链式操作后的切片数据，方便与普通切片交互
// Returns:
//   - 返回当前内部切片的引用，类型为 []T
//
// 注意：返回的是内部切片的引用，修改返回值会影响 SliceChain 内部数据
func (sc *SliceChain[T]) Data() []T {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return IF(len(sc.data) > 0, sc.data, []T{}) // 返回非nil空切片，避免断言错误
}

// String 实现 fmt.Stringer 接口，方便打印 SliceChain 内容
// Returns:
//   - 返回当前切片数据的字符串表示
func (sc *SliceChain[T]) String() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return fmt.Sprintf("%v", sc.data)
}

// TransformSlice 将切片的每个元素通过函数转换为新类型
// 使用泛型支持任意类型转换，如 []int -> []string
func TransformSlice[T any, R any](slice []T, transform func(T) R) []R {
	if len(slice) == 0 {
		return []R{}
	}
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = transform(v)
	}
	return result
}

// FilterSlice 过滤切片，保留满足条件的元素
// 与 SliceRemove 的区别：FilterSlice 保留满足条件的，SliceRemove 也是保留满足条件的
// 这个函数提供更直观的命名
func FilterSlice[T any](slice []T, predicate func(T) bool) []T {
	if len(slice) == 0 {
		return slice
	}
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// TransformAndFilterSlice 组合转换和过滤操作
//
// Deprecated: 使用 SliceFilterMap 替代，单次回调更简洁，避免中间切片分配
// 迁移示例：
//
//	// 旧：
//	result := TransformAndFilterSlice(slice,
//	    func(v T) R { return transform(v) },
//	    func(r R) bool { return r != zero })
//	// 新：
//	result := SliceFilterMap(slice, func(v T, _ int) (R, bool) {
//	    r := transform(v)
//	    return r, r != zero
//	})
//
// 注意：SliceFilterMap 回调接收 (item, index)，TransformAndFilterSlice 只接收 item
func TransformAndFilterSlice[T any, R any](slice []T, transform func(T) R, predicate func(R) bool) []R {
	if len(slice) == 0 {
		return []R{}
	}
	return SliceFilterMap(slice, func(v T, _ int) (R, bool) {
		r := transform(v)
		return r, predicate(r)
	})
}

// TransformAndCompactSlice 转换切片并移除零值
// 这是常见的模式：转换后过滤掉零值元素
//
// 参数:
//   - slice: 输入切片
//   - transform: 转换函数
//
// 返回:
//   - 转换后移除零值的切片
//
// 示例:
//
//	clients := []*Client{...}
//	ips := TransformAndCompactSlice(clients, func(c *Client) string { return getIP(c) })
func TransformAndCompactSlice[T any, R comparable](slice []T, transform func(T) R) []R {
	var zero R
	return TransformAndFilterSlice(slice, transform, func(r R) bool {
		return r != zero
	})
}

// ReduceSlice 将切片归约为单个值
func ReduceSlice[T any, R any](slice []T, initial R, accumulator func(R, T) R) R {
	result := initial
	for _, v := range slice {
		result = accumulator(result, v)
	}
	return result
}

// FlattenSlice 扁平化嵌套切片
func FlattenSlice[T any](slices [][]T) []T {
	if len(slices) == 0 {
		return []T{}
	}

	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// ReverseSlice 反转切片（返回新切片）
func ReverseSlice[T any](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// ReverseSliceInPlace 原地反转切片
func ReverseSliceInPlace[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// TakeSlice 获取前 n 个元素
func TakeSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}
	return slice[:n]
}

// TakeLastSlice 获取后 n 个元素
func TakeLastSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}
	return slice[len(slice)-n:]
}

// SkipSlice 跳过前 n 个元素
func SkipSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return slice
	}
	if n >= len(slice) {
		return []T{}
	}
	return slice[n:]
}

// SkipLastSlice 跳过后 n 个元素
func SkipLastSlice[T any](slice []T, n int) []T {
	if n <= 0 {
		return slice
	}
	if n >= len(slice) {
		return []T{}
	}
	return slice[:len(slice)-n]
}

// CompactSlice 移除切片中的零值元素
func CompactSlice[T comparable](slice []T) []T {
	var zero T
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != zero {
			result = append(result, v)
		}
	}
	return result
}

// PartitionSlice 将切片分为两部分：满足条件的和不满足条件的
func PartitionSlice[T any](slice []T, predicate func(T) bool) (truthy, falsy []T) {
	truthy = make([]T, 0)
	falsy = make([]T, 0)
	for _, v := range slice {
		if predicate(v) {
			truthy = append(truthy, v)
		} else {
			falsy = append(falsy, v)
		}
	}
	return
}

// GroupSliceBy 按照键函数对切片元素分组
func GroupSliceBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// FindSlice 查找第一个满足条件的元素
func FindSlice[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindLastSlice 查找最后一个满足条件的元素
func FindLastSlice[T any](slice []T, predicate func(T) bool) (T, bool) {
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(slice[i]) {
			return slice[i], true
		}
	}
	var zero T
	return zero, false
}

// AllSlice 检查是否所有元素都满足条件
func AllSlice[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// AnySlice 检查是否有任何元素满足条件
func AnySlice[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if predicate(v) {
			return true
		}
	}
	return false
}

// NoneSlice 检查是否没有元素满足条件
func NoneSlice[T any](slice []T, predicate func(T) bool) bool {
	return !AnySlice(slice, predicate)
}

// CountSlice 计算满足条件的元素数量
//
// Deprecated: 使用 SliceCountBy 替代，命名风格统一为 SliceXxx
// 迁移示例：
//
//	// 旧：
//	count := CountSlice(slice, func(v T) bool { return v > 0 })
//	// 新：
//	count := SliceCountBy(slice, func(v T) bool { return v > 0 })
func CountSlice[T any](slice []T, predicate func(T) bool) int {
	return SliceCountBy(slice, predicate)
}

// IndexOfSlice 返回元素在切片中的索引，不存在返回 -1
func IndexOfSlice[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// LastIndexOfSlice 返回元素在切片中最后一次出现的索引，不存在返回 -1
func LastIndexOfSlice[T comparable](slice []T, item T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

// RemoveSliceAt 移除指定索引的元素
func RemoveSliceAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// ContainsAnySlice 检查切片是否包含任意一个指定元素
func ContainsAnySlice[T comparable](slice []T, items ...T) bool {
	for _, item := range items {
		if SliceContainsComparable(slice, item) {
			return true
		}
	}
	return false
}

// ContainsAllSlice 检查切片是否包含所有指定元素
func ContainsAllSlice[T comparable](slice []T, items ...T) bool {
	for _, item := range items {
		if !SliceContainsComparable(slice, item) {
			return false
		}
	}
	return true
}

// UniqueSliceBy 使用自定义键函数去重
// keyFunc 用于提取每个元素的唯一标识
func UniqueSliceBy[T any, K comparable](slice []T, keyFunc func(T) K) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[K]struct{}, len(slice))
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		key := keyFunc(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// AddUniqueSlice 向切片中添加元素（如果不存在）
// 返回新切片和是否实际添加了元素
func AddUniqueSlice[T comparable](slice []T, items ...T) ([]T, bool) {
	if len(items) == 0 {
		return slice, false
	}

	// 构建已存在元素的集合
	existing := make(map[T]struct{}, len(slice))
	for _, item := range slice {
		existing[item] = struct{}{}
	}

	// 添加新元素
	added := false
	for _, item := range items {
		if _, exists := existing[item]; !exists {
			slice = append(slice, item)
			existing[item] = struct{}{}
			added = true
		}
	}

	return slice, added
}

// EqualUnorderedSlice 检查两个切片是否包含相同元素（忽略顺序）
//
// Deprecated: 使用 SliceElementsMatch 替代，命名更清晰（elements match）
// 迁移示例：
//
//	// 旧：
//	equal := EqualUnorderedSlice(a, b)
//	// 新：
//	equal := SliceElementsMatch(a, b)
func EqualUnorderedSlice[T comparable](a, b []T) bool {
	return SliceElementsMatch(a, b)
}

// DedupeValues 从切片中按键提取去重后的值列表
// key 函数返回 (值, 是否包含)；ok=false 时跳过该项（常用于过滤零值）
func DedupeValues[T any, K comparable](items []T, key func(T) (K, bool)) []K {
	if len(items) == 0 {
		return nil
	}
	set := make(map[K]struct{}, len(items))
	result := make([]K, 0, len(items))
	for _, item := range items {
		k, ok := key(item)
		if !ok {
			continue
		}
		if _, exists := set[k]; exists {
			continue
		}
		set[k] = struct{}{}
		result = append(result, k)
	}
	return result
}

// ============================================================================
// 转换 + 去重/展平 组合优化（避免中间切片分配）
// ============================================================================

// SliceUniqMap 转换并去重，一次遍历完成 Map+Uniq
// 相比分别调用 TransformSlice + SliceRemoveDuplicates 减少一次中间切片分配
func SliceUniqMap[T any, R comparable](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, 0, len(collection))
	seen := make(map[R]struct{}, len(collection))
	for i := range collection {
		r := iteratee(collection[i], i)
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			result = append(result, r)
		}
	}
	return result
}

// SliceFlatMap 映射并展平，一次遍历完成 Map+Flatten
// iteratee 返回切片，所有返回切片被展平到一个结果切片
func SliceFlatMap[T, R any](collection []T, iteratee func(item T, index int) []R) []R {
	result := make([]R, 0, len(collection))
	for i := range collection {
		result = append(result, iteratee(collection[i], i)...)
	}
	return result
}

// SliceFilterMap 过滤+映射组合
// 回调返回 (映射值, 是否保留)，与 TransformAndFilterSlice 等价但参数顺序更直观
func SliceFilterMap[T, R any](collection []T, callback func(item T, index int) (R, bool)) []R {
	result := make([]R, 0, len(collection))
	for i := range collection {
		if r, ok := callback(collection[i], i); ok {
			result = append(result, r)
		}
	}
	return result
}

// ============================================================================
// 遍历控制
// ============================================================================

// SliceForEachWhile 可中断遍历
// iteratee 返回 false 时中断遍历（类似 do-while）
func SliceForEachWhile[T any](collection []T, iteratee func(item T, index int) bool) {
	for i := range collection {
		if !iteratee(collection[i], i) {
			break
		}
	}
}

// SliceTimes 生成 N 个元素
// 调用 iteratee n 次，返回结果切片（iteratee 接收索引作为参数）
func SliceTimes[T any](count int, iteratee func(index int) T) []T {
	result := make([]T, count)
	for i := 0; i < count; i++ {
		result[i] = iteratee(i)
	}
	return result
}

// SliceInterleave 交替合并多个切片
// round-robin 交替从每个切片取元素，依次追加到结果
// 例：[1,2,3], [4,5], [6,7,8,9] => [1,4,6,2,5,7,3,8,9]
func SliceInterleave[T any, Slice ~[]T](collections ...Slice) Slice {
	if len(collections) == 0 {
		return Slice{}
	}
	maxSize := 0
	totalSize := 0
	for i := range collections {
		size := len(collections[i])
		totalSize += size
		if size > maxSize {
			maxSize = size
		}
	}
	if maxSize == 0 {
		return Slice{}
	}
	result := make(Slice, 0, totalSize)
	for i := 0; i < maxSize; i++ {
		for j := range collections {
			if len(collections[j])-1 < i {
				continue
			}
			result = append(result, collections[j][i])
		}
	}
	return result
}

// ============================================================================
// 切片 ↔ map 互转
// ============================================================================

// SliceToMap 切片转 map（KeyBy）
// iteratee 返回 key，value 为原元素；同 key 后者覆盖前者
func SliceToMap[T any, K comparable](collection []T, iteratee func(item T) K) map[K]T {
	result := make(map[K]T, len(collection))
	for i := range collection {
		result[iteratee(collection[i])] = collection[i]
	}
	return result
}

// SliceFilterToMap 过滤后转 map
// 回调返回 (key, value, 是否保留)，仅保留的元素进入 map
func SliceFilterToMap[T any, K comparable, V any](collection []T, transform func(item T, index int) (K, V, bool)) map[K]V {
	result := make(map[K]V, len(collection))
	for i := range collection {
		k, v, ok := transform(collection[i], i)
		if ok {
			result[k] = v
		}
	}
	return result
}

// ============================================================================
// 条件丢弃
// ============================================================================

// SliceDropWhile 从开头丢弃满足条件的元素，直到遇到第一个不满足的
func SliceDropWhile[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	i := 0
	for ; i < len(collection); i++ {
		if !predicate(collection[i]) {
			break
		}
	}
	return collection[i:]
}

// SliceDropRightWhile 从末尾丢弃满足条件的元素，直到遇到第一个不满足的
func SliceDropRightWhile[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	i := len(collection) - 1
	for ; i >= 0; i-- {
		if !predicate(collection[i]) {
			break
		}
	}
	return collection[:i+1]
}

// ============================================================================
// 反向过滤 / 计数
// ============================================================================

// SliceReject 反向过滤，保留不满足条件的元素（与 FilterSlice 相反）
func SliceReject[T any, Slice ~[]T](collection Slice, predicate func(item T, index int) bool) Slice {
	result := make(Slice, 0, len(collection))
	for i := range collection {
		if !predicate(collection[i], i) {
			result = append(result, collection[i])
		}
	}
	return result
}

// SliceCountBy 按条件计数
func SliceCountBy[T any](collection []T, predicate func(item T) bool) int {
	count := 0
	for i := range collection {
		if predicate(collection[i]) {
			count++
		}
	}
	return count
}

// SliceCountValues 按值计数，返回 map[值]出现次数
func SliceCountValues[T comparable](collection []T) map[T]int {
	result := make(map[T]int, len(collection))
	for i := range collection {
		result[collection[i]]++
	}
	return result
}

// SliceCountValuesBy 按键计数，等价于 Map + CountValues
func SliceCountValuesBy[T any, U comparable](collection []T, mapper func(item T) U) map[U]int {
	result := make(map[U]int, len(collection))
	for i := range collection {
		result[mapper(collection[i])]++
	}
	return result
}

// ============================================================================
// 子集 / 替换 / 有序检查
// ============================================================================

// SliceSubset 返回从 offset 开始最多 length 个元素的子切片
// offset 为负表示从末尾倒数；不 panic，越界返回空切片
func SliceSubset[T any, Slice ~[]T](collection Slice, offset int, length uint) Slice {
	size := len(collection)
	if offset < 0 {
		offset = size + offset
		if offset < 0 {
			offset = 0
		}
	}
	if offset > size {
		return Slice{}
	}
	remain := uint(size - offset)
	if length > remain {
		length = remain
	}
	return collection[offset : offset+int(length)]
}

// SliceReplace 返回副本，将前 n 个 old 替换为 nEw（n<0 表示全部替换）
func SliceReplace[T comparable, Slice ~[]T](collection Slice, old, nEw T, n int) Slice {
	result := make(Slice, len(collection))
	copy(result, collection)
	for i := range result {
		if result[i] == old && n != 0 {
			result[i] = nEw
			n--
		}
	}
	return result
}

// SliceReplaceAll 返回副本，将所有 old 替换为 nEw
func SliceReplaceAll[T comparable, Slice ~[]T](collection Slice, old, nEw T) Slice {
	return SliceReplace(collection, old, nEw, -1)
}

// SliceIsSorted 检查切片是否按升序排列
func SliceIsSorted[T types.Ordered](collection []T) bool {
	for i := 1; i < len(collection); i++ {
		if collection[i-1] > collection[i] {
			return false
		}
	}
	return true
}

// SliceIsSortedByKey 检查切片是否按键函数升序排列
func SliceIsSortedByKey[T any, K types.Ordered](collection []T, iteratee func(item T) K) bool {
	size := len(collection)
	if size <= 1 {
		return true
	}
	prev := iteratee(collection[0])
	for i := 1; i < size; i++ {
		cur := iteratee(collection[i])
		if prev > cur {
			return false
		}
		prev = cur
	}
	return true
}

// ============================================================================
// 切割 / 前后缀
// ============================================================================

// SliceCut 围绕第一个 separator 切割，返回前/后两段及是否找到
// separator 为空时返回 (空, collection, true)
func SliceCut[T comparable, Slice ~[]T](collection, separator Slice) (before, after Slice, found bool) {
	if len(separator) == 0 {
		return Slice{}, collection, true
	}
	for i := 0; i+len(separator) <= len(collection); i++ {
		match := true
		for j := 0; j < len(separator); j++ {
			if collection[i+j] != separator[j] {
				match = false
				break
			}
		}
		if match {
			return collection[:i], collection[i+len(separator):], true
		}
	}
	return collection, Slice{}, false
}

// SliceHasPrefix 检查切片是否以 prefix 开头
func SliceHasPrefix[T comparable](collection, prefix []T) bool {
	if len(collection) < len(prefix) {
		return false
	}
	for i := range prefix {
		if collection[i] != prefix[i] {
			return false
		}
	}
	return true
}

// SliceHasSuffix 检查切片是否以 suffix 结尾
func SliceHasSuffix[T comparable](collection, suffix []T) bool {
	if len(collection) < len(suffix) {
		return false
	}
	start := len(collection) - len(suffix)
	for i := range suffix {
		if collection[start+i] != suffix[i] {
			return false
		}
	}
	return true
}

// ============================================================================
// 唯一 / 重复查找
// ============================================================================

// SliceFindUniques 返回只出现一次的元素（保持原顺序）
func SliceFindUniques[T comparable, Slice ~[]T](collection Slice) Slice {
	isDup := make(map[T]bool, len(collection))
	for i := range collection {
		if existed, ok := isDup[collection[i]]; ok {
			if !existed {
				isDup[collection[i]] = true
			}
		} else {
			isDup[collection[i]] = false
		}
	}
	result := make(Slice, 0, len(isDup))
	for i := range collection {
		if !isDup[collection[i]] {
			result = append(result, collection[i])
		}
	}
	return result
}

// SliceFindUniquesBy 按键函数返回只出现一次的元素
func SliceFindUniquesBy[T any, U comparable, Slice ~[]T](collection Slice, iteratee func(item T) U) Slice {
	isDup := make(map[U]bool, len(collection))
	keys := make([]U, len(collection))
	for i := range collection {
		key := iteratee(collection[i])
		keys[i] = key
		if existed, ok := isDup[key]; ok {
			if !existed {
				isDup[key] = true
			}
		} else {
			isDup[key] = false
		}
	}
	result := make(Slice, 0, len(isDup))
	for i := range collection {
		if !isDup[keys[i]] {
			result = append(result, collection[i])
		}
	}
	return result
}

// SliceFindDuplicates 返回每个重复元素的首次出现（保持原顺序）
func SliceFindDuplicates[T comparable, Slice ~[]T](collection Slice) Slice {
	isDup := make(map[T]bool, len(collection))
	for i := range collection {
		if existed, ok := isDup[collection[i]]; ok {
			if !existed {
				isDup[collection[i]] = true
			}
		} else {
			isDup[collection[i]] = false
		}
	}
	result := make(Slice, 0, len(isDup))
	for i := range collection {
		if isDup[collection[i]] {
			result = append(result, collection[i])
			isDup[collection[i]] = false
		}
	}
	return result
}

// SliceFindDuplicatesBy 按键函数返回每个重复元素的首次出现
func SliceFindDuplicatesBy[T any, U comparable, Slice ~[]T](collection Slice, iteratee func(item T) U) Slice {
	isDup := make(map[U]bool, len(collection))
	keys := make([]U, len(collection))
	for i := range collection {
		key := iteratee(collection[i])
		keys[i] = key
		if existed, ok := isDup[key]; ok {
			if !existed {
				isDup[key] = true
			}
		} else {
			isDup[key] = false
		}
	}
	result := make(Slice, 0, len(isDup))
	for i := range collection {
		if isDup[keys[i]] {
			result = append(result, collection[i])
			isDup[keys[i]] = false
		}
	}
	return result
}

// ============================================================================
// 自定义比较的 Min/Max
// ============================================================================

// SliceMinBy 自定义比较的最小值，comparison(a,b) 返回 true 表示 a<b
func SliceMinBy[T any](collection []T, comparison func(a, b T) bool) T {
	var min T
	if len(collection) == 0 {
		return min
	}
	min = collection[0]
	for i := 1; i < len(collection); i++ {
		if comparison(collection[i], min) {
			min = collection[i]
		}
	}
	return min
}

// SliceMaxBy 自定义比较的最大值，comparison(a,b) 返回 true 表示 a>b
func SliceMaxBy[T any](collection []T, comparison func(a, b T) bool) T {
	var max T
	if len(collection) == 0 {
		return max
	}
	max = collection[0]
	for i := 1; i < len(collection); i++ {
		if comparison(collection[i], max) {
			max = collection[i]
		}
	}
	return max
}

// SliceMinIndexBy 自定义比较的最小值及其索引
func SliceMinIndexBy[T any](collection []T, comparison func(a, b T) bool) (T, int) {
	var min T
	if len(collection) == 0 {
		return min, -1
	}
	min = collection[0]
	idx := 0
	for i := 1; i < len(collection); i++ {
		if comparison(collection[i], min) {
			min = collection[i]
			idx = i
		}
	}
	return min, idx
}

// SliceMaxIndexBy 自定义比较的最大值及其索引
func SliceMaxIndexBy[T any](collection []T, comparison func(a, b T) bool) (T, int) {
	var max T
	if len(collection) == 0 {
		return max, -1
	}
	max = collection[0]
	idx := 0
	for i := 1; i < len(collection); i++ {
		if comparison(collection[i], max) {
			max = collection[i]
			idx = i
		}
	}
	return max, idx
}

// SliceMinOrdered 返回 Ordered 类型切片的最小值（空切片返回零值）
func SliceMinOrdered[T types.Ordered](collection []T) T {
	var min T
	if len(collection) == 0 {
		return min
	}
	min = collection[0]
	for i := 1; i < len(collection); i++ {
		if cmp.Compare(collection[i], min) < 0 {
			min = collection[i]
		}
	}
	return min
}

// SliceMaxOrdered 返回 Ordered 类型切片的最大值（空切片返回零值）
func SliceMaxOrdered[T types.Ordered](collection []T) T {
	var max T
	if len(collection) == 0 {
		return max
	}
	max = collection[0]
	for i := 1; i < len(collection); i++ {
		if cmp.Compare(collection[i], max) > 0 {
			max = collection[i]
		}
	}
	return max
}

// ============================================================================
// 首尾 / 第 N 个元素
// ============================================================================

// SliceFirst 返回首个元素及是否存在
func SliceFirst[T any](collection []T) (T, bool) {
	if len(collection) == 0 {
		var t T
		return t, false
	}
	return collection[0], true
}

// SliceLast 返回末尾元素及是否存在
func SliceLast[T any](collection []T) (T, bool) {
	if len(collection) == 0 {
		var t T
		return t, false
	}
	return collection[len(collection)-1], true
}

// SliceFirstOr 返回首个元素或 fallback
func SliceFirstOr[T any](collection []T, fallback T) T {
	if len(collection) == 0 {
		return fallback
	}
	return collection[0]
}

// SliceLastOr 返回末尾元素或 fallback
func SliceLastOr[T any](collection []T, fallback T) T {
	if len(collection) == 0 {
		return fallback
	}
	return collection[len(collection)-1]
}

// SliceNth 返回第 nth 个元素，nth 为负表示从末尾倒数
// 越界返回错误
func SliceNth[T any](collection []T, nth int) (T, error) {
	l := len(collection)
	var t T
	if nth >= l || -nth > l {
		return t, fmt.Errorf("nth: %d out of slice bounds", nth)
	}
	if nth >= 0 {
		return collection[nth], nil
	}
	return collection[l+nth], nil
}

// SliceNthOr 返回第 nth 个元素或 fallback（越界不报错）
func SliceNthOr[T any](collection []T, nth int, fallback T) T {
	l := len(collection)
	if nth >= l || -nth > l {
		return fallback
	}
	if nth >= 0 {
		return collection[nth]
	}
	return collection[l+nth]
}

// ============================================================================
// 双向差集 / 元素匹配
// ============================================================================

// SliceDifference 双向差集，返回 (list1有而list2无, list2有而list1无)
func SliceDifference[T comparable, Slice ~[]T](list1, list2 Slice) (Slice, Slice) {
	seenLeft := make(map[T]struct{}, len(list1))
	seenRight := make(map[T]struct{}, len(list2))
	for i := range list1 {
		seenLeft[list1[i]] = struct{}{}
	}
	for i := range list2 {
		seenRight[list2[i]] = struct{}{}
	}
	left := make(Slice, 0)
	right := make(Slice, 0)
	for i := range list1 {
		if _, ok := seenRight[list1[i]]; !ok {
			left = append(left, list1[i])
		}
	}
	for i := range list2 {
		if _, ok := seenLeft[list2[i]]; !ok {
			right = append(right, list2[i])
		}
	}
	return left, right
}

// SliceUnionMulti 多个切片的并集（去重，保持首次出现顺序）
// 与 SliceUnion 区别：支持变参多个切片
func SliceUnionMulti[T comparable, Slice ~[]T](lists ...Slice) Slice {
	var capLen int
	for _, list := range lists {
		capLen += len(list)
	}
	result := make(Slice, 0, capLen)
	seen := make(map[T]struct{}, capLen)
	for i := range lists {
		for j := range lists[i] {
			if _, ok := seen[lists[i][j]]; !ok {
				seen[lists[i][j]] = struct{}{}
				result = append(result, lists[i][j])
			}
		}
	}
	return result
}

// SliceElementsMatch 检查两切片是否包含相同元素集合（忽略顺序，含重复计数）
func SliceElementsMatch[T comparable](list1, list2 []T) bool {
	if len(list1) != len(list2) {
		return false
	}
	if len(list1) == 0 {
		return true
	}
	counters := make(map[T]int, len(list1))
	for i := range list1 {
		counters[list1[i]]++
	}
	for i := range list2 {
		counters[list2[i]]--
	}
	for _, count := range counters {
		if count != 0 {
			return false
		}
	}
	return true
}

// SliceWithoutBy 按键排除元素
// 排除 iteratee(item) 结果在 exclude 列表中的元素
func SliceWithoutBy[T any, K comparable, Slice ~[]T](collection Slice, iteratee func(item T) K, exclude ...K) Slice {
	excludeMap := make(map[K]struct{}, len(exclude))
	for _, e := range exclude {
		excludeMap[e] = struct{}{}
	}
	result := make(Slice, 0, len(collection))
	for i := range collection {
		if _, ok := excludeMap[iteratee(collection[i])]; !ok {
			result = append(result, collection[i])
		}
	}
	return result
}

// SlicePartitionBy 按键函数分区，返回各分区切片（保持首次出现顺序）
// 与 PartitionSlice 区别：PartitionBy 返回 []Slice，键由 iteratee 决定
func SlicePartitionBy[T any, K comparable, Slice ~[]T](collection Slice, iteratee func(item T) K) []Slice {
	result := []Slice{}
	seen := map[K]int{}
	for i := range collection {
		key := iteratee(collection[i])
		idx, ok := seen[key]
		if !ok {
			idx = len(result)
			seen[key] = idx
			result = append(result, Slice{})
		}
		result[idx] = append(result[idx], collection[i])
	}
	return result
}

// SliceGroupByMap 按键函数分组，返回 map[key][]value
// 与 GroupSliceBy 区别：value 类型可与原元素不同（iteratee 同时提取 key 和 value）
func SliceGroupByMap[T any, K comparable, V any](collection []T, iteratee func(item T) (K, V)) map[K][]V {
	result := make(map[K][]V, len(collection))
	for i := range collection {
		k, v := iteratee(collection[i])
		result[k] = append(result[k], v)
	}
	return result
}
