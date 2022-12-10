package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestEmail(t *testing.T) {
	setup()
	defer teardown()
	emailNotifications = 0

	opts.EmailMarketing = "test"
	m := map[string]*big.Int{"ETH": big.NewInt(1)}
	err := sendMarketingEmail("tom@tomp2p.net", m, []string{"tomp2p"})
	assert.Nil(t, err)
	assert.Equal(t, 1, emailNotifications)
	assert.Equal(t, "test", lastMailTo)
}

func TestEmailTwiceSendOne(t *testing.T) {
	setup()
	defer teardown()
	emailNotifications = 0

	opts.EmailMarketing = "test"
	m := map[string]*big.Int{"ETH": big.NewInt(1)}
	err := sendMarketingEmail("tom@tomp2p.net", m, []string{"tomp2p"})
	assert.Nil(t, err)
	err = sendMarketingEmail("tom@tomp2p.net", m, []string{"tomp2p"})
	assert.Nil(t, err)
	assert.Equal(t, 1, emailNotifications)
	assert.Equal(t, "test", lastMailTo)
}

func TestEmailTwiceSendTwo(t *testing.T) {
	setup()
	defer teardown()
	emailNotifications = 0

	opts.EmailMarketing = "live"
	m := map[string]*big.Int{"ETH": big.NewInt(1)}
	err := sendMarketingEmail("tom@tomp2p.net", m, []string{"tomp2p"})
	assert.Nil(t, err)
	debug = true
	secondsAdd = WaitToSendEmail + 1
	err = sendMarketingEmail("tom@tomp2p.net", m, []string{"tomp2p"})
	assert.Nil(t, err)
	assert.Equal(t, 2, emailNotifications)
	assert.Equal(t, "tom@tomp2p.net", lastMailTo)
}

func TestTopUpOneEmail(t *testing.T) {
	setup()
	defer teardown()
	emailNotifications = 0

	u := User{
		Id:                uuid.UUID{},
		InvitedId:         nil,
		StripeId:          nil,
		PaymentCycleInId:  &uuid.UUID{},
		PaymentCycleOutId: uuid.UUID{},
		Email:             "tom2@tomp2p.net",
		Name:              nil,
		Image:             nil,
		PaymentMethod:     nil,
		Last4:             nil,
		CreatedAt:         time.Time{},
		Claims:            nil,
		Role:              nil,
	}
	err := insertUser(&u)
	assert.Nil(t, err)

	err = sendTopUpInvited(u)
	assert.Nil(t, err)
	assert.Equal(t, 1, emailNotifications)
	assert.Equal(t, "tom2@tomp2p.net", lastMailTo)
	err = sendTopUpInvited(u)
	assert.Nil(t, err)
	assert.Equal(t, 1, emailNotifications)
}

func TestTopUpTwoEmail(t *testing.T) {
	setup()
	defer teardown()
	emailNotifications = 0

	u := User{
		Id:                uuid.UUID{},
		InvitedId:         nil,
		StripeId:          nil,
		PaymentCycleInId:  &uuid.UUID{},
		PaymentCycleOutId: uuid.UUID{},
		Email:             "tom2@tomp2p.net",
		Name:              nil,
		Image:             nil,
		PaymentMethod:     nil,
		Last4:             nil,
		CreatedAt:         time.Time{},
		Claims:            nil,
		Role:              nil,
	}
	err := insertUser(&u)
	assert.Nil(t, err)

	err = sendTopUpInvited(u)
	assert.Nil(t, err)
	assert.Equal(t, 1, emailNotifications)
	assert.Equal(t, "tom2@tomp2p.net", lastMailTo)
	u.PaymentCycleInId = &uuid.UUID{1}
	err = sendTopUpInvited(u)
	assert.Nil(t, err)
	assert.Equal(t, 2, emailNotifications)
}
