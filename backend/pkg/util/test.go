package util

import (
	dbLib "github.com/flatfeestack/go-lib/database"
	"log/slog"
	"os"
)

func SetupTestDb() {
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())

	err = dbLib.InitDb("sqlite3", file.Name(), "")
	if err != nil {
		slog.Error("DB error",
			slog.Any("error", err))
	}
}

func CloseTestDb() {
	err := dbLib.DB.Close()
	if err != nil {
		slog.Warn("Could not start resource",
			slog.Any("error", err))
	}

	if err != nil {
		slog.Warn("Could not start resource",
			slog.Any("error", err))
	}
}

func SetupTestData() {
	err := dbLib.RunSQL("init.sql")
	if err != nil {
		slog.Error("Could not run init.sql scripts",
			slog.Any("error", err))
	}
}
func TeardownTestData() {
	err := dbLib.RunSQL("delAll_test.sql")
	if err != nil {
		slog.Error("Could not run delAll_test.sql",
			slog.Any("error", err))
	}
}
