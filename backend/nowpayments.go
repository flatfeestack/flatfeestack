package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
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

type InvoiceRequest struct {
	PriceAmount    float64 `json:"price_amount"`
	PriceCurrency  string  `json:"price_currency"`
	PayCurrency    string  `json:"pay_currency"`
	IpnCallbackUrl string  `json:"ipn_callback_url"`
}

type InvoiceResponse struct {
	Id               string `json:"id"`
	OrderId          string `json:"order_id"`
	OrderDescription string `json:"order_description"`
	PriceAmount      string `json:"price_amount"`
	PriceCurrency    string `json:"price_currency"`
	PayCurrency      string `json:"pay_currency"`
	IpnCallbackUrl   string `json:"ipn_callback_url"`
	InvoiceUrl       string `json:"invoice_url"`
	SuccessUrl       string `json:"success_url"`
	CancelUrl        string `json:"cancel_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type InvoiceDB struct {
	NowpaymentsInvoiceId int64
	PaymentCycleId       *uuid.UUID
	PaymentId            sql.NullInt64

	PriceAmount   int64  // price amount in microUSD
	PriceCurrency string // USD

	PayAmount   sql.NullInt64 // how much to pay in crypto (nanoCrypto)
	PayCurrency string        // in which currency

	ActuallyPaid    sql.NullInt64  // how much the user really paid (incl. fees)
	OutcomeAmount   sql.NullInt64  // how much we get in our wallet
	OutcomeCurrency sql.NullString // in which wallet the pay-in goes

	PaymentStatus string
	CreatedAt     string
	LastUpdate    string
}

type PaymentInformation struct {
	Freq  int
	Seats int
	Plan  *Plan
}

type NowpaymentWebhookResponse struct {
	ActuallyPaid     float64     `json:"actually_paid"`
	InvoiceId        int64       `json:"invoice_id"`
	OrderDescription interface{} `json:"order_description"`
	OrderId          interface{} `json:"order_id"`
	OutcomeAmount    float64     `json:"outcome_amount"`
	OutcomeCurrency  string      `json:"outcome_currency"`
	PayAddress       string      `json:"pay_address"`
	PayAmount        float64     `json:"pay_amount"`
	PayCurrency      string      `json:"pay_currency"`
	PaymentId        int64       `json:"payment_id"`
	PaymentStatus    string      `json:"payment_status"`
	PriceAmount      float64     `json:"price_amount"`
	PriceCurrency    string      `json:"price_currency"`
	PurchaseId       string      `json:"purchase_id"`
}

func nowpaymentPayment(w http.ResponseWriter, r *http.Request, user *User) {
	paymentInformation, err := getPaymentInformation(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)

	usdPrice := getPrice(paymentInformation)

	priceCurrency := "USD"
	payCurrency := data["currency"] //get from request

	paymentCycleId, err := insertNewPaymentCycle(user.Id, 0, paymentInformation.Seats, paymentInformation.Freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invoice := InvoiceRequest{float64(usdPrice), priceCurrency, payCurrency, ""}

	err = createNowpaymentsInvoice(invoice, paymentCycleId)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create invoice: %v", err)
		return
	}
}

func createNowpaymentsInvoice(invoice InvoiceRequest, paymentCycleId *uuid.UUID) error {
	invoiceUrl := "https://api.nowpayments.io/v1/invoice"
	apiToken := opts.NowpaymentsToken
	invoice.IpnCallbackUrl = "https://95a5-2a02-aa16-947f-b800-30e6-cfcb-a1fc-c204.ngrok.io/hooks/nowpayments"
	invoiceData, err := json.Marshal(invoice)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", invoiceUrl, strings.NewReader(string(invoiceData)))
	req.Header.Set("x-api-key", apiToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	var data InvoiceResponse
	json.NewDecoder(response.Body).Decode(&data)
	id, err := strconv.ParseInt(data.Id, 10, 64)
	priceAmount, err := strconv.ParseInt(data.PriceAmount, 10, 64)

	db := InvoiceDB{
		NowpaymentsInvoiceId: id,
		PaymentCycleId:       paymentCycleId,

		PriceAmount:   priceAmount * 1_000_000_000,
		PriceCurrency: data.PriceCurrency,

		PayCurrency:   data.PayCurrency,
		PaymentStatus: "CREATED",
		CreatedAt:     data.CreatedAt,
		LastUpdate:    data.CreatedAt,
	}

	_, err = insertNewInvoice(db)

	return err
}

func nowpaymentsWebhook(w http.ResponseWriter, r *http.Request) {
	nowpaymentsSignature := r.Header.Get("x-nowpayments-sig")
	var data NowpaymentWebhookResponse
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not parse webhook data: %v", err)
		return
	}

	jsonData, err := json.Marshal(data)

	fmt.Println(string(jsonData))

	isWebhookVerified := verifyNowpaymentsWebhook(jsonData, nowpaymentsSignature)
	if isWebhookVerified && !isWebhookVerified {
		return
	}

	// maybe simplify
	invoice, _ := getInvoice(data.InvoiceId)
	userId, _ := findUserIdByInvoice(invoice.NowpaymentsInvoiceId)
	user, _ := findUserById(userId)

	amount := int64(data.OutcomeAmount * 1_000_000_000) // nanoCrypto

	switch data.PaymentStatus {
	case "finished":
		err := nowpaymentSuccess(user, *invoice.PaymentCycleId, amount, data.PayCurrency)
		if err != nil {
			return
		}
		updateInvoiceFromWebhook(data)
	case "expired":
		updateInvoiceFromWebhook(data)
	default:
		log.Printf("Unhandled event type: %s\n", data.PaymentStatus)
		w.WriteHeader(http.StatusOK)
	}
}

// Helper
func closeCycle2(uid uuid.UUID, oldPaymentCycleId uuid.UUID, newPaymentCycleId uuid.UUID, currency string) (*UserBalance, error) {
	// get user balance for each currency
	fmt.Println(currency)
	currencies, _ := findAllCurreniesFromUserBalance(oldPaymentCycleId)
	if !contains(currencies, currency) {
		currencies = append(currencies, currency)
	}

	var ubNew *UserBalance

	for _, currency := range currencies {
		oldSum, err := findSumUserBalance2(uid, oldPaymentCycleId, currency)
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

func updateInvoiceFromWebhook(data NowpaymentWebhookResponse) {
	invoice, _ := getInvoice(data.InvoiceId)

	invoice.PaymentId.Int64 = data.PaymentId
	invoice.PriceAmount = int64(data.PriceAmount * 1_000_000)
	invoice.PriceCurrency = data.PriceCurrency
	invoice.PayAmount.Int64 = int64(data.PayAmount * 1_000_000_000)
	invoice.PayCurrency = data.PayCurrency
	invoice.ActuallyPaid.Int64 = int64(data.ActuallyPaid * 1_000_000_000)
	invoice.OutcomeAmount.Int64 = int64(data.OutcomeAmount * 1_000_000_000)
	invoice.OutcomeCurrency.String = data.OutcomeCurrency
	invoice.PaymentStatus = data.PaymentStatus
	invoice.LastUpdate = timeNow().Format(time.RFC3339)

	err := updateInvoice(invoice)
	if err != nil {
		return
	}
}

func getPaymentInformation(r *http.Request) (PaymentInformation, error) {
	p := mux.Vars(r)
	f := p["freq"]
	s := p["seats"]
	seats, err := strconv.Atoi(s)
	if err != nil {
		return PaymentInformation{}, errors.New(fmt.Sprintf("Cannot convert number seats: %v", +seats))
	}

	freq, err := strconv.Atoi(f)
	if err != nil {
		return PaymentInformation{}, errors.New(fmt.Sprintf("Cannot convert number freq: %v", freq))
	}

	var plan *Plan
	for _, v := range plans {
		if v.Freq == freq {
			plan = &v
			break
		}
	}
	if plan == nil {
		return PaymentInformation{}, errors.New(fmt.Sprintf("No matching plan found: %v, available: %v", freq, plans))
	}

	return PaymentInformation{freq, seats, plan}, nil
}

func getPrice(paymentInformation PaymentInformation) int64 {
	var f1 big.Float
	f1.SetString("100")
	cents, _ := f1.Mul(&paymentInformation.Plan.Price, &f1).Int64()
	cents = cents * int64(paymentInformation.Seats)
	usd := cents / 100
	return usd
}

func verifyNowpaymentsWebhook(data []byte, nowpaymentsSignature string) bool {
	key := opts.NowpaymentsIpnKey
	mac := hmac.New(sha512.New, []byte(key))

	io.WriteString(mac, string(data))
	expectedMAC := mac.Sum(nil)

	return hex.EncodeToString(expectedMAC) == nowpaymentsSignature
}

func nowpaymentSuccess(u *User, newPaymentCycleId uuid.UUID, amount int64, currency string) error {
	_, err := findUserBalancesAndType(newPaymentCycleId, "PAYMENT", currency) // TODO: do wee need currency check?
	if err != nil {
		return err
	}
	/*if ub != nil {
		log.Printf("We already processed this event, we can safely ignore it: %v", ub)
		return nil
	}*/

	ubNew, err := closeCycle2(u.Id, u.PaymentCycleId, newPaymentCycleId, currency)
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

	isNewCurrencyPayment := true
	totalDaysLeft := 365 // add one year by default, because payment is successful

	dailyPayments, _ := findDailyPaymentByPaymentCycleId(u.PaymentCycleId)
	for _, dailyPayment := range dailyPayments {
		if dailyPayment.Currency == currency {
			isNewCurrencyPayment = false
		}
		totalDaysLeft += int(dailyPayment.DaysLeft)
		dailyPayment.PaymentCycleId = newPaymentCycleId
		dailyPayment.LastUpdate = time.Now()
		err = insertDailyPayment(dailyPayment)
	}

	daysLeft, err := findDaysLeftForCurrency(newPaymentCycleId, currency)
	newDaysLeft := daysLeft + 365
	balance, err := findSumUserBalance2(u.Id, newPaymentCycleId, currency)
	newDailyPaymentAmount := balance / newDaysLeft

	newDailyPayment := DailyPayment{newPaymentCycleId, currency, newDailyPaymentAmount, newDaysLeft, time.Now()}

	if isNewCurrencyPayment {
		err = insertDailyPayment(newDailyPayment)
	} else {
		err = updateDailyPayment(newDailyPayment)
	}

	err = updatePaymentCycleDaysLeft(newPaymentCycleId, totalDaysLeft)
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

func crontester(w http.ResponseWriter, r *http.Request) {

	yesterdayStop, _ := time.Parse(time.RFC3339, "2021-10-31T23:59:59+00:00")
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	/*	repos, _ := runDailyAnalysisCheck(time.Now(), 5)
		log.Printf("Daily Analysis Check found %v entries", len(repos))*/

	/*	for _, v := range repos {
		if v.Url == nil {
			log.Printf("URL is nil of %v", v.Id)
			continue
		}
		if v.Branch == nil {
			log.Printf("Branch is nil of %v", v.Id)
			continue
		}
		_ = analysisRequest(v.Id, *v.Url, *v.Branch)
	}*/

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)
	nr, err := runDailyFutureLeftover(yesterdayStart, yesterdayStop, time.Now())
	if err != nil {
		log.Printf("error")
	}
	log.Printf("Daily Repo Hours inserted %v entries", nr)
}
