package tgo

import (
	"time"
)

func UtilTimeGetDate(t time.Time) time.Time {

	year, month, day := t.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func UtilTimeIsToday(t time.Time) bool {
	if t.Format("20060102") == time.Now().Format("20060102") {
		return true
	}
	return false
}
