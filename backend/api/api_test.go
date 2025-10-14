package api

import (
	"backend/client"
	"backend/db"
	"backend/util"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testDb := db.NewTestDb()
	code := m.Run()
	testDb.CloseTestDb()
	os.Exit(code)
}

func SetupAnalysisTestServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/analyze":
			var request db.AnalysisRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.Nil(t, err)

			err = json.NewEncoder(w).Encode(client.AnalysisResponse2{RequestId: request.Id})
			require.Nil(t, err)
		default:
			http.NotFound(w, r)
		}
	}))
	return server
}
func insertTestUser(t *testing.T, email string) *db.UserDetail {
	u := db.User{
		Id:    uuid.New(),
		Email: email,
	}
	ud := db.UserDetail{
		User:     u,
		StripeId: util.StringPointer("strip-id"),
	}

	err := db.InsertUser(&ud)
	assert.Nil(t, err)
	u2, err := db.FindUserById(u.Id)
	assert.Nil(t, err)
	return u2
}

func insertTestFoundation(t *testing.T, email string, multiplierDailyLimit int) *db.UserDetail {
	u := db.User{
		Id:    uuid.New(),
		Email: email,
	}
	ud := db.UserDetail{
		User:                 u,
		StripeId:             stringPointer("strip-id"),
		Multiplier:           true,
		MultiplierDailyLimit: multiplierDailyLimit,
	}

	err := db.InsertFoundation(&ud)
	assert.Nil(t, err)
	u2, err := db.FindUserById(u.Id)
	assert.Nil(t, err)
	return u2
}

func insertPayInEvent(t *testing.T, externalId uuid.UUID, userId uuid.UUID, status string, currency string, amount int64, seats int64, freq int64) *db.PayInEvent {
	ub := db.PayInEvent{
		Id:         uuid.New(),
		ExternalId: externalId,
		UserId:     userId,
		Balance:    big.NewInt(amount),
		Status:     status,
		Currency:   currency,
		Seats:      seats,
		Freq:       freq,
		CreatedAt:  time.Time{},
	}
	err := db.InsertPayInEvent(ub)
	assert.Nil(t, err)
	return &ub
}

func TestStrategyDeductMax(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()

	u := insertTestUser(t, "test@test.test")
	e1 := uuid.New()
	insertPayInEvent(t, e1, u.Id, db.PayInSuccess, "USD", 12, 2, 2)
	e2 := uuid.New()
	insertPayInEvent(t, e2, u.Id, db.PayInSuccess, "CHF", 16, 4, 4)

	mAdd, err := db.FindSumPaymentByCurrency(u.Id, db.PayInSuccess)
	assert.Nil(t, err)
	mFut, err := db.FindSumFutureSponsors(u.Id)
	assert.Nil(t, err)
	mSub, err := db.FindSumDailySponsors(u.Id)
	assert.Nil(t, err)

	s, i, b, err := StrategyDeductMax(u.Id, mAdd, mSub, mFut)
	assert.Nil(t, err)

	assert.Equal(t, "CHF", s)
	assert.Equal(t, int64(16), i)
	assert.Equal(t, big.NewInt(1), b)
}
