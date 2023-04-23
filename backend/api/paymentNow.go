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

func NowPayment(w http.ResponseWriter, r *http.Request, user *db.User) {
	freq, seats, plan, err := paymentInformation(r)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Cannot get payment informations for now payments: %v", err)
		return
	}

	var data map[string]string
	err = json.NewDecoder(r.Body).Decode(&data)

	paymentCycleId := uuid.New()
	err = db.InsertNewPaymentCycleIn(paymentCycleId, seats, freq, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Cannot insert payment for %v: %v\n", user.Id, err)
		return
	}

	currUSDBalance := int64(0)
	if user.PaymentCycleInId != nil {
		currUSDBalance, err = currentUSDBalance(*user.PaymentCycleInId)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Cannot get current USD balance %v: %v\n", user.Id, err)
			return
		}
	}
	price := (plan.Price * float64(seats)) - float64(currUSDBalance)

	payCurrency := data["currency"]
	paymentResponse, err := createNowPayment(price, payCurrency, &paymentCycleId, &user.Id)

	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "could not create payment: %v", err)
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
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func createNowPayment(price float64, payCurrency string, paymentCycleId *uuid.UUID, uid *uuid.UUID) (*PaymentResponse, error) {
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

func NowWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not read body: %v", err)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	nowSignature := r.Header.Get("x-nowpayments-sig")
	err = verifyNowWebhook(body, nowSignature)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Wrong signature: %v", err)
		return
	}

	var data clnt.WebhookResponse
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not parse webhook data: %v", err)
		return
	}

	//todo: nil check
	paymentCycleInId := *data.OrderId
	pc, err := db.FindPaymentCycle(paymentCycleInId)

	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not parse freq: %v", err)
		return
	}
	freq := pc.Freq
	seat := pc.Seats

	//todo: nil check
	userId := *data.OrderDescription
	user, err := db.FindUserById(userId)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not find user: %v", err)
		return
	}
	//outcome_amount - this parameter shows the amount that will be (or is already) received on your Outcome Wallet once the transaction is settled.
	amount, err := minCrypto(data.OutcomeCurrency, data.OutcomeAmount)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not find currency: %v", err)
		return
	}

	switch data.PaymentStatus {
	case "finished":
		err = PaymentSuccess(user.Id, user.PaymentCycleInId, paymentCycleInId, amount, strings.ToUpper(data.PayCurrency), seat, freq, big.NewInt(0))
		if err != nil {
			utils.WriteErrorf(w, http.StatusInternalServerError, "Could not process nowpayment success: %v", err)
			return
		}

		err = SendToBrowser(user.Id, paymentCycleInId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)
		}

		clnt.SendPaymentNowFinished(user, data)
	case "partially_paid":
		var ub *db.UserBalance
		ub, err = db.FindBalance(*data.OrderId, user.Id, "PART_PAID", strings.ToUpper(data.PayCurrency))
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Error find user balance: %v, %v\n", user.Id, err)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}
		if data.OrderId == nil {
			log.Printf("payment cycle is nil: %v", data)
			return
		}

		ubNew := db.UserBalance{
			PaymentCycleInId: *data.OrderId,
			UserId:           user.Id,
			Balance:          big.NewInt(0),
			BalanceType:      "PART_PAID",
			Currency:         strings.ToUpper(data.PayCurrency),
			CreatedAt:        utils.TimeNow(),
		}

		err = db.InsertUserBalance(ubNew)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Insert user balance for %v: %v\n", user.Id, err)
			return
		}

		err = SendToBrowser(user.Id, *data.OrderId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)
		}

		clnt.SendPaymentNowPartially(*user, data)
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
		var ub *db.UserBalance
		ub, err = db.FindBalance(*data.OrderId, user.Id, "FAILED_"+suf, strings.ToUpper(data.PayCurrency))
		if err != nil {
			log.Printf("Error find user balance: %v, %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ub != nil {
			log.Printf("We already processed this event, we can safely ignore it: %v", ub)
			return
		}
		if data.OrderId == nil {
			log.Printf("payment cycle is nil: %v", data)
			return
		}

		ubNew := db.UserBalance{
			PaymentCycleInId: *data.OrderId,
			UserId:           user.Id,
			Balance:          big.NewInt(0),
			BalanceType:      "FAILED_" + suf,
			Currency:         strings.ToUpper(data.PayCurrency),
			CreatedAt:        utils.TimeNow(),
		}

		err = db.InsertUserBalance(ubNew)
		if err != nil {
			log.Printf("Insert user balance for %v: %v\n", user.Id, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = SendToBrowser(user.Id, *data.OrderId)
		if err != nil {
			log.Infof("browser seems offline, need to send email %v", err)
		}

		clnt.SendPaymentNowRefunded(*user, data, suf)
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

	freq, err := strconv.ParseInt(f, 10, 64)
	if err != nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("Cannot convert number freq: %v", freq))
	}

	var plan *Plan
	for _, v := range Plans {
		if v.Freq == freq {
			plan = &v
			break
		}
	}
	if plan == nil {
		return 0, 0, nil, errors.New(fmt.Sprintf("No matching plan found: %v, available: %v", freq, Plans))
	}

	return freq, seats, plan, nil
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
