/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-20 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-27 09:17:37
 * @FilePath: \go-toolbox\pkg\pbmo\doc.go
 * @Description: 包文档
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

// Package pbmo 在 protobuf 生成结构体与领域/GORM 模型之间做字段映射转换
//
// 本包不参与 protobuf 线格式编解码：不调用 proto.Marshal/Unmarshal/Clone，
// 也不走 protojson对 google.golang.org/protobuf 的依赖仅用于 WKT 类型
// （timestamppb / wrapperspb / structpb）与一次 proto.Message 接口判定
//
// # 性能模型
//
// 转换计划按 reflect.Type 编译一次后缓存，热路径只做一次原子读 + 一次闭包调用，
// 稳态零反射字段按偏移合并为连续 memmove 段，WKT 子对象的多次分配被合并为
// 单次复合分配——因此 Model→PB 的 wrapper 路径可以比手写代码更快（分配更少）
//
// # 字段别名语义（重要）
//
// 为保持热路径零拷贝，标量与切片类字段**不做防御性复制**：
//
//   - 标量、定长数组：按值 memmove 到目标，相互独立
//   - []byte 与其他切片字段：**只拷贝切片头（指针 + 长度 + 容量）**，
//     源与目标因此共享同一底层数组修改 src[0] 会改变 dst[0]，反之亦然
//   - map 字段：拷贝 map 头，源与目标指向同一个 map
//   - 指针字段：拷贝指针值，源与目标指向同一对象
//
// 这是有意的设计（跨节点扇出等场景据此省掉整份载荷复制），但调用方必须知道：
// 若转换后仍会改写源对象，且目标的生命周期更长，则需要自行深拷贝相关字段
//
// PBToUpdates / ModelToUpdates 输出的 map 值是**值快照**（标量经值拷贝后装箱），
// 不随源结构体后续改动而变化；但指针类字段的值仍是共享指针
//
// # 目标归属与 Into 变体
//
// ToPB / FromPB 必须为结果分配目标并返回，该分配必然逃逸，是这两条路径
// 100% 的垃圾来源与并行扩展的实际天花板需要稳态零分配时改用
// ToPBInto / FromPBInto 写入调用方持有的目标，但要遵守其文档中的清零契约：
// 拷贝计划只写有映射关系的字段，复用脏目标会残留上一轮的值
package pbmo
