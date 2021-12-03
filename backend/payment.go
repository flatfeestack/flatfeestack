package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/setupintent"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ClientSecretBody struct {
	ClientSecret string `json:"client_secret"`
}

type PayoutInfoDTO struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

func paymentCycle(w http.ResponseWriter, r *http.Request, user *User) {
	pc, err := findPaymentCycle(user.PaymentCycleId)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not find user balance: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(pc)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func topup(w http.ResponseWriter, r *http.Request, user *User) {
	if len(user.Claims.InviteEmails) == 0 {
		log.Printf("no invitations")
		return
	}

	pc, err := findPaymentCycle(user.PaymentCycleId)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not find user balance: %v", err)
		return
	}

	if pc != nil && pc.DaysLeft > 0 {
		log.Printf("enough funds")
		return
	}

	for k, inviteEmail := range user.Claims.InviteEmails {
		freq, err := strconv.Atoi(user.Claims.InviteMeta[k])
		if err != nil {
			log.Printf("findSumUserBalances: %v", err)
			continue
		}
		ok, paymentCycleId := topupWithSponsor(user, freq, inviteEmail)
		if !ok {
			continue
		}

		go func() {
			err = sendToBrowser(user.Id, *paymentCycleId)
			if err != nil {
				log.Printf("could not notify client %v", user.Id)
			}
		}()

		break
	}
}

func topupWithSponsor(u *User, freq int, inviteEmail string) (bool, *uuid.UUID) {
	sponsor, err := findUserByEmail(inviteEmail)
	if err != nil {
		log.Printf("findUserByEmail: %v", err)
		return false, nil
	}

	//parent has enough funds go for it!
	parentBalances, err := findSumUserBalances(sponsor.Id, sponsor.PaymentCycleId)
	if err != nil {
		log.Printf("findSumUserBalances: %v", err)
		return false, nil
	}

	dailyPayments, err := findDailyPaymentByPaymentCycleId(sponsor.PaymentCycleId)
	if err != nil {
		log.Printf("dailyPayment: %v", err)
		return false, nil
	}

	var currency string
	var balance int64
	var dailyPaymentAmount int64
	parentHasEnoughFunds := false
	for _, parentBalance := range parentBalances {
		if parentHasEnoughFunds {
			break
		}
		for _, dailyPayment := range dailyPayments {
			if parentBalance.Currency == dailyPayment.Currency {
				tempAmount := int64(freq) * dailyPayment.Amount
				if parentBalance.Balance > tempAmount {
					currency = parentBalance.Currency
					balance = tempAmount
					dailyPaymentAmount = dailyPayment.Amount
					parentHasEnoughFunds = true
				}
			}
		}
	}

	if !parentHasEnoughFunds {
		log.Printf("parent has not enough funding")
		//TODO: notify parent
		return false, nil
	}

	newPaymentCycleId, err := insertNewPaymentCycle(u.Id, freq, 1, freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", u.Id, err)
		return false, nil
	}
	dp := DailyPayment{PaymentCycleId: *newPaymentCycleId, Currency: currency, Amount: dailyPaymentAmount, DaysLeft: freq, LastUpdate: timeNow()}
	err = insertDailyPayment(dp)
	if err != nil {
		log.Printf("dailyPayment: %v", err)
		return false, nil
	}

	ubNew, err := closeCycle(u.Id, u.PaymentCycleId, *newPaymentCycleId, currency)
	if err != nil {
		return false, nil
	}

	ubNew.PaymentCycleId = sponsor.PaymentCycleId
	ubNew.UserId = sponsor.Id
	ubNew.Balance = -balance
	ubNew.BalanceType = "SPONSOR"
	ubNew.Currency = currency
	err = insertUserBalance(*ubNew)
	if err != nil {
		log.Printf("transferBalance: %v", err)
		return false, nil
	}

	ubNew.UserId = u.Id
	ubNew.Balance = balance
	ubNew.PaymentCycleId = *newPaymentCycleId
	ubNew.FromUserId = &sponsor.Id
	err = insertUserBalance(*ubNew)
	if err != nil {
		log.Printf("transferBalance: %v", err)
		return false, nil
	}
	err = updatePaymentCycleId(u.Id, newPaymentCycleId, &sponsor.Id)
	if err != nil {
		log.Printf("transferBalance: %v", err)
		return false, nil
	}

	return true, newPaymentCycleId
}

//https://stripe.com/docs/payments/save-and-reuse
func setupStripe(w http.ResponseWriter, r *http.Request, user *User) {
	//create a user at stripe if the user does not exist yet
	if user.StripeId == nil || opts.StripeAPISecretKey != "" {
		stripe.Key = opts.StripeAPISecretKey
		params := &stripe.CustomerParams{}
		c, err := customer.New(params)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
			return
		}
		user.StripeId = &c.ID
		err = updateUser(user)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
			return
		}
	}

	usage := string(stripe.SetupIntentUsageOnSession)
	params := &stripe.SetupIntentParams{
		Customer: user.StripeId,
		Usage:    &usage,
	}
	intent, err := setupintent.New(params)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func stripePaymentInitial(w http.ResponseWriter, r *http.Request, user *User) {

	if user.PaymentMethod == nil {
		writeErr(w, http.StatusInternalServerError, "No payment method defined for user: %v", user.Id)
		return
	}

	p := mux.Vars(r)
	f := p["freq"]
	s := p["seats"]
	seats, err := strconv.Atoi(s)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Cannot convert number seats: %v", seats)
		return
	}

	freq, err := strconv.Atoi(f)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Cannot convert number freq: %v", seats)
		return
	}
	var plan *Plan
	for _, v := range plans {
		if v.Freq == freq {
			plan = &v
			break
		}
	}
	if plan == nil {
		writeErr(w, http.StatusInternalServerError, "No matching plan found: %v, available: %v", freq, plans)
		return
	}

	var f1 big.Float
	f1.SetString("100")
	cents, _ := f1.Mul(&plan.Price, &f1).Int64()
	cents = cents * int64(seats)

	//cents := stripe.Int64(int64(plan.Price.Int64()))
	params := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(cents),
		Currency:         stripe.String(string(stripe.CurrencyUSD)),
		Customer:         user.StripeId,
		PaymentMethod:    user.PaymentMethod,
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	paymentCycleId, err := insertNewPaymentCycle(user.Id, 0, seats, freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["paymentCycleId"] = paymentCycleId.String()
	params.Params.Metadata["fee"] = strconv.Itoa(plan.FeePrm)
	params.Params.Metadata["freq"] = strconv.Itoa(freq)

	intent, err := paymentintent.New(params)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func stripePaymentRecurring(user User) (*ClientSecretBody, error) {
	pc, err := findPaymentCycle(user.PaymentCycleId)
	if err != nil {
		return nil, err
	}

	var plan *Plan
	for _, v := range plans {
		if v.Freq == pc.Freq {
			plan = &v
			break
		}
	}
	if plan == nil {
		return nil, fmt.Errorf("no matching plan found: %v, available: %v", pc.Freq, plans)
	}

	var f1 big.Float
	f1.SetString("100")
	cents, _ := f1.Mul(&plan.Price, &f1).Int64()
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(cents * int64(pc.Seats)),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
	}

	paymentCycleId, err := insertNewPaymentCycle(user.Id, pc.Freq, pc.Seats, pc.Freq, timeNow())
	if err != nil {
		return nil, err
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["paymentCycleId"] = paymentCycleId.String()
	params.Params.Metadata["fee"] = strconv.Itoa(plan.FeePrm)
	params.Params.Metadata["freq"] = strconv.Itoa(plan.Freq)

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	return &cs, nil
}

func stripeWebhook(w http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), opts.StripeWebhookSecretKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		log.Printf("Error evaluating signed webhook request: %v", err)
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":

		uid, newPaymentCycleId, amount, feePrm, freq, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		u, err := findUserById(uid)
		if err != nil || u == nil {
			log.Printf("User does not exist: %v, %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// amount(cent) * 10000(mUSD) * feePrm / 1000
		fee := (amount * int64(feePrm) / 1000) + 1 //round up

		err = paymentSuccess(u, newPaymentCycleId, amount*10000, "USD", freq, fee*10000)
		if err != nil {
			log.Printf("User sum balance cann run for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	// ... handle other event types
	case "payment_intent.authentication_required":
	case "payment_intent.requires_payment_method":
		//again
		fmt.Printf("CASE-STRIP: %v", event)
		uid, newPaymentCycleId, _, _, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		u, err := findUserById(uid)
		if err != nil || u == nil {
			log.Printf("User does not exist: %v, %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ub, err := findUserBalancesAndType(newPaymentCycleId, "AUTHREQ", "USD")
		if err != nil {
			log.Printf("Error find user balance: %v, %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}

		ubNew := UserBalance{
			PaymentCycleId: newPaymentCycleId,
			UserId:         uid,
			Balance:        0,
			BalanceType:    "AUTHREQ",
			Currency:       "USD",
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := u.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/user/payments"
		other["lang"] = "en"

		defaultMessage := "Authentication is required, please go to the following site to continue: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-authreq_", "Authentication requested",
			"template-plain-authreq_", defaultMessage,
			"template-html-authreq_", other["lang"])

		go func() {
			insertEmailSent(u.Id, "authreq-"+newPaymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	//case "payment_intent.requires_action":
	//3d secure - this is handled by strip, we just get notified
	//	log.Printf("stripe handles 3d secure for %v", event.Data)
	case "payment_intent.insufficient_funds":
		uid, newPaymentCycleId, _, _, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		u, err := findUserById(uid)
		if err != nil || u == nil {
			log.Printf("User does not exist: %v, %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ub, err := findUserBalancesAndType(newPaymentCycleId, "NOFUNDS", "USD")
		if err != nil {
			log.Printf("Error find user balance: %v, %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}

		ubNew := UserBalance{
			PaymentCycleId: newPaymentCycleId,
			UserId:         uid,
			Balance:        0,
			BalanceType:    "NOFUNDS",
			Currency:       "USD",
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := u.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/dashboard/profile"
		other["lang"] = "en"

		defaultMessage := "Your credit card does not have sufficient funds. If you have enough funds, please go to: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-nofund_", "Insufficient funds",
			"template-plain-nofund_", defaultMessage,
			"template-html-nofund_", other["lang"])

		go func() {
			insertEmailSent(u.Id, "nofund-"+newPaymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()

	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
	}
}

func parseStripeData(data json.RawMessage) (uuid.UUID, uuid.UUID, int64, int, int, error) {
	var pi stripe.PaymentIntent
	err := json.Unmarshal(data, &pi)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, 0, 0, fmt.Errorf("Error parsing webhook JSON: %v\n", err)
	}
	uidRaw := pi.Metadata["userId"]
	uid, err := uuid.Parse(uidRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, 0, 0, fmt.Errorf("Error parsing uid: %v, available %v\n", err, pi.Metadata)
	}
	newPaymentCycleIdRaw := pi.Metadata["paymentCycleId"]
	newPaymentCycleId, err := uuid.Parse(newPaymentCycleIdRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, 0, 0, fmt.Errorf("Error parsing newPaymentCycleId: %v, available %v\n", err, pi.Metadata)
	}
	feePrm, err := strconv.Atoi(pi.Metadata["fee"])
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, 0, 0, fmt.Errorf("Error parsing fee: %v, available %v\n", pi.Metadata["fee"], pi.Metadata)
	}

	freq, err := strconv.Atoi(pi.Metadata["freq"])
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, 0, 0, fmt.Errorf("Error parsing freq: %v, available %v\n", pi.Metadata["freq"], pi.Metadata)
	}
	return uid, newPaymentCycleId, pi.Amount, feePrm, freq, nil
}

var upgrader = websocket.Upgrader{
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
		log.Printf("closing")
		lock.Lock()
		delete(clients, user.Id)
		lock.Unlock()
		return nil
	})

	conn.SetPongHandler(func(appData string) error {
		log.Printf(appData)
		return nil
	})

	go func() {
		err = sendToBrowser(user.Id, user.PaymentCycleId)
		if err != nil {
			log.Printf("could not notify client %v", user.Id)
		}
	}()
}

type UserBalances struct {
	PaymentCycle PaymentCycle  `json:"paymentCycle"`
	UserBalances []UserBalance `json:"userBalances"`
	Total        int64         `json:"total"`
	DaysLeft     int           `json:"daysLeft"`
}

func sendToBrowser(userId uuid.UUID, paymentCycleId uuid.UUID) error {
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

	total := int64(0)
	for _, v := range userBalances {
		total += v.Balance
	}

	pc, err := findPaymentCycle(paymentCycleId)
	if err != nil {
		conn.Close()
		return err
	}

	if pc == nil {
		return nil //nothing to do
	}

	err = conn.WriteJSON(UserBalances{PaymentCycle: *pc, UserBalances: userBalances, Total: total, DaysLeft: pc.DaysLeft})
	if err != nil {
		conn.Close()
		return err
	}

	return nil
}

func cancelSub(w http.ResponseWriter, r *http.Request, user *User) {
	err := updateFreq(user.PaymentCycleId, 0)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not cancel subscription: %v", err)
		return
	}
}

func statusSponsoredUsers(w http.ResponseWriter, r *http.Request, user *User) {
	userStatus, err := findSponsoredUserBalances(user.Id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(userStatus)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func getPayoutInfos(w http.ResponseWriter, r *http.Request, email string) {
	infos, err := findPayoutInfos()
	var result []PayoutInfoDTO
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		log.Printf("Could not find payout infos: %v", err)
		return
	}
	for _, v := range infos {
		r := PayoutInfoDTO{Currency: v.Currency}
		if v.Currency == "USD" {
			r.Amount = float64(v.Amount) / usdFactor
		} else {
			r.Amount = float64(v.Amount) / cryptoFactor
		}
		result = append(result, r)
	}
	writeJson(w, result)
}

func monthlyPayout(w http.ResponseWriter, r *http.Request, email string) {
	chunkSize := 100
	var container = make([][]PayoutCrypto, len(supportedCurrencies))
	var usdPayouts []PayoutCrypto

	m := mux.Vars(r)
	h := m["exchangeRate"]
	if h == "" {
		writeErr(w, http.StatusBadRequest, "Parameter exchangeRate not set: %v", m)
		return
	}

	exchangeRate, _, err := big.ParseFloat(h, 10, 128, big.ToZero)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Parameter exchangeRate not set: %v", m)
		return
	}

	payouts, err := findMonthlyBatchJobPayout()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
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
					writeErr(w, http.StatusInternalServerError, err.Error())
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
				writeErr(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}
}

// Helper
func cryptoPayout(pts []PayoutToService, batchId uuid.UUID, currency string) error {
	res, err := payoutRequest(pts, currency)
	res.Currency = currency

	if err != nil {
		err1 := err.Error()
		err2 := insertPayoutResponse(&PayoutsResponse{
			BatchId:   batchId,
			Error:     &err1,
			CreatedAt: timeNow(),
		})
		return fmt.Errorf("error %v/%v", err, err2)
	}
	return insertPayoutResponse(&PayoutsResponse{
		BatchId:   batchId,
		TxHash:    res.TxHash,
		Error:     nil,
		CreatedAt: timeNow(),
		Payouts:   *res,
	})
}

func closeCycle(uid uuid.UUID, oldPaymentCycleId uuid.UUID, newPaymentCycleId uuid.UUID, currency string) (*UserBalance, error) {
	currencies, err := findAllCurrenciesFromUserBalance(oldPaymentCycleId)
	if err != nil {
		return nil, err
	}
	if !contains(currencies, currency) {
		currencies = append(currencies, currency)
	}

	var ubNew *UserBalance
	for _, currency := range currencies {
		oldSum, err := findSumUserBalanceByCurrency(uid, oldPaymentCycleId, currency)
		if err != nil {
			return nil, err
		}

		ubNew = &UserBalance{
			PaymentCycleId: oldPaymentCycleId,
			UserId:         uid,
			Balance:        -oldSum,
			BalanceType:    "CLOSE_CYCLE",
			Currency:       currency,
			CreatedAt:      timeNow(),
		}

		if oldSum > 0 {
			err := insertUserBalance(*ubNew)
			if err != nil {
				return nil, err
			}
			ubNew.Balance = oldSum
			ubNew.PaymentCycleId = newPaymentCycleId
			ubNew.BalanceType = "CARRY_OVER"
			ubNew.Currency = currency
			err = insertUserBalance(*ubNew)
			if err != nil {
				return nil, err
			}
		}
	}

	return ubNew, nil
}

func paymentSuccess(u *User, newPaymentCycleId uuid.UUID, amount int64, currency string, freq int, fee int64) error {
	ub, err := findUserBalancesAndType(newPaymentCycleId, "PAYMENT", currency)
	if err != nil {
		return err
	}
	if ub != nil {
		log.Printf("We already processed this event, we can safely ignore it: %v", ub)
		return nil
	}

	pc, err := findPaymentCycle(newPaymentCycleId)
	if err != nil {
		log.Printf("Payment Cycle not found: %v", err)
		return nil
	}

	ubNew, err := closeCycle(u.Id, u.PaymentCycleId, newPaymentCycleId, currency)
	if err != nil {
		return err
	}

	ubNew.PaymentCycleId = newPaymentCycleId
	ubNew.BalanceType = "PAYMENT"
	ubNew.Balance = amount
	ubNew.Currency = currency
	err = insertUserBalance(*ubNew)
	if err != nil {
		return err
	}

	if currency == "USD" {
		ubNew.BalanceType = "FEE"
		ubNew.Balance = -fee
		err = insertUserBalance(*ubNew)
		if err != nil {
			return err
		}
	}

	isNewCurrencyPayment := true
	paymentCycleDaysLeft := freq * pc.Seats
	dailyPaymentDaysLeft := freq * pc.Seats

	dailyPayments, err := findDailyPaymentByPaymentCycleId(u.PaymentCycleId)
	if err != nil {
		return err
	}
	// migrate remaining dailyPayments to new paymentCycle
	for _, dailyPayment := range dailyPayments {
		if dailyPayment.Currency == currency {
			isNewCurrencyPayment = false
			dailyPaymentDaysLeft += dailyPayment.DaysLeft
		}
		paymentCycleDaysLeft += dailyPayment.DaysLeft
		dailyPayment.PaymentCycleId = newPaymentCycleId
		dailyPayment.LastUpdate = time.Now()
		err = insertDailyPayment(dailyPayment)
	}

	balance, err := findSumUserBalanceByCurrency(u.Id, newPaymentCycleId, currency)
	if err != nil {
		return err
	}
	newDailyPaymentAmount := balance / int64(dailyPaymentDaysLeft)

	newDailyPayment := DailyPayment{newPaymentCycleId, currency, newDailyPaymentAmount, dailyPaymentDaysLeft, time.Now()}

	if isNewCurrencyPayment {
		err = insertDailyPayment(newDailyPayment)
		if err != nil {
			return err
		}
	} else {
		err = updateDailyPayment(newDailyPayment)
		if err != nil {
			return err
		}
	}

	err = updatePaymentCycleDaysLeft(newPaymentCycleId, int64(paymentCycleDaysLeft))
	if err != nil {
		return err
	}
	err = updatePaymentCycleId(u.Id, &newPaymentCycleId, nil)
	if err != nil {
		return err
	}

	go func(uid uuid.UUID) {
		err = sendToBrowser(uid, newPaymentCycleId)
		if err != nil {
			log.Printf("could not notify client %v", uid)
		}
	}(u.Id)
	return nil
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
