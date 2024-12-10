package cloneruntime

import (
	"time"
)

func Time(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
}

func Assign[T any](t T) T {
	return t
}
