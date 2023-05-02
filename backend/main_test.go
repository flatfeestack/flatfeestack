package main

import (
	"backend/clients"
	db "backend/db"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	opts = &Opts{}
	opts.Env = "test"
	opts.DBDriver = "sqlite3"
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())
	opts.DBPath = file.Name()

	err = db.InitDb("sqlite3", file.Name(), "db/init.sql")
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
	err := db.RunSQL("db/init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}

func teardown() {
	err := db.RunSQL("db/delAll_test.sql")
	if err != nil {
		log.Fatalf("Could not run delAll_test.sql: %s", err)
	}
	clients.EmailNotifications = 0
	clients.EmailNoNotifications = 0
}
