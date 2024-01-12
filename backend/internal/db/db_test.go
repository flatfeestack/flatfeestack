package db

import (
	dbLib "github.com/flatfeestack/go-lib/database"
	"log/slog"
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
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())

	err = dbLib.InitDb("sqlite3", file.Name(), "")
	if err != nil {
		slog.Error("DB error",
			slog.Any("error", err))
	}

	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		slog.Warn("Could not start resource",
			slog.Any("error", err))
	}

	if err != nil {
		slog.Warn("Could not start resource",
			slog.Any("error", err))
	}

	os.Exit(code)
}

func setup() {
	err := dbLib.RunSQL("init.sql")
	if err != nil {
		slog.Error("Could not run init.sql scripts",
			slog.Any("error", err))
	}
}
func teardown() {
	err := dbLib.RunSQL("delAll_test.sql")
	if err != nil {
		slog.Error("Could not run delAll_test.sql",
			slog.Any("error", err))
	}
}
