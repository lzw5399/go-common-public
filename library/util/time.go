package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FormatTime struct {
	TimeStamp int64
	Date      string
	DateTime  string
	Hour      int
}

func NowTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetDateMs(myTime int64) int64 {
	tmpMs := myTime / 1000
	finalMs := tmpMs * 1000

	return finalMs
}

func GetSpecialMs(myTime int64) int64 {
	if myTime == 0 {
		return 0
	}

	t := time.Unix(myTime/1000, 0)
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()

	ms := hour*3600000 + min*60000 + sec*1000

	return int64(ms)
}

func GetTimeStr(myTimeSec int64, format string) string {
	tm := time.Unix(myTimeSec, 0)
	return tm.Format("2006/01/02")
}

func FormatYearMonth() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
}

// GetTheLastMonthForYear 获取上一个月份
func GetTheLastMonthForYear(mon time.Month, year int) (time.Month, int) {
	if mon == 1 {
		return 12, year - 1
	}

	return mon - 1, year
}

func CountDayForMon(year int, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30

		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return
}

// GetTheLastMondayForNowGap 获取上一个星期一到现在相差多少天
func GetTheLastMondayForNowGap(weekDay int) int {
	return weekDay + 7 - 1
}

func FormatUnixTime(timeStamp int64) FormatTime {
	t := time.Unix(timeStamp/1000, 0)
	return FormatTime{
		TimeStamp: timeStamp,
		Date:      t.Format("2006-01-02"),
		DateTime:  t.Format("2006-01-02 15:04:05"),
		Hour:      GetDateHour(t.Format("2006-01-02 15:04:05")),
	}
}

func FormatNowTime() FormatTime {
	now := time.Now()
	ts := now.UnixNano() / 1000000
	return FormatTime{
		TimeStamp: ts,
		Date:      now.Format("2006-01-02"),
		DateTime:  now.Format("2006-01-02 15:04:05"),
		Hour:      GetDateHour(now.Format("2006-01-02 15:04:05")),
	}
}

func DefaultTime() FormatTime {
	dt, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	return FormatTime{
		TimeStamp: (dt.Unix() - 8*60*60) * 1000,
		Date:      "2006-01-02",
		DateTime:  "2006-01-02 15:04:05",
		Hour:      GetDateHour("2006-01-02 15:04:05"),
	}
}

func GetDateHour(dateTime string) int {
	ts := strings.Split(dateTime, " ")
	if len(ts) < 2 {
		return 0
	}
	ds := strings.Split(ts[1], ":")
	if len(ds) < 3 {
		return 0
	}
	hour, err := strconv.Atoi(ds[0])
	if err != nil {
		return 0
	}
	return hour
}

func GetPreDay(day int) string {
	nTime := time.Now()
	preTime := nTime.AddDate(0, 0, day)
	return preTime.Format("2006-01-02")
}

func FomartYearMonth() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
}

func FomartTimeYearMonth(timeStamp int64) string {
	t := time.Unix(timeStamp/1000, 0)
	return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
}

func GetDayStartTime(timeStamp int64) int64 {
	t := time.Unix(timeStamp, 0)
	dayTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return dayTime.UnixMilli()
}

// GetMonthStartMs 获取当前时间月初的毫秒级时间戳
func GetMonthStartMs(ms int64) int64 {
	// 获取当前时间的毫秒级时间戳
	t := time.UnixMilli(ms)

	year := t.Year()
	mon := t.Month()
	loc, _ := time.LoadLocation("Local")

	strTime := fmt.Sprintf("%d-%d-01 00:00:00", year, mon)
	if mon < 10 {
		strTime = fmt.Sprintf("%d-0%d-01 00:00:00", year, mon)
	}
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", strTime, loc)
	monStartMs := startTime.Unix() * 1000

	return monStartMs
}

// GetMonthEndMs 获取当前时间月末的毫秒级时间戳
func GetMonthEndMs(ms int64) int64 {
	return GetNextMonthStartMs(ms) - 1
}

// GetNextMonthStartMs 获取指定时间戳的下个月初的毫秒级时间戳
func GetNextMonthStartMs(startMs int64) int64 {
	// 获取起始时间
	start := time.UnixMilli(startMs)

	// 获取当前年份和月份
	year := start.Year()
	month := start.Month()
	// 定义本地时区
	loc, _ := time.LoadLocation("Local")

	// 如果是12月，则下一个月是下一年的1月
	if month == time.December {
		year++
		month = time.January
	} else {
		// 否则，月份加1
		month++
	}

	// 格式化日期为下个月初的日期
	strTime := fmt.Sprintf("%d-%d-01 00:00:00", year, month)
	if month < 10 {
		strTime = fmt.Sprintf("%d-0%d-01 00:00:00", year, month)
	}

	// 解析时间字符串为time.Time对象
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", strTime, loc)
	// 计算并返回毫秒级时间戳
	nextMonthStartMs := startTime.Unix() * 1000

	return nextMonthStartMs
}

// GetNextMonthEndMs 获取指定时间戳的下个月末的毫秒级时间戳
func GetNextMonthEndMs(startMs int64) int64 {
	// 获取下下个月初的毫秒级时间戳 - 1
	return GetNextMonthStartMs(GetNextMonthStartMs(startMs)) - 1
}

// GetCurDayEndMs 获取当前时间的当天结束的毫秒级时间戳
func GetCurDayEndMs(ms int64) int64 {
	// 获取当前时间的毫秒级时间戳
	t := time.UnixMilli(ms)

	// 获取当前时间的年月日
	year := t.Year()
	mon := t.Month()
	day := t.Day()

	// 定义本地时区
	loc, _ := time.LoadLocation("Local")

	// 格式化日期为当天结束的日期
	var strTime string
	if mon < 10 {
		if day < 10 {
			strTime = fmt.Sprintf("%d-0%d-0%d 23:59:59", year, mon, day)
		} else {
			strTime = fmt.Sprintf("%d-0%d-%d 23:59:59", year, mon, day)
		}
	} else {
		if day < 10 {
			strTime = fmt.Sprintf("%d-%d-0%d 23:59:59", year, mon, day)
		} else {
			strTime = fmt.Sprintf("%d-%d-%d 23:59:59", year, mon, day)
		}
	}

	// 解析时间字符串为time.Time对象
	endTime, _ := time.ParseInLocation("2006-01-02 15:04:05", strTime, loc)
	// 计算并返回毫秒级时间戳
	curDayEndMs := endTime.Unix() * 1000

	return curDayEndMs
}
