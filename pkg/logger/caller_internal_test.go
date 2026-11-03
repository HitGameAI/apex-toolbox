/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-26 02:10:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-09-30 15:28:36
 * @FilePath: \go-toolbox\pkg\logger\caller_internal_test.go
 * @Description: 内部帧判定的路径无关性回归
 *
 * 覆盖缺陷：isInternalCaller 原按文件路径子串匹配 "/go-toolbox/pkg/logger/" 识别本包帧，
 * 只要检出目录不叫 go-toolbox/pkg/logger 该判定即失效（git worktree、vendor 副本、
 * 改名的 clone、workspace 名不同的 CI、改名的 fork），届时库自身的帧被当作
 * 外部调用者，每条日志的 caller 静默指向库内部而非用户代码。
 *
 * 本文件的用例全部构造「文件路径不含 go-toolbox/pkg/logger」的输入，故在任意检出路径下
 * 都能复现该缺陷——不依赖测试运行时的实际目录名。
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"strings"
	"testing"
)

// TestSelfModPrefix_Derived 根包路径必须成功推导
//
// 推导失败（空值）会让本库帧判定整体退化为「全部不是内部帧」，
// caller 将指向 logger 自身的最内层函数。
func TestSelfModPrefix_Derived(t *testing.T) {
	if selfModPrefix == "" {
		t.Fatal("selfModPrefix 推导为空：本库帧判定会整体失效")
	}
	if strings.HasSuffix(selfModPrefix, ".") {
		t.Fatalf("selfModPrefix 不得含末尾 '.'：边界字符由 isSelfFrame 判定，实得 %q", selfModPrefix)
	}
	if !strings.Contains(selfModPrefix, "go-toolbox/pkg/logger") {
		t.Fatalf("selfModPrefix 应含模块名，实得 %q", selfModPrefix)
	}
}

// TestIsSelfFrame_Subpackages 子包帧必须与根包帧一样被识别为本库
//
// 这是包结构拆分的回归：判定若退化成「前缀 + 末尾 '.'」，
// <mod>/sink.xxx 这类子包帧不再匹配，子包里的函数会被当作外部调用者，
// caller 静默指向库内部（如 sink/console.go:214）而非用户代码。
func TestIsSelfFrame_Subpackages(t *testing.T) {
	for _, fn := range []string{
		selfModPrefix + ".(*Logger).Info",
		selfModPrefix + ".appendJSONString",
		selfModPrefix + "/sink.(*consoleWriter).Write",
		selfModPrefix + "/sink.Flush",
		selfModPrefix + "/level.Level.String",
		selfModPrefix + "/console.(*Group).Info",
		selfModPrefix + "/internal/redact.AppendJSONValue",
		selfModPrefix + "/internal/encode.AppendTextSafe",
	} {
		if !isSelfFrame(fn) {
			t.Errorf("本库帧未被识别，caller 会指向库内部: %q", fn)
		}
	}
}

// TestIsSelfFrame_PrefixNeighbours 模块路径以本库为前缀的「邻居模块」不得误判
//
// 这是边界检查的回归，而且不是假想输入：go-logger-benchmark 是本库的真实
// 下游，它的 runtime 名称以 "github.com/HitGameAI/apex-toolbox/pkg/logger" 开头。
// 判定若退化成裸 HasPrefix，基准自身的帧会被当作内部帧跳过。
func TestIsSelfFrame_PrefixNeighbours(t *testing.T) {
	for _, fn := range []string{
		selfModPrefix + "-benchmark.BenchmarkGoLogger_KV10",
		selfModPrefix + "-extra.Helper",
		selfModPrefix + "_test.TestExternal",
		selfModPrefix, // 恰好等长：索引越界的边界
	} {
		if isSelfFrame(fn) {
			t.Errorf("邻居模块帧被误判为本库，其调用点会被跳过: %q", fn)
		}
	}
}

// TestIsInternalCaller_PathIndependent 本包帧在路径不含 go-toolbox/pkg/logger 时仍须被识别
//
// 这是原缺陷的直接回归：旧实现在此输入下返回 false（判为外部调用者）。
func TestIsInternalCaller_PathIndependent(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		function string
	}{
		{
			name:     "git worktree 路径",
			file:     "/private/tmp/wt-abc123/logger.go",
			function: selfModPrefix + "." + "(*Logger).writeJSONEntry",
		},
		{
			name:     "vendor 副本路径",
			file:     "/app/vendor/github.com/kamalyesogging/logger.go",
			function: selfModPrefix + "." + "appendJSONString",
		},
		{
			name:     "改名的检出目录",
			file:     "/home/ci/workspace/build-7/logger.go",
			function: selfModPrefix + "." + "(*Logger).appendJSONHeader",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if strings.Contains(c.file, "go-toolbox/pkg/logger") {
				t.Fatalf("用例自身失效：file 不应含 go-toolbox/pkg/logger，实得 %q", c.file)
			}
			if !isInternalCaller(c.file, c.function) {
				t.Errorf("本包帧未被识别为内部（按路径判定的回归）\n  file=%q\n  func=%q", c.file, c.function)
			}
		})
	}
}

// TestIsInternalCaller_UserCodeNotSkipped 用户代码不得被判为内部帧
//
// 反向保护：若判定过宽，caller 会继续向外回溯并越过真实调用点。
func TestIsInternalCaller_UserCodeNotSkipped(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		function string
	}{
		{"普通用户代码", "/app/svc/handler.go", "example.com/app/svc.(*Handler).Serve"},
		{"路径恰好含 go-toolbox/pkg/logger 的用户代码", "/src/go-toolbox/pkg/logger-demo/main.go", "example.com/demo.main"},
		{"外部测试包", "/app/x/foo_x.go", "github.com/HitGameAI/apex-toolbox/pkg/logger_test.TestExternal"},
		{"同名前缀的其他模块", "/app/y/y.go", "github.com/kamalyeso-logger-extra.Helper"},
		{"下游基准仓库", "/app/b/bench_kv_test.go", "github.com/kamalyeso-logger-benchmark.benchGoLogger"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if isInternalCaller(c.file, c.function) {
				t.Errorf("用户帧被误判为内部，caller 会越过真实调用点\n  file=%q\n  func=%q", c.file, c.function)
			}
		})
	}
}

// TestIsInternalCaller_TestFileExempt 本包 _test.go 必须豁免
//
// 本包内测试函数的 runtime 名称同样带本库根包前缀，若无 _test.go 豁免
// 会被判为内部帧，导致测试里的日志 caller 指向 logger 内部。
// 该豁免必须先于包路径判定。
func TestIsInternalCaller_TestFileExempt(t *testing.T) {
	if isInternalCaller("/anywhere/logger_test.go", selfModPrefix+"."+"TestShowCaller") {
		t.Error("本包 _test.go 被判为内部帧：豁免未生效或顺序错误")
	}
}

// TestIsInternalCaller_Stdlib 标准库帧仍须跳过
func TestIsInternalCaller_Stdlib(t *testing.T) {
	cases := []struct{ file, function string }{
		{"/usr/local/go/src/runtime/panic.go", "runtime.gopanic"},
		{"/usr/local/go/src/reflect/value.go", "reflect.Value.call"},
		{"/usr/local/go/src/testing/testing.go", "testing.tRunner"},
	}
	for _, c := range cases {
		if !isInternalCaller(c.file, c.function) {
			t.Errorf("标准库帧未被跳过: %q", c.file)
		}
	}
}
