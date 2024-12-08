package api

import (
	"backend/internal/client"
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/setupintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

type ClientSecretBody struct {
	ClientSecret string `json:"clientSecret"`
}

type PaymentStripeHandler struct {
	e                      *client.EmailClient
	stripeAPISecretKey     string
	stripeWebhookSecretKey string
}

func NewPaymentHandler(e *client.EmailClient, stripeAPISecretKey0 string, stripeWebhookSecretKey string) *PaymentStripeHandler {
	return &PaymentStripeHandler{e, stripeAPISecretKey0, stripeWebhookSecretKey}
}

func (p *PaymentStripeHandler) SetupStripe(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	// https://stripe.com/docs/payments/save-and-reuse
	//create a user at stripe if the user does not exist yet
	if user.StripeId == nil || p.stripeAPISecretKey != "" {
		stripe.Key = p.stripeAPISecretKey
		params := &stripe.CustomerParams{}
		c, err := customer.New(params)
		if err != nil {
			slog.Error("Error while creating new customer",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		user.StripeId = &c.ID
		err = db.UpdateStripe(user)
		if err != nil {
			slog.Error("Error while updating stripe",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
		slog.Error("Error while creating setup intent",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func StripePaymentInitial(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	if user.PaymentMethod == nil {
		slog.Error("No payment method defined for user",
			slog.String("user", user.Id.String()))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong, no payment method set. Please try again.")
		return
	}

	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		slog.Error("Cannot get payment informations for stripe payments",
			slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cents := util.UsdBaseToCent(plan.PriceBase) * seats

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
		CreatedAt:  util.TimeNow(),
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		slog.Error("Cannot insert payment",
			slog.String("userId", user.Id.String()),
			slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//params.Params.Metadata = map[string]string{}
	//params.Params.Metadata["userId"] = user.Id.String()
	//params.Params.Metadata["externalId"] = e.String()
	//params.Params.Metadata["fee"] = strconv.FormatInt(plan.FeePrm, 10)
	//params.Params.Metadata["freq"] = strconv.FormatInt(freq, 10)
	//params.Params.Metadata["seats"] = strconv.FormatInt(seats, 10)
	params.Metadata = map[string]string{
		"userId":     user.Id.String(),
		"externalId": e.String(),
		"fee":        strconv.FormatInt(plan.FeePrm, 10),
		"freq":       strconv.FormatInt(freq, 10),
		"seats":      strconv.FormatInt(seats, 10),
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		slog.Error("Error while creating new Payment intent",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	cs := ClientSecretBody{ClientSecret: intent.ClientSecret}
	db.UpdateClientSecret(user.Id, intent.ClientSecret)
	err = json.NewEncoder(w).Encode(cs)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func StripePaymentRecurring(user db.UserDetail) error {

	_, seats, freq, _, err := db.FindLatestDailyPayment(user.Id, "USD")
	if err != nil {
		return err
	}

	plan := findPlan(freq)
	cents := util.UsdBaseToCent(plan.PriceBase) * seats

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
		CreatedAt:  util.TimeNow(),
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

func (p *PaymentStripeHandler) StripeWebhook(w http.ResponseWriter, req *http.Request) {
	// https://stripe.com/docs/testing#cards
	// regular card for testing: 4242 4242 4242 4242
	// 3d secure with auth required: 4000 0027 6000 3184
	// trigger 3d secure: 4000 0000 0000 3063
	// failed: 4000 0000 0000 0341 (checked)
	// insufficient funds: 4000 0000 0000 9995 (checked)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("Error reading request body",
			slog.Any("error", err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), p.stripeWebhookSecretKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		slog.Error("Error evaluating signed webhook request",
			slog.Any("error", err))
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		externalId, feePrm, err := parseStripeData(event.Data.Raw)
		if err != nil {
			slog.Error("Parser err from stripe",
				slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			slog.Error("Payin does not exist",
				slog.String("externalId", externalId.String()), slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//Fee calculation
		fee := new(big.Int).Mul(payInEvent.Balance, big.NewInt(feePrm))
		fee = new(big.Int).Div(fee, big.NewInt(1000)) //we have promill
		fee = new(big.Int).Add(fee, big.NewInt(1))    //round up

		err = db.PaymentSuccess(externalId, fee)
		if err != nil {
			slog.Error("User sum balance cannot run",
				slog.String("externalId", externalId.String()), slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p.e.SendStripeSuccess(payInEvent.UserId, externalId)
	// ... handle other event types
	case "payment_intent.requires_action":
		//again
		externalId, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			slog.Error("Parser err from stripe",
				slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			slog.Error("payin does not exist",
				slog.String("externalId", externalId.String()), slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInAction
		payInEvent.CreatedAt = util.TimeNow()
		err = db.InsertPayInEvent(*payInEvent)
		if err != nil {
			slog.Error("Insert payin does not exist",
				slog.String("externalId", externalId.String()), slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p.e.SendStripeAction(payInEvent.UserId, externalId)
	//case "payment_intent.requires_action":
	//3d secure - this is handled by strip, we just get notified
	case "payment_intent.payment_failed":
		//requires_payment_method is when action is required but fails, dont report it twice
		if event.Data.Object["status"] == "requires_payment_method" {
			slog.Info("Payment failed due to requires_payment_method",
				slog.Any("data", event.Data.Object))
			w.WriteHeader(http.StatusOK)
			return
		}
		externalId, _, err := parseStripeData(event.Data.Raw)
		if err != nil {
			slog.Error("Parser err from stripe",
				slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)
		if err != nil {
			slog.Error("Payin does not exist",
				slog.String("externalId", externalId.String()),
				slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInMethod
		payInEvent.CreatedAt = util.TimeNow()
		err = db.InsertPayInEvent(*payInEvent)
		if err != nil {
			slog.Error("Insert payin does not exist",
				slog.String("externalId", externalId.String()),
				slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p.e.SendStripeFailed(payInEvent.UserId, externalId)
	default:
		slog.Error("Unhandled event type",
			slog.String("type", string(event.Type)))
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

	balance := big.NewInt(util.UsdCentToBase(pi.Amount))
	if payInEvent.Balance.Cmp(balance) != 0 {
		return uuid.Nil, 0, fmt.Errorf("balance do not match: %v != %v", balance, payInEvent.Balance)
	}

	return externalId, fee, nil
}
