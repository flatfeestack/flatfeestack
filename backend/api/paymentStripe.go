package api

import (
	clnt "backend/clients"
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/setupintent"
	"github.com/stripe/stripe-go/v74/webhook"
	"io"
	"math/big"
	"net/http"
	"strconv"
)

type ClientSecretBody struct {
	ClientSecret string `json:"clientSecret"`
}

var (
	stripeAPISecretKey     string
	stripeWebhookSecretKey string
)

func InitStripe(stripeAPISecretKey0 string, stripeWebhookSecretKey0 string) {
	stripeAPISecretKey = stripeAPISecretKey0
	stripeWebhookSecretKey = stripeWebhookSecretKey0
}

func SetupStripe(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	// https://stripe.com/docs/payments/save-and-reuse
	//create a user at stripe if the user does not exist yet
	if user.StripeId == nil || stripeAPISecretKey != "" {
		stripe.Key = stripeAPISecretKey
		params := &stripe.CustomerParams{}
		c, err := customer.New(params)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
			return
		}
		user.StripeId = &c.ID
		err = db.UpdateStripe(user)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
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
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func StripePaymentInitial(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	if user.PaymentMethod == nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "No payment method defined for user: %v", user.Id)
		return
	}

	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		log.Printf("Cannot get payment informations for stripe payments: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cents := utils.UsdBaseToCent(plan.PriceBase) * seats

	params := &stripe.PaymentIntentParams{
		Amount:           stripe.Int64(int64(cents)),
		Currency:         stripe.String(string(stripe.CurrencyUSD)),
		Customer:         user.StripeId,
		PaymentMethod:    user.PaymentMethod,
		SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
	}

	e := uuid.New()
	payInEvent := db.PayInEvent{
		Id:         uuid.New(),
		ExternalId: e,
		UserId:     user.Id,
		Balance:    big.NewInt(plan.PriceBase * seats),
		Currency:   "USD",
		Status:     db.PayInRequest,
		Seats:      seats,
		Freq:       freq,
		CreatedAt:  utils.TimeNow(),
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["externalId"] = e.String()
	params.Params.Metadata["fee"] = strconv.FormatInt(plan.FeePrm, 10)
	params.Params.Metadata["freq"] = strconv.FormatInt(freq, 10)
	params.Params.Metadata["seats"] = strconv.FormatInt(seats, 10)

	intent, err := paymentintent.New(params)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	db.UpdateClientSecret(user.Id, intent.ClientSecret)
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func StripePaymentRecurring(user db.UserDetail) error {

	_, seats, freq, _, err := db.FindLatestDailyPayment(user.Id, "USD")
	if err != nil {
		return err
	}

	plan := findPlan(freq)
	cents := utils.UsdBaseToCent(plan.PriceBase) * seats

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(cents),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Customer:      user.StripeId,
		PaymentMethod: user.PaymentMethod,
		ClientSecret:  user.StripeClientSecret,
	}

	e := uuid.New()
	payInEvent := db.PayInEvent{
		Id:         uuid.New(),
		ExternalId: e,
		UserId:     user.Id,
		Balance:    big.NewInt(plan.PriceBase * int64(user.Seats)),
		Currency:   "USD",
		Status:     db.PayInRequest,
		Seats:      seats,
		Freq:       freq,
		CreatedAt:  utils.TimeNow(),
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		return err
	}

	params.Params.Metadata = map[string]string{}
	params.Params.Metadata["userId"] = user.Id.String()
	params.Params.Metadata["externalId"] = e.String()
	params.Params.Metadata["fee"] = strconv.FormatInt(plan.FeePrm, 10)
	params.Params.Metadata["freq"] = strconv.FormatInt(plan.Freq, 10)
	params.Params.Metadata["seats"] = strconv.Itoa(user.Seats)

	_, err = paymentintent.New(params)
	if err != nil {
		return err
	}

	return nil
}

func StripeWebhook(w http.ResponseWriter, req *http.Request) {
	// https://stripe.com/docs/testing#cards
	// regular card for testing: 4242 4242 4242 4242
	// 3d secure with auth required: 4000 0027 6000 3184
	// trigger 3d secure: 4000 0000 0000 3063
	// failed: 4000 0000 0000 0341 (checked)
	// insufficient funds: 4000 0000 0000 9995 (checked)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), stripeWebhookSecretKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		log.Printf("Error evaluating signed webhook request: %v", err)
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		externalId, feePrm, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			log.Printf("payin does not exist: %v, %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//Fee calculation
		fee := new(big.Int).Mul(payInEvent.Balance, big.NewInt(feePrm))
		fee = new(big.Int).Div(fee, big.NewInt(1000)) //we have promill
		fee = new(big.Int).Add(fee, big.NewInt(1))    //round up

		err = db.PaymentSuccess(externalId, fee)
		if err != nil {
			log.Printf("User sum balance cann run for %v: %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clnt.SendStripeSuccess(payInEvent.UserId, externalId)
	// ... handle other event types
	case "payment_intent.requires_action":
		//again
		externalId, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			log.Printf("payin does not exist: %v, %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInAction
		payInEvent.CreatedAt = utils.TimeNow()
		err = db.InsertPayInEvent(*payInEvent)
		if err != nil {
			log.Printf("insert payin does not exist: %v, %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clnt.SendStripeAction(payInEvent.UserId, externalId)
	//case "payment_intent.requires_action":
	//3d secure - this is handled by strip, we just get notified
	case "payment_intent.payment_failed":
		//requires_payment_method is when action is required but fails, dont report it twice
		if event.Data.Object["status"] == "requires_payment_method" {
			log.Infof("Payment failed due to requires_payment_method: %v\n", event.Data.Object)
			w.WriteHeader(http.StatusOK)
			return
		}
		externalId, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			log.Printf("Parer err from stripe: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			log.Printf("payin does not exist: %v, %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInMethod
		payInEvent.CreatedAt = utils.TimeNow()
		err = db.InsertPayInEvent(*payInEvent)
		if err != nil {
			log.Printf("insert payin does not exist: %v, %v\n", externalId, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clnt.SendStripeFailed(payInEvent.UserId, externalId)
	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
	}
}

func parseStripeData(data json.RawMessage) (uuid.UUID, int64, error) {
	var pi stripe.PaymentIntent
	err := json.Unmarshal(data, &pi)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing webhook JSON: %v\n", err)
	}
	uidRaw := pi.Metadata["userId"]
	uid, err := uuid.Parse(uidRaw)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing uid: %v, available %v\n", err, pi.Metadata)
	}
	externalIdRaw := pi.Metadata["externalId"]
	externalId, err := uuid.Parse(externalIdRaw)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing newPaymentCycleId: %v, available %v\n", err, pi.Metadata)
	}

	freq, err := strconv.ParseInt(pi.Metadata["freq"], 10, 64)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing freq: %v, available %v, %v\n", pi.Metadata["freq"], pi.Metadata, err)
	}

	seats, err := strconv.ParseInt(pi.Metadata["seats"], 10, 64)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing seats: %v, available %v, %v\n", pi.Metadata["seats"], pi.Metadata, err)
	}

	fee, err := strconv.ParseInt(pi.Metadata["fee"], 10, 64)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing fee: %v, available %v, %v\n", pi.Metadata["fee"], pi.Metadata, err)
	}

	payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("Error parsing seats: %v, available %v, %v\n", pi.Metadata["seats"], pi.Metadata, err)
	}

	if payInEvent.Seats != seats {
		return uuid.Nil, 0, fmt.Errorf("seats do not match: %v != %v", seats, payInEvent.Seats)
	}

	if payInEvent.Freq != freq {
		return uuid.Nil, 0, fmt.Errorf("freq do not match: %v != %v", freq, payInEvent.Freq)
	}

	if payInEvent.UserId != uid {
		return uuid.Nil, 0, fmt.Errorf("userId do not match: %v != %v", uid, payInEvent.UserId)
	}

	balance := big.NewInt(utils.UsdCentToBase(pi.Amount))
	if payInEvent.Balance.Cmp(balance) != 0 {
		return uuid.Nil, 0, fmt.Errorf("balance do not match: %v != %v", balance, payInEvent.Balance)
	}

	return externalId, fee, nil
}
