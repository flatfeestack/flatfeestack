package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	t1 = time.Time{}.Add(time.Duration(1) * time.Second)
	t2 = time.Time{}.Add(time.Duration(2) * time.Second)
	t3 = time.Time{}.Add(time.Duration(3) * time.Second)
	t4 = time.Time{}.Add(time.Duration(4) * time.Second)
)

func TestSponsorTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	s1 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   &t1,
		UnSponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   &t2,
		UnSponsorAt: &t2,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	//we want to insertOrUpdateSponsor, but we are already sponsoring this repo
	err = InsertOrUpdateSponsor(&s2)
	assert.NotNil(t, err)

}

func TestUnSponsorTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t3,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	//we want to unsponsor, but we already unsponsored it
	err = InsertOrUpdateSponsor(&s3)
	assert.NotNil(t, err)

}

func TestUnSponsorWrong(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	s1 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t1,
	}

	//we want to unsponsor, but we are currently not sponsoring this repo
	err := InsertOrUpdateSponsor(&s1)
	assert.NotNil(t, err)
}

func TestSponsorWrongOrder(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t1,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	//we want to unsponsor, but the unsponsor date is before the sponsor date
	err = InsertOrUpdateSponsor(&s2)
	assert.NotNil(t, err)

}

func TestSponsorWrongOrderActive(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t4,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t3,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	//we want to unsponsor, but we already unsponsored it at 0001-01-01 00:00:04 +0000 UTC
	err = InsertOrUpdateSponsor(&s3)
	assert.NotNil(t, err)

}

func TestSponsorCorrect(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		UnSponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t3,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s3)
	assert.Nil(t, err)

	rs, err := FindSponsoredReposByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Id, rs[0].Id)

	s4 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r2.Id,
		EventType:   Inactive,
		UnSponsorAt: &t4,
	}
	err = InsertOrUpdateSponsor(&s4)
	assert.Nil(t, err)

	rs, err = FindSponsoredReposByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestTwoRepos(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)

	rs, err := FindSponsoredReposByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))
}

func TestSponsorsBetween(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)

	res, err := FindSponsorsBetween(t3, t4)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res[0].RepoIds))
}

func TestSponsorsBetween2(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)

	res, err := FindSponsorsBetween(t2, t4)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res[0].RepoIds))
}

func TestSponsorsBetween3(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r2.Id,
		EventType:   Inactive,
		UnSponsorAt: &t3,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s3)
	assert.Nil(t, err)

	res, err := FindSponsorsBetween(t2, t4)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res[0].RepoIds))
}

func TestSponsorsBetween4(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		SponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		SponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r2.Id,
		EventType:   Inactive,
		UnSponsorAt: &t4,
	}

	err := InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s3)
	assert.Nil(t, err)

	res, err := FindSponsorsBetween(t3, t4)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res[0].RepoIds))
}
