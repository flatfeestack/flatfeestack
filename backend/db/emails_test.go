package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestGitEmailInsert(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uuid.New(), uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	assert.NotNil(t, err)
}

func TestGitEmailFind(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uuid.New(), uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	emails, err := FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(emails))
}

func TestGitEmailDelete(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	err = InsertGitEmail(uuid.New(), uid, "email2", stringPointer("A"), time.Now())
	err = DeleteGitEmail(uid, "email2")
	assert.Nil(t, err)
	emails, err := FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(emails))
	err = DeleteGitEmail(uid, "email1")
	assert.Nil(t, err)
	emails, err = FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(emails))
}

func TestGitEmailConfirm(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())

	err = ConfirmGitEmail("email1", "AB", time.Time{})
	assert.NotNil(t, err)

	err = ConfirmGitEmail("email1", "A", time.Time{})
	assert.Nil(t, err)
}

func TestInsertUnclaimed(t *testing.T) {
	setup()
	defer teardown()

	r := insertTestRepo(t)

	err := InsertUnclaimed(uuid.New(), "test", r.Id, big.NewInt(1), "XBTC", time.Time{}, time.Time{})
	assert.Nil(t, err)

	err = InsertUnclaimed(uuid.New(), "test", r.Id, big.NewInt(1), "XBTC", time.Time{}.Add(1), time.Time{})
	assert.Nil(t, err)

	m, err := FindMarketingEmails()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(m))
	assert.Equal(t, big.NewInt(2), m[0].Balances["XBTC"])
}
