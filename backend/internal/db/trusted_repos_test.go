package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
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
		TrustAt:   &t1,
		UnTrustAt: &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t2,
		UnTrustAt: &t2,
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
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t2,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t3,
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
		UnTrustAt: &t1,
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
		TrustAt:   &t2,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t1,
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
		TrustAt:   &t2,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t4,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t3,
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
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r.Id,
		EventType: Inactive,
		UnTrustAt: &t2,
	}

	tr3 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t3,
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
		UnTrustAt: &t4,
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
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       u.Id,
		RepoId:    r2.Id,
		EventType: Active,
		TrustAt:   &t2,
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
//		SponsorAt: &t1,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t2,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t3, t4)
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
//		SponsorAt: &t1,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t2,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t2, t4)
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
//		SponsorAt: &t1,
//	}
//
//	s2 := SponsorEvent{
//		Id:        uuid.New(),
//		Uid:       u.Id,
//		RepoId:    r2.Id,
//		EventType: Active,
//		SponsorAt: &t2,
//	}
//
//	s3 := SponsorEvent{
//		Id:          uuid.New(),
//		Uid:         u.Id,
//		RepoId:      r2.Id,
//		EventType:   Inactive,
//		UnSponsorAt: &t3,
//	}
//
//	err := InsertOrUpdateSponsor(&s1)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s2)
//	assert.Nil(t, err)
//	err = InsertOrUpdateSponsor(&s3)
//	assert.Nil(t, err)
//
//	res, err := FindSponsorsBetween(t2, t4)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(res[0].RepoIds))
//}
