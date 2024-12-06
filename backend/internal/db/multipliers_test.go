package db

import (
	"sort"
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

	m1 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Active,
		MultiplierAt:   &t001,
		UnMultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Active,
		MultiplierAt:   &t002,
		UnMultiplierAt: &t002,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	//we want to InsertOrUpdateMultiplierRepo, but we are already set a multiplier for this repo
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.NotNil(t, err)

}

func TestUnsetMultiplierTwice(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t002,
	}

	m3 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)
	// we want to UnMultiply, but we already UnMultiplied it
	err = InsertOrUpdateMultiplierRepo(&m3)
	assert.NotNil(t, err)

}

func TestUnsetMultiplierWrong(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	m1 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t001,
	}

	//we want to unMultiply, but we are currently not multiplying this repo
	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.NotNil(t, err)
}

func TestMultiplierWrongOrder(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	m2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t001,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	// we want to unMultiply, but the unMultiply date is before the multiply date
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.NotNil(t, err)

}

func TestMultiplierWrongOrderActive(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	m2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t004,
	}

	m3 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)
	//we want to unMultiply, but we already unMultiplied it at 0001-01-01 00:00:04 +0000 UTC
	err = InsertOrUpdateMultiplierRepo(&m3)
	assert.NotNil(t, err)

}

func TestMultiplierCorrect(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepo(t)
	r2 := insertTestRepoGitUrl(t, "git-url2")

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t002,
	}

	m3 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t003,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m3)
	assert.Nil(t, err)

	rs, err := FindMultiplierRepoByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Id, rs[0].Id)

	m4 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            u.Id,
		RepoId:         r2.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t004,
	}
	err = InsertOrUpdateMultiplierRepo(&m4)
	assert.Nil(t, err)

	rs, err = FindMultiplierRepoByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestTwoMultipliedRepos(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	u := insertTestUser(t, "email")
	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          u.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)

	rs, err := FindMultiplierRepoByUserId(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(rs))
}

func TestGetAllFoundationsSupportingReposEmpty(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	_ = insertTestUser(t, "email4")
	_ = insertTestUser(t, "email5")

	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")
	r3 := insertTestRepoGitUrl(t, "git-url3")
	r4 := insertTestRepoGitUrl(t, "git-url4")

	list := []uuid.UUID{r.Id, r2.Id, r3.Id, r4.Id}

	userList, parts, err := GetAllFoundationsSupportingRepos(list)
	assert.Nil(t, err)

	expected := []Foundation{}

	assert.Equal(t, 0, parts)

	assert.Equal(t, len(expected), len(userList), "The number of users returned should match the expected number")

	for _, expectedUser := range expected {
		assert.Contains(t, userList, expectedUser, "Expected user should be in the user list")
	}
}

func TestGetAllFoundationsSupportingReposMany(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	foundation := insertTestFoundation(t, "email", 200)
	foundation2 := insertTestFoundation(t, "email2", 400)
	foundation3 := insertTestFoundation(t, "email3", 1000)
	_ = insertTestUser(t, "email4")
	_ = insertTestUser(t, "email5")

	r := insertTestRepoGitUrl(t, "git-url")
	r2 := insertTestRepoGitUrl(t, "git-url2")
	r3 := insertTestRepoGitUrl(t, "git-url3")
	//r4 := insertTestRepoGitUrl(t, "git-url4")

	list := []uuid.UUID{r.Id, r2.Id, r3.Id}

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation2.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	m3 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	m4 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation3.Id,
		RepoId:       r3.Id,
		EventType:    Active,
		MultiplierAt: &t002,
	}

	err := InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m3)
	assert.Nil(t, err)
	err = InsertOrUpdateMultiplierRepo(&m4)
	assert.Nil(t, err)

	userList, parts, err := GetAllFoundationsSupportingRepos(list)
	assert.Nil(t, err)

	expected := []Foundation{
		{Id: foundation.Id, MultiplierDailyLimit: 200, RepoIds: []uuid.UUID{r2.Id, r.Id}},
		{Id: foundation2.Id, MultiplierDailyLimit: 400, RepoIds: []uuid.UUID{r.Id}},
		{Id: foundation3.Id, MultiplierDailyLimit: 1000, RepoIds: []uuid.UUID{r3.Id}},
	}

	for _, f := range expected {
		sort.Slice(f.RepoIds, func(i, j int) bool { return f.RepoIds[i].String() < f.RepoIds[j].String() })
	}
	for _, f := range userList {
		sort.Slice(f.RepoIds, func(i, j int) bool { return f.RepoIds[i].String() < f.RepoIds[j].String() })
	}

	assert.Equal(t, 4, parts)

	assert.Equal(t, len(expected), len(userList), "The number of users returned should match the expected number")

	assert.ElementsMatch(t, expected, userList, "The contents of the user list should match the expected set")

	m5 := MultiplierEvent{
		Id:             uuid.New(),
		Uid:            foundation3.Id,
		RepoId:         r3.Id,
		EventType:      Inactive,
		UnMultiplierAt: &t003,
	}

	err = InsertOrUpdateMultiplierRepo(&m5)
	assert.Nil(t, err)

	userList, parts, err = GetAllFoundationsSupportingRepos(list)
	assert.Nil(t, err)

	expected = []Foundation{
		{Id: foundation.Id, MultiplierDailyLimit: 200, RepoIds: []uuid.UUID{r2.Id, r.Id}},
		{Id: foundation2.Id, MultiplierDailyLimit: 400, RepoIds: []uuid.UUID{r.Id}},
	}

	for _, f := range expected {
		sort.Slice(f.RepoIds, func(i, j int) bool { return f.RepoIds[i].String() < f.RepoIds[j].String() })
	}
	for _, f := range userList {
		sort.Slice(f.RepoIds, func(i, j int) bool { return f.RepoIds[i].String() < f.RepoIds[j].String() })
	}

	assert.Equal(t, 3, parts)

	assert.Equal(t, len(expected), len(userList), "The number of users returned should match the expected number")

	assert.ElementsMatch(t, expected, userList, "The contents of the user list should match the expected set")
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
