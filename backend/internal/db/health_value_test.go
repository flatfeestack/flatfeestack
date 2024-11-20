package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestData(r *Repo) *RepoHealthMetrics {
	newMetricsId := uuid.New()
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05.999999999")
	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", formatted)

	newRepoMetrics := RepoHealthMetrics{
		Id:                  newMetricsId,
		RepoId:              r.Id,
		CreatedAt:           parsedTime,
		ContributerCount:    rand.Intn(100) + 1,
		CommitCount:         rand.Intn(100) + 1,
		SponsorCount:        rand.Intn(100) + 1,
		RepoStarCount:       rand.Intn(100) + 1,
		RepoMultiplierCount: rand.Intn(100) + 1,
		RepoWeight:          rand.Float64(),
	}

	return &newRepoMetrics

}

// done
func TestInsertTrustValue(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)
}

func TestInsertTrustValueDuplicateId(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)
	err = InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Error(t, err)
}

func TestFindRepoHealthMetricsByRepoId(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)

	result, err := FindRepoHealthMetricsByRepoId(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, result, newRepoMetrics)

	// newRepoMetrics2 := getTestData(r)
	// assert.NotEqual(t, newRepoMetrics, newRepoMetrics2)
	//
	// err = InsertRepoHealthMetrics(*newRepoMetrics2)
	// assert.Nil(t, err)
	//
	// result, err = FindRepoHealthMetricsByRepoId(r.Id)
	// assert.Nil(t, err)
	// assert.Len(t, result, 2)
	// assert.Equal(t, result[0].RepoId, result[1].RepoId)
	// assert.NotEmpty(t, result[0], result[1])
}

func TestFindRepoHealthMetricsByRepoIdHistory(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)

	result, err := FindRepoHealthMetricsByRepoIdHistory(r.Id)
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, result[0], *newRepoMetrics)

	newRepoMetrics2 := getTestData(r)
	assert.NotEqual(t, newRepoMetrics, newRepoMetrics2)

	err = InsertRepoHealthMetrics(*newRepoMetrics2)
	assert.Nil(t, err)

	result, err = FindRepoHealthMetricsByRepoIdHistory(r.Id)
	assert.Nil(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, result[0].RepoId, result[1].RepoId)

}

// done
func TestUpdateTrustValue(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	err := InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Nil(t, err)

	alteredRepoMetrics := *newRepoMetrics
	alteredRepoMetrics.RepoMultiplierCount = rand.Intn(100) + 1
	alteredRepoMetrics.ContributerCount = rand.Intn(100) + 1
	alteredRepoMetrics.RepoStarCount = rand.Intn(100) + 1
	assert.NotEqual(t, newRepoMetrics, alteredRepoMetrics)

	err = UpdateRepoHealthMetrics(alteredRepoMetrics)
	assert.Nil(t, err)

	result, err := FindRepoHealthMetricsById(alteredRepoMetrics.Id)
	assert.Nil(t, err)
	assert.Equal(t, alteredRepoMetrics, *result)

}

// done
func TestFindTrustValueById(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := FindRepoHealthMetricsById(newRepoMetrics.Id)
	assert.Nil(t, err)
	assert.Equal(t, newRepoMetrics, result)

}

// done
func TestGetAllTrustValues(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := GetAllRepoHealthMetrics()
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	for _ = range 5 {
		_ = InsertRepoHealthMetrics(*getTestData(insertTestRepo(t)))
	}
	result, err = GetAllRepoHealthMetrics()
	assert.Nil(t, err)
	assert.Len(t, result, 6)
}

// not finished, need to merge multiplier to finish the test
func TestGetInternalMetrics(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)

	// for _ = range 10 {
	// 	_ = insertTestRepo(t)
	// }

	// uid1 := insertTestUser(t, "email1").Id
	// uid2 := insertTestUser(t, "email2").Id
	// uid3 := insertTestUser(t, "email3").Id

	// _ = InsertContribution(uid1, uid3, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
	// _ = InsertContribution(uid2, uid3, r.Id, big.NewInt(3), "XBTC", time.Time{}, time.Time{})

	// uids := []uuid.UUID{uid1, uid2, uid3}

	// for i := range uids {
	// 	uid := uids[i%len(uids)]
	// 	_ = InsertGitEmail(uuid.New(), uid, fmt.Sprintf("email%d", i), stringPointer("A"), time.Now())
	// }

	// a := AnalysisRequest{
	// 	Id:       uuid.New(),
	// 	RepoId:   r.Id,
	// 	DateFrom: day1,
	// 	DateTo:   day2,
	// 	GitUrl:   *r.GitUrl,
	// }
	// _ = InsertAnalysisRequest(a, time.Now())

	// _ = InsertAnalysisResponse(a.Id, a.RepoId, "email3", []string{"tom"}, 0.5, time.Now())

	// s1 := SponsorEvent{
	// 	Id:        uuid.New(),
	// 	Uid:       uid1,
	// 	RepoId:    r.Id,
	// 	EventType: Active,
	// 	SponsorAt: &t1,
	// }

	// s2 := SponsorEvent{
	// 	Id:        uuid.New(),
	// 	Uid:       uid2,
	// 	RepoId:    r.Id,
	// 	EventType: Active,
	// 	SponsorAt: &t2,
	// }

	// _ = InsertOrUpdateSponsor(&s1)
	// _ = InsertOrUpdateSponsor(&s2)

	// t1 := TrustEvent{
	// 	Id:        uuid.New(),
	// 	Uid:       uid1,
	// 	RepoId:    r.Id,
	// 	EventType: Active,
	// 	TrustAt:   &t1,
	// }

	// _ = InsertOrUpdateTrustRepo(&t1)

	internalMetrics, _ := GetInternalMetrics(r.Id, false)
	assert.Equal(t, 0, internalMetrics.SponsorCount)
	assert.Equal(t, 0, internalMetrics.RepoStarCount)
	assert.Equal(t, 0, internalMetrics.RepoMultiplierCount)
	assert.Equal(t, 0.0, internalMetrics.RepoWeight)
}
