package util

import (
	dbLib "github.com/flatfeestack/go-lib/database"
	"log/slog"
	"os"
)

type TestDb struct {
	file *os.File
}

func NewTestDb() *TestDb {
	file, err := os.CreateTemp("", "test-db-ffs-*.sqlite")

	err = dbLib.InitDb("sqlite3", file.Name(), "")
	if err != nil {
		slog.Error("DB error",
			slog.Any("error", err))
	}
	return &TestDb{file}
}

func (t *TestDb) CloseTestDb() {
	defer os.Remove(t.file.Name())
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
	err := dbLib.RunSQL("../../migrations/init.sql")
	if err != nil {
		slog.Error("Could not run init.sql scripts",
			slog.Any("error", err))
	}
}
func TeardownTestData() {
	err := dbLib.RunSQL("../../migrations/delAll_test.sql")
	if err != nil {
		slog.Error("Could not run delAll_test.sql",
			slog.Any("error", err))
	}
}
