package utils

import (
	"context"
	"database/sql"
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

func InitDatabase() (testcontainers.Container, error) {
	// Start a PostgreSQL container
	ctx := context.Background()
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
		},
	}
	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %v", err)
	}

	// Get the container's host and port
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL container port: %v", err)
	}

	// Set up the database connection
	dbLib.DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Int()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	return container, nil
}

func Setup() {
	// Run SQL scripts to initialize test data
	err := dbLib.RunSQL("../db/init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}

func Teardown() {
	// Run SQL script to delete all test data
	err := dbLib.RunSQL("../db/drop.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}
