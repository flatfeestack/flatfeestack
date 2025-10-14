package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func insertPayInEvent(t *testing.T, externalId uuid.UUID, userId uuid.UUID, status string, currency string) *PayInEvent {
	ub := PayInEvent{
		Id:         uuid.New(),
		ExternalId: externalId,
		UserId:     userId,
		Balance:    big.NewInt(1),
		Status:     status,
		Currency:   currency,
		CreatedAt:  time.Time{},
	}
	err := InsertPayInEvent(ub)
	assert.Nil(t, err)
	return &ub
}

func TestPayment(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	e := uuid.New()

	ub := insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	ubs, err := FindPayInUser(u.Id)
	assert.Nil(t, err)

	assert.Equal(t, ub.Balance, ubs[0].Balance)
	assert.Equal(t, 1, len(ubs))
}

func TestPaymentTwiceFailed(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	e := uuid.New()

	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	ub2 := PayInEvent{
		Id:         uuid.New(),
		ExternalId: e,
		UserId:     uuid.New(),
		Balance:    big.NewInt(1),
		Status:     "TEST",
		Currency:   "currency",
		CreatedAt:  time.Time{},
	}
	err := InsertPayInEvent(ub2)

	assert.NotNil(t, err)
}

func TestTwoPaymentTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")

	e := uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	e = uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	ubs, err := FindPayInUser(u.Id)
	assert.Nil(t, err)

	assert.Equal(t, big.NewInt(1), ubs[0].Balance)
	assert.Equal(t, 2, len(ubs))
}

func TestFindSumUserBalanceByCurrency(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")

	e := uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	e = uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XETH")

	e = uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	e = uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XETH")

	e = uuid.New()
	insertPayInEvent(t, e, u.Id, "TEST", "XBTC")

	m, err := FindSumPaymentByCurrency(u.Id, "TEST")
	assert.Nil(t, err)
	assert.Equal(t, m["XBTC"], big.NewInt(3))
	assert.Equal(t, m["XETH"], big.NewInt(2))
}

func TestLatestCurrency(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	u := uuid.New()
	ub := PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     u,
		Balance:    big.NewInt(1),
		Status:     PayInRequest,
		Currency:   "XBTC",
		Seats:      2,
		Freq:       3,
		CreatedAt:  time.Time{},
	}
	err := InsertPayInEvent(ub)
	assert.Nil(t, err)

	ub = PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     u,
		Balance:    big.NewInt(2),
		Status:     PayInRequest,
		Currency:   "XETH",
		Seats:      2,
		Freq:       3,
		CreatedAt:  time.Time{},
	}

	ub = PayInEvent{
		Id:         uuid.New(),
		ExternalId: uuid.New(),
		UserId:     u,
		Balance:    big.NewInt(12),
		Status:     PayInRequest,
		Currency:   "XETH",
		Seats:      2,
		Freq:       3,
		CreatedAt:  time.Time{}.Add(1),
	}
	err = InsertPayInEvent(ub)
	assert.Nil(t, err)

	b, _, _, c, err := FindLatestDailyPayment(u, "XETH")
	assert.Nil(t, err)
	assert.Nil(t, b)
	err = PaymentSuccess(ub.ExternalId, big.NewInt(0))
	assert.Nil(t, err)
	b, _, _, c, err = FindLatestDailyPayment(u, "XETH")
	assert.Nil(t, err)
	//why 7? we have a payin of 42, with 2 seats -> 42/2=21 and 3 days -> 21/3=7,
	//so we pay daily 7 XETH. We will round down, so 41 would lead to daily payout of 6.
	assert.Equal(t, big.NewInt(2), b)
	c1 := time.Time{}.Add(1)
	assert.Equal(t, &c1, c)
}
