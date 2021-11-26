package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Time 字符串 -> 时间
func Time(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}

func MustTime(s string) time.Time {
	tm, err := Time(s)
	if err != nil {
		return time.Now()
	}
	return tm
}

func Test_getPlacePath(t *testing.T) {
	assert.Equal(t, getPlacePath(MustTime("2016-01-02 15:04:05")), "2016/01/2016-01-02")
	assert.Equal(t, getPlacePath(MustTime("2019-11-22 15:04:05")), "2019/11/2019-11-22")
}

func Test_rename(t *testing.T) {
	// rename("./t2", "./a")
	// rename("/Volumes/home/Photos/Mobile/iPhone", "./a")
	// afp://OASSIS._afpovertcp._tcp.local/home/Photos/Mobile/iPhone
	// rename("//James@OASSIS._afpovertcp._tcp.local/home/Photos/Mobile/iPhone", "./a")
}
