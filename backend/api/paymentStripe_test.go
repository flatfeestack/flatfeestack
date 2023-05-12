package api

import (
	"backend/db"
	"backend/utils"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v74"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestPostHookStripe(t *testing.T) {
	stripeWebhookSecretKey = "webhooksecret"

	t.Run("stripe confirms successful payment", func(t *testing.T) {
		setup()
		defer teardown()

		userDetail := insertTestUser(t, "hello@world.com")
		payInEvent := insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInRequest, "USD", Plans[0].PriceBase, 1, Plans[0].Freq)

		metadata := make(map[string]string)
		metadata["userId"] = userDetail.Id.String()
		metadata["externalId"] = payInEvent.ExternalId.String()
		metadata["freq"] = strconv.FormatInt(365, 10)
		metadata["seats"] = strconv.Itoa(1)
		metadata["fee"] = strconv.FormatInt(40, 10)

		paymentIntent := stripe.PaymentIntent{
			Amount:   utils.UsdBaseToCent(Plans[0].PriceBase),
			Metadata: metadata,
		}

		jsonBytes, err := json.Marshal(paymentIntent)
		require.Nil(t, err)

		eventData := stripe.EventData{
			Raw: json.RawMessage(jsonBytes),
		}
		event := stripe.Event{
			APIVersion: "2022-11-15",
			Data:       &eventData,
			Type:       "payment_intent.succeeded",
		}
		body, err := json.Marshal(event)
		require.Nil(t, err)

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		hasher := hmac.New(sha256.New, []byte("webhooksecret"))
		hasher.Write(append([]byte(timestamp+"."), body...))
		hmacBytes := hasher.Sum(nil)
		hmacString := hex.EncodeToString(hmacBytes)

		request, _ := http.NewRequest(http.MethodPost, "/hooks/stripe", bytes.NewReader(body))
		request.Header.Set("Stripe-Signature", "t="+timestamp+",v1="+hmacString)
		response := httptest.NewRecorder()

		StripeWebhook(response, request)

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
	})
}
