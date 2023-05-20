package api

import (
	"backend/db"
	"backend/utils"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	file, err := os.CreateTemp("", "sqlite")
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Errorf("cannot remove file: %v", err)
		}
	}(file.Name())

	err = dbLib.InitDb("sqlite3", file.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	os.Exit(code)
}

func setup() {
	err := dbLib.RunSQL("../db/init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}
func teardown() {
	err := dbLib.RunSQL("../db/delAll_test.sql")
	if err != nil {
		log.Fatalf("Could not run delAll_test.sql: %s", err)
	}
}

func insertTestUser(t *testing.T, email string) *db.UserDetail {
	u := db.User{
		Id:    uuid.New(),
		Email: email,
	}
	ud := db.UserDetail{
		User:     u,
		StripeId: utils.StringPointer("strip-id"),
	}

	err := db.InsertUser(&ud)
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
	setup()
	defer teardown()

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
