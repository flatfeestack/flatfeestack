package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Payment struct {
	PriceAmount    string `json:"price_amount"`
	PriceCurrency  string `json:"price_currency"`
	PayCurrency    string `json:"pay_currency"`
	IpnCallbackUrl string `json:"ipn_callback_url"`
}

type Invoice struct {
	PriceAmount    string `json:"price_amount"`
	PriceCurrency  string `json:"price_currency"`
	PayCurrency    string `json:"pay_currency"`
	OrderId        string `json:"order_id"`
	IpnCallbackUrl string `json:"ipn_callback_url"`
}

func nowpaymentPayment(w http.ResponseWriter, r *http.Request, user *User) {

	priceAmount := "120" // calculate -> (120USD * seats) - current amount
	priceCurrency := "USD"
	payCurrency := "Neo"    //get from request
	newPaymentCycleId := "" // create

	invoice := Invoice{priceAmount, priceCurrency, payCurrency, newPaymentCycleId, ""}

	createNowpaymentsInvoice(invoice, user)
}

func createNowpaymentsInvoice(invoice Invoice, user *User) {
	/*paymentUrl := "https://api.sandbox.nowpayments.io/v1/payment"
	sandboxToken := ""

	payment := Payment{"120", "USD", "Neo", "https://316c-31-10-156-230.ngrok.io/nowpayments/ipn"}
	paymentData, err := json.Marshal(payment)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", paymentUrl, strings.NewReader(string(paymentData)))
	req.Header.Set("x-api-key", sandboxToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()*/

	invoiceUrl := "https://api.nowpayments.io/v1/invoice"
	apiToken := ""
	invoice.IpnCallbackUrl = "https://316c-31-10-156-230.ngrok.io/nowpayments/nowpaymentsWebhook"

	//invoice := Payment{"120", "USD", "Neo", "https://316c-31-10-156-230.ngrok.io/nowpayments/nowpaymentsWebhook"}
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
