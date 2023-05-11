package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

// string mapping
const (
	PayInRequest   string = "REQUEST"
	PayInSuccess          = "SUCCESS"
	PayInFee              = "FEE"
	PayInAction           = "ACTION"
	PayInMethod           = "METHOD"
	PayInPartially        = "PARTIALLY"
	PayInExpired          = "EXPIRED"
	PayInFailed           = "FAILED"
	PayInRefunded         = "REFUNDED"
)

type PayInEvent struct {
	Id         uuid.UUID `json:"id"`
	ExternalId uuid.UUID `json:"externalId"`
	UserId     uuid.UUID `json:"userId"`
	Balance    *big.Int  `json:"balance"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	Seats      int64     `json:"seats"`
	Freq       int64     `json:"freq"`
	CreatedAt  time.Time `json:"createdAt"`
}

type PayTransferEvent struct {
	UserFromId uuid.UUID `json:"userFromId"`
	UserToId   uuid.UUID `json:"userToId"`
	Balance    *big.Int  `json:"balance"`
	Currency   string    `json:"currency"`
	CreatedAt  time.Time `json:"createdAt"`
}

func InsertPayInEvent(payInEvent PayInEvent) error {
	stmt, err := db.Prepare(`INSERT INTO payment_in_event(id, external_id, user_id, balance,  
                                                                currency, status, seats, freq, created_at) 
                                   VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payment_in_event for %v statement event: %v", payInEvent.UserId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := payInEvent.Balance.String()
	res, err = stmt.Exec(payInEvent.Id, payInEvent.ExternalId, payInEvent.UserId, b, payInEvent.Currency,
		payInEvent.Status, payInEvent.Seats, payInEvent.Freq, payInEvent.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindPayInUser(userId uuid.UUID) ([]PayInEvent, error) {
	s := `SELECT balance, currency, status, seats, freq, created_at 
          FROM payment_in_event 
          WHERE user_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	payInEvents := []PayInEvent{}
	for rows.Next() {
		var payInEvent PayInEvent
		payInEvent.UserId = userId
		b := ""
		err = rows.Scan(&b, &payInEvent.Currency, &payInEvent.Status,
			&payInEvent.Seats, &payInEvent.Freq, &payInEvent.CreatedAt)
		payInEvent.Balance, _ = new(big.Int).SetString(b, 10)
		if err != nil {
			return nil, err
		}
		payInEvents = append(payInEvents, payInEvent)
	}
	return payInEvents, nil
}

func FindPayInExternal(externalId uuid.UUID, status string) (*PayInEvent, error) {
	var payInEvent PayInEvent
	var b string
	err := db.
		QueryRow(`SELECT user_id, balance, currency, status, seats, freq, created_at 
          FROM payment_in_event 
          WHERE external_id = $1 and status = $2`, externalId, status).
		Scan(&payInEvent.UserId, &b, &payInEvent.Currency, &payInEvent.Status,
			&payInEvent.Seats, &payInEvent.Freq, &payInEvent.CreatedAt)

	b1, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return nil, fmt.Errorf("not a big.int %v", b1)
	}
	payInEvent.Balance = b1
	payInEvent.ExternalId = externalId
	payInEvent.Status = status

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &payInEvent, nil
	default:
		return nil, err
	}
}

func FindSumPaymentByCurrency(userId uuid.UUID, status string) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
               FROM payment_in_event 
               WHERE user_id = $1 AND status = $2
               GROUP BY currency`, userId, status)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func FindLatestDailyPayment(userId uuid.UUID, currency string) (*big.Int, int64, int64, *time.Time, error) {
	var d string
	var seats int64
	var freq int64
	var c time.Time
	err := db.
		QueryRow(`SELECT balance, seats, freq, created_at
               FROM payment_in_event 
               WHERE user_id = $1 AND currency = $2 AND status = $3
               ORDER BY created_at DESC
               LIMIT 1`, userId, currency, PayInSuccess).
		Scan(&d, &seats, &freq, &c)

	var db *big.Int
	if d != "" {
		d1, ok := new(big.Int).SetString(d, 10)
		log.Printf("last payed in balance is %v for currency %v", d1, currency)
		if !ok {
			return nil, 0, 0, nil, fmt.Errorf("not a big.int %v", d1)
		}
		db = new(big.Int).Div(d1, big.NewInt(seats))
		db = new(big.Int).Div(db, big.NewInt(freq))
		log.Printf("daily spending balance is %v", db)
	} else {
		log.Printf("nothing found for userId %v", userId)
		db = big.NewInt(0)
	}

	switch err {
	case sql.ErrNoRows:
		return nil, 0, 0, nil, nil
	case nil:
		return db, seats, freq, &c, nil
	default:
		return nil, 0, 0, nil, err
	}
}

func PaymentSuccess(externalId uuid.UUID, fee *big.Int) error {
	//closes the current cycle and opens a new one, rolls over all currencies
	payInEvent, err := FindPayInExternal(externalId, PayInRequest)
	if err != nil {
		return nil
	}
	payInEvent.Id = uuid.New()
	payInEvent.Status = PayInSuccess
	payInEvent.Balance = new(big.Int).Sub(payInEvent.Balance, fee)
	err = InsertPayInEvent(*payInEvent)
	if err != nil {
		return nil
	}
	//now also store fee
	payInEvent.Id = uuid.New()
	payInEvent.Status = PayInFee
	payInEvent.Balance = fee
	return InsertPayInEvent(*payInEvent)
}

func sumTotalEarnedAmountForContributionIds(contributionIds []uuid.UUID) (*big.Int, error) {
	var c string
	err := db.
		QueryRow(`SELECT COALESCE(SUM(balance), 0) AS c FROM daily_contribution WHERE id = ANY($1)`, pq.Array(contributionIds)).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return big.NewInt(0), nil
	case nil:
		b1, ok := new(big.Int).SetString(c, 10)
		if !ok {
			return big.NewInt(0), fmt.Errorf("not a big.int %v", b1)
		}
		return b1, nil
	default:
		return big.NewInt(0), err
	}
}
