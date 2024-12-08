package db

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"backend/pkg/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Active = iota + 1
	Inactive
)

var (
	DB *sql.DB
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

func SetupRepo(url string) (*uuid.UUID, error) {
	r := Repo{
		Id:          uuid.New(),
		Url:         util.StringPointer(url),
		GitUrl:      util.StringPointer(url),
		Source:      util.StringPointer("github"),
		Name:        util.StringPointer("name"),
		Description: util.StringPointer("desc"),
		CreatedAt:   time.Time{},
	}
	err := InsertOrUpdateRepo(&r)
	if err != nil {
		return nil, err
	}
	return &r.Id, nil
}

func CloseDb() {
	err := DB.Close()
	slog.Warn("could not close the db", slog.Any("error", err))
}

func InitDb(dbDriver string, dbPath string, dbScripts string) error {
	// Open the connection
	var err error
	DB, err = sql.Open(dbDriver, dbPath)
	if err != nil {
		return err
	}

	//we wait for ten seconds to connect
	err = DB.Ping()
	now := time.Now()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		err = DB.Ping()
	}
	if err != nil {
		return err
	}

	files := strings.Split(dbScripts, ":")
	err = RunSQL(files...)
	if err != nil {
		return err
	}

	slog.Info("Successfully connected!")
	return nil
}

func RunSQL(files ...string) error {
	for _, file := range files {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			//https://stackoverflow.com/questions/12682405/strip-out-c-style-comments-from-a-byte
			re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/|(?s)--.*?\n|(?s)#.*?\n")

			newBytes := re.ReplaceAll(fileBytes, nil)

			requests := strings.Split(string(newBytes), ";")
			for _, request := range requests {
				request = strings.TrimSpace(request)
				if len(request) > 0 {
					_, err := DB.Exec(request)
					if err != nil {
						return fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			slog.Info("ignoring file %v (%v)", slog.String("file", file), slog.Any("error", err))
		}
	}

	return nil
}

func CloseAndLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		slog.Info("could not close: %v", slog.Any("error", err))
	}
}
