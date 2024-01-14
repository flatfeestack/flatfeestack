package util

import "time"

var (
	delta int
)

func AddTimeNowSeconds(deltaSeconds int) {
	delta += deltaSeconds
}

func ResetTimeNow() {
	delta = 0
}

func SecondsAdd() int {
	return delta
}

func TimeNow() time.Time {
	if debug {
		return time.Now().Add(time.Duration(delta) * time.Second).UTC()
	} else {
		return time.Now().UTC()
	}
}
