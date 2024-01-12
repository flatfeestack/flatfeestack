package db

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"math/big"
	"time"
)

const (
	Active = iota + 1
	Inactive
)

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

//type Balance struct {
//	Balance       *big.Int
//	DailySpending *big.Int
//}

func CreateUser(email string, now time.Time) (*UserDetail, error) {
	user := User{
		Id:        uuid.New(),
		Email:     email,
		CreatedAt: now,
	}
	userDetail := UserDetail{
		User: user,
	}

	err := InsertUser(&userDetail)
	if err != nil {
		return nil, err
	}
	slog.Info("user %v created",
		slog.Any("user", user))
	return &userDetail, nil
}

func handleErrMustInsertOne(res sql.Result) error {
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nr == 0 {
		return fmt.Errorf("0 rows affacted, need at least 1")
	} else if nr != 1 {
		return fmt.Errorf("Only 1 row needs to be affacted, got %v", nr)
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
