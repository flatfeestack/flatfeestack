package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func GetRepoHealthMetricsById(w http.ResponseWriter, r *http.Request) {
	id := uuid.MustParse(r.PathValue("id"))
	trustValue, err := db.FindRepoHealthMetricsById(id)

	if trustValue == nil {
		slog.Error("Trust Value not found %s",
			slog.String("id", id.String()))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		slog.Error("Could not fetch trust value",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, trustValue)
	}
}

func UpdateTrustValue(w http.ResponseWriter, r *http.Request, trustValue *db.RepoHealthMetrics) {
	if trustValue.Id == uuid.Nil {
		slog.Error("RepoId is missing",
			slog.String("Trust Value id", trustValue.Id.String()))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	err := db.UpdateRepoHealthMetrics(*trustValue)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not update trust value: %v", err)
		return
	}
	trustValue, err = db.FindRepoHealthMetricsById(trustValue.Id)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not find trust value: %v", err)
		return
	}

	util.WriteJson(w, trustValue)
}

func GetAllTrustValuesUnique(w http.ResponseWriter, r *http.Request) {

}

func manageRepoHealthMetrics(data []FlatFeeWeight, repoId uuid.UUID) error {
	//var repoHealthMetrics db.TrustValueMetrics

	contributerCount := 0
	var commitCount int
	var repoWeight float64
	for _, email := range data {
		contributerCount++
		commitCount += email.CommitCount
		repoWeight += email.Weight
	}

	repoHealthMetrics, err := manageInternalHealthMetrics(repoId)
	if err != nil {
		return err
	}

	repoHealthMetrics.Id = uuid.New()
	repoHealthMetrics.RepoId = repoId
	repoHealthMetrics.ContributerCount = contributerCount
	repoHealthMetrics.CommitCount = commitCount
	repoHealthMetrics.RepoWeight = repoWeight

	err = db.InsertRepoHealthMetrics(*repoHealthMetrics)
	if err != nil {
		return err
	}

	return nil
}

func manageInternalHealthMetrics(repoId uuid.UUID) (*db.RepoHealthMetrics, error) {
	internalHealthMetric, err := db.GetInternalMetrics(repoId)
	if err != nil {
		return nil, err
	}

	repoHealthMetrics := db.RepoHealthMetrics{
		SponsorCount:        internalHealthMetric.SponsorCount,
		RepoStarCount:       internalHealthMetric.RepoStarCount,
		RepoMultiplierCount: internalHealthMetric.RepoMultiplierCount,
	}

	return &repoHealthMetrics, nil
}
