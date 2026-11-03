---
name: netx-network
description: 网络工具包，提供本地IP获取、公网IP获取、客户端IP提取，当需要获取本机/私有/公网IP地址、或从HTTP请求提取客户端IP时使用
---

# netx - 网络工具

提供本地IP、私有IP、公网IP获取与HTTP客户端IP提取

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/netx"
```

获取IP：
```go
ips, err := netx.GetLocalInterfaceIPs()              // []net.IP
privateIP, err := netx.GetPrivateIP()                // 私有IP
publicIP, err := netx.GetConNetPublicIP()            // 公网IP
privateIP, publicIP, err := netx.GetLocalInterfaceIPAndExternalIP()
```

从请求获取客户端IP：
```go
clientIP := netx.GetClientIP(r)
```

## 完整API索引

### 函数（ip.go）

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `GetLocalInterfaceIPs` | `func() ([]net.IP, error)` | 查询本机网卡所有IP（排除回环地址） |
| `GetPrivateIP` | `func() (string, error)` | 获取私有IP地址（未找到返回错误） |
| `GetLocalInterfaceIPAndExternalIP` | `func(urls ...string) (privateIP string, publicIP string, err error)` | 同时获取本地私有IP和公网IP |
| `GetConNetPublicIP` | `func(urls ...string) (string, error)` | 联网获取公网IP（默认 `http://myexternalip.com/raw`） |
| `GetClientIP` | `func(r *http.Request) string` | 从HTTP请求提取客户端IP |

## 注意事项

- `GetLocalInterfaceIPs` 返回 `[]net.IP` 而非字符串切片，且会排除回环地址
- `GetPrivateIP` 遍历 `net.Interfaces()`，仅返回 `IsPrivate()` 的IP，未找到时返回 `"未找到私有 IP"` 错误
- `GetConNetPublicIP` 需要外部网络访问，HTTP超时3秒；可传入自定义URL作为首个参数
- `GetLocalInterfaceIPAndExternalIP` 返回三个值（私有IP、公网IP、错误），任一步骤失败立即返回
- `GetClientIP` 依次检查：X-Forwarded-For（取首个IP）→ X-Real-IP → RemoteAddr（`net.SplitHostPort` 去端口）
