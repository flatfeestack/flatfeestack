package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"

	"github.com/google/uuid"
)

type RepoHealthValue struct {
	RepoId      uuid.UUID `json:"repoid"`
	HealthValue float64   `json:"healthvalue"`
}

func GetRepoHealthValueByRepoId(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	repoId := uuid.MustParse(r.PathValue("id"))
	healthValue, err := getRepoHealthValue(repoId)

	//log.Printf("My healthvalue: %v", healthValue)
	log.Printf("My error: %v", err)

	if healthValue == nil {
		slog.Error("Health Value not found %s",
			slog.String("id", repoId.String()))
		//util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
		util.WriteJson(w, &RepoHealthValue{RepoId: repoId, HealthValue: 0})
	} else if err != nil {
		slog.Error("Could not fetch health value",
			slog.Any("error", err))
		//util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		util.WriteJson(w, &RepoHealthValue{RepoId: repoId, HealthValue: 0})
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
	var repoHealthMetrics *db.RepoHealthMetrics
	var err error

	for _, email := range data {
		contributerCount++
		commitCount += email.CommitCount
		repoWeight += email.Weight
	}

	repoHealthMetrics, err = manageInternalHealthMetrics(repoId, true)
	if err != nil {
		log.Printf("This is an arrow: %v", err)
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

func manageInternalHealthMetrics(repoId uuid.UUID, isPostgres bool) (*db.RepoHealthMetrics, error) {
	internalHealthMetric, err := db.GetInternalMetrics(repoId, isPostgres)
	if err != nil {
		return &db.RepoHealthMetrics{
			SponsorCount:        0,
			RepoStarCount:       0,
			RepoMultiplierCount: 0,
		}, err
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
		tmp := calculateRepoHealthValue(healthThreshold, &metrics)
		repoHealthHistory = append(
			repoHealthHistory,
			RepoHealthValue{
				RepoId:      repoId,
				HealthValue: tmp})
	}
	return repoHealthHistory, nil
}

func getRepoHealthValue(repoId uuid.UUID) (*RepoHealthValue, error) {
	healthMetrics, err := db.FindRepoHealthMetricsByRepoId(repoId)
	if err != nil {
		return nil, fmt.Errorf("couldn't get repo health metrics: %v", err)
	}

	if healthMetrics == nil {
		return &RepoHealthValue{
			RepoId:      repoId,
			HealthValue: 0,
		}, nil
	}

	healthThreshold, err := db.GetLatestThresholds()
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest threshold values: %v", err)
	}
	healthValue := calculateRepoHealthValue(healthThreshold, healthMetrics)
	return &RepoHealthValue{
		RepoId:      repoId,
		HealthValue: healthValue,
	}, nil
}

func calculateRepoHealthValue(threshold *db.RepoHealthThreshold, metrics *db.RepoHealthMetrics) float64 {
	healthValue := 0.0

	healthMetrics := []int{
		metrics.ContributerCount,
		metrics.CommitCount,
		metrics.SponsorCount,
		metrics.RepoStarCount,
		metrics.RepoMultiplierCount,
	}

	healthThreshold := [][]int{
		{threshold.ThContributerCount.Lower, threshold.ThContributerCount.Upper},
		{threshold.ThCommitCount.Lower, threshold.ThCommitCount.Upper},
		{threshold.ThSponsorDonation.Lower, threshold.ThSponsorDonation.Upper},
		{threshold.ThRepoStarCount.Lower, threshold.ThRepoStarCount.Upper},
		{threshold.ThRepoMultiplier.Lower, threshold.ThRepoMultiplier.Upper},
	}

	for i := range 5 {
		var partialHealthValue float64
		currentMetric := healthMetrics[i]
		currentThresholdLower := healthThreshold[i][0]
		currentThresholdUpper := healthThreshold[i][1]

		if currentMetric < currentThresholdLower {
			partialHealthValue = 0.0
		} else if currentMetric > currentThresholdUpper {
			partialHealthValue = 2.0
		} else {
			thresholdDifference := currentThresholdUpper - currentThresholdLower + 1
			normalizedCurrentMetric := currentMetric - currentThresholdLower + 1
			partialHealthValue = 2 / float64(thresholdDifference) * float64(normalizedCurrentMetric)
		}
		healthValue += partialHealthValue
	}
	return math.Round(healthValue*100) / 100

	//return &RepoHealthValue{
	//	RepoId:      metrics.RepoId,
	//	HealthValue: healthValue,
	//}, nil
}
