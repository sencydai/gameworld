package base

import (
	"time"
)

const (
	DATETIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT     = "2006-01-02"
	TIME_FORMAT     = "15:04:05"
)

func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

func FormatDateTime(t time.Time) string {
	return t.Format(DATETIME_FORMAT)
}

func FormatDate(t time.Time) string {
	return t.Format(DATE_FORMAT)
}

func FormatTime(t time.Time) string {
	return t.Format(TIME_FORMAT)
}

func Parse(layout string, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

func ParseDateTime(value string) (time.Time, error) {
	return time.ParseInLocation(DATETIME_FORMAT, value, time.Local)
}

func ParseDate(value string) (time.Time, error) {
	return time.ParseInLocation(DATE_FORMAT, value, time.Local)
}

func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation(TIME_FORMAT, value, time.Local)
}

func Unix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return d1 == d2 && m1 == m2 && y1 == y2
}

func CheckMomentHappend(t time.Time, hour, min, sec int) bool {
	now := time.Now()
	moment := time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, 0, time.Local)
	if !moment.Before(now) {
		return false
	}
	if moment.After(t) {
		return true
	}
	moment = moment.AddDate(0, 0, 1)
	return moment.Before(now)
}

func GetMomentDelay(hour, min, sec int) time.Duration {
	now := time.Now()
	moment := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, time.Local)
	if moment.After(now) {
		return moment.Sub(now)
	}
	return time.Hour*24 - now.Sub(moment)
}
