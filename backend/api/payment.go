package api

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

type PayoutInfoDTO struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

type UserBalanceDto struct {
	UserId           uuid.UUID  `json:"userId"`
	Balance          *big.Int   `json:"balance"`
	PaymentCycleInId *uuid.UUID `json:"paymentCycleInId"`
	BalanceType      string     `json:"balanceType"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type UserBalances struct {
	PaymentCycle db.PaymentCycle     `json:"paymentCycle"`
	UserBalances []UserBalanceDto    `json:"userBalances"`
	Total        map[string]*big.Int `json:"total"`
	DaysLeft     int64               `json:"daysLeft"`
}

type TotalUserBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func FakePayment(w http.ResponseWriter, r *http.Request, _ string) {
	log.Printf("fake payment")
	m := mux.Vars(r)
	n := m["email"]
	s := m["seats"]

	u, err := db.FindUserByEmail(n)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	seats, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	e := uuid.New()
	payInEvent := db.PayInEvent{
		Id:         uuid.New(),
		ExternalId: e,
		UserId:     u.Id,
		Balance:    big.NewInt(Plans[0].PriceBase),
		Currency:   "USD",
		Status:     db.PayInRequest,
		Seats:      seats,
		Freq:       365,
		CreatedAt:  time.Time{},
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	err = PaymentSuccess(e, big.NewInt(1))
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	return
}

func StrategyDeductMax(userId uuid.UUID, balances map[string]*big.Int, subs map[string]*big.Int, futSub map[string]*big.Int) (string, int64, *big.Int, error) {
	var maxDailySpending *big.Int
	var maxFreq = int64(0)
	var maxCurrency string

	//
	for currency, newBalance := range balances {
		if subs[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, subs[currency])
		}
		if futSub[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, futSub[currency])
		}

		ds, _, _, _, err := db.FindLatestDailyPayment(userId, currency)
		if err != nil {
			return "N/A", 0, nil, err
		}
		freq := new(big.Int).Div(newBalance, ds).Int64()
		if freq > 0 {
			if freq > maxFreq {
				maxFreq = freq
				maxDailySpending = ds
				maxCurrency = currency
			}
		}
	}

	if maxDailySpending == nil {
		return "N/A", 0, nil, nil
	}
	return maxCurrency, maxFreq, maxDailySpending, nil
}

func CancelSub(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	err := db.UpdateSeatsFreq(user.Id, user.Seats, 0)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not cancel subscription: %v", err)
		return
	}
}

func StatusSponsoredUsers(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	userStatus, err := db.FindSponsoredUserBalances(user.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(userStatus)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func PaymentSuccess(externalId uuid.UUID, fee *big.Int) error {
	//closes the current cycle and opens a new one, rolls over all currencies
	payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
	if err != nil {
		return nil
	}
	payInEvent.Status = db.PayInSuccess
	payInEvent.Balance = new(big.Int).Sub(payInEvent.Balance, fee)
	err = db.InsertPayInEvent(*payInEvent)
	if err != nil {
		return nil
	}
	//now also store fee
	payInEvent.Status = db.PayInFee
	payInEvent.Balance = fee
	err = db.InsertPayInEvent(*payInEvent)
	if err != nil {
		return nil
	}

	return nil
}
