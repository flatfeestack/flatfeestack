package api

import (
	"backend/internal/client"
	"backend/internal/db"
	"backend/pkg/util"
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PaymentRequest struct {
	PriceAmount      float64    `json:"price_amount"`
	PriceCurrency    string     `json:"price_currency"`
	PayCurrency      string     `json:"pay_currency"`
	IpnCallbackUrl   string     `json:"ipn_callback_url"`
	OrderId          *uuid.UUID `json:"order_id"`
	OrderDescription *uuid.UUID `json:"order_description"`
}

type PaymentResponse struct {
	PaymentId        string  `json:"payment_id"`
	PaymentStatus    string  `json:"payment_status"`
	PayAddress       string  `json:"pay_address"`
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayAmount        float64 `json:"pay_amount"`
	PayCurrency      string  `json:"pay_currency"`
	OrderId          string  `json:"order_id"`
	OrderDescription string  `json:"order_description"`
	IpnCallbackUrl   string  `json:"ipn_callback_url"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	PurchaseId       string  `json:"purchase_id"`
}

type PaymentResponse2 struct {
	PayAddress  string   `json:"payAddress"`
	PayAmount   *big.Int `json:"payAmount"`
	PayCurrency string   `json:"payCurrency"`
}

var (
	nowpaymentsApiUrl         string
	nowpaymentsToken          string
	nowpaymentsIpnCallbackUrl string
	nowpaymentsIpnKey         string
)

type PaymentNowHandler struct {
	e                         *client.EmailClient
	nowpaymentsApiUrl         string
	nowpaymentsToken          string
	nowpaymentsIpnCallbackUrl string
	nowpaymentsIpnKey         string
}

func NewPaymentNowHandler(e *client.EmailClient, nowpaymentsApiUrl0 string, nowpaymentsToken0 string, nowpaymentsIpnCallbackUrl0 string, nowpaymentsIpnKey0 string) *PaymentNowHandler {
	return &PaymentNowHandler{
		e,
		nowpaymentsApiUrl0,
		nowpaymentsToken0,
		nowpaymentsIpnCallbackUrl0,
		nowpaymentsIpnKey0}
}

func NowPayment(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		slog.Error("Cannot get payment information for now payments",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with retrieving the payment information. Please try again.")
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)

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

	price := plan.Price * float64(seats)

	payCurrency := data["currency"]
	paymentResponse, err := createNowPayment(price, payCurrency, &e, &user.Id)

	if err != nil {
		slog.Error("could not create payment",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with creating the payment. Please try again.")
		return
	}

	amount, err := minCrypto(paymentResponse.PayCurrency, paymentResponse.PayAmount)

	paymentResponse2 := PaymentResponse2{
		PayAddress:  paymentResponse.PayAddress,
		PayAmount:   amount,
		PayCurrency: paymentResponse.PayCurrency,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(paymentResponse2)
	if err != nil {
		slog.Error("Could not encode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}
}

func createNowPayment(price float64, payCurrency string, externId *uuid.UUID, uid *uuid.UUID) (*PaymentResponse, error) {
	paymentUrl := nowpaymentsApiUrl + "/payment"
	apiToken := nowpaymentsToken
	if apiToken == "" {
		return nil, fmt.Errorf("now Paymenst API token is empty")
	}

	pr := PaymentRequest{
		PriceAmount:      price,
		PriceCurrency:    "USD",
		PayCurrency:      payCurrency,
		IpnCallbackUrl:   nowpaymentsIpnCallbackUrl,
		OrderId:          externId,
		OrderDescription: uid,
	}

	paymentData, err := json.Marshal(pr)

	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", paymentUrl, strings.NewReader(string(paymentData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusCreated {
		b, err := io.ReadAll(response.Body)
		return nil, fmt.Errorf("bad request: %v (%v)", string(b), err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("Body ReadCloser",
				slog.Any("error", err))
		}
	}(response.Body)

	var data *PaymentResponse
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	data.PriceCurrency = strings.ToUpper(data.PriceCurrency)
	data.PayCurrency = strings.ToUpper(data.PayCurrency)

	return data, nil
}

func (p *PaymentNowHandler) NowWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Could not read body",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	nowSignature := r.Header.Get("x-nowpayments-sig")
	err = verifyNowWebhook(body, nowSignature)
	if err != nil {
		slog.Error("Wrong signature",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	var data client.WebhookResponse
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		slog.Error("Could not parse webhook data",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	if data.OrderId == nil {
		slog.Error("No orderId set for WebhookResponse")
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	externalId := *data.OrderId
	payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)

	if err != nil {
		slog.Error("Error while finding pay in external",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	switch data.PaymentStatus {
	case "finished":
		err = db.PaymentSuccess(externalId, big.NewInt(0))
		if err != nil {
			slog.Error("Could not process now payment success",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
			return
		}
		p.e.SendPaymentNowFinished(payInEvent.UserId, data)
	case "partially_paid":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInPartially
		payInEvent.CreatedAt = util.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		p.e.SendPaymentNowPartially(payInEvent.UserId, data)
	case "expired":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInExpired
		payInEvent.CreatedAt = util.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		p.e.SendPaymentNowRefunded(payInEvent.UserId, "expired", externalId)
	case "failed":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInFailed
		payInEvent.CreatedAt = util.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		p.e.SendPaymentNowRefunded(payInEvent.UserId, "failed", externalId)
	case "refunded":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInRefunded
		payInEvent.CreatedAt = util.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		p.e.SendPaymentNowRefunded(payInEvent.UserId, "refunded", externalId)
	default:
		slog.Error("Unhandled event type",
			slog.String("status", data.PaymentStatus))
		w.WriteHeader(http.StatusOK)
	}
}

func minCrypto(currency string, balance float64) (*big.Int, error) {
	i, err := util.GetFactor(currency)
	if err != nil {
		return nil, err
	}
	f := new(big.Float).SetInt(i)
	amountF := new(big.Float).Mul(big.NewFloat(balance), f)

	amount := new(big.Int)
	amountF.Int(amount)
	return amount, nil
}

// Helper
func paymentInformation(r *http.Request) (int64, int64, *Plan, error) {

	freqEsc := r.PathValue("freq")
	freqStr, err := url.QueryUnescape(freqEsc)

	if err != nil {
		slog.Error("Query unescape payment freq",
			slog.String("freq", freqEsc),
			slog.Any("error", err))
		return 0, 0, nil, err
	}

	seatsEsc := r.PathValue("seats")
	seatsStr, err := url.QueryUnescape(seatsEsc)

	if err != nil {
		slog.Error("Query unescape payment seats",
			slog.String("seats", seatsStr),
			slog.Any("error", err))
		return 0, 0, nil, err
	}

	seats, err := strconv.ParseInt(seatsStr, 10, 64)
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot convert number seats: %v", +seats))
	}
	if seats > 1_000_000 {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot have more than 1m seats: %v", +seats))
	}

	freq, err := strconv.ParseInt(freqStr, 10, 64)
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot convert number freq: %v", freq))
	}

	plan := findPlan(freq)
	if plan == nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("No matching plan found: %v, available: %v", freq, Plans))
	}

	return freq, seats, plan, nil
}

func findPlan(freq int64) *Plan {
	var plan *Plan
	for _, v := range Plans {
		if v.Freq == freq {
			plan = &v
			break
		}
	}
	return plan
}

func verifyNowWebhook(data []byte, sig string) error {
	key := nowpaymentsIpnKey
	mac := hmac.New(sha512.New, []byte(key))

	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	//marshal it again, it will be sorted
	//https://stackoverflow.com/questions/18668652/how-to-produce-json-with-sorted-keys-in-go
	data, err = json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = mac.Write(data)
	if err != nil {
		return err
	}

	expectedMAC := mac.Sum(nil)
	expected := hex.EncodeToString(expectedMAC)
	if expected != sig {
		return fmt.Errorf("signatures do not match calc(%s) != sent(%s)", expected, sig)
	}
	return nil
}
