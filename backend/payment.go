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
	"time"
)

type ClientSecretBody struct {
	ClientSecret string `json:"client_secret"`
}

//https://stripe.com/docs/payments/save-and-reuse
func setupStripe(w http.ResponseWriter, r *http.Request, user *User) {
	//create a user at stripe if the user does not exist yet
	if user.StripeId == nil || opts.StripeSecret != "" {
		stripe.Key = opts.StripeSecret
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
	p := mux.Vars(r)
	f := p["freq"]
	s := p["seats"]
	seats, err := strconv.Atoi(s)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Cannot convert number seats: %v", seats)
		return
	}

	freq := 0
	if f == "quarterly" {
		freq = 30 //3 month
	} else if f == "yearly" {
		freq = 120 //1 year
	} else {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", f)
		return
	}

	paymentCycleId, err := insertNewPaymentCycle(user.Id, freq, seats, freq, timeNow())
	if err != nil {
		log.Printf("User does not exist: %v\n", user.Id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(seats * freq * 100)),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		Confirm:       stripe.Bool(false),
		OffSession:    stripe.Bool(false),
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

/*func stripePaymentRecurring(user *User, noConfirm bool) (*ClientSecretBody, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(user.Seats * user.Freq * 100)),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		Confirm:       stripe.Bool(noConfirm),
		OffSession:    stripe.Bool(noConfirm),
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	return &cs, nil
}*/

func stripeWebhook(w http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), "whsec_9HJx5EoyhE1K3UFBnTxpOSr0lscZMHJL")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		log.Printf("Error evaluating signed webhook request: %v", err)
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		uidRaw := pi.Metadata["userId"]
		uid, err := uuid.Parse(uidRaw)
		if err != nil {
			log.Printf("Error parsing uid: %v\n", uidRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newPaymentCycleIdRaw := pi.Metadata["paymentCycleId"]
		newPaymentCycleId, err := uuid.Parse(newPaymentCycleIdRaw)
		if err != nil {
			log.Printf("Error parsing uid: %v\n", newPaymentCycleIdRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		u, err := findUserById(uid)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		amount := pi.Amount * 10000 //pi.Amount is in cent, we need in microUSD

		oldSum := int64(0)
		//uuid.Nil does not work: https://github.com/google/uuid/issues/45
		if !isUUIDZero(u.PaymentCycleId) {
			oldSum, err = findSumUserBalance(u.PaymentCycleId, u.Id)
			if err != nil {
				log.Printf("User does not exist: %v\n", uid)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ubNew := UserBalance{
			PaymentCycleId: newPaymentCycleId,
			UserId:         u.Id,
			Balance:        oldSum,
			Day:            timeNow(),
			BalanceType:    "REM",
			CreatedAt:      timeNow(),
		}

		if oldSum > 0 {
			err = insertUserBalance(ubNew)
			if err != nil {
				log.Printf("User does not exist: %v\n", uid)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		ubNew.BalanceType = "PAY"
		ubNew.Balance = amount
		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = updatePaymentCycleId(uid, &newPaymentCycleId)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sendToBrowser(uid)

	// ... handle other event types
	case "payment_intent.authentication_required":
		//TODO: send email

	case "payment_intent.insufficient_funds":
		//TODO: send email

	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[uuid.UUID]*websocket.Conn)
var lock = sync.Mutex{}

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
		lock.Lock()
		defer lock.Unlock()
		delete(clients, user.Id)
		return nil
	})
	sendToBrowser(user.Id)
}

func sendToBrowser(userId uuid.UUID) error {
	lock.Lock()
	defer lock.Unlock()

	conn := clients[userId]

	if conn == nil {
		return fmt.Errorf("cannot get websockt for client %v", userId)
	}

	err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		delete(clients, userId)
		return err
	}

	userBalances, err := findUserBalances(userId)
	if err != nil {
		delete(clients, userId)
		return err
	}

	err = conn.WriteJSON(userBalances)
	if err != nil {
		delete(clients, userId)
		return err
	}

	return nil
}

func cancelSub(w http.ResponseWriter, r *http.Request, user *User) {
	err := updatePaymentCycleId(user.Id, nil)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not cancel subscription: %v", err)
		return
	}
}
