package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	t01 = time.Time{}.Add(time.Duration(1) * time.Second)
	t02 = time.Time{}.Add(time.Duration(2) * time.Second)
	t03 = time.Time{}.Add(time.Duration(3) * time.Second)
	t04 = time.Time{}.Add(time.Duration(4) * time.Second)
)

func TestTrustedRepoTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t01,
		UnTrustAt: &t01,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t02,
		UnTrustAt: &t02,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	//we want to InsertOrUpdateTrustRepo, but we are already trusting this repo
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.NotNil(t, err)

}

func TestUnTrustedTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t01,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t02,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t03,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	//we want to untrust, but we already untrusted it
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.NotNil(t, err)

}

func TestUnTrustWrong(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t01,
	}

	//we want to untrust, but we are currently not trusting this repo
	err := InsertOrUpdateTrustRepo(&tr1)
	assert.NotNil(t, err)
}

func TestTrustWrongOrder(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t02,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t01,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	//we want to untrunst, but the untrust date is before the trust date
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.NotNil(t, err)

}

func TestTrustWrongOrderActive(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t02,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t04,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t03,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	//we want to untrust, but we already untrusted it at 0001-01-01 00:00:04 +0000 UTC
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.NotNil(t, err)

}

func TestTrustCorrect(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t01,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t02,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t03,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr3)
	assert.Nil(t, err)

	rs, err := FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Id, rs[0].Id)

	tr4 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Inactive,
		UnTrustAt: &t04,
	}
	err = InsertOrUpdateTrustRepo(&tr4)
	assert.Nil(t, err)

	rs, err = FindTrustedRepos()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestTwoTrustedRepos(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t01,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t02,
	}

	err := InsertOrUpdateTrustRepo(&tr1)
	assert.Nil(t, err)
	err = InsertOrUpdateTrustRepo(&tr2)
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
//		SponsorAt: &t01,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t02,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t03, t04)
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
//		SponsorAt: &t01,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t02,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t02, t04)
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
//		SponsorAt: &t01,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t02,
//	}
//
//	s3 := SponsorEvent{
//		Id:          uuid.New(),
//		Uid:         u.Id,
//		RepoId:      r2.Id,
//		EventType:   Inactive,
//		UnSponsorAt: &t03,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s3)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t02, t04)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(res[0].RepoIds))
//}
