package db

import (
	"backend/pkg/util"
	"os"
	"testing"
	"time"
)

var (
	day1 = time.Time{}
	day2 = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day3 = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day4 = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day5 = time.Time{}.Add(time.Duration(4*24) * time.Hour)
	day6 = time.Time{}.Add(time.Duration(5*24) * time.Hour)
)

func TestMain(m *testing.M) {
	util.SetupTestDb()
	code := m.Run()
	util.CloseTestDb()
	os.Exit(code)
}
