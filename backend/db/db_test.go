package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	day1 = time.Time{}
	day2 = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day3 = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day4 = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day5 = time.Time{}.Add(time.Duration(4*24) * time.Hour)
	day6 = time.Time{}.Add(time.Duration(5*24) * time.Hour)
	db *DB
)

type TestDb struct {
	container *postgres.PostgresContainer
	db        *DB
	ctx       context.Context
}

func TestMain(m *testing.M) {
	testDb := NewTestDb()
	defer testDb.Close()

	db = testDb.db
	err := db.RunMigrations(); 
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func NewTestDb() *TestDb {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		panic(fmt.Errorf("failed to start postgres container: %w", err))
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		panic(fmt.Errorf("failed to get connection string: %w", err))
	}

	db, err := New("postgres", connStr)
	if err != nil {
		postgresContainer.Terminate(ctx)
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	fmt.Printf("PostgreSQL test container started: %s\n", connStr)

	return &TestDb{
		container: postgresContainer,
		db:        db,
		ctx:       ctx,
	}
}


func (td *TestDb) DB() *DB {
	return td.db
}

func (td *TestDb) Close() {
	if td.db != nil {
		td.db.Close()
	}
	if td.container != nil {
		td.container.Terminate(td.ctx)
	}
}

// Helper to truncate all tables between tests (faster than recreating container)
func TruncateAll(db *DB, t *testing.T) {
	tables := []string{
		"user_emails_sent", "invite", "future_contribution", "unclaimed",
		"daily_contribution", "repo_metrics", "analysis_request",
		"multiplier_event", "trust_event", "sponsor_event", "git_email",
		"payment_in_event", "repo", "users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}
}
