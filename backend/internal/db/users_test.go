package db

import (
	"testing"

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
