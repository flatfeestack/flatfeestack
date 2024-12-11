package api

import (
	"backend/internal/db"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, "[{\"currency\":\"USD\",\"repoId\":\"00000000-0000-0000-0000-000000000000\",\"repoName\":\"N/A\",\"balance\":0,\"totalbalance\":12}]\n", string(body))
	})

	t.Run("user made a pay-in and some got distributed to contributors", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 12, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertContribution(userDetail.Id, userDetail.Id, *repoId, big.NewInt(3370000), "USD", time.Now(), time.Now(), false)
		require.Nil(t, err)

		request, err := http.NewRequest(http.MethodGet, "/users/me/balance", nil)
		assert.Nil(t, err)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":3370000,\"totalbalance\":-3369988}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})

	t.Run("user made a pay-in and has future contribution", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 12, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(userDetail.Id, *repoId, big.NewInt(12), "USD", time.Now(), time.Now(), false)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":12,\"totalbalance\":0}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})

	t.Run("user made a pay-in, has future contribution and distributed funds", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@flatfeestack.io")
		insertPayInEvent(t, uuid.New(), userDetail.Id, db.PayInSuccess, "USD", 400, 2, 2)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(userDetail.Id, *repoId, big.NewInt(100), "USD", time.Now(), time.Now(), false)
		require.Nil(t, err)

		err = db.InsertContribution(userDetail.Id, userDetail.Id, *repoId, big.NewInt(200), "USD", time.Now(), time.Now(), false)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balance", nil)
		response := httptest.NewRecorder()

		UserBalance(response, request, userDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":300,\"totalbalance\":100}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})
}

func TestGetFoundationBalance(t *testing.T) {
	t.Run("foundation made no payments", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		foundationDetail := insertTestFoundation(t, "hello@flatfeestack.io", 200)
		request, _ := http.NewRequest(http.MethodPost, "/users/me/balanceFoundation", nil)
		response := httptest.NewRecorder()

		FoundationBalance(response, request, foundationDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[]\n", string(body))
	})

	t.Run("foundation made a pay-in but did not distribute anything", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		foundationDetail := insertTestFoundation(t, "hello@flatfeestack.io", 200)
		insertPayInEvent(t, uuid.New(), foundationDetail.Id, db.PayInSuccess, "USD", 20, 1, 1)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balanceFoundation", nil)
		response := httptest.NewRecorder()

		FoundationBalance(response, request, foundationDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)
		assert.Equal(t, "[{\"currency\":\"USD\",\"repoId\":\"00000000-0000-0000-0000-000000000000\",\"repoName\":\"N/A\",\"balance\":0,\"totalbalance\":20}]\n", string(body))
	})

	t.Run("foundation made a pay-in and some got payed out to contributors", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		foundationDetail := insertTestFoundation(t, "hello@flatfeestack.io", 200000000)
		insertPayInEvent(t, uuid.New(), foundationDetail.Id, db.PayInSuccess, "USD", 12, 1, 1)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertContribution(foundationDetail.Id, foundationDetail.Id, *repoId, big.NewInt(3370000), "USD", time.Now(), time.Now(), true)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balanceFoundation", nil)
		response := httptest.NewRecorder()

		FoundationBalance(response, request, foundationDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":3370000,\"totalbalance\":-3369988}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})

	t.Run("foundation made a pay-in and has future contribution", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		foundationDetail := insertTestFoundation(t, "hello@flatfeestack.io", 2000)
		insertPayInEvent(t, uuid.New(), foundationDetail.Id, db.PayInSuccess, "USD", 12, 1, 1)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(foundationDetail.Id, *repoId, big.NewInt(12), "USD", time.Now(), time.Now(), true)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balanceFoundation", nil)
		response := httptest.NewRecorder()

		FoundationBalance(response, request, foundationDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":12,\"totalbalance\":0}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})

	t.Run("foundation made a pay-in, has future contribution and distributed funds", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		foundationDetail := insertTestFoundation(t, "hello@flatfeestack.io", 2000)
		insertPayInEvent(t, uuid.New(), foundationDetail.Id, db.PayInSuccess, "USD", 400, 1, 1)
		repoId, err := db.SetupRepo("github.com/hello-world")
		require.Nil(t, err)

		err = db.InsertFutureContribution(foundationDetail.Id, *repoId, big.NewInt(100), "USD", time.Now(), time.Now(), true)
		require.Nil(t, err)

		err = db.InsertContribution(foundationDetail.Id, foundationDetail.Id, *repoId, big.NewInt(200), "USD", time.Now(), time.Now(), true)
		require.Nil(t, err)

		request, _ := http.NewRequest(http.MethodPost, "/users/me/balanceFoundation", nil)
		response := httptest.NewRecorder()

		FoundationBalance(response, request, foundationDetail)

		assert.Equal(t, 200, response.Code)

		body, _ := io.ReadAll(response.Body)

		expected := fmt.Sprintf("[{\"currency\":\"USD\",\"repoId\":\"%s\",\"repoName\":\"name\",\"balance\":300,\"totalbalance\":100}]\n", repoId)
		assert.Equal(t, expected, string(body))
	})
}
