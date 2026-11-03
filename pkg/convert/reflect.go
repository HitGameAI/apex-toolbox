/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-09 01:15:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-20 13:22:11
 * @FilePath: \go-toolbox\pkg\convert\reflect.go
 * @Description:
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package convert

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// 错误信息常量，方便统一维护和复用
var (
	ErrDstNilPointer        = errors.New("dst must be a non-nil pointer")
	ErrDstNil               = errors.New("dst must not be nil")
	ErrSrcNil               = errors.New("src must not be nil")
	ErrTypeMismatchStrict   = "type mismatch: cannot assign %s to %s"
	ErrUnsupportedType      = "unsupported or mismatched type: %s -> %s"
	ErrNegativeIntToUint    = "negative int %d cannot assign to %s"
	ErrTypeMismatchToBool   = "type mismatch: cannot assign %s to bool"
	ErrTypeMismatchToString = "type mismatch: cannot assign %s to string"
	ErrTypeMismatchToSlice  = "type mismatch: cannot assign %s to slice %s"
	ErrTypeMismatchToMap    = "type mismatch: cannot assign %s to map %s"
)

// TransformFieldsOptions 定义转换选项
type TransformFieldsOptions struct {
	StrictTypeCheck bool   // 是否严格类型检查，默认 false
	TimeFormat      string // 时间格式，默认 time.DateTime
	TransTagName    string // 支持字段标签名
}

// SetStrictTypeCheck 设置严格类型检查
func (o *TransformFieldsOptions) SetStrictTypeCheck(strict bool) *TransformFieldsOptions {
	o.StrictTypeCheck = strict
	return o
}

// SetTimeFormat 设置时间格式
func (o *TransformFieldsOptions) SetTimeFormat(format string) *TransformFieldsOptions {
	o.TimeFormat = format
	return o
}

// Transformer 负责执行转换操作，封装目标对象、源数据及转换选项
type Transformer struct {
	dst  any                     // 目标对象，通常是指向结构体的指针，用于接收转换结果
	src  any                     // 源数据，可以是结构体、map 等，用于读取数据进行转换
	opts *TransformFieldsOptions // 转换选项，控制转换行为，如严格类型检查、时间格式等
	mu   sync.RWMutex            // 读写锁，保护 Transformer 内部状态的并发安全
}

// NewTransformer 创建空 Transformer 实例
func NewTransformer() *Transformer {
	return &Transformer{}
}

// SetDst 设置目标结构体指针，必须是非 nil 指针
func (t *Transformer) SetDst(dst any) *Transformer {
	return syncx.WithLockReturnValue(&t.mu, func() *Transformer {
		t.dst = dst
		return t
	})
}

// SetSrc 设置源数据，支持结构体、map 等
func (t *Transformer) SetSrc(src any) *Transformer {
	return syncx.WithLockReturnValue(&t.mu, func() *Transformer {
		t.src = src
		return t
	})
}

// SetOptions 设置转换选项，支持链式调用
func (t *Transformer) SetOptions(opts *TransformFieldsOptions) *Transformer {
	return syncx.WithLockReturnValue(&t.mu, func() *Transformer {
		t.opts = opts
		return t
	})
}

// GetDst 获取目标结构体指针，线程安全
func (t *Transformer) GetDst() any {
	return syncx.WithRLockReturnValue(&t.mu, func() any {
		return t.dst
	})
}

// GetSrc 获取源数据，线程安全
func (t *Transformer) GetSrc() any {
	return syncx.WithRLockReturnValue(&t.mu, func() any {
		return t.src
	})
}

// GetOptions 获取转换选项，线程安全
func (t *Transformer) GetOptions() *TransformFieldsOptions {
	return syncx.WithRLockReturnValue(&t.mu, func() *TransformFieldsOptions {
		return t.opts
	})
}

// Transform 执行转换操作，将源数据转换到目标对象，返回转换过程中的错误
// 该方法使用带尝试加锁的方式保证并发安全，未能获取锁时返回 ErrLockNotAcquired
func (t *Transformer) Transform() error {
	return syncx.WithTryLock(&t.mu, func() error {
		if t.dst == nil {
			return ErrDstNil
		}

		// 默认转换选项
		defaultOptions := &TransformFieldsOptions{}

		if t.opts == nil {
			t.opts = defaultOptions
		}

		if t.opts.TimeFormat == "" {
			t.opts.TimeFormat = time.DateTime
		}

		dstVal := reflect.ValueOf(t.dst)
		// 目标必须是非 nil 指针
		if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
			return ErrDstNilPointer
		}

		if t.src == nil {
			return ErrSrcNil
		}
		srcVal := reflect.ValueOf(t.src)

		// 调用核心转换函数，执行递归字段赋值
		if err := transformValue(dstVal.Elem(), srcVal, t.opts); err != nil {
			return err
		}
		return nil
	})
}

// TransformFields 兼容旧用法，返回 error
func TransformFields(dst any, src any, opts *TransformFieldsOptions) error {
	return NewTransformer().
		SetDst(dst).
		SetSrc(src).
		SetOptions(opts).
		Transform()
}

// transformValue 递归执行值转换
//
// 参数:
//   - dst: reflect.Value，目标值，必须是可设置的（settable）
//   - src: reflect.Value，源值，可以是结构体、map、基本类型等
//   - opts: *TransformFieldsOptions，转换选项，支持严格类型检查、时间格式、标签名等
//
// 返回:
//   - error，转换过程中遇到的错误，若成功返回 nil
//
// 说明:
//
//	本函数通过反射递归遍历目标结构体字段，根据 opts.TransTagName 标签决定字段映射关系。
//	支持基本类型、time.Time->string、指针、slice、map 等多种类型转换。
//	严格模式下遇到类型不匹配会返回错误，非严格模式则忽略类型不匹配。
func transformValue(dst, src reflect.Value, opts *TransformFieldsOptions) error {
	// 如果源值无效（零值、未初始化等），直接返回，不做赋值
	if !src.IsValid() {
		return nil
	}

	// 自动解引用源指针，直到得到非指针类型的值
	for src.Kind() == reflect.Ptr {
		if src.IsNil() {
			// 如果指针为 nil，则将目标值置为对应类型的零值
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		src = src.Elem()
	}

	// 这里只做基础的类型匹配检查，复杂业务检查放在后续逻辑中
	if err := checkStrictTypeMatch(src, dst, opts); err != nil {
		return err
	}

	// 根据目标值的类型，执行不同的转换逻辑
	switch dst.Kind() {
	case reflect.Struct:
		// 目标是结构体时，且源也是结构体，则递归转换字段
		if src.Kind() == reflect.Struct {
			for i := 0; i < dst.NumField(); i++ {
				dstField := dst.Field(i)            // 目标结构体字段值
				dstFieldType := dst.Type().Field(i) // 目标结构体字段类型信息

				// 跳过不可设置的字段（私有字段或只读字段）
				if !dstField.CanSet() || dstFieldType.PkgPath != "" {
					continue
				}

				// 获取目标字段上的转换标签，支持自定义字段映射
				tag := dstFieldType.Tag.Get(opts.TransTagName)
				if tag == "-" {
					// 标签为 "-" 表示跳过该字段
					continue
				}

				// 默认源字段名为目标字段名
				srcFieldName := dstFieldType.Name
				if tag != "" {
					// 如果标签非空且不是 "-", 使用标签指定的字段名作为源字段名
					srcFieldName = tag
				}

				// 从源结构体中查找对应字段
				srcField := src.FieldByName(srcFieldName)
				if !srcField.IsValid() {
					// 源结构体无对应字段，跳过
					continue
				}

				// 递归转换字段值
				if err := transformValue(dstField, srcField, opts); err != nil {
					return err
				}
			}
			return nil
		}
		// 目标是结构体但源不是结构体，忽略或返回 nil（不做严格报错）

	case reflect.String:
		// 支持将 time.Time 类型转换为字符串格式
		if src.Type() == reflect.TypeOf(time.Time{}) {
			t := src.Interface().(time.Time)
			dst.SetString(t.Format(opts.TimeFormat))
			return nil
		}

		// 源是字符串，直接赋值
		if src.Kind() == reflect.String {
			dst.SetString(src.String())
			return nil
		}

		// 其他类型，调用 MustString 强制转换为字符串
		dst.SetString(MustString(src.Interface(), opts.TimeFormat))
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 使用 MustIntT 进行强制转换为 int64
		val, err := MustIntT[int64](src.Interface(), nil)
		if err == nil {
			dst.SetInt(val)
			return nil
		}
		// 转换失败则忽略赋值

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 使用 MustIntT 先转换为 int64，再赋给无符号类型
		val, err := MustIntT[int64](src.Interface(), nil)
		if err == nil && val >= 0 {
			dst.SetUint(uint64(val))
			return nil
		}
		// 转换失败或负数值，忽略赋值

	case reflect.Float32, reflect.Float64:
		// 使用 MustFloatT 强制转换为 float64
		val, err := MustFloatT[float64](src.Interface(), RoundNone)
		if err == nil {
			dst.SetFloat(val)
			return nil
		}
		// 转换失败忽略赋值

	case reflect.Bool:
		// 使用 MustBool 强制转换为 bool
		val := MustBool(src.Interface())
		dst.SetBool(val)
		return nil

	case reflect.Ptr:
		// 目标是指针类型，先确保指针非 nil，再递归转换指针指向的元素
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return transformValue(dst.Elem(), src, opts)

	case reflect.Slice:
		// 目标是切片，源也是切片时，递归转换每个元素
		if src.Kind() == reflect.Slice {
			if src.IsNil() {
				// 源切片为 nil，目标赋 nil 切片
				dst.Set(reflect.Zero(dst.Type()))
				return nil
			}
			newSlice := reflect.MakeSlice(dst.Type(), src.Len(), src.Cap())
			for i := 0; i < src.Len(); i++ {
				if err := transformValue(newSlice.Index(i), src.Index(i), opts); err != nil {
					return err
				}
			}
			dst.Set(newSlice)
			return nil
		}
		// 源不是切片，忽略赋值

	case reflect.Map:
		// 目标是 map，源也是 map 时，递归转换每个键值对
		if src.Kind() == reflect.Map {
			if src.IsNil() {
				// 源 map 为 nil，目标赋零值
				dst.Set(reflect.Zero(dst.Type()))
				return nil
			}
			newMap := reflect.MakeMapWithSize(dst.Type(), src.Len())
			for _, key := range src.MapKeys() {
				valSrc := src.MapIndex(key)
				valDst := reflect.New(dst.Type().Elem()).Elem()
				if err := transformValue(valDst, valSrc, opts); err != nil {
					return err
				}
				newMap.SetMapIndex(key, valDst)
			}
			dst.Set(newMap)
			return nil
		}
		// 源不是 map，忽略赋值

	default:
		// 其他类型，如果类型完全匹配且可赋值，直接赋值
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
		}
		// 不支持的类型忽略赋值
	}

	return nil
}

// checkStrictTypeMatch 严格模式下判断 src 与 dst 类型是否一致，不一致返回错误
func checkStrictTypeMatch(src, dst reflect.Value, opts *TransformFieldsOptions) error {
	if !opts.StrictTypeCheck {
		return nil
	}

	return types.CheckTypeCompatibility(src.Type(), dst.Type())
}
