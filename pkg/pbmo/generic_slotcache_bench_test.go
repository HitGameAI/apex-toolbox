/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2026-10-22 00:00:00
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2026-10-27 11:35:07
 * @FilePath: \go-toolbox\pkg\pbmo\generic_slotcache_bench_test.go
 * @Description: 多类型对热路径槽位缓存回归基准
 *
 * 为什么需要它：lastEntries 是直接映射（1-way）槽组，同槽的两个类型对互相驱逐、永不命中
 * 单类型对基准**永远命中槽位**，因此看不到这类退化——槽位数被改小、哈希被改坏、
 * 或 ensureEntry 回源代价变高，单对基准全都测不出来
 * 本基准以 64 个交错访问的类型对覆盖该盲区：
 *   - 32 槽时命中率约 16%，ToPB 从 ~27ns 劣化到 ~91ns，且条目重建占全部分配字节的 46%
 *   - 512 槽时命中率约 93%，恢复到 ~23ns
 * 判读：Pairs64 若显著高于 Pairs1（约 2× 以上）或 B/op 明显高于 Pairs1，即为槽位缓存退化
 *
 * 类型对用 phantom 标记泛型生成（零尺寸 [0]T 字段只参与类型身份，不影响拷贝计划），避免 64 组结构体声明的样板膨胀
 *
 * Copyright (c) 2026 by kamalyes, All Rights Reserved.
 */

package pbmo

import "testing"

// slotPB / slotM 槽位基准专用类型载体
// [0]T 为零尺寸 phantom 字段：仅用于区分类型身份，不参与字段映射
type slotPB[T any] struct {
	Id     uint64
	Name   string
	Status int32
	_      [0]T
}

type slotM[T any] struct {
	ID     uint64 `pbmo:"Id"`
	Name   string
	Status int
	_      [0]T
}

// 槽位基准的 phantom 标记类型
type (
	slotMk0  struct{}
	slotMk1  struct{}
	slotMk2  struct{}
	slotMk3  struct{}
	slotMk4  struct{}
	slotMk5  struct{}
	slotMk6  struct{}
	slotMk7  struct{}
	slotMk8  struct{}
	slotMk9  struct{}
	slotMk10 struct{}
	slotMk11 struct{}
	slotMk12 struct{}
	slotMk13 struct{}
	slotMk14 struct{}
	slotMk15 struct{}
	slotMk16 struct{}
	slotMk17 struct{}
	slotMk18 struct{}
	slotMk19 struct{}
	slotMk20 struct{}
	slotMk21 struct{}
	slotMk22 struct{}
	slotMk23 struct{}
	slotMk24 struct{}
	slotMk25 struct{}
	slotMk26 struct{}
	slotMk27 struct{}
	slotMk28 struct{}
	slotMk29 struct{}
	slotMk30 struct{}
	slotMk31 struct{}
	slotMk32 struct{}
	slotMk33 struct{}
	slotMk34 struct{}
	slotMk35 struct{}
	slotMk36 struct{}
	slotMk37 struct{}
	slotMk38 struct{}
	slotMk39 struct{}
	slotMk40 struct{}
	slotMk41 struct{}
	slotMk42 struct{}
	slotMk43 struct{}
	slotMk44 struct{}
	slotMk45 struct{}
	slotMk46 struct{}
	slotMk47 struct{}
	slotMk48 struct{}
	slotMk49 struct{}
	slotMk50 struct{}
	slotMk51 struct{}
	slotMk52 struct{}
	slotMk53 struct{}
	slotMk54 struct{}
	slotMk55 struct{}
	slotMk56 struct{}
	slotMk57 struct{}
	slotMk58 struct{}
	slotMk59 struct{}
	slotMk60 struct{}
	slotMk61 struct{}
	slotMk62 struct{}
	slotMk63 struct{}
)

// slotConvFns 64 个类型对的转换入口，交错调用以暴露槽位互驱逐
var slotConvFns = [...]func(){
	func() { _, _ = ToPB[slotM[slotMk0], slotPB[slotMk0]](&slotM[slotMk0]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk1], slotPB[slotMk1]](&slotM[slotMk1]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk2], slotPB[slotMk2]](&slotM[slotMk2]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk3], slotPB[slotMk3]](&slotM[slotMk3]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk4], slotPB[slotMk4]](&slotM[slotMk4]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk5], slotPB[slotMk5]](&slotM[slotMk5]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk6], slotPB[slotMk6]](&slotM[slotMk6]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk7], slotPB[slotMk7]](&slotM[slotMk7]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk8], slotPB[slotMk8]](&slotM[slotMk8]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk9], slotPB[slotMk9]](&slotM[slotMk9]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk10], slotPB[slotMk10]](&slotM[slotMk10]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk11], slotPB[slotMk11]](&slotM[slotMk11]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk12], slotPB[slotMk12]](&slotM[slotMk12]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk13], slotPB[slotMk13]](&slotM[slotMk13]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk14], slotPB[slotMk14]](&slotM[slotMk14]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk15], slotPB[slotMk15]](&slotM[slotMk15]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk16], slotPB[slotMk16]](&slotM[slotMk16]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk17], slotPB[slotMk17]](&slotM[slotMk17]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk18], slotPB[slotMk18]](&slotM[slotMk18]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk19], slotPB[slotMk19]](&slotM[slotMk19]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk20], slotPB[slotMk20]](&slotM[slotMk20]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk21], slotPB[slotMk21]](&slotM[slotMk21]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk22], slotPB[slotMk22]](&slotM[slotMk22]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk23], slotPB[slotMk23]](&slotM[slotMk23]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk24], slotPB[slotMk24]](&slotM[slotMk24]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk25], slotPB[slotMk25]](&slotM[slotMk25]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk26], slotPB[slotMk26]](&slotM[slotMk26]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk27], slotPB[slotMk27]](&slotM[slotMk27]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk28], slotPB[slotMk28]](&slotM[slotMk28]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk29], slotPB[slotMk29]](&slotM[slotMk29]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk30], slotPB[slotMk30]](&slotM[slotMk30]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk31], slotPB[slotMk31]](&slotM[slotMk31]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk32], slotPB[slotMk32]](&slotM[slotMk32]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk33], slotPB[slotMk33]](&slotM[slotMk33]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk34], slotPB[slotMk34]](&slotM[slotMk34]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk35], slotPB[slotMk35]](&slotM[slotMk35]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk36], slotPB[slotMk36]](&slotM[slotMk36]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk37], slotPB[slotMk37]](&slotM[slotMk37]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk38], slotPB[slotMk38]](&slotM[slotMk38]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk39], slotPB[slotMk39]](&slotM[slotMk39]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk40], slotPB[slotMk40]](&slotM[slotMk40]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk41], slotPB[slotMk41]](&slotM[slotMk41]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk42], slotPB[slotMk42]](&slotM[slotMk42]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk43], slotPB[slotMk43]](&slotM[slotMk43]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk44], slotPB[slotMk44]](&slotM[slotMk44]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk45], slotPB[slotMk45]](&slotM[slotMk45]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk46], slotPB[slotMk46]](&slotM[slotMk46]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk47], slotPB[slotMk47]](&slotM[slotMk47]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk48], slotPB[slotMk48]](&slotM[slotMk48]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk49], slotPB[slotMk49]](&slotM[slotMk49]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk50], slotPB[slotMk50]](&slotM[slotMk50]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk51], slotPB[slotMk51]](&slotM[slotMk51]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk52], slotPB[slotMk52]](&slotM[slotMk52]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk53], slotPB[slotMk53]](&slotM[slotMk53]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk54], slotPB[slotMk54]](&slotM[slotMk54]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk55], slotPB[slotMk55]](&slotM[slotMk55]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk56], slotPB[slotMk56]](&slotM[slotMk56]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk57], slotPB[slotMk57]](&slotM[slotMk57]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk58], slotPB[slotMk58]](&slotM[slotMk58]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk59], slotPB[slotMk59]](&slotM[slotMk59]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk60], slotPB[slotMk60]](&slotM[slotMk60]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk61], slotPB[slotMk61]](&slotM[slotMk61]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk62], slotPB[slotMk62]](&slotM[slotMk62]{ID: 1, Name: "n"}) },
	func() { _, _ = ToPB[slotM[slotMk63], slotPB[slotMk63]](&slotM[slotMk63]{ID: 1, Name: "n"}) },
}

// benchSlotPairs 以前 k 个类型对循环交错调用
func benchSlotPairs(b *testing.B, k int) {
	// 预热：确保全部类型对已注册，基准只测稳态查找与转换
	for i := 0; i < k; i++ {
		slotConvFns[i]()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slotConvFns[i%k]()
	}
}

func BenchmarkSlotCache_Pairs1(b *testing.B)  { benchSlotPairs(b, 1) }
func BenchmarkSlotCache_Pairs16(b *testing.B) { benchSlotPairs(b, 16) }
func BenchmarkSlotCache_Pairs64(b *testing.B) { benchSlotPairs(b, 64) }
