/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2025-12-18 19:36:05
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\doc.go
 * @Description: 包文档 - 三条入口的分工与取舍
 *
 * Copyright (c) 2025 by kamalyes, All Rights Reserved.
 */

// Package desensitize 提供数据脱敏：零第三方依赖，按结构体 tag、字段名或整段报文三种方式工作
//
// # 三条入口，判定依据不同
//
// 这三条路解决的不是同一个问题，取舍方向也不同，选错会既误擦又漏擦：
//
//	入口                        判定依据              适用场景        取舍
//	Desensitization(obj)       结构体 tag 逐字段声明    自己的 Model    最准，不猜
//	ActionForKey(key)          字段名严格匹配          日志 KV         保守，宁可漏擦
//	ActionForKeyLoose(key)     字段名 + 内部词根        报文 / 响应体    激进，宁可误擦
//	DataMasker.Mask(body)      同上，整段重写          请求 / 响应体    激进
//
// # 为什么 Model 不走字段名启发式
//
// 字段名规则表（keyrules.go）是按**日志**调出来的，它的取舍对 Model 是错的：
//
//   - name 不入表：通用日志里 service_name、table_name、file_name 远多于人名
//   - status_code、cache_key、trace_id 进豁免表：它们是排障主线索，擦掉日志就失去价值
//
// 而在一个 Model 里，Name 字段通常就是人名，Address 通常就是收货地址
// 所以 Model 这条路**完全不碰规则表**，只认你写的 tag：
//
//	type User struct {
//	    Name  string `desensitize:"name"`
//	    Phone string `desensitize:"mobilePhone"`
//	    Email string `desensitize:"email"`
//	    Memo  string // 没有 tag 就不动
//	}
//	err := desensitize.Desensitization(&user)
//
// 你掌握 schema 时显式声明永远比猜更准，这也是 tag 存在的理由
//
// # 什么时候才该用字段名启发式
//
// 只有在**拿不到 schema** 时才猜：日志的 KV 参数来自各处调用方，
// HTTP 响应体可能是上游返回的任意 JSON，这些场合没有 tag 可依
//
// 两种强度对应两种代价：
//
//   - 日志用 ActionForKey：误擦会让排障线索静默消失，所以只认精确与后缀匹配，
//     并单列运维豁免表兜底
//   - 报文用 ActionForKeyLoose：漏擦等于把凭据写进了日志或响应，
//     所以额外扫字段名内部的敏感词根，my_token_field 这类也会命中
//
// # 零依赖
//
// 主模块不引入任何第三方库，使用者编译本库时第三方包数为 0
// testify 仅测试使用，不会进入使用者的构建图
//
// # 迁移出处
//
// 本库由两处代码合并而成，都在 2025 年底提取：
//
//   - 按 tag 与按类型的脱敏（adapter.go、desensitize.go、masker.go、model.go、
//     constants.go）取自 github.com/HitGameAI/apex-toolbox v0.17.0 的 pkg/desensitize，
//     该包自 v0.17.0 起未再变动
//   - 按字段名的规则表与判定（keyrules.go）取自 github.com/HitGameAI/apex-toolbox/pkg/logger
//     的 sanitize.go，其中不依赖 Logger 的纯逻辑部分
//
// 提取后 go-toolbox 的 go-toolbox/pkg/argus 依赖与 go-toolbox/pkg/logger 的规则副本都不再需要，
// 200 余条安全规则从此只有一处来源
package desensitize
