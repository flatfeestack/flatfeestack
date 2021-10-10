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
	PriceAmount   string `json:"price_amount"`
	PriceCurrency string `json:"price_currency"`
	PayCurrency   string `json:"pay_currency"`
}

func nowpaymentstest(w http.ResponseWriter, r *http.Request) {
	paymentUrl := "https://api.sandbox.nowpayments.io/v1/payment"
	sandboxToken := "JA6GXF8-DY4MJ9H-H6NQD53-E89DQJ2"

	payment := Payment{"120", "USD", "Eth", "https://21f9-31-10-156-230.ngrok.io/nowpayments/ipn"}
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
	defer response.Body.Close()

	/*invoiceUrl := "https://api.nowpayments.io/v1/invoice"
	apiToken := "R6WFT2D-7TC4ZK2-K2P43ES-P4AD2G8"

	invoice := Payment{"120", "USD", "Eth"}
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
	defer response.Body.Close()*/

	// https://api.sandbox.nowpayments.io/v1/payment

	// price_amount (required) - price in Fiat currency
	// price_currency (required) - Fiat currency price
	// pay_amount (optional) - price in cryptocurrency
	// pay_currency (required) - cryptocurrency price
	// ipn_callback_url (optional) - url to receive callbacks, should contain "http" or "https", eg. "https://nowpayments.io"
	// order_id (optional) - inner store order ID, e.g. "RGDBP-21314"
	// order_description (optional) - inner store order description, e.g. "Apple Macbook Pro 2019 x 1"
	// purchase_id (optional) - id of purchase for which you want to create aother payment, only used for several payments for one order
	// payout_address (optional) - in case you want to receive funds on an external address, you can specify it in this parameter
	// payout_currency (optional) - currency of your external payout_address, required when payout_adress is specified.
	// payout_extra_id(optional) - extra id or memo or tag for external payout_address.
	// case(optional) - case which you want to test.

	// https://api.nowpayments.io/v1/invoice
	// price_amount = 120
	// price_currency = USD
	// pay_currency = Neo, Tezos, Eth
	// ipn_callback_url
	// order_id (optional) - internal store order ID, e.g. "RGDBP-21314"
	// order_description (optional) - internal store order description, e.g. "Apple Macbook Pro 2019 x 1"
	// success_url(optional) - url where the customer will be redirected after successful payment.
	// cancel_url(optional) - url where the customer will be redirected after failed payment.
}

func ipn(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)

	key := "SYqQtcEUdoPbfjdsI6aIdOZFE4I4XG07"
	mac := hmac.New(sha512.New, []byte(key))

	json, err := json.Marshal(data)
	// mac.Write(json)
	io.WriteString(mac, string(json))
	expectedMAC := mac.Sum(nil)

	sig := r.Header.Get("X-Nowpayments-Sig")
	fmt.Println("------------------")
	fmt.Println(string(sig))
	fmt.Println("------------------")
	fmt.Println(hex.EncodeToString(expectedMAC))
	fmt.Println("------------------")
	fmt.Println(hex.EncodeToString(expectedMAC) == string(sig))
	fmt.Println("------------------")
	fmt.Println(err)
	fmt.Println("------------------")
	fmt.Println(string(json))
	fmt.Println("------------------")
	fmt.Println(data)

	//fmt.Print(json)

	//fmt.Print(sig)
	//fmt.Println(r.Body)
	//fmt.Println(r.GetBody())
}
