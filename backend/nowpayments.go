package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"log"
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
	OrderDescription string     `json:"order_description"`
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

type NowpaymentWebhookResponse struct {
	ActuallyPaid     float64    `json:"actually_paid"`
	InvoiceId        int64      `json:"invoice_id"`
	OrderDescription string     `json:"order_description"`
	OrderId          *uuid.UUID `json:"order_id"`
	OutcomeAmount    float64    `json:"outcome_amount"`
	OutcomeCurrency  string     `json:"outcome_currency"`
	PayAddress       string     `json:"pay_address"`
	PayAmount        float64    `json:"pay_amount"`
	PayCurrency      string     `json:"pay_currency"`
	PaymentId        int64      `json:"payment_id"`
	PaymentStatus    string     `json:"payment_status"`
	PriceAmount      float64    `json:"price_amount"`
	PriceCurrency    string     `json:"price_currency"`
	PurchaseId       string     `json:"purchase_id"`
}

func nowpaymentsPayment(w http.ResponseWriter, r *http.Request, user *User) {
	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		log.Printf("Cannot get payment informations for now payments: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)

	paymentCycleId, err := insertNewPaymentCycle(user.Id, seats, freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	currentUSDBalance, err := currentUSDBalance(user.PaymentCycleId)
	if err != nil {
		log.Printf("Cannot get current USD balance %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	price := (plan.Price * float64(seats)) - float64(currentUSDBalance)

	payCurrency := data["currency"]
	paymentResponse, err := createNowPayment(price, payCurrency, paymentCycleId, &user.Id, freq)

	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create payment: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(paymentResponse)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func createNowPayment(price float64, payCurrency string, paymentCycleId *uuid.UUID, uid *uuid.UUID, freq int64) (*PaymentResponse, error) {
	paymentUrl := opts.NowpaymentsApiUrl + "/payment"
	apiToken := opts.NowpaymentsToken
	if apiToken == "" {
		return nil, fmt.Errorf("now Paymenst API token is empty")
	}

	pr := PaymentRequest{
		PriceAmount:      price,
		PriceCurrency:    "USD",
		PayCurrency:      payCurrency,
		IpnCallbackUrl:   opts.NowpaymentsIpnCallbackUrl,
		OrderId:          paymentCycleId,
		OrderDescription: uid.String() + "#" + strconv.Itoa(int(freq)),
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
	log.Printf("aoeu: %v", string(paymentData))
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

func nowpaymentsWebhook(w http.ResponseWriter, r *http.Request) {
	nowpaymentsSignature := r.Header.Get("x-nowpayments-sig")
	var data NowpaymentWebhookResponse
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not parse webhook data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Could not convert to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isWebhookVerified, err := verifyNowpaymentsWebhook(jsonData, nowpaymentsSignature)
	if debug {
		isWebhookVerified = true
	}

	data.PayCurrency = strings.ToUpper(data.PayCurrency)
	data.PriceCurrency = strings.ToUpper(data.PriceCurrency)
	data.OutcomeCurrency = strings.ToUpper(data.OutcomeCurrency)
	data.PaymentStatus = strings.ToUpper(data.PaymentStatus)

	if err != nil {
		log.Printf("Could not verify Webhook data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isWebhookVerified {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	index := strings.Index(data.OrderDescription, "#")

	userId, err := uuid.Parse(data.OrderDescription[:index])
	if err != nil {
		log.Printf("Could not find uuid: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	freq, err := strconv.ParseInt(data.OrderDescription[index+1:], 10, 64)
	if err != nil {
		log.Printf("Could not parse freq: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := findUserById(userId)
	if err != nil {
		log.Printf("Could not find user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	amount, err := minCrypto(data.OutcomeCurrency, data.OutcomeAmount)
	if err != nil {
		log.Printf("Could not find currency: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch data.PaymentStatus {
	case "FINISHED":
		if err != nil {
			log.Printf("Could not process nowpayment success: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = paymentSuccess(user, *data.OrderId, amount, data.PayCurrency, freq, big.NewInt(0))
		if err != nil {
			log.Printf("Could not process nowpayment success: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["lang"] = "en"

		defaultMessage := "Your payment was successful, you can start supporting your favorit repositories!"
		e := prepareEmail(email, other,
			"template-subject-payment-success_", "Payment successful",
			"template-plain-payment-success_", defaultMessage,
			"template-html-payment-success_", other["lang"])

		s := *data.OrderId
		go func() {
			insertEmailSent(user.Id, "payment-success-"+s.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "EXPIRED":
		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/user/payments"
		other["lang"] = "en"

		defaultMessage := "Your payment expired. To start a new payment go to: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-payment-expired_", "Payment expired",
			"template-plain-payment-expired_", defaultMessage,
			"template-html-payment-expired_", other["lang"])

		s := *data.OrderId
		go func() {
			insertEmailSent(user.Id, "payment-expired-"+s.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "PARTIALLY_PAID":
		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = data.PayAddress
		other["lang"] = "en"

		defaultMessage := "Your payment is partially paid. Please transfer the missing amount over the following address: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-payment-expired_", "Partially paid",
			"template-plain-partially-paid_", defaultMessage,
			"template-html-partially-paid_", other["lang"])

		s := *data.OrderId
		go func() {
			insertEmailSent(user.Id, "partially-paid-"+s.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "FAILED":
		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["lang"] = "en"

		defaultMessage := "Your payment has failed. Please contact support@flatfeestack.io"
		e := prepareEmail(email, other,
			"template-subject-payment-failed_", "Payment failed",
			"template-plain-partially-failed_", defaultMessage,
			"template-html-partially-failed_", other["lang"])

		s := *data.OrderId
		go func() {
			insertEmailSent(user.Id, "payment-failed-"+s.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "REFUNDED":
		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["lang"] = "en"

		defaultMessage := "You got your money refunded."
		e := prepareEmail(email, other,
			"template-subject-payment-refunded_", "Payment refunded",
			"template-plain-partially-refunded_", defaultMessage,
			"template-html-partially-refunded_", other["lang"])

		s := *data.OrderId
		go func() {
			insertEmailSent(user.Id, "payment-refunded-"+s.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	default:
		log.Printf("Unhandled event type: %s\n", data.PaymentStatus)
		w.WriteHeader(http.StatusOK)
	}
}

func minCrypto(currency string, balance float64) (*big.Int, error) {
	i, err := getFactor(currency)
	if err != nil {
		return nil, err
	}
	f := new(big.Float).SetInt(i)
	amount := new(big.Int)
	amountF := new(big.Float).Mul(big.NewFloat(balance), f)
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

	freq, err := strconv.ParseInt(f, 10, 64)
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot convert number freq: %v", freq))
	}

	var plan *Plan
	for _, v := range plans {
		if v.Freq == freq {
			plan = &v
			break
		}
	}
	if plan == nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("No matching plan found: %v, available: %v", freq, plans))
	}

	return freq, seats, plan, nil
}

func verifyNowpaymentsWebhook(data []byte, nowpaymentsSignature string) (bool, error) {
	key := opts.NowpaymentsIpnKey
	mac := hmac.New(sha512.New, []byte(key))

	_, err := io.WriteString(mac, string(data))
	if err != nil {
		return false, err
	}
	expectedMAC := mac.Sum(nil)

	return hex.EncodeToString(expectedMAC) == nowpaymentsSignature, nil
}
