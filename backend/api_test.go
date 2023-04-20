package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostFakePayment(t *testing.T) {
	debug = true

	t.Run("should throw error if user does not exist", func(t *testing.T) {
		setup()
		defer teardown()

		request, _ := http.NewRequest(http.MethodPost, "/admin/fake/payment/shouldnotexist/1", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"email": "shouldnotexist",
			"seats": "1",
		}
		request = mux.SetURLVars(request, vars)

		fakePayment(response, request, "ignored")

		assert.Equal(t, 400, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Containsf(t, string(body), "Unable to find user from given e-mail address", "error message %s")
	})

	t.Run("should create new payment cycle", func(t *testing.T) {
		setup()
		defer teardown()

		u := User{
			Id:    uuid.New(),
			Email: "email",
		}

		err := insertUser(&u)
		require.Nil(t, err)
		oldUserBalances, err := findUserBalances(u.Id)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/admin/fake/payment/email/1", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"email": u.Email,
			"seats": "1",
		}
		request = mux.SetURLVars(request, vars)

		fakePayment(response, request, "ignored")

		assert.Equal(t, 200, response.Code)

		userBalances, err := findUserBalances(u.Id)
		require.Nil(t, err)
		assert.Equal(t, len(oldUserBalances)+1, len(userBalances))

		createdUserBalance := userBalances[len(userBalances)-1]
		assert.Equal(t, "PAY", createdUserBalance.BalanceType)
		assert.Equal(t, big.NewInt(120000000), createdUserBalance.Balance)
		assert.Equal(t, "USD", createdUserBalance.Currency)
	})
}

func TestPostPayoutRequest(t *testing.T) {
	debug = true

	t.Run("should complain when requesting invalid currency", func(t *testing.T) {
		setup()
		defer teardown()

		payOutId := uuid.New()
		user := User{
			Id:                uuid.New(),
			StripeId:          stringPointer("strip-id"),
			PaymentCycleOutId: payOutId,
			Email:             "email",
		}

		request, _ := http.NewRequest(http.MethodPost, "/users/me/request-payout/yikes", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"targetCurrency": "yikes",
		}
		request = mux.SetURLVars(request, vars)

		requestPayout(response, request, &user)

		assert.Equal(t, 400, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Containsf(t, string(body), "Unsupported currency requested", "error message %s")
	})

	t.Run("should raise when new contributions are less than 1 dollar", func(t *testing.T) {
		setup()
		defer teardown()

		payOutId := uuid.New()

		repoId, err := setupRepo("github.com/hello-world")
		require.Nil(t, err)

		user := User{
			Id:                uuid.New(),
			StripeId:          stringPointer("strip-id"),
			PaymentCycleOutId: payOutId,
			Email:             "email",
		}
		err = insertUser(&user)
		require.Nil(t, err)

		paymentCycleId, err := insertNewPaymentCycleIn(1, 365, timeNow())
		require.Nil(t, err)

		err = insertContribution(user.Id, user.Id, *repoId, paymentCycleId, payOutId, big.NewInt(1), "USD", timeDayPlusOne(timeNow()), timeNow())
		require.Nil(t, err)

		setupPayoutTestServer(t)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/request-payout/USD", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"targetCurrency": "USD",
		}
		request = mux.SetURLVars(request, vars)

		requestPayout(response, request, &user)

		assert.Equal(t, 422, response.Code)

		contributions, err := findContributions(user.Id, true)
		require.Nil(t, err)

		for _, contribution := range contributions {
			assert.Equal(t, sql.NullTime{}, contribution.ClaimedAt)
		}
	})

	t.Run("should get signature and create payout request", func(t *testing.T) {
		setup()
		defer teardown()

		payOutId := uuid.New()

		repoId, err := setupRepo("github.com/hello-world")
		require.Nil(t, err)

		user := User{
			Id:                uuid.New(),
			StripeId:          stringPointer("strip-id"),
			PaymentCycleOutId: payOutId,
			Email:             "email",
		}
		err = insertUser(&user)
		require.Nil(t, err)

		paymentCycleId, err := insertNewPaymentCycleIn(1, 365, timeNow())
		require.Nil(t, err)

		err = insertContribution(user.Id, user.Id, *repoId, paymentCycleId, payOutId, big.NewInt(3370000), "USD", timeDayPlusOne(timeNow()), timeNow())
		require.Nil(t, err)

		setupPayoutTestServer(t)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/request-payout/USD", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"targetCurrency": "USD",
		}
		request = mux.SetURLVars(request, vars)

		requestPayout(response, request, &user)

		assert.Equal(t, 200, response.Code)

		contributions, err := findContributions(user.Id, true)
		require.Nil(t, err)

		for _, contribution := range contributions {
			assert.NotNil(t, contribution.ClaimedAt)
		}
	})
}
