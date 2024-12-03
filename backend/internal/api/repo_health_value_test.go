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
		ContributorCount:    a,
		CommitCount:         b,
		SponsorCount:        c,
		RepoStarCount:       d,
		RepoMultiplierCount: e,
		RepoWeight:          rand.Int(),
	}

	return &newRepoMetrics

}

func stringPointer(s string) *string {
	return &s
}

func insertTestRepo(t *testing.T) *db.Repo {
	return insertTestRepoGitUrl(t, "git-url")
}

func insertTestRepoGitUrl(t *testing.T, gitUrl string) *db.Repo {
	rid := uuid.New()
	r := db.Repo{
		Id:          rid,
		Url:         stringPointer("url"),
		GitUrl:      stringPointer(gitUrl),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := db.InsertOrUpdateRepo(&r)
	assert.Nil(t, err)
	r2, err := db.FindRepoById(r.Id)
	assert.Nil(t, err)
	return r2
}

func TestGetRepoHealthValueByRepoIdNilRepoId(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()

	var repoId uuid.UUID
	result, err := getRepoHealthValue(repoId)
	compareValue := returnZeroHealthValue(uuid.MustParse("00000000-0000-0000-0000-000000000000"))

	assert.NotNil(t, result)
	assert.Equal(t, result, compareValue)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "couldn't get partial health values for repo with id 00000000-0000-0000-0000-000000000000: couldn't get repo health metrics: repoId is empty")
}
func TestGetRepoHealthValueByRepoIdHealthMetricsEmpty(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()

	r := insertTestRepo(t)
	result, err := getRepoHealthValue(r.Id)
	assert.Error(t, err)
	assert.Equal(t, result.HealthValue, float64(0))
}

func TestCalculateRepoHealthValue(t *testing.T) {
	db.SetupTestData()
	defer db.TeardownTestData()

	// For testing purposes, the thresholds are fixed
	threshold := db.RepoHealthThreshold{
		ThContributorCount: &db.Threshold{Upper: 13.0, Lower: 4.0},
		ThCommitCount:      &db.Threshold{Upper: 130.0, Lower: 40.0},
		ThSponsorDonation:  &db.Threshold{Upper: 20.0, Lower: 5.0},
		ThRepoStarCount:    &db.Threshold{Upper: 20.0, Lower: 5.0},
		ThRepoMultiplier:   &db.Threshold{Upper: 20.0, Lower: 5.0},
	}
	metrics := getTestData(3, 131, 5, 20, 12)

	partialResult := calculatePartialHealthValues(&threshold, metrics)
	result := calculateRepoHealthValue(*partialResult)
	assert.Equal(t, result, float64(5.13))
}
