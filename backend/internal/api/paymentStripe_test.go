package api

import (
	"backend/internal/client"
	"backend/internal/db"
	"backend/pkg/util"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var (
	p = NewPaymentHandler(client.NewEmailClient("", "", "", "", "", "", ""), "webhooksecret", "webhooksecret")
)

func TestStripeConfirmsSuccessfulPayment(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	userDetail := insertTestUser(t, "hello@world.com")
	payInEvent := insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInRequest, "USD", Plans[1].PriceBase, 1, Plans[1].Freq)
	event, err := generateWebhookPayload(userDetail.Id.String(), payInEvent.ExternalId.String(), "payment_intent.succeeded")
	require.Nil(t, err)

	body, err := json.Marshal(event)
	require.Nil(t, err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hmacString := generateStripeSignature(&timestamp, &body)

	request, _ := http.NewRequest(http.MethodPost, "/hooks/stripe", bytes.NewReader(body))
	request.Header.Set("Stripe-Signature", getStripeSignatureHeaderContent(&timestamp, &hmacString))
	response := httptest.NewRecorder()

	p.StripeWebhook(response, request)

	assert.Equal(t, 200, response.Code)

	// this should create two events now
	// one that confirms the pay in
	// and another one that stores the fees
	successPayIn, err := db.FindPayInExternal(payInEvent.ExternalId, db.PayInSuccess)
	assert.Nil(t, err)
	assert.NotNil(t, successPayIn)
	assert.Equal(t, int64(120451199), successPayIn.Balance.Int64())

	feePayIn, err := db.FindPayInExternal(payInEvent.ExternalId, db.PayInFee)
	assert.Nil(t, err)
	assert.NotNil(t, feePayIn)
	assert.Equal(t, int64(5018801), feePayIn.Balance.Int64())
}

func TestStripeRequiresActionToContinuePayment(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	userDetail := insertTestUser(t, "hello@world.com")
	payInEvent := insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInRequest, "USD", Plans[1].PriceBase, 1, Plans[1].Freq)
	event, err := generateWebhookPayload(userDetail.Id.String(), payInEvent.ExternalId.String(), "payment_intent.requires_action")
	require.Nil(t, err)

	body, err := json.Marshal(event)
	require.Nil(t, err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hmacString := generateStripeSignature(&timestamp, &body)

	request, _ := http.NewRequest(http.MethodPost, "/hooks/stripe", bytes.NewReader(body))
	request.Header.Set("Stripe-Signature", getStripeSignatureHeaderContent(&timestamp, &hmacString))
	response := httptest.NewRecorder()

	p.StripeWebhook(response, request)

	assert.Equal(t, 200, response.Code)

	// this should create a pay in action event
	actionPayIn, err := db.FindPayInExternal(payInEvent.ExternalId, db.PayInAction)
	assert.Nil(t, err)
	assert.NotNil(t, actionPayIn)
	assert.Equal(t, Plans[1].PriceBase, actionPayIn.Balance.Int64())
}

func TestStripeMissesPaymentMethod(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	userDetail := insertTestUser(t, "hello@world.com")
	payInEvent := insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInRequest, "USD", Plans[1].PriceBase, 1, Plans[1].Freq)
	event, err := generateWebhookPayload(userDetail.Id.String(), payInEvent.ExternalId.String(), "payment_intent.payment_failed")
	require.Nil(t, err)

	event.Data.Object = make(map[string]interface{})
	event.Data.Object["status"] = "requires_payment_method"

	body, err := json.Marshal(event)
	require.Nil(t, err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hmacString := generateStripeSignature(&timestamp, &body)

	request, _ := http.NewRequest(http.MethodPost, "/hooks/stripe", bytes.NewReader(body))
	request.Header.Set("Stripe-Signature", getStripeSignatureHeaderContent(&timestamp, &hmacString))
	response := httptest.NewRecorder()

	p.StripeWebhook(response, request)

	assert.Equal(t, 200, response.Code)
}

func TestStripeHasIssue(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	userDetail := insertTestUser(t, "hello@world.com")
	payInEvent := insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInRequest, "USD", Plans[1].PriceBase, 1, Plans[1].Freq)
	event, err := generateWebhookPayload(userDetail.Id.String(), payInEvent.ExternalId.String(), "payment_intent.payment_failed")
	require.Nil(t, err)

	body, err := json.Marshal(event)
	require.Nil(t, err)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hmacString := generateStripeSignature(&timestamp, &body)

	request, _ := http.NewRequest(http.MethodPost, "/hooks/stripe", bytes.NewReader(body))
	request.Header.Set("Stripe-Signature", getStripeSignatureHeaderContent(&timestamp, &hmacString))
	response := httptest.NewRecorder()

	p.StripeWebhook(response, request)

	assert.Equal(t, 200, response.Code)

	// this should create a pay in method event
	actionPayIn, err := db.FindPayInExternal(payInEvent.ExternalId, db.PayInMethod)
	assert.Nil(t, err)
	assert.NotNil(t, actionPayIn)
	assert.Equal(t, Plans[1].PriceBase, actionPayIn.Balance.Int64())
}

func generateStripeSignature(timestamp *string, body *[]byte) string {
	hasher := hmac.New(sha256.New, []byte("webhooksecret"))
	hasher.Write(append([]byte(*timestamp+"."), *body...))
	hmacBytes := hasher.Sum(nil)
	return hex.EncodeToString(hmacBytes)
}

func getStripeSignatureHeaderContent(timestamp *string, hmacString *string) string {
	return "t=" + *timestamp + ",v1=" + *hmacString
}

func generateWebhookPayload(userId string, externalId string, eventType stripe.EventType) (stripe.Event, error) {
	metadata := make(map[string]string)
	metadata["userId"] = userId
	metadata["externalId"] = externalId
	metadata["freq"] = strconv.FormatInt(365, 10)
	metadata["seats"] = strconv.Itoa(1)
	metadata["fee"] = strconv.FormatInt(40, 10)

	paymentIntent := stripe.PaymentIntent{
		Amount:   util.UsdBaseToCent(Plans[1].PriceBase),
		Metadata: metadata,
	}

	jsonBytes, err := json.Marshal(paymentIntent)
	if err != nil {
		return stripe.Event{}, err
	}

	eventData := stripe.EventData{
		Raw: json.RawMessage(jsonBytes),
	}
	return stripe.Event{
		APIVersion: "2023-10-16",
		Data:       &eventData,
		Type:       eventType,
	}, nil
}
