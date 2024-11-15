package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	t001 = time.Time{}.Add(time.Duration(1) * time.Second)
	t002 = time.Time{}.Add(time.Duration(2) * time.Second)
	t003 = time.Time{}.Add(time.Duration(3) * time.Second)
	t004 = time.Time{}.Add(time.Duration(4) * time.Second)
)

func TestSetMultiplierRepoTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Active,
		MultiplierAt:   &t001,
		UnMultiplierAt: &t001,
	}

	tr2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Active,
		MultiplierAt:   &t002,
		UnMultiplierAt: &t002,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	//we want to InsertOrUpdateMultiplierRepo, but we are already trusting this repo
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.NotNil(t, err)

}

func TestUnsetMultiplierTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	tr2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t002,
	}

	tr3 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.Nil(t, err)
	//we want to untrust, but we already untrusted it
	err = InsertOrUpdateMultiplierRepo(&tr3)
	assert.NotNil(t, err)

}

func TestUnsetMultiplierWrong(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t001,
	}

	//we want to untrust, but we are currently not trusting this repo
	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.NotNil(t, err)
}

func TestMultiplierWrongOrder(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	tr2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t001,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	//we want to untrunst, but the untrust date is before the trust date
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.NotNil(t, err)

}

func TestMultiplierWrongOrderActive(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	tr2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t004,
	}

	tr3 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.Nil(t, err)
	//we want to untrust, but we already untrusted it at 0001-01-01 00:00:04 +0000 UTC
	err = InsertOrUpdateMultiplierRepo(&tr3)
	assert.NotNil(t, err)

}

func TestMultiplierCorrect(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	tr2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t002,
	}

	tr3 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&tr3)
	assert.Nil(t, err)

	rs, err := FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Id, rs[0].Id)

	tr4 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r2.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t004,
	}
	err = InsertOrUpdateMultiplierRepo(&tr4)
	assert.Nil(t, err)

	rs, err = FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestTwoMultipliedRepos(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	tr2 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	err := InsertOrUpdateMultiplierRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&tr2)
	assert.Nil(t, err)

	rs, err := FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))
}

//func TestSponsorsBetween(t *testing.T) {
//	SetupTestData()
//	defer TeardownTestData()
//	u := insertTestUser(t, "email")
//	r := insertTestRepoGitUrl(t, "git-url")
//	r2 := insertTestRepoGitUrl(t, "git-url2")
//
//	s1 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r.Id,
//		EventType: Active,
//		SponsorAt: &t001,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t002,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t003, t004)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(res[0].RepoIds))
//}
//
//func TestSponsorsBetween2(t *testing.T) {
//	SetupTestData()
//	defer TeardownTestData()
//	u := insertTestUser(t, "email")
//	r := insertTestRepoGitUrl(t, "git-url")
//	r2 := insertTestRepoGitUrl(t, "git-url2")
//
//	s1 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r.Id,
//		EventType: Active,
//		SponsorAt: &t001,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t002,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t002, t004)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(res[0].RepoIds))
//}
//
//func TestSponsorsBetween3(t *testing.T) {
//	SetupTestData()
//	defer TeardownTestData()
//	u := insertTestUser(t, "email")
//	r := insertTestRepoGitUrl(t, "git-url")
//	r2 := insertTestRepoGitUrl(t, "git-url2")
//
//	s1 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r.Id,
//		EventType: Active,
//		SponsorAt: &t001,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t002,
//	}
//
//	s3 := SponsorEvent{
//		Id:          uuid.New(),
//		Uid:         u.Id,
//		RepoId:      r2.Id,
//		EventType:   Inactive,
//		UnSponsorAt: &t003,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s3)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t002, t004)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(res[0].RepoIds))
//}
