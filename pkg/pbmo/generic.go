/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-23 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-27 11:35:07
 * @FilePath: \go-toolbox\pkg\pbmo\generic.go
 * @Description: 泛型便捷函数 - 一行注册、一行转换
 * 利用 reflect.Type 做 key + sync.Map 缓存，消除 Model 层样板代码
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// typePair 类型对，用于缓存 key
type typePair [2]reflect.Type

// converterCache 转换器缓存：typePair -> *BidiConverter
var converterCache sync.Map

// entrySlotCount 热路径缓存槽位数，按类型对哈希分片降低多类型对互驱逐
//
// 槽组是**直接映射（1-way）**的：每槽仅存一个条目，故同槽的两个类型对会互相驱逐、
// 永不命中
// 槽位数必须显著大于进程活跃类型对数，否则退化为每次调用都回源重建
// 实测（57 个活跃类型对，消费方真实规模）：32 槽命中率仅约 16%，单次 ToPB
// 从 27ns 劣化到 91ns（3.4×），且 ensureEntry 的条目重建占全部分配字节的 46%
//
// 512 槽把同等规模的命中率抬到约 93%，静态开销反而略降：
// 原先 32 槽 ×136B（128B 填充 + 8B 指针）= 4352B，现 512 槽 ×8B = 4096B
//
// 移除缓存行填充的权衡：填充只在**写**频繁时才有价值，而写只发生于未命中——
// 这恰是本次要消除的工况
// 命中率 16%→93% 意味着写占比 84%→7%，
// 伪共享的暴露面随之下降一个数量级，远超 8 槽共享一行带来的损失
//
// 注意：单类型对基准永远命中槽位，看不到本问题——回归需多类型对基准覆盖
const entrySlotCount = 512

// entrySlot 热路径缓存槽位
type entrySlot struct {
	entry atomic.Pointer[cachedConverter]
}

// lastEntries 最近使用的转换条目槽组
var lastEntries [entrySlotCount]entrySlot

// cachedConverter 转换器热路径缓存条目
// 一次原子读同时取得转换器与双向字段缓存，稳态零分配
// 槽位由 invalidateCache 全量清空感知失效，热路径免代数检查
// 两侧纯拷贝标志与合并闭包预烘焙，热路径免查 slowEntries 直接分发
type cachedConverter struct {
	key       fastKey
	converter *BidiConverter
	pbToModel *fieldCache
	modelToPB *fieldCache
	pbMerged  func(srcBase, dstBase unsafe.Pointer)
	mMerged   func(srcBase, dstBase unsafe.Pointer)
	pbPure    bool
	mPure     bool
}

// fastKey 热路径槽位 key，取类型对指针描述符地址组合
// 指针类型描述符全局唯一，等值比较退化为两次纯指针比较
// 避免 typePair 接口等值比较引入的 ifaceeq 间接调用开销
type fastKey struct {
	model uintptr
	pb    uintptr
}

// derefType 指针类型解引用，与注册中心 key 语义保持一致
func derefType(t reflect.Type) reflect.Type {
	// 循环剥到底，不是只剥一层
	//
	// 类型参数本身可以是指针：SafeToPBs[*Model] 这种用法下，入口里的 new(M) 会得到
	// **Model，只剥一层剩 *Model，再交给 buildFieldCache 就是
	// reflect: NumField of non-struct type *Model 的 panic
	//
	// 这个 panic 只在某些测试顺序下出现，因为 converterCache 的键是规范过的类型对：
	// 先跑的调用若已用正确类型建好转换器，后面的调用直接复用、永远走不到构造那一步
	// 实测 -shuffle=3 必现，而 1 / 2 / 5 不现
	//
	// [EN] Dereferences all the way down rather than one level. A type parameter may itself be
	// a pointer, as in SafeToPBs[*Model], where new(M) yields **Model and stripping one level
	// leaves *Model, which panics in buildFieldCache. The panic surfaced only under some test
	// orders because converterCache is keyed by the normalised type pair: an earlier call that
	// built the converter from the correct type means later calls reuse it and never
	// construct one. Reproduced reliably with -shuffle=3 and not with 1, 2 or 5
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// slotIndexOf 计算类型对槽位下标
// rtype 指针低位因内存对齐恒零，移位后取有效位参与散列
// 黄金比例乘数将乘积高低位充分混合，避免全部类型对聚在同一槽位
func slotIndexOf(mk, pk uintptr) int {
	h := ((mk >> 4) * 0x9E3779B97F4A7C15) ^ ((pk >> 4) * 0xC2B2AE3D27D4EB4F)
	return int((h ^ (h >> 29)) & (entrySlotCount - 1))
}

// typeWordOf 提取 T 指针类型描述符地址
// nil 指针装箱仅写静态类型字与零数据字，类型字为链接期确定的只读静态地址
// 热路径内联为单次常量加载，取代 reflect.TypeFor 的反射调用链
func typeWordOf[T any]() uintptr {
	var e any = (*T)(nil)
	type ifaceHeader struct {
		Type unsafe.Pointer
		_    unsafe.Pointer
	}
	return uintptr((*ifaceHeader)(unsafe.Pointer(&e)).Type)
}

// typePairOf 规范类型对，冷路径专用
func typePairOf[M any, P any]() typePair {
	return typePair{derefType(reflect.TypeFor[M]()), derefType(reflect.TypeFor[P]())}
}

// newCachedConverter 确保字段缓存已构建并生成热路径条目
// 预判两侧纯拷贝可走性并烘焙合并闭包，热路径免查 slowEntries
func newCachedConverter(key fastKey, c *BidiConverter) *cachedConverter {
	pbToModel, modelToPB := c.fieldCaches()
	return &cachedConverter{
		key:       key,
		converter: c,
		pbToModel: pbToModel,
		modelToPB: modelToPB,
		pbMerged:  pbToModel.mergedCopyFunc,
		mMerged:   modelToPB.mergedCopyFunc,
		pbPure:    pbToModel.mergedCopyFunc != nil && len(pbToModel.slowEntries) == 0,
		mPure:     modelToPB.mergedCopyFunc != nil && len(modelToPB.slowEntries) == 0,
	}
}

// ensureEntry 回源 converterCache 构建热路径条目并写入槽位，冷路径专用
// 槽位未命中或被 invalidateCache 清空后调用，并发重复构建引用同一转换器幂等无害
func ensureEntry[M any, P any](mk, pk uintptr) *cachedConverter {
	tp := typePairOf[M, P]()
	var c *BidiConverter
	if v, ok := converterCache.Load(tp); ok {
		c = v.(*BidiConverter)
	} else {
		actual, _ := converterCache.LoadOrStore(tp,
			NewBidiConverter(new(P), new(M)).WithAutoTimeConversion(true))
		c = actual.(*BidiConverter)
	}
	e := newCachedConverter(fastKey{mk, pk}, c)
	lastEntries[slotIndexOf(mk, pk)].entry.Store(e)
	return e
}

// loadEntry 确保类型对的热路径缓存条目就绪，批量入口的单次查找点
// 命中仅一次原子读，未命中或被清空回源重建
func loadEntry[M any, P any]() *cachedConverter {
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	if e := lastEntries[slotIndexOf(mk, pk)].entry.Load(); e != nil &&
		e.key.model == mk && e.key.pb == pk {
		return e
	}
	return ensureEntry[M, P](mk, pk)
}

// lookupCachedConverter 按运行时元素类型查找热路径缓存条目
// key 与 typeWordOf 同源：modelType 经 PointerTo 规范为指针描述符，槽位命中零装箱
// 未命中回源 converterCache，仍无则返回 nil 不创建
// 用于元素类型与类型参数不一致的批量入口（如 SafeToPBs 的指针切片）
func lookupCachedConverter[P any](modelType reflect.Type) *cachedConverter {
	mk := uintptr(ifaceDataPtr(reflect.PointerTo(modelType)))
	pk := typeWordOf[P]()
	key := fastKey{mk, pk}
	slot := &lastEntries[slotIndexOf(mk, pk)]
	if e := slot.entry.Load(); e != nil && e.key == key {
		return e
	}
	var c *BidiConverter
	if v, ok := converterCache.Load(typePair{modelType, derefType(reflect.TypeFor[P]())}); ok {
		c = v.(*BidiConverter)
	} else {
		return nil
	}
	e := newCachedConverter(key, c)
	slot.entry.Store(e)
	return e
}

// getOrInitConverter 获取或初始化转换器
func getOrInitConverter[M any, P any]() *BidiConverter {
	return loadEntry[M, P]().converter
}

// Register[M, P] 注册 Model-PB 转换对（使用默认配置）
func Register[M any, P any]() *BidiConverter {
	return getOrInitConverter[M, P]()
}

// RegisterWith[M, P] 注册 Model-PB 转换对（自定义配置）
func RegisterWith[M any, P any](opts ...Option) *BidiConverter {
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	c := NewBidiConverter(new(P), new(M), opts...)
	converterCache.Store(typePairOf[M, P](), c)
	lastEntries[slotIndexOf(mk, pk)].entry.Store(newCachedConverter(fastKey{mk, pk}, c))
	return c
}

// ToPB 将 Model 转换为 PB 消息，nil 输入返回 nil
// 类型字静态提取一次原子读命中即分发，纯拷贝直达合并闭包，稳态零反射
//
// 目标由本函数分配并返回，该分配必然逃逸；需要稳态零分配时改用 ToPBInto
// 切片/map/指针类字段与源共享底层数据，不做防御性复制（见包文档「字段别名语义」）
func ToPB[M any, P any](m *M) (*P, error) {
	if m == nil {
		return nil, nil
	}
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	e := lastEntries[slotIndexOf(mk, pk)].entry.Load()
	if e == nil || e.key.model != mk || e.key.pb != pk {
		e = ensureEntry[M, P](mk, pk)
	}
	info := new(P)
	if e.mPure {
		e.mMerged(unsafe.Pointer(m), unsafe.Pointer(info))
		return info, nil
	}
	return info, e.converter.convertPtrCached(unsafe.Pointer(m), unsafe.Pointer(info), e.modelToPB)
}

// FromPB 将 PB 消息转换为 Model，nil 输入返回 nil
// 类型字静态提取一次原子读命中即分发，纯拷贝直达合并闭包，稳态零反射
//
// 目标由本函数分配并返回，该分配必然逃逸；需要稳态零分配时改用 FromPBInto
// 切片/map/指针类字段与源共享底层数据，不做防御性复制（见包文档「字段别名语义」）
func FromPB[P any, M any](pb *P) (*M, error) {
	if pb == nil {
		return nil, nil
	}
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	e := lastEntries[slotIndexOf(mk, pk)].entry.Load()
	if e == nil || e.key.model != mk || e.key.pb != pk {
		e = ensureEntry[M, P](mk, pk)
	}
	model := new(M)
	if e.pbPure {
		e.pbMerged(unsafe.Pointer(pb), unsafe.Pointer(model))
		return model, nil
	}
	return model, e.converter.convertPtrCached(unsafe.Pointer(pb), unsafe.Pointer(model), e.pbToModel)
}

// ToPBInto 将 Model 转换到调用方提供的 PB 目标，免去 ToPB 的目标堆分配
//
// 与 ToPB 的唯一差异是目标归属ToPB 必须 new(P) 后返回，该返回值必然逃逸——
// 它是该路径 100% 的垃圾与约 47% 的单次耗时，也是并行扩展的真实天花板
// （无锁竞争，纯分配/GC 压力：同一基准关掉 GC 后 1→10 核扩展比从 1.31× 升到 2.80×）
// 本函数写入调用方持有的 dst，配合对象池复用即稳态零分配
//
// ⚠️ 调用方契约：dst 的清零责任在调用方
// 拷贝计划只写「存在映射关系的字段」，未映射字段保持 dst 原有值不变——
// 复用一个脏目标会残留上一轮的数据复用前请显式置零（*dst = P{}），
// 或确认该类型对的全部字段均被映射覆盖
//
// m 为 nil 时不触碰 dst 并返回 nil（与 ToPB 的 nil→nil 语义一致）；
// dst 为 nil 返回 ErrPBMessageNil
//
// 字段别名语义与 ToPB 完全一致：标量与 []byte/切片头经合并 memmove 直拷，
// 切片类字段与源共享底层数组，不做防御性复制（见包文档「字段别名语义」）
func ToPBInto[M any, P any](m *M, dst *P) error {
	if m == nil {
		return nil
	}
	if dst == nil {
		return ErrPBMessageNil
	}
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	e := lastEntries[slotIndexOf(mk, pk)].entry.Load()
	if e == nil || e.key.model != mk || e.key.pb != pk {
		e = ensureEntry[M, P](mk, pk)
	}
	if e.mPure {
		e.mMerged(unsafe.Pointer(m), unsafe.Pointer(dst))
		return nil
	}
	return e.converter.convertPtrCached(unsafe.Pointer(m), unsafe.Pointer(dst), e.modelToPB)
}

// FromPBInto 将 PB 消息转换到调用方提供的 Model 目标，免去 FromPB 的目标堆分配
//
// 与 ToPBInto 对偶，动机、契约与别名语义同上：dst 清零责任在调用方，
// 未映射字段保持原值；pb 为 nil 时不触碰 dst 并返回 nil；dst 为 nil 返回 ErrModelNil
func FromPBInto[P any, M any](pb *P, dst *M) error {
	if pb == nil {
		return nil
	}
	if dst == nil {
		return ErrModelNil
	}
	mk := typeWordOf[M]()
	pk := typeWordOf[P]()
	e := lastEntries[slotIndexOf(mk, pk)].entry.Load()
	if e == nil || e.key.model != mk || e.key.pb != pk {
		e = ensureEntry[M, P](mk, pk)
	}
	if e.pbPure {
		e.pbMerged(unsafe.Pointer(pb), unsafe.Pointer(dst))
		return nil
	}
	return e.converter.convertPtrCached(unsafe.Pointer(pb), unsafe.Pointer(dst), e.pbToModel)
}

// ToPBs 批量将 Model 指针切片转换为 PB 切片
func ToPBs[M any, P any](models []*M) ([]*P, error) {
	pbs := make([]*P, len(models))
	if len(models) == 0 {
		return pbs, nil
	}
	e := loadEntry[M, P]()
	values := make([]P, len(models))
	if e.mPure {
		for i, m := range models {
			p := &values[i]
			if m != nil {
				e.mMerged(unsafe.Pointer(m), unsafe.Pointer(p))
			}
			pbs[i] = p
		}
		return pbs, nil
	}
	for i, m := range models {
		p := &values[i]
		if err := e.converter.convertPtrCached(unsafe.Pointer(m), unsafe.Pointer(p), e.modelToPB); err != nil {
			return nil, NewBatchError("元素 %d: %v", i, err)
		}
		pbs[i] = p
	}
	return pbs, nil
}

// FromPBs 批量将 PB 切片转换为 Model 指针切片
// nil 元素会被跳过，对应的 model 保持零值
func FromPBs[P any, M any](pbs []*P) ([]*M, error) {
	models := make([]*M, len(pbs))
	if len(pbs) == 0 {
		return models, nil
	}
	e := loadEntry[M, P]()
	values := make([]M, len(pbs))
	if e.pbPure {
		for i, pb := range pbs {
			m := &values[i]
			if pb != nil {
				e.pbMerged(unsafe.Pointer(pb), unsafe.Pointer(m))
			}
			models[i] = m
		}
		return models, nil
	}
	for i, pb := range pbs {
		m := &values[i]
		if pb == nil {
			models[i] = m
			continue
		}
		if err := e.converter.convertPtrCached(unsafe.Pointer(pb), unsafe.Pointer(m), e.pbToModel); err != nil {
			return nil, NewBatchError("元素 %d: %v", i, err)
		}
		models[i] = m
	}
	return models, nil
}

// convertBatchItem 转换单个元素并写入批量结果，SafeToPBs/SafeFromPBs 共用
// pureFn 非 nil 时走零错误快速路径，否则走 convertPtrCached 慢速路径
// 转换成功返回元素指针，失败返回 nil 并记录失败项
func convertBatchItem[T any](
	srcPtr unsafe.Pointer, dstVal *T, index int,
	pureFn func(unsafe.Pointer, unsafe.Pointer),
	c *BidiConverter, cache *fieldCache,
	result *BatchResult,
) *T {
	var err error
	if pureFn != nil {
		pureFn(srcPtr, unsafe.Pointer(dstVal))
	} else {
		err = c.convertPtrCached(srcPtr, unsafe.Pointer(dstVal), cache)
	}

	item := BatchItem{Index: index}
	if err != nil {
		item.Success = false
		item.Error = err
		result.FailureCount++
		result.Results = append(result.Results, item)
		return nil
	}
	item.Success = true
	item.Value = dstVal
	result.SuccessCount++
	result.Results = append(result.Results, item)
	return dstVal
}

// SafeToPBs 安全批量将 Model 切片转换为 PB 切片，不因单个失败而中断
// 自动支持 []M（值切片）和 []*M（指针切片），nil 指针元素标记为失败
func SafeToPBs[M any, P any](models []M) ([]*P, *BatchResult) {
	pbs := make([]*P, len(models))
	result := &BatchResult{Results: make([]BatchItem, 0, len(models))}
	if len(models) == 0 {
		return pbs, result
	}
	values := make([]P, len(models))

	mType := reflect.TypeOf(models).Elem()

	if mType.Kind() == reflect.Ptr {
		// 槽位 key 与元素真实类型同源，未命中回源 loadEntry 惰性创建
		e := lookupCachedConverter[P](mType.Elem())
		if e == nil {
			e = loadEntry[M, P]()
		}
		var pureFn func(unsafe.Pointer, unsafe.Pointer)
		if e.mPure {
			pureFn = e.mMerged
		}
		// 切片数据指针按元素大小步进，避免逐元素 reflect.ValueOf 构造
		base := unsafe.SliceData(models)
		elemSize := unsafe.Sizeof(*base)
		for i := range models {
			item := BatchItem{Index: i}
			// 元素为指针类型，槽位数据即元素指针值
			srcPtr := *(*unsafe.Pointer)(addPtr(unsafe.Pointer(base), uintptr(i)*elemSize))
			if srcPtr == nil {
				item.Success = false
				item.Error = NewConversionError("元素 %d: 输入为 nil", i)
				result.FailureCount++
				result.Results = append(result.Results, item)
				continue
			}
			if p := convertBatchItem(srcPtr, &values[i], i, pureFn, e.converter, e.modelToPB, result); p != nil {
				pbs[i] = p
			}
		}
	} else {
		e := loadEntry[M, P]()
		var pureFn func(unsafe.Pointer, unsafe.Pointer)
		if e.mPure {
			pureFn = e.mMerged
		}
		for i := range models {
			if p := convertBatchItem(unsafe.Pointer(&models[i]), &values[i], i, pureFn, e.converter, e.modelToPB, result); p != nil {
				pbs[i] = p
			}
		}
	}

	return pbs, result
}

// SafeFromPBs 安全批量将 PB 切片转换为 Model 切片，不因单个失败而中断
// 接受 []*P（指针切片），nil 元素标记为失败
func SafeFromPBs[P any, M any](pbs []*P) ([]*M, *BatchResult) {
	models := make([]*M, len(pbs))
	result := &BatchResult{Results: make([]BatchItem, 0, len(pbs))}
	if len(pbs) == 0 {
		return models, result
	}
	values := make([]M, len(pbs))
	e := loadEntry[M, P]()
	var pureFn func(unsafe.Pointer, unsafe.Pointer)
	if e.pbPure {
		pureFn = e.pbMerged
	}

	for i, pb := range pbs {
		if pb == nil {
			item := BatchItem{Index: i}
			item.Success = false
			item.Error = NewConversionError("元素 %d: 输入为 nil", i)
			result.FailureCount++
			result.Results = append(result.Results, item)
			continue
		}
		if m := convertBatchItem(unsafe.Pointer(pb), &values[i], i, pureFn, e.converter, e.pbToModel, result); m != nil {
			models[i] = m
		}
	}

	return models, result
}

// ConverterFor 获取已注册的转换器
func ConverterFor[M any, P any]() *BidiConverter {
	return getOrInitConverter[M, P]()
}

// PBToUpdates 将 PB 消息转换为 map[string]interface{}
// 遍历 proto 字段，使用 protobuf json tag 名作为 key（snake_case）
// 自动跳过零值、protobuf 内部字段，自动解包 wrapper 类型
// 字段 key 与偏移按类型缓存于更新计划，热路径零 tag 解析
func PBToUpdates[P any](pb *P) map[string]interface{} {
	if pb == nil {
		return map[string]interface{}{}
	}
	return buildUpdatesFromPlan(getUpdatesPlan[P](false), unsafe.Pointer(pb))
}

// ModelToUpdates 将 Model 结构体转换为 map[string]interface{}
// 使用 gorm column tag 或 json tag 作为 key，跳过零值字段和不可导出字段
// 字段 key 与偏移按类型缓存于更新计划，热路径零 tag 解析
func ModelToUpdates[M any](m *M) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return buildUpdatesFromPlan(getUpdatesPlan[M](true), unsafe.Pointer(m))
}

// buildUpdatesFromPlan 按预构建计划输出更新字段
// 空值跳过语义与旧反射实现一致：nil 指针、空字符串、零值、空集合不输出
// updatesPresizeThreshold 触发「精确计数预分配」的计划字段数下限
//
// map 的 bucket 分配对容量提示存在台阶：提示 ≤ 8 时不额外分配（单 bucket 内联），
// ≥ 9 时才申请 bucket 数组实测（go1.26.2，插入 4 项）：
//
//	hint=4/6/8 → 0 B/op、0 allocs/op
//	hint=9/12  → 616 B/op、3 allocs/op
//
// 因此只有计划字段数跨过该台阶时，按实际条数预分配才可能省下 bucket；
// 否则计数那一趟是纯开销（小结构体的字段填充率通常还很高）
const updatesPresizeThreshold = 8

// buildUpdatesFromPlan 按预构建计划输出更新字段
// 空值跳过语义与旧反射实现一致：nil 指针、空字符串、零值、空集合不输出
func buildUpdatesFromPlan(plan *updatesPlan, base unsafe.Pointer) map[string]interface{} {
	// 计划覆盖该类型的全部字段，而偏量更新通常只写其中少数几个：
	// 按 len(plan.fields) 预分配会为用不到的条目多申请一级 bucket
	// 跨过 bucket 台阶时先数一遍非空字段再精确预分配——empty 是零分配特化闭包
	// （一次偏移读 + 一次比较，约 1–2ns/字段），这一趟远比多出的 bucket 便宜
	// 未跨台阶则直接按计划长度预分配，不付计数开销（见 updatesPresizeThreshold）
	size := len(plan.fields)
	if size > updatesPresizeThreshold {
		n := 0
		for i := range plan.fields {
			if !plan.fields[i].empty(base) {
				n++
			}
		}
		size = n
	}

	updates := make(map[string]interface{}, size)
	for i := range plan.fields {
		entry := &plan.fields[i]
		if entry.empty(base) {
			continue
		}
		updates[entry.key] = entry.emit(base)
	}
	return updates
}
