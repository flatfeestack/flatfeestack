package db

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Active = iota + 1
	Inactive
)

type DB struct {
	*sql.DB
}

type RepoBalance struct {
	Repo            Repo                `json:"repo"`
	CurrencyBalance map[string]*big.Int `json:"currencyBalance"`
}

type PayoutRequest struct {
	UserId       uuid.UUID
	BatchId      uuid.UUID
	Currency     string
	ExchangeRate big.Float
	Tea          int64
	Address      string
	CreatedAt    time.Time
}

type GitEmail struct {
	Email       string     `json:"email"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type UserBalanceCore struct {
	UserId   uuid.UUID `json:"userId"`
	Balance  *big.Int  `json:"balance"`
	Currency string    `json:"currency"`
}

type PaymentCycle struct {
	Id    uuid.UUID `json:"id"`
	Seats int64     `json:"seats"`
	Freq  int64     `json:"freq"`
}

type PayoutInfo struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

func (db *DB) RunMigrations() error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Migrations completed successfully")
	return nil
}

func handleErrMustInsertOne(res sql.Result) error {
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nr == 0 {
		return fmt.Errorf("0 rows affacted, need at least 1")
	} else if nr != 1 {
		return fmt.Errorf("only 1 row needs to be affacted, got %v", nr)
	}
	return nil
}

func stringPointer(s string) *string {
	return &s
}

// https://stackoverflow.com/a/33072822
type JsonNullTime struct {
	sql.NullTime
}

func (v JsonNullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (db *DB) CloseDb() {
	err := db.Close()
	slog.Warn("could not close the db", slog.Any("error", err))
}

func New(dbDriver string, dbPath string) (db *DB, err error) {
	// Open the connection
	db1, err := sql.Open(dbDriver, dbPath)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 40; i++ { // 40 * 250ms = 10s
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("database connection failed after 10s: %w", err)
	}

	slog.Info("Successfully connected!")
	return &DB{DB: db1}, nil
}

func CloseAndLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		slog.Info("could not close: %v", slog.Any("error", err))
	}
}
