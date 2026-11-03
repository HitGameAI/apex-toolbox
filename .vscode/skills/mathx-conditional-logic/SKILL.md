---
name: mathx-conditional-logic
description: 数学工具与三元运算包，提供泛型条件表达式、空值/零值默认值、切片统计与变换、map 操作、位运算、概率与克隆当需要在表达式中做三元判断、处理空指针零值默认值、计算百分比/均值/最值、或做切片/map 变换时使用
---

# mathx - 数学工具与三元运算

提供泛型条件运算（IF/IfDo/IfChain 系列）、空值零值安全默认值、切片变换与统计、map 操作、位运算、概率与深克隆

> `mathx` 的 nil/零值/函数类型判断已下沉到 `types` 与 `go-argus`（validator），例如 `IfNil`、`IfCEmpty` 通过 `validator.IsNil` / `validator.IsCEmpty` 判断

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/mathx"
```

三元表达式：
```go
result := mathx.IF(age >= 18, "adult", "minor")
```

安全默认值：
```go
val := mathx.IfNotEmpty(name, "default")
num := mathx.IfNotZero(count, 10)
```

链式条件构建器：
```go
val, ok := mathx.NewIFChain[int]().When(x > 0).ThenReturn(1).When(x < 0).ThenReturn(-1).Execute()
```

## 完整API索引

### 函数

#### 条件运算（基础）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IF[T]` | `func(condition bool, trueVal, falseVal T) T` | 三元运算符泛型版本 |
| `IfDo[T]` | `func(condition bool, do DoFunc[T], defaultVal T) T` | 条件为 true 时执行函数返回结果 |
| `IfDoAF[T]` | `func(condition bool, do, defaultFunc DoFunc[T]) T` | 条件为 true 执行 do，否则执行 defaultFunc |
| `IfDoWithError[T]` | `func(condition bool, do DoFuncWithError[T], defaultVal T) (T, error)` | 条件执行带错误返回 |
| `IfDoAsync[T]` | `func(condition bool, do DoFunc[T], defaultVal ...T) <-chan T` | 条件异步执行，可选默认值 |
| `IfDoAsyncWithTimeout[T]` | `func(condition bool, do DoFunc[T], timeoutMs int, defaultVal ...T) <-chan T` | 异步执行带超时控制 |
| `IfDoWithErrorAsync[T]` | `func(condition bool, do DoFuncWithError[T], defaultVal T) <-chan ResultWithError[T]` | 异步执行带错误返回 |
| `IfDoWithErrorDefault[T]` | `func(condition bool, do DoFuncWithError[T], defaultVal T) T` | 条件执行，失败或错误返回默认值 |
| `ReturnIfErr[T]` | `func(val T, err error) (T, error)` | err 非 nil 返回零值和 err |
| `IfElse[T]` | `func(conditions []bool, values []T, defaultVal T) T` | 多条件链式判断（数组形式） |
| `IfChain[T]` | `func(pairs []ConditionValue[T], defaultVal T) T` | 多条件链式判断（结构体形式） |
| `IfCall[T]` | `func(condition bool, result T, err error, callbacks ...func(T, error))` | 按条件选择性调用回调 |
| `IfExec` | `func(condition bool, action func())` | 条件为 true 时执行副作用 |
| `IfExecElse` | `func(condition bool, onTrue func(), onFalse ...func())` | 条件执行不同副作用分支 |
| `MarshalJSONOrDefault` | `func(value any, defaultVal string) string` | 序列化为 JSON，失败/空值返回默认值 |
| `IfStrFmt` | `func(condition bool, trueFormat string, trueArgs []any, falseFormat string, falseArgs []any) (string, []any)` | 条件选择格式化字符串 |

#### 链式构建器

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `When` | `func(condition bool) *IfChainer` | 创建副作用链式构建器 |
| `WhenValue[T]` | `func(condition bool) *IfValueChainer[T]` | 创建带返回值的链式构建器 |
| `NewIFChain[T]` | `func() *IFChainBuilder[T]` | 创建链式条件构建器 |
| `IFChain` | `func() *IFChainBuilder[any]` | 全局链式构建器入口（自动推断 any） |
| `IFChainFor[T]` | `func() *IFChainBuilder[T]` | 为特定类型创建链式构建器 |
| `IFErrorChain` | `func() *IFChainBuilder[error]` | 错误处理专用链式构建器 |
| `IFNilChain` | `func() *IFChainBuilder[any]` | 返回 nil 的链式构建器 |

#### 安全默认值

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IfNotEmpty` | `func(str, defaultVal string) string` | 空字符串时返回默认值 |
| `IfNotZero[T]` | `func(val, defaultVal T) T` | 零值时返回默认值（comparable） |
| `IfLeZero[T]` | `func(val, defaultVal T) T` | 小于等于零时返回默认值（Numerical） |
| `IfLtZero[T]` | `func(val, defaultVal T) T` | 小于零时返回默认值 |
| `IfGeZero[T]` | `func(val, defaultVal T) T` | 大于等于零时返回默认值 |
| `IfGtZero[T]` | `func(val, defaultVal T) T` | 大于零时返回默认值 |
| `IfSafeIndex[T]` | `func(slice []T, index int, defaultVal T) T` | 安全索引访问，越界返回默认值 |
| `IfSafeKey[K,V]` | `func(m map[K]V, key K, defaultVal V) V` | 安全键访问，不存在返回默认值 |
| `DefaultIfNilPtr[T]` | `func(param *T, defaultValue T) *T` | nil 指针返回指向默认值的指针 |
| `IfNotNil[T]` | `func(val *T, defaultVal T) T` | 非 nil 指针返回解引用值 |
| `IfProtoTimeOr` | `func(protoTime *timestamppb.Timestamp, duration time.Duration) time.Time` | proto 时间戳转 time.Time |
| `IfProtoTimeOrPtr` | `func(protoTime *timestamppb.Timestamp, duration time.Duration) *time.Time` | proto 时间戳转 *time.Time |
| `IfTimeToProto` | `func(t *time.Time, duration time.Duration) *timestamppb.Timestamp` | time.Time 转 proto 时间戳 |

#### validator 系列条件判断

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IfEmpty[T]` | `func(val, defaultVal T) T` | 空值时返回默认值 |
| `IfNotEmptyValue[T]` | `func(val, defaultVal T) T` | 非空值时返回原值 |
| `IfEmptyDo[T]` | `func(val T, do DoFunc[T]) T` | 空值时执行函数 |
| `IfAllEmpty[T]` | `func(values []interface{}, trueVal, falseVal T) T` | 全部为空返回 trueVal |
| `IfHasEmpty[T]` | `func(values []interface{}, trueVal, falseVal T) T` | 包含空值返回 trueVal |
| `IfNil[T]` | `func(val interface{}, trueVal, falseVal T) T` | nil 时返回 trueVal |
| `IfNotNilValue[T]` | `func(val interface{}, trueVal, falseVal T) T` | 非 nil 时返回 trueVal |
| `IfCEmpty[T,R]` | `func(val T, trueVal, falseVal R) R` | 可比较类型零值时返回 trueVal |
| `IfNotCEmpty[T,R]` | `func(val T, trueVal, falseVal R) R` | 可比较类型非零值时返回 trueVal |
| `IfIPAllowed[T]` | `func(ip string, cidrList []string, trueVal, falseVal T) T` | IP 白名单检查 |
| `IfSafeFieldName[T]` | `func(field string, trueVal, falseVal T) T` | 安全字段名检查 |
| `IfAllowedField[T]` | `func(field string, allowedFields []string, trueVal, falseVal T) T` | 字段白名单检查 |
| `IfContainsChinese[T]` | `func(str string, trueVal, falseVal T) T` | 中文字符检查 |
| `IfUndefined[T]` | `func(str string, trueVal, falseVal T) T` | "undefined" 字符串检查 |
| `IfNull[T]` | `func(str string, trueVal, falseVal T) T` | "null" 字符串检查 |
| `IfNullOrUndefined[T]` | `func(str string, trueVal, falseVal T) T` | "null" 或 "undefined" 检查 |

#### 数值比较与区间

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IfGt[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 大于比较（Numerical） |
| `IfGe[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 大于等于比较 |
| `IfLt[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 小于比较 |
| `IfLe[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 小于等于比较 |
| `IfEq[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 等于比较（comparable） |
| `IfNe[T,R]` | `func(a, b T, trueVal, falseVal R) R` | 不等于比较 |
| `IfBetween[T]` | `func(val, min, max T, trueVal, falseVal T) T` | 区间检查（Numerical） |
| `IfClamp[T]` | `func(val, min, max T) T` | 将值限制在 [min, max] 范围内 |
| `IfDefaultAndClamp[T]` | `func(val, defaultVal, min, max T) T` | 先应用默认值再做范围限制 |
| `IfCast[R]` | `func(val any, defaultVal R) R` | 类型断言转换 |

#### 切片/集合条件运算

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IfContains[T]` | `func(slice []T, target T, trueVal, falseVal T) T` | 包含检查 |
| `IfAny[T]` | `func(conditions []bool, trueVal, falseVal T) T` | 任意条件满足 |
| `IfAll[T]` | `func(conditions []bool, trueVal, falseVal T) T` | 所有条件满足 |
| `IfCount[T]` | `func(conditions []bool, threshold int, trueVal, falseVal T) T` | 计数条件判断 |
| `IfEmptySlice[T,R]` | `func(slice []T, trueVal, falseVal R) R` | 空切片检查 |
| `IfLenGt[T,R]` | `func(slice []T, threshold int, trueVal, falseVal R) R` | 长度大于检查 |
| `IfLenEq[T,R]` | `func(slice []T, length int, trueVal, falseVal R) R` | 长度等于检查 |
| `IfErrOrNil[T,R]` | `func(val T, err error, trueVal, falseVal R) R` | 错误或零值检查 |
| `IfCountGt[R]` | `func(count, threshold int64, trueVal, falseVal R) R` | 计数大于检查 |
| `IfMulti[T,R]` | `func(target T, values []T, trueVal, falseVal R) R` | 多值匹配 |
| `IfMap[T,R]` | `func(condition bool, val T, mapper func(T) R, defaultVal R) R` | 条件映射 |
| `IfMapElse[T,R]` | `func(condition bool, val T, trueMapper func(T) R, falseMapper ...func(T) R) R` | 双向映射 |
| `IfFilter[T]` | `func(useFilter bool, slice []T, predicate func(T) bool) []T` | 条件过滤 |
| `IfValidate[T,R]` | `func(val T, validator func(T) bool, validVal, invalidVal R) R` | 验证三元 |
| `IfSwitch[K,V]` | `func(key K, cases map[K]V, defaultVal V) V` | 开关式映射 |
| `IfTryParse[T,R]` | `func(input T, parser func(T) (R, error), defaultVal R) R` | 尝试解析 |
| `IfPipeline[T]` | `func(condition bool, input T, funcs []func(T) T, defaultVal T) T` | 管道式执行 |
| `IfMemoized[T]` | `func(condition bool, key string, cache map[string]T, computeFn func() T, defaultVal T) T` | 带缓存的三元运算 |

#### 数值函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Min[T]` | `func(values ...T) T` | 返回多个数值最小值（Numerical） |
| `Max[T]` | `func(values ...T) T` | 返回多个数值最大值 |
| `AtLeast[T]` | `func(x, lower T) T` | 返回 x 和 lower 的最小值 |
| `AtMost[T]` | `func(x, upper T) T` | 返回 x 和 upper 的最大值 |
| `Between[T]` | `func(x, lower, upper T) T` | 将 x 限制在 [lower, upper] 范围内 |
| `Abs[T]` | `func(x T) T` | 绝对值（Numerical） |
| `Decimals[T]` | `func(num T, digit int) string` | 转换为指定小数位的字符串 |
| `ParseInt64` | `func(s string) (int64, error)` | 字符串解析为 int64 |
| `LongestCommonPrefix` | `func(a, b string) int` | 两个字符串最长公共前缀长度 |
| `CountPathSegments` | `func(path string, prefixes ...string) int` | 计算路径中参数数量（默认 `:` 和 `*`） |
| `ZeroValue[T]` | `func() T` | 返回类型 T 的零值 |
| `EqualSlices[T]` | `func(a, b []T) bool` | 比较两个切片是否相等 |
| `SafeGetIndexWithErr[T]` | `func(slice []T, index int) (T, error)` | 安全获取索引元素 |
| `SafeGetIndexOrDefault[T]` | `func(slice []T, index int, defaultVal T) T` | 安全获取或返回默认值 |
| `SafeGetIndexOrDefaultNoSpace` | `func(slice []string, index int, defaultVal string) string` | 安全获取并去空格 |
| `MergeLayeredScalar[T,V]` | `func(layers []*T, fieldGetter func(*T) V) V` | 多层级标量合并 |

#### 统计函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Percentile` | `func(values []float64, p float64) float64` | 计算百分位数 |
| `Percentiles` | `func(values []float64, percentiles ...float64) map[float64]float64` | 批量计算多个百分位数 |
| `Percentage` | `func(part, total uint64) float64` | 计算百分比 |
| `FormatPercentage` | `func(part, total uint64, precision int) string` | 格式化百分比字符串 |
| `Mean` | `func(values []float64) float64` | 计算均值 |
| `StdDev` | `func(values []float64) float64` | 计算标准差 |
| `MinSlice` | `func(values []float64) float64` | 计算最小值（Deprecated，用 SliceMinOrdered） |
| `MaxSlice` | `func(values []float64) float64` | 计算最大值（Deprecated，用 SliceMaxOrdered） |
| `SliceMinMax[T]` | `func(list []T, f types.MinMaxFunc[T]) (T, error)` | 泛型切片最值 |
| `SummarizeStats` | `func(values []float64) StatsSummary` | 生成统计摘要 |
| `SortByCount[T]` | `func(items []T, getCount func(T) uint64)` | 按计数降序排序 |
| `SortByKey[T,K]` | `func(items []T, getKey func(T) K)` | 按键升序排序 |
| `SortByKeyDesc[T,K]` | `func(items []T, getKey func(T) K)` | 按键降序排序 |
| `SortByKeyDescUnique[T,K,ID]` | `func(items []T, getKey func(T) K, getID func(T) ID) []T` | 按键降序排序并去重 |

#### 切片变换（SliceXxx 系列）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SliceMinOrdered[T]` | `func(collection []T) T` | Ordered 类型切片最小值 |
| `SliceMaxOrdered[T]` | `func(collection []T) T` | Ordered 类型切片最大值 |
| `SliceMinBy[T]` | `func(collection []T, comparison func(a, b T) bool) T` | 自定义比较最小值 |
| `SliceMaxBy[T]` | `func(collection []T, comparison func(a, b T) bool) T` | 自定义比较最大值 |
| `SliceMinIndexBy[T]` | `func(collection []T, comparison func(a, b T) bool) (T, int)` | 自定义比较最小值及索引 |
| `SliceMaxIndexBy[T]` | `func(collection []T, comparison func(a, b T) bool) (T, int)` | 自定义比较最大值及索引 |
| `SliceUniq[T]` | `func(list T) T` | 集合去重 |
| `SliceWithout[T]` | `func(list T, exclude ...E) T` | 排除指定值 |
| `SliceIntersect[T]` | `func(list1, list2 T) T` | 交集 |
| `SliceContainsComparable[T]` | `func(slice []T, element T) bool` | 包含检查（支持任意 comparable） |
| `SliceHasDuplicates[T]` | `func(slice []T) bool` | 是否存在重复元素 |
| `SliceRemoveEmpty[T]` | `func(slice []T) []T` | 移除空元素（反射判断） |
| `CompactSlice[T]` | `func(slice []T) []T` | 移除零值元素（comparable） |
| `SliceChunk[T]` | `func(slice []T, size int) [][]T` | 分块 |
| `SliceEqual[T]` | `func(a, b []T) bool` | 比较切片是否相等 |
| `SliceFisherYates[T]` | `func(slice []T, maxRetries int) error` | Fisher-Yates 洗牌 |
| `SliceDiffSetSorted[T]` | `func(arr1, arr2 []T) []T` | 已排序数组对称差集 |
| `SliceDifference[T]` | `func(list1, list2 T) (T, T)` | 双向差集 |
| `SliceUnionMulti[T]` | `func(lists ...T) T` | 多切片并集（保持首次出现顺序） |
| `SliceElementsMatch[T]` | `func(list1, list2 []T) bool` | 是否包含相同元素集合（含重复计数） |
| `TransformSlice[T,R]` | `func(slice []T, transform func(T) R) []R` | 映射切片 |
| `FilterSlice[T]` | `func(slice []T, predicate func(T) bool) []T` | 过滤切片 |
| `ReduceSlice[T,R]` | `func(slice []T, initial R, accumulator func(R, T) R) R` | 归约 |
| `FlattenSlice[T]` | `func(slices [][]T) []T` | 扁平化嵌套切片 |
| `ReverseSlice[T]` | `func(slice []T) []T` | 反转切片（返回新切片） |
| `ReverseSliceInPlace[T]` | `func(slice []T)` | 原地反转切片 |
| `TakeSlice[T]` | `func(slice []T, n int) []T` | 获取前 n 个元素 |
| `TakeLastSlice[T]` | `func(slice []T, n int) []T` | 获取后 n 个元素 |
| `SkipSlice[T]` | `func(slice []T, n int) []T` | 跳过前 n 个元素 |
| `SkipLastSlice[T]` | `func(slice []T, n int) []T` | 跳过后 n 个元素 |
| `PartitionSlice[T]` | `func(slice []T, predicate func(T) bool) ([]T, []T)` | 分区为满足/不满足两部分 |
| `GroupSliceBy[T,K]` | `func(slice []T, keyFunc func(T) K) map[K][]T` | 按键分组 |
| `FindSlice[T]` | `func(slice []T, predicate func(T) bool) (T, bool)` | 查找首个满足条件的元素 |
| `FindLastSlice[T]` | `func(slice []T, predicate func(T) bool) (T, bool)` | 查找最后一个满足条件的元素 |
| `AllSlice[T]` | `func(slice []T, predicate func(T) bool) bool` | 是否全部满足 |
| `AnySlice[T]` | `func(slice []T, predicate func(T) bool) bool` | 是否任意满足 |
| `NoneSlice[T]` | `func(slice []T, predicate func(T) bool) bool` | 是否全部不满足 |
| `SliceCountBy[T]` | `func(collection []T, predicate func(T) bool) int` | 按条件计数 |
| `SliceCountValues[T]` | `func(collection []T) map[T]int` | 按值计数 |
| `SliceCountValuesBy[T,U]` | `func(collection []T, mapper func(T) U) map[U]int` | 按键计数 |
| `IndexOfSlice[T]` | `func(slice []T, item T) int` | 返回元素索引 |
| `LastIndexOfSlice[T]` | `func(slice []T, item T) int` | 返回元素最后索引 |
| `RemoveSliceAt[T]` | `func(slice []T, index int) []T` | 移除指定索引元素 |
| `ContainsAnySlice[T]` | `func(slice []T, items ...T) bool` | 包含任意一个 |
| `ContainsAllSlice[T]` | `func(slice []T, items ...T) bool` | 包含全部 |
| `UniqueSliceBy[T,K]` | `func(slice []T, keyFunc func(T) K) []T` | 按键去重 |
| `AddUniqueSlice[T]` | `func(slice []T, items ...T) ([]T, bool)` | 添加不存在的元素 |
| `SliceUniqMap[T,R]` | `func(collection []T, iteratee func(T, int) R) []R` | 转换并去重 |
| `SliceFlatMap[T,R]` | `func(collection []T, iteratee func(T, int) []R) []R` | 映射并展平 |
| `SliceFilterMap[T,R]` | `func(collection []T, callback func(T, int) (R, bool)) []R` | 过滤+映射组合 |
| `SliceForEachWhile[T]` | `func(collection []T, iteratee func(T, int) bool)` | 可中断遍历 |
| `SliceTimes[T]` | `func(count int, iteratee func(int) T) []T` | 生成 N 个元素 |
| `SliceInterleave[T]` | `func(collections ...Slice) Slice` | 交替合并多个切片 |
| `SliceToMap[T,K]` | `func(collection []T, iteratee func(T) K) map[K]T` | 切片转 map |
| `SliceFilterToMap[T,K,V]` | `func(collection []T, transform func(T, int) (K, V, bool)) map[K]V` | 过滤后转 map |
| `SliceDropWhile[T]` | `func(collection Slice, predicate func(T) bool) Slice` | 从开头丢弃满足条件的 |
| `SliceDropRightWhile[T]` | `func(collection Slice, predicate func(T) bool) Slice` | 从末尾丢弃满足条件的 |
| `SliceReject[T]` | `func(collection Slice, predicate func(T, int) bool) Slice` | 反向过滤 |
| `SliceSubset[T]` | `func(collection Slice, offset int, length uint) Slice` | 子切片 |
| `SliceReplace[T]` | `func(collection Slice, old, nEw T, n int) Slice` | 替换前 n 个 |
| `SliceReplaceAll[T]` | `func(collection Slice, old, nEw T) Slice` | 替换全部 |
| `SliceIsSorted[T]` | `func(collection []T) bool` | 检查是否升序 |
| `SliceIsSortedByKey[T,K]` | `func(collection []T, iteratee func(T) K) bool` | 按键检查是否升序 |
| `SliceCut[T]` | `func(collection, separator Slice) (before, after Slice, found bool)` | 围绕分隔符切割 |
| `SliceHasPrefix[T]` | `func(collection, prefix []T) bool` | 前缀检查 |
| `SliceHasSuffix[T]` | `func(collection, suffix []T) bool` | 后缀检查 |
| `SliceFindUniques[T]` | `func(collection Slice) Slice` | 返回只出现一次的元素 |
| `SliceFindUniquesBy[T,U]` | `func(collection Slice, iteratee func(T) U) Slice` | 按键返回只出现一次的元素 |
| `SliceFindDuplicates[T]` | `func(collection Slice) Slice` | 返回重复元素的首次出现 |
| `SliceFindDuplicatesBy[T,U]` | `func(collection Slice, iteratee func(T) U) Slice` | 按键返回重复元素 |
| `SliceWithoutBy[T,K]` | `func(collection Slice, iteratee func(T) K, exclude ...K) Slice` | 按键排除 |
| `SlicePartitionBy[T,K]` | `func(collection Slice, iteratee func(T) K) []Slice` | 按键分区 |
| `SliceGroupByMap[T,K,V]` | `func(collection []T, iteratee func(T) (K, V)) map[K][]V` | 按键分组（value 可转换） |
| `DedupeValues[T,K]` | `func(items []T, key func(T) (K, bool)) []K` | 按键提取去重后的值列表 |
| `RepeatField[T]` | `func(field T, count int) []T` | 生成重复元素的切片 |
| `InsertionSort` | `func(arr []int)` | 插入排序 |
| `QuickSort` | `func(arr []int, low, high int)` | 快速排序 |
| `BubbleSort` | `func(arr []int)` | 冒泡排序 |

#### 切片过滤与查找

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `MapSliceByKey[T,K]` | `func(slice []T, keyFunc func(T) K) map[K]T` | 切片转 map |
| `FilterSliceByFunc[T]` | `func(slice []T, predicate func(T) bool) []T` | 通用过滤函数 |
| `NewSliceFilter[T]` | `func(slice []T) *SliceFilter[T]` | 创建切片过滤器（策略模式） |
| `FindUpdate[T,V]` | `func(item *T, dataMap map[string]V, getKey func(*T, ...any) string, keyArgs ...any) *FindResult[T,V]` | 查找并支持回写 |

#### map 操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `DeepMergeMap` | `func(target, source map[string]interface{}, options *MapMergeOptions) (map[string]interface{}, error)` | 深度合并两个 map |
| `ShallowMergeMap[K,V]` | `func(maps ...map[K]V) map[K]V` | 浅合并多个 map（Deprecated，用 MapAssign） |
| `ConvertMapKeysToString` | `func(data interface{}) interface{}` | 递归转换 map 键为字符串 |
| `GetNestedMapValue[T]` | `func(m map[string]interface{}, keys ...string) (T, bool)` | 从嵌套 map 获取值 |
| `SetNestedMapValue` | `func(m map[string]interface{}, value interface{}, keys ...string)` | 在嵌套 map 设置值 |
| `FlattenMap` | `func(m map[string]interface{}, separator string) map[string]interface{}` | 扁平化嵌套 map |
| `UnflattenMap` | `func(m map[string]interface{}, separator string) map[string]interface{}` | 还原扁平化为嵌套 |
| `MapClone[K,V]` | `func(m map[K]V) map[K]V` | 浅拷贝 map |
| `MapKeys[K,V]` | `func(in ...map[K]V) []K` | 提取所有键 |
| `MapUniqKeys[K,V]` | `func(in ...map[K]V) []K` | 提取去重键 |
| `MapValues[K,V]` | `func(in ...map[K]V) []V` | 提取所有值 |
| `MapUniqValues[K,V]` | `func(in ...map[K]V) []V` | 提取去重值 |
| `MapHasKey[K,V]` | `func(in map[K]V, key K) bool` | 检查键是否存在 |
| `MapValueOr[K,V]` | `func(in map[K]V, key K, fallback V) V` | 返回值或 fallback |
| `MapPickBy[K,V,M]` | `func(in M, predicate func(K, V) bool) M` | 按 predicate 保留 |
| `MapOmitBy[K,V,M]` | `func(in M, predicate func(K, V) bool) M` | 按 predicate 移除 |
| `MapPickByKeys[K,V,M]` | `func(in M, keys []K) M` | 按键列表保留 |
| `MapOmitByKeys[K,V,M]` | `func(in M, keys []K) M` | 按键列表移除 |
| `MapPickByValues[K,V,M]` | `func(in M, values []V) M` | 按值列表保留 |
| `MapOmitByValues[K,V,M]` | `func(in M, values []V) M` | 按值列表移除 |
| `MapEntries[K,V]` | `func(in map[K]V) []MapEntry[K,V]` | map 转键值对切片 |
| `MapFromEntries[K,V]` | `func(entries []MapEntry[K,V]) map[K]V` | 键值对切片转 map |
| `MapInvert[K,V]` | `func(in map[K]V) map[V]K` | 反转键值 |
| `MapAssign[K,V,M]` | `func(maps ...M) M` | 从左到右合并多个 map |
| `MapToSlice[K,V,R]` | `func(in map[K]V, iteratee func(K, V) R) []R` | map 转切片 |
| `MapFilterToSlice[K,V,R]` | `func(in map[K]V, iteratee func(K, V) (R, bool)) []R` | 过滤+转切片 |
| `MapFilterKeys[K,V]` | `func(in map[K]V, predicate func(K, V) bool) []K` | 过滤返回键切片 |
| `MapFilterValues[K,V]` | `func(in map[K]V, predicate func(K, V) bool) []V` | 过滤返回值切片 |
| `MapMapEntries[K1,V1,K2,V2]` | `func(in map[K1]V1, iteratee func(K1, V1) (K2, V2)) map[K2]V2` | 同时转换键和值 |
| `MapMapKeys[K,V,R]` | `func(in map[K]V, iteratee func(V, K) R) map[R]V` | 转换所有键 |
| `MapMapValues[K,V,R]` | `func(in map[K]V, iteratee func(V, K) R) map[K]R` | 转换所有值 |
| `MergeLayeredKeyValues[T,KV]` | `func(layers []*T, fieldGetter func(*T) []KV, keyFieldName, valueFieldName string) []KV` | 多层级键值对合并 |
| `NewLayeredMerger[T,KV]` | `func(keyFieldName, valueFieldName string) *LayeredMerger[T,KV]` | 创建多层级合并器 |

#### 位运算

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetBit64` | `func(min, max, step uint) uint64` | 生成 64 位掩码 |
| `Bit64ToArray` | `func(bit uint64) []uint` | 掩码转位数组 |
| `GetBitBig` | `func(min, max, step uint) *big.Int` | 生成大整数掩码 |
| `BitToArrayBig` | `func(bits *big.Int) []uint` | big.Int 掩码转位数组 |

#### 概率与熵

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `NewProba` | `func() *Proba` | 创建概率器 |
| `NewUnstable` | `func(deviation float64) Unstable` | 创建不稳定随机器 |
| `CalcEntropy` | `func(m map[interface{}]int) float64` | 计算概率分布的熵值 |
| `Clone` | `func(value interface{}, seen map[uintptr]interface{}) interface{}` | 深拷贝任意类型值 |

### 类型

| 导出名称 | 说明 |
|---|---|
| `DoFunc[T]` | 条件执行函数类型 `func() T` |
| `DoFuncWithError[T]` | 带错误的条件执行函数类型 `func() (T, error)` |
| `ConditionValue[T]` | 条件值结构体（Cond/Value） |
| `ResultWithError[T]` | 结果与错误封装（Result/Err） |
| `IfChainer` | 副作用链式调用构建器 |
| `IfValueChainer[T]` | 带返回值的链式调用构建器 |
| `IFChainBuilder[T]` | 链式条件构建器 |
| `IFChainBuilderCondition[T]` | 条件构建器中间态 |
| `IFChainBuilderAction[T]` | 操作构建器中间态 |
| `StatsSummary` | 统计摘要类型（Count/Min/Max/Mean/StdDev/P50/P90/P95/P99） |
| `MapMergeStrategy` | map 合并策略枚举 |
| `MapMergeOptions` | map 合并选项 |
| `MapEntry[K,V]` | map 键值对结构体 |
| `LayeredMerger[T,KV]` | 多层级键值对合并器 |
| `Predicate[T]` | 判断条件函数类型 `func(T) bool` |
| `FilterCallback[T]` | 过滤回调函数类型 `func(*T)` |
| `SliceFilter[T]` | 策略模式切片过滤器 |
| `FindResult[T,V]` | 查找结果类型 |
| `SliceChain[T]` | 链式切片操作结构体 |
| `Proba` | 概率器类型 |
| `Unstable` | 不稳定随机器类型 |

### 常量/变量

| 导出名称 | 值/类型 | 说明 |
|---|---|---|
| `MapMergeStrategyOverwrite` | MapMergeStrategy | 覆盖策略：源覆盖目标 |
| `MapMergeStrategyKeepExisting` | MapMergeStrategy | 保持现有：保留目标值 |
| `MapMergeStrategyError` | MapMergeStrategy | 冲突报错策略 |

### 关键类型方法

**IfChainer**: `Then(action func()) *IfChainer`, `Else(action func()) *IfChainer`, `Do()`

**IfValueChainer[T]**: `ThenReturn(val T)`, `ElseReturn(val T)`, `ThenDo(fn func() T)`, `ElseDo(fn func() T)`, `Get() T`

**IFChainBuilder[T]**: `When(condition bool)`, `Execute() (T, bool)`, `MustExecute() T`, `ExecuteOr(defaultValue T) T`, `HasResult() bool`

**IFChainBuilderCondition[T]**: `Then(action func())`, `ThenReturn(value T, action ...func())`, `ThenReturnNil(action ...func())`

**IFChainBuilderAction[T]**: `Return(value T)`, `ReturnNil()`, `ContinueChain()`

**Proba**: `TrueOnProba(proba float64) bool`

**Unstable**: `AroundDuration(base time.Duration) time.Duration`, `AroundInt(base int64) int64`

**SliceFilter[T]**: `UseAnd()`, `UseOr()`, `UseCustom(combine func([]bool) bool)`, `Condition(preds ...Predicate[T])`, `OnMatch(cb FilterCallback[T])`, `OnNotMatch(cb FilterCallback[T])`, `Result() []T`

**SliceChain[T]**: `Append(elements ...T)`, `Uniq()`, `RemoveValue(value T)`, `RemoveEmpty()`, `Filter(f func(T) bool)`, `Sort(less func(a, b T) bool)`, `Data() []T`, `String() string`

**FindResult[T,V]**: `Item() *T`, `IfFound(f func(*T, *V))`, `OrElse(f func(*T))`, `Then(onFound, onNotFound)`, `Stop()`, `When(cond bool, f)`, `Do(f)`

## 注意事项

- `IF` 即时求值两个分支，如需延迟求值用 `IfDo`
- `IfSafeIndex` 对越界索引返回默认值，不会 panic
- `Percentile` 要求切片已排序或内部会先排序
- `Percentiles` 返回 `map[float64]float64`，键为百分位值
- `Percentage` / `FormatPercentage` 参数为 `uint64` 类型
- `SliceFisherYates` 返回 `error`，洗牌失败（达到最大重试仍与原切片相同）时返回错误
- `Clone` 深拷贝使用 `seen map` 避免循环引用，需传入 `map[uintptr]interface{}` 参数
- 通用 nil/零值/函数类型判断优先放在 `types` 与 `go-argus`（validator），`mathx` 只保留条件表达式语义
- 链式构建器 `IFChainBuilder` 首个匹配条件执行后，后续条件自动跳过
