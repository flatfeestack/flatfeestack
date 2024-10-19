package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func insertInvite(t *testing.T, fromEmail string, toEmail string) uuid.UUID {
	id := uuid.New()
	err := InsertInvite(id, fromEmail, toEmail, time.Time{})
	assert.Nil(t, err)
	return id
}

func TestInviteInsert(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	insertInvite(t, "from", "to")
	insertInvite(t, "from2", "to")
	insertInvite(t, "from", "to2")
	insertInvite(t, "from2", "to2")

	id := uuid.New()
	err := InsertInvite(id, "from2", "to2", time.Time{})
	assert.NotNil(t, err)

}

func TestInviteFindMyAny(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	insertInvite(t, "from", "to")
	insertInvite(t, "from2", "to")
	insertInvite(t, "from", "to2")
	insertInvite(t, "from2", "to2")
	insertInvite(t, "to", "from")

	i, err := FindMyInvitations("from")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(i))

	i, err = FindInvitationsByAnyEmail("to")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(i))
}

func TestDeleteInvite(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	insertInvite(t, "from", "to")
	insertInvite(t, "from2", "to")
	insertInvite(t, "to", "from")

	err := DeleteInvite("from", "to")
	assert.Nil(t, err)

	i, err := FindInvitationsByAnyEmail("to")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(i))
}

func TestConfirmInvite(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	insertInvite(t, "from", "to")

	t1 := time.Time{}.Add(time.Duration(1) * time.Second)
	err := UpdateConfirmInviteAt("from", "to", t1)
	assert.Nil(t, err)

	i, err := FindInvitationsByAnyEmail("to")
	assert.Nil(t, err)
	assert.Equal(t, t1.Unix(), i[0].ConfirmedAt.Unix())
}
