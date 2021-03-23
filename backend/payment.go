package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/setupintent"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(seats * freq * 100)),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		Confirm:       stripe.Bool(false),
		OffSession:    stripe.Bool(false),
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["seats"] = s
	params.Params.Metadata["freq"] = strconv.Itoa(freq)
	params.Params.Metadata["uid"] = user.Id.String()

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

func stripePaymentRecurring(user *User, noConfirm bool) (*ClientSecretBody, error) {
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
}

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
		uidRaw := pi.Metadata["uid"]
		if uidRaw == "" {
			log.Printf("Error parsing webhook, no uid: %v\n", uidRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		uid, err := uuid.Parse(uidRaw)
		if err != nil {
			log.Printf("Error parsing uid: %v\n", uidRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		seatRaw := pi.Metadata["seats"]
		if seatRaw == "" {
			log.Printf("Error parsing webhook, no uid: %v\n", uidRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		seats, err := strconv.Atoi(seatRaw)
		if err != nil {
			log.Printf("Error parsing seats: %v\n", uidRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		freqRaw := pi.Metadata["freq"]
		if freqRaw == "" {
			log.Printf("Error parsing freq, no uid: %v\n", freqRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		freq, err := strconv.Atoi(freqRaw)
		if err != nil {
			log.Printf("Error parsing freq: %v\n", freqRaw)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		u, err := findUserById(uid)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		u.Freq = freq
		u.Seats = seats
		b := *u.Balance
		amount := b + (pi.Amount * 10000) //pi.Amount is in cent, we need in microUSD

		oldSum, err := findSumUserBalance(u.PaymentCycleId, u.Id)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newPaymentCycleId := uuid.New()
		ubNew := UserBalance{
			paymentCycleId: newPaymentCycleId,
			userId:         u.Id,
			balance:        oldSum,
			day:            timeNow(),
			balanceType:    "REM",
			createdAt:      timeNow(),
		}
		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ubNew.balanceType = "PAY"
		ubNew.balance = amount
		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		u.PaymentCycleId = newPaymentCycleId
		err = updateUser(u)
		if err != nil {
			log.Printf("User does not exist: %v\n", uid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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
