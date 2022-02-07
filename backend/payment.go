package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type PayoutInfoDTO struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

//returns current payment cycle that is active (there was exactly one payment for this)
func paymentCycle(w http.ResponseWriter, _ *http.Request, user *User) {
	pc, err := findPaymentCycle(user.PaymentCycleInId)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not find user balance: %v", err)
		return
	}

	if pc == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pc)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

//calculates the maximum of days that is left with any currency, returns the max with currency
func maxDaysLeft(paymentCycleId *uuid.UUID) (string, int64, error) {
	//daily, err := findDailyPaymentByPaymentCycleId(paymentCycleId)
	//if err != nil {
	//	return "", 0, err
	//}
	_, err := findSumUserBalanceByCurrency(paymentCycleId)
	if err != nil {
		return "", 0, err
	}

	max := int64(0)
	cur := ""
	/*for currency, balance := range balances {
		if daily[currency] != nil {
			d := new(big.Int).Div(balance.Balance, daily[currency])
			if d.Int64() > max {
				max = d.Int64()
				cur = currency
			}
		}
	}*/
	return cur, max, nil
}

func strategyDeductMax(balances map[string]*Balance, subs map[string]*Balance) (string, int64, *big.Int, error) {
	var maxBalance *Balance
	var maxFreq = int64(0)
	var maxCurrency string

	for currency, balance := range balances {
		newBalance := balance.Balance
		if subs[currency] != nil {
			newBalance = new(big.Int).Sub(balance.Balance, subs[currency].Balance)
		}
		freq := new(big.Int).Div(newBalance, balance.Split).Int64()
		if freq > 0 {
			if freq > maxFreq {
				maxFreq = freq
				maxBalance = balance
				maxCurrency = currency
			}
		}
	}

	if maxBalance != nil {
		return maxCurrency, maxFreq, maxBalance.Split, nil
	}

	return "", 0, nil, fmt.Errorf("not enough balance %v, %v", balances, subs)
}

func currentUSDBalance(paymentCycleId *uuid.UUID) (int64, error) {
	total, err := findSumUserBalanceByCurrency(paymentCycleId)
	if err != nil {
		return 0, err
	}
	if total["USD"] == nil {
		return 0, nil
	}

	f := new(big.Int).Exp(big.NewInt(10), big.NewInt(supportedCurrencies["USD"].FactorPow), nil)
	t := new(big.Int).Div(total["USD"].Balance, f)
	return t.Int64(), nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"access_token"},
}

var clients = make(map[uuid.UUID]*websocket.Conn)
var lock = sync.Mutex{}

func wsNoAuth(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	conn.CloseHandler()(4001, "Unauthorized")
}

// serveWs handles websocket requests from the peer.
func ws(w http.ResponseWriter, r *http.Request, user *User) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("could not upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lock.Lock()
	clients[user.Id] = conn
	lock.Unlock()
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("closing connection for %v", user.Id)
		lock.Lock()
		delete(clients, user.Id)
		lock.Unlock()
		return nil
	})

	conn.SetPongHandler(func(appData string) error {
		log.Printf(appData)
		return nil
	})

	notifyBrowser(user.Id, user.PaymentCycleInId)
}

type UserBalanceDto struct {
	UserId           uuid.UUID  `json:"userId"`
	Balance          *big.Int   `json:"balance"`
	PaymentCycleInId *uuid.UUID `json:"paymentCycleInId"`
	FromUserId       *uuid.UUID `json:"fromUserId"`
	BalanceType      string     `json:"balanceType"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type UserBalances struct {
	PaymentCycle PaymentCycle        `json:"paymentCycle"`
	UserBalances []UserBalanceDto    `json:"userBalances"`
	Total        map[string]*big.Int `json:"total"`
	DaysLeft     int64               `json:"daysLeft"`
}

type TotalUserBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func sendToBrowser(userId uuid.UUID, paymentCycleInId *uuid.UUID) error {
	lock.Lock()
	conn := clients[userId]
	lock.Unlock()

	if conn == nil {
		return fmt.Errorf("cannot get websockt for client %v", userId)
	}

	userBalances, err := findUserBalances(userId)
	if err != nil {
		conn.Close()
		return err
	}

	var userBalancesDto []UserBalanceDto
	for _, ub := range userBalances {
		r := UserBalanceDto{
			UserId:           ub.UserId,
			PaymentCycleInId: ub.PaymentCycleInId,
			FromUserId:       ub.FromUserId,
			BalanceType:      ub.BalanceType,
			Balance:          ub.Balance,
			Currency:         ub.Currency,
			CreatedAt:        ub.CreatedAt,
		}
		userBalancesDto = append(userBalancesDto, r)
	}

	total := map[string]*big.Int{}
	for currency, _ := range supportedCurrencies {
		total[currency] = big.NewInt(0)
		for _, ub := range userBalancesDto {
			if ub.Currency == currency && ub.Balance != nil {
				total[currency] = new(big.Int).Add(total[currency], ub.Balance)
			}
		}
	}

	var pc PaymentCycle
	if isUUIDZero(paymentCycleInId) {
		pc, err = findPaymentCycleLast(userId)
		if err != nil {
			conn.Close()
			return err
		}
	} else {
		pcp, err := findPaymentCycle(paymentCycleInId)
		if err != nil {
			conn.Close()
			return err
		}
		pc = *pcp
	}

	_, daysLeft, err := maxDaysLeft(paymentCycleInId)
	if err != nil {
		conn.Close()
		return err
	}

	err = conn.WriteJSON(UserBalances{PaymentCycle: pc, UserBalances: userBalancesDto, Total: total, DaysLeft: daysLeft})
	if err != nil {
		conn.Close()
		return err
	}

	return nil
}

func cancelSub(w http.ResponseWriter, r *http.Request, user *User) {
	err := updateFreq(user.PaymentCycleInId, 0)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not cancel subscription: %v", err)
		return
	}
}

func statusSponsoredUsers(w http.ResponseWriter, r *http.Request, user *User) {
	userStatus, err := findSponsoredUserBalances(user.Id)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(userStatus)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func getPayoutInfos(w http.ResponseWriter, r *http.Request, email string) {
	infos, err := findPayoutInfos()
	var result []PayoutInfoDTO
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, err.Error())
		log.Printf("Could not find payout infos: %v", err)
		return
	}
	for _, v := range infos {
		r := PayoutInfoDTO{Currency: v.Currency, Amount: v.Amount}
		result = append(result, r)
	}
	writeJson(w, result)
}

func monthlyPayout(w http.ResponseWriter, r *http.Request, email string) {
	//TODO: enable again
	/*chunkSize := 100
	var container = make([][]PayoutCrypto, len(supportedCurrencies))
	var usdPayouts []PayoutCrypto

	m := mux.Vars(r)
	h := m["exchangeRate"]
	if h == "" {
		writeErrorf(w, http.StatusBadRequest, "Parameter exchangeRate not set: %v", m)
		return
	}

	exchangeRate, _, err := big.ParseFloat(h, 10, 128, big.ToZero)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Parameter exchangeRate not set: %v", m)
		return
	}

	payouts, err := findMonthlyBatchJobPayout()
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, err.Error())
		return
	}

	// group container by currency [[eth], [neo], [tez]]
	for _, payout := range payouts {
		if payout.Currency == "USD" {
			usdPayouts = append(usdPayouts, payout)
			continue
		}
		for i, currency := range supportedCurrencies {
			if payout.Currency == currency.ShortName {
				container[i] = append(container[i], payout)
			}
		}
	}

	for _, usdPayout := range usdPayouts {
		for i := 0; i < len(container[0]); i++ { // ETH at position 1
			if usdPayout.Address == container[0][i].Address {
				e, _ := exchangeRate.Float64()
				container[0][i].Meta = append(container[0][i].Meta, PayoutMeta{Currency: "USD", Tea: usdPayout.Tea})
				container[0][i].Meta = append(container[0][i].Meta, PayoutMeta{Currency: "ETH", Tea: container[0][i].Tea})
				usdInEth := float64(usdPayout.Tea) / float64(usdFactor) * e * cryptoFactor
				container[0][i].Tea += int64(usdInEth)
			}
		}
	}

	for _, payouts := range container {
		if len(payouts) <= 0 {
			continue
		}
		currency := payouts[0].Currency

		for i := 0; i < len(payouts); i += chunkSize {
			end := i + chunkSize
			if end > len(payouts) {
				end = len(payouts)
			}
			var pts []PayoutToService
			batchId := uuid.New()
			for _, payout := range payouts[i:end] {
				request := PayoutRequest{
					UserId:    payout.UserId,
					BatchId:   batchId,
					Currency:  currency,
					Tea:       payout.Tea,
					Address:   payout.Address,
					CreatedAt: timeNow(),
				}

				err := insertPayoutRequest(&request)
				if err != nil {
					writeErrorf(w, http.StatusInternalServerError, err.Error())
					return
				}

				pt := PayoutToService{
					Address: payout.Address,
					Tea:     payout.Tea,
					Meta:    payout.Meta,
				}

				pts = append(pts, pt)
			}
			err := cryptoPayout(pts, batchId, currency)
			if err != nil {
				writeErrorf(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}*/
}

// Helper
func cryptoPayout(pts []PayoutToService, batchId uuid.UUID, currency string) error {
	res, err := payoutRequest(pts, currency)
	if err != nil {
		err1 := err.Error()
		err2 := insertPayoutResponse(&PayoutsResponse{
			BatchId:   batchId,
			Error:     &err1,
			CreatedAt: timeNow(),
		})
		return fmt.Errorf("error %v/%v", err, err2)
	}

	res.Currency = currency
	return insertPayoutResponse(&PayoutsResponse{
		BatchId:   batchId,
		TxHash:    res.TxHash,
		Error:     nil,
		CreatedAt: timeNow(),
		Payouts:   *res,
	})
}

//Closes the current cycle and carries over all open currencies
func closeCycle(uid uuid.UUID, currentPaymentCycleInId *uuid.UUID, newPaymentCycleInId *uuid.UUID) error {
	if isUUIDZero(currentPaymentCycleInId) {
		return nil
	}
	currencies, err := findSumUserBalanceByCurrency(currentPaymentCycleInId)
	if err != nil {
		return err
	}

	var ubNew *UserBalance
	ubNew = &UserBalance{
		PaymentCycleInId: currentPaymentCycleInId,
		UserId:           uid,
		CreatedAt:        timeNow(),
	}
	for k, currency := range currencies {
		ubNew.Balance = new(big.Int).Neg(currency.Balance)
		ubNew.BalanceType = "CLOSE_CYCLE"
		ubNew.PaymentCycleInId = currentPaymentCycleInId
		ubNew.Currency = k
		ubNew.Split = currency.Split

		err := insertUserBalance(*ubNew)
		if err != nil {
			return err
		}

		if currency.Balance.Cmp(big.NewInt(0)) > 0 {
			ubNew.Balance = currency.Balance
			ubNew.PaymentCycleInId = newPaymentCycleInId
			ubNew.BalanceType = "CARRY_OVER"
			ubNew.Currency = k
			err = insertUserBalance(*ubNew)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func paymentSuccess(uid uuid.UUID, oldPaymentCycleInId *uuid.UUID, newPaymentCycleInId *uuid.UUID, balance *big.Int, currency string, seat int64, freq int64, fee *big.Int) error {
	//closes the current cycle and opens a new one, rolls over all currencies
	err := closeCycle(uid, oldPaymentCycleInId, newPaymentCycleInId)
	if err != nil {
		return err
	}

	ubNew := UserBalance{}
	ubNew.PaymentCycleInId = newPaymentCycleInId
	ubNew.BalanceType = "PAYMENT"
	ubNew.Balance = balance
	ubNew.Currency = currency
	ubNew.UserId = uid
	balanceSub := new(big.Int).Sub(balance, fee)
	ubNew.Split = new(big.Int).Div(balanceSub, big.NewInt(freq*seat))
	err = insertUserBalance(ubNew)
	if err != nil {
		return err
	}

	if fee.Cmp(big.NewInt(0)) > 0 {
		ubNew.BalanceType = "FEE"
		ubNew.Balance = new(big.Int).Neg(fee)
		err = insertUserBalance(ubNew)
		if err != nil {
			return err
		}
	}

	err = updatePaymentCycleInId(uid, newPaymentCycleInId)
	if err != nil {
		return err
	}

	return nil
}

func notifyBrowser(uid uuid.UUID, paymentCycleId *uuid.UUID) {
	go func(uid uuid.UUID, paymentCycleId *uuid.UUID) {
		err := sendToBrowser(uid, paymentCycleId)
		if err != nil {
			log.Warnf("could not notify client %v, %v", uid, err)
		}
	}(uid, paymentCycleId)
}
