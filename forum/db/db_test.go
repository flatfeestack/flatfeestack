package database_test

import (
	"context"
	"forum/utils"
	dbLib "github.com/flatfeestack/go-lib/database"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	container, err := utils.InitDatabase()
	if err != nil {
		log.Error(err)
		panic(err)
	}

	// Run tests
	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer container.Terminate(ctx)

	os.Exit(code)
}
