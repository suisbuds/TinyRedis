package pkg

import (
	"time"
)

/*
封装时间库
*/

const YYYY_MM_DD_HH_MM_SS = "2008-01-01 12:00:00"


func TimeNow() time.Time {
	return time.Now()
}

func ParseTimeSecondFormat(timeStr string) (time.Time, error) {
	return time.ParseInLocation(YYYY_MM_DD_HH_MM_SS, timeStr, time.Local)
}

func TimeSecondFormat(t time.Time) string {
	return t.Format(YYYY_MM_DD_HH_MM_SS)
}
