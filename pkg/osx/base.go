/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2023-07-28 00:50:58
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-12-04 19:31:22
 * @FilePath: \go-toolbox\pkg\osx\base.go
 * @Description: 操作系统相关工具函数
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package osx

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/random"
	"github.com/HitGameAI/apex-toolbox/pkg/stringx"
)

// WorkerIDConfig WorkerID 配置
type WorkerIDConfig struct {
	maxWorkerID     atomic.Int64 // WorkerID 最大值（默认 1024）
	maxDatacenterID atomic.Int64 // DatacenterID 最大值（默认 32）
	maxSnowflakeID  atomic.Int64 // 雪花算法 ID 最大值（默认 32）
}

// RunTimeCaller 运行时调用栈信息
type RunTimeCaller struct {
	Pc       uintptr // 程序计数器
	File     string  // 文件路径
	Line     int     // 行号
	FuncName string  // 函数名
}

var (
	globalWorkerIDConfig = &WorkerIDConfig{}
	callerPool           = sync.Pool{
		New: func() any {
			return &RunTimeCaller{}
		},
	}
)

func init() {
	// 设置默认值
	globalWorkerIDConfig.maxWorkerID.Store(1024)
	globalWorkerIDConfig.maxDatacenterID.Store(32)
	globalWorkerIDConfig.maxSnowflakeID.Store(32)
}

// SetMaxWorkerID 设置 WorkerID 的最大值（范围限制）
// 默认值: 1024
func SetMaxWorkerID(max int64) {
	globalWorkerIDConfig.maxWorkerID.Store(mathx.IfLeZero(max, 1024))
}

// SetMaxDatacenterID 设置 DatacenterID 的最大值（范围限制）
// 默认值: 32（雪花算法标准）
func SetMaxDatacenterID(max int64) {
	globalWorkerIDConfig.maxDatacenterID.Store(mathx.IfLeZero(max, 32))
}

// SetMaxSnowflakeWorkerID 设置雪花算法 WorkerID 的最大值（范围限制）
// 默认值: 32（雪花算法标准的5位）
func SetMaxSnowflakeWorkerID(max int64) {
	globalWorkerIDConfig.maxSnowflakeID.Store(mathx.IfLeZero(max, 32))
}

// GetMaxWorkerID 获取 WorkerID 的最大值
func GetMaxWorkerID() int64 {
	return globalWorkerIDConfig.maxWorkerID.Load()
}

// GetMaxDatacenterID 获取 DatacenterID 的最大值
func GetMaxDatacenterID() int64 {
	return globalWorkerIDConfig.maxDatacenterID.Load()
}

// GetMaxSnowflakeWorkerID 获取雪花算法 WorkerID 的最大值
func GetMaxSnowflakeWorkerID() int64 {
	return globalWorkerIDConfig.maxSnowflakeID.Load()
}

// 主机名缓存
//
// os.Hostname 是系统调用，实测约 730 纳秒、3 次分配而主机名是内核属性，
// 进程生命周期内不会变（真要改也得重启服务），每次调用都去问一遍没有意义
// SafeGetHostName 还要在此之上做一遍字符替换，结果同样是恒定的，一并缓存
//
// 只缓存成功的结果：首次失败可能是暂时的，不该把失败固化下来
// 用 atomic.Pointer 而不是 sync.Once，这样失败后下次调用仍会重试
//
// 注意这里缓存的是主机名，不是 GetWorkerId / GetDatacenterId 的结果——
// 那两个的契约是按优先级读环境变量，调用方可以在运行期改环境变量来影响它们，
// 缓存会改变语义它们实际也只在生成器构造时调一次
//
// [EN] Hostname cache. os.Hostname is a syscall, measured at about 730
// nanoseconds and three allocations, while the hostname is a kernel property
// that does not change during a process's life. SafeGetHostName's character
// replacement on top of it is equally constant, so both are cached. Only
// successful results are cached — a first failure may be transient and should
// not be frozen in — which is why this uses atomic.Pointer rather than
// sync.Once, leaving a later call free to retry. Note it is the hostname that is
// cached, not the results of GetWorkerId or GetDatacenterId: those are
// contracted to read environment variables by priority, so caching them would
// change their semantics
var (
	cachedHostName     atomic.Pointer[string]
	cachedSafeHostName atomic.Pointer[string]
)

// GetHostName 获取主机名，如果失败则返回错误
//
// 结果在进程内缓存，只在首次成功时读取一次系统调用
func GetHostName() (string, error) {
	if p := cachedHostName.Load(); p != nil {
		return *p, nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("无法获取主机名: %v", err)
	}
	cachedHostName.CompareAndSwap(nil, &hostname)
	return hostname, nil
}

// SafeGetHostName 安全获取主机名，失败时返回随机字符串
//
// 成功的结果在进程内缓存失败时每次都会新生成随机串，不缓存——否则一次暂时
// 的失败会把这个随机串固化成进程的身份
func SafeGetHostName() string {
	if p := cachedSafeHostName.Load(); p != nil {
		return *p
	}
	output, err := GetHostName()
	if err != nil || output == "" {
		// 如果获取主机名失败或返回空字符串，则生成随机字符串
		return stringx.ReplaceSpecialChars(random.FRandAlphaString(8), 'x')
	}
	safe := stringx.ReplaceSpecialChars(output, 'x')
	cachedSafeHostName.CompareAndSwap(nil, &safe)
	return safe
}

// HashUnixMicroCipherText 生成基于时间戳和主机名的哈希密文
func HashUnixMicroCipherText() string {
	var (
		nowUnixMicro = time.Now().UnixMicro()
		hostName     = SafeGetHostName()
		randStr      = random.RandString(10, 4)
		plainText    = fmt.Sprintf("%s%s%d", hostName, randStr, nowUnixMicro)
		cipherText   = stringx.CalculateMD5Hash(plainText)
	)
	return cipherText
}

// GetWorkerId 获取唯一的 Worker ID (智能增强版)
// 优先级: K8s Pod Name > K8s Hostname > 环境变量 > 主机名哈希
// 支持环境变量: POD_NAME, HOSTNAME, WORKER_ID, NODE_ID, POD_ORDINAL
// 返回范围: 0 到 (maxWorkerID-1)，默认 0-1023
func GetWorkerId() int64 {
	maxWorkerID := globalWorkerIDConfig.maxWorkerID.Load()

	// 1. 优先从 K8s Pod Name 提取序号（如 myapp-0, myapp-1）
	if podName := os.Getenv("POD_NAME"); podName != "" {
		if id := extractOrdinalFromName(podName); id >= 0 {
			return id % maxWorkerID
		}
	}

	// 2. 从 K8s Hostname 提取序号
	if hostname := os.Getenv("HOSTNAME"); hostname != "" {
		if id := extractOrdinalFromName(hostname); id >= 0 {
			return id % maxWorkerID
		}
	}

	// 3. 尝试从环境变量获取
	envVars := []string{"WORKER_ID", "NODE_ID", "POD_ORDINAL"}
	for _, envVar := range envVars {
		if workerID := os.Getenv(envVar); workerID != "" {
			if id, err := mathx.ParseInt64(workerID); err == nil {
				return mathx.Abs(id) % maxWorkerID
			}
		}
	}

	// 4. 从主机名生成唯一的 Worker ID
	hostName := SafeGetHostName()
	hash := sha256.Sum256([]byte(hostName))
	hostNameHash := int64(binary.BigEndian.Uint64(hash[:8]))

	return mathx.Abs(hostNameHash) % maxWorkerID
}

// extractOrdinalFromName 从名称中提取序号
// 支持格式: myapp-0, myapp-1, myapp-statefulset-0 等
// 返回 -1 表示未找到序号
func extractOrdinalFromName(name string) int64 {
	// 从后往前查找最后一个 '-' 后的数字
	lastDash := strings.LastIndex(name, "-")
	if lastDash == -1 || lastDash == len(name)-1 {
		return -1
	}

	ordinalStr := name[lastDash+1:]
	if id, err := mathx.ParseInt64(ordinalStr); err == nil {
		return id
	}

	return -1
}

// GetDatacenterId 获取数据中心ID
// 优先级: 自定义环境变量 > 容器编排平台 > 数据中心标识 > 默认值1
// 注意：不使用 REGION/ZONE 等区域变量，因为同一区域内的多台机器值相同，无法区分节点
// 返回范围: 0 到 (maxDatacenterID-1)，默认 0-31（雪花算法标准）
func GetDatacenterId() int64 {
	maxDatacenterID := globalWorkerIDConfig.maxDatacenterID.Load()

	// 1. 优先使用自定义环境变量（数字类型，直接使用）
	customEnvVars := []string{"DATACENTER_ID", "DC_ID", "DATA_CENTER_ID"}
	for _, envVar := range customEnvVars {
		if dcID := os.Getenv(envVar); dcID != "" {
			if id, err := mathx.ParseInt64(dcID); err == nil {
				return mathx.Abs(id) % maxDatacenterID
			}
		}
	}

	// 2. 从容器编排平台环境变量推导（这些能区分不同节点/环境）
	orchestrationEnvVars := []string{
		// Kubernetes - 命名空间和集群名可以区分不同环境
		"KUBERNETES_NAMESPACE", "K8S_NAMESPACE", "POD_NAMESPACE",
		"KUBERNETES_CLUSTER_NAME", "K8S_CLUSTER_NAME",

		// Docker Swarm - 节点 ID 可以区分不同节点
		"DOCKER_SWARM_NODE_ID", "SWARM_NODE_ID",

		// Nomad - DC 和命名空间可以区分
		"NOMAD_DC", "NOMAD_NAMESPACE",

		// Mesos - 任务 ID 可以区分
		"MESOS_TASK_ID", "MARATHON_APP_ID",

		// OpenShift - 命名空间可以区分
		"OPENSHIFT_BUILD_NAMESPACE", "OPENSHIFT_DEPLOYMENT_NAMESPACE",
	}
	for _, envVar := range orchestrationEnvVars {
		if value := os.Getenv(envVar); value != "" {
			hash := sha256.Sum256([]byte(value))
			return int64(binary.BigEndian.Uint32(hash[:4])) % maxDatacenterID
		}
	}

	// 3. 从明确的数据中心/集群标识推导
	datacenterEnvVars := []string{
		// 数据中心/机房标识
		"DATACENTER", "DC", "IDC", "SITE_ID",

		// 集群标识
		"CLUSTER_ID",
	}
	for _, envVar := range datacenterEnvVars {
		if value := os.Getenv(envVar); value != "" {
			hash := sha256.Sum256([]byte(value))
			return int64(binary.BigEndian.Uint32(hash[:4])) % maxDatacenterID
		}
	}

	// 4. 默认返回 1
	return 1
}

// GetWorkerIdForSnowflake 获取适合雪花算法的 WorkerID（0-31，5位）
func GetWorkerIdForSnowflake() int64 {
	maxSnowflakeID := globalWorkerIDConfig.maxSnowflakeID.Load()
	return GetWorkerId() % maxSnowflakeID
}

// GetServerNode 获取当前服务节点标识（K8s 环境下为 Pod 名称）
// 用于 HTTP 响应头 X-Server-Node 和 gRPC metadata x-server-node 透传，
// 便于在分布式环境下定位处理请求的具体节点
//
// 优先级: POD_NAME 环境变量 > HOSTNAME 环境变量 > os.Hostname() > 随机生成的标识
//
// 与 GetWorkerId 的区别：
//   - GetWorkerId 返回数字 ID（提取序号或哈希取模），用于雪花算法
//   - GetServerNode 返回原始节点名称（保留 Pod 名称原文），用于链路追踪和流量定位
func GetServerNode() string {
	// 1. K8s Downward API 注入的 POD_NAME（最准确）
	if node := os.Getenv("POD_NAME"); node != "" {
		return node
	}
	// 2. K8s 默认注入的 HOSTNAME（通常等于 Pod 名称）
	if node := os.Getenv("HOSTNAME"); node != "" {
		return node
	}
	// 3. 回退到系统主机名（K8s 环境下与 Pod 名称一致）
	// 走 GetHostName 而不是直呼 os.Hostname，以便用上主机名缓存
	// 上面两步读环境变量，按契约不缓存
	if host, err := GetHostName(); err == nil && host != "" {
		return host
	}
	// 4. 兜底：基于时间戳和主机名哈希生成短标识，避免空值
	return "node-" + HashUnixMicroCipherText()[:8]
}

// StableHashSlot 根据输入字符串 s 和范围 [minNum, maxNum]，返回一个稳定且范围内的整数
// 使用加密哈希 sha256，抗碰撞更强
// 如果 maxNum < minNum，会 panic
// 如果 maxNum == minNum，直接返回 minNum
func StableHashSlot(s string, minNum, maxNum int) int {
	if maxNum < minNum {
		panic("maxNum must be >= minNum")
	}
	if maxNum == minNum {
		return minNum
	}

	var (
		hashBytes = sha256.Sum256([]byte(s))
		// 取哈希结果的前8字节转为uint64
		hashVal   = binary.BigEndian.Uint64(hashBytes[:8])
		rangeSize = maxNum - minNum + 1
		result    = int(hashVal%uint64(rangeSize)) + minNum
	)
	// 防御 int 转换溢出导致负值，确保结果至少为 1
	return mathx.IfLeZero(result, 1)
}

// GetRuntimeCaller 获取调用栈信息，调用者使用完需调用 Release() 归还对象
func GetRuntimeCaller(skip int) *RunTimeCaller {
	caller := callerPool.Get().(*RunTimeCaller)
	caller.init(skip)
	return caller
}

// init 初始化 RunTimeCaller 内容
//
// 用 runtime.Callers 而不是 runtime.Caller
//
// runtime.Caller 内部要走 CallersFrames，实测每次 248 字节、2 次分配，而对象池
// 省下的那一次分配只有几十字节——分配其实全在这里改成把程序计数器取进栈上数组，
// 再用 FuncForPC.FileLine 解析，实测 187ns / 248B / 2 次分配降到 112ns / 零分配
//
// skip 的换算：runtime.Caller(s) 等价于 runtime.Callers(s+1, ...)，因为
// runtime.Callers 的第 0 帧是它自己这个差一错不会报错，只会把调用点指偏，
// 所以 TestRuntimeCallerSkipMapping 钉住了它
//
// 返回地址要回退一个字节才落在调用指令上，这与 runtime.Caller 内部的处理一致
// 已验证两者在文件、行号、函数名上完全一致，包括经过被内联的包装函数和闭包
//
// [EN] Uses runtime.Callers rather than runtime.Caller. runtime.Caller goes
// through CallersFrames, measured at 248 bytes and two allocations per call —
// which is where the allocations actually were, dwarfing the few dozen bytes the
// object pool saves. Taking the program counter into a stack array and resolving
// it with FuncForPC.FileLine instead measures 112ns and no allocations, down
// from 187ns. On the skip arithmetic: runtime.Caller(s) is equivalent to
// runtime.Callers(s+1, ...) because frame 0 of runtime.Callers is itself. That
// off-by-one raises no error and merely misreports the call site, so
// TestRuntimeCallerSkipMapping pins it. The return address is rolled back one
// byte to land on the call instruction, as runtime.Caller does internally. File,
// line and function name were verified identical between the two, including
// through inlined wrappers and closures
func (c *RunTimeCaller) init(skip int) {
	var pcs [1]uintptr
	if runtime.Callers(skip+1, pcs[:]) < 1 {
		c.Pc = 0
		c.File = "unknown_file"
		c.Line = 0
		c.FuncName = "unknown_func"
		return
	}
	pc := pcs[0] - 1

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		c.Pc = pc
		c.File = "unknown_file"
		c.Line = 0
		c.FuncName = "unknown_func"
		return
	}
	c.Pc = pc
	c.File, c.Line = fn.FileLine(pc)

	fullName := fn.Name()
	// 只保留函数名（去掉包路径）
	if lastSlash := strings.LastIndex(fullName, "/"); lastSlash != -1 {
		fullName = fullName[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
		fullName = fullName[lastDot+1:]
	}
	c.FuncName = fullName
}

// Release 放回对象池，调用者必须调用
func (c *RunTimeCaller) Release() {
	// 清理字段，防止内存泄漏
	c.Pc = 0
	c.File = ""
	c.Line = 0
	c.FuncName = ""
	callerPool.Put(c)
}

// String 返回调用栈的格式化字符串，包含函数名、文件名和行号
func (c *RunTimeCaller) String() string {
	return fmt.Sprintf("FuncName:%s, File:%s:%d", c.FuncName, c.File, c.Line)
}

// Command 执行系统命令
func Command(bin string, argv []string, baseDir string) ([]byte, error) {
	cmd := exec.Command(bin, argv...)
	cmd.Dir = mathx.IF(baseDir != "", baseDir, "")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.Bytes(), fmt.Errorf("command failed: %s: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
