/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-02-28 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-05-13 13:15:15
 * @FilePath: \go-toolbox\pkg\serializer\json.go
 * @Description: JSON解析和转换工具
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package serializer

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"

	validator "github.com/HitGameAI/apex-toolbox/pkg/argus"
	"github.com/HitGameAI/apex-toolbox/pkg/protox"
	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
	"github.com/HitGameAI/apex-toolbox/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const DefaultJSONText = "{}"

type jsonFieldMeta struct {
	name       string
	quotedName []byte
	omitEmpty  bool
	needsProto bool
	skip       bool
}

var (
	// needsProtoJSONCache 缓存类型是否包含 protobuf 消息，避免重复递归反射
	needsProtoJSONCache sync.Map
	// jsonFieldsCache 缓存结构体字段 JSON 元信息，减少 tag 解析和字段名转义开销
	jsonFieldsCache sync.Map
)

// JSONMarshal 使用标准 JSON 或 proto-aware JSON 序列化值
func JSONMarshal[T any](value T) ([]byte, error) {
	v := reflect.ValueOf(&value).Elem()
	if !needsProtoJSON(v.Type()) {
		return json.Marshal(value)
	}
	return marshalJSONReflect(v)
}

// JSONUnmarshal 使用标准 JSON 或 proto-aware JSON 反序列化到目标值
func JSONUnmarshal[T any](data []byte, target *T) error {
	if target == nil {
		return NewJSONNilTargetError()
	}

	v := reflect.ValueOf(target).Elem()
	if !needsProtoJSON(v.Type()) {
		return json.Unmarshal(data, target)
	}
	return scanJSONReflect(data, v)
}

// NormalizeJSONText 在 JSON 字符串前添加默认值，确保 JSON 格式正确
func NormalizeJSONText(value string, defaultValue ...string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return NormalizeJSONDefault(defaultValue...)
}

// NormalizeJSONDefault 返回默认 JSON 字符串，确保 JSON 格式正确
func NormalizeJSONDefault(defaultValue ...string) string {
	if len(defaultValue) == 0 {
		return DefaultJSONText
	}
	value := strings.TrimSpace(defaultValue[0])
	if value == "" || validator.ValidateJSON([]byte(value)) != nil {
		return DefaultJSONText
	}
	return value
}

// scanJSONReflect 根据反射类型递归扫描 JSON，遇到 protobuf 消息时使用 protojson
func scanJSONReflect(data []byte, v reflect.Value) error {
	if validator.IsJSONNull(data) {
		return scanJSONNull(v)
	}

	if msg, ok := asProtoMessage(v, true); ok {
		return unmarshalProtoJSON(data, msg)
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return scanJSONReflect(data, v.Elem())
	case reflect.Struct:
		return scanJSONStruct(data, v)
	case reflect.Slice:
		return scanJSONSlice(data, v)
	case reflect.Array:
		return scanJSONArray(data, v)
	case reflect.Map:
		return scanJSONMap(data, v)
	default:
		return json.Unmarshal(data, v.Addr().Interface())
	}
}

// marshalJSONReflect 根据反射类型递归编码 JSON，遇到 protobuf 消息时使用 protojson
func marshalJSONReflect(v reflect.Value) ([]byte, error) {
	if !v.IsValid() {
		return []byte("null"), nil
	}

	if msg, ok := asProtoMessage(v, false); ok {
		if msg == nil {
			return []byte("null"), nil
		}
		return protojson.Marshal(msg)
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return []byte("null"), nil
		}
		return marshalJSONReflect(v.Elem())
	case reflect.Struct:
		if !needsProtoJSON(v.Type()) {
			return json.Marshal(v.Interface())
		}
		return marshalJSONStruct(v)
	case reflect.Slice:
		if v.IsNil() {
			return []byte("null"), nil
		}
		fallthrough
	case reflect.Array:
		if !needsProtoJSON(v.Type().Elem()) {
			return json.Marshal(v.Interface())
		}
		return marshalJSONSlice(v)
	case reflect.Map:
		if v.IsNil() {
			return []byte("null"), nil
		}
		if !needsProtoJSON(v.Type().Elem()) {
			return json.Marshal(v.Interface())
		}
		return marshalJSONMap(v)
	default:
		return json.Marshal(v.Interface())
	}
}

// scanJSONStruct 优先使用快速字段扫描，失败时回退到标准 map 解码路径
func scanJSONStruct(data []byte, v reflect.Value) error {
	err := scanJSONStructFast(data, v)
	if err == nil {
		return nil
	}
	if isJSONStructScanError(err) {
		return err
	}
	return scanJSONStructMap(data, v)
}

// scanJSONStructFast 通过字节扫描定位对象字段，避免构造中间 map
func scanJSONStructFast(data []byte, v reflect.Value) error {
	fields := jsonFieldsInfo(v.Type())
	i, done, err := scanJSONObjectStart(data)
	if err != nil || done {
		return err
	}

	for i < len(data) {
		next, done, err := scanJSONStructEntry(data, i, v, fields)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		i = next
	}
	return NewJSONUnexpectedEndObjectError()
}

func scanJSONObjectStart(data []byte) (next int, done bool, err error) {
	i := validator.SkipJSONSpaces(data, 0)
	if i >= len(data) || data[i] != '{' {
		return 0, false, NewJSONExpectedObjectError()
	}
	i = validator.SkipJSONSpaces(data, i+1)
	return i, i < len(data) && data[i] == '}', nil
}

func scanJSONStructEntry(data []byte, i int, v reflect.Value, fields []jsonFieldMeta) (next int, done bool, err error) {
	keyStart, keyEnd, err := scanJSONObjectKey(data, i)
	if err != nil {
		return 0, false, err
	}
	valueStart, valueEnd, err := scanJSONObjectValue(data, keyEnd)
	if err != nil {
		return 0, false, err
	}
	if err := scanJSONStructField(data, valueStart, valueEnd, v, fields, data[keyStart:keyEnd]); err != nil {
		return 0, false, err
	}
	return scanJSONObjectNext(data, valueEnd)
}

func scanJSONObjectKey(data []byte, i int) (start int, end int, err error) {
	start = validator.SkipJSONSpaces(data, i)
	end, err = validator.ScanJSONString(data, start)
	return start, end, err
}

func scanJSONObjectValue(data []byte, keyEnd int) (start int, end int, err error) {
	i := validator.SkipJSONSpaces(data, keyEnd)
	if i >= len(data) || data[i] != ':' {
		return 0, 0, NewJSONExpectedObjectKeySeparatorError()
	}
	start = validator.SkipJSONSpaces(data, i+1)
	end, err = validator.ScanJSONValueEnd(data, start)
	return start, end, err
}

func scanJSONStructField(data []byte, valueStart int, valueEnd int, v reflect.Value, fields []jsonFieldMeta, quotedKey []byte) error {
	fieldIndex := matchJSONField(fields, quotedKey)
	if fieldIndex < 0 {
		return validateUnknownJSONValue(data[valueStart:valueEnd])
	}
	return scanJSONKnownStructField(data[valueStart:valueEnd], v.Field(fieldIndex), fields[fieldIndex])
}

func validateUnknownJSONValue(data []byte) error {
	if json.Valid(data) {
		return nil
	}
	return NewJSONInvalidUnknownFieldValueError()
}

func scanJSONKnownStructField(data []byte, field reflect.Value, meta jsonFieldMeta) error {
	if meta.skip || !field.CanSet() {
		return nil
	}
	if err := scanJSONFieldValue(data, field, meta); err != nil {
		return NewJSONFieldError(meta.name, err)
	}
	return nil
}

func scanJSONObjectNext(data []byte, valueEnd int) (next int, done bool, err error) {
	i := validator.SkipJSONSpaces(data, valueEnd)
	if i >= len(data) {
		return 0, false, NewJSONUnexpectedEndObjectError()
	}
	if data[i] == '}' {
		return 0, true, nil
	}
	if data[i] != ',' {
		return 0, false, NewJSONExpectedObjectNextError()
	}
	return i + 1, false, nil
}

func scanJSONStructMap(data []byte, v reflect.Value) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	t := v.Type()
	fields := jsonFieldsInfo(t)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		meta := fields[i]
		if meta.skip || !field.CanSet() {
			continue
		}

		rawValue, ok := raw[meta.name]
		if !ok {
			continue
		}

		if err := scanJSONFieldValue(rawValue, field, meta); err != nil {
			return NewJSONFieldError(meta.name, err)
		}
	}
	return nil
}

// matchJSONField 使用已转义字段名优先匹配，必要时再反转义比较
func matchJSONField(fields []jsonFieldMeta, quotedKey []byte) int {
	for i, meta := range fields {
		if !meta.skip && bytes.Equal(quotedKey, meta.quotedName) {
			return i
		}
	}

	key, err := strconv.Unquote(string(quotedKey))
	if err != nil {
		return -1
	}
	for i, meta := range fields {
		if !meta.skip && key == meta.name {
			return i
		}
	}
	return -1
}

func marshalJSONStruct(v reflect.Value) ([]byte, error) {
	t := v.Type()
	fields := jsonFieldsInfo(t)
	buf := make([]byte, 0, v.NumField()*16)
	buf = append(buf, '{')
	wroteField := false

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		meta := fields[i]
		if meta.skip || !field.CanInterface() {
			continue
		}
		if meta.omitEmpty && validator.IsEmptyValue(field) {
			continue
		}

		data, err := marshalJSONReflect(field)
		if err != nil {
			return nil, NewJSONFieldError(meta.name, err)
		}
		if wroteField {
			buf = append(buf, ',')
		}
		buf = append(buf, meta.quotedName...)
		buf = append(buf, ':')
		buf = append(buf, data...)
		wroteField = true
	}
	buf = append(buf, '}')
	return buf, nil
}

func scanJSONSlice(data []byte, v reflect.Value) error {
	i, done, err := scanJSONArrayStart(data)
	if err != nil {
		return err
	}

	elemType := v.Type().Elem()
	result := reflect.MakeSlice(v.Type(), 0, 0)
	itemIndex := 0
	for !done {
		valueStart, valueEnd, err := scanJSONArrayValue(data, i)
		if err != nil {
			return err
		}
		elem := reflect.New(elemType).Elem()
		if !validator.IsJSONNull(data[valueStart:valueEnd]) {
			if err := scanJSONValue(data[valueStart:valueEnd], elem, elemType); err != nil {
				return NewJSONItemError(itemIndex, err)
			}
		}
		result = reflect.Append(result, elem)
		i, done, err = scanJSONArrayNext(data, valueEnd)
		if err != nil {
			return err
		}
		itemIndex++
	}
	v.Set(result)
	return nil
}

func scanJSONArray(data []byte, v reflect.Value) error {
	i, done, err := scanJSONArrayStart(data)
	if err != nil {
		return err
	}

	elemType := v.Type().Elem()
	itemIndex := 0
	for !done {
		if itemIndex >= v.Len() {
			return NewJSONArrayTooLongError(itemIndex+1, v.Len())
		}
		valueStart, valueEnd, err := scanJSONArrayValue(data, i)
		if err != nil {
			return err
		}
		if !validator.IsJSONNull(data[valueStart:valueEnd]) {
			if err := scanJSONValue(data[valueStart:valueEnd], v.Index(itemIndex), elemType); err != nil {
				return NewJSONItemError(itemIndex, err)
			}
		}
		i, done, err = scanJSONArrayNext(data, valueEnd)
		if err != nil {
			return err
		}
		itemIndex++
	}
	return nil
}

func scanJSONArrayStart(data []byte) (next int, done bool, err error) {
	i := validator.SkipJSONSpaces(data, 0)
	if i >= len(data) || data[i] != '[' {
		return 0, false, NewJSONExpectedArrayError()
	}
	i = validator.SkipJSONSpaces(data, i+1)
	return i, i < len(data) && data[i] == ']', nil
}

func scanJSONArrayValue(data []byte, i int) (start int, end int, err error) {
	start = validator.SkipJSONSpaces(data, i)
	end, err = validator.ScanJSONValueEnd(data, start)
	return start, end, err
}

func scanJSONArrayNext(data []byte, valueEnd int) (next int, done bool, err error) {
	i := validator.SkipJSONSpaces(data, valueEnd)
	if i >= len(data) {
		return 0, false, NewJSONUnexpectedEndObjectError()
	}
	if data[i] == ']' {
		return 0, true, nil
	}
	if data[i] != ',' {
		return 0, false, NewJSONExpectedArrayNextError()
	}
	return i + 1, false, nil
}

func marshalJSONSlice(v reflect.Value) ([]byte, error) {
	buf := make([]byte, 0, v.Len()*8)
	buf = append(buf, '[')
	for i := 0; i < v.Len(); i++ {
		data, err := marshalJSONReflect(v.Index(i))
		if err != nil {
			return nil, NewJSONItemError(i, err)
		}
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, data...)
	}
	buf = append(buf, ']')
	return buf, nil
}

func scanJSONMap(data []byte, v reflect.Value) error {
	if v.Type().Key().Kind() != reflect.String {
		return json.Unmarshal(data, v.Addr().Interface())
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	elemType := v.Type().Elem()
	result := reflect.MakeMapWithSize(v.Type(), len(rawMap))
	for key, raw := range rawMap {
		elem := reflect.New(elemType).Elem()
		if !validator.IsJSONNull(raw) {
			if err := scanJSONValue(raw, elem, elemType); err != nil {
				return NewJSONKeyError(key, err)
			}
		}
		result.SetMapIndex(reflect.ValueOf(key).Convert(v.Type().Key()), elem)
	}
	v.Set(result)
	return nil
}

func marshalJSONMap(v reflect.Value) ([]byte, error) {
	if v.Type().Key().Kind() != reflect.String {
		return nil, NewJSONMapKeyUnsupportedError(v.Type().Key().String())
	}

	buf := make([]byte, 0, v.Len()*16)
	buf = append(buf, '{')
	wroteItem := false
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		data, err := marshalJSONReflect(iter.Value())
		if err != nil {
			return nil, NewJSONKeyError(key, err)
		}
		if wroteItem {
			buf = append(buf, ',')
		}
		buf = append(buf, stringx.QuoteJSONBytes(key)...)
		buf = append(buf, ':')
		buf = append(buf, data...)
		wroteItem = true
	}
	buf = append(buf, '}')
	return buf, nil
}

func scanJSONValue(data []byte, v reflect.Value, t reflect.Type) error {
	if needsProtoJSON(t) {
		return scanJSONReflect(data, v)
	}
	return json.Unmarshal(data, v.Addr().Interface())
}

func scanJSONFieldValue(data []byte, field reflect.Value, meta jsonFieldMeta) error {
	if meta.needsProto {
		return scanJSONReflect(data, field)
	}
	return json.Unmarshal(data, field.Addr().Interface())
}

func scanJSONNull(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
		v.Set(reflect.Zero(v.Type()))
	default:
		if v.CanAddr() {
			return json.Unmarshal([]byte("null"), v.Addr().Interface())
		}
	}
	return nil
}

func unmarshalProtoJSON(data []byte, msg proto.Message) error {
	if err := protojson.Unmarshal(data, msg); err != nil {
		var wrapped map[string]json.RawMessage
		if json.Unmarshal(data, &wrapped) == nil {
			if value, ok := wrapped["Data"]; ok && len(value) > 0 && !validator.IsJSONNull(value) {
				return protojson.Unmarshal(value, msg)
			}
		}
		return err
	}
	return nil
}

func asProtoMessage(v reflect.Value, createIfNil bool) (proto.Message, bool) {
	if !v.IsValid() {
		return nil, false
	}

	switch v.Kind() {
	case reflect.Ptr:
		if protox.IsProtoMessageType(v.Type()) {
			if v.IsNil() {
				if !createIfNil {
					return nil, true
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			return v.Interface().(proto.Message), true
		}
	case reflect.Struct:
		if protox.IsProtoMessageType(reflect.PointerTo(v.Type())) && v.CanAddr() {
			return v.Addr().Interface().(proto.Message), true
		}
	case reflect.Interface:
		if !v.IsNil() {
			return asProtoMessage(v.Elem(), createIfNil)
		}
	}
	return nil, false
}

func needsProtoJSON(t reflect.Type) bool {
	if t == nil {
		return false
	}
	if cached, ok := needsProtoJSONCache.Load(t); ok {
		return cached.(bool)
	}
	needs := needsProtoJSONSeen(t, make(map[reflect.Type]bool))
	needsProtoJSONCache.Store(t, needs)
	return needs
}

func needsProtoJSONSeen(t reflect.Type, seen map[reflect.Type]bool) bool {
	if !markProtoJSONTypeSeen(t, seen) {
		return false
	}

	switch t.Kind() {
	case reflect.Ptr:
		return needsProtoJSONPtr(t, seen)
	case reflect.Struct:
		return needsProtoJSONStruct(t, seen)
	case reflect.Array, reflect.Slice, reflect.Map:
		return needsProtoJSONSeen(t.Elem(), seen)
	default:
		return false
	}
}

func markProtoJSONTypeSeen(t reflect.Type, seen map[reflect.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	return true
}

func needsProtoJSONPtr(t reflect.Type, seen map[reflect.Type]bool) bool {
	return protox.IsProtoMessageType(t) || needsProtoJSONSeen(t.Elem(), seen)
}

func needsProtoJSONStruct(t reflect.Type, seen map[reflect.Type]bool) bool {
	if protox.IsProtoMessageType(reflect.PointerTo(t)) {
		return true
	}
	return needsProtoJSONFields(t, seen)
}

func needsProtoJSONFields(t reflect.Type, seen map[reflect.Type]bool) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if types.IsExportedField(field) && needsProtoJSONSeen(field.Type, seen) {
			return true
		}
	}
	return false
}

func jsonFieldsInfo(t reflect.Type) []jsonFieldMeta {
	if cached, ok := jsonFieldsCache.Load(t); ok {
		return cached.([]jsonFieldMeta)
	}

	fields := make([]jsonFieldMeta, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fields[i] = jsonFieldInfo(t.Field(i))
	}
	jsonFieldsCache.Store(t, fields)
	return fields
}

func jsonFieldInfo(field reflect.StructField) jsonFieldMeta {
	if !types.IsExportedField(field) {
		return jsonFieldMeta{skip: true}
	}

	tag := field.Tag.Get("json")
	if tag == "-" {
		return jsonFieldMeta{skip: true}
	}

	meta := jsonFieldMeta{name: field.Name}
	if tag != "" {
		meta.name = types.JSONFieldName(field)
		meta.omitEmpty = types.HasJSONTagOption(field, "omitempty", "omitzero")
	}
	meta.needsProto = needsProtoJSON(field.Type)
	meta.quotedName = stringx.QuoteJSONBytes(meta.name)
	return meta
}
