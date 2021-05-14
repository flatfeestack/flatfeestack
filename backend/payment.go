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
	"net/http"
	"strconv"
	"sync"
)

type ClientSecretBody struct {
	ClientSecret string `json:"client_secret"`
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

	balance, err := findSumUserBalance(user.Id, user.PaymentCycleId)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not find user balance: %v", err)
		return
	}
	if balance >= mUSDPerDay {
		log.Printf("enough funds")
		return
	}

	for k, inviteEmail := range user.Claims.InviteEmails {
		sponsor, err := findUserByEmail(inviteEmail)
		if err != nil {
			log.Printf("findUserByEmail: %v", err)
			continue
		}

		//parent has enough funds go for it!
		sum, err := findSumUserBalance(sponsor.Id, sponsor.PaymentCycleId)
		if err != nil {
			log.Printf("findSumUserBalance: %v", err)
			continue
		}
		freq, err := strconv.Atoi(user.Claims.InviteMeta[k])
		if err != nil {
			log.Printf("findSumUserBalance: %v", err)
			continue
		}

		balance := int64(freq * mUSDPerDay)
		if sum < balance {
			log.Printf("parent has not enough funding")
			//TODO: notify parent
			continue
		}

		paymentCycleId, err := insertNewPaymentCycle(user.Id, freq, 1, freq, timeNow())
		if err != nil {
			log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ub := UserBalance{
			PaymentCycleId: sponsor.PaymentCycleId,
			UserId:         sponsor.Id,
			Balance:        -balance,
			BalanceType:    "SPONSOR",
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ub)
		if err != nil {
			log.Printf("transferBalance: %v", err)
			continue
		}

		ub.UserId = user.Id
		ub.Balance = balance
		ub.PaymentCycleId = *paymentCycleId
		err = insertUserBalance(ub)
		if err != nil {
			log.Printf("transferBalance: %v", err)
			continue
		}
		err = updatePaymentCycleId(user.Id, paymentCycleId, &sponsor.Id)
		if err != nil {
			log.Printf("transferBalance: %v", err)
			continue
		}

		go func() {
			err = sendToBrowser(user.Id, sponsor.PaymentCycleId)
			if err != nil {
				log.Printf("could not notify client %v", user.Id)
			}
		}()
	}
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
	match := false
	for _, v := range plans {
		if v.Freq == freq {
			match = true
			break
		}
	}
	if !match {
		writeErr(w, http.StatusInternalServerError, "Cannot convert number freq: %v", seats)
		return
	}

	cents := stripe.Int64(int64(seats * freq * mUSDPerDay / 10000))
	params := &stripe.PaymentIntentParams{
		Amount:           cents,
		Currency:         stripe.String(string(stripe.CurrencyUSD)),
		Customer:         user.StripeId,
		PaymentMethod:    user.PaymentMethod,
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	paymentCycleId, err := insertNewPaymentCycle(user.Id, freq, seats, freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["paymentCycleId"] = paymentCycleId.String()

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

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(pc.Seats * pc.Freq * 100)),
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

		uid, newPaymentCycleId, amount, err := parseStripeData(event.Data.Raw)
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

		oldSum, err := findSumUserBalance(uid, u.PaymentCycleId)
		if err != nil {
			log.Printf("User sum balance cann run for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ubNew := UserBalance{
			PaymentCycleId: u.PaymentCycleId,
			UserId:         u.Id,
			Balance:        -oldSum,
			BalanceType:    "CLOSE_CYCLE",
			CreatedAt:      timeNow(),
		}

		if oldSum > 0 {
			err = insertUserBalance(ubNew)
			if err != nil {
				log.Printf("Insert user balance1 for %v: %v\n", uid, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ubNew.Balance = oldSum
			ubNew.PaymentCycleId = newPaymentCycleId
			ubNew.BalanceType = "CARRY_OVER"
			err = insertUserBalance(ubNew)
			if err != nil {
				log.Printf("Insert user balance2 for %v: %v\n", uid, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ubNew.PaymentCycleId = newPaymentCycleId
		ubNew.BalanceType = "PAYMENT"
		ubNew.Balance = amount
		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance3 for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = updatePaymentCycleId(uid, &newPaymentCycleId, nil)
		if err != nil {
			log.Printf("Update payment cycle for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			err = sendToBrowser(uid, newPaymentCycleId)
			if err != nil {
				log.Printf("could not notify client %v", uid)
			}
		}()

	// ... handle other event types
	case "payment_intent.authentication_required":
	case "payment_intent.requires_payment_method":
		//again
		fmt.Printf("CASE-STRIP: %v", event)
		uid, newPaymentCycleId, _, err := parseStripeData(event.Data.Raw)
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

		ubNew := UserBalance{
			PaymentCycleId: u.PaymentCycleId,
			UserId:         uid,
			Balance:        0,
			BalanceType:    "AUTHREQ",
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := *u.Email
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
		uid, newPaymentCycleId, _, err := parseStripeData(event.Data.Raw)
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

		ubNew := UserBalance{
			PaymentCycleId: u.PaymentCycleId,
			UserId:         uid,
			Balance:        0,
			BalanceType:    "NOFUNDS",
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := *u.Email
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

func parseStripeData(data json.RawMessage) (uuid.UUID, uuid.UUID, int64, error) {
	var pi stripe.PaymentIntent
	err := json.Unmarshal(data, &pi)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, fmt.Errorf("Error parsing webhook JSON: %v\n", err)
	}
	uidRaw := pi.Metadata["userId"]
	uid, err := uuid.Parse(uidRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, fmt.Errorf("Error parsing uid: %v\n", err)
	}
	newPaymentCycleIdRaw := pi.Metadata["paymentCycleId"]
	newPaymentCycleId, err := uuid.Parse(newPaymentCycleIdRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, fmt.Errorf("Error parsing newPaymentCycleId: %v\n", err)
	}
	return uid, newPaymentCycleId, pi.Amount * 10000, nil
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
