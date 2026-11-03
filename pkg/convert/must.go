/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-08-03 21:32:26
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-05-27 18:09:55
 * @FilePath: \go-toolbox\pkg\convert\must.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package convert

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
)

// MustString 强制转为字符串
// 参数:
//   - v: 要转换的值，支持 string、[]byte、error、bool、数字、time.Time 等
//   - timeLayout: 可选的时间格式化布局，默认使用 time.RFC3339
//
// 返回值: 转换后的字符串
func MustString[T any](v T, timeLayout ...string) string {
	switch v := any(v).(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case nil:
		return ""
	case bool:
		return strconv.FormatBool(v)
	default:
		return convertToString(v, timeLayout...)
	}
}

// convertToString 将其他类型转换为字符串
func convertToString[T any](v T, timeLayout ...string) string {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Struct:
		return convertStructToString(val, timeLayout...)
	case reflect.Ptr:
		return convertPtrToString(val, timeLayout...)
	}
	return convertFallback(v)
}

// convertStructToString 将结构体转换为字符串
func convertStructToString(val reflect.Value, timeLayout ...string) string {
	if val.Type() == reflect.TypeOf(time.Time{}) {
		return formatTime(val.Interface().(time.Time), timeLayout...)
	}
	return convertFallback(val.Interface())
}

// convertPtrToString 将指针类型转换为字符串
func convertPtrToString(val reflect.Value, timeLayout ...string) string {
	if val.IsNil() {
		return ""
	}

	// 处理 *timestamppb.Timestamp：用鸭子类型断言，避免为一个分支把整个
	// protobuf 引入 convert（convert 被 mathx/random/contextx 等基础包依赖）
	// [EN] Duck-typed assertion instead of importing protobuf for a single branch;
	// convert is depended on by many foundational packages
	if ts, ok := val.Interface().(interface{ AsTime() time.Time }); ok {
		return formatTime(ts.AsTime(), timeLayout...)
	}

	// 递归处理其他指针类型，使用 MustString 而不是 convertToString
	return MustString(val.Elem().Interface(), timeLayout...)
}

// formatTime 格式化时间
func formatTime(t time.Time, timeLayout ...string) string {
	if len(timeLayout) > 0 {
		return t.Format(timeLayout[0])
	}
	return t.Format(time.RFC3339)
}

// convertFallback 默认转换方式
func convertFallback(v any) string {
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%v", v)
}

// RoundMode 是一个枚举类型，用于指定取整的方式
type RoundMode int

const (
	RoundNone    RoundMode = iota // 不进行四舍五入，保持原值
	RoundNearest                  // 四舍五入到最接近的整数
	RoundDown                     // 向下取整
	RoundUp                       // 向上取整
)

var defaultRoundMode = RoundNone

// MustIntT 将任意类型转换为数字类型 T
// 参数:
//   - value: 要转换的值，支持 int/uint 系列、float 系列、string
//   - mode: 取整模式，nil 时默认为 RoundDown
//
// 返回值: 转换后的数字和可能的错误
func MustIntT[T types.Numerical](value any, mode *RoundMode) (T, error) {
	if mode == nil {
		mode = &defaultRoundMode
	}

	var zero T
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return T(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return T(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return applyRoundMode[T](v.Float(), *mode), nil
	case reflect.String:
		return parseStringToInt[T](v.String(), *mode)
	default:
		return zero, fmt.Errorf("unsupported conversion: %v (type %T)", value, value)
	}
}

// applyRoundMode 应用取整模式（内部辅助函数）
func applyRoundMode[T types.Numerical](value float64, mode RoundMode) T {
	switch mode {
	case RoundUp:
		return T(math.Ceil(value))
	case RoundDown:
		return T(math.Floor(value))
	case RoundNearest:
		return T(math.Round(value))
	default:
		return T(value)
	}
}

// parseStringToInt 将字符串解析为整数（处理跨平台一致性问题）
func parseStringToInt[T types.Numerical](v string, mode RoundMode) (T, error) {
	var zero T
	var floatValue float64
	if err := ParseFloat(v, &floatValue); err != nil {
		// 直接回传不再包一层：ParseFloat 的错误已经带了同样的 "failed to parse %q"
		// 前缀，再包一次会得到
		//   failed to parse "abc": failed to parse "abc": strconv.ParseFloat: ...
		// 前缀重复一遍，且多付一轮格式化与装箱的分配
		//
		// [EN] Returned as is rather than wrapped again: ParseFloat's error
		// already carries the same "failed to parse %q" prefix, so wrapping
		// duplicated the prefix in the message and paid another round of
		// formatting and boxing allocations
		return zero, err
	}
	return Float64ToInt[T](floatValue, mode)
}

// MustFloatT 将值转换为浮点数类型 T
// 参数:
//   - value: 要转换的值，支持 string、float 系列、int 系列
//   - mode: 取整模式
//
// 返回值: 转换后的浮点数和可能的错误
func MustFloatT[T types.Float](value any, mode RoundMode) (T, error) {
	f, err := ToFloat64(value)
	if err != nil {
		return 0, err
	}

	switch mode {
	case RoundNone:
		return T(f), nil
	case RoundNearest:
		return T(math.Round(f)), nil
	case RoundUp:
		return T(math.Ceil(f)), nil
	case RoundDown:
		return T(math.Floor(f)), nil
	default:
		return 0, fmt.Errorf("未知的四舍五入模式")
	}
}

// ToFloat64 将各种类型转换为 float64
func ToFloat64(value any) (float64, error) {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(v.String(), 64)
	default:
		return 0, fmt.Errorf("不支持的输入类型: %T", value)
	}
}

// Float64ToInt 将浮点数转换为整数类型，进行范围检查
func Float64ToInt[T types.Numerical](value float64, mode RoundMode) (T, error) {
	var zero T
	convertedValue := applyRoundMode[float64](value, mode)

	// 无符号类型不允许负值
	switch any(zero).(type) {
	case uint, uint8, uint16, uint32, uint64:
		if convertedValue < 0 {
			return zero, fmt.Errorf("value %f is negative, cannot convert to unsigned type %T", convertedValue, zero)
		}
	}

	// 检查范围（仅对 int64 类型检查，其他类型由 Go 自动处理）
	if convertedValue < float64(math.MinInt64) || convertedValue > float64(math.MaxInt64) {
		return zero, fmt.Errorf("value %f out of range for type %T", convertedValue, zero)
	}

	return T(convertedValue), nil
}

// ParseFloat 将字符串解析为浮点数，进行 NaN 和 Inf 检查
func ParseFloat[T types.Float](v string, value *T) error {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		// %w 而不是 %v：消息文本一模一样，但底层的 *strconv.NumError 能被
		// errors.As 取出，调用方可以区分语法错误与范围溢出
		return fmt.Errorf("failed to parse %q: %w", v, err)
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("invalid float value: %q", v)
	}
	*value = T(f)
	return nil
}

// MustBool 强制转为 bool，支持布尔值、字符串、数字
func MustBool[T any](v T) bool {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.String:
		return validator.IsTrueString(val.String())
	default:
		flag, err := MustIntT[int](v, nil)
		return err == nil && flag != 0
	}
}

// MustJSONIndent 转 json 返回 []byte
func MustJSONIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// MustJSON 转 json 返回 []byte
func MustJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// NumberSliceToStringSlice Number切片转String
func NumberSliceToStringSlice[T types.Numerical](numbers []T) []string {
	if len(numbers) == 0 {
		return []string{}
	}
	result := make([]string, len(numbers))
	for i, n := range numbers {
		v := reflect.ValueOf(n)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[i] = strconv.FormatInt(v.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[i] = strconv.FormatUint(v.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			result[i] = strconv.FormatFloat(v.Float(), 'f', -1, 64)
		default:
			// 兜底，理论上 types.Numerical 约束不会进入此分支
			result[i] = fmt.Sprintf("%v", n)
		}
	}
	return result
}

// StringSliceToNumberSlice 将字符串切片转换为数字切片
//
// 直接走字符串特化的解析，不再绕 MustIntT
//
// MustIntT 的入参是 any，而这里静态上已经知道元素是 string把 string 放进
// any 会分配一次（字符串头装不进一个接口字），于是每个元素都要多付一次分配，
// 再加一次 reflect.ValueOf 与 Kind 分支——泛型函数本来握着类型信息，却在调用
// MustIntT 时丢掉了
// 十个元素实测由 247ns / 11 次分配降到 135ns / 1 次分配，只剩结果切片那一次
//
// [EN] Parses with the string-specialised path instead of going through
// MustIntT, whose parameter is any. The element type is statically known to be
// string here, and putting a string into an any allocates — a string header does
// not fit in one interface word — so every element paid an allocation plus a
// reflect.ValueOf and Kind switch. A generic function holds the type information
// and threw it away at the call. Measured over ten elements, 247ns and 11
// allocations become 135ns and the single allocation for the result slice
func StringSliceToNumberSlice[T types.Numerical](input []string, mode *RoundMode) ([]T, error) {
	if len(input) == 0 {
		return []T{}, nil
	}
	// 取整模式只解一次，不必每个元素都判一遍 nil
	m := defaultRoundMode
	if mode != nil {
		m = *mode
	}
	result := make([]T, len(input))
	for i, s := range input {
		v, err := parseStringToInt[T](s, m)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// StringSliceToFloatSlice 将字符串切片转换为浮点数切片
//
// 同 StringSliceToNumberSlice，绕开 MustFloatT 的 any 入参以免逐元素装箱
func StringSliceToFloatSlice[T types.Float](input []string, mode RoundMode) ([]T, error) {
	if len(input) == 0 {
		return []T{}, nil
	}
	result := make([]T, len(input))
	for i, s := range input {
		v, err := parseStringToFloat[T](s, mode)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// parseStringToFloat 解析字符串为浮点并套用取整模式
//
// 与 MustFloatT 对字符串的处理逐步等价用的是 strconv.ParseFloat 而不是本包的
// ParseFloat：后者会拒绝 NaN 与 Inf，换过去会悄悄缩小可接受的输入集
// 模式分支与未知模式的错误也与 MustFloatT 保持一致
//
// [EN] Equivalent step for step to how MustFloatT handles a string. It uses
// strconv.ParseFloat rather than this package's own ParseFloat, which rejects
// NaN and Inf — swapping them would quietly narrow the accepted input set. The
// mode branches and the unknown-mode error match MustFloatT as well
func parseStringToFloat[T types.Float](s string, mode RoundMode) (T, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	switch mode {
	case RoundNone:
		return T(f), nil
	case RoundNearest:
		return T(math.Round(f)), nil
	case RoundUp:
		return T(math.Ceil(f)), nil
	case RoundDown:
		return T(math.Floor(f)), nil
	default:
		return 0, fmt.Errorf("未知的四舍五入模式")
	}
}

// sliceMapErr 切片映射转换（带错误处理）
func sliceMapErr[T any, R any](slice []T, fn func(T) (R, error)) ([]R, error) {
	if len(slice) == 0 {
		return []R{}, nil
	}
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		r, err := fn(v)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// AnySliceToInterfaceSlice 将任意类型的切片或数组转换为 []any
// 支持所有类型切片: []string, []int, []int32, []int64, []uint, []bool, 自定义类型切片等
// 支持数组类型: [3]int, [5]string 等
// 如果传入的不是切片/数组类型或为空，返回空切片
func AnySliceToInterfaceSlice(slice any) []any {
	if slice == nil {
		return []any{}
	}

	v := reflect.ValueOf(slice)
	// 支持切片和数组
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return []any{}
	}

	length := v.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = v.Index(i).Interface()
	}
	return result
}

// StringSliceToInterfaceSlice 将 []string 转换为 []any
// 为了向后兼容保留，内部调用 AnySliceToInterfaceSlice
func StringSliceToInterfaceSlice(slice []string) []any {
	return AnySliceToInterfaceSlice(slice)
}

// ToNumberSlice 支持输入 string 或 []string，自动拆分和转换
// 参数:
//   - input: 字符串或字符串切片
//   - desolator: 分隔符（仅当 input 为 string 时有效）
//
// 返回值: 数字切片和可能的错误
func ToNumberSlice[T types.Numerical](input any, desolator string) ([]T, error) {
	strSlice, err := normalizeToStringSlice(input, desolator)
	if err != nil {
		return nil, err
	}

	return sliceMapErr(strSlice, func(s string) (T, error) {
		trimmed := strings.TrimSpace(s)
		return MustIntT[T](trimmed, &defaultRoundMode)
	})
}

// MustToNumberSlice 不返回错误的版本，转换失败时 panic
func MustToNumberSlice[T types.Numerical](input any, desolator string) []T {
	nums, err := ToNumberSlice[T](input, desolator)
	if err != nil {
		panic(err)
	}
	return nums
}

// normalizeToStringSlice 将输入标准化为字符串切片
func normalizeToStringSlice(input any, separator string) ([]string, error) {
	switch v := input.(type) {
	case string:
		if v == "" {
			return []string{}, nil
		}
		return strings.Split(v, separator), nil
	case []string:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported input type %T, want string or []string", input)
	}
}

// InterfaceSliceToStringSlice 将 []any 转换为 []string
func InterfaceSliceToStringSlice(slice []any) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = MustString(v)
	}
	return result
}

// InterfaceSliceToIntSlice 将 []any 转换为 []int
func InterfaceSliceToIntSlice(slice []any, mode *RoundMode) []int {
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if num, err := MustIntT[int](v, mode); err == nil {
			result = append(result, num)
		}
	}
	return result
}

// InterfaceMapToStringMap 将 map[any]any 转换为 map[string]any
func InterfaceMapToStringMap(m map[any]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if key, ok := k.(string); ok {
			result[key] = v
		}
	}
	return result
}

// MustConvertTo 通用的泛型类型转换函数
// 自动将 value 转换为目标类型 T
//
// 参数:
//   - value: 要转换的值
//
// 返回值:
//   - 转换后的值和是否成功的标志
//
// 支持的类型:
//   - string: 字符串
//   - bool: 布尔值
//   - int, int8, int16, int32, int64: 有符号整数
//   - uint, uint8, uint16, uint32, uint64: 无符号整数
//   - float32, float64: 浮点数
//   - []byte: 字节切片
//   - map[string]any: 字典类型
//   - []any: 切片类型
//
// 示例:
//
//	str, ok := MustConvertTo[string](123)           // "123", true
//	num, ok := MustConvertTo[int]("567")            // 567, true
//	flag, ok := MustConvertTo[bool]("true")         // true, true
//	f64, ok := MustConvertTo[float64]("3.14")       // 3.14, true
//	bytes, ok := MustConvertTo[[]byte]("hello")     // []byte("hello"), true
//	m, ok := MustConvertTo[map[string]any](data)    // map, true
func MustConvertTo[T types.Convertible](value any) (T, bool) {
	var zero T

	// 如果值已经是目标类型，直接返回
	if v, ok := value.(T); ok {
		return v, true
	}

	// 使用反射判断目标类型的 Kind
	typ := reflect.TypeOf(zero)
	var result any

	switch typ.Kind() {
	case reflect.String:
		result = MustString(value)
	case reflect.Bool:
		result = MustBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = convertToNumerical(typ, value)
		if result == nil {
			return zero, false
		}
	case reflect.Float32, reflect.Float64:
		result = convertToFloating(typ, value)
		if result == nil {
			return zero, false
		}
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 { // []byte
			result = convertToBytes(value)
		} else { // []any
			result = convertToSlice(value)
		}
		if result == nil {
			return zero, false
		}
	case reflect.Map:
		result = convertToMap(value)
		if result == nil {
			return zero, false
		}
	default:
		return zero, false
	}

	if v, ok := result.(T); ok {
		return v, true
	}
	return zero, false
}

// convertToNumerical 转换为数字类型（整数）
func convertToNumerical(typ reflect.Type, value any) any {
	switch typ.Kind() {
	case reflect.Int:
		v, err := MustIntT[int](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Int8:
		v, err := MustIntT[int8](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Int16:
		v, err := MustIntT[int16](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Int32:
		v, err := MustIntT[int32](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Int64:
		v, err := MustIntT[int64](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Uint:
		v, err := MustIntT[uint](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Uint8:
		v, err := MustIntT[uint8](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Uint16:
		v, err := MustIntT[uint16](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Uint32:
		v, err := MustIntT[uint32](value, nil)
		if err != nil {
			return nil
		}
		return v
	case reflect.Uint64:
		v, err := MustIntT[uint64](value, nil)
		if err != nil {
			return nil
		}
		return v
	}
	return nil
}

// convertToFloating 转换为浮点数类型
func convertToFloating(typ reflect.Type, value any) any {
	switch typ.Kind() {
	case reflect.Float32:
		v, err := MustFloatT[float32](value, RoundNone)
		if err != nil {
			return nil
		}
		return v
	case reflect.Float64:
		v, err := MustFloatT[float64](value, RoundNone)
		if err != nil {
			return nil
		}
		return v
	}
	return nil
}

// convertToBytes 转换为字节切片
func convertToBytes(value any) any {
	if s, ok := value.(string); ok {
		return []byte(s)
	}
	if b, ok := value.([]byte); ok {
		return b
	}
	return nil
}

// convertToMap 转换为字典类型
func convertToMap(value any) any {
	if m, ok := value.(map[string]any); ok {
		return m
	}
	if m, ok := value.(map[interface{}]interface{}); ok {
		return InterfaceMapToStringMap(m)
	}
	return nil
}

// convertToSlice 转换为切片类型
func convertToSlice(value any) any {
	if s, ok := value.([]any); ok {
		return s
	}
	result := AnySliceToInterfaceSlice(value)
	if len(result) == 0 && value != nil {
		return nil
	}
	return result
}
