/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-23 11:52:30
 * @FilePath: \go-toolbox\pkg\desensitize\adapter.go
 * @Description: 脱敏器注册表与结构体 tag 驱动的脱敏入口
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */

package desensitize

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// ==== 脱敏器接口 | Desensitizer ====

// Desensitizer 脱敏器接口，输入一个值返回脱敏后的值
type Desensitizer interface {
	Desensitize(value string) string
}

// DefaultDesensitizer 默认脱敏适配器，按脱敏类型分派到 Desensitize
type DefaultDesensitizer struct {
	desensitizerType DesensitizeType
}

// NewDefaultDesensitizer 以内置脱敏类型构造一个脱敏器
//
// 用于把内置类型注册到自定义的 tag 名下：
//
//	desensitize.RegisterDesensitizer("myType",
//	    desensitize.NewDefaultDesensitizer(desensitize.CustomExtension))
//
// 没有这个构造函数时包外无路可走：DefaultDesensitizer 的字段不导出，
// 复合字面量 &DefaultDesensitizer{CustomExtension} 在包外编译不过
// 上游文档里一直写着那种用法，但它从来没能编译 —— 这里把这条路补通
//
// [EN] Wraps a built-in masking type so it can be registered under a custom tag
// name. Without this there is no way to do it from outside the package: the
// field is unexported, so the composite literal shown in upstream's docs never
// actually compiled
func NewDefaultDesensitizer(t DesensitizeType) Desensitizer {
	return &DefaultDesensitizer{desensitizerType: t}
}

// Desensitize 按指定的脱敏类型处理输入值
func (e *DefaultDesensitizer) Desensitize(value string) string {
	return Desensitize(value, e.desensitizerType)
}

// DesensitizerFactory 参数化脱敏器工厂
// 根据标签中的整数参数创建脱敏器，如 jump(3,-4,1) 中的 3、-4、1
type DesensitizerFactory func(params ...int) Desensitizer

// ==== 注册表：读路径无锁 | Lock-free read path ====

// registryState 注册表的只读快照
//
// 原实现用 sync.RWMutex 且**走的是写锁**：OperateByRule 每个字段取一次独占锁，
// 于是并发脱敏完全串行化，顺序基准根本看不出来
// 改为「写时拷贝 + atomic.Pointer」：读路径一次原子加载，写路径（注册）加锁复制整表
// 注册是启动期行为、脱敏是稳态热路径，这个不对称正适合写时拷贝
//
// gen 是代数：每次注册递增，供类型计划判断自己是否已过期
// 没有它，加了计划缓存后「先脱敏、再注册自定义脱敏器」就会对已缓存的类型永久失效
//
// [EN] The original used an RWMutex but took the *write* lock: OperateByRule
// grabbed it per field, serialising all concurrent masking — invisible to a
// sequential benchmark. Now copy-on-write behind an atomic.Pointer: one atomic
// load to read, a full copy under a mutex to register. Registration happens at
// startup while masking is the steady-state hot path, so the asymmetry fits.
// gen is bumped on every registration so cached per-type plans can detect
// staleness; without it, registering a desensitizer after first use would never
// take effect for an already-cached type
type registryState struct {
	byName    map[string]Desensitizer
	factories map[string]DesensitizerFactory
	gen       uint64
}

var (
	registry   atomic.Pointer[registryState]
	registryMu sync.Mutex // 仅写者持有
)

// loadRegistry 取当前注册表快照，读路径唯一的同步开销
//
// 懒初始化而非依赖 init() 顺序：同包内 init() 按文件名排序执行，
// 而 cipher.go 的 init() 也要调 RegisterDesensitizer——
// 靠 adapter < cipher 的字母序侥幸正确，改个文件名就会在 nil 快照上崩
// [EN] Lazily initialised rather than relying on init() order: init functions
// run in filename order within a package, and cipher.go's init also registers a
// desensitizer. Depending on adapter sorting before cipher would break on a
// rename, panicking on a nil snapshot
func loadRegistry() *registryState {
	if state := registry.Load(); state != nil {
		return state
	}
	registry.CompareAndSwap(nil, builtinRegistry())
	return registry.Load()
}

// mutateRegistry 在锁内复制整表并应用变更，然后原子换入
func mutateRegistry(apply func(byName map[string]Desensitizer, factories map[string]DesensitizerFactory)) {
	registryMu.Lock()
	defer registryMu.Unlock()

	cur := registry.Load()
	if cur == nil {
		cur = builtinRegistry()
	}
	next := &registryState{
		byName:    make(map[string]Desensitizer, len(cur.byName)+1),
		factories: make(map[string]DesensitizerFactory, len(cur.factories)+1),
		gen:       cur.gen + 1,
	}
	for k, v := range cur.byName {
		next.byName[k] = v
	}
	for k, v := range cur.factories {
		next.factories[k] = v
	}
	apply(next.byName, next.factories)
	registry.Store(next)
}

// RegisterDesensitizer 注册脱敏器（支持覆盖内置项）
// 注册会令已缓存的类型计划作废并在下次使用时重建
func RegisterDesensitizer(desensitizerType string, desensitizer Desensitizer) {
	mutateRegistry(func(byName map[string]Desensitizer, _ map[string]DesensitizerFactory) {
		byName[desensitizerType] = desensitizer
	})
}

// RegisterDesensitizerFactory 注册参数化脱敏器工厂
// name 为标签中括号前的名称，如 jump
func RegisterDesensitizerFactory(name string, factory DesensitizerFactory) {
	mutateRegistry(func(_ map[string]Desensitizer, factories map[string]DesensitizerFactory) {
		factories[name] = factory
	})
}

// ==== 跳步脱敏器 | Jump desensitizer ====

// JumpDesensitizer 跳步脱敏器
// 按开始、结束、跳步参数对字符进行掩码，参数含义见 SensitizeJump
type JumpDesensitizer struct {
	start int // 开始索引（含），支持负数
	end   int // 结束索引（不含），支持负数（-N 表示保留末尾 N 个字符，0 表示到末尾）
	step  int // 跳步数
}

// Desensitize 按跳步规则对输入值脱敏
func (j *JumpDesensitizer) Desensitize(value string) string {
	return SensitizeJump(value, j.start, j.end, j.step)
}

// NewJumpDesensitizer 创建跳步脱敏器
// 参数依次为开始、结束、跳步，均可省略，省略时默认从头到尾连续掩码
func NewJumpDesensitizer(params ...int) Desensitizer {
	jumper := &JumpDesensitizer{step: 1}
	if len(params) > 0 {
		jumper.start = params[0]
	}
	if len(params) > 1 {
		jumper.end = params[1]
	}
	if len(params) > 2 {
		jumper.step = params[2]
	}
	return jumper
}

// ==== 标签解析 | Tag parsing ====

// errDesensitizerNotFound / errFactoryNotFound 预分配的哨兵错误
// 原实现每次失败都 errors.New 一次，而无 tag 字段走的恰是失败路径：
// 一个十字段结构体里七个无 tag，就是每次脱敏七次无谓分配
// [EN] Preallocated sentinels; the original allocated a fresh error on every
// miss, and untagged fields take exactly that path — seven needless allocations
// per call for a ten-field struct with three tags
var (
	errDesensitizerNotFound = errors.New("desensitizer not found")
	errFactoryNotFound      = errors.New("desensitizer factory not found")
)

// parseParameterizedTag 解析带参数的脱敏标签，如 jump(3,-4,1)
// 返回规则名称、整数参数列表、是否为带参数形式
//
// 原实现用正则 `^([^\s(),]+)\s*\(([^)]*)\)$` 做 FindStringSubmatch，
// 每个字段每次脱敏都要跑一遍，且普通标签（如 email）也得先白跑一次才知道不匹配
// 手写扫描等价且不分配：先找 '('，无则立刻判定非参数形式
//
// [EN] The original ran a regexp FindStringSubmatch per field per call, and a
// plain tag such as email had to fail that match first. Hand scanning is
// equivalent and allocation-free: locate '(' and bail out immediately when absent
func parseParameterizedTag(tag string) (name string, params []int, ok bool) {
	open := strings.IndexByte(tag, '(')
	if open <= 0 || tag[len(tag)-1] != ')' {
		return "", nil, false
	}

	// 名称部分不得含空白、括号与逗号，与原正则的 [^\s(),]+ 一致
	name = tag[:open]
	for i := 0; i < len(name); i++ {
		switch name[i] {
		case ' ', '\t', '\n', '\r', '\v', '\f', '(', ')', ',':
			return "", nil, false
		}
	}

	// 参数部分不得再含 ')'，与原正则的 [^)]* 一致
	argStr := tag[open+1 : len(tag)-1]
	if strings.IndexByte(argStr, ')') >= 0 {
		return "", nil, false
	}

	argStr = strings.TrimSpace(argStr)
	if argStr == "" {
		return name, nil, true // 无参数形式，如 jump()
	}
	for {
		comma := strings.IndexByte(argStr, ',')
		part := argStr
		if comma >= 0 {
			part = argStr[:comma]
		}
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return "", nil, false // 参数解析失败视为非法标签
		}
		params = append(params, value)
		if comma < 0 {
			break
		}
		argStr = argStr[comma+1:]
	}
	return name, params, true
}

// resolveDesensitizer 按标签解析脱敏器，支持普通名称与带参数形式
func resolveDesensitizer(state *registryState, tag string) (Desensitizer, error) {
	// 带参数形式：通过工厂创建，如 jump(3,-4,1)
	if name, params, ok := parseParameterizedTag(tag); ok {
		if factory, exists := state.factories[name]; exists {
			return factory(params...), nil
		}
		return nil, errFactoryNotFound
	}
	// 普通形式：优先精确匹配已注册的脱敏器
	if desensitizer, exists := state.byName[tag]; exists {
		return desensitizer, nil
	}
	// 普通形式未注册时，尝试以默认参数使用工厂，如 jump
	if factory, exists := state.factories[tag]; exists {
		return factory(), nil
	}
	return nil, errDesensitizerNotFound
}

// OperateByRule 根据规则对单个值脱敏
// desensitizerType 支持普通名称（如 email）与带参数形式（如 jump(3,-4,1)）
// in 应为字符串
func OperateByRule(desensitizerType string, in interface{}) (interface{}, error) {
	operator, err := resolveDesensitizer(loadRegistry(), desensitizerType)
	if err != nil {
		return nil, err
	}
	return operator.Desensitize(in.(string)), nil
}

// TagName 结构体字段上声明脱敏规则所用的 tag 名
//
// 对外导出是为了让下游不必硬编码这个字面量：go-toolbox/pkg/pbmo 这类库要在自己的字段
// 计划里读同一个 tag，若各自写死字符串，这里改名下游不会报错、只会静默不再
// 命中——脱敏静默失效正是最危险的失败方式
// 本包内部读 tag 也走这个常量，保证只有一处来源
//
// [EN] Name of the struct tag that declares a masking rule. It is exported so
// downstream libraries reading the same tag in their own field plans need not
// hardcode the literal: if each side spelled it out and this name changed,
// nothing would fail to compile and masking would simply stop matching in
// silence, which is the most dangerous way for it to break. This package reads
// the tag through the same constant, so there is a single source of truth
const TagName = "desensitize"

// MaskFunc 已绑定某个标签的脱敏函数，可长期持有并并发调用
type MaskFunc func(value string) string

// resolvedRule 某次解析的结果与其注册表代数
type resolvedRule struct {
	gen uint64
	d   Desensitizer
}

// ResolveRule 解析一次标签，返回可长期复用、并发安全的脱敏函数
//
// 适用于上游在「计划构建期」解析、在热路径上反复调用的场景：
// OperateByRule 每次调用都要重新查注册表、经 interface{} 装箱、并在未命中时
// 返回 error，这些开销对每条记录都付一遍；而标签到脱敏器的映射只取决于标签
// 本身与注册表代数，解析一次即可
//
// 返回的函数仍按代数自检：注册表变更（后注册或覆盖同名脱敏器）后自动重解析，
// 因此不会把早期解析的结果永久固化——这是与「调用方自己缓存 Desensitizer」
// 的关键区别
// 稳态开销是两次原子读加一次整数比较，重解析失败时沿用上次结果
//
// [EN] Resolves a tag once and returns a reusable, concurrency-safe masking
// function, for callers that resolve while building a plan and then call it per
// record. OperateByRule re-reads the registry, boxes through interface{} and
// allocates an error on a miss for every single call, yet the tag-to-desensitizer
// mapping depends only on the tag and the registry generation. The returned
// function still re-checks that generation, so a desensitizer registered or
// overridden later takes effect instead of being frozen out — the key difference
// from caching a Desensitizer yourself. Steady state is two atomic loads and an
// integer compare; if re-resolution fails the previous result is kept
// refreshRule 代数变化后的重解析，置于变量之后以便测试强制其失败
//
// 注册只增不减，因此一个解析成功过的标签不会再解析失败，这条分支实际不可达
// 但它的行为必须被闸住：重解析失败时要沿用上次结果，而不是把 nil 脱敏器
// 发布出去让热路径崩溃
//
// [EN] Re-resolution sits behind a variable so tests can force its failure
// branch. Registration only ever adds, so a tag that resolved once cannot fail
// later and the branch is unreachable in practice; its behaviour still matters,
// as it must keep the previous result rather than publishing a nil desensitizer
// that would crash the hot path
var refreshRule = resolveDesensitizer

func ResolveRule(tag string) (MaskFunc, error) {
	state := loadRegistry()
	d, err := resolveDesensitizer(state, tag)
	if err != nil {
		return nil, err
	}
	var cache atomic.Pointer[resolvedRule]
	cache.Store(&resolvedRule{gen: state.gen, d: d})
	return func(value string) string {
		cur := loadRegistry()
		got := cache.Load()
		if got.gen != cur.gen {
			// 代数变了：重解析并发布；并发下多个 goroutine 可能各解析一次，
			// 结果等价，最后一个写入生效
			if nd, nerr := refreshRule(cur, tag); nerr == nil {
				got = &resolvedRule{gen: cur.gen, d: nd}
				cache.Store(got)
			}
		}
		return got.d.Desensitize(value)
	}, nil
}

// TagMisuseError 字段的 desensitize tag 与字段类型不匹配
//
// 脱敏器只接受字符串，因此带 tag 的字段（或切片元素、映射值）必须是字符串形态
// 命名字符串类型（type Name string）同样可用，判定按 reflect.Kind 而非具体类型
type TagMisuseError struct {
	Struct string // 结构体类型名
	Field  string // 字段名
	Tag    string // 该字段的 desensitize tag
	Type   string // 字段的实际类型
}

func (e *TagMisuseError) Error() string {
	return "desensitize: " + e.Struct + "." + e.Field + " 的 tag " + strconv.Quote(e.Tag) +
		" 要求字符串字段，实际类型为 " + e.Type
}

// ==== 类型计划缓存 | Per-type plan cache ====

// fieldOp 字段的处理方式
type fieldOp uint8

const (
	opApply      fieldOp = iota // 直接对字段值套用脱敏器
	opRecurse                   // 结构体字段，递归下钻
	opPtrStruct                 // 结构体指针字段，非 nil 时解引用后递归
	opElems                     // 切片/数组，逐元素套用脱敏器
	opStructs                   // 切片/数组且元素为结构体，逐元素递归
	opPtrStructs                // 切片/数组且元素为结构体指针，逐元素解引用后递归
	opMap                       // 映射，逐值套用脱敏器
)

// fieldPlan 单个字段的处理计划
type fieldPlan struct {
	index int
	op    fieldOp
	d     Desensitizer // opRecurse / opStructs 时为 nil
}

// typePlan 一个结构体类型的处理计划
//
// 原实现每次脱敏都要：遍历全部字段、逐个读 tag、为每个字段跑一次正则、
// 取一次独占锁、查两次 map，失败路径还分配一个 error
// 而这些结果只取决于类型本身与注册表代数，按类型缓存一次即可
//
// [EN] The original re-read every tag, ran a regexp, took an exclusive lock and
// did two map lookups per field on every call, allocating an error on the miss
// path. All of it depends only on the type and the registry generation, so it is
// computed once per type
type typePlan struct {
	gen    uint64
	fields []fieldPlan
	err    error // 类型本身的 tag 用法错误，每次调用都原样返回
}

// typePlans 类型 → 计划，读多写极少
var typePlans sync.Map // reflect.Type -> *typePlan

// planFor 取（或构建）指定类型的处理计划
// 注册表代数变化时重建，使后注册的脱敏器对已缓存类型同样生效
func planFor(t reflect.Type, state *registryState) *typePlan {
	if cached, ok := typePlans.Load(t); ok {
		if plan := cached.(*typePlan); plan.gen == state.gen {
			return plan
		}
	}
	plan := buildTypePlan(t, state)
	typePlans.Store(t, plan)
	return plan
}

// buildTypePlan 构建类型计划，只收录真正需要动作的字段
//
// 保留原实现的判定顺序与语义：
//   - 不可导出字段跳过（原实现靠 CanSet 判断，入参是非 nil 指针的 Elem，二者等价）
//   - 结构体字段无条件递归，与 tag 无关
//   - 切片/数组元素是结构体时递归，否则逐元素套用 tag
//   - 其余类型套用 tag，解析不出脱敏器则原样跳过（原实现是忽略 error）
func buildTypePlan(t reflect.Type, state *registryState) *typePlan {
	plan := &typePlan{gen: state.gen}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(TagName)
		ft := field.Type

		switch ft.Kind() {
		case reflect.Struct:
			plan.fields = append(plan.fields, fieldPlan{index: i, op: opRecurse})
			continue
		case reflect.Ptr:
			// 结构体指针字段同样下钻
			//
			// 上游原实现对 *Struct 字段走 default 分支：无 tag 时解析失败被静默
			// 跳过、有 tag 时在 in.(string) 断言上 panic
			// 也就是说 `Billing *Address` 这种 Go 里极常见的形状从来没被脱敏过，
			// 嵌套在里面的姓名手机号原样留着——不报错、不可见，属于静默漏擦
			// 这里改为与值类型结构体一致地递归，tag 同样被忽略
			//
			// [EN] Upstream sent *Struct fields to the default branch: silently
			// skipped when untagged, panicking on the string assertion when
			// tagged. A shape as common as `Billing *Address` was therefore
			// never masked, leaving nested names and phone numbers in place
			// without any error — a silent miss. Now it recurses like a value
			// struct, with the tag ignored the same way
			if ft.Elem().Kind() == reflect.Struct {
				plan.fields = append(plan.fields, fieldPlan{index: i, op: opPtrStruct})
				continue
			}
		case reflect.Slice, reflect.Array:
			switch ft.Elem().Kind() {
			case reflect.Struct:
				plan.fields = append(plan.fields, fieldPlan{index: i, op: opStructs})
				continue
			case reflect.Ptr:
				// []*Item 与 []Item 同样常见，同样此前被整体跳过
				if ft.Elem().Elem().Kind() == reflect.Struct {
					plan.fields = append(plan.fields, fieldPlan{index: i, op: opPtrStructs})
					continue
				}
			}
		}

		// 剩余形态都要先解析出脱敏器，解析不出就整字段跳过
		d, err := resolveDesensitizer(state, tag)
		if err != nil {
			continue
		}
		// 脱敏器只接受字符串，目标不是字符串形态即为 tag 用法错误
		//
		// 原实现在运行时才撞上 in.(string) 断言失败而 panic，等于「字段打错 tag
		// 就崩掉进程」；而这件事只取决于类型，建计划时一次就能查出
		// 因此改为返回明确错误：既不崩溃，也绝不静默跳过——
		// 静默跳过会让开发者以为擦了、实际没擦，那是安全问题
		//
		// [EN] A desensitizer only accepts strings, so a non-string target is a
		// tag misuse. The original hit the failed type assertion at runtime and
		// panicked; this depends on the type alone and is caught once at plan
		// time. Reporting an error neither crashes nor silently skips — silently
		// skipping would leave the developer believing a field was masked
		op := opApply
		target := ft
		switch ft.Kind() {
		case reflect.Slice, reflect.Array:
			op, target = opElems, ft.Elem()
		case reflect.Map:
			op, target = opMap, ft.Elem()
		}
		if target.Kind() != reflect.String {
			plan.err = &TagMisuseError{
				Struct: t.String(),
				Field:  field.Name,
				Tag:    tag,
				Type:   ft.String(),
			}
			return plan
		}
		plan.fields = append(plan.fields, fieldPlan{index: i, op: op, d: d})
	}
	return plan
}

// ==== 结构体脱敏入口 | Struct entry point ====

// Desensitization 按结构体 tag 对对象原地脱敏
// obj 须为指向结构体的非 nil 指针
func Desensitization(obj interface{}) error {
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return errors.New("expected a non-nil pointer to a struct")
	}
	return desensitizeStruct(value.Elem(), loadRegistry())
}

// desensitizeStruct 按计划处理一个结构体值
func desensitizeStruct(sv reflect.Value, state *registryState) error {
	if sv.Kind() != reflect.Struct {
		// 与原实现一致：非结构体指针走到 NumField 会 panic，这里保持同样的失败形态
		return errors.New("expected a non-nil pointer to a struct")
	}

	plan := planFor(sv.Type(), state)
	if plan.err != nil {
		return plan.err
	}
	for _, fp := range plan.fields {
		fv := sv.Field(fp.index)
		switch fp.op {
		case opRecurse:
			if err := desensitizeStruct(fv, state); err != nil {
				return err
			}
		case opPtrStruct:
			// nil 指针跳过，不得解引用
			if fv.IsNil() {
				continue
			}
			if err := desensitizeStruct(fv.Elem(), state); err != nil {
				return err
			}
		case opStructs:
			for j := 0; j < fv.Len(); j++ {
				if err := desensitizeStruct(fv.Index(j), state); err != nil {
					return err
				}
			}
		case opPtrStructs:
			for j := 0; j < fv.Len(); j++ {
				elem := fv.Index(j)
				if elem.IsNil() {
					continue
				}
				if err := desensitizeStruct(elem.Elem(), state); err != nil {
					return err
				}
			}
		case opElems:
			for j := 0; j < fv.Len(); j++ {
				elem := fv.Index(j)
				elem.SetString(fp.d.Desensitize(elem.String()))
			}
		case opMap:
			for _, key := range fv.MapKeys() {
				masked := reflect.New(fv.Type().Elem()).Elem()
				masked.SetString(fp.d.Desensitize(fv.MapIndex(key).String()))
				fv.SetMapIndex(key, masked)
			}
		default:
			fv.SetString(fp.d.Desensitize(fv.String()))
		}
	}
	return nil
}

// ==== 内置脱敏器 | Built-in desensitizers ====

// builtinRegistry 构造只含内置脱敏器与工厂的初始快照
func builtinRegistry() *registryState {
	byName := map[string]Desensitizer{
		"email":        &DefaultDesensitizer{Email},
		"phoneNumber":  &DefaultDesensitizer{PhoneNumber},
		"phone":        &DefaultDesensitizer{PhoneNumber},
		"name":         &DefaultDesensitizer{ChineseName},
		"identityCard": &DefaultDesensitizer{IDCard},
		"idCard":       &DefaultDesensitizer{IDCard},
		"mobilePhone":  &DefaultDesensitizer{MobilePhone},
		"mobile":       &DefaultDesensitizer{MobilePhone},
		"address":      &DefaultDesensitizer{Address},
		"password":     &DefaultDesensitizer{Password},
		"carLicense":   &DefaultDesensitizer{CarLicense},
		"bankCard":     &DefaultDesensitizer{BankCard},
		"ipv4":         &DefaultDesensitizer{IPV4},
		"ipv6":         &DefaultDesensitizer{IPV6},
		"ip":           &DefaultDesensitizer{IPV4},
		// PEMKey 此前没有任何注册名：desensitize:"pemKey" 这个 tag 解析不到，
		// 既不报错也不脱敏，私钥与证书内容就这样原样落进输出
		// [EN] PEMKey had no registered name at all, so a desensitize:"pemKey"
		// tag resolved to nothing — neither failing nor masking — and private
		// keys and certificates went through untouched
		"pemKey":     &DefaultDesensitizer{PEMKey},
		"apiKey":     &DefaultDesensitizer{APIKey},
		"apikey":     &DefaultDesensitizer{APIKey},
		"api_key":    &DefaultDesensitizer{APIKey},
		"secret":     &DefaultDesensitizer{Secret},
		"secretKey":  &DefaultDesensitizer{Secret},
		"secret_key": &DefaultDesensitizer{Secret},
		"userId":     &DefaultDesensitizer{UserID},
		"userid":     &DefaultDesensitizer{UserID},
		"playerId":   &DefaultDesensitizer{PlayerID},
		"playerid":   &DefaultDesensitizer{PlayerID},
		"orderNo":    &DefaultDesensitizer{OrderNo},
		"orderno":    &DefaultDesensitizer{OrderNo},
		"uuid":       &DefaultDesensitizer{UUID},
		"openId":     &DefaultDesensitizer{OpenID},
		"openid":     &DefaultDesensitizer{OpenID},
		"unionId":    &DefaultDesensitizer{OpenID},
		"account":    &DefaultDesensitizer{Account},
	}
	// 跳步脱敏器工厂：标签格式为 jump(开始,结束,跳步)，如 desensitize:"jump(3,-4,1)"
	factories := map[string]DesensitizerFactory{
		"jump": NewJumpDesensitizer,
	}
	return &registryState{byName: byName, factories: factories, gen: 1}
}
