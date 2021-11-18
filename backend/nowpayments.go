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
	Freq          int
	InvoiceUrl    sql.NullString
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

func nowpaymentsPayment(w http.ResponseWriter, r *http.Request, user *User) {
	paymentInformation, err := getPaymentInformation(r)
	if err != nil {
		log.Printf("Cannot get payment informations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)

	usdPrice := getPrice(paymentInformation)

	priceCurrency := "USD"
	payCurrency := data["currency"]

	paymentCycleId, err := insertNewPaymentCycle(user.Id, 0, paymentInformation.Seats, paymentInformation.Freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invoice := InvoiceRequest{float64(usdPrice), priceCurrency, payCurrency, ""}
	invoiceUrl, err := createNowpaymentsInvoice(invoice, paymentCycleId, paymentInformation.Freq)

	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create invoice: %v", err)
		return
	}
	writeJsonStr(w, `{ "invoice_url": "`+invoiceUrl+`" }`)
}

func createNowpaymentsInvoice(invoice InvoiceRequest, paymentCycleId *uuid.UUID, freq int) (string, error) {
	invoiceUrl := opts.NowpaymentsApiUrl + "/invoice"
	apiToken := opts.NowpaymentsToken
	invoice.IpnCallbackUrl = opts.NowpaymentsIpnCallbackUrl
	invoiceData, err := json.Marshal(invoice)

	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", invoiceUrl, strings.NewReader(string(invoiceData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", apiToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Body ReadCloser: %v", err)
		}
	}(response.Body)

	var data InvoiceResponse
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	id, err := strconv.ParseInt(data.Id, 10, 64)
	if err != nil {
		return "", err
	}
	priceAmount, err := strconv.ParseInt(data.PriceAmount, 10, 64)
	if err != nil {
		return "", err
	}
	invoiceDb := InvoiceDB{
		NowpaymentsInvoiceId: id,
		PaymentCycleId:       paymentCycleId,
		PriceAmount:          priceAmount * cryptoFactor,
		PriceCurrency:        data.PriceCurrency,
		PayCurrency:          data.PayCurrency,
		PaymentStatus:        "CREATED",
		Freq:                 freq,
		CreatedAt:            data.CreatedAt,
		LastUpdate:           data.CreatedAt,
	}
	invoiceDb.InvoiceUrl.String = data.InvoiceUrl

	err = insertNewInvoice(invoiceDb)

	return data.InvoiceUrl, err
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
	if err != nil {
		log.Printf("Could not verify Webhook data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isWebhookVerified {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invoice, err := getInvoice(data.InvoiceId)
	if err != nil {
		log.Printf("Could not get invoice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userId, err := findUserIdByInvoice(invoice.NowpaymentsInvoiceId)
	if err != nil {
		log.Printf("Could not find user ID by invoice: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := findUserById(userId)
	if err != nil {
		log.Printf("Could not find user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	amount := int64(data.OutcomeAmount * cryptoFactor) // nanoCrypto

	switch data.PaymentStatus {
	case "finished":
		err := nowpaymentsSuccess(user, *invoice.PaymentCycleId, amount, data.PayCurrency, invoice.Freq, 0)
		if err != nil {
			log.Printf("Could not process nowpayment success: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = updateInvoiceFromWebhook(data)
		if err != nil {
			log.Printf("Could update Invoice: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case "expired":
		err := updateInvoiceFromWebhook(data)
		if err != nil {
			log.Printf("Could update Invoice: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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

		go func() {
			insertEmailSent(user.Id, "payment-expired-"+invoice.PaymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "partially_paid":
		err := updateInvoiceFromWebhook(data)
		if err != nil {
			log.Printf("Could update Invoice: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = invoice.InvoiceUrl.String
		other["lang"] = "en"

		defaultMessage := "Your payment is partially paid. Please transfer the missing amount over the following link: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-payment-expired_", "Partially paid",
			"template-plain-partially-paid_", defaultMessage,
			"template-html-partially-paid_", other["lang"])

		go func() {
			insertEmailSent(user.Id, "partially-paid-"+invoice.PaymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "failed":
		err := updateInvoiceFromWebhook(data)
		if err != nil {
			log.Printf("Could update Invoice: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["lang"] = "en"

		defaultMessage := "Your payment has failed. Please contact support@flatfeestack.io"
		e := prepareEmail(email, other,
			"template-subject-payment-failed_", "Payment failed",
			"template-plain-partially-failed_", defaultMessage,
			"template-html-partially-failed_", other["lang"])

		go func() {
			insertEmailSent(user.Id, "payment-failed-"+invoice.PaymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()
	case "refunded":
		err := updateInvoiceFromWebhook(data)
		if err != nil {
			log.Printf("Could update Invoice: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["lang"] = "en"

		defaultMessage := "You got your money refunded."
		e := prepareEmail(email, other,
			"template-subject-payment-refunded_", "Payment refunded",
			"template-plain-partially-refunded_", defaultMessage,
			"template-html-partially-refunded_", other["lang"])

		go func() {
			insertEmailSent(user.Id, "payment-refunded-"+invoice.PaymentCycleId.String(), timeNow())
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

// Helper
func closeCycleCrypto(uid uuid.UUID, oldPaymentCycleId uuid.UUID, newPaymentCycleId uuid.UUID, currency string) (*UserBalance, error) {
	currencies, err := findAllCurreniesFromUserBalance(oldPaymentCycleId)
	if err != nil {
		return nil, err
	}
	if !contains(currencies, currency) {
		currencies = append(currencies, currency)
	}

	var ubNew *UserBalance
	for _, currency := range currencies {
		oldSum, err := findSumUserBalanceByCurrency(uid, oldPaymentCycleId, currency)
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

func updateInvoiceFromWebhook(data NowpaymentWebhookResponse) error {
	invoice, _ := getInvoice(data.InvoiceId)

	invoice.PaymentId.Int64 = data.PaymentId
	invoice.PriceAmount = int64(data.PriceAmount * usdFactor)
	invoice.PriceCurrency = data.PriceCurrency
	invoice.PayAmount.Int64 = int64(data.PayAmount * cryptoFactor)
	invoice.PayCurrency = data.PayCurrency
	invoice.ActuallyPaid.Int64 = int64(data.ActuallyPaid * cryptoFactor)
	invoice.OutcomeAmount.Int64 = int64(data.OutcomeAmount * cryptoFactor)
	invoice.OutcomeCurrency.String = data.OutcomeCurrency
	invoice.PaymentStatus = data.PaymentStatus
	invoice.LastUpdate = timeNow().Format(time.RFC3339)

	err := updateInvoice(*invoice)
	if err != nil {
		return err
	}
	return nil
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

func nowpaymentsSuccess(u *User, newPaymentCycleId uuid.UUID, amount int64, currency string, freq int, fee int64) error {
	_, err := findUserBalancesAndType(newPaymentCycleId, "PAYMENT", currency)
	if err != nil {
		return err
	}
	/*	if ub != nil {
		log.Printf("We already processed this event, we can safely ignore it: %v", ub)
		return nil
	}*/

	ubNew, err := closeCycleCrypto(u.Id, u.PaymentCycleId, newPaymentCycleId, currency)
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

	if currency == "USD" {
		ubNew.BalanceType = "FEE"
		ubNew.Balance = -fee
		err = insertUserBalance(*ubNew)
		if err != nil {
			return err
		}
	}

	isNewCurrencyPayment := true
	totalDaysLeft := freq // add one year by default, because payment is successful

	dailyPayments, err := findDailyPaymentByPaymentCycleId(u.PaymentCycleId)
	if err != nil {
		return err
	}
	for _, dailyPayment := range dailyPayments {
		if dailyPayment.Currency == currency {
			isNewCurrencyPayment = false
		}
		totalDaysLeft += dailyPayment.DaysLeft
		dailyPayment.PaymentCycleId = newPaymentCycleId
		dailyPayment.LastUpdate = time.Now()
		err = insertDailyPayment(dailyPayment)
	}

	daysLeft, err := findDaysLeftForCurrency(newPaymentCycleId, currency)
	newDaysLeft := daysLeft + freq
	balance, err := findSumUserBalanceByCurrency(u.Id, newPaymentCycleId, currency)
	newDailyPaymentAmount := balance / int64(newDaysLeft)
	if currency == "USD" {
		newDailyPaymentAmount = mUSDPerDay
	}

	newDailyPayment := DailyPayment{newPaymentCycleId, currency, newDailyPaymentAmount, newDaysLeft, time.Now()}

	if isNewCurrencyPayment {
		err = insertDailyPayment(newDailyPayment)
		if err != nil {
			return err
		}
	} else {
		err = updateDailyPayment(newDailyPayment)
		if err != nil {
			return err
		}
	}

	err = updatePaymentCycleDaysLeft(newPaymentCycleId, int64(totalDaysLeft))
	if err != nil {
		return err
	}
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

// Wallet
func getUserWallets(w http.ResponseWriter, _ *http.Request, user *User) {
	userWallets, err := findWalletsByUserId(user.Id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJson(w, userWallets)
}

func addUserWallet(w http.ResponseWriter, r *http.Request, user *User) {
	var data Wallet
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = insertWallet(user.Id, data.Currency, data.Address, false)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// ToDo: verify that wallet is from the right user
func deleteUserWallet(w http.ResponseWriter, r *http.Request, user *User) {
	p := mux.Vars(r)
	f := p["uuid"]
	id, _ := uuid.Parse(f)

	err := deleteWallet(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
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

//Todo: remove only for cron testing
func crontester(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return
	}

	yesterdayStop, _ := time.Parse(time.RFC3339, data["yesterdayStop"])
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	_, err = runDailyUserBalance(yesterdayStart, yesterdayStop, time.Now())

	_, err = runDailyDaysLeftDailyPayment()

	_, err = runDailyDaysLeftPaymentCycle()

	_, err = runDailyRepoBalance(yesterdayStart, yesterdayStop, time.Now())

	_, err = runDailyRepoWeight(yesterdayStart, yesterdayStop, time.Now())

	_, err = runDailyUserPayout(yesterdayStart, yesterdayStop, time.Now())

	// _, err = runDailyUserContribution(yesterdayStart, yesterdayStop, time.Now())

	_, err = runDailyFutureLeftover(yesterdayStart, yesterdayStop, time.Now())

	//_, err = runDailyEmailPayout(yesterdayStart, yesterdayStop, time.Now())

	// _, err = runDailyAnalysisCheck(time.Now(), 5) --> sollte keine änderungen brauchen

	// _, err = runDailyTopupReminderUser()

	// _, err = runDailyMarketing(yesterdayStart) --> currency änderung
}
