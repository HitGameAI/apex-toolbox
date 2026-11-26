package units

import "errors"

// 定义错误变量，方便统一管理和复用
var (
	ErrInvalidSizeFormat = errors.New("无效的大小格式")
	ErrNegativeSize      = errors.New("大小不能为负数")
	ErrInvalidUnitSuffix = errors.New("无效的单位后缀")
)
