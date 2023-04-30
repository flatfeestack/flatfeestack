package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestUserNotFound(t *testing.T) {
	setup()
	defer teardown()
	insertTestUser(t, "email")

	u2, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)
}

func TestUserFound(t *testing.T) {
	setup()
	defer teardown()
	insertTestUser(t, "email")

	u3, err := FindUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)
}

func TestUserUpdate(t *testing.T) {
	setup()
	defer teardown()
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
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	UpdateSeatsFreq(u.Id, 12, 13)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, u2.Seats, 12)
}

func TestUserUpdateInvite(t *testing.T) {
	setup()
	defer teardown()
	u := insertTestUser(t, "email")
	i := uuid.New()
	UpdateUserInviteId(u.Id, i)
	u2, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, u2.InvitedId, &i)
}
