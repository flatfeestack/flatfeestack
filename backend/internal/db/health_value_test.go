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
	assert.Equal(t, result[0], *newRepoMetrics)
	assert.Len(t, result, 1)

	newRepoMetrics2 := getTestData(r)
	assert.NotEqual(t, newRepoMetrics, newRepoMetrics2)

	err = InsertRepoHealthMetrics(*newRepoMetrics2)
	assert.Nil(t, err)

	result, err = FindRepoHealthMetricsByRepoId(r.Id)
	assert.Nil(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, result[0].RepoId, result[1].RepoId)
	assert.NotEmpty(t, result[0], result[1])
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
	SetupTestData()
	defer TeardownTestData()
	r := insertTestRepo(t)
	newRepoMetrics := getTestData(r)
	_ = InsertRepoHealthMetrics(*newRepoMetrics)
	result, err := GetAllTrustValues()
	assert.Nil(t, err)
	assert.Len(t, result, 1)
	for _ = range 5 {
		_ = InsertRepoHealthMetrics(*getTestData(insertTestRepo(t)))
	}
	result, err = GetAllTrustValues()
	assert.Nil(t, err)
	assert.Len(t, result, 6)
}
