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
		ThContributerCount: &Threshold{Upper: rand.Float32() + 1, Lower: rand.Float32() + 1},
		ThCommitCount:      &Threshold{Upper: rand.Float32() + 1, Lower: rand.Float32() + 1},
		ThSponsorDonation:  &Threshold{Upper: rand.Float32() + 1, Lower: rand.Float32() + 1},
		ThRepoStarCount:    &Threshold{Upper: rand.Float32() + 1, Lower: rand.Float32() + 1},
		ThRepoMultiplier:   &Threshold{Upper: rand.Float32() + 1, Lower: rand.Float32() + 1},
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
}

func TestGetLatestThresholds(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	for range 5 {
		_ = InsertRepoHealthThreshold(*getRepoHealthThresholdtTestData())
	}
	newRepoHealthThreshold := getRepoHealthThresholdtTestData()
	_ = InsertRepoHealthThreshold(*newRepoHealthThreshold)

	res, err := GetLatestThresholds()
	assert.Nil(t, err)
	assert.Equal(t, newRepoHealthThreshold, res)
}

func TestGetRepoThresholdHistory(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	for range 5 {
		_ = InsertRepoHealthThreshold(*getRepoHealthThresholdtTestData())
	}
	res, err := GetRepoThresholdHistory()
	for _, value := range res {
		t.Logf("%v\n", value.ThCommitCount)
	}
	assert.Nil(t, err)
	assert.Len(t, res, 5)
}
