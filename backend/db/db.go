package db

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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
	db  *sql.DB
	agg string
)

type SponsorEvent struct {
	Id          uuid.UUID  `json:"id"`
	Uid         uuid.UUID  `json:"uid"`
	RepoId      uuid.UUID  `json:"repo_id"`
	EventType   uint8      `json:"event_type"`
	SponsorAt   time.Time  `json:"sponsor_at"`
	UnSponsorAt *time.Time `json:"un_sponsor_at"`
}

type RepoBalance struct {
	Repo            Repo                `json:"repo"`
	CurrencyBalance map[string]*big.Int `json:"currencyBalance"`
}

type Repo struct {
	Id          uuid.UUID `json:"uuid"`
	Url         *string   `json:"url"`
	GitUrl      *string   `json:"gitUrl"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Score       uint32    `json:"score"`
	Source      *string   `json:"source"`
	CreatedAt   time.Time
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

type UserBalance struct {
	UserId           uuid.UUID  `json:"userId"`
	Balance          *big.Int   `json:"balance"`
	Split            *big.Int   `json:"split"`
	PaymentCycleInId *uuid.UUID `json:"paymentCycleInId"`
	FromUserId       *uuid.UUID `json:"fromUserId"`
	BalanceType      string     `json:"balanceType"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type UserStatus struct {
	UserId   uuid.UUID `json:"userId"`
	Email    string    `json:"email"`
	Name     *string   `json:"name,omitempty"`
	DaysLeft int       `json:"daysLeft"`
}

type PaymentCycle struct {
	Id    uuid.UUID `json:"id"`
	Seats int64     `json:"seats"`
	Freq  int64     `json:"freq"`
}

type Contribution struct {
	RepoName         string    `json:"repoName"`
	RepoUrl          string    `json:"repoUrl"`
	SponsorName      *string   `json:"sponsorName,omitempty"`
	SponsorEmail     string    `json:"sponsorEmail"`
	ContributorName  *string   `json:"contributorName,omitempty"`
	ContributorEmail string    `json:"contributorEmail"`
	Balance          *big.Int  `json:"balance"`
	Currency         string    `json:"currency"`
	PaymentCycleInId uuid.UUID `json:"paymentCycleInId"`
	Day              time.Time `json:"day"`
}

type PayoutInfo struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

type Balance struct {
	Balance *big.Int
	Split   *big.Int
}

type AnalysisResponse struct {
	Id        uuid.UUID
	RequestId uuid.UUID `json:"request_id"`
	DateFrom  time.Time
	DateTo    time.Time
	GitEmail  string
	GitNames  []string
	Weight    float64
}

type AnalysisRequest struct {
	Id         uuid.UUID
	RepoId     uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
	GitUrl     string
	ReceivedAt *time.Time
	Error      *string
}

func CreateUser(email string, now time.Time) (*User, error) {
	user := User{
		Id:                uuid.New(),
		PaymentCycleOutId: uuid.New(),
		Email:             email,
		CreatedAt:         now,
	}

	err := InsertUser(&user)
	if err != nil {
		return nil, err
	}
	log.Printf("user %v created", user)
	return &user, nil
}

func handleErrMustInsertOne(res sql.Result) error {
	nr, err := res.RowsAffected()
	if nr == 0 || err != nil {
		return err
	}
	if nr != 1 {
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

	//this will create or alter tables
	//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	for _, file := range strings.Split(dbScripts, ":") {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := ioutil.ReadFile(file)
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
			log.Infof("ignoring file [%v] (%v)", file, err)
		}
	}

	if dbDriver == "sqlite3" {
		//https://database.guide/sqlite-json_group_array/
		agg = "json_group_array"
	} else if dbDriver == "postgres" {
		//https://www.postgresql.org/docs/9.5/functions-aggregate.html
		agg = "json_agg"
	}

	log.Infof("Successfully connected!")
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
