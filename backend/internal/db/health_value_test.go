package db

import (
	"database/sql"
	"fmt"
	"math/big"
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
		ActiveFFSUserCount:  rand.Intn(100) + 1,
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
	err = InsertRepoHealthMetrics(*newRepoMetrics)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "error occured trying to insert: UNIQUE constraint failed: repo_health_metrics.id")
	//r2 := insertTestRepo(t)
}

func TestFindRepoHealthMetricsByRepoIdEmptyRepoId(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	var repoId uuid.UUID
	res, err := FindRepoHealthMetricsByRepoId(repoId)
	assert.Nil(t, res)
	assert.Equal(t, err.Error(), "repoId is empty")
}

func TestFindRepoHealthMetricsByRepoIdMissing(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	res, err := FindRepoHealthMetricsByRepoId(r.Id)
	assert.Nil(t, res)
	assert.Nil(t, err)

	res, err = FindRepoHealthMetricsByRepoId(uuid.New())
	assert.Nil(t, res)
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
func TestFindRepoHealthMetricsByRepoIdHistoryEmpty(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)

	result, err := FindRepoHealthMetricsByRepoIdHistory(r.Id)
	assert.Nil(t, err)
	assert.Empty(t, result)
}

// done
func TestUpdateHealthValue(t *testing.T) {
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
func TestFindHealthValueById(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := FindRepoHealthMetricsById(newRepoMetrics.Id)
	assert.Nil(t, err)
	assert.Equal(t, newRepoMetrics, result)

}

func TestScanRepohealthMetricsEmptyRows(t *testing.T) {
	var rows *sql.Rows
	result, err := scanRepoHealthMetrics(rows)
	assert.Nil(t, err)
	assert.Empty(t, result)
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
	assert.Len(t, result, 8)
	for _ = range 5 {
		_ = InsertRepoHealthMetrics(*getTestData(insertTestRepo(t)))
	}
	result, err = GetAllRepoHealthMetrics()
	assert.Nil(t, err)
	assert.Len(t, result, 13)
}

// not finished, need to merge multiplier to finish the test
func TestGetInternalMetrics(t *testing.T) {
	t.Run("empty situation", func(t *testing.T) {
		SetupTestData()
		defer TeardownTestData()
		r := insertTestRepo(t)

		internalMetrics, _ := GetInternalMetrics(r.Id, false)
		assert.Equal(t, 0, internalMetrics.SponsorCount)
		assert.Equal(t, 0, internalMetrics.RepoStarCount)
		assert.Equal(t, 0, internalMetrics.RepoMultiplierCount)
		assert.Equal(t, 0, internalMetrics.ActiveFFSUserCount)
	})

	t.Run("not an active user staring and contributions", func(t *testing.T) {
		SetupTestData()
		defer TeardownTestData()
		r := insertTestRepo(t)

		uid1 := insertTestUser(t, "email1").Id
		uid2 := insertTestUser(t, "email2").Id
		uid3 := insertTestUser(t, "email3").Id

		_ = InsertContribution(uid1, uid3, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
		_ = InsertContribution(uid2, uid3, r.Id, big.NewInt(3), "XBTC", time.Time{}, time.Time{})

		s1 := SponsorEvent{
			Id:        uuid.New(),
			Uid:       uid1,
			RepoId:    r.Id,
			EventType: Active,
			SponsorAt: &t1,
		}

		s2 := SponsorEvent{
			Id:        uuid.New(),
			Uid:       uid2,
			RepoId:    r.Id,
			EventType: Active,
			SponsorAt: &t2,
		}

		_ = InsertOrUpdateSponsor(&s1)
		_ = InsertOrUpdateSponsor(&s2)

		internalMetrics, _ := GetInternalMetrics(r.Id, false)
		assert.Equal(t, 2, internalMetrics.SponsorCount)
		assert.Equal(t, 0, internalMetrics.RepoStarCount)
		assert.Equal(t, 0, internalMetrics.RepoMultiplierCount)
		assert.Equal(t, 0, internalMetrics.ActiveFFSUserCount)
	})

	t.Run("active/inactive user staring", func(t *testing.T) {
		SetupTestData()
		defer TeardownTestData()
		r := insertTestRepo(t)
		r2 := insertTestRepo(t)

		uid1 := insertTestUser(t, "email1").Id
		uid2 := insertTestUser(t, "email2").Id
		uid3 := insertTestUser(t, "email3").Id

		tr1 := TrustEvent{
			Id:        uuid.New(),
			Uid:       uid3,
			RepoId:    r2.Id,
			EventType: Active,
			TrustAt:   &t1,
		}

		_ = InsertOrUpdateTrustRepo(&tr1)

		currentTime := time.Now()
		previousTime := currentTime.AddDate(0, -3, -1)
		_ = InsertContribution(uid1, uid3, r2.Id, big.NewInt(2), "XBTC", currentTime, currentTime)
		_ = InsertContribution(uid2, uid3, r2.Id, big.NewInt(3), "XBTC", currentTime, previousTime)

		s1 := SponsorEvent{
			Id:        uuid.New(),
			Uid:       uid1,
			RepoId:    r.Id,
			EventType: Active,
			SponsorAt: &t1,
		}

		s2 := SponsorEvent{
			Id:        uuid.New(),
			Uid:       uid2,
			RepoId:    r.Id,
			EventType: Active,
			SponsorAt: &t2,
		}

		_ = InsertOrUpdateSponsor(&s1)
		_ = InsertOrUpdateSponsor(&s2)

		internalMetrics, _ := GetInternalMetrics(r.Id, false)
		assert.Equal(t, 0, internalMetrics.SponsorCount)
		assert.Equal(t, 1, internalMetrics.RepoStarCount)
		assert.Equal(t, 0, internalMetrics.RepoMultiplierCount)
		assert.Equal(t, 0, internalMetrics.ActiveFFSUserCount)
	})

	t.Run("active/inactive user multiplier", func(t *testing.T) {
		SetupTestData()
		defer TeardownTestData()
		r := insertTestRepo(t)
		r2 := insertTestRepo(t)

		uid1 := insertTestUser(t, "email1").Id
		uid2 := insertTestUser(t, "email2").Id
		uid3 := insertTestUser(t, "email3").Id

		tr1 := TrustEvent{
			Id:        uuid.New(),
			Uid:       uid3,
			RepoId:    r2.Id,
			EventType: Active,
			TrustAt:   &t1,
		}

		_ = InsertOrUpdateTrustRepo(&tr1)

		currentTime := time.Now()
		previousTime := currentTime.AddDate(0, -3, -1)
		_ = InsertContribution(uid1, uid3, r2.Id, big.NewInt(2), "XBTC", currentTime, currentTime)
		_ = InsertContribution(uid2, uid3, r2.Id, big.NewInt(3), "XBTC", currentTime, previousTime)

		s1 := MultiplierEvent{
			Id:           uuid.New(),
			Uid:          uid1,
			RepoId:       r.Id,
			EventType:    Active,
			MultiplierAt: &t1,
		}

		s2 := MultiplierEvent{
			Id:           uuid.New(),
			Uid:          uid2,
			RepoId:       r.Id,
			EventType:    Active,
			MultiplierAt: &t2,
		}

		_ = InsertOrUpdateMultiplierRepo(&s1)
		_ = InsertOrUpdateMultiplierRepo(&s2)

		internalMetrics, _ := GetInternalMetrics(r.Id, false)
		assert.Equal(t, 0, internalMetrics.SponsorCount)
		assert.Equal(t, 0, internalMetrics.RepoStarCount)
		assert.Equal(t, 1, internalMetrics.RepoMultiplierCount)
		assert.Equal(t, 0, internalMetrics.ActiveFFSUserCount)
	})

	t.Run("test active FFS user count", func(t *testing.T) {
		SetupTestData()
		defer TeardownTestData()
		r := insertTestRepo(t)
		r2 := insertTestRepo(t)
		r3 := insertTestRepo(t)

		a := AnalysisRequest{
			Id:       uuid.New(),
			RepoId:   r.Id,
			DateFrom: day1,
			DateTo:   day2,
			GitUrl:   *r.GitUrl,
		}
		_ = InsertAnalysisRequest(a, time.Now())

		_ = InsertAnalysisResponse(a.Id, a.RepoId, "email1", []string{"tom"}, 0.5, time.Now())
		_ = InsertAnalysisResponse(a.Id, a.RepoId, "email2", []string{"classi"}, 0.5, time.Now())

		uid1 := insertTestUser(t, "email1").Id
		uid2 := insertTestUser(t, "email2").Id
		uid3 := insertTestUser(t, "email3").Id

		uids := []uuid.UUID{uid1, uid2, uid3}

		for i := range uids {
			uid := uids[i%len(uids)]
			_ = InsertGitEmail(uuid.New(), uid, fmt.Sprintf("email%d", i+1), stringPointer("A"), time.Now())
		}

		tr1 := TrustEvent{
			Id:        uuid.New(),
			Uid:       uid3,
			RepoId:    r2.Id,
			EventType: Active,
			TrustAt:   &t1,
		}

		_ = InsertOrUpdateTrustRepo(&tr1)

		currentTime := time.Now()
		previousTime := currentTime.AddDate(0, -3, -1)
		_ = InsertContribution(uid1, uid3, r2.Id, big.NewInt(2), "XBTC", currentTime, currentTime)
		_ = InsertContribution(uid2, uid3, r2.Id, big.NewInt(3), "XBTC", currentTime, previousTime)
		_ = InsertContribution(uid1, uid2, r3.Id, big.NewInt(3), "XBTC", currentTime, currentTime)

		internalMetricsFocusSponsorEvent, _ := GetInternalMetrics(r.Id, false)
		assert.Equal(t, 0, internalMetricsFocusSponsorEvent.SponsorCount)
		assert.Equal(t, 0, internalMetricsFocusSponsorEvent.RepoStarCount)
		assert.Equal(t, 0, internalMetricsFocusSponsorEvent.RepoMultiplierCount)
		assert.Equal(t, 2, internalMetricsFocusSponsorEvent.ActiveFFSUserCount)
	})
}
