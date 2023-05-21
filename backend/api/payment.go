package api

import (
	"backend/clients"
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
	Currency string   `json:"currency"`
	Balance  *big.Int `json:"balance"`
}

const (
	UserBalancesError = "Oops something went wrong with retrieving user balances. Please try again."
	PaymentError      = "Oops something went wrong with the payment. Please try again."
	PayoutError       = "Oops something went wrong with the payout. Please try again."
)

func FakePayment(w http.ResponseWriter, r *http.Request, _ string) {
	log.Printf("fake payment")
	m := mux.Vars(r)
	n := m["email"]
	s := m["seats"]

	u, err := db.FindUserByEmail(n)
	if err != nil {
		log.Errorf("Error while finding user by email: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	seats, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Errorf("Error while parsing int: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
		log.Errorf("Error with inserting pay in event: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	err = db.PaymentSuccess(e, big.NewInt(1))
	if err != nil {
		log.Errorf("Error with payment success: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PaymentError)
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
		log.Printf("newB %v / ds %v", newBalance, ds)
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
		log.Errorf("Could not cancel subscription: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong while canceling the subscription. Please try again.")
		return
	}
}

func StatusSponsoredUsers(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	userStatus, err := db.FindSponsoredUserBalances(user.Id)
	if err != nil {
		log.Errorf("Error while finding sponsored user balances: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	err = json.NewEncoder(w).Encode(userStatus)
	if err != nil {
		log.Errorf("Could not encode json: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func RequestPayout(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	m := mux.Vars(r)
	targetCurrency := m["targetCurrency"]

	currencyMetadata, ok := utils.SupportedCurrencies[targetCurrency]
	if !ok {
		log.Errorf("Unsupported currency requested")
		utils.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong with the selected currency. Please try again.")
		return
	}

	ownContributionIds, err := db.FindOwnContributionIds(user.Id, targetCurrency)
	if err != nil {
		log.Errorf("Error while trying to retrieve own contribution ids: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	}

	totalEarnedAmount, err := db.SumTotalEarnedAmountForContributionIds(ownContributionIds)
	if err != nil {
		log.Errorf("Unable to retrieve already earned amount in target currency: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	}

	if targetCurrency == "USD" {
		// For USDC, 10^18 units are one dollar
		// See explorer https://explorer.bitquery.io/bsc/token/0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d
		// And https://docs.openzeppelin.com/contracts/4.x/erc20#a-note-on-decimals
		// FlatFeeStack already calculates in micro dollars, so we need to blow up the value a bit
		usdcDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18-currencyMetadata.FactorPow), nil)

		// Divide the float by the power of 10
		totalEarnedAmount.Mul(totalEarnedAmount, usdcDecimals)
	}

	signature, err := clients.RequestPayout(user.Id, totalEarnedAmount, currencyMetadata.PayoutName)
	if err != nil {
		log.Errorf("Error when generating signature: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	}

	err = db.MarkContributionAsClaimed(ownContributionIds)
	if err != nil {
		log.Errorf("Error when marking contributions as claimed: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PayoutError)
		return
	} else {
		utils.WriteJson(w, signature)
	}
}

func PaymentEvent(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	payInEvents, err := db.FindPayInUser(user.Id)
	if err != nil {
		log.Errorf("Error while finding pay in user: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, PaymentError)
		return
	}

	err = json.NewEncoder(w).Encode(payInEvents)
	if err != nil {
		log.Errorf("Could not encode json: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func UserBalance(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	mAdd, err := db.FindSumPaymentByCurrency(user.Id, db.PayInSuccess)
	if err != nil {
		log.Errorf("Error while finding sum payments by currency: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	//either the user spent it on a repo that does not have any devs who can claim
	mFut, err := db.FindSumFutureSponsors(user.Id)
	if err != nil {
		log.Errorf("Error while finding sum future sponsors: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}
	//or the user spent it on for a repo with a dev who can claim
	mSub, err := db.FindSumDailySponsors(user.Id)
	if err != nil {
		log.Errorf("Error while finding sum daily sponsors: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, UserBalancesError)
		return
	}

	totalUserBalances := []TotalUserBalance{}
	for currency, _ := range mAdd {
		if mSub[currency] != nil {
			mAdd[currency] = new(big.Int).Sub(mAdd[currency], mSub[currency])
			mSub[currency] = big.NewInt(0)
		}
		if mFut[currency] != nil {
			mAdd[currency] = new(big.Int).Sub(mAdd[currency], mFut[currency])
			mFut[currency] = big.NewInt(0)
		}
		//
		totalUserBalances = append(totalUserBalances, TotalUserBalance{
			Currency: currency,
			Balance:  mAdd[currency],
		})
	}

	//now sanity check if all the deducted mSub/mFut are set to zero
	for currency, balance := range mSub {
		if balance.Cmp(big.NewInt(0)) != 0 {
			log.Errorf("Something is off with: %v", currency)
			utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}
	for currency, balance := range mFut {
		if balance.Cmp(big.NewInt(0)) != 0 {
			log.Errorf("Something is off with: %v", currency)
			utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}

	err = json.NewEncoder(w).Encode(totalUserBalances)
	if err != nil {
		log.Errorf("Could not encode json: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

}
