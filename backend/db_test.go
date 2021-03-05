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

	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		Email:             stringPointer("email"),
		Subscription:      stringPointer("sub"),
		SubscriptionState: stringPointer("state"),
		PayoutETH:         stringPointer("0x123"),
	}

	err := insertUser(&u, "A")
	assert.Nil(t, err)

	u2, err := findUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)

	u3, err := findUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)

	u.Email = stringPointer("email2")
	err = updateUser(&u)
	assert.Nil(t, err)

	u4, err := findUserByEmail("email2")
	assert.Nil(t, err)
	assert.NotNil(t, u4)

	u5, err := findUserById(u.Id)
	assert.Nil(t, err)
	assert.NotNil(t, u5)
}

func TestSponsor(t *testing.T) {
	setup()
	defer teardown()

	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		Email:             stringPointer("email"),
		Subscription:      stringPointer("sub"),
		SubscriptionState: stringPointer("state"),
		PayoutETH:         stringPointer("0x123"),
	}

	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer("url"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertUser(&u, "A")
	assert.Nil(t, err)
	id, err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)
	assert.NotNil(t, id)

	s1 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   time.Time{}.Add(time.Duration(1) * time.Second),
		UnsponsorAt: time.Time{}.Add(time.Duration(1) * time.Second),
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   time.Time{}.Add(time.Duration(2) * time.Second),
		UnsponsorAt: time.Time{}.Add(time.Duration(2) * time.Second),
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   time.Time{}.Add(time.Duration(3) * time.Second),
		UnsponsorAt: time.Time{}.Add(time.Duration(3) * time.Second),
	}

	err1, err2 := insertOrUpdateSponsor(&s1)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	err1, err2 = insertOrUpdateSponsor(&s2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	err1, err2 = insertOrUpdateSponsor(&s3)
	assert.Nil(t, err1)
	assert.Nil(t, err2)

	rs, err := findSponsoredReposById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))

	s4 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   time.Unix(4, 0),
		UnsponsorAt: time.Unix(4, 0),
	}
	err1, err2 = insertOrUpdateSponsor(&s4)
	assert.Nil(t, err1)
	assert.Nil(t, err2)

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
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	id, err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)
	assert.NotNil(t, id)

	r2, err := findRepoById(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, r2)

	r3, err := findRepoById(r.Id)
	assert.Nil(t, err)
	assert.NotNil(t, r3)
}

func saveTestUser(t *testing.T, email string) uuid.UUID {
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		Email:             stringPointer(email),
		Subscription:      stringPointer("sub"),
		SubscriptionState: stringPointer("state"),
		PayoutETH:         stringPointer("0x123"),
	}

	err := insertUser(&u, "A")
	assert.Nil(t, err)
	return u.Id
}

func TestGitEmail(t *testing.T) {
	setup()
	defer teardown()

	uid := saveTestUser(t, "email1")

	err := insertGitEmail(uuid.New(), uid, "email1", "A", timeNow())
	assert.Nil(t, err)
	err = insertGitEmail(uuid.New(), uid, "email2", "A", timeNow())
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
