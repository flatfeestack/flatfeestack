package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	setup()
	defer teardown()

	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: &payOutId,
		Email:             "email",
	}

	err := insertUser(&u)
	assert.Nil(t, err)

	u2, err := findUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)

	u3, err := findUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)

	u.Email = "email2"
	err = updateUser(&u)
	assert.Nil(t, err)

	//cannot change Email
	u4, err := findUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u4)

	u5, err := findUserById(u.Id)
	assert.Nil(t, err)
	assert.NotNil(t, u5)
}

func TestSponsor(t *testing.T) {
	setup()
	defer teardown()

	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: &payOutId,
		Email:             "email",
	}

	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Branch:      stringPointer("branch"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertUser(&u)
	assert.Nil(t, err)
	id, err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)
	assert.NotNil(t, id)

	t1 := time.Time{}.Add(time.Duration(1) * time.Second)
	t2 := time.Time{}.Add(time.Duration(2) * time.Second)
	t3 := time.Time{}.Add(time.Duration(3) * time.Second)
	t4 := time.Time{}.Add(time.Duration(4) * time.Second)

	s1 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   t1,
		UnSponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   t2,
		UnSponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   t3,
		UnSponsorAt: &t3,
	}

	err = insertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = insertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	err = insertOrUpdateSponsor(&s3)
	assert.Nil(t, err)

	rs, err := findSponsoredReposById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))

	s4 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   t4,
		UnSponsorAt: &t4,
	}
	err = insertOrUpdateSponsor(&s4)
	assert.Nil(t, err)

	rs, err = findSponsoredReposById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestRepo(t *testing.T) {
	setup()
	defer teardown()
	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Branch:      stringPointer("branch"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	id, err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)
	assert.NotNil(t, id)

	r2, err := findRepoById(uuid.New())
	assert.NotNil(t, err)
	assert.Nil(t, r2)

	r3, err := findRepoById(r.Id)
	assert.Nil(t, err)
	assert.NotNil(t, r3)
}

func saveTestUser(t *testing.T, email string) uuid.UUID {
	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: &payOutId,
		Email:             email,
	}

	err := insertUser(&u)
	assert.Nil(t, err)
	return u.Id
}

func TestGitEmail(t *testing.T) {
	setup()
	defer teardown()

	uid := saveTestUser(t, "email1")

	err := insertGitEmail(uid, "email1", stringPointer("A"), timeNow())
	assert.Nil(t, err)
	err = insertGitEmail(uid, "email2", stringPointer("A"), timeNow())
	assert.Nil(t, err)
	emails, err := findGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(emails))
	err = deleteGitEmail(uid, "email2")
	assert.Nil(t, err)
	emails, err = findGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(emails))
	err = deleteGitEmail(uid, "email1")
	assert.Nil(t, err)
	emails, err = findGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(emails))
}
