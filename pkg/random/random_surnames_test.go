/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-06-17 13:15:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-06-17 13:55:16
 * @FilePath: \go-toolbox\pkg\random\random_surnames_test.go
 * @Description: 随机姓氏管理器单元测试
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */
package random

import (
	"strings"
	"testing"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/stretchr/testify/assert"
)

func TestNewSurnameManager(t *testing.T) {
	// 使用默认数据创建管理器
	manager := NewSurnameManager()
	if len(manager.Data()) != len(SurnameData) {
		assert.Equal(t, len(SurnameData), len(manager.Data()), "默认数据长度应一致")
	}

	// 追加额外数据，使用非默认的姓氏以避免重复
	extra := []SurnameInfo{
		{Surname: "测试", Initial: "c", CorrectPinyin: "ceshi", AllPinyins: []string{"cèshì"}},
		{Surname: "示例", Initial: "s", CorrectPinyin: "shili", AllPinyins: []string{"shìlì"}},
	}
	manager = NewSurnameManager(extra)

	expectedLen := len(SurnameData) + len(extra)
	assert.Equal(t, expectedLen, len(manager.Data()), "追加数据后长度应正确")

	// 验证追加的数据是否存在
	foundCeshi, foundShili := false, false
	for _, info := range manager.Data() {
		if info.Surname == "测试" {
			foundCeshi = true
		}
		if info.Surname == "示例" {
			foundShili = true
		}
	}
	assert.True(t, foundCeshi, "应包含追加的姓氏 '测试'")
	assert.True(t, foundShili, "应包含追加的姓氏 '示例'")
}

func TestFilterBySurname(t *testing.T) {
	mgr := NewSurnameManager()

	// 正常匹配
	result := mgr.FilterBySurname("赵")
	assert.NotEmpty(t, result.Data(), "按姓氏过滤应返回非空结果")
	for _, info := range result.Data() {
		assert.Equal(t, "赵", info.Surname, "过滤结果姓氏应匹配")
	}

	// 不存在的姓氏
	emptyResult := mgr.FilterBySurname("不存在")
	assert.Empty(t, emptyResult.Data(), "不存在的姓氏应返回空结果")
}

func TestFilterByPinyin(t *testing.T) {
	mgr := NewSurnameManager()

	// 精确匹配 CorrectPinyin
	result := mgr.FilterByPinyin("zhao")
	assert.NotEmpty(t, result.Data(), "按拼音过滤应返回非空结果")
	for _, info := range result.Data() {
		cp := strings.ToLower(info.CorrectPinyin)
		allPinyin := make([]string, len(info.AllPinyins))
		for i, p := range info.AllPinyins {
			allPinyin[i] = strings.ToLower(p)
		}
		assert.True(t,
			cp == "zhao" || mathx.SliceContains(allPinyin, "zhao"),
			"过滤结果拼音应包含目标拼音")
	}

	// 大小写匹配
	result2 := mgr.FilterByPinyin("ZHAO")
	assert.NotEmpty(t, result2.Data(), "拼音过滤应不区分大小写")

	// 不存在的拼音
	emptyResult := mgr.FilterByPinyin("不存在拼音")
	assert.Empty(t, emptyResult.Data(), "不存在的拼音应返回空结果")
}

func TestToJSON(t *testing.T) {
	mgr := NewSurnameManager()

	jsonStr, err := mgr.ToJSON()
	assert.NoError(t, err, "ToJSON 不应返回错误")
	assert.Contains(t, jsonStr, "赵", "JSON字符串应包含姓氏‘赵’")
}

func TestPrint(t *testing.T) {
	mgr := NewSurnameManager()
	// 这里仅测试不panic，输出内容人工验证
	mgr.Print()
}

func TestAppendData(t *testing.T) {
	// 创建一个初始管理器，包含两个姓氏
	manager := NewSurnameManager()
	initialCount := len(manager.Data())

	// 追加新的姓氏：阿巴 和 阿哦
	newSurnames := []SurnameInfo{
		{Surname: "阿巴", Initial: "a", CorrectPinyin: "aba", AllPinyins: []string{"ābā"}},
		{Surname: "阿哦", Initial: "a", CorrectPinyin: "ao", AllPinyins: []string{"āo"}},
	}
	manager.AppendData(newSurnames...)

	// 断言追加后数据总长度增加了
	assert.Equal(t, initialCount+len(newSurnames), len(manager.Data()))

	// 断言追加的数据正确存在
	data := manager.Data()
	foundSun := false
	foundZhou := false
	for _, info := range data {
		if info.Surname == "阿巴" {
			foundSun = true
			assert.Equal(t, "aba", info.CorrectPinyin)
		}
		if info.Surname == "阿哦" {
			foundZhou = true
			assert.Equal(t, "ao", info.CorrectPinyin)
		}
	}
	assert.True(t, foundSun, "追加的姓氏 孙 应该存在")
	assert.True(t, foundZhou, "追加的姓氏 周 应该存在")
}
