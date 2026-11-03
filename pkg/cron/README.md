# Cron 表达式解析器

高性能的 Cron 表达式解析库，支持标准 Cron 和 Quartz 风格的扩展语法

## 特性

- ✅ **标准 Cron 格式**：支持 5 字段（分 时 日 月 周）和 6 字段（秒 分 时 日 月 周）
- ✅ **Quartz 扩展语法**：支持 L、W、#、LW 等特殊字符
- ✅ **时区支持**：可指定时区解析和执行
- ✅ **描述符**：支持 @yearly、@monthly、@daily 等预定义表达式
- ✅ **命名字段**：支持月份（JAN-DEC）和星期（SUN-SAT）名称
- ✅ **高性能**：使用位运算优化，O(1) 时间复杂度匹配
- ✅ **类型安全**：使用泛型实现，编译时类型检查

## 安装

```bash
go get github.com/HitGameAI/apex-toolbox/pkg/cron
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/cron"
)

func main() {
    // 解析标准 Cron 表达式（5 字段）
    schedule, err := cron.ParseCronStandard("0 9 * * MON-FRI")
    if err != nil {
        panic(err)
    }
    
    // 获取下次执行时间
    now := time.Now()
    next := schedule.Next(now)
    fmt.Printf("下次执行时间: %s\n", next.Format("2006-01-02 15:04:05"))
}
```

### 带秒的 Cron 表达式

```go
// 解析 6 字段表达式（包含秒）
schedule, err := cron.ParseCronWithSeconds("*/5 * * * * ?")
if err != nil {
    panic(err)
}
```

## Cron 表达式格式

### 字段说明

#### 标准格式（5 字段）

```bash
分 时 日 月 周
```

| 字段 | 允许值 | 允许的特殊字符          | 备注                     |
|------|--------|-------------------------|-------------------------|
| 分   | 0-59   | * / , -                 | -                       |
| 时   | 0-23   | * / , -                 | -                       |
| 日   | 1-31   | * ? / , - L W C         | -                       |
| 月   | 1-12   | JAN-DEC * / , -         | -                       |
| 周   | 0-7    | SUN-SAT * ? / , - L #   | 0/7=周日，1=周一，2=周二，3=周三，4=周四，5=周五，6=周六 |

#### Quartz 格式（6 字段）

```bash
秒 分 时 日 月 周
```

| 字段 | 允许值 | 允许的特殊字符          | 备注                     |
|------|--------|-------------------------|-------------------------|
| 秒   | 0-59   | * / , -                 | -                       |
| 分   | 0-59   | * / , -                 | -                       |
| 时   | 0-23   | * / , -                 | -                       |
| 日   | 1-31   | * ? / , - L W C         | -                       |
| 月   | 1-12   | JAN-DEC * / , -         | -                       |
| 周   | 0-7    | SUN-SAT * ? / , - L #   | 0/7=周日，1=周一，2=周二，3=周三，4=周四，5=周五，6=周六 |

### 特殊字符说明

| 字符 | 含义                                         | 示例                                                                                                                   |
|------|----------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `*`  | 表示匹配该字段的任意值                        | 在分这个字段使用 `*`，即表示每分钟都会触发                                                                              |
| `?`  | 表示不关心该字段，但只能用在日和周字段         | 表示不指定值当2个子表达式其中之一被指定了值以后，为了避免冲突，需要将另一个子表达式的值设为 `?`                         |
| `-`  | 表示区间范围                                  | 在分这个字段使用 5-20，表示从 5 分到 20 分钟每分钟触发一次                                                               |
| `/`  | 表示起始时间开始触发，然后每隔固定时间触发一次 | 在分这个字段使用 5/20，表示从 5 分钟开始，每 20 分钟触发一次，即 5、25、45 等分别触发一次                                 |
| `,`  | 表示列出枚举值                                | 在分这个字段使用 5,20，则意味着在 5 和 20 分每分钟触发一次                                                              |
| `L`  | 表示最后，只能出在日和周字段                  | 在星期这个字段使用 5L，意味着在最后的一个星期四触发                                                                      |
| `W`  | 表示有效工作日(周一到周五)，只能出现在日字段   | 在日这个字段使用 5W，如果 5 号是星期六，则将在最近的工作日周五（4号）触发如果 5 号是周日，则在 6 号（周一）触发如果 5 号在星期一到星期五中的一天，则就在 5 号触发 |
| `LW` | 这两个字符可以连用，表示某个月最后一个工作日   | 在日这个字段使用 LW，想在某个月的最后一个工作日触发                                                                      |
| `#`  | 用于确定每个月第几个星期几                    | 在星期这个字段使用 4#2，表示某月的第二个星期三                                                                           |

## 常用示例

### 基础示例

```go
// 每隔 5 秒执行一次
"*/5 * * * * ?"

// 每隔 1 分钟执行一次
"0 */1 * * * ?"

// 每月 1 日的凌晨 2 点执行一次
"0 0 2 1 * ?"

// 周一到周五每天上午 10:15 执行
"0 15 10 ? * MON-FRI"

// 每天 23 点执行一次
"0 0 23 * * ?"

// 每天凌晨 1 点执行一次
"0 0 1 * * ?"

// 每月 1 日凌晨 1 点执行一次
"0 0 1 1 * ?"
```

### 特殊字符示例

```go
// 每月最后一天 23 点执行一次
"0 0 23 L * ?"

// 每周星期天凌晨 1 点执行一次
"0 0 1 ? * L"

// 每月的最后一个星期五上午 10:15 执行
"0 15 10 ? * 6L"

// 2002 年至 2006 年的每个月的最后一个星期五上午 10:15 执行
"0 15 10 ? * 6L 2002-2006"

// 每月的第三个星期五上午 10:15 执行
"0 15 10 ? * 6#3"

// 每月 15 日上午 10:15 触发
"0 15 10 15 * ?"

// 每月最后一日的上午 10:15 触发
"0 15 10 L * ?"
```

### 复杂时间段示例

```go
// 在 26 分、29 分、33 分执行一次
"0 26,29,33 * * * ?"

// 每天的 0 点、13 点、18 点、21 点都执行一次
"0 0 0,13,18,21 * * ?"

// 每天上午 10 点，下午 2 点，4 点执行一次
"0 0 10,14,16 * * ?"

// 朝九晚五工作时间内每半小时执行一次
"0 0/30 9-17 * * ?"

// 每个星期三中午 12 点执行一次
"0 0 12 ? * WED"

// 每天中午 12 点触发
"0 0 12 * * ?"

// 每天上午 10:15 触发
"0 15 10 ? * *"
"0 15 10 * * ?"

// 2005 年的每天上午 10:15 触发
"0 15 10 * * ? 2005"
```

### 分钟级精细控制

```go
// 每天下午 2 点到 2:59 期间的每 1 分钟触发
"0 * 14 * * ?"

// 每天下午 2 点到 2:55 期间的每 5 分钟触发
"0 0/5 14 * * ?"

// 每天下午 2 点到 2:55 和下午 6 点到 6:55 期间的每 5 分钟触发
"0 0/5 14,18 * * ?"

// 每天下午 2 点到 2:05 期间的每 1 分钟触发
"0 0-5 14 * * ?"

// 每年三月的星期三的下午 2:10 和 2:44 触发
"0 10,44 14 ? 3 WED"
```

## 预定义描述符

库提供了丰富的预定义描述符，可以直接使用：

### 标准描述符

```go
"@secondly"  // 等同于 "* * * * * *" - 每秒
"@minutely"  // 等同于 "0 * * * * *" - 每分钟
"@hourly"    // 等同于 "0 0 * * * *" - 每小时
"@daily"     // 等同于 "0 0 0 * * *" - 每天午夜
"@midnight"  // 同 @daily
"@weekly"    // 等同于 "0 0 0 * * 0" - 每周日午夜
"@monthly"   // 等同于 "0 0 0 1 * *" - 每月 1 号午夜
"@yearly"    // 等同于 "0 0 0 1 1 *" - 每年 1 月 1 日午夜
"@annually"  // 同 @yearly
```

### 工作日与周末

```go
"@workdays"       // 工作日(周一到周五)午夜
"@weekends"       // 周末(周六和周日)午夜
"@workdays_9am"   // 工作日早 9 点
"@workdays_6pm"   // 工作日晚 6 点
"@workday_start"  // 工作日开始(周一到周五早 9 点)
"@workday_end"    // 工作日结束(周一到周五晚 6 点)
"@weekend_10am"   // 周末早 10 点
"@weekend_morning" // 周末早晨(周六日早 10 点)
"@weekend_evening" // 周末晚上(周六日晚 8 点)
```

### 具体星期

```go
"@monday"    // 每周一午夜
"@tuesday"   // 每周二午夜
"@wednesday" // 每周三午夜
"@thursday"  // 每周四午夜
"@friday"    // 每周五午夜
"@saturday"  // 每周六午夜
"@sunday"    // 每周日午夜
```

### 每日特定时间

```go
"@daily_9am"  // 每天早 9 点
"@daily_12pm" // 每天中午 12 点
"@daily_6pm"  // 每天晚 6 点
"@daily_11pm" // 每天晚 11 点
```

### 时间段描述符

```go
"@business_hours" // 营业时间(9-17 点)
"@morning_hours"  // 早晨时段(6-12 点)
"@afternoon"      // 下午时段(12-18 点)
"@evening"        // 晚上时段(18-23 点)
"@peak_hours"     // 高峰期(8-9 点和 17-18 点)
"@off_peak"       // 低谷期(0-6 点和 22-23 点)
```

### 特殊时刻

```go
"@night"          // 深夜时段(凌晨 2 点)
"@dawn"           // 黎明时段(早晨 6 点)
"@noon"           // 中午 12 点
"@dusk"           // 黄昏时段(傍晚 6 点)
"@late_night"     // 深夜时段(23 点-凌晨 1 点)
"@early_morning"  // 清晨时段(早晨 5-7 点)
"@lunch_time"     // 午餐时间(11 点-13 点)
"@dinner_time"    // 晚餐时间(18 点-20 点)
```

### 月度相关

```go
"@first_of_month"  // 每月第 1 天午夜
"@mid_month"       // 每月 15 号午夜
"@monthly_report"  // 每月 1 号凌晨 2 点(报表生成)
"@month_end"       // 每月最后几天(28-31 日近似)
```

### 季度与年度

```go
"@quarterly"      // 每季度首日凌晨 2 点
"@quarter_start"  // 每季度开始(1/4/7/10 月的 1 号)
"@quarter_end"    // 每季度结束(3/6/9/12 月最后几天)
"@yearly_report"  // 每年 1 月 1 日凌晨 2 点
"@year_start"     // 每年开始(1 月 1 日午夜)
"@year_end"       // 每年结束(12 月 31 日午夜)
```

### 季节相关

```go
"@spring_start" // 春季开始(3 月 1 日)
"@summer_start" // 夏季开始(6 月 1 日)
"@autumn_start" // 秋季开始(9 月 1 日)
"@winter_start" // 冬季开始(12 月 1 日)
```

### 间隔执行

```go
"@every 5s"    // 每 5 秒（自定义间隔，支持 Go duration 格式）
"@every 1m"    // 每 1 分钟
"@every 1h"    // 每 1 小时
"@every 24h"   // 每 24 小时
"@every 1h30m" // 每 1 小时 30 分钟

// 预定义间隔别名
"@every_10s"   // 每 10 秒
"@every_30s"   // 每 30 秒
"@every_5min"  // 每 5 分钟
"@every_15min" // 每 15 分钟
"@every_30min" // 每 30 分钟
"@every_2h"    // 每 2 小时
"@every_3h"    // 每 3 小时
"@every_6h"    // 每 6 小时
"@every_12h"   // 每 12 小时
```

## API 文档

### 解析函数

```go
// 解析标准 Cron 表达式(5 字段)
func ParseCronStandard(spec string) (CronSchedule, error)

// 解析带秒的 Cron 表达式(6 字段)
func ParseCronWithSeconds(spec string) (CronSchedule, error)

// 创建自定义解析器
func NewCronParser(options CronParseOption) *CronParser

// 使用解析器解析表达式
func (p *CronParser) Parse(spec string) (CronSchedule, error)
```

### CronParseOption 解析选项

```go
type CronParseOption int

const (
    CronSecond        // 秒字段，默认 0
    CronSecondOptional // 可选秒字段，默认 0
    CronMinute        // 分钟字段，默认 0
    CronHour          // 小时字段，默认 0
    CronDom           // 月中日期字段，默认 *
    CronMonth         // 月份字段，默认 *
    CronDow           // 星期字段，默认 *
    CronDowOptional   // 可选星期字段，默认 *
    CronDescriptor    // 允许描述符，如 @monthly, @weekly 等
)
```

> 预定义解析器：`CronStandardParser`（5 字段）和 `CronSecondParser`（6 字段）已包含 `CronDescriptor` 选项注意不能同时配置多个可选字段（`CronSecondOptional` 和 `CronDowOptional`），否则会 panic

### CronSchedule 接口

```go
type CronSchedule interface {
    // Next 返回下次激活时间，晚于给定时间
    Next(t time.Time) time.Time
}
```

### CronSpecSchedule 结构

```go
type CronSpecSchedule struct {
    // 各字段的位掩码
    Second, Minute, Hour, Dom, Month, Dow uint64
    Location                              *time.Location
    // Quartz 特殊字符标记
    LastDay        bool // L - 月份最后一天
    LastWeekday    bool // LW - 月份最后一个工作日
    NearestWeekday int  // W - 最近的工作日(如 15W，存储 15)
    LastDow        int  // 星期字段的 L(如 6L，存储 6 表示最后一个星期五)
    NthDow         int  // 第几个星期 X(如 6#3，高位存储 6，低位存储 3)
}
```

> 还提供了 `NewZeroCronSpecSchedule(loc)` 构造函数和 `WithSecond/WithMinute/WithHour/WithDom/WithMonth/WithDow` 链式设置方法，用于描述符内部构建调度对象

### CronEverySchedule 结构

```go
// 间隔调度（@every duration）
type CronEverySchedule struct {
    Duration time.Duration
}
```

### 辅助函数

```go
// 获取表达式别名
func GetExpression(alias string) (string, bool)

// 解析字段为位集合（泛型，支持逗号分隔的列表）
func ParseFieldToBits[T types.Numerical](field string, bounds types.Bounds[T], starBit uint64) (uint64, error)

// 解析单个表达式（支持 number | start-end[/step] | * | ?）
func ParseExprToBits[T types.Numerical](expr string, bounds types.Bounds[T], starBit uint64) (uint64, error)

// 解析可能包含特殊字符(L/W/#)的字段
// 返回值：bits, lastDay, lastWeekday, nearestWeekday, lastDow, nthDow, error
func ParseFieldWithSpecialChars(fieldStr string, bounds cronBounds, starBit uint64, isDOM bool, isDOW bool) (
    bits uint64,
    lastDay bool,
    lastWeekday bool,
    nearestWeekday int,
    lastDow int,
    nthDow int,
    err error,
)

// 解析整数或命名值（泛型，names 为命名映射表）
func ParseIntOrName[T types.Numerical](expr string, names map[string]T) (T, error)
```

## 预定义表达式常量

库提供了常用 Cron 表达式的字符串常量，可直接使用：

### 基础周期

```go
EverySecond    = "* * * * * *"    // 每秒
EveryMinute    = "0 * * * * *"    // 每分钟
Every5Minutes  = "0 */5 * * * *"  // 每5分钟
Every15Minutes = "0 */15 * * * *" // 每15分钟
Every30Minutes = "0 */30 * * * *" // 每30分钟
EveryHour      = "0 0 * * * *"    // 每小时
Every2Hours    = "0 0 */2 * * *"  // 每2小时
Every3Hours    = "0 0 */3 * * *"  // 每3小时
Every6Hours    = "0 0 */6 * * *"  // 每6小时
Every12Hours   = "0 0 */12 * * *" // 每12小时
EveryDay       = "0 0 0 * * *"    // 每天午夜
EveryWeek      = "0 0 0 * * 0"    // 每周日午夜
EveryMonth     = "0 0 0 1 * *"    // 每月1号午夜
EveryYear      = "0 0 0 1 1 *"    // 每年1月1日午夜
```

### 工作日与具体星期

```go
WorkdaysOnly  = "0 0 0 * * 1-5"  // 工作日午夜
WeekendsOnly  = "0 0 0 * * 0,6"  // 周末午夜
WorkdaysAt9AM = "0 0 9 * * 1-5"  // 工作日早9点
WorkdaysAt6PM = "0 0 18 * * 1-5" // 工作日晚6点
Monday    = "0 0 0 * * 1" // 每周一
Tuesday   = "0 0 0 * * 2" // 每周二
Wednesday = "0 0 0 * * 3" // 每周三
Thursday  = "0 0 0 * * 4" // 每周四
Friday    = "0 0 0 * * 5" // 每周五
Saturday  = "0 0 0 * * 6" // 每周六
Sunday    = "0 0 0 * * 0" // 每周日
```

### 特定时段与月度

```go
BusinessHours  = "0 0 9-17 * * *"      // 营业时间
MorningHours   = "0 0 6-12 * * *"      // 早晨时段
AfternoonHours = "0 0 12-18 * * *"     // 下午时段
EveningHours   = "0 0 18-23 * * *"     // 晚上时段
PeakHours      = "0 0 8-9,17-18 * * *" // 高峰期
OffPeakHours   = "0 0 0-6,22-23 * * *" // 低谷期
FirstDayOfMonth = "0 0 0 1 * *"  // 每月第1天
LastDayOfMonth  = "0 0 0 L * *"  // 每月最后一天
MidMonth        = "0 0 0 15 * *" // 每月15号
QuarterlyReport = "0 0 2 1 */3 *" // 每季度首日凌晨2点
YearlyReport    = "0 0 2 1 1 *"   // 每年1月1日凌晨2点
```

## 使用示例

### 示例 1：基本调度

```go
package main

import (
    "fmt"
    "time"
    "github.com/HitGameAI/apex-toolbox/pkg/cron"
)

func main() {
    // 每天上午 10:15 执行
    schedule, err := cron.ParseCronWithSeconds("0 15 10 * * ?")
    if err != nil {
        panic(err)
    }
    
    now := time.Now()
    for i := 0; i < 5; i++ {
        next := schedule.Next(now)
        fmt.Printf("第 %d 次执行: %s\n", i+1, next.Format("2006-01-02 15:04:05"))
        now = next
    }
}
```

### 示例 2：使用描述符

```go
// 使用预定义描述符
schedule, _ := cron.ParseCronStandard("@daily")

// 使用自定义描述符
schedule, _ := cron.ParseCronStandard("@workdays_9am")

// 使用间隔描述符
schedule, _ := cron.ParseCronStandard("@every 5m")
```

### 示例 3：时区支持

```go
// 设置时区
location, _ := time.LoadLocation("Asia/Shanghai")
parser := cron.NewCronParser(
    cron.CronSecond | cron.CronMinute | cron.CronHour | 
    cron.CronDom | cron.CronMonth | cron.CronDow,
)

// 带时区的表达式
schedule, _ := parser.Parse("CRON_TZ=Asia/Shanghai 0 15 10 * * ?")
```

### 示例 4：Quartz 特殊字符

```go
// 每月最后一天
schedule, _ := cron.ParseCronWithSeconds("0 0 0 L * ?")

// 每月最后一个星期五
schedule, _ := cron.ParseCronWithSeconds("0 15 10 ? * 6L")

// 每月第三个星期五
schedule, _ := cron.ParseCronWithSeconds("0 15 10 ? * 6#3")

// 每月 15 号最近的工作日
schedule, _ := cron.ParseCronWithSeconds("0 0 12 15W * ?")

// 每月最后一个工作日
schedule, _ := cron.ParseCronWithSeconds("0 0 12 LW * ?")
```

### 示例 5：自定义解析器

```go
// 创建只支持特定字段的解析器
parser := cron.NewCronParser(
    cron.CronMinute | cron.CronHour | cron.CronDow,
)

// 只需要 3 个字段：分 时 周
schedule, err := parser.Parse("15 10 MON-FRI")
```

## 性能优化

### 位运算优化

本库使用位集合(bit set)来表示时间字段，具有以下优势：

- **O(1) 匹配**：检查某个值是否匹配只需要一次位操作
- **内存高效**：一个 uint64 可以表示 64 个不同的值
- **快速计算**：利用 CPU 位运算指令，比循环快得多

```go
// 检查某个值是否匹配
func (s *CronSpecSchedule) matches(value uint, bits uint64) bool {
    return bits&(1<<value) != 0
}
```

### 预计算位掩码

常用的时间范围被预先计算并缓存：

```go
var (
    cronStarBit      = uint64(1) << 63          // 星号标记位
    cronAllWeekdays  = mathx.GetBit64(0, 6, 1)  // 周日到周六：0-6
    cronAllMonths    = mathx.GetBit64(1, 12, 1) // 1月到12月：1-12
    cronAllHours     = mathx.GetBit64(0, 23, 1) // 0-23小时
    cronAllDaysOfMon = mathx.GetBit64(1, 31, 1) // 1-31日
)
```

## 测试

运行测试：

```bash
cd pkg/cron
go test -v
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 限制和注意事项

1. **年份字段**：目前年份字段(2002-2006)被标记为 TODO，暂不支持
2. **DOW 值 7 兼容**：✅ 已支持！在星期字段中，7 自动转换为 0(周日)，完全兼容 Quartz 标准
3. **负值**：所有字段的值不能为负，负值或零步长会被拒绝并返回错误
4. **L 在列表中**：特殊字符 L、W、# 不能出现在逗号分隔的列表中

### DOW=7 的星期日兼容性

为了与 Quartz Cron 完全兼容，本库支持使用 `7` 表示星期日：

```go
// 以下两种写法等价，都表示每周日执行
"0 0 0 * * 0"  // 标准写法：0 表示周日
"0 0 0 * * 7"  // Quartz 兼容：7 自动转换为 0

// 在范围表达式中也支持
"0 0 0 * * 1-7"  // 等同于 1-6,0 (周一到周日)

// 错误示例：大于 7 的值会被拒绝
"0 0 0 * * 8"   // ❌ 错误：星期值超出范围 [0-7]
```

## 兼容性

- ✅ 兼容标准 Unix Cron 格式(5 字段)
- ✅ 兼容 Quartz Scheduler 格式(6 字段)
- ✅ 兼容 Spring @Scheduled Cron 表达式
- ✅ 部分兼容 robfig/cron/v3(可作为替代品)

## 许可证

Copyright (c) 2025 by kamalyes, All Rights Reserved.

## 参考资料

- [Quartz Scheduler Cron 表达式](http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html)
- [Cron 维基百科](https://en.wikipedia.org/wiki/Cron)
- [Spring @Scheduled](https://docs.spring.io/spring-framework/docs/current/reference/html/integration.html#scheduling-cron-expression)
