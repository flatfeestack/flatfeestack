package api

import (
	"backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func FakePayment(w http.ResponseWriter, r *http.Request, email string) {
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
	paymentCycleInId := uuid.New()
	err = db.InsertNewPaymentCycleIn(paymentCycleInId, 365, seats, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	ubNew := db.UserBalance{
		PaymentCycleInId: paymentCycleInId,
		UserId:           u.Id,
		Balance:          big.NewInt(2970),
		BalanceType:      "PAY",
		CreatedAt:        utils.TimeNow(),
	}

	err = db.InsertUserBalance(ubNew)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	err = db.UpdatePaymentCycleInId(u.Id, paymentCycleInId)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	return
}

type PayoutInfoDTO struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

// returns current payment cycle that is active (there was exactly one payment for this)
func PaymentCycle(w http.ResponseWriter, _ *http.Request, user *db.User) {
	if user.PaymentCycleInId == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	pc, err := db.FindPaymentCycle(*user.PaymentCycleInId)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not find user balance: %v", err)
		return
	}

	if pc == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pc)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

// calculates the maximum of days that is left with any currency, returns the max with currency
func maxDaysLeft(paymentCycleId uuid.UUID) (string, int64, error) {
	//daily, err := findDailyPaymentByPaymentCycleId(paymentCycleId)
	//if err != nil {
	//	return "", 0, err
	//}
	_, err := db.FindSumUserBalanceByCurrency(paymentCycleId)
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

func StrategyDeductMax(balances map[string]*db.Balance, subs map[string]*big.Int, futSub map[string]*big.Int) (string, int64, *big.Int) {
	var maxBalance *db.Balance
	var maxFreq = int64(0)
	var maxCurrency string

	for currency, balance := range balances {
		newBalance := balance.Balance
		if subs[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, subs[currency])
		}
		if futSub[currency] != nil {
			newBalance = new(big.Int).Sub(newBalance, futSub[currency])
		}
		freq := new(big.Int).Div(newBalance, balance.DailySpending).Int64()
		if freq > 0 {
			if freq > maxFreq {
				maxFreq = freq
				maxBalance = balance
				maxCurrency = currency
			}
		}
	}

	if maxBalance == nil {
		return "N/A", 0, nil
	}
	return maxCurrency, maxFreq, maxBalance.DailySpending
}

func currentUSDBalance(paymentCycleId uuid.UUID) (int64, error) {
	total, err := db.FindSumUserBalanceByCurrency(paymentCycleId)
	if err != nil {
		return 0, err
	}
	if total["USD"] == nil {
		return 0, nil
	}

	f := new(big.Int).Exp(big.NewInt(10), big.NewInt(utils.SupportedCurrencies["USD"].FactorPow), nil)
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

func WsNoAuth(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	conn.CloseHandler()(4001, "Unauthorized")
}

// serveWs handles websocket requests from the peer.
func WebSocket(w http.ResponseWriter, r *http.Request, user *db.User) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("could not upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.PaymentCycleInId == nil {
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

	notifyBrowser(user.Id, *user.PaymentCycleInId)
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

func SendToBrowser(userId uuid.UUID, paymentCycleInId uuid.UUID) error {
	lock.Lock()
	conn := clients[userId]
	lock.Unlock()

	if conn == nil {
		return fmt.Errorf("cannot get websockt for client %v", userId)
	}

	userBalances, err := db.FindUserBalances(userId)
	if err != nil {
		conn.Close()
		return err
	}

	var userBalancesDto []UserBalanceDto
	for _, ub := range userBalances {
		r := UserBalanceDto{
			UserId:           ub.UserId,
			PaymentCycleInId: &ub.PaymentCycleInId,
			BalanceType:      ub.BalanceType,
			Balance:          ub.Balance,
			Currency:         ub.Currency,
			CreatedAt:        ub.CreatedAt,
		}
		userBalancesDto = append(userBalancesDto, r)
	}

	total := map[string]*big.Int{}
	for currency, _ := range utils.SupportedCurrencies {
		total[currency] = big.NewInt(0)
		for _, ub := range userBalancesDto {
			if ub.Currency == currency && ub.Balance != nil {
				total[currency] = new(big.Int).Add(total[currency], ub.Balance)
			}
		}
	}

	var pc db.PaymentCycle
	if utils.IsUUIDZero(paymentCycleInId) {
		pc, err = db.FindPaymentCycleLast(userId)
		if err != nil {
			conn.Close()
			return err
		}
	} else {
		pcp, err := db.FindPaymentCycle(paymentCycleInId)
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

func CancelSub(w http.ResponseWriter, r *http.Request, user *db.User) {
	err := db.UpdateFreq(user.PaymentCycleInId, 0)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not cancel subscription: %v", err)
		return
	}
}

func StatusSponsoredUsers(w http.ResponseWriter, r *http.Request, user *db.User) {
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

func MonthlyPayout(w http.ResponseWriter, r *http.Request, email string) {
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
	/*res, err := payoutRequest(pts, currency)
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
	})*/
	return nil
}

// Closes the current cycle and carries over all open currencies
func closeCycle(uid uuid.UUID, currentPaymentCycleInId uuid.UUID, newPaymentCycleInId uuid.UUID) error {
	if utils.IsUUIDZero(currentPaymentCycleInId) {
		return nil
	}
	currencies, err := db.FindSumUserBalanceByCurrency(currentPaymentCycleInId)
	if err != nil {
		return err
	}

	var ubNew *db.UserBalance
	ubNew = &db.UserBalance{
		PaymentCycleInId: currentPaymentCycleInId,
		UserId:           uid,
		CreatedAt:        utils.TimeNow(),
	}
	for k, currency := range currencies {
		ubNew.Balance = new(big.Int).Neg(currency.Balance)
		ubNew.BalanceType = "CLOSE_CYCLE"
		ubNew.PaymentCycleInId = currentPaymentCycleInId
		ubNew.Currency = k
		ubNew.DailySpending = currency.DailySpending

		err := db.InsertUserBalance(*ubNew)
		if err != nil {
			return err
		}

		if currency.Balance.Cmp(big.NewInt(0)) > 0 {
			ubNew.Balance = currency.Balance
			ubNew.PaymentCycleInId = newPaymentCycleInId
			ubNew.BalanceType = "CARRY_OVER"
			ubNew.Currency = k
			err = db.InsertUserBalance(*ubNew)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func PaymentSuccess(uid uuid.UUID, oldPaymentCycleInId *uuid.UUID, newPaymentCycleInId uuid.UUID, balance *big.Int, currency string, seat int64, freq int64, fee *big.Int) error {
	//closes the current cycle and opens a new one, rolls over all currencies
	if oldPaymentCycleInId != nil {
		err := closeCycle(uid, *oldPaymentCycleInId, newPaymentCycleInId)
		if err != nil {
			return err
		}
	}

	ubNew := db.UserBalance{}
	ubNew.PaymentCycleInId = newPaymentCycleInId
	ubNew.BalanceType = "PAYMENT"
	ubNew.Balance = balance
	ubNew.Currency = currency
	ubNew.UserId = uid
	balanceSub := new(big.Int).Sub(balance, fee)
	ubNew.DailySpending = new(big.Int).Div(balanceSub, big.NewInt(freq*seat))
	err := db.InsertUserBalance(ubNew)
	if err != nil {
		return err
	}

	if fee.Cmp(big.NewInt(0)) > 0 {
		ubNew.BalanceType = "FEE"
		ubNew.Balance = new(big.Int).Neg(fee)
		err = db.InsertUserBalance(ubNew)
		if err != nil {
			return err
		}
	}

	err = db.UpdatePaymentCycleInId(uid, newPaymentCycleInId)
	if err != nil {
		return err
	}

	return nil
}

func notifyBrowser(uid uuid.UUID, paymentCycleId uuid.UUID) {
	go func(uid uuid.UUID, paymentCycleId uuid.UUID) {
		err := SendToBrowser(uid, paymentCycleId)
		if err != nil {
			log.Warnf("could not notify client %v, %v", uid, err)
		}
	}(uid, paymentCycleId)
}
