package util

import (
	"errors"
	"time"
)

const GoBeginTimeString = "2006-01-02,15:04:05"

const (
	OneDay    int64 = 24 * OneHour
	OneHour   int64 = 60 * OneMinute
	OneMinute int64 = 60

	OneDayUint32    uint32 = 24 * OneHourUint32
	OneHourUint32   uint32 = 60 * OneMinuteUint32
	OneMinuteUint32 uint32 = 60
)

type Time struct {
	Unix         int64  // unix时间戳
	TodayUnix    int64  // 今日零点的时间戳（注意：当时间戳为0时，在北京时区，这个值是负的）
	TodaySeconds uint32 // 今日从零点算起已过的秒数
	Month        int    // 当前月
	WeekDay      int    // 0周天，1周一，2周二...
	DayID        int64  // 当前时区当前DayID（从1970-01-01开始算）
	WeekID       int64  // 当前时间当前WeekID（从1970-01-01开始算）
	MonthID      int64  // 当前时间当前MonthID（从1970-01-01开始算）
}

func NewTime(t time.Time) *Time {
	ext := new(Time)
	ext.Unix = t.Unix()
	ext.TodaySeconds = uint32(t.Hour()*3600 + t.Minute()*60 + t.Second())
	ext.TodayUnix = ext.Unix - int64(ext.TodaySeconds)

	ext.Month = int(t.Month())
	ext.WeekDay = int(t.Weekday())

	_, offset := t.Zone()
	ext.DayID = (t.Unix() + int64(offset)) / (24 * 3600)
	ext.WeekID = (ext.DayID + 3) / 7 // DayID为0时，是周四，所以加3除7来算WeekID，这样，0,1,2,3是第0周，4,5,6,7,8,9,10是第1周（注：周一到周日算一周）
	ext.MonthID = int64(t.Year())*12 + (int64(t.Month()) - 1) - 1970*12

	return ext
}

func BaseTimeUint32() uint32 {
	t := time.Now()
	return uint32(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix())
}

func ToBaseTime(useTime uint32) time.Time {
	t := time.Unix(int64(useTime), 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func NowUint32() uint32 {
	return uint32(time.Now().Unix())
}

// 毫秒
func NowMs() int64 {
	return time.Now().UnixMilli()
}

func NowInt64() int64 {
	return time.Now().UnixNano()
}

// 时间文本转为Unix时间（秒）
func TimeStrToUnix(str string) (uint32, error) {
	var parseTime, err = time.ParseInLocation(GoBeginTimeString, str, time.Local)
	if err != nil {
		return 0, err
	}
	t := parseTime.Unix()
	if t < 0 {
		return 0, errors.New("too early time")
	}
	return uint32(t), nil
}

// Unix时间（秒）转为时间文本
func UnixToTimeStr(t int64) string {
	return time.Unix(t, 0).Format(GoBeginTimeString)
}

func GetNextMonthStartTime(t time.Time) int64 {
	year, month, _ := t.Date()

	// 如果当前月份是12月，则下个月的年份加1，月份设置为1
	if month == time.December {
		year += 1
		month = time.January
	} else {
		month += 1
	}

	// 设置时间为下个月的零点零分零秒
	nextMonthStart := time.Date(year, month, 1, 0, 0, 0, 0, t.Location()).Unix()

	return nextMonthStart
}

// 获取对应日期的周一0点时间戳
func GetWeekStartTime(t time.Time) int64 {
	// 计算距离上一个周一还有多少天
	daysSinceMonday := int(t.Weekday()) - 1
	if daysSinceMonday < 0 {
		daysSinceMonday = 6
	}

	// 使用Add函数减去相应的天数
	prevMonday := t.AddDate(0, 0, -daysSinceMonday)

	prevMondayTimestamp := time.Date(prevMonday.Year(), prevMonday.Month(), prevMonday.Day(), 0, 0, 0, 0, prevMonday.Location()).Unix()

	return prevMondayTimestamp
}
