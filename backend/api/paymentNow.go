package api

import (
	clnt "backend/clients"
	db "backend/db"
	"backend/utils"
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"math/big"
	"net/http"
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

func InitNow(nowpaymentsApiUrl0 string, nowpaymentsToken0 string, nowpaymentsIpnCallbackUrl0 string, nowpaymentsIpnKey0 string) {
	nowpaymentsApiUrl = nowpaymentsApiUrl0
	nowpaymentsToken = nowpaymentsToken0
	nowpaymentsIpnCallbackUrl = nowpaymentsIpnCallbackUrl0
	nowpaymentsIpnKey = nowpaymentsIpnKey0
}

func NowPayment(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		log.Errorf("Cannot get payment information for now payments: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with retrieving the payment information. Please try again.")
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
		CreatedAt:  utils.TimeNow(),
	}

	err = db.InsertPayInEvent(payInEvent)
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	price := plan.Price * float64(seats)

	payCurrency := data["currency"]
	paymentResponse, err := createNowPayment(price, payCurrency, &e, &user.Id)

	if err != nil {
		log.Errorf("could not create payment: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with creating the payment. Please try again.")
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
		log.Errorf("Could not encode json: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
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
			log.Printf("Body ReadCloser: %v", err)
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

func NowWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Could not read body: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	nowSignature := r.Header.Get("x-nowpayments-sig")
	err = verifyNowWebhook(body, nowSignature)
	if err != nil {
		log.Errorf("Wrong signature: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	var data clnt.WebhookResponse
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Errorf("Could not parse webhook data: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	if data.OrderId == nil {
		log.Errorf("No orderId set for WebhookResponse")
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	externalId := *data.OrderId
	payInEvent, err := db.FindPayInExternal(externalId, db.PayInRequest)

	if err != nil {
		log.Errorf("Error while finding pay in external: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	switch data.PaymentStatus {
	case "finished":
		err = db.PaymentSuccess(externalId, big.NewInt(0))
		if err != nil {
			log.Errorf("Could not process now payment success: %v", err)
			utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
			return
		}
		clnt.SendPaymentNowFinished(payInEvent.UserId, data)
	case "partially_paid":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInPartially
		payInEvent.CreatedAt = utils.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		clnt.SendPaymentNowPartially(payInEvent.UserId, data)
	case "expired":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInExpired
		payInEvent.CreatedAt = utils.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		clnt.SendPaymentNowRefunded(payInEvent.UserId, "expired", externalId)
	case "failed":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInFailed
		payInEvent.CreatedAt = utils.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		clnt.SendPaymentNowRefunded(payInEvent.UserId, "failed", externalId)
	case "refunded":
		payInEvent.Id = uuid.New()
		payInEvent.Status = db.PayInRefunded
		payInEvent.CreatedAt = utils.TimeNow()
		db.InsertPayInEvent(*payInEvent)
		clnt.SendPaymentNowRefunded(payInEvent.UserId, "refunded", externalId)
	default:
		log.Printf("Unhandled event type: %s\n", data.PaymentStatus)
		w.WriteHeader(http.StatusOK)
	}
}

func minCrypto(currency string, balance float64) (*big.Int, error) {
	i, err := utils.GetFactor(currency)
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
	p := mux.Vars(r)
	f := p["freq"]
	s := p["seats"]
	seats, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot convert number seats: %v", +seats))
	}
	if seats > 1_000_000 {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot have more than 1m seats: %v", +seats))
	}

	freq, err := strconv.ParseInt(f, 10, 64)
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
