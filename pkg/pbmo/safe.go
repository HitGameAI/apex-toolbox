/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\safe.go
 * @Description: 安全转换器 - nil 与 typed-nil 安全的转换入口，零第三方依赖
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
)

// SafeConverter 安全转换器
// 在转换前拦截 nil 与 typed-nil 指针，避免转换引擎 panic
type SafeConverter struct {
	*BidiConverter
}

// NewSafeConverter 创建安全转换器
func NewSafeConverter(pbType, modelType interface{}, opts ...Option) *SafeConverter {
	return &SafeConverter{
		BidiConverter: NewBidiConverter(pbType, modelType, opts...),
	}
}

// checkNotNil 校验输入非 nil 且非 typed-nil 指针，返回 nil 表示有效
func checkNotNil(v interface{}, nilMsg, invalidMsg string) error {
	if v == nil {
		return NewNilValueError("%s", nilMsg)
	}
	// 检测 typed-nil 指针（interface 非 nil 但底层指针为 nil）
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return NewNilValueError("%s", invalidMsg)
	}
	return nil
}

// SafeConvertPBToModel 安全的 PB -> Model 转换
// 自动处理 nil 与 typed-nil 指针，避免 panic
// typed-nil 检查已覆盖 safe Access 的 IsValid 语义，不再重复构造反射包装
func (sc *SafeConverter) SafeConvertPBToModel(pb interface{}, modelPtr interface{}) error {
	if err := checkNotNil(pb, "PB消息不能为空", "PB消息无效"); err != nil {
		return err
	}
	return sc.BidiConverter.ConvertPBToModel(pb, modelPtr)
}

// SafeConvertModelToPB 安全的 Model -> PB 转换
// 自动处理 nil 与 typed-nil 指针，避免 panic
func (sc *SafeConverter) SafeConvertModelToPB(model interface{}, pbPtr interface{}) error {
	if err := checkNotNil(model, "Model不能为空", "Model无效"); err != nil {
		return err
	}
	return sc.BidiConverter.ConvertModelToPB(model, pbPtr)
}

// 原先这里还有 SafeFieldAccess / SafeGetField / SafeGetNestedField 三个函数，
// 返回类型是 *safe.Access，函数体就是 safe.New(obj).Field(name)
// 一行转发：本库没有增加任何语义，只是把上游的名字换了一遍，因此移除
// 需要链式安全访问的调用方直接用上游即可，而 safe 包本身零第三方依赖，
// 直接引它不会给构建图添任何东西：
//
//	import "github.com/HitGameAI/apex-toolbox/pkg/safe"
//	safe.New(obj).Field("Value")        // 原 SafeGetField(obj, "Value")
//	safe.New(obj).At("User.Profile.Name") // 原 SafeGetNestedField
//
// [EN] Three one-line passthroughs returning *safe.Access used to live here.
// They added no semantics of their own beyond renaming upstream's, so they are
// gone; call upstream directly as shown above. pkg/safe itself has no
// third-party dependencies, so importing it adds nothing to the build graph
