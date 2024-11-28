package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getRepoHealthThresholdtTestData() *RepoHealthThreshold {
	newId := uuid.New()
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05.999999999")
	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", formatted)

	newThresholdData := RepoHealthThreshold{
		Id:                 newId,
		CreatedAt:          parsedTime,
		ThContributorCount: &Threshold{Upper: rand.Int() + 1, Lower: rand.Int() + 1},
		ThCommitCount:      &Threshold{Upper: rand.Int() + 1, Lower: rand.Int() + 1},
		ThSponsorDonation:  &Threshold{Upper: rand.Int() + 1, Lower: rand.Int() + 1},
		ThRepoStarCount:    &Threshold{Upper: rand.Int() + 1, Lower: rand.Int() + 1},
		ThRepoMultiplier:   &Threshold{Upper: rand.Int() + 1, Lower: rand.Int() + 1},
	}

	return &newThresholdData
}

// tested
func TestInsertRepoHealthThreshold(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	newRepoHealthThreshold := getRepoHealthThresholdtTestData()
	err := InsertRepoHealthThreshold(*newRepoHealthThreshold)
	assert.Nil(t, err)
	err = InsertRepoHealthThreshold(*newRepoHealthThreshold)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "error inserting repo health threshold: UNIQUE constraint failed: repo_health_threshold.id")
}

func TestInsertRepoHealthThresholdNil(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	var newRepoHealthThreshold RepoHealthThreshold
	err := InsertRepoHealthThreshold(newRepoHealthThreshold)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Threshold values can't be empty, aborting")
}

func TestGetLatestThresholds(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	res, err := GetLatestThresholds()
	assert.Nil(t, err)

	for range 5 {
		_ = InsertRepoHealthThreshold(*getRepoHealthThresholdtTestData())
	}
	newRepoHealthThreshold := getRepoHealthThresholdtTestData()
	_ = InsertRepoHealthThreshold(*newRepoHealthThreshold)

	res, err = GetLatestThresholds()
	assert.Nil(t, err)
	assert.Equal(t, newRepoHealthThreshold, res)
}

func TestGetRepoThresholdHistory(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	iterations := 5
	for range iterations {
		_ = InsertRepoHealthThreshold(*getRepoHealthThresholdtTestData())
	}
	res, err := GetRepoThresholdHistory()
	assert.Nil(t, err)
	assert.Len(t, res, 13)
}

func TestInitialValue(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	res, err := GetFirstRepoHealthThreshold()
	assert.Nil(t, err)
	assert.Equal(t, uuid.MustParse("b7244c4a-dadd-45f5-bd12-0fcefb5d66c2"), res.Id)
}
