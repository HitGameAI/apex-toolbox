---
name: stringx-string-operations
description: 字符串操作链式工具，提供链式调用、子串提取、隐藏脱敏、填充对齐、显示宽度计算、前后缀判断、替换、分割、命名风格转换、快速数字/时间格式化、JSON 字符串引用、域名处理、字段解析等当需要对字符串做链式变换、截取、脱敏、格式化对齐、命名转换或构造 JSON 字段名时使用
---

# stringx - 字符串操作链式工具

提供链式字符串变换、子串提取、脱敏隐藏、填充对齐、显示宽度计算、前后缀匹配、替换、分割、命名风格转换、快速格式化、JSON 字符串引用、域名处理与字段解析

## 快速开始

```go
import "github.com/HitGameAI/apex-toolbox/pkg/stringx"
```

链式调用：
```go
result := stringx.New("hello world").ToUpperChain().ReplaceAllChain("WORLD", "GO").String()
```

子串提取与隐藏：
```go
before := stringx.SubBefore("user@host", "@", false)
hidden := stringx.Hide("13812345678", 3, 7)
```

## 完整API索引

### 函数

#### 构造与基础

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `New` | `func(value string) *StringX` | 创建 StringX 链式对象（返回指针） |

#### 大小写转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ToLower` | `func(str string) string` | 转小写（委托 strings.ToLower） |
| `ToUpper` | `func(str string) string` | 转大写（委托 strings.ToUpper） |
| `ToTitle` | `func(str string) string` | 转标题格式（每个单词首字母大写） |

#### 命名风格转换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ToPascalCase` | `func(s string) string` | 转帕斯卡命名法（UserName） |
| `ToCamelCase` | `func(s string) string` | 转驼峰命名法（userName） |
| `ToSnakeCase` | `func(s string) string` | 转蛇形命名法（user_name） |
| `ToKebabCase` | `func(s string) string` | 转短横线命名法（user-name） |
| `ConvertCharacterStyle` | `func(input string, caseType CharacterStyle) string` | 按 CharacterStyle 转换命名风格 |
| `NormalizeFieldName` | `func(fieldName string) []string` | 返回所有命名风格变体切片 |

#### 修剪操作

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Trim` | `func(str string) string` | 去除两端空白 |
| `EqualsTrimIgnoreCase` | `func(str1, str2 string) bool` | 去除首尾空白后忽略大小写比较 |
| `TrimStart` | `func(str string) string` | 去除前导空白（空格/制表符/换行） |
| `TrimEnd` | `func(str string) string` | 去除尾部空白 |
| `CleanEmpty` | `func(str string) string` | 清除所有空格字符 |
| `TrimProtocol` | `func(url string) string` | 去除协议前缀（http/https/ftp/ws/wss/file 等） |
| `TrimAll` | `func(str, strToRemove string) string` | 去除所有指定子串 |
| `TrimAny` | `func(str string, strsToRemove []string) string` | 去除多个指定子串 |
| `TrimAllLineBreaks` | `func(str string) string` | 去除所有换行符（\r \n） |
| `TrimNewlines` | `func(str string) string` | 去除首尾换行符 |
| `TrimStartNewlines` | `func(str string) string` | 去除开头换行符 |
| `TrimEndNewlines` | `func(str string) string` | 去除结尾换行符 |
| `TrimPrefix` | `func(str, prefix string) string` | 去除指定前缀 |
| `TrimPrefixIgnoreCase` | `func(str, prefix string) string` | 去除指定前缀（忽略大小写） |
| `TrimSuffix` | `func(str, suffix string) string` | 去除指定后缀 |
| `TrimSuffixIgnoreCase` | `func(str, suffix string) string` | 去除指定后缀（忽略大小写） |
| `TrimSymbols` | `func(str string) string` | 去除所有符号（正则 `[^\w]+`） |

#### 子串提取

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `SubBefore` | `func(str, separator string, isLastSeparator bool) string` | 提取分隔符前的子串 |
| `SubAfter` | `func(str, separator string, isLastSeparator bool) string` | 提取分隔符后的子串 |
| `SubBetween` | `func(str, before, after string) string` | 提取两标记之间的子串 |
| `SubBetweenAll` | `func(str, prefix, suffix string) []string` | 提取所有两标记之间的子串 |
| `SubString` | `func(s string, start, length int) string` | 按位置截取子串 |

#### 前后缀判断

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `StartWith` | `func(str, prefix string) bool` | 判断前缀 |
| `StartWithIgnoreCase` | `func(str, prefix string) bool` | 判断前缀（忽略大小写） |
| `StartWithAny` | `func(str string, prefixes []string) bool` | 是否以任一前缀开头 |
| `StartWithAnyIgnoreCase` | `func(str string, prefixes []string) bool` | 是否以任一前缀开头（忽略大小写） |
| `EndWith` | `func(str, suffix string) bool` | 判断后缀 |
| `EndWithIgnoreCase` | `func(str, suffix string) bool` | 判断后缀（忽略大小写） |
| `EndWithAny` | `func(str string, suffixes []string) bool` | 是否以任一后缀结尾 |
| `EndWithAnyIgnoreCase` | `func(str string, suffixes []string) bool` | 是否以任一后缀结尾（忽略大小写） |

#### 包含判断

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Contains` | `func(value, searchStr string) bool` | 是否包含子串 |
| `ContainsIgnoreCase` | `func(value, searchStr string) bool` | 是否包含子串（忽略大小写） |
| `ContainsAny` | `func(value string, searchStrs []string) bool` | 是否包含任一子串 |
| `ContainsAnyIgnoreCase` | `func(str string, searchStrs []string) bool` | 是否包含任一子串（忽略大小写） |
| `ContainsAll` | `func(str string, searchStrs []string) bool` | 是否包含所有子串 |
| `ContainsBlank` | `func(str string) bool` | 是否包含空白符（含全角空格/不间断空格） |
| `GetContainsStr` | `func(str string, searchStrs []string) string` | 返回第一个包含的子串 |

#### 替换

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Replace` | `func(source, searchStr, replacement string, replaceCount int) string` | 替换前 n 个匹配 |
| `ReplaceAll` | `func(source, searchStr, replacement string) string` | 替换所有匹配 |
| `ReplaceWithIndex` | `func(str string, startIndex, endIndex int, replacedStr string) string` | 按索引范围替换 |
| `ReplaceWithMatcher` | `func(str, regex string, replaceFun func(string) string) string` | 按正则表达式替换 |
| `ReplaceSpecialChars` | `func(str string, replaceValue rune) string` | 替换特殊字符（中英文标点）为指定 rune |
| `Hide` | `func(str string, startInclude, endExclude int) string` | 隐藏指定区间字符为 `*`（脱敏） |

#### 填充与截断

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Pad` | `func(input string, minLength int, paddler ...*Paddler) string` | 填充到指定长度（用 `*` 填充，可选位置） |
| `Fill` | `func(str, char string, length int, isPre bool) string` | 填充到指定长度 |
| `FillBefore` | `func(str, char string, length int) string` | 前置填充到指定长度 |
| `FillAfter` | `func(str, char string, length int) string` | 后置填充到指定长度 |
| `Truncate` | `func(str string, maxBytes int) string` | 截断到指定字节长度 |
| `TruncateAppendEllipsis` | `func(str string, maxChars int) string` | 截断到指定字符数并追加 `...` |
| `TruncateMessage` | `func(content string, maxLen int) string` | 截断消息用于日志显示 |
| `Repeat` | `func(str string, count int) string` | 重复字符串 |
| `RepeatByLength` | `func(str string, padLen int) string` | 重复到指定长度 |
| `RepeatAndJoin` | `func(str, delimiter string, count int) string` | 重复并通过分界符连接 |

#### 格式化

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Format` | `func(template string, params map[string]interface{}) string` | 通过 map 格式化字符串（`{key}` 占位符） |
| `IndexedFormat` | `func(template string, params []interface{}) string` | 有序格式化（`{0}` `{1}` 占位符） |
| `AddPrefixIfNot` | `func(str, prefix string) string` | 不以 prefix 开头时补充前缀 |
| `AddSuffixIfNot` | `func(str, suffix string) string` | 不以 suffix 结尾时补充后缀 |
| `SanitizeSlug` | `func(name string) string` | 格式化为 URL 友好的 slug |
| `InsertSpaces` | `func(str string, interval int) string` | 按间隔插入空格 |

#### 索引查找

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `IndexOf` | `func(str, subStr string) int` | 查找子串位置 |
| `IndexOfIgnoreCase` | `func(str, subStr string) int` | 查找子串位置（忽略大小写） |
| `LastIndexOf` | `func(str, subStr string) int` | 查找最后出现位置 |
| `LastIndexOfIgnoreCase` | `func(str, subStr string) int` | 查找最后出现位置（忽略大小写） |
| `SafeIndexOfByRange` | `func(str, subStr string, options ...func(*searchOptions)) int` | 指定范围内查找（避免溢出） |
| `IndexOfByRange` | `func(str, subStr string, start, end int) int` | 指定范围查找 |
| `IndexOfByRangeStart` | `func(str, subStr string, start int) int` | 从指定位置查找 |
| `OrdinalIndexOf` | `func(str, subStr string, ordinal int, start ...int) int` | 查找第 n 次出现位置 |

#### 分割

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Split` | `func(str, separator string) []string` | 分割字符串 |
| `SplitLimit` | `func(str, separator string, limit int) []string` | 分割并限制分片数 |
| `SplitTrim` | `func(str, separator string) []string` | 分割并去除空白项 |
| `SplitTrimLimit` | `func(str, separator string, limit int) []string` | 分割、去除空白项、限制分片数 |
| `SplitAfterMapping` | `func[T any](str, separator string, mapping func(s string) (T, error)) []T` | 分割后通过 mapping 转换 |
| `SplitByLen` | `func(str string, length int) []string` | 按长度截取为多份 |
| `Cut` | `func(str string, n int) []string` | 切分为 n 等份 |
| `UniqueStringSlice` | `func(slice []string) []string` | 去重（去除空串） |

#### 比较与统计

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Equals` | `func(str1, str2 string) bool` | 比较（大小写敏感） |
| `EqualsIgnoreCase` | `func(str1, str2 string) bool` | 比较（忽略大小写） |
| `EqualsAny` | `func(str1 string, str2 []string) bool` | 与任一字符串相同 |
| `EqualsAnyIgnoreCase` | `func(str1 string, str2 []string) bool` | 与任一字符串相同（忽略大小写） |
| `EqualsAt` | `func(value string, position int, subStr string) bool` | 指定位置字符匹配 |
| `Count` | `func(str, searchStr string) int` | 统计子串出现次数 |
| `CompareIgnoreCase` | `func(str1, str2 string) int` | 比较用于排序（忽略大小写） |
| `Length` | `func(str string) int` | 计算 rune 长度 |

#### 显示宽度

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `RuneWidth` | `func(r rune) int` | 返回单个 rune 显示宽度（基于 East Asian Width） |
| `DisplayWidth` | `func(s string) int` | 计算字符串显示宽度（CJK 计 2，ASCII 计 1） |

#### 域名处理

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ExtractDomainPrefix` | `func(fullDomain, primaryDomain string) string` | 提取域名前缀（根域名返回 `@`） |
| `IsSubdomain` | `func(subdomain, primaryDomain string) bool` | 判断是否为子域名 |
| `SplitDomain` | `func(fullDomain, primaryDomain string) (string, string)` | 分割为前缀和主域名 |

#### 字段解析

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `ParseFieldInt` | `func(field string, min, max int) (int, error)` | 解析数字字段（带范围校验） |
| `ParseFieldIntOrWildcard` | `func(field, wildcard string, min, max int) (int, error)` | 解析数字字段（支持通配符，通配符返回 -1） |

#### 工具函数

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `Reverse` | `func(str string) string` | 反转字符串 |
| `Coalesce` | `func(s ...string) string` | 高性能字符串拼接 |
| `CalculateMD5Hash` | `func(input string) string` | 计算 MD5 哈希 |
| `ToSliceByte` | `func(s string) []byte` | 零拷贝转字节切片（unsafe） |
| `ToInt` | `func(s string) (int, error)` | 字符串转 int |
| `ExtractValue` | `func(extra, key, searchPrefix string) string` | 从字符串提取键值 |
| `FindKeysByValue` | `func(data map[string]string, searchValue string) []string` | 按值反查键 |
| `NormalizeSQLDirection` | `func(direction, defaultDirection string) string` | 规范化排序方向（ASC/DESC） |
| `QuoteJSONBytes` | `func(str string) []byte` | 按 JSON 规则转义并返回带双引号的字节切片 |

#### 快速格式化

| 导出名称 | 签名 | 说明 |
|---|---|---|
| `FastItoa` | `func(val int) string` | 快速整数转字符串（小整数使用缓存） |
| `FastFloat` | `func(val float64, prec int) string` | 浮点数转字符串 |
| `FastAppendInt` | `func(buf []byte, val int) []byte` | 将整数追加到 buffer |
| `FastFormatTime` | `func(buf []byte, t time.Time) []byte` | 格式化时间为 `YYYY/M/D H:MM:SS ` |
| `FastFormatTimeISO` | `func(buf []byte, t time.Time) []byte` | 格式化时间为 ISO 样式 `YYYY-MM-DD HH:MM:SS` |
| `FastFormatTimeCompact` | `func(buf []byte, t time.Time) []byte` | 格式化为紧凑数字样式 `YYYYMMDDHHMMSS` |

### 类型

| 导出名称 | 说明 |
|---|---|
| `StringX` | 链式字符串操作类型（指针语义，方法修改内部 value） |
| `PadPosition` | 填充位置常量类型（`Left`/`Right`/`Middle`） |
| `Paddler` | 填充器结构体（含 `Position` 字段） |
| `CharacterStyle` | 命名风格枚举类型 |
| `searchOptions` | 索引查找选项（非导出） |

### 常量/变量

| 导出名称 | 值 | 说明 |
|---|---|---|
| `Left` | `PadPosition(0)` | 左填充 |
| `Right` | `PadPosition(1)` | 右填充 |
| `Middle` | `PadPosition(2)` | 中间填充 |
| `SnakeCharacterStyle` | `CharacterStyle(0)` | 蛇形命名法 |
| `StudlyCharacterStyle` | `CharacterStyle(1)` | 首字母大写风格 |
| `CamelCharacterStyle` | `CharacterStyle(2)` | 驼峰命名法 |
| `KebabCharacterStyle` | `CharacterStyle(3)` | 短横线命名法 |
| `PascalCharacterStyle` | `CharacterStyle(4)` | 帕斯卡命名法 |
| `RootDomainPrefix` | `"@"` | 根域名前缀标识符 |

### StringX 链式方法

`StringX` 支持以下链式方法（返回 `*StringX` 的方法可继续链式调用）：

**变换类（返回 `*StringX`）**：
`ToLowerChain`, `ToUpperChain`, `ToTitleChain`, `ReverseChain`, `ConvertCharacterStyleChain`, `TrimChain`, `TrimStartChain`, `TrimEndChain`, `CleanEmptyChain`, `TrimProtocolChain`, `TrimAllChain`, `TrimAnyChain`, `TrimAllLineBreaksChain`, `TrimNewlinesChain`, `TrimStartNewlinesChain`, `TrimEndNewlinesChain`, `TrimPrefixChain`, `TrimPrefixIgnoreCaseChain`, `TrimSuffixChain`, `TrimSuffixIgnoreCaseChain`, `TrimSymbolsChain`, `ReplaceChain`, `ReplaceAllChain`, `ReplaceWithMatcherChain`, `ReplaceSpecialCharsChain`, `HideChain`, `SubBeforeChain`, `SubAfterChain`, `SubBetweenChain`, `SubStringChain`, `PadChain`(无), `FillBeforeChain`, `FillAfterChain`, `FormatChain`, `IndexedFormatChain`, `TruncateChain`, `TruncateAppendEllipsisChain`, `AddPrefixIfNotChain`, `AddSuffixIfNotChain`, `SanitizeSlugChain`, `InsertSpacesChain`, `RepeatChain`, `RepeatByLengthChain`, `RepeatAndJoinChain`, `CoalesceChain`

**判断类（返回 `bool`）**：
`EqualsChain`, `EqualsIgnoreCaseChain`, `EqualsAnyChain`, `EqualsAnyIgnoreCaseChain`, `EqualsAtChain`, `EqualsTrimIgnoreCaseChain`, `IsBlankChain`, `ContainsChain`, `ContainsIgnoreCaseChain`, `ContainsAnyChain`, `ContainsAnyIgnoreCaseChain`, `ContainsAllChain`, `ContainsBlankChain`, `StartWith`(无), `GetContainsStrChain`

**取值类（返回值）**：
`Value() string`, `String() string`, `LengthChain() int`, `CountChain(searchStr string) int`, `CompareIgnoreCaseChain(str2 string) int`, `DisplayWidthChain() int`

### 索引查找选项函数

| 函数 | 签名 | 说明 |
|---|---|---|
| `WithStart` | `func(start int) func(*searchOptions)` | 配置起始位置 |
| `WithEnd` | `func(end int) func(*searchOptions)` | 配置结束位置 |

## 常用示例

详细用法参阅 → [reference.md](reference.md)

## 注意事项

- `New` 返回 `*StringX`（指针），链式方法修改内部 `value` 字段，非不可变对象
- `SubBefore`/`SubAfter` 的第三个参数 `isLastSeparator` 控制是否取最后一个分隔符位置
- `StartWithAny`/`EndWithAny` 接收 `[]string` 切片，非可变参数
- `Pad` 使用 `*` 作为填充字符，通过可选 `*Paddler` 指定位置（默认 `Middle`）
- `ReplaceWithMatcher` 接收正则表达式字符串和替换函数 `func(string) string`
- `ReplaceSpecialChars` 接收 `rune` 类型替换值
- `Hide` 的 `startInclude` 包含、`endExclude` 不包含（即 `[start, end)` 区间）
- `NormalizeFieldName` 返回 `[]string`（所有命名风格变体），非单个字符串
- `DisplayWidth` 基于 East Asian Width 标准，CJK 字符计宽为 2，Emoji 计 2，ASCII 计 1
- `ToSliceByte` 使用 unsafe 零拷贝，结果切片不可修改
- `SplitAfterMapping` 为泛型函数，支持自定义映射转换
- 快速格式化函数（`FastItoa`/`FastFormatTime*`）使用预分配缓存，适合高并发热路径
