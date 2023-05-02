package db

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"io"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	Active = iota + 1
	Inactive
)

var (
	db *sql.DB
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
	log.Printf("user %v created", user)
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

func handleErr(res sql.Result) (int64, error) {
	nr, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return nr, nil
}

func Close() error {
	return db.Close()
}

func Exec(query string, args ...any) (sql.Result, error) {
	return db.Exec(query, args)
}

// stringPointer connection with postgres db
func InitDb(dbDriver string, dbPath string, dbScripts string) error {
	// Open the connection
	var err error
	db, err = sql.Open(dbDriver, dbPath)
	if err != nil {
		return err
	}

	//we wait for ten seconds to connect
	err = db.Ping()
	now := time.Now()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		time.Sleep(time.Second)
		err = db.Ping()
	}
	if err != nil {
		return err
	}

	files := strings.Split(dbScripts, ":")
	err = RunSQL(files...)
	if err != nil {
		return err
	}

	log.Infof("Successfully connected!")
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
					_, err := db.Exec(request)
					if err != nil {
						return fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			log.Printf("ignoring file %v (%v)", file, err)
		}
	}
	return nil
}

func closeAndLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("could not close: %v", err)
	}
}

func stringPointer(s string) *string {
	return &s
}

func extractGitUrls(repos []Repo) []string {
	gitUrls := []string{}
	for _, v := range repos {
		if v.GitUrl != nil {
			gitUrls = append(gitUrls, *v.GitUrl)
		}
	}
	return gitUrls
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
