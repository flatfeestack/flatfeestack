package api

import (
	"backend/internal/db"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func getTestData(a, b, c, d, e int) *db.RepoHealthMetrics {
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05.999999999")
	parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", formatted)

	newRepoMetrics := db.RepoHealthMetrics{
		Id:                  uuid.New(),
		RepoId:              uuid.New(),
		CreatedAt:           parsedTime,
		ContributerCount:    a,
		CommitCount:         b,
		SponsorCount:        c,
		RepoStarCount:       d,
		RepoMultiplierCount: e,
		RepoWeight:          rand.Float64(),
	}

	return &newRepoMetrics

}

// i'm da best
func TestCalculateRepoHealthValue(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()

	// For testing purposes, the thresholds are fixed
	threshold := db.RepoHealthThreshold{
		ThContributerCount: &db.Threshold{Upper: 13.0, Lower: 4.0},
		ThCommitCount:      &db.Threshold{Upper: 130.0, Lower: 40.0},
		ThSponsorDonation:  &db.Threshold{Upper: 20.0, Lower: 5.0},
		ThRepoStarCount:    &db.Threshold{Upper: 20.0, Lower: 5.0},
		ThRepoMultiplier:   &db.Threshold{Upper: 20.0, Lower: 5.0},
	}
	metrics := getTestData(3, 131, 5, 20, 12)

	result, err := calculateRepoHealthValue(&threshold, metrics)
	assert.Nil(t, err)
	assert.Equal(t, result.HealthValue, float32(5.13))

}
