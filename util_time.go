package tgo

import (
	"time"
)

func UtilTimeGetDate(t time.Time) time.Time {

	year, month, day := t.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func UtilTimeIsToday(t time.Time) bool {
	return UtilTimeSameDay(t, time.Now())
}

func UtilTimeSameDay(t1 time.Time, t2 time.Time) bool {
	if UtilTimeDiffDay(t1, t2) == 0 {
		return true
	}
	return false
}

func UtilTimeDiffDay(t1 time.Time, t2 time.Time) int {
	return int(UtilTimeGetDate(t2).Sub(UtilTimeGetDate(t1)) / (24 * time.Hour))
}
