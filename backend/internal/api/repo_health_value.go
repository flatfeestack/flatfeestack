package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func GetTrustValueById(w http.ResponseWriter, r *http.Request) {
	var id uuid.UUID
	id = uuid.MustParse(r.PathValue("id"))
	//convertedTrustValueId, err := strconv.Atoi(id)

	//if err != nil {
	//	slog.Error("Invalid user ID",
	//		slog.Any("error", err))
	//	util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	//}

	trustValue, err := db.FindTrustValueById(id)

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

	repoHealthMetrics := manageInternalHealthMetrics(repoId)
	repoHealthMetrics.Id = uuid.New()
	repoHealthMetrics.RepoId = repoId
	repoHealthMetrics.ContributerCount = contributerCount
	repoHealthMetrics.CommitCount = commitCount
	repoHealthMetrics.RepoWeight = repoWeight

	return nil
}

func manageInternalHealthMetrics(repoId uuid.UUID) *db.RepoHealthMetrics {
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
