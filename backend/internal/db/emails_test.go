package db

import (
	"backend/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestGitEmailInsert(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uuid.New(), uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())
	assert.NotNil(t, err)
}

func TestCountExistingOrConfirmedGitEmail(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	uid := insertTestUser(t, "foo@bar.ch").Id
	uid2 := insertTestUser(t, "bar@foo.ch").Id
	uid3 := insertTestUser(t, "meh@foo.ch").Id

	c, err := CountExistingOrConfirmedGitEmail(uid, "foo@bar.ch")
	assert.Equal(t, 0, c)
	err = InsertGitEmail(uuid.New(), uid, "foo@bar.ch", stringPointer("A"), time.Now())
	assert.Nil(t, err)

	c, err = CountExistingOrConfirmedGitEmail(uid2, "foo@bar.ch")
	assert.Equal(t, 0, c)
	err = InsertGitEmail(uuid.New(), uid2, "foo@bar.ch", stringPointer("B"), time.Now())
	assert.Nil(t, err)

	c, err = CountExistingOrConfirmedGitEmail(uid, "foo@bar.ch")
	assert.Equal(t, 1, c)

	err = ConfirmGitEmail("foo@bar.ch", "A", time.Time{})
	assert.Nil(t, err)

	c, err = CountExistingOrConfirmedGitEmail(uid3, "foo@bar.ch")
	assert.Equal(t, 1, c)
}

func TestGitEmailFind(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

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
	util.SetupTestData()
	defer util.TeardownTestData()

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

func TestGitEmailFromUserEmailsSentDelete(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	uid := insertTestUser(t, "foo@bar.ch").Id

	err := InsertEmailSent(uuid.New(), &uid, "foo@bar.ch", "add-gitfoo@bar.ch", time.Now())
	err = InsertEmailSent(uuid.New(), &uid, "bar@foo.ch", "add-gitbar@foo.ch", time.Now())
	err = DeleteGitEmailFromUserEmailsSent(uid, "bar@foo.ch")
	assert.Nil(t, err)
	c, err := CountEmailSentById(uid, "add-gitfoo@bar.ch")
	assert.Nil(t, err)
	assert.Equal(t, 1, c)
	c, err = CountEmailSentById(uid, "add-gitbar@foo.ch")
	assert.Nil(t, err)
	assert.Equal(t, 0, c)
}

func TestGitEmailConfirm(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uuid.New(), uid, "email1", stringPointer("A"), time.Now())

	err = ConfirmGitEmail("email1", "AB", time.Time{})
	assert.NotNil(t, err)

	err = ConfirmGitEmail("email1", "A", time.Time{})
	assert.Nil(t, err)
}

func TestInsertUnclaimed(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

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
