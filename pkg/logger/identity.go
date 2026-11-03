/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-01-26 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-01 18:30:05
 * @FilePath: \go-toolbox\pkg\logger\identity.go
 * @Description: 服务身份固定字段（service / instance / env）
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */
package logger

import (
	"os"
	"strings"
)

// 为什么需要身份字段
//
// 多实例部署下，"哪个服务、哪个副本打的这条日志"是定位问题的第一步。
// 此前只有 prefix 一个自由字符串，拼在消息前面：既无法作为结构化字段检索，
// 多实例聚合时也无从区分副本。
//
// 三个字段在设置时一次性预编码成 JSON 片段，每条日志只是一次 memcpy，
// 因此常驻字段不带来逐条的编码成本。

const (
	// FieldService 服务名字段
	FieldService = "service"
	// FieldInstance 实例（副本）标识字段
	FieldInstance = "instance"
	// FieldEnv 部署环境字段
	FieldEnv = "env"
)

// instanceEnvKeys 推导实例标识时依次尝试的环境变量
// POD_NAME 由 Kubernetes downward API 注入，HOSTNAME 在容器里即容器名
var instanceEnvKeys = []string{"INSTANCE_ID", "POD_NAME", "HOSTNAME"}

// WithService 设置服务名（输出为 service 字段）
func (l *Logger) WithService(service string) *Logger {
	l.service = strings.TrimSpace(service)
	l.rebuildIdentity()
	return l
}

// WithInstance 设置实例标识（输出为 instance 字段）
// 传空串表示按 INSTANCE_ID / POD_NAME / HOSTNAME 依次推导，都取不到时回落到主机名
func (l *Logger) WithInstance(instance string) *Logger {
	if instance = strings.TrimSpace(instance); instance == "" {
		instance = DetectInstance()
	}
	l.instance = instance
	l.rebuildIdentity()
	return l
}

// WithEnv 设置部署环境（输出为 env 字段，如 prod / staging / dev）
func (l *Logger) WithEnv(env string) *Logger {
	l.env = strings.TrimSpace(env)
	l.rebuildIdentity()
	return l
}

// DetectInstance 推导当前实例标识
// 依次尝试 INSTANCE_ID、POD_NAME、HOSTNAME 环境变量，再回落到主机名，最后返回 unknown
func DetectInstance() string {
	for _, key := range instanceEnvKeys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	if hostname, err := os.Hostname(); err == nil {
		if hostname = strings.TrimSpace(hostname); hostname != "" {
			return hostname
		}
	}
	return "unknown"
}

// rebuildIdentity 重建身份字段的预编码片段
// 仅在配置变更时调用，日志路径上只读该片段
func (l *Logger) rebuildIdentity() {
	if l.service == "" && l.instance == "" && l.env == "" {
		l.identityJSON = nil
		l.identityText = nil
		return
	}

	jsonBuf := make([]byte, 0, 64)
	textBuf := make([]byte, 0, 48)
	for _, f := range [...]struct{ key, value string }{
		{FieldService, l.service},
		{FieldInstance, l.instance},
		{FieldEnv, l.env},
	} {
		if f.value == "" {
			continue
		}
		jsonBuf = append(jsonBuf, ',')
		jsonBuf = appendJSONKey(jsonBuf, f.key)
		jsonBuf = appendJSONString(jsonBuf, f.value)

		textBuf = append(textBuf, '[')
		textBuf = appendTextSafe(textBuf, f.value)
		textBuf = append(textBuf, ']')
	}
	if len(textBuf) > 0 {
		textBuf = append(textBuf, ' ')
	}
	l.identityJSON = jsonBuf
	l.identityText = textBuf
}

// Service 返回已配置的服务名
func (l *Logger) Service() string { return l.service }

// Instance 返回已配置的实例标识
func (l *Logger) Instance() string { return l.instance }

// Env 返回已配置的部署环境
func (l *Logger) Env() string { return l.env }
