package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func setup() {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	err = initDB(db, "../init.sql")
	if err != nil {
		log.Fatal(err)
	}
}
func teardown() {
	db.Close()
}

func TestUser(t *testing.T) {
	setup()
	defer teardown()

	u := User{
		Id:                uuid.New(),
		StripeId:          create("strip-id"),
		Email:             create("email"),
		Subscription:      create("sub"),
		SubscriptionState: create("state"),
		PayoutETH:         create("0x123"),
	}

	err := SaveUser(&u)
	assert.Nil(t, err)

	u2, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)

	u3, err := FindUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)

	u.Email = create("email2")
	err = UpdateUser(&u)
	assert.Nil(t, err)

	u4, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.NotNil(t, u4)

	u5, err := FindUserByID(u.Id)
	assert.Nil(t, err)
	assert.NotNil(t, u5)
}

func TestSponsor(t *testing.T) {
	setup()
	defer teardown()

	u := User{
		Id:                uuid.New(),
		StripeId:          create("strip-id"),
		Email:             create("email"),
		Subscription:      create("sub"),
		SubscriptionState: create("state"),
		PayoutETH:         create("0x123"),
	}

	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		OrigFrom:    create("github"),
		Url:         create("url"),
		Name:        create("name"),
		Description: create("desc"),
	}
	err := SaveUser(&u)
	assert.Nil(t, err)
	err = SaveRepo(&r)
	assert.Nil(t, err)

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: SPONSOR,
		CreatedAt: time.Unix(1, 0),
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: UNSPONSOR,
		CreatedAt: time.Unix(2, 0),
	}

	s3 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: SPONSOR,
		CreatedAt: time.Unix(3, 0),
	}

	err = Sponsor(&s1)
	assert.Nil(t, err)
	err = Sponsor(&s2)
	assert.Nil(t, err)
	err = Sponsor(&s3)
	assert.Nil(t, err)

	rs, err := GetSponsoredReposById(u.Id, SPONSOR)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))

	s4 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: UNSPONSOR,
		CreatedAt: time.Unix(4, 0),
	}
	err = Sponsor(&s4)
	assert.Nil(t, err)

	rs, err = GetSponsoredReposById(u.Id, SPONSOR)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestRepo(t *testing.T) {
	setup()
	defer teardown()
	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		OrigFrom:    create("github"),
		Url:         create("url"),
		Name:        create("name"),
		Description: create("desc"),
	}
	err := SaveRepo(&r)
	assert.Nil(t, err)

	r2, err := FindRepoByID(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, r2)

	r3, err := FindRepoByID(r.Id)
	assert.Nil(t, err)
	assert.NotNil(t, r3)
}

func TestGitEmail(t *testing.T) {
	setup()
	defer teardown()
	uid := uuid.New()
	err := SaveGitEmail(uuid.New(), uid, "email1")
	assert.Nil(t, err)
	err = SaveGitEmail(uuid.New(), uid, "email2")
	assert.Nil(t, err)
	emails, err := FindGitEmails(uid)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(emails))
	err = DeleteGitEmail(uid, "email2")
	assert.Nil(t, err)
	emails, err = FindGitEmails(uid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(emails))
	err = DeleteGitEmail(uid, "email1")
	assert.Nil(t, err)
	emails, err = FindGitEmails(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(emails))
}

func create(s string) *string {
	return &s
}

func initDB(db *sql.DB, file string) error {
	//this will create or alter tables
	//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(file); err == nil {
		file, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		requests := strings.Split(string(file), ";")
		for _, request := range requests {
			request = strings.Replace(request, "\n", "", -1)
			request = strings.Replace(request, "\t", "", -1)
			if !strings.HasPrefix(request, "#") {
				_, err = db.Exec(request)
				if err != nil {
					return fmt.Errorf("[%v] %v", request, err)
				}
			}
		}
	}

	return nil
}
