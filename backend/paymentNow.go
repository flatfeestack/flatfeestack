package main

import (
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
	"io/ioutil"
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

type NowpaymentWebhookResponse struct {
	ActuallyPaid     float64    `json:"actually_paid"`
	InvoiceId        int64      `json:"invoice_id"`
	OrderDescription *uuid.UUID `json:"order_description"`
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

func nowPayment(w http.ResponseWriter, r *http.Request, user *User) {
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
	paymentResponse, err := createNowPayment(price, payCurrency, paymentCycleId, &user.Id)

	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create payment: %v", err)
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
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func createNowPayment(price float64, payCurrency string, paymentCycleId *uuid.UUID, uid *uuid.UUID) (*PaymentResponse, error) {
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

func nowWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	nowSignature := r.Header.Get("x-nowpayments-sig")
	err = verifyNowWebhook(body, nowSignature)
	if err != nil {
		log.Printf("Wrong signature: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data NowpaymentWebhookResponse
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Could not parse webhook data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := *data.OrderDescription
	pc, err := findPaymentCycle(*data.OrderId)

	if err != nil {
		log.Printf("Could not parse freq: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	freq := pc.Freq
	seat := pc.Seats

	user, err := findUserById(userId)
	if err != nil {
		log.Printf("Could not find user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//outcome_amount - this parameter shows the amount that will be (or is already) received on your Outcome Wallet once the transaction is settled.
	amount, err := minCrypto(data.OutcomeCurrency, data.OutcomeAmount)
	if err != nil {
		log.Printf("Could not find currency: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch data.PaymentStatus {
	case "finished":
		if err != nil {
			log.Printf("Could not process nowpayment success: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = paymentSuccess(user, *data.OrderId, amount, strings.ToUpper(data.PayCurrency), seat, freq, big.NewInt(0))
		if err != nil {
			log.Printf("Could not process nowpayment success: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		email := user.Email
		var other = map[string]string{}
		other["email"] = email
		other["url"] = opts.EmailLinkPrefix + "/user/payments"
		other["lang"] = "en"

		defaultMessage := "Crypto payment successful. See your payment here: " + other["url"]
		e := prepareEmail(email, other,
			"template-subject-success_", "Payment successful",
			"template-plain-success_", defaultMessage,
			"template-html-success_", other["lang"])

		err = sendToBrowser(user.Id, *data.OrderId)
		if err != nil {
			log.Debugf("browser offline, best effort, we write a email to %s anyway", email)
		}
		go func(uid uuid.UUID, paymentCycleId uuid.UUID, e EmailRequest) {
			insertEmailSent(uid, "success-"+paymentCycleId.String(), timeNow())
			err = sendEmail(opts.EmailUrl, e)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}(user.Id, *data.OrderId, e)

	case "partially_paid":
		ub, err := findUserBalancesAndType(*data.OrderId, "PART_PAID", strings.ToUpper(data.PayCurrency))
		if err != nil {
			log.Printf("Error find user balance: %v, %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}

		ubNew := UserBalance{
			PaymentCycleId: *data.OrderId,
			UserId:         user.Id,
			Balance:        big.NewInt(0),
			BalanceType:    "PART_PAID",
			Currency:       strings.ToUpper(data.PayCurrency),
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = sendToBrowser(user.Id, *data.OrderId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)

			email := user.Email
			var other = map[string]string{}
			other["email"] = email
			other["url"] = opts.EmailLinkPrefix + "/user/payments"
			other["lang"] = "en"

			defaultMessage := fmt.Sprintf("Only partial payment received (%v) of (%v), please send the rest (%v) to: ", data.ActuallyPaid, data.PayAmount, data.PayAmount-data.ActuallyPaid)
			e := prepareEmail(email, other,
				"template-subject-part_paid_", "Partially paid",
				"template-plain-part_paid_", defaultMessage,
				"template-html-part_paid_", other["lang"])

			go func(uid uuid.UUID, paymentCycleId uuid.UUID, e EmailRequest) {
				insertEmailSent(uid, "failed-"+paymentCycleId.String(), timeNow())
				err = sendEmail(opts.EmailUrl, e)
				if err != nil {
					log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
				}
			}(user.Id, *data.OrderId, e)
		}
	case "expired":
	case "failed":
	case "refunded":
		suf := "NONE"
		switch data.PaymentStatus {
		case "expired":
			suf = "EXP"
		case "failed":
			suf = "FAIL"
		case "refunded":
			suf = "REF"
		}
		ub, err := findUserBalancesAndType(*data.OrderId, "FAILED_"+suf, strings.ToUpper(data.PayCurrency))
		if err != nil {
			log.Printf("Error find user balance: %v, %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}

		ubNew := UserBalance{
			PaymentCycleId: *data.OrderId,
			UserId:         user.Id,
			Balance:        big.NewInt(0),
			BalanceType:    "FAILED_" + suf,
			Currency:       strings.ToUpper(data.PayCurrency),
			CreatedAt:      timeNow(),
		}

		err = insertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = sendToBrowser(user.Id, *data.OrderId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)

			email := user.Email
			var other = map[string]string{}
			other["email"] = email
			other["url"] = opts.EmailLinkPrefix + "/user/payments"
			other["lang"] = "en"

			defaultMessage := fmt.Sprintf("Payment %v, please check payment: %s", data.PaymentStatus, other["url"])
			e := prepareEmail(email, other,
				"template-subject-failed_", "Payment "+data.PaymentStatus,
				"template-plain-failed_", defaultMessage,
				"template-html-failed_", other["lang"])

			go func(uid uuid.UUID, paymentCycleId uuid.UUID, e EmailRequest) {
				insertEmailSent(uid, "failed-"+suf+"-"+paymentCycleId.String(), timeNow())
				err = sendEmail(opts.EmailUrl, e)
				if err != nil {
					log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
				}
			}(user.Id, *data.OrderId, e)
		}
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

func verifyNowWebhook(data []byte, sig string) error {
	key := opts.NowpaymentsIpnKey
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
