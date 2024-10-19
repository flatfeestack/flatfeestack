package api

import (
	"backend/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUserBalance(t *testing.T) {
	t.Run("user made no payments", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[]\n", string(body))
	})

	t.Run("user made a pay-in but did not distribute anything", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 12, 2, 2)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[{\"currency\":\"USD\",\"balance\":12}]\n", string(body))
	})

	t.Run("user made a pay-in and some got distributed to contributors", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 12, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertContribution(userDetail.Id, userDetail.Id, *repoId, big.NewInt(3370000), "USD", time.Now(), time.Now())
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[{\"currency\":\"USD\",\"balance\":-3369988}]\n", string(body))
	})

	t.Run("user made a pay-in and has future contribution", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 12, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(userDetail.Id, *repoId, big.NewInt(12), "USD", time.Now(), time.Now())
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[{\"currency\":\"USD\",\"balance\":0}]\n", string(body))
	})

	t.Run("user made a pay-in, has future contribution and distributed funds", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 400, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(userDetail.Id, *repoId, big.NewInt(100), "USD", time.Now(), time.Now())
		require.Nil(t, err)

		err = db.InsertContribution(userDetail.Id, userDetail.Id, *repoId, big.NewInt(200), "USD", time.Now(), time.Now())
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[{\"currency\":\"USD\",\"balance\":100}]\n", string(body))
	})
}
