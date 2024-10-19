package db

import (
	"fmt"
	"log/slog"
	"os"
)

type TestDb struct {
	file *os.File
}

func NewTestDb() *TestDb {
	file, err := os.CreateTemp("", "test-db-ffs-*.sqlite")
	if err != nil {
		slog.Error("File error",
			slog.Any("error", err))
	}
	fmt.Printf("SQLite DB path: %v", file.Name())
	err = InitDb("sqlite3", file.Name(), "")
	if err != nil {
		slog.Error("DB error",
			slog.Any("error", err))
	}
	return &TestDb{file}
}

func (t *TestDb) CloseTestDb() {
	defer func() {
		err := os.Remove(t.file.Name())
		if err != nil {
			slog.Error("DB error",
				slog.Any("error", err))
		}
	}()
	err := DB.Close()
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
	err := RunSQL("../../db/init.sql")
	if err != nil {
		slog.Error("Could not run init.sql scripts",
			slog.Any("error", err))
	}
}
func TeardownTestData() {
	err := RunSQL("../../db/delAll_test.sql")
	if err != nil {
		slog.Error("Could not run delAll_test.sql",
			slog.Any("error", err))
	}
}
