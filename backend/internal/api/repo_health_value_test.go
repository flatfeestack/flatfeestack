package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestGetRepoHealthMetricsById(t *testing.T) {
	t.Run("should return t", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()
		r := insertTestRepo(t)

		insertTestRepoHealthMetrics(t)
	})
}

func insertTestRepoHealthMetrics(t *testing.T) {
	repoId := uuid.New()
	randomString := "foobar"
	newRepo := db.Repo{
		id:     repoId,
		Url:    &randomString,
		GitUrl: &randomString,
		Name:   &randomString,
		Source: "github.com", 
		CreatedAt:,
	}
	newRepoMetrics := db.RepoHealthMetrics{
		Id:                  uuid.New(),
		RepoId:              uuid.New(),
		ContributerCount:    1,
		CommitCount:         2,
		SponsorCount:        3,
		RepoStarCount:       4,
		RepoMultiplierCount: 5,
		RepoWeight:          6,
	}

}

func TestUpdateTrustValue(w http.ResponseWriter, r *http.Request, trustValue *db.RepoHealthMetrics) {
	if trustValue.Id == uuid.Nil {
		slog.Error("RepoId is missing",
			slog.String("Trust Value id", trustValue.Id.String()))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	err := db.UpdateTrustValue(*trustValue)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not update trust value: %v", err)
		return
	}
	trustValue, err = db.FindTrustValueById(trustValue.Id)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not find trust value: %v", err)
		return
	}

	util.WriteJson(w, trustValue)
}

func TestGetAllTrustValuesUnique(w http.ResponseWriter, r *http.Request) {

}

func TestmanageRepoHealthMetrics(data []FlatFeeWeight, repoId uuid.UUID) error {
	//var repoHealthMetrics db.TrustValueMetrics

	contributerCount := 0
	var commitCount int
	var repoWeight float64
	for _, email := range data {
		contributerCount++
		commitCount += email.CommitCount
		repoWeight += email.Weight
	}

	repoHealthMetrics := manageInternalHealthMetrics(repoId)
	repoHealthMetrics.Id = uuid.New()
	repoHealthMetrics.RepoId = repoId
	repoHealthMetrics.ContributerCount = contributerCount
	repoHealthMetrics.CommitCount = commitCount
	repoHealthMetrics.RepoWeight = repoWeight

	return nil
}

func TestmanageInternalHealthMetrics(repoId uuid.UUID) *db.RepoHealthMetrics {
	// get magical internal metrics

	// plox mino
	var sponsorCount, repoStarCount, repoMultiplierCount int
	//sponsorCount, repoStarCount, repoMultiplierCount := db.GetInternalMetrics(repoid)

	repoHealthMetrics := db.RepoHealthMetrics{
		SponsorCount:        sponsorCount,
		RepoStarCount:       repoStarCount,
		RepoMultiplierCount: repoMultiplierCount,
	}

	return &repoHealthMetrics
}
