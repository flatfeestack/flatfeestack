package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
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
		PaymentCycleOutId: payOutId,
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
		PaymentCycleOutId: payOutId,
		Email:             "email",
	}

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertUser(&u)
	assert.Nil(t, err)
	err = insertOrUpdateRepo(&r)
	assert.Nil(t, err)

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
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)

	r2, err := findRepoById(uuid.New())
	assert.Nil(t, err)
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
		PaymentCycleOutId: payOutId,
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

func TestAnalysisRequest(t *testing.T) {
	setup()
	defer teardown()

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)

	ar, err := findLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Nil(t, ar)

	a := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r.Id,
		DateFrom:  day1,
		DateTo:    day2,
		GitUrl:    *r.GitUrl,
	}
	err = insertAnalysisRequest(a, timeNow())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err := findLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a.RequestId)
	assert.Equal(t, abd.RepoId, a.RepoId)
}

func TestAnalysisRequest2(t *testing.T) {
	setup()
	defer teardown()

	r1 := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertOrUpdateRepo(&r1)
	assert.Nil(t, err)

	r2 := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url2"),
		GitUrl:      stringPointer("gitUrl2"),
		Source:      stringPointer("source2"),
		Name:        stringPointer("name2"),
		Description: stringPointer("desc2"),
	}
	err = insertOrUpdateRepo(&r2)
	assert.Nil(t, err)

	a1 := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r1.Id,
		DateFrom:  day1,
		DateTo:    day2,
		GitUrl:    *r1.GitUrl,
	}
	err = insertAnalysisRequest(a1, timeNow())
	assert.Nil(t, err)

	a2 := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r1.Id,
		DateFrom:  day2,
		DateTo:    day3,
		GitUrl:    *r1.GitUrl,
	}
	err = insertAnalysisRequest(a2, timeNow())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err := findLatestAnalysisRequest(r1.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a2.RequestId)
	assert.Equal(t, abd.RepoId, a2.RepoId)

	a3 := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r2.Id,
		DateFrom:  day3,
		DateTo:    day4,
		GitUrl:    *r2.GitUrl,
	}
	err = insertAnalysisRequest(a3, timeNow())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err = findLatestAnalysisRequest(r1.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a2.RequestId)
	assert.Equal(t, abd.RepoId, a2.RepoId)

	a4 := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r2.Id,
		DateFrom:  day4,
		DateTo:    day5,
		GitUrl:    *r2.GitUrl,
	}
	err = insertAnalysisRequest(a4, timeNow())
	assert.Nil(t, err)

	alar, err := findAllLatestAnalysisRequest(day2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(alar))
	assert.Equal(t, alar[0].RepoId, r1.Id)
	assert.Equal(t, alar[1].RepoId, r2.Id)

	alar, err = findAllLatestAnalysisRequest(day3)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(alar))
	assert.Equal(t, alar[0].RepoId, r1.Id)
	assert.Equal(t, alar[1].RepoId, r2.Id)
}

func TestAnalysisResponse(t *testing.T) {
	setup()
	defer teardown()

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := insertOrUpdateRepo(&r)
	assert.Nil(t, err)

	a := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    r.Id,
		DateFrom:  day1,
		DateTo:    day2,
		GitUrl:    *r.GitUrl,
	}
	err = insertAnalysisRequest(a, timeNow())
	assert.Nil(t, err)

	err = insertAnalysisResponse(a.RequestId, "tom", []string{"tom"}, 0.5, timeNow())
	assert.Nil(t, err)
	err = insertAnalysisResponse(a.RequestId, "tom", []string{"tom"}, 0.4, timeNow())
	assert.NotNil(t, err)
	err = insertAnalysisResponse(a.RequestId, "tom2", []string{"tom2"}, 0.4, timeNow())
	assert.Nil(t, err)

	ar, err := findAnalysisResults(a.RequestId)
	assert.Equal(t, 2, len(ar))
	assert.Equal(t, ar[0].GitNames[0], "tom")
	assert.Equal(t, ar[1].GitNames[0], "tom2")

	err = updateAnalysisRequest(a.RequestId, day2, stringPointer("test"))
	assert.Nil(t, err)

	alr, err := findLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, day3.Nanosecond(), alr.ReceivedAt.Nanosecond())
}

func TestSumTotalEarnedAmountForContributionIds(t *testing.T) {
	t.Run("returns 0 without any elements in array", func(t *testing.T) {
		setup()
		defer teardown()

		array := []uuid.UUID{}
		result, err := sumTotalEarnedAmountForContributionIds(array)

		require.Nil(t, err)
		assert.Equal(t, big.NewInt(0), result)
	})

	t.Run("returns sum of received contributions", func(t *testing.T) {
		setup()
		defer teardown()
		payOutId := uuid.New()

		repoId, err := setupRepo("github.com/hello-world")
		require.Nil(t, err)

		user := User{
			Id:                uuid.New(),
			StripeId:          stringPointer("strip-id"),
			PaymentCycleOutId: payOutId,
			Email:             "email",
		}
		err = insertUser(&user)
		require.Nil(t, err)

		paymentCycleId, err := insertNewPaymentCycleIn(1, 365, timeNow())
		require.Nil(t, err)

		err = insertContribution(user.Id, user.Id, *repoId, paymentCycleId, payOutId, big.NewInt(1), "USD", timeDayPlusOne(timeNow()), timeNow())
		require.Nil(t, err)

		ids, err := findOwnContributionIds(user.Id, "USD")
		require.Nil(t, err)
		assert.Equal(t, 1, len(ids))

		result, err := sumTotalEarnedAmountForContributionIds(ids)

		require.Nil(t, err)
		assert.Equal(t, big.NewInt(1), result)
	})
}
