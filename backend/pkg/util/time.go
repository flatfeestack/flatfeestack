package util

import "time"

var (
	SecondsAdd int
)

func TimeNow() time.Time {
	if debug {
		return time.Now().Add(time.Duration(SecondsAdd) * time.Second).UTC()
	} else {
		return time.Now().UTC()
	}
}
