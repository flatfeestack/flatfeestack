package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertPayInEvent(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	payInEvent := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(10000),
		Currency:   "USD",
		Status:     PayInRequest,
		Seats:      5,
		Freq:       30,
		CreatedAt:  time.Now(),
	}

	err := db.InsertPayInEvent(payInEvent)
	require.NoError(t, err)
}

func TestFindPayInUser(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	payInEvent1 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(5000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent1))

	payInEvent2 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(3000),
		Currency:   "EUR",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent2))

	events, err := db.FindPayInUser(user.Id)
	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestFindPayInExternal(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")
	externalId := uuid.New()

	payInEvent := PayInEvent{
		Id:         uuid.New(),
		ExternalId: externalId,
		UserId:     user.Id,
		Balance:    big.NewInt(7000),
		Currency:   "USD",
		Status:     PayInRequest,
		Seats:      3,
		Freq:       7,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent))

	found, err := db.FindPayInExternal(externalId, PayInRequest)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.Id, found.UserId)
	assert.Equal(t, big.NewInt(7000), found.Balance)
	assert.Equal(t, "USD", found.Currency)
}

func TestFindPayInExternal_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindPayInExternal(uuid.New(), PayInRequest)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindSumPaymentByCurrency(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	payInEvent1 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(5000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent1))

	payInEvent2 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(3000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent2))

	payInEvent3 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(2000),
		Currency:   "EUR",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent3))

	balances, err := db.FindSumPaymentByCurrency(user.Id, PayInSuccess)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(8000), balances["USD"])
	assert.Equal(t, big.NewInt(2000), balances["EUR"])
}

func TestFindSumPaymentByCurrency_NoCurrency(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	balances, err := db.FindSumPaymentByCurrency(user.Id, PayInSuccess)
	require.NoError(t, err)
	assert.Empty(t, balances)
}

func TestFindSumPaymentByCurrencyWithDate(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	now := time.Now()
	payInEvent := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(1000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      2,
		Freq:       1,
		CreatedAt:  now,
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent))

	balances, err := db.FindSumPaymentByCurrencyWithDate(user.Id, PayInSuccess)
	require.NoError(t, err)
	require.NotNil(t, balances["USD"])
	assert.Equal(t, big.NewInt(2000), balances["USD"].Balance)
	assert.False(t, balances["USD"].CreatedAt.IsZero())
}

func TestFindSumPaymentByCurrencyFoundationWithDate(t *testing.T) {
	TruncateAll(db, t)

	foundation := createTestUser(t, db, "foundation@example.com")

	payInEvent1 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     foundation.Id,
		Balance:    big.NewInt(5000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      2,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent1))

	payInEvent2 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     foundation.Id,
		Balance:    big.NewInt(3000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       30,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent2))

	balances, err := db.FindSumPaymentByCurrencyFoundationWithDate(foundation.Id, PayInSuccess)
	require.NoError(t, err)
	require.NotNil(t, balances["USD"])
	assert.Equal(t, big.NewInt(10000), balances["USD"].Balance)
}

func TestFindSumPaymentFromFoundation(t *testing.T) {
	TruncateAll(db, t)

	foundation := createTestUser(t, db, "foundation@example.com")

	payInEvent1 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     foundation.Id,
		Balance:    big.NewInt(8000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent1))

	payInEvent2 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     foundation.Id,
		Balance:    big.NewInt(5000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      1,
		Freq:       30,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent2))

	balance, err := db.FindSumPaymentFromFoundation(foundation.Id, PayInSuccess, "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(8000), balance)
}

func TestFindSumPaymentFromFoundation_NoPayments(t *testing.T) {
	TruncateAll(db, t)

	foundation := createTestUser(t, db, "foundation@example.com")

	balance, err := db.FindSumPaymentFromFoundation(foundation.Id, PayInSuccess, "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestFindLatestDailyPayment(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	oldPayment := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(6000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      2,
		Freq:       30,
		CreatedAt:  time.Now().AddDate(0, 0, -5),
	}
	require.NoError(t, db.InsertPayInEvent(oldPayment))

	latestPayment := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     user.Id,
		Balance:    big.NewInt(9000),
		Currency:   "USD",
		Status:     PayInSuccess,
		Seats:      3,
		Freq:       30,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(latestPayment))

	balance, seats, freq, createdAt, err := db.FindLatestDailyPayment(user.Id, "USD")
	require.NoError(t, err)
	require.NotNil(t, createdAt)
	assert.Equal(t, int64(3), seats)
	assert.Equal(t, int64(30), freq)
	assert.Equal(t, big.NewInt(100), balance)
}

func TestFindLatestDailyPayment_NotFound(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")

	balance, seats, freq, createdAt, err := db.FindLatestDailyPayment(user.Id, "JPY")
	require.NoError(t, err)
	assert.Nil(t, balance)
	assert.Equal(t, int64(0), seats)
	assert.Equal(t, int64(0), freq)
	assert.Nil(t, createdAt)
}

func TestPaymentSuccess(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "payment@example.com")
	externalId := uuid.New()

	requestEvent := PayInEvent{
		Id:         uuid.New(),
		ExternalId: externalId,
		UserId:     user.Id,
		Balance:    big.NewInt(10000),
		Currency:   "USD",
		Status:     PayInRequest,
		Seats:      1,
		Freq:       1,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(requestEvent))

	fee := big.NewInt(500)
	err := db.PaymentSuccess(externalId, fee)
	require.NoError(t, err)

	successEvent, err := db.FindPayInExternal(externalId, PayInSuccess)
	require.NoError(t, err)
	require.NotNil(t, successEvent)
	assert.Equal(t, big.NewInt(9500), successEvent.Balance)

	feeEvent, err := db.FindPayInExternal(externalId, PayInFee)
	require.NoError(t, err)
	require.NotNil(t, feeEvent)
	assert.Equal(t, fee, feeEvent.Balance)
}

func TestPaymentSuccess_RequestNotFound(t *testing.T) {
	TruncateAll(db, t)

	err := db.PaymentSuccess(uuid.New(), big.NewInt(100))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment request not found")
}