/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-06-16 21:08:23
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-06-16 21:08:23
 * @FilePath: \go-toolbox\pkg\pbmo\updates_plan.go
 * @Description: Updates 构建计划缓存 - 按类型预解析字段 key 与偏移量
 * 热路径零 tag 解析、零 reflect.Value 构造，空值判断按字段种类特化为直读比较
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"time"
	"unsafe"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// updatesPlanKey 更新计划缓存 key（rtype 指针 + 构建方向位）
// reflect.Type 底层 *rtype 指针至少 2 字节对齐，低位恒为 0，可安全承载方向标记
// 单字 key 装入接口为零拷贝直接存储，等值比较退化为一次指针比较，热路径零装箱
type updatesPlanKey uintptr

const (
	updatesPlanPBBit uintptr = 1 // PB 方向标记位，Model 方向保持纯 rtype 指针
)

// updatesPlanKeyOf 生成更新计划缓存 key，顺带返回解引用后的结构体类型
func updatesPlanKeyOf[M any](isModel bool) (updatesPlanKey, reflect.Type) {
	t := reflect.TypeFor[M]()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	k := updatesPlanKey(ifaceDataPtr(t))
	if !isModel {
		k |= updatesPlanKey(updatesPlanPBBit)
	}
	return k, t
}

// updateFieldEntry 单字段的更新构建计划
type updateFieldEntry struct {
	key    string                                // 预解析的 map key（proto name / gorm column / json tag）
	offset uintptr                               // 字段在结构体中的偏移量
	empty  func(base unsafe.Pointer) bool        // 特化空值判断（零分配）
	emit   func(base unsafe.Pointer) interface{} // 装箱并解包，输出最终 map 值
}

// updatesPlan 单类型的更新构建计划
type updatesPlan struct {
	fields []updateFieldEntry
}

// updatesPlans 全局更新计划缓存，每个类型只构建一次
var updatesPlans syncMap[updatesPlanKey, *updatesPlan]

// stringWrapperPtrType *wrapperspb.StringValue 类型，空值判断走专用语义
var stringWrapperPtrType = reflect.TypeOf((*wrapperspb.StringValue)(nil))

// wrapperspbPkgPath protobuf wrapper 包路径
const wrapperspbPkgPath = "google.golang.org/protobuf/types/known/wrapperspb"

// getUpdatesPlan 获取（或构建）指定类型的更新计划
func getUpdatesPlan[M any](isModel bool) *updatesPlan {
	key, t := updatesPlanKeyOf[M](isModel)
	if p, ok := updatesPlans.Load(key); ok {
		return p
	}
	p, _ := updatesPlans.LoadOrStore(key, buildUpdatesPlan(t, isModel))
	return p
}

// buildUpdatesPlan 构建更新计划：解析 key、特化空值判断与值输出
// 仅保留可导出且 key 可解析的字段，与旧反射实现的遍历语义一致
func buildUpdatesPlan(t reflect.Type, isModel bool) *updatesPlan {
	plan := &updatesPlan{fields: make([]updateFieldEntry, 0, t.NumField())}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		var key string
		if isModel {
			key = resolveModelKey(field)
			if key == "-" {
				continue
			}
		} else {
			key = resolvePBKey(field)
			if key == "" || key == "-" {
				continue
			}
		}
		plan.fields = append(plan.fields, updateFieldEntry{
			key:    key,
			offset: field.Offset,
			empty:  makeUpdateEmptyCheck(field.Type, field.Offset),
			emit:   makeUpdateEmitter(field.Type, field.Offset, isModel),
		})
	}
	return plan
}

// maxUpdateCheckDepth 单字段 struct 下钻深度上限，防自引用类型递归爆栈
const maxUpdateCheckDepth = 8

// makeUpdateEmptyCheck 按字段类型生成特化空值判断
func makeUpdateEmptyCheck(ft reflect.Type, off uintptr) func(unsafe.Pointer) bool {
	return makeUpdateEmptyCheckAt(ft, off, 0)
}

// isEmptyAt 可比较类型零值判断闭包工厂，按目标类型从偏移读值与零值比较
func isEmptyAt[T comparable](off uintptr) func(unsafe.Pointer) bool {
	return func(base unsafe.Pointer) bool {
		var zero T
		return *(*T)(addPtr(base, off)) == zero
	}
}

// makeUpdateEmptyCheckAt 按字段类型生成特化空值判断，depth 记录单字段 struct 下钻层数
// 语义与 isEmptyUpdateValue 一致：nil 指针、空字符串、零值、空集合、零值时间视为空
func makeUpdateEmptyCheckAt(ft reflect.Type, off uintptr, depth int) func(unsafe.Pointer) bool {
	switch ft.Kind() {
	case reflect.String:
		return func(base unsafe.Pointer) bool {
			s := *(*string)(addPtr(base, off))
			return isEmptyUpdateString(s)
		}
	case reflect.Bool:
		return isEmptyAt[bool](off)
	case reflect.Int:
		return isEmptyAt[int](off)
	case reflect.Int8:
		return isEmptyAt[int8](off)
	case reflect.Int16:
		return isEmptyAt[int16](off)
	case reflect.Int32:
		return isEmptyAt[int32](off)
	case reflect.Int64:
		return isEmptyAt[int64](off)
	case reflect.Uint:
		return isEmptyAt[uint](off)
	case reflect.Uint8:
		return isEmptyAt[uint8](off)
	case reflect.Uint16:
		return isEmptyAt[uint16](off)
	case reflect.Uint32:
		return isEmptyAt[uint32](off)
	case reflect.Uint64:
		return isEmptyAt[uint64](off)
	case reflect.Uintptr:
		return isEmptyAt[uintptr](off)
	case reflect.Float32:
		return isEmptyAt[float32](off)
	case reflect.Float64:
		return isEmptyAt[float64](off)
	case reflect.Ptr:
		if ft == stringWrapperPtrType {
			return func(base unsafe.Pointer) bool {
				sv := *(**wrapperspb.StringValue)(addPtr(base, off))
				return sv == nil || isEmptyUpdateString(sv.Value)
			}
		}
		elemCheck := makeUpdateDerefEmptyCheck(ft.Elem(), depth)
		return func(base unsafe.Pointer) bool {
			p := *(*unsafe.Pointer)(addPtr(base, off))
			return p == nil || elemCheck(p)
		}
	case reflect.Slice:
		// slice 运行时头直读 Len，免去逐次 reflect.NewAt 与 ptrTo 缓存查找
		return func(base unsafe.Pointer) bool {
			return (*sliceHeader)(addPtr(base, off)).Len == 0
		}
	case reflect.Map, reflect.Array:
		// map 头依赖 runtime 内部布局不直读，数组恒非空语义与原实现一致
		return func(base unsafe.Pointer) bool {
			return reflect.NewAt(ft, addPtr(base, off)).Elem().Len() == 0
		}
	case reflect.Struct:
		if ft == timeType {
			return func(base unsafe.Pointer) bool {
				tm := *(*time.Time)(addPtr(base, off))
				return tm.IsZero() || tm.Unix() <= 0
			}
		}
		// 单导出字段 struct 下钻载体字段复用已有特化，语义与 IsZero 全字段零值一致
		// 如 sqltypes.JSON[T]{Data T} 包装类型，避免 IsZero 的递归反射遍历
		if depth < maxUpdateCheckDepth && ft.NumField() == 1 {
			if f := ft.Field(0); f.IsExported() {
				return makeUpdateEmptyCheckAt(f.Type, off+f.Offset, depth+1)
			}
		}
		return func(base unsafe.Pointer) bool {
			return reflect.NewAt(ft, addPtr(base, off)).Elem().IsZero()
		}
	default:
		// Interface / Func / Chan / 复数等动态语义类型，回退通用判断保证语义一致
		return func(base unsafe.Pointer) bool {
			return isEmptyReflectValue(reflect.NewAt(ft, addPtr(base, off)).Elem())
		}
	}
}

// makeUpdateDerefEmptyCheck 生成对指针已解引用值的空值判断
// 递归展开多级指针，与 IsEmptyValue 的循环解引用语义一致
func makeUpdateDerefEmptyCheck(et reflect.Type, depth int) func(unsafe.Pointer) bool {
	if et.Kind() == reflect.Ptr {
		inner := makeUpdateDerefEmptyCheck(et.Elem(), depth)
		return func(p unsafe.Pointer) bool {
			pp := *(*unsafe.Pointer)(p)
			return pp == nil || inner(pp)
		}
	}
	return makeUpdateEmptyCheckAt(et, 0, depth)
}

// makeUpdateEmitter 按字段类型生成值输出闭包（装箱 + 解包）
// PB 侧 wrapper 输出 Value 字段值，Model 侧指针解引用输出底层值
// 基础类型与 time.Time 直接从偏移读值装箱，避免 reflect.Value 构造开销
func makeUpdateEmitter(ft reflect.Type, off uintptr, isModel bool) func(unsafe.Pointer) interface{} {
	if !isModel && ft.Kind() == reflect.Ptr &&
		ft.Elem().Kind() == reflect.Struct && ft.Elem().PkgPath() == wrapperspbPkgPath {
		if valueField, ok := ft.Elem().FieldByName("Value"); ok {
			vt, vo := valueField.Type, valueField.Offset
			if emit := makeDirectEmitter(vt, 0); emit != nil {
				return func(base unsafe.Pointer) interface{} {
					p := *(*unsafe.Pointer)(addPtr(base, off))
					if p == nil {
						return nil
					}
					return emit(addPtr(p, vo))
				}
			}
			return func(base unsafe.Pointer) interface{} {
				p := *(*unsafe.Pointer)(addPtr(base, off))
				if p == nil {
					return nil
				}
				return reflect.NewAt(vt, addPtr(p, vo)).Elem().Interface()
			}
		}
	}
	if isModel && ft.Kind() == reflect.Ptr {
		et := ft.Elem()
		if emit := makeDirectEmitter(et, 0); emit != nil {
			return func(base unsafe.Pointer) interface{} {
				p := *(*unsafe.Pointer)(addPtr(base, off))
				if p == nil {
					// typed nil 保持原值语义，输出装箱后的指针本身
					return reflect.NewAt(ft, addPtr(base, off)).Elem().Interface()
				}
				return emit(p)
			}
		}
		return func(base unsafe.Pointer) interface{} {
			p := *(*unsafe.Pointer)(addPtr(base, off))
			if p == nil {
				// typed nil 保持原值语义，输出装箱后的指针本身
				return reflect.NewAt(ft, addPtr(base, off)).Elem().Interface()
			}
			return reflect.NewAt(et, p).Elem().Interface()
		}
	}
	if ft.Kind() == reflect.Ptr {
		// PB 方向非 wrapper 指针直接输出指针装箱，构建期缓存类型词
		// 消除逐次 reflect.NewAt 与 packEface 反射链
		typ := ifaceTypeWord(ft)
		return func(base unsafe.Pointer) interface{} {
			p := *(*unsafe.Pointer)(addPtr(base, off))
			if p == nil {
				// typed nil 保持原值语义
				return reflect.NewAt(ft, addPtr(base, off)).Elem().Interface()
			}
			return makeIface(typ, p)
		}
	}
	if emit := makeDirectEmitter(ft, off); emit != nil {
		return emit
	}
	return func(base unsafe.Pointer) interface{} {
		return reflect.NewAt(ft, addPtr(base, off)).Elem().Interface()
	}
}

// emitDirect 基础类型直读装箱闭包工厂，按目标类型从偏移读值并装箱
func emitDirect[T any](off uintptr) func(unsafe.Pointer) interface{} {
	return func(base unsafe.Pointer) interface{} {
		return *(*T)(addPtr(base, off))
	}
}

// makeDirectEmitter 基础类型与 time.Time 的直读装箱闭包
// 按字段 Kind 直接从偏移读取并装箱，无法特化时返回 nil 交由 reflect 回退
func makeDirectEmitter(ft reflect.Type, off uintptr) func(unsafe.Pointer) interface{} {
	switch ft.Kind() {
	case reflect.String:
		return emitDirect[string](off)
	case reflect.Bool:
		return emitDirect[bool](off)
	case reflect.Int:
		return emitDirect[int](off)
	case reflect.Int8:
		return emitDirect[int8](off)
	case reflect.Int16:
		return emitDirect[int16](off)
	case reflect.Int32:
		return emitDirect[int32](off)
	case reflect.Int64:
		return emitDirect[int64](off)
	case reflect.Uint:
		return emitDirect[uint](off)
	case reflect.Uint8:
		return emitDirect[uint8](off)
	case reflect.Uint16:
		return emitDirect[uint16](off)
	case reflect.Uint32:
		return emitDirect[uint32](off)
	case reflect.Uint64:
		return emitDirect[uint64](off)
	case reflect.Uintptr:
		return emitDirect[uintptr](off)
	case reflect.Float32:
		return emitDirect[float32](off)
	case reflect.Float64:
		return emitDirect[float64](off)
	case reflect.Interface:
		if ft.NumMethod() == 0 {
			return emitDirect[interface{}](off)
		}
		return nil
	case reflect.Struct:
		if ft == timeType {
			return emitDirect[time.Time](off)
		}
		return nil
	default:
		return nil
	}
}
