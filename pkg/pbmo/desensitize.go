/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-09-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2023-09-26 00:00:00
 * @FilePath: \go-toolbox\pkg\pbmo\desensitize.go
 * @Description: 把脱敏编织进 unsafe 拷贝计划，转换与脱敏合为单遍
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

package pbmo

import (
	"reflect"
	"unsafe"

	"github.com/HitGameAI/apex-toolbox/pkg/desensitize"
)

// resolveFieldMask 在计划构建期解析字段的脱敏函数，无需脱敏时返回 nil
//
// 标签取自「源字段优先、否则目标字段」：protoc 生成的 PB 结构体不带
// desensitize tag，所以实际生效的总是 Model 一侧的声明，Model→PB 与 PB→Model
// 两个方向都由同一处标签说话，不必再按方向分叉
//
// 仅处理 reflect.String 形态的目标字段，其余形态返回 nil：
// 指针形态（*string）在 fieldSameType 下是指针直拷，就地改写会污染源对象共享的
// 那块内存；切片、映射与结构体形态则由各自的嵌套拷贝计划负责搬运，
// 而嵌套计划按「源类型+目标类型」全进程共享，承载不了 per-converter 的开关
// 这些形态需要脱敏时在输出层调 desensitize.Desensitization
//
// [EN] Resolves a field's masking function once, while the plan is built, and
// returns nil when the field needs none. The tag is read from the source field
// first and the destination field otherwise: protoc-generated structs carry no
// desensitize tag, so the Model side always governs and both directions are
// driven by the same declaration without branching on direction. Only
// string-kinded destination fields are handled: a *string under fieldSameType is
// a pointer copy, so masking through it would corrupt memory still shared with
// the source, while slice, map and struct shapes are moved by their own nested
// copy plans, which are shared process-wide by type pair and so cannot carry a
// per-converter switch. Mask those at the output layer with
// desensitize.Desensitization
func resolveFieldMask(srcField, dstField reflect.StructField, enabled bool) desensitize.MaskFunc {
	if !enabled {
		return nil
	}
	tag := srcField.Tag.Get(desensitize.TagName)
	if tag == "" {
		tag = dstField.Tag.Get(desensitize.TagName)
	}
	if tag == "" || tag == "-" {
		return nil
	}
	if dstField.Type.Kind() != reflect.String {
		return nil
	}
	mask, err := desensitize.ResolveRule(tag)
	if err != nil {
		// 标签名未注册：保持原样而不是静默写空值，注册后下次建计划即生效
		return nil
	}
	return mask
}

// wrapMaskCopy 在字段拷贝之后就地脱敏目标字段
//
// 只读写 dstBase 上的目标偏移，srcBase 全程只读，因此调用方传入的源对象不会被
// 改坏——这正是「转换中途脱敏会破坏管道数据」这一顾虑的解法
// 字符串是不可变值，写目标字段只是换一个 header，不触碰源字符串的底层字节
//
// [EN] Masks the destination field in place right after the field is copied.
// Only the destination offset is written and srcBase stays read-only, so the
// caller's source object is never corrupted — which is what makes masking safe
// mid-conversion. Strings are immutable, so writing the destination field just
// swaps a header and never touches the source's backing bytes
func wrapMaskCopy(inner fieldCopyFunc, dstOffset uintptr, mask desensitize.MaskFunc) fieldCopyFunc {
	if inner == nil {
		return nil
	}
	// 参数顺序随 fieldCopyFunc：源在前、目标在后
	return func(srcBase, dstBase unsafe.Pointer) {
		inner(srcBase, dstBase)
		p := (*string)(addPtr(dstBase, dstOffset))
		*p = mask(*p)
	}
}

// maybeWrapMask 需要脱敏时包装拷贝闭包，否则原样返回
//
// 把「是否脱敏」的判断收在计划构建期：无 tag 的字段拿回的是原闭包本身，
// 既不多一层间接调用，也不多一次分支——这是「开关开着也零代价」的落点
//
// [EN] Wraps the copy closure only when the field needs masking and returns it
// untouched otherwise, keeping the decision at plan-build time: an untagged
// field gets its original closure back, with no extra indirection and no extra
// branch. This is what keeps the feature free when the switch is on but no
// field carries a tag
func maybeWrapMask(inner fieldCopyFunc, dstOffset uintptr, mask desensitize.MaskFunc) fieldCopyFunc {
	if mask == nil {
		return inner
	}
	return wrapMaskCopy(inner, dstOffset, mask)
}

// applyMaskToField 在 reflect 慢路径上脱敏目标字段
//
// 慢路径的字段没有 copyFunc，拿不到 wrapMaskCopy 的包装，必须单独收口：
// 否则「哪些字段能生成 unsafe 闭包」这一实现细节会决定「哪些字段会被脱敏」，
// 同一个 tag 在不同字段组合下擦与不擦——这正是最容易被忽略的漏擦
//
// [EN] Slow-path fields have no copyFunc and therefore never pass through
// wrapMaskCopy, so masking has to be applied here as well. Without it, whether
// a field can be given an unsafe closure would silently decide whether its tag
// is honoured, making the same tag mask in one field layout and leak in another
func applyMaskToField(dstField reflect.Value, mask desensitize.MaskFunc) {
	if mask == nil || dstField.Kind() != reflect.String || !dstField.CanSet() {
		return
	}
	dstField.SetString(mask(dstField.String()))
}
