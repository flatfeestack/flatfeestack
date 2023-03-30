package main

import (
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
