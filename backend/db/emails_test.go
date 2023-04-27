package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGitEmailInsert(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())
	assert.NotNil(t, err)
}

func TestGitEmailFind(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	emails, err := FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(emails))
}

func TestGitEmailDelete(t *testing.T) {
	setup()
	defer teardown()

	uid := insertTestUser(t, "email1").Id

	err := InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())
	err = InsertGitEmail(uid, "email2", stringPointer("A"), time.Now())
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

	err := InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())

	err = ConfirmGitEmail("email1", "AB", time.Time{})
	assert.NotNil(t, err)

	err = ConfirmGitEmail("email1", "A", time.Time{})
	assert.Nil(t, err)

}
