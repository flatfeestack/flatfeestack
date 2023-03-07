package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/setupintent"
	"github.com/stripe/stripe-go/v74/webhook"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
)

type ClientSecretBody struct {
	ClientSecret string `json:"clientSecret"`
}

// https://stripe.com/docs/payments/save-and-reuse
func setupStripe(w http.ResponseWriter, r *http.Request, user *User) {
	//create a user at stripe if the user does not exist yet
	if user.StripeId == nil || opts.StripeAPISecretKey != "" {
		stripe.Key = opts.StripeAPISecretKey
		params := &stripe.CustomerParams{}
		c, err := customer.New(params)
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
			return
		}
		user.StripeId = &c.ID
		err = updateUser(user)
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
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
		writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func stripePaymentInitial(w http.ResponseWriter, r *http.Request, user *User) {
	if user.PaymentMethod == nil {
		writeErrorf(w, http.StatusInternalServerError, "No payment method defined for user: %v", user.Id)
		return
	}

	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		log.Printf("Cannot get payment informations for stripe payments: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	currentUSDBalance, err := currentUSDBalance(user.PaymentCycleInId)
	if err != nil {
		log.Printf("Cannot get current USD balance %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cents := ((plan.Price * float64(seats)) - float64(currentUSDBalance)) * 100

	params := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(int64(cents)),
		Currency:         stripe.String(string(stripe.CurrencyUSD)),
		Customer:         user.StripeId,
		PaymentMethod:    user.PaymentMethod,
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	paymentCycleId, err := insertNewPaymentCycleIn(seats, freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["paymentCycleId"] = paymentCycleId.String()
	params.Params.Metadata["fee"] = strconv.FormatInt(plan.FeePrm, 10)
	params.Params.Metadata["freq"] = strconv.FormatInt(freq, 10)
	params.Params.Metadata["seats"] = strconv.FormatInt(seats, 10)

	intent, err := paymentintent.New(params)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

// https://stripe.com/docs/testing#cards
// regular card for testing: 4242 4242 4242 4242
// 3d secure with auth required: 4000 0027 6000 3184
// trigger 3d secure: 4000 0000 0000 3063
// failed: 4000 0000 0000 0341 (checked)
// insufficient funds: 4000 0000 0000 9995 (checked)
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
		uid, newPaymentCycleInId, amount, feePrm, seat, freq, err := parseStripeData(event.Data.Raw)
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

		err = paymentSuccess(u.Id, u.PaymentCycleInId, newPaymentCycleInId, big.NewInt(amount*10000), "USD", seat, freq, big.NewInt(fee*10000))
		if err != nil {
			log.Printf("User sum balance cann run for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = sendToBrowser(u.Id, newPaymentCycleInId)
		if err != nil {
			log.Debugf("browser offline, best effort, we write a email to %s anyway", err)
		}

		sendStripeSuccess(*u, newPaymentCycleInId)
	// ... handle other event types
	case "payment_intent.requires_action":
		//again
		uid, newPaymentCycleInId, _, _, _, _, err := parseStripeData(event.Data.Raw)
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

		ub, err := findUserBalancesAndType(newPaymentCycleInId, "REQACT", "USD")
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
			PaymentCycleInId: newPaymentCycleInId,
			UserId:           uid,
			Balance:          big.NewInt(0),
			BalanceType:      "REQACT",
			Currency:         "USD",
			CreatedAt:        timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = sendToBrowser(uid, newPaymentCycleInId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)
		}

		sendStripeAction(*u, newPaymentCycleInId)
	//case "payment_intent.requires_action":
	//3d secure - this is handled by strip, we just get notified
	case "payment_intent.payment_failed":
		//requires_payment_method is when action is required but fails, dont report it twice
		if event.Data.Object["status"] == "requires_payment_method" {
			log.Infof("Payment failed due to requires_payment_method: %v\n", event.Data.Object)
			w.WriteHeader(http.StatusOK)
			return
		}
		uid, newPaymentCycleInId, _, _, _, _, err := parseStripeData(event.Data.Raw)
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

		ub, err := findUserBalancesAndType(newPaymentCycleInId, "FAILED", "USD")
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
			PaymentCycleInId: newPaymentCycleInId,
			UserId:           uid,
			Balance:          big.NewInt(0),
			BalanceType:      "FAILED",
			Currency:         "USD",
			CreatedAt:        timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", uid, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = sendToBrowser(uid, newPaymentCycleInId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)
		}

		sendStripeFailed(*u, newPaymentCycleInId)
	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
	}
}

func parseStripeData(data json.RawMessage) (uuid.UUID, *uuid.UUID, int64, int64, int64, int64, error) {
	var pi stripe.PaymentIntent
	err := json.Unmarshal(data, &pi)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing webhook JSON: %v\n", err)
	}
	uidRaw := pi.Metadata["userId"]
	uid, err := uuid.Parse(uidRaw)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing uid: %v, available %v\n", err, pi.Metadata)
	}
	newPaymentCycleIdRaw := pi.Metadata["paymentCycleId"]
	newPaymentCycleId, err := uuid.Parse(newPaymentCycleIdRaw)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing newPaymentCycleId: %v, available %v\n", err, pi.Metadata)
	}
	feePrm, err := strconv.ParseInt(pi.Metadata["fee"], 10, 64)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing fee: %v, available %v\n", pi.Metadata["fee"], pi.Metadata)
	}

	freq, err := strconv.ParseInt(pi.Metadata["freq"], 10, 64)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing freq: %v, available %v, %v\n", pi.Metadata["freq"], pi.Metadata, err)
	}

	seats, err := strconv.ParseInt(pi.Metadata["seats"], 10, 64)
	if err != nil {
		return uuid.Nil, nil, 0, 0, 0, 0, fmt.Errorf("Error parsing seats: %v, available %v, %v\n", pi.Metadata["seats"], pi.Metadata, err)
	}
	return uid, &newPaymentCycleId, pi.Amount, feePrm, seats, freq, nil
}

func stripePaymentRecurring(user User) (*ClientSecretBody, error) {
	pc, err := findPaymentCycle(user.PaymentCycleInId)
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

	currentUSDBalance, err := currentUSDBalance(user.PaymentCycleInId)
	if err != nil {
		return nil, err
	}
	cents := ((plan.Price * float64(pc.Seats)) - float64(currentUSDBalance)) * 100

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(cents)),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		Confirm:       stripe.Bool(true),
		OffSession:    stripe.Bool(true),
	}

	paymentCycleId, err := insertNewPaymentCycleIn(pc.Seats, pc.Freq, timeNow())
	if err != nil {
		return nil, err
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["paymentCycleId"] = paymentCycleId.String()
	params.Params.Metadata["fee"] = strconv.FormatInt(plan.FeePrm, 10)
	params.Params.Metadata["freq"] = strconv.FormatInt(plan.Freq, 10)
	params.Params.Metadata["seats"] = strconv.FormatInt(pc.Seats, 10)

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	return &cs, nil
}
