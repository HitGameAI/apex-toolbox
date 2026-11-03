/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-04-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-04-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\wrappers.go
 * @Description: Protobuf Wrappers 类型自动转换 - *T ↔ *wrapperspb.XxxValue
 * 同时提供公开的 DerefXxxVal / XxxValuePtr / PtrToXxxValue / XxxValueToPtr 辅助函数
 * 在 convertField 中自动识别并转换，无需手动编写辅助函数
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"unsafe"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// wrapperspb.XxxValue 指针类型变量，用于反射类型匹配
var (
	int32ValuePtrType  = reflect.TypeOf((*wrapperspb.Int32Value)(nil))  // *wrapperspb.Int32Value 类型
	int64ValuePtrType  = reflect.TypeOf((*wrapperspb.Int64Value)(nil))  // *wrapperspb.Int64Value 类型
	uint32ValuePtrType = reflect.TypeOf((*wrapperspb.UInt32Value)(nil)) // *wrapperspb.UInt32Value 类型
	uint64ValuePtrType = reflect.TypeOf((*wrapperspb.UInt64Value)(nil)) // *wrapperspb.UInt64Value 类型
	floatValuePtrType  = reflect.TypeOf((*wrapperspb.FloatValue)(nil))  // *wrapperspb.FloatValue 类型
	doubleValuePtrType = reflect.TypeOf((*wrapperspb.DoubleValue)(nil)) // *wrapperspb.DoubleValue 类型
	boolValuePtrType   = reflect.TypeOf((*wrapperspb.BoolValue)(nil))   // *wrapperspb.BoolValue 类型
	stringValuePtrType = reflect.TypeOf((*wrapperspb.StringValue)(nil)) // *wrapperspb.StringValue 类型
)

// wrapPtrCopy 生成 *T → *W 包装闭包：nil 透传，非 nil 分配包装对象并按偏移写入 Value
func wrapPtrCopy[S any, W any](valueOff, srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		srcPtr := *(**S)(addPtr(srcBase, srcOffset))
		if srcPtr == nil {
			*(**W)(addPtr(dstBase, dstOffset)) = nil
			return
		}
		w := new(W)
		*(*S)(addPtr(unsafe.Pointer(w), valueOff)) = *srcPtr
		*(**W)(addPtr(dstBase, dstOffset)) = w
	}
}

// unwrapPtrCopy 生成 *W → *T 解包闭包：nil 透传，非 nil 读取 Value 并取地址
func unwrapPtrCopy[S any, W any](valueOff, srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		srcPtr := *(**W)(addPtr(srcBase, srcOffset))
		if srcPtr == nil {
			*(**S)(addPtr(dstBase, dstOffset)) = nil
			return
		}
		val := *(*S)(addPtr(unsafe.Pointer(srcPtr), valueOff))
		*(**S)(addPtr(dstBase, dstOffset)) = &val
	}
}

// wrapValCopy 生成 T → *W 包装闭包：分配包装对象并按偏移写入 Value
func wrapValCopy[S any, W any](valueOff, srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		w := new(W)
		*(*S)(addPtr(unsafe.Pointer(w), valueOff)) = *(*S)(addPtr(srcBase, srcOffset))
		*(**W)(addPtr(dstBase, dstOffset)) = w
	}
}

// unwrapValCopy 生成 *W → T 解包闭包：nil 写零值，非 nil 读取 Value
func unwrapValCopy[S any, W any](valueOff, srcOffset, dstOffset uintptr) fieldCopyFunc {
	return func(srcBase, dstBase unsafe.Pointer) {
		srcPtr := *(**W)(addPtr(srcBase, srcOffset))
		if srcPtr == nil {
			var zero S
			*(*S)(addPtr(dstBase, dstOffset)) = zero
			return
		}
		*(*S)(addPtr(dstBase, dstOffset)) = *(*S)(addPtr(unsafe.Pointer(srcPtr), valueOff))
	}
}

// wrapperConverter 单个 Wrapper 类型的转换器
// wrapperType: *wrapperspb.XxxValue 的反射类型
// ptrType: *T（Go 基本类型指针）的反射类型
// wrap: 将 *T 包装为 *wrapperspb.XxxValue（Model → PB 方向）
// unwrap: 将 *wrapperspb.XxxValue 解包为 *T（PB → Model 方向）
type wrapperConverter struct {
	wrapperType reflect.Type                      // *wrapperspb.XxxValue 反射类型
	ptrType     reflect.Type                      // *T Go 基本类型指针反射类型
	valueType   reflect.Type                      // T Go 基本类型值反射类型
	wrap        func(reflect.Value) reflect.Value // *T → *wrapperspb.XxxValue 包装函数
	unwrap      func(reflect.Value) reflect.Value // *wrapperspb.XxxValue → *T 解包函数
	valueWrap   func(reflect.Value) reflect.Value // T → *wrapperspb.XxxValue 值类型包装函数
	valueUnwrap func(reflect.Value) reflect.Value // *wrapperspb.XxxValue → T 值类型解包函数
}

// wrapperConverters 所有 Wrapper 类型的转换器列表，在 init 中初始化
var wrapperConverters []wrapperConverter

// wrapperLookupMap 用于 O(1) 类型查找的映射表
var (
	wrapperByPtrType     map[reflect.Type]*wrapperConverter // 指针类型到转换器的映射
	wrapperByWrapperType map[reflect.Type]*wrapperConverter // 包装类型到转换器的映射
	wrapperByValueType   map[reflect.Type]*wrapperConverter // 值类型到转换器的映射
)

// newWrapperConverter 以泛型工厂构建单个 Wrapper 类型的反射转换器
// wrap 为对应的 wrapperspb 构造函数（如 wrapperspb.Int32），四个方向闭包由工厂统一生成
func newWrapperConverter[S any, W any](wrap func(S) *W, unwrap func(*W) S) wrapperConverter {
	wrapperType := reflect.TypeOf((*W)(nil))
	ptrType := reflect.TypeOf((*S)(nil))
	valueType := ptrType.Elem()
	return wrapperConverter{
		wrapperType: wrapperType,
		ptrType:     ptrType,
		valueType:   valueType,
		wrap: func(v reflect.Value) reflect.Value {
			if v.IsNil() {
				return reflect.Zero(wrapperType)
			}
			return reflect.ValueOf(wrap(*v.Interface().(*S)))
		},
		unwrap: func(v reflect.Value) reflect.Value {
			if v.IsNil() {
				return reflect.Zero(ptrType)
			}
			val := unwrap(v.Interface().(*W))
			return reflect.ValueOf(&val)
		},
		valueWrap: func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(wrap(v.Interface().(S)))
		},
		valueUnwrap: func(v reflect.Value) reflect.Value {
			if v.IsNil() {
				return reflect.Zero(valueType)
			}
			return reflect.ValueOf(unwrap(v.Interface().(*W)))
		},
	}
}

// init 注册所有 *T ↔ *wrapperspb.XxxValue 的转换器
// 当源值为 nil 指针时，wrap/unwrap 返回对应类型的零值（nil 指针）
func init() {
	wrapperConverters = []wrapperConverter{
		newWrapperConverter(wrapperspb.Int32, func(w *wrapperspb.Int32Value) int32 { return w.Value }),
		newWrapperConverter(wrapperspb.Int64, func(w *wrapperspb.Int64Value) int64 { return w.Value }),
		newWrapperConverter(wrapperspb.UInt32, func(w *wrapperspb.UInt32Value) uint32 { return w.Value }),
		newWrapperConverter(wrapperspb.UInt64, func(w *wrapperspb.UInt64Value) uint64 { return w.Value }),
		newWrapperConverter(wrapperspb.Float, func(w *wrapperspb.FloatValue) float32 { return w.Value }),
		newWrapperConverter(wrapperspb.Double, func(w *wrapperspb.DoubleValue) float64 { return w.Value }),
		newWrapperConverter(wrapperspb.Bool, func(w *wrapperspb.BoolValue) bool { return w.Value }),
		newWrapperConverter(wrapperspb.String, func(w *wrapperspb.StringValue) string { return w.Value }),
	}

	// 初始化指针类型到包装器的映射
	wrapperByPtrType = make(map[reflect.Type]*wrapperConverter, len(wrapperConverters))
	wrapperByWrapperType = make(map[reflect.Type]*wrapperConverter, len(wrapperConverters))
	wrapperByValueType = make(map[reflect.Type]*wrapperConverter, len(wrapperConverters))
	for i := range wrapperConverters {
		wc := &wrapperConverters[i]
		wrapperByPtrType[wc.ptrType] = wc
		wrapperByWrapperType[wc.wrapperType] = wc
		wrapperByValueType[wc.valueType] = wc
	}
}

// tryConvertWrapper 尝试在 *T 和 *wrapperspb.XxxValue 之间自动转换
// 使用 O(1) 的 map 查找替代线性扫描，提升性能
// 返回值：
//   - handled: 是否已处理该转换（true 表示匹配成功并完成转换）
//   - error: 转换过程中的错误
func tryConvertWrapper(src, dst reflect.Value) (bool, error) {
	srcType := src.Type()
	dstType := dst.Type()

	if wc, ok := wrapperByPtrType[srcType]; ok {
		if dstType == wc.wrapperType {
			result := wc.wrap(src)
			if result.IsValid() {
				dst.Set(result)
			}
			return true, nil
		}
	}

	// 包装类型到包装器的映射
	if wc, ok := wrapperByWrapperType[srcType]; ok {
		if dstType == wc.ptrType {
			result := wc.unwrap(src)
			if result.IsValid() {
				dst.Set(result)
			}
			return true, nil
		}

		// 值类型到包装器的映射
		if dstType == wc.valueType {
			result := wc.valueUnwrap(src)
			if result.IsValid() {
				dst.Set(result)
			}
			return true, nil
		}
	}

	// 值类型到包装器的映射
	if wc, ok := wrapperByValueType[srcType]; ok {
		if dstType == wc.wrapperType {
			result := wc.valueWrap(src)
			dst.Set(result)
			return true, nil
		}
	}

	return false, nil
}

// derefWrapperVal 从 *wrapperspb.XxxValue 提取基础值，nil 指针返回零值
func derefWrapperVal[S any, W any, P interface {
	*W
	GetValue() S
}](v P) S {
	if v == nil {
		var zero S
		return zero
	}
	return v.GetValue()
}

// wrapperValToPtr 将 *wrapperspb.XxxValue 解包为基础值指针，nil 指针透传
func wrapperValToPtr[S any, W any, P interface {
	*W
	GetValue() S
}](v P) *S {
	if v == nil {
		return nil
	}
	val := v.GetValue()
	return &val
}

// ptrToWrapperVal 将基础值指针包装为 *wrapperspb.XxxValue，nil 指针透传
func ptrToWrapperVal[S any, W any](wrap func(S) *W, v *S) *W {
	if v == nil {
		return nil
	}
	return wrap(*v)
}

// DerefInt32Val 从 *wrapperspb.Int32Value 中提取整数值
// 当源值为 nil 指针时，DerefInt32Val 返回 0
func DerefInt32Val(v *wrapperspb.Int32Value) int32 {
	return derefWrapperVal(v)
}

// DerefInt64Val 从 *wrapperspb.Int64Value 中提取整数值
// 当源值为 nil 指针时，DerefInt64Val 返回 0
func DerefInt64Val(v *wrapperspb.Int64Value) int64 {
	return derefWrapperVal(v)
}

// DerefUInt32Val 从 *wrapperspb.UInt32Value 中提取无符号整数值
// 当源值为 nil 指针时，DerefUInt32Val 返回 0
func DerefUInt32Val(v *wrapperspb.UInt32Value) uint32 {
	return derefWrapperVal(v)
}

// DerefUInt64Val 从 *wrapperspb.UInt64Value 中提取无符号整数值
// 当源值为 nil 指针时，DerefUInt64Val 返回 0
func DerefUInt64Val(v *wrapperspb.UInt64Value) uint64 {
	return derefWrapperVal(v)
}

// DerefFloatVal 从 *wrapperspb.FloatValue 中提取浮点数值
// 当源值为 nil 指针时，DerefFloatVal 返回 0.0
func DerefFloatVal(v *wrapperspb.FloatValue) float32 {
	return derefWrapperVal(v)
}

// DerefDoubleVal 从 *wrapperspb.DoubleValue 中提取浮点数值
// 当源值为 nil 指针时，DerefDoubleVal 返回 0.0
func DerefDoubleVal(v *wrapperspb.DoubleValue) float64 {
	return derefWrapperVal(v)
}

// DerefBoolVal 从 *wrapperspb.BoolValue 中提取布尔值
// 当源值为 nil 指针时，DerefBoolVal 返回 false
func DerefBoolVal(v *wrapperspb.BoolValue) bool {
	return derefWrapperVal(v)
}

// DerefStringVal 从 *wrapperspb.StringValue 中提取字符串值
// 当源值为 nil 指针时，DerefStringVal 返回空字符串 ""
func DerefStringVal(v *wrapperspb.StringValue) string {
	return derefWrapperVal(v)
}

// DerefSlice 将 []*T 转换为 []T，nil 元素转为零值
func DerefSlice[T any](ptrs []*T) []T {
	result := make([]T, len(ptrs))
	for i, p := range ptrs {
		if p != nil {
			result[i] = *p
		}
	}
	return result
}

// Int32ValueToPtr 将 *wrapperspb.Int32Value 转换为 *int32
// 当源值为 nil 指针时，Int32ValueToPtr 返回 nil 指针
func Int32ValueToPtr(v *wrapperspb.Int32Value) *int32 {
	return wrapperValToPtr(v)
}

// Int64ValueToPtr 将 *wrapperspb.Int64Value 转换为 *int64
// 当源值为 nil 指针时，Int64ValueToPtr 返回 nil 指针
func Int64ValueToPtr(v *wrapperspb.Int64Value) *int64 {
	return wrapperValToPtr(v)
}

// UInt32ValueToPtr 将 *wrapperspb.UInt32Value 转换为 *uint32
// 当源值为 nil 指针时，UInt32ValueToPtr 返回 nil 指针
func UInt32ValueToPtr(v *wrapperspb.UInt32Value) *uint32 {
	return wrapperValToPtr(v)
}

// UInt64ValueToPtr 将 *wrapperspb.UInt64Value 转换为 *uint64
// 当源值为 nil 指针时，UInt64ValueToPtr 返回 nil 指针
func UInt64ValueToPtr(v *wrapperspb.UInt64Value) *uint64 {
	return wrapperValToPtr(v)
}

// FloatValueToPtr 将 *wrapperspb.FloatValue 转换为 *float32
// 当源值为 nil 指针时，FloatValueToPtr 返回 nil 指针
func FloatValueToPtr(v *wrapperspb.FloatValue) *float32 {
	return wrapperValToPtr(v)
}

// DoubleValueToPtr 将 *wrapperspb.DoubleValue 转换为 *float64
// 当源值为 nil 指针时，DoubleValueToPtr 返回 nil 指针
func DoubleValueToPtr(v *wrapperspb.DoubleValue) *float64 {
	return wrapperValToPtr(v)
}

// BoolValueToPtr 将 *wrapperspb.BoolValue 转换为 *bool
// 当源值为 nil 指针时，BoolValueToPtr 返回 nil 指针
func BoolValueToPtr(v *wrapperspb.BoolValue) *bool {
	return wrapperValToPtr(v)
}

// StringValueToPtr 将 *wrapperspb.StringValue 转换为 *string
// 当源值为 nil 指针时，StringValueToPtr 返回 nil 指针
func StringValueToPtr(v *wrapperspb.StringValue) *string {
	return wrapperValToPtr(v)
}

// PtrToInt32Value 将 *int32 转换为 *wrapperspb.Int32Value
// 当源值为 nil 指针时，PtrToInt32Value 返回 nil 指针
func PtrToInt32Value(v *int32) *wrapperspb.Int32Value {
	return ptrToWrapperVal(wrapperspb.Int32, v)
}

// PtrToInt64Value 将 *int64 转换为 *wrapperspb.Int64Value
// 当源值为 nil 指针时，PtrToInt64Value 返回 nil 指针
func PtrToInt64Value(v *int64) *wrapperspb.Int64Value {
	return ptrToWrapperVal(wrapperspb.Int64, v)
}

// PtrToUInt32Value 将 *uint32 转换为 *wrapperspb.UInt32Value
// 当源值为 nil 指针时，PtrToUInt32Value 返回 nil 指针
func PtrToUInt32Value(v *uint32) *wrapperspb.UInt32Value {
	return ptrToWrapperVal(wrapperspb.UInt32, v)
}

// PtrToUInt64Value 将 *uint64 转换为 *wrapperspb.UInt64Value
// 当源值为 nil 指针时，PtrToUInt64Value 返回 nil 指针
func PtrToUInt64Value(v *uint64) *wrapperspb.UInt64Value {
	return ptrToWrapperVal(wrapperspb.UInt64, v)
}

// PtrToFloatValue 将 *float32 转换为 *wrapperspb.FloatValue
// 当源值为 nil 指针时，PtrToFloatValue 返回 nil 指针
func PtrToFloatValue(v *float32) *wrapperspb.FloatValue {
	return ptrToWrapperVal(wrapperspb.Float, v)
}

// PtrToDoubleValue 将 *float64 转换为 *wrapperspb.DoubleValue
// 当源值为 nil 指针时，PtrToDoubleValue 返回 nil 指针
func PtrToDoubleValue(v *float64) *wrapperspb.DoubleValue {
	return ptrToWrapperVal(wrapperspb.Double, v)
}

// PtrToBoolValue 将 *bool 转换为 *wrapperspb.BoolValue
// 当源值为 nil 指针时，PtrToBoolValue 返回 nil 指针
func PtrToBoolValue(v *bool) *wrapperspb.BoolValue {
	return ptrToWrapperVal(wrapperspb.Bool, v)
}

// PtrToStringValue 将 *string 转换为 *wrapperspb.StringValue
// 当源值为 nil 指针时，PtrToStringValue 返回 nil 指针
func PtrToStringValue(v *string) *wrapperspb.StringValue {
	return ptrToWrapperVal(wrapperspb.String, v)
}
