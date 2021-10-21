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

type Invoice struct {
	PriceAmount    int64  `json:"price_amount"`
	PriceCurrency  string `json:"price_currency"`
	PayCurrency    string `json:"pay_currency"`
	IpnCallbackUrl string `json:"ipn_callback_url"`
}

type InvoiceDB struct {
	NowpaymentsInvoiceId int64
	PaymentCycleId       *uuid.UUID
	PriceAmount          int64
	PriceCurrency        string
	PayAmount            int64
	ActuallyPaid         int64
	PayCurrency          string
	CreatedAt            string
	paidAt               string
}

type PaymentInformation struct {
	Freq  int
	Seats int
	Plan  *Plan
}

func nowpaymentPayment(w http.ResponseWriter, r *http.Request, user *User) {
	paymentInformation, err := getPaymentInformation(r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	usdPrice := getPrice(paymentInformation)

	priceCurrency := "USD"
	payCurrency := "XTZ" //get from request

	paymentCycleId, err := insertNewPaymentCycle(user.Id, paymentInformation.Freq, paymentInformation.Seats, paymentInformation.Freq, timeNow())
	if err != nil {
		log.Printf("Cannot insert payment for %v: %v\n", user.Id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invoice := Invoice{usdPrice, priceCurrency, payCurrency, ""}

	createNowpaymentsInvoice(invoice, paymentCycleId)
}

func createNowpaymentsInvoice(invoice Invoice, paymentCycleId *uuid.UUID) {
	invoiceUrl := "https://api.nowpayments.io/v1/invoice"
	apiToken := "XXX"
	invoice.IpnCallbackUrl = "https://316c-31-10-156-230.ngrok.io/nowpayments/nowpaymentsWebhook"

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

	var data map[string]string
	json.NewDecoder(response.Body).Decode(&data)
	id, err := strconv.ParseInt(data["id"], 10, 64)
	priceAmount, err := strconv.ParseInt(data["price_amount"], 10, 64)

	db := InvoiceDB{
		NowpaymentsInvoiceId: id,
		PaymentCycleId:       paymentCycleId,
		PriceAmount:          120 * 1_000_000,
		PayAmount:            priceAmount * 1_000_000_000,
		PayCurrency:          data["pay_currency"],
		PriceCurrency:        data["price_currency"],
		CreatedAt:            data["created_at"],
	}

	_, _ = insertNewInvoice(db)
	fmt.Println("-------------------------------")
	fmt.Println(data)
	fmt.Println("-------------------------------")
	fmt.Println(db)
	fmt.Println("-------------------------------")
}

func nowpaymentsWebhook(w http.ResponseWriter, r *http.Request) {
	// get data to json
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)

	// verify request
	key := ""
	mac := hmac.New(sha512.New, []byte(key))

	json, err := json.Marshal(data)
	io.WriteString(mac, string(json))
	expectedMAC := mac.Sum(nil)

	sig := r.Header.Get("x-nowpayments-sig")
	fmt.Println("------------------")
	fmt.Println(string(sig))
	fmt.Println("------------------")
	fmt.Println(hex.EncodeToString(expectedMAC))
	fmt.Println("------------------")

	fmt.Println("------------------")
	fmt.Println(err)
	fmt.Println("------------------")
	fmt.Println(string(json))
	fmt.Println("------------------")
	fmt.Println(data)

	// close payment cycle --> close cycle for all currencies the user has

	// insertUserBalance

	// updatePaymentCycleId
}

// Helper
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
