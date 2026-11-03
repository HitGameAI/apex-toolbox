/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-06-16 21:38:23
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-06-16 21:38:23
 * @FilePath: \go-toolbox\pkg\pbmo\copy_plan.go
 * @Description: 分组拷贝计划 - 同尺寸字段统一聚合为移动组，双侧连续段单次 memmove 覆盖
 * 整数转换与时间互转保持静态泛型实例分组，偏移烘焙进闭包免切片循环
 * 产生堆分配的字段聚合为复合分配组，StructOf 合成复合类型单次分配覆盖全部子对象
 * 不可分组种类保留独立闭包走 singles 通路，全程字段写入互不重叠顺序无关
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"sort"
	"strconv"
	"time"
	"unsafe"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// groupCopyFunc 组拷贝体，连续段与零散偏移对全部烘焙进闭包
// 整组仅一次间接调用，调用点免切片加载与边界检查
type groupCopyFunc func(dstBase, srcBase unsafe.Pointer)

// pureFieldCopy 纯转换组单字段拷贝体，偏移以参数传入
// 参数序与 convertValue 一致（dstOff 在前），泛型转换核可直接实例化复用
// 分组执行器将偏移烘焙进闭包后以立即数调用，免切片加载与边界检查
type pureFieldCopy func(dstBase, srcBase unsafe.Pointer, dstOff, srcOff uintptr)

// copyGroup 同拷贝体字段分组
type copyGroup struct {
	copy groupCopyFunc
}

// groupSrc 分组分类所需的字段视图，统一 fieldMappingEntry 与 structFieldEntry
type groupSrc struct {
	kind      fieldKind
	srcType   reflect.Type
	dstType   reflect.Type
	srcOff    uintptr
	dstOff    uintptr
	groupable bool          // 拷贝体是否为可共享的纯拷贝，transform 等自定义闭包为 false
	legacy    fieldCopyFunc // 不可分组时回退的独立闭包
}

// ══════════════════════════════════════════════════════════════════════════════
// 泛型拷贝核：每种拷贝语义唯一实现
// 泛型按类型实例化为静态代码，直拷闭包静态调用可被内联，组拷贝体循环内联同一语义
// ══════════════════════════════════════════════════════════════════════════════

// copyValue 同类型字段标量拷贝
func copyValue[T any](dstBase, srcBase unsafe.Pointer, dstOff, srcOff uintptr) {
	*(*T)(addPtr(dstBase, dstOff)) = *(*T)(addPtr(srcBase, srcOff))
}

// intKindConv 整数种类约束，覆盖全部有符号与无符号整数
type intKindConv interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// convertValue 整数转换标量
func convertValue[S, D intKindConv](dstBase, srcBase unsafe.Pointer, dstOff, srcOff uintptr) {
	*(*D)(addPtr(dstBase, dstOff)) = D(*(*S)(addPtr(srcBase, srcOff)))
}

// makeIntCopy 生成整数转换的单态闭包，偏移烘焙进闭包免切片传参
func makeIntCopy[S, D intKindConv](srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		convertValue[S, D](dstBase, srcBase, dstOffset, srcOffset)
	}
}

// intCopyFromSrc 按源种类分派（第一级），标准整数返回单态闭包，非标准返回 nil
func intCopyFromSrc(srcKind, dstKind reflect.Kind, srcOffset, dstOffset uintptr) fieldCopyFunc {
	switch srcKind {
	case reflect.Int:
		return intCopyToDst[int](dstKind, srcOffset, dstOffset)
	case reflect.Int8:
		return intCopyToDst[int8](dstKind, srcOffset, dstOffset)
	case reflect.Int16:
		return intCopyToDst[int16](dstKind, srcOffset, dstOffset)
	case reflect.Int32:
		return intCopyToDst[int32](dstKind, srcOffset, dstOffset)
	case reflect.Int64:
		return intCopyToDst[int64](dstKind, srcOffset, dstOffset)
	case reflect.Uint:
		return intCopyToDst[uint](dstKind, srcOffset, dstOffset)
	case reflect.Uint8:
		return intCopyToDst[uint8](dstKind, srcOffset, dstOffset)
	case reflect.Uint16:
		return intCopyToDst[uint16](dstKind, srcOffset, dstOffset)
	case reflect.Uint32:
		return intCopyToDst[uint32](dstKind, srcOffset, dstOffset)
	case reflect.Uint64:
		return intCopyToDst[uint64](dstKind, srcOffset, dstOffset)
	}
	return nil
}

// intCopyToDst 按目标种类分派（第二级），生成对应类型对的单态闭包
func intCopyToDst[S intKindConv](dstKind reflect.Kind, srcOffset, dstOffset uintptr) fieldCopyFunc {
	switch dstKind {
	case reflect.Int:
		return makeIntCopy[S, int](srcOffset, dstOffset)
	case reflect.Int8:
		return makeIntCopy[S, int8](srcOffset, dstOffset)
	case reflect.Int16:
		return makeIntCopy[S, int16](srcOffset, dstOffset)
	case reflect.Int32:
		return makeIntCopy[S, int32](srcOffset, dstOffset)
	case reflect.Int64:
		return makeIntCopy[S, int64](srcOffset, dstOffset)
	case reflect.Uint:
		return makeIntCopy[S, uint](srcOffset, dstOffset)
	case reflect.Uint8:
		return makeIntCopy[S, uint8](srcOffset, dstOffset)
	case reflect.Uint16:
		return makeIntCopy[S, uint16](srcOffset, dstOffset)
	case reflect.Uint32:
		return makeIntCopy[S, uint32](srcOffset, dstOffset)
	case reflect.Uint64:
		return makeIntCopy[S, uint64](srcOffset, dstOffset)
	}
	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// 时间转换核：四种方向各一个语义函数，组拷贝与直拷闭包共用
// ══════════════════════════════════════════════════════════════════════════════

// timeFromTS 解引用时间戳为 time.Time，nil 返回 false
func timeFromTS(ts *timestamppb.Timestamp) (time.Time, bool) {
	if ts == nil {
		return time.Time{}, false
	}
	return ts.AsTime(), true
}

// tsFromTime 零值时间返回 nil，其余转换为时间戳
func tsFromTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// timePtrFromTS 时间戳转 *time.Time，nil 时间戳返回 nil
func timePtrFromTS(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// tsFromTimePtr nil 或零值的 *time.Time 返回 nil，其余转换为时间戳
func tsFromTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// allocSlice 同布局切片分配核，泛型按元素类型静态实例化
// make 静态分配替代 reflect.MakeSlice，省去标志位检查与 reflect 包装开销
// 命名切片类型与底层种类内存布局一致，写入目标字段头后语义等价
func allocSlice[T any](n int) sliceHeader {
	s := make([]T, n)
	return sliceHeader{Data: unsafe.Pointer(unsafe.SliceData(s)), Len: n, Cap: n}
}

// ptrSliceAlloc 指针元素切片分配核
// 单指针布局（Ptr/Map/Chan/Func/UnsafePointer）统一以 []unsafe.Pointer 分配
// Go 1.25 reflect.MakeSlice 的 Value 携带堆逃逸的 slice 头，单次调用产生两笔分配
// make 分配即黑，写入黑色对象指针到黑色数组无需屏障，GC 语义安全
func ptrSliceAlloc(n int) sliceHeader {
	s := make([]unsafe.Pointer, n)
	return sliceHeader{Data: unsafe.Pointer(unsafe.SliceData(s)), Len: n, Cap: n}
}

// scalarSliceAlloc 按元素种类返回切片分配闭包，未覆盖种类返回 nil 回退 reflect
func scalarSliceAlloc(elem reflect.Type) func(int) sliceHeader {
	switch elem.Kind() {
	case reflect.Bool:
		return allocSlice[bool]
	case reflect.Int:
		return allocSlice[int]
	case reflect.Int8:
		return allocSlice[int8]
	case reflect.Int16:
		return allocSlice[int16]
	case reflect.Int32:
		return allocSlice[int32]
	case reflect.Int64:
		return allocSlice[int64]
	case reflect.Uint:
		return allocSlice[uint]
	case reflect.Uint8:
		return allocSlice[uint8]
	case reflect.Uint16:
		return allocSlice[uint16]
	case reflect.Uint32:
		return allocSlice[uint32]
	case reflect.Uint64:
		return allocSlice[uint64]
	case reflect.Float32:
		return allocSlice[float32]
	case reflect.Float64:
		return allocSlice[float64]
	case reflect.String:
		return allocSlice[string]
	case reflect.Ptr, reflect.UnsafePointer, reflect.Map, reflect.Chan, reflect.Func:
		return ptrSliceAlloc
	}
	return nil
}

// moveRun 双侧同步连续段：段内相邻字段的源侧与目标侧偏移同步按字段尺寸递增
// 连续段单次 memmove 覆盖多字段拷贝，宽段触发向量化整块搬运
type moveRun struct {
	srcOff uint32
	dstOff uint32
	bytes  uint32
}

// moveRunMinBytes 成段搬运的最小字节数，低于该值拆散为逐字段零散处理
// 短段 memmove 调用开销高于逐字段常量宽度拷贝，成段收益随段宽增长
const moveRunMinBytes = 64

// extractMoveRuns 按源偏移排序抽取双侧同步连续段，返回段列表与剩余零散偏移对
// pairs 为平铺 (srcOff, dstOff) 序列，段内偏移差恒定保证整段搬运语义等价逐字段拷贝
func extractMoveRuns(pairs []uintptr, sz uintptr) ([]moveRun, []uintptr) {
	type moveCand struct{ srcOff, dstOff uint32 }
	cands := make([]moveCand, len(pairs)/2)
	for i := range cands {
		cands[i] = moveCand{uint32(pairs[2*i]), uint32(pairs[2*i+1])}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].srcOff < cands[j].srcOff })
	var runs []moveRun
	rest := make([]uintptr, 0, len(pairs))
	stride := uint32(sz)
	for i := 0; i < len(cands); {
		j := i + 1
		for j < len(cands) && cands[j].srcOff == cands[j-1].srcOff+stride &&
			cands[j].dstOff == cands[j-1].dstOff+stride {
			j++
		}
		if n := uint32(j - i); n > 1 && n*stride >= moveRunMinBytes {
			runs = append(runs, moveRun{cands[i].srcOff, cands[i].dstOff, n * stride})
		} else {
			for k := i; k < j; k++ {
				rest = append(rest, uintptr(cands[k].srcOff), uintptr(cands[k].dstOff))
			}
		}
		i = j
	}
	return runs, rest
}

// moveRuns 逐段整块搬运连续段
func moveRuns(dstBase, srcBase unsafe.Pointer, runs []moveRun) {
	for i := range runs {
		copy((*[1 << 28]byte)(addPtr(dstBase, uintptr(runs[i].dstOff)))[:runs[i].bytes],
			(*[1 << 28]byte)(addPtr(srcBase, uintptr(runs[i].srcOff)))[:runs[i].bytes])
	}
}

// moveWord 可整字搬运的定宽单元，数组字覆盖 16/24 字节 string 头与三字字段
type moveWord interface {
	uint8 | uint16 | uint32 | uint64 | [2]uint64 | [3]uint64
}

// unrolledMovePairs 移动组的完全展开拷贝体
// 偏移对以局部变量值捕获烘焙进闭包，运行期免循环迭代与切片边界检查
// n 为偏移对数，超出展开上限返回 nil 回退循环实现
func unrolledMovePairs[W moveWord](runs []moveRun, pairs []uintptr, n int) groupCopyFunc {
	var a [16]uintptr
	copy(a[:], pairs)
	p := a
	switch n {
	case 1:
		s0, d0 := p[0], p[1]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
		}
	case 2:
		s0, d0, s1, d1 := p[0], p[1], p[2], p[3]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
		}
	case 3:
		s0, d0, s1, d1, s2, d2 := p[0], p[1], p[2], p[3], p[4], p[5]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
		}
	case 4:
		s0, d0, s1, d1, s2, d2, s3, d3 := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
			*(*W)(addPtr(dstBase, d3)) = *(*W)(addPtr(srcBase, s3))
		}
	case 5:
		s0, d0, s1, d1, s2, d2, s3, d3, s4, d4 := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
			*(*W)(addPtr(dstBase, d3)) = *(*W)(addPtr(srcBase, s3))
			*(*W)(addPtr(dstBase, d4)) = *(*W)(addPtr(srcBase, s4))
		}
	case 6:
		s0, d0, s1, d1, s2, d2, s3, d3, s4, d4, s5, d5 := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9], p[10], p[11]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
			*(*W)(addPtr(dstBase, d3)) = *(*W)(addPtr(srcBase, s3))
			*(*W)(addPtr(dstBase, d4)) = *(*W)(addPtr(srcBase, s4))
			*(*W)(addPtr(dstBase, d5)) = *(*W)(addPtr(srcBase, s5))
		}
	case 7:
		s0, d0, s1, d1, s2, d2, s3, d3, s4, d4, s5, d5, s6, d6 := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7], p[8], p[9], p[10], p[11], p[12], p[13]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
			*(*W)(addPtr(dstBase, d3)) = *(*W)(addPtr(srcBase, s3))
			*(*W)(addPtr(dstBase, d4)) = *(*W)(addPtr(srcBase, s4))
			*(*W)(addPtr(dstBase, d5)) = *(*W)(addPtr(srcBase, s5))
			*(*W)(addPtr(dstBase, d6)) = *(*W)(addPtr(srcBase, s6))
		}
	case 8:
		s0, d0, s1, d1, s2, d2, s3, d3 := p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7]
		s4, d4, s5, d5, s6, d6, s7, d7 := p[8], p[9], p[10], p[11], p[12], p[13], p[14], p[15]
		return func(dstBase, srcBase unsafe.Pointer) {
			moveRuns(dstBase, srcBase, runs)
			*(*W)(addPtr(dstBase, d0)) = *(*W)(addPtr(srcBase, s0))
			*(*W)(addPtr(dstBase, d1)) = *(*W)(addPtr(srcBase, s1))
			*(*W)(addPtr(dstBase, d2)) = *(*W)(addPtr(srcBase, s2))
			*(*W)(addPtr(dstBase, d3)) = *(*W)(addPtr(srcBase, s3))
			*(*W)(addPtr(dstBase, d4)) = *(*W)(addPtr(srcBase, s4))
			*(*W)(addPtr(dstBase, d5)) = *(*W)(addPtr(srcBase, s5))
			*(*W)(addPtr(dstBase, d6)) = *(*W)(addPtr(srcBase, s6))
			*(*W)(addPtr(dstBase, d7)) = *(*W)(addPtr(srcBase, s7))
		}
	}
	return nil
}

// makeMoveWordCopy 按定宽单元构建移动组拷贝体
// 偏移对数在展开上限内时完全展开烘焙进闭包，否则以循环兜底
func makeMoveWordCopy[W moveWord](runs []moveRun, pairs []uintptr) groupCopyFunc {
	if n := len(pairs) / 2; n >= 1 && n <= 8 {
		if fn := unrolledMovePairs[W](runs, pairs, n); fn != nil {
			return fn
		}
	}
	return func(dstBase, srcBase unsafe.Pointer) {
		moveRuns(dstBase, srcBase, runs)
		for i := 0; i < len(pairs); i += 2 {
			*(*W)(addPtr(dstBase, pairs[i+1])) = *(*W)(addPtr(srcBase, pairs[i]))
		}
	}
}

// makeMoveGroup 按字段字节尺寸构建移动组拷贝体，闭包捕获连续段与零散偏移对
// 常规尺寸以定宽单元整字搬运，连续段由 moveRuns 单次 memmove 覆盖，非常规尺寸走 memmove 兜底
func makeMoveGroup(sz uintptr, runs []moveRun, pairs []uintptr) groupCopyFunc {
	switch sz {
	case 1:
		return makeMoveWordCopy[uint8](runs, pairs)
	case 2:
		return makeMoveWordCopy[uint16](runs, pairs)
	case 4:
		return makeMoveWordCopy[uint32](runs, pairs)
	case 8:
		return makeMoveWordCopy[uint64](runs, pairs)
	case 16:
		return makeMoveWordCopy[[2]uint64](runs, pairs)
	case 24:
		return makeMoveWordCopy[[3]uint64](runs, pairs)
	}
	return func(dstBase, srcBase unsafe.Pointer) {
		moveRuns(dstBase, srcBase, runs)
		for i := 0; i < len(pairs); i += 2 {
			copy((*[1 << 28]byte)(addPtr(dstBase, pairs[i+1]))[:sz],
				(*[1 << 28]byte)(addPtr(srcBase, pairs[i]))[:sz])
		}
	}
}

// isMoveKind 判定字段种类是否为纯移动语义（拷贝前后位模式不变）
// fieldConvertible 仅接纳同 kind 组合，int32↔float32 等跨类数值可转换但位模式不等
// 同尺寸移动类按字节尺寸聚合为移动组，不再按类型实例区分分组
func isMoveKind(kind fieldKind, srcType, dstType reflect.Type) bool {
	switch kind {
	case fieldSameType, fieldAssignable, fieldStringFallback:
		return true
	case fieldConvertible:
		return srcType.Kind() == dstType.Kind()
	}
	return false
}

// memmoveDirect 按字节尺寸整块拷贝的直拷闭包
func memmoveDirect(sz, srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		copy((*[1 << 28]byte)(addPtr(dstBase, dstOffset))[:sz],
			(*[1 << 28]byte)(addPtr(srcBase, srcOffset))[:sz])
	}
}

// intGroupFromSrc 按源种类分派转换组展开器（第一级），标准整数返回特化展开器
func intGroupFromSrc(srcKind, dstKind reflect.Kind) func(pairs []uintptr) groupCopyFunc {
	switch srcKind {
	case reflect.Int:
		return intGroupToDst[int](dstKind)
	case reflect.Int8:
		return intGroupToDst[int8](dstKind)
	case reflect.Int16:
		return intGroupToDst[int16](dstKind)
	case reflect.Int32:
		return intGroupToDst[int32](dstKind)
	case reflect.Int64:
		return intGroupToDst[int64](dstKind)
	case reflect.Uint:
		return intGroupToDst[uint](dstKind)
	case reflect.Uint8:
		return intGroupToDst[uint8](dstKind)
	case reflect.Uint16:
		return intGroupToDst[uint16](dstKind)
	case reflect.Uint32:
		return intGroupToDst[uint32](dstKind)
	case reflect.Uint64:
		return intGroupToDst[uint64](dstKind)
	}
	return nil
}

// intGroupToDst 按目标种类分派（第二级），返回对应类型对的特化展开器
func intGroupToDst[S intKindConv](dstKind reflect.Kind) func(pairs []uintptr) groupCopyFunc {
	switch dstKind {
	case reflect.Int:
		return unrolledConvertGroup[S, int]
	case reflect.Int8:
		return unrolledConvertGroup[S, int8]
	case reflect.Int16:
		return unrolledConvertGroup[S, int16]
	case reflect.Int32:
		return unrolledConvertGroup[S, int32]
	case reflect.Int64:
		return unrolledConvertGroup[S, int64]
	case reflect.Uint:
		return unrolledConvertGroup[S, uint]
	case reflect.Uint8:
		return unrolledConvertGroup[S, uint8]
	case reflect.Uint16:
		return unrolledConvertGroup[S, uint16]
	case reflect.Uint32:
		return unrolledConvertGroup[S, uint32]
	case reflect.Uint64:
		return unrolledConvertGroup[S, uint64]
	}
	return nil
}

// pureIntegerCopyGeneric 宽度泛化整数转换单字段拷贝体
// 不同 kind 组合共享同一代码指针，以 kind 对作分组辅助键避免错误合并
func pureIntegerCopyGeneric(srcKind, dstKind reflect.Kind) (pureFieldCopy, uintptr) {
	srcSigned := IsSignedInt(srcKind)
	dstSigned := IsSignedInt(dstKind)
	srcSize := kindToSize(srcKind)
	dstSize := kindToSize(dstKind)

	return func(dstBase, srcBase unsafe.Pointer, dstOff, srcOff uintptr) {
		srcPtr := addPtr(srcBase, srcOff)
		dstPtr := addPtr(dstBase, dstOff)

		var ival int64
		var uval uint64

		if srcSigned {
			switch srcSize {
			case 1:
				ival = int64(*(*int8)(srcPtr))
			case 2:
				ival = int64(*(*int16)(srcPtr))
			case 4:
				ival = int64(*(*int32)(srcPtr))
			case 8:
				ival = *(*int64)(srcPtr)
			}
			if dstSigned {
				switch dstSize {
				case 1:
					*(*int8)(dstPtr) = int8(ival)
				case 2:
					*(*int16)(dstPtr) = int16(ival)
				case 4:
					*(*int32)(dstPtr) = int32(ival)
				case 8:
					*(*int64)(dstPtr) = ival
				}
			} else {
				uval = uint64(ival)
				switch dstSize {
				case 1:
					*(*uint8)(dstPtr) = uint8(uval)
				case 2:
					*(*uint16)(dstPtr) = uint16(uval)
				case 4:
					*(*uint32)(dstPtr) = uint32(uval)
				case 8:
					*(*uint64)(dstPtr) = uval
				}
			}
		} else {
			switch srcSize {
			case 1:
				uval = uint64(*(*uint8)(srcPtr))
			case 2:
				uval = uint64(*(*uint16)(srcPtr))
			case 4:
				uval = uint64(*(*uint32)(srcPtr))
			case 8:
				uval = *(*uint64)(srcPtr)
			}
			if dstSigned {
				ival = int64(uval)
				switch dstSize {
				case 1:
					*(*int8)(dstPtr) = int8(ival)
				case 2:
					*(*int16)(dstPtr) = int16(ival)
				case 4:
					*(*int32)(dstPtr) = int32(ival)
				case 8:
					*(*int64)(dstPtr) = ival
				}
			} else {
				switch dstSize {
				case 1:
					*(*uint8)(dstPtr) = uint8(uval)
				case 2:
					*(*uint16)(dstPtr) = uint16(uval)
				case 4:
					*(*uint32)(dstPtr) = uint32(uval)
				case 8:
					*(*uint64)(dstPtr) = uval
				}
			}
		}
	}, uintptr(srcKind)<<8 | uintptr(dstKind)
}

// pureTSToTimeField *timestamppb.Timestamp → time.Time 单字段拷贝体
func pureTSToTimeField(dstBase, srcBase unsafe.Pointer, dstOff, srcOff uintptr) {
	if t, ok := timeFromTS(*(**timestamppb.Timestamp)(addPtr(srcBase, srcOff))); ok {
		*(*time.Time)(addPtr(dstBase, dstOff)) = t
	}
}

// groupBuilderOf 按字段种类返回转换组特化展开器与分组辅助键
// 移动类字段由 buildGroupedCopyFunc 按字节尺寸统一聚合，此处仅处理转换语义
// 产生堆分配的时间与 wrapper 种类由复合分配组接管，不再进入转换组
// 返回 nil 表示闭包捕获字段外状态（wrapper 解包/嵌套/集合/transform 等），保留独立通路
func groupBuilderOf(kind fieldKind, srcType, dstType reflect.Type) (func(pairs []uintptr) groupCopyFunc, uintptr) {
	switch kind {
	case fieldInteger:
		return intGroupBuilder(srcType.Kind(), dstType.Kind())
	case fieldTSToTime:
		return unrolledTSToTimeGroup, 0
	default:
		return nil, 0
	}
}

// intGroupBuilder 整数转换组展开器入口，标准组合两级分派到泛型特化展开器
// 罕见组合回退宽度泛化循环体，kind 对作辅助键避免不同组合错误合并
func intGroupBuilder(srcKind, dstKind reflect.Kind) (func(pairs []uintptr) groupCopyFunc, uintptr) {
	if b := intGroupFromSrc(srcKind, dstKind); b != nil {
		return b, 0
	}
	pure, aux := pureIntegerCopyGeneric(srcKind, dstKind)
	return func(pairs []uintptr) groupCopyFunc {
		return func(dstBase, srcBase unsafe.Pointer) {
			for i := 0; i < len(pairs); i += 2 {
				pure(dstBase, srcBase, pairs[i+1], pairs[i])
			}
		}
	}, aux
}

// unrolledConvertGroup 整数转换组特化展开器
// 偏移烘焙进闭包，convertValue 静态调用被内联为单条 mov，与手写代码等价
// 偏移对数超过展开上限时回退循环，循环体内静态调用同样可被内联
func unrolledConvertGroup[S, D intKindConv](pairs []uintptr) groupCopyFunc {
	switch len(pairs) / 2 {
	case 1:
		s0, d0 := pairs[0], pairs[1]
		return func(dstBase, srcBase unsafe.Pointer) {
			convertValue[S, D](dstBase, srcBase, d0, s0)
		}
	case 2:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		return func(dstBase, srcBase unsafe.Pointer) {
			convertValue[S, D](dstBase, srcBase, d0, s0)
			convertValue[S, D](dstBase, srcBase, d1, s1)
		}
	case 3:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		s2, d2 := pairs[4], pairs[5]
		return func(dstBase, srcBase unsafe.Pointer) {
			convertValue[S, D](dstBase, srcBase, d0, s0)
			convertValue[S, D](dstBase, srcBase, d1, s1)
			convertValue[S, D](dstBase, srcBase, d2, s2)
		}
	case 4:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		s2, d2 := pairs[4], pairs[5]
		s3, d3 := pairs[6], pairs[7]
		return func(dstBase, srcBase unsafe.Pointer) {
			convertValue[S, D](dstBase, srcBase, d0, s0)
			convertValue[S, D](dstBase, srcBase, d1, s1)
			convertValue[S, D](dstBase, srcBase, d2, s2)
			convertValue[S, D](dstBase, srcBase, d3, s3)
		}
	default:
		return func(dstBase, srcBase unsafe.Pointer) {
			for i := 0; i < len(pairs); i += 2 {
				convertValue[S, D](dstBase, srcBase, pairs[i+1], pairs[i])
			}
		}
	}
}

// unrolledTSToTimeGroup *timestamppb.Timestamp → time.Time 组展开器
// 偏移烘焙进闭包，体内静态直调单字段拷贝体，免间接调用
func unrolledTSToTimeGroup(pairs []uintptr) groupCopyFunc {
	switch len(pairs) / 2 {
	case 1:
		s0, d0 := pairs[0], pairs[1]
		return func(dstBase, srcBase unsafe.Pointer) {
			pureTSToTimeField(dstBase, srcBase, d0, s0)
		}
	case 2:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		return func(dstBase, srcBase unsafe.Pointer) {
			pureTSToTimeField(dstBase, srcBase, d0, s0)
			pureTSToTimeField(dstBase, srcBase, d1, s1)
		}
	case 3:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		s2, d2 := pairs[4], pairs[5]
		return func(dstBase, srcBase unsafe.Pointer) {
			pureTSToTimeField(dstBase, srcBase, d0, s0)
			pureTSToTimeField(dstBase, srcBase, d1, s1)
			pureTSToTimeField(dstBase, srcBase, d2, s2)
		}
	case 4:
		s0, d0 := pairs[0], pairs[1]
		s1, d1 := pairs[2], pairs[3]
		s2, d2 := pairs[4], pairs[5]
		s3, d3 := pairs[6], pairs[7]
		return func(dstBase, srcBase unsafe.Pointer) {
			pureTSToTimeField(dstBase, srcBase, d0, s0)
			pureTSToTimeField(dstBase, srcBase, d1, s1)
			pureTSToTimeField(dstBase, srcBase, d2, s2)
			pureTSToTimeField(dstBase, srcBase, d3, s3)
		}
	default:
		return func(dstBase, srcBase unsafe.Pointer) {
			for i := 0; i < len(pairs); i += 2 {
				pureTSToTimeField(dstBase, srcBase, pairs[i+1], pairs[i])
			}
		}
	}
}

// 复合分配组：单次转换内全部子对象分配聚合为一次复合对象分配
//
// wrapper 包装/解包、时间戳互转、值→指针、嵌套结构体指针等字段各自分配独立子对象，
// N 个字段产生 N 次分配与 N 次间接调用，是热路径 mallocgc 与 GC 压力的主要来源
// 构建期以 reflect.StructOf 合成一个容纳全部子对象的复合类型，运行期仅分配复合对象，
// 子对象取复合体字段地址写入目标，&composite.field 语义与普通堆结构体取址一致，
// GC 经 findObject 解析内部指针定位复合体基址，存活期由目标引用保持
// ══════════════════════════════════════════════════════════════════════════════

// allocSpecKind 复合分配字段种类，值拷贝语义与对应独立闭包逐一对齐
type allocSpecKind uint8

const (
	allocWrapPtr     allocSpecKind = iota // *T → *wrapperspb.XxxValue 包装
	allocUnwrapPtr                        // *wrapperspb.XxxValue → *T 解包
	allocWrapVal                          // T → *wrapperspb.XxxValue 值包装
	allocTimeToTS                         // time.Time → *timestamppb.Timestamp
	allocTimePtrToTS                      // *time.Time → *timestamppb.Timestamp
	allocTSToTimePtr                      // *timestamppb.Timestamp → *time.Time
	allocToPtrScalar                      // T → *T 标量，零值写 nil
	allocToPtrOpaque                      // T → *T 非标量（slice/map 等），无零值检查
	allocStructPtr                        // *StructA → *StructB，子缓存直转
	allocStructToPtr                      // StructA → *StructB，子缓存直转
)

// allocSpec 复合分配组的单字段规格，构建期一次解析，运行期只读
type allocSpec struct {
	kind     allocSpecKind
	srcKind  reflect.Kind // 源标量种类，float 零值需类型化比较
	srcOff   uint32
	dstOff   uint32
	compOff  uint32 // 复合体内子对象字段偏移
	valueOff uint32 // wrapper Value 字段偏移
	size     uint32 // 值字节数
	sc       *structFieldCache
	hasSlow  bool
}

// isZeroScalar 源标量是否为零值，float 以类型化比较覆盖 -0.0 等价 0.0 的语义
// string 以 len==0 判空，空串可能携带非空数据指针，逐字节比较会误判
func isZeroScalar(p unsafe.Pointer, k reflect.Kind) bool {
	switch k {
	case reflect.Float32:
		return *(*float32)(p) == 0
	case reflect.Float64:
		return *(*float64)(p) == 0
	case reflect.String:
		return len(*(*string)(p)) == 0
	default:
		b := (*[8]byte)(p)
		for i := uintptr(0); i < kindToSize(k); i++ {
			if b[i] != 0 {
				return false
			}
		}
		return true
	}
}

// allocStructSpec 结构体直转字段的复合规格，子缓存缺失 mergedFn 时回退独立闭包
// kind 由调用方按字段种类预设，此处仅挂载子缓存与慢路径标记
func allocStructSpec(spec allocSpec, srcElem, dstElem reflect.Type) (allocSpec, reflect.Type, bool) {
	sc := loadStructFieldCache(srcElem, dstElem)
	if sc.mergedFn == nil {
		return spec, nil, false
	}
	spec.sc = sc
	spec.hasSlow = len(sc.slowEntries) > 0
	return spec, dstElem, true
}

// allocSpecOf 解析可复合分配字段视图为规格，不可处理的种类返回 ok=false 回退独立闭包
func allocSpecOf(s *groupSrc) (allocSpec, reflect.Type, bool) {
	spec := allocSpec{srcOff: uint32(s.srcOff), dstOff: uint32(s.dstOff), srcKind: s.srcType.Kind()}
	switch s.kind {
	case fieldWrapper:
		wc, dir := lookupWrapper(s.srcType, s.dstType)
		if wc == nil {
			return spec, nil, false
		}
		f, ok := wc.wrapperType.Elem().FieldByName("Value")
		if !ok {
			return spec, nil, false
		}
		spec.valueOff = uint32(f.Offset)
		spec.size = uint32(wc.valueType.Size())
		switch dir {
		case 0: // *T → *W
			spec.kind = allocWrapPtr
			return spec, wc.wrapperType.Elem(), true
		case 1: // *W → *T
			spec.kind = allocUnwrapPtr
			return spec, wc.valueType, true
		case 2: // T → *W
			spec.kind = allocWrapVal
			return spec, wc.wrapperType.Elem(), true
		default: // *W → T 无分配
			return spec, nil, false
		}
	case fieldTimeToTS:
		spec.kind = allocTimeToTS
		return spec, timestampType, true
	case fieldTimePtrToTS:
		spec.kind = allocTimePtrToTS
		return spec, timestampType, true
	case fieldTSToTimePtr:
		spec.kind = allocTSToTimePtr
		return spec, timeType, true
	case fieldToPtr:
		elem := s.dstType.Elem()
		if s.srcType == elem {
			switch elem.Kind() {
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
				spec.kind = allocToPtrScalar
				spec.size = uint32(elem.Size())
				return spec, elem, true
			case reflect.Struct:
				spec.kind = allocStructToPtr
				return allocStructSpec(spec, s.srcType, elem)
			default:
				spec.kind = allocToPtrOpaque
				spec.size = uint32(elem.Size())
				return spec, elem, true
			}
		}
		if s.srcType.Kind() == reflect.Struct && elem.Kind() == reflect.Struct {
			spec.kind = allocStructToPtr
			return allocStructSpec(spec, s.srcType, elem)
		}
		return spec, nil, false
	case fieldStructPtr:
		spec.kind = allocStructPtr
		return allocStructSpec(spec, s.srcType.Elem(), s.dstType.Elem())
	}
	return spec, nil, false
}

// buildAllocGroup 构建复合分配组执行闭包
// 复合类型字段与规格一一对应，compOff 取自复合类型字段偏移
// 闭包单次分配复合体后按规格逐字段完成解引用、零值判定、值填充与目标指针写入
func buildAllocGroup(specs []allocSpec, compType reflect.Type) groupCopyFunc {
	for i := range specs {
		specs[i].compOff = uint32(compType.Field(i).Offset)
	}
	return func(dstBase, srcBase unsafe.Pointer) {
		comp := reflect.New(compType).UnsafePointer()
		for i := range specs {
			s := &specs[i]
			switch s.kind {
			case allocWrapPtr:
				srcPtr := *(*unsafe.Pointer)(addPtr(srcBase, uintptr(s.srcOff)))
				if srcPtr == nil {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				w := addPtr(comp, uintptr(s.compOff))
				copy((*[1 << 28]byte)(addPtr(w, uintptr(s.valueOff)))[:s.size],
					(*[1 << 28]byte)(srcPtr)[:s.size])
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = w
			case allocUnwrapPtr:
				srcPtr := *(*unsafe.Pointer)(addPtr(srcBase, uintptr(s.srcOff)))
				if srcPtr == nil {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				v := addPtr(comp, uintptr(s.compOff))
				copy((*[1 << 28]byte)(v)[:s.size],
					(*[1 << 28]byte)(addPtr(srcPtr, uintptr(s.valueOff)))[:s.size])
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = v
			case allocWrapVal:
				w := addPtr(comp, uintptr(s.compOff))
				copy((*[1 << 28]byte)(addPtr(w, uintptr(s.valueOff)))[:s.size],
					(*[1 << 28]byte)(addPtr(srcBase, uintptr(s.srcOff)))[:s.size])
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = w
			case allocTimeToTS:
				t := *(*time.Time)(addPtr(srcBase, uintptr(s.srcOff)))
				if t.IsZero() {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				ts := (*timestamppb.Timestamp)(addPtr(comp, uintptr(s.compOff)))
				ts.Seconds = int64(t.Unix())
				ts.Nanos = int32(t.Nanosecond())
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = unsafe.Pointer(ts)
			case allocTimePtrToTS:
				tp := *(**time.Time)(addPtr(srcBase, uintptr(s.srcOff)))
				if tp == nil || tp.IsZero() {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				ts := (*timestamppb.Timestamp)(addPtr(comp, uintptr(s.compOff)))
				ts.Seconds = int64(tp.Unix())
				ts.Nanos = int32(tp.Nanosecond())
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = unsafe.Pointer(ts)
			case allocTSToTimePtr:
				tsPtr := *(**timestamppb.Timestamp)(addPtr(srcBase, uintptr(s.srcOff)))
				if tsPtr == nil {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				t := (*time.Time)(addPtr(comp, uintptr(s.compOff)))
				*t = tsPtr.AsTime()
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = unsafe.Pointer(t)
			case allocToPtrScalar:
				srcAddr := addPtr(srcBase, uintptr(s.srcOff))
				if isZeroScalar(srcAddr, s.srcKind) {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				v := addPtr(comp, uintptr(s.compOff))
				copy((*[1 << 28]byte)(v)[:s.size], (*[1 << 28]byte)(srcAddr)[:s.size])
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = v
			case allocToPtrOpaque:
				v := addPtr(comp, uintptr(s.compOff))
				copy((*[1 << 28]byte)(v)[:s.size],
					(*[1 << 28]byte)(addPtr(srcBase, uintptr(s.srcOff)))[:s.size])
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = v
			case allocStructPtr:
				srcPtr := *(*unsafe.Pointer)(addPtr(srcBase, uintptr(s.srcOff)))
				if srcPtr == nil {
					*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = nil
					continue
				}
				dstPtr := addPtr(comp, uintptr(s.compOff))
				s.sc.mergedFn(srcPtr, dstPtr)
				if s.hasSlow {
					applyStructSlowEntriesUnsafe(srcPtr, dstPtr, s.sc)
				}
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = dstPtr
			case allocStructToPtr:
				srcAddr := addPtr(srcBase, uintptr(s.srcOff))
				dstPtr := addPtr(comp, uintptr(s.compOff))
				s.sc.mergedFn(srcAddr, dstPtr)
				if s.hasSlow {
					applyStructSlowEntriesUnsafe(srcAddr, dstPtr, s.sc)
				}
				*(*unsafe.Pointer)(addPtr(dstBase, uintptr(s.dstOff))) = dstPtr
			}
		}
	}
}

// buildGroupedCopyFunc 将字段视图按拷贝体特征分组构建合并执行闭包
// 移动类字段按字节尺寸聚合为移动组，finalize 时抽取双侧同步连续段以单次 memmove 覆盖
// 转换类字段以 (拷贝体代码指针, 辅助键) 聚合，同组字段共享一次间接调用
// 产生堆分配的字段聚合为复合分配组，单次复合分配覆盖全部子对象
// 不可分组种类收集为 singles，复用 buildMergedCopyFuncFromFns 的展开执行
// 各字段仅写自身目标位置互不重叠，组间执行顺序不影响结果
func buildGroupedCopyFunc(srcs []groupSrc) func(srcBase, dstBase unsafe.Pointer) {
	type groupAcc struct {
		isMove  bool
		sz      uintptr
		builder func(pairs []uintptr) groupCopyFunc
		pairs   []uintptr
	}
	var groups []groupAcc
	var singles []fieldCopyFunc
	var allocSpecs []allocSpec
	var allocFieldTypes []reflect.Type

	type groupKey struct {
		code uintptr
		aux  uintptr
	}
	index := make(map[groupKey]int, len(srcs))
	for i := range srcs {
		s := &srcs[i]
		var key groupKey
		var builder func(pairs []uintptr) groupCopyFunc
		isMove := false
		if s.groupable {
			if isMoveKind(s.kind, s.srcType, s.dstType) && s.srcType.Size() == s.dstType.Size() {
				isMove = true
				key = groupKey{aux: s.srcType.Size()}
			} else if spec, ft, ok := allocSpecOf(s); ok {
				allocSpecs = append(allocSpecs, spec)
				allocFieldTypes = append(allocFieldTypes, ft)
				continue
			} else {
				var aux uintptr
				builder, aux = groupBuilderOf(s.kind, s.srcType, s.dstType)
				if builder == nil {
					if s.legacy != nil {
						singles = append(singles, s.legacy)
					}
					continue
				}
				key = groupKey{code: reflect.ValueOf(builder).Pointer(), aux: aux}
			}
		} else {
			if s.legacy != nil {
				singles = append(singles, s.legacy)
			}
			continue
		}
		id, ok := index[key]
		if !ok {
			id = len(groups)
			index[key] = id
			groups = append(groups, groupAcc{isMove: isMove, sz: key.aux, builder: builder})
		}
		g := &groups[id]
		g.pairs = append(g.pairs, s.srcOff, s.dstOff)
	}

	// finalize 移动组：抽取双侧同步连续段烘焙进拷贝体，零散字段保留逐对搬运
	// finalize 转换组：偏移对传入特化展开器烘焙进闭包，体内静态直调免间接调用
	// finalize 复合分配组：StructOf 合成复合类型，单次分配覆盖全部子对象
	var compiled []copyGroup
	for i := range groups {
		g := &groups[i]
		if g.isMove {
			runs, rest := extractMoveRuns(g.pairs, g.sz)
			compiled = append(compiled, copyGroup{copy: makeMoveGroup(g.sz, runs, rest)})
		} else {
			compiled = append(compiled, copyGroup{copy: g.builder(g.pairs)})
		}
	}
	if len(allocSpecs) > 0 {
		fields := make([]reflect.StructField, len(allocSpecs))
		for i := range allocFieldTypes {
			fields[i] = reflect.StructField{Name: "F" + strconv.Itoa(i), Type: allocFieldTypes[i]}
		}
		compiled = append(compiled, copyGroup{copy: buildAllocGroup(allocSpecs, reflect.StructOf(fields))})
	}

	singlesFn := buildMergedCopyFuncFromFns(singles)
	if n := len(compiled); n <= 8 {
		return unrolledGroupsExec(compiled, singlesFn, n)
	}
	if singlesFn == nil {
		return func(srcBase, dstBase unsafe.Pointer) {
			for i := range compiled {
				compiled[i].copy(dstBase, srcBase)
			}
		}
	}
	return func(srcBase, dstBase unsafe.Pointer) {
		for i := range compiled {
			compiled[i].copy(dstBase, srcBase)
		}
		singlesFn(srcBase, dstBase)
	}
}

// unrolledGroupsExec 将组数不超过展开上限的执行闭包直调各组，免循环迭代与切片索引开销
// 构建期组数与 singles 存在性均固定，闭包内分支可被完美预测
func unrolledGroupsExec(compiled []copyGroup, singlesFn func(srcBase, dstBase unsafe.Pointer), n int) func(srcBase, dstBase unsafe.Pointer) {
	switch n {
	case 0:
		return singlesFn
	case 1:
		c0 := compiled[0].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 2:
		c0, c1 := compiled[0].copy, compiled[1].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 3:
		c0, c1, c2 := compiled[0].copy, compiled[1].copy, compiled[2].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 4:
		c0, c1, c2, c3 := compiled[0].copy, compiled[1].copy, compiled[2].copy, compiled[3].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			c3(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 5:
		c0, c1 := compiled[0].copy, compiled[1].copy
		c2, c3 := compiled[2].copy, compiled[3].copy
		c4 := compiled[4].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			c3(dstBase, srcBase)
			c4(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 6:
		c0, c1 := compiled[0].copy, compiled[1].copy
		c2, c3 := compiled[2].copy, compiled[3].copy
		c4, c5 := compiled[4].copy, compiled[5].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			c3(dstBase, srcBase)
			c4(dstBase, srcBase)
			c5(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	case 7:
		c0, c1 := compiled[0].copy, compiled[1].copy
		c2, c3 := compiled[2].copy, compiled[3].copy
		c4, c5 := compiled[4].copy, compiled[5].copy
		c6 := compiled[6].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			c3(dstBase, srcBase)
			c4(dstBase, srcBase)
			c5(dstBase, srcBase)
			c6(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	default:
		c0, c1 := compiled[0].copy, compiled[1].copy
		c2, c3 := compiled[2].copy, compiled[3].copy
		c4, c5 := compiled[4].copy, compiled[5].copy
		c6, c7 := compiled[6].copy, compiled[7].copy
		return func(srcBase, dstBase unsafe.Pointer) {
			c0(dstBase, srcBase)
			c1(dstBase, srcBase)
			c2(dstBase, srcBase)
			c3(dstBase, srcBase)
			c4(dstBase, srcBase)
			c5(dstBase, srcBase)
			c6(dstBase, srcBase)
			c7(dstBase, srcBase)
			if singlesFn != nil {
				singlesFn(srcBase, dstBase)
			}
		}
	}
}
