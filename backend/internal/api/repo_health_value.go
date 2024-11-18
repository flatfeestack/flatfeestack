package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type RepoHealthValue struct {
	RepoId      uuid.UUID `json:"repoid"`
	HealthValue float32   `json:"healthvalue"`
}

func GetRepoHealthValueByRepoId(w http.ResponseWriter, r *http.Request) {
	repoId := uuid.MustParse(r.PathValue("id"))
	healthValue, err := getRepoHealthValue(repoId)

	if healthValue == nil {
		slog.Error("Repo Health Metrics not found %s",
			slog.String("id", repoId.String()))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		slog.Error("Could not fetch trust value",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, healthValue)
	}
}

func UpdateRepoHealthValue(w http.ResponseWriter, r *http.Request, repoHealthMetrics *db.RepoHealthMetrics) {
	if repoHealthMetrics.Id == uuid.Nil {
		slog.Error("RepoId is missing",
			slog.String("Repo Health Metrics id", repoHealthMetrics.Id.String()))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	err := db.UpdateRepoHealthMetrics(*repoHealthMetrics)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not update trust value: %v", err)
		return
	}
	repoHealthMetrics, err = db.FindRepoHealthMetricsById(repoHealthMetrics.Id)
	if err != nil {
		util.WriteErrorf(w, http.StatusNoContent, "Could not find trust value: %v", err)
		return
	}

	util.WriteJson(w, repoHealthMetrics)
}

func manageRepoHealthMetrics(data []FlatFeeWeight, repoId uuid.UUID) error {
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
	internalHealthMetric, err := db.GetInternalMetrics(repoId, true)
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

func getRepoHealthValueHistory(repoId uuid.UUID) ([]RepoHealthValue, error) {
	healthMetrics, err := db.FindRepoHealthMetricsByRepoIdHistory(repoId)
	if err != nil {
		return nil, err
	}
	healthThreshold, err := db.GetLatestThresholds()
	if err != nil {
		return nil, err
	}

	var repoHealthHistory []RepoHealthValue
	for _, metrics := range healthMetrics {
		tmp, err := calculateRepoHealthValue(healthThreshold, &metrics)
		if err != nil {
			return nil, err
		}
		repoHealthHistory = append(repoHealthHistory, *tmp)
	}
	return repoHealthHistory, nil
}

func getRepoHealthValue(repoId uuid.UUID) (*RepoHealthValue, error) {
	healthMetrics, err := db.FindRepoHealthMetricsByRepoId(repoId)
	if err != nil {
		return nil, err
	}
	healthThreshold, err := db.GetLatestThresholds()
	if err != nil {
		return nil, err
	}

	repoHealthValue, err := calculateRepoHealthValue(healthThreshold, healthMetrics)
	if err != nil {
		return nil, err
	}
	return repoHealthValue, nil
}

func calculateRepoHealthValue(threshold *db.RepoHealthThreshold, metrics *db.RepoHealthMetrics) (*RepoHealthValue, error) {
	healthValueObject := RepoHealthValue{
		RepoId:      metrics.RepoId,
		HealthValue: 0.0,
	}
	healthMetrics := []int{
		metrics.ContributerCount,
		metrics.CommitCount,
		metrics.SponsorCount,
		metrics.RepoStarCount,
		metrics.RepoMultiplierCount,
	}

	healthThreshold := [][]float32{
		{threshold.ThContributerCount.Lower, threshold.ThContributerCount.Upper},
		{threshold.ThCommitCount.Lower, threshold.ThCommitCount.Upper},
		{threshold.ThSponsorDonation.Lower, threshold.ThSponsorDonation.Upper},
		{threshold.ThRepoStarCount.Lower, threshold.ThRepoStarCount.Upper},
		{threshold.ThRepoMultiplier.Lower, threshold.ThRepoMultiplier.Upper},
	}

	for i := range 5 {
		var partialHealthValue float32
		currentMetric := float32(healthMetrics[i])
		currentThresholdLower := healthThreshold[i][0]
		currentThresholdUpper := healthThreshold[i][1]

		if currentMetric <= currentThresholdLower {
			partialHealthValue = 0.0
		} else if currentMetric >= currentThresholdUpper {
			partialHealthValue = 2.0
		} else {
			thresholdDifference := currentThresholdUpper - currentThresholdLower - 1
			normalizedCurrentMetric := currentMetric - currentThresholdLower + 1
			partialHealthValue = (2 / thresholdDifference) * normalizedCurrentMetric
		}
		healthValueObject.HealthValue += partialHealthValue
	}

	return &healthValueObject, nil
}
