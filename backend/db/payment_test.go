package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func insertUserBalance(t *testing.T, userId uuid.UUID, paymentCycleIn uuid.UUID, bType string, currency string) *UserBalance {
	ub := UserBalance{
		UserId:           userId,
		Balance:          big.NewInt(1),
		DailySpending:    big.NewInt(2),
		PaymentCycleInId: paymentCycleIn,
		BalanceType:      bType,
		Currency:         currency,
		CreatedAt:        time.Time{},
	}
	err := InsertUserBalance(ub)
	assert.Nil(t, err)
	return &ub
}

func insertPaymentCycle(t *testing.T) uuid.UUID {
	id := uuid.New()
	err := InsertNewPaymentCycleIn(id, 4, 4, time.Time{})
	assert.Nil(t, err)
	return id
}

func TestPayment(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)

	ub := insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	ubs, err := FindUserBalances(u.Id)
	assert.Nil(t, err)

	assert.Equal(t, ub.Balance, ubs[0].Balance)
	assert.Equal(t, 1, len(ubs))
}

func TestPaymentTwice(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)
	pid = insertPaymentCycle(t)
	ub := insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	ubs, err := FindUserBalances(u.Id)
	assert.Nil(t, err)

	assert.Equal(t, ub.Balance, ubs[0].Balance)
	assert.Equal(t, 1, len(ubs))
}

func TestTwoPaymentTwice(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)
	ub := insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	pid = insertPaymentCycle(t)
	ub = insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	ubs, err := FindUserBalances(u.Id)
	assert.Nil(t, err)

	assert.Equal(t, ub.Balance, ubs[0].Balance)
	assert.Equal(t, 2, len(ubs))
}

func TestPaymentFailed(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)

	insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	ub2 := UserBalance{
		UserId:           u.Id,
		Balance:          big.NewInt(1),
		DailySpending:    big.NewInt(2),
		PaymentCycleInId: pid,
		BalanceType:      "TEST",
		Currency:         "XBTC",
		CreatedAt:        time.Time{},
	}
	err := InsertUserBalance(ub2)
	assert.NotNil(t, err)
}

func TestFindPaymentCycle(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)
	//on successful payment, the payment cycle id gets updated
	err := UpdatePaymentCycleInId(u.Id, pid)
	assert.Nil(t, err)
	p, err := FindPaymentCycleLast(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, pid, p.Id)
}

func TestUserBalanceAndType(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)
	insertUserBalance(t, u.Id, pid, "TEST", "XBTC")

	insertUserBalance(t, u.Id, pid, "TEST", "XBTC2")

	insertUserBalance(t, u.Id, pid, "TEST2", "XBTC")

	pid2 := insertPaymentCycle(t)
	insertUserBalance(t, u.Id, pid2, "TEST", "XBTC")

	ub, err := FindBalance(pid, u.Id, "TEST", "XBTC")
	assert.Nil(t, err)
	assert.Equal(t, ub.Balance, big.NewInt(1))
}

func TestFindSumUserBalanceByCurrency(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	pid := insertPaymentCycle(t)
	insertUserBalance(t, u.Id, pid, "TEST1", "XBTC")
	insertUserBalance(t, u.Id, pid, "TEST2", "XBTC")

	insertUserBalance(t, u.Id, pid, "TEST1", "CHF")
	insertUserBalance(t, u.Id, pid, "TEST2", "CHF")

	m, err := FindSumUserBalanceByCurrency(pid)
	assert.Nil(t, err)
	assert.Equal(t, m["XBTC"].Balance, big.NewInt(2))
	assert.Equal(t, m["CHF"].Balance, big.NewInt(2))
}
