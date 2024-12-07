package db

import (
	"backend/pkg/util"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func insertTestUser(t *testing.T, email string) *UserDetail {
	u := User{
		Id:    uuid.New(),
		Email: email,
	}
	ud := UserDetail{
		User:     u,
		StripeId: stringPointer("strip-id"),
	}

	err := InsertUser(&ud)
	assert.Nil(t, err)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	return u2
}

func insertTestFoundation(t *testing.T, email string, multiplierDailyLimit int) *UserDetail {
	u := User{
		Id:    uuid.New(),
		Email: email,
	}
	ud := UserDetail{
		User:                 u,
		StripeId:             stringPointer("strip-id"),
		Multiplier:           true,
		MultiplierDailyLimit: multiplierDailyLimit,
	}

	err := InsertFoundation(&ud)
	assert.Nil(t, err)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	return u2
}

func TestUserNotFound(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	insertTestUser(t, "email")

	u2, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)
}

func TestUserFound(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	insertTestUser(t, "email")

	u3, err := FindUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)
}

func TestUserUpdate(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")

	u.Email = "email2"
	u.StripeId = stringPointer("strip-id2")
	err := UpdateStripe(u)
	assert.Nil(t, err)

	//email should not change, only the strip id
	u4, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u4)

	u5, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.NotNil(t, u5)
	assert.Equal(t, *u5.StripeId, "strip-id2")
}

func TestUserUpdateSeat(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	UpdateSeatsFreq(u.Id, 12, 13)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, u2.Seats, 12)
}

func TestUserUpdateInvite(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	i := uuid.New()
	UpdateUserInviteId(u.Id, i)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, u2.InvitedId, &i)
}

func TestUserMultiplierSet(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	UpdateMultiplier(u.Id, true)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, true, u2.Multiplier)
}

func TestUserMultiplierUnset(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	UpdateMultiplier(u.Id, true)
	UpdateMultiplier(u.Id, false)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, false, u2.Multiplier)
}

func TestUserMultiplierDailyLimit(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	UpdateMultiplierDailyLimit(u.Id, 1200)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1200, u2.MultiplierDailyLimit)
}

func TestUserMultiplierDailyLimitChange(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	UpdateMultiplierDailyLimit(u.Id, 1200)
	UpdateMultiplierDailyLimit(u.Id, 600)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 600, u2.MultiplierDailyLimit)
}

func TestCheckDailyLimitStillAdheredToNil(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	checked, err := CheckDailyLimitStillAdheredTo(nil, big.NewInt(0), yesterdayStart)
	assert.Equal(t, fmt.Errorf("foundation cannot be nil"), err)
	assert.Equal(t, false, checked)
}

func TestCheckDailyLimitStillAdheredToNoRowsLessAmount(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(190), yesterdayStart)
	assert.Nil(t, err)
	assert.Equal(t, true, checked)
}

func TestCheckDailyLimitStillAdheredToNoRowsMoreAmount(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(210), yesterdayStart)
	assert.Nil(t, err)
	assert.Equal(t, false, checked)
}

func TestCheckDailyLimitStillAdheredToNoContributions(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)
	theLastDayStart := yesterdayStop.AddDate(0, 0, -2)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	contributor := insertTestUser(t, "email")
	contributor2 := insertTestUser(t, "email2")
	contributor3 := insertTestUser(t, "email3")
	contributor4 := insertTestUser(t, "email4")

	r := insertTestRepoGitUrl(t, "git-url")

	err := InsertContribution(userFoundation.Id, contributor.Id, r.Id, big.NewInt(2), "XBTC", yesterdayStart, time.Time{}, false)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor2.Id, r.Id, big.NewInt(6), "XBTC", yesterdayStart, time.Time{}, false)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor3.Id, r.Id, big.NewInt(100), "XBTC", yesterdayStart, time.Time{}, false)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor4.Id, r.Id, big.NewInt(1000), "XBTC", theLastDayStart, time.Time{}, false)
	assert.Nil(t, err)

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(20), yesterdayStart)
	assert.Nil(t, err)

	assert.Equal(t, true, checked)
}

func TestCheckDailyLimitStillAdheredToInRange(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)
	theLastDayStart := yesterdayStop.AddDate(0, 0, -2)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	contributor := insertTestUser(t, "email")
	contributor2 := insertTestUser(t, "email2")
	contributor3 := insertTestUser(t, "email3")
	contributor4 := insertTestUser(t, "email4")

	r := insertTestRepoGitUrl(t, "git-url")

	err := InsertContribution(userFoundation.Id, contributor.Id, r.Id, big.NewInt(2), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor2.Id, r.Id, big.NewInt(6), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor3.Id, r.Id, big.NewInt(100), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor4.Id, r.Id, big.NewInt(1000), "XBTC", theLastDayStart, time.Time{}, true)
	assert.Nil(t, err)

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(20), yesterdayStart)
	assert.Nil(t, err)

	assert.Equal(t, true, checked)
}

func TestCheckDailyLimitStillAdheredToInRangeOuterLimit(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)
	theLastDayStart := yesterdayStop.AddDate(0, 0, -2)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	contributor := insertTestUser(t, "email")
	contributor2 := insertTestUser(t, "email2")
	contributor3 := insertTestUser(t, "email3")
	contributor4 := insertTestUser(t, "email4")

	r := insertTestRepoGitUrl(t, "git-url")

	err := InsertContribution(userFoundation.Id, contributor.Id, r.Id, big.NewInt(29), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor2.Id, r.Id, big.NewInt(50), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor3.Id, r.Id, big.NewInt(100), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor4.Id, r.Id, big.NewInt(1000), "XBTC", theLastDayStart, time.Time{}, true)
	assert.Nil(t, err)

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(20), yesterdayStart)
	assert.Nil(t, err)

	assert.Equal(t, true, checked)
}

func TestCheckDailyLimitStillAdheredToInRangeJustNotInRange(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	now := util.TimeNow()
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)
	theLastDayStart := yesterdayStop.AddDate(0, 0, -2)

	userFoundation := insertTestFoundation(t, "email5", 200)

	foundation := Foundation{
		Id:                   userFoundation.Id,
		MultiplierDailyLimit: userFoundation.MultiplierDailyLimit,
	}

	contributor := insertTestUser(t, "email")
	contributor2 := insertTestUser(t, "email2")
	contributor3 := insertTestUser(t, "email3")
	contributor4 := insertTestUser(t, "email4")

	r := insertTestRepoGitUrl(t, "git-url")

	err := InsertContribution(userFoundation.Id, contributor.Id, r.Id, big.NewInt(30), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor2.Id, r.Id, big.NewInt(50), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor3.Id, r.Id, big.NewInt(100), "XBTC", yesterdayStart, time.Time{}, true)
	assert.Nil(t, err)

	err = InsertContribution(userFoundation.Id, contributor4.Id, r.Id, big.NewInt(1000), "XBTC", theLastDayStart, time.Time{}, true)
	assert.Nil(t, err)

	checked, err := CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(21), yesterdayStart)
	assert.Nil(t, err)

	assert.Equal(t, false, checked)
}
