package db

import (
	log "github.com/sirupsen/logrus"
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

	err = InitDb("sqlite3", file.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = db.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	os.Exit(code)
}

func setup() {
	err := RunSQL("init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}
func teardown() {
	err := RunSQL("delAll_test.sql")
	if err != nil {
		log.Fatalf("Could not run delAll_test.sql: %s", err)
	}
}
