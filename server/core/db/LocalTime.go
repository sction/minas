// db包提供了数据库相关的基础功能
package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

// LocalTime 自定义时间类型，用于格式化数据库中的时间字段
type LocalTime struct {
	time.Time
}

// Now 返回当前时间的LocalTime类型
func (t LocalTime) Now() LocalTime {
	t = LocalTime{time.Now()}
	return t
}

// MarshalJSON 将LocalTime类型转换为JSON字符串
// 实现json.Marshaler接口
// 返回:
//   - []byte: JSON字符串的字节数组
//   - error: 错误信息
func (t LocalTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format(TimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON 将JSON字符串解析为LocalTime类型
// 实现json.Unmarshaler接口
// 参数:
//   - data: JSON字符串的字节数组
//
// 返回:
//   - error: 错误信息
func (t *LocalTime) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	// 处理空字符串、null或空白情况
	if str == `""` || str == "null" || str == `" "` || len(str) <= 2 {
		*t = LocalTime{time.Time{}} // 设置为零值时间
		return nil
	}

	now, err := time.ParseInLocation(`"`+TimeFormat+`"`, str, time.Local)
	*t = LocalTime{now}
	return
}

// Value 用于数据库写入操作
// 实现driver.Valuer接口
// 返回:
//   - driver.Value: 数据库中存储的值
//   - error: 错误信息
func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan 用于数据库读取操作
// 实现sql.Scanner接口
// 参数:
//   - v: 数据库中的值
//
// 返回:
//   - error: 错误信息
func (t *LocalTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = LocalTime{Time: value}
		return nil
	}
	return fmt.Errorf("无法将 %v 转换为 time.Time", v)
}
