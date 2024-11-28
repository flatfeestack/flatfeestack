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

type PartialHealthValues struct {
	RepoId              uuid.UUID `json:"repoid"`
	ContributorValue    float64   `json:"contributorvalue"`
	CommitValue         float64   `json:"commitvalue"`
	SponsorValue        float64   `json:"sponsorvalue"`
	RepoStarValue       float64   `json:"repostarvalue"`
	RepoMultiplierValue float64   `json:"repomultipliervalue"`
}

func GetRepoHealthValueByRepoId(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	repoId := uuid.MustParse(r.PathValue("id"))
	healthValue, err := getRepoHealthValue(repoId)

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
func GetRepoMetricsById(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	repoId := uuid.MustParse(r.PathValue("id"))
	repoMetrics, err := db.FindRepoHealthMetricsByRepoId(repoId)

	if repoMetrics == nil {
		slog.Error("repo metrics not found %s",
			slog.String("id", repoId.String()))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
		//util.WriteJson(w, &RepoHealthValue{RepoId: repoId, HealthValue: 0})
	} else if err != nil {
		slog.Error("Could not fetch repo metrics",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		//util.WriteJson(w, &RepoHealthValue{RepoId: repoId, HealthValue: 0})
	} else {
		util.WriteJson(w, repoMetrics)
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
	contributorCount := 0
	var commitCount int
	var repoWeight float64
	var repoHealthMetrics *db.RepoHealthMetrics
	var err error

	for _, email := range data {
		contributorCount++
		commitCount += email.CommitCount
		repoWeight += email.Weight
	}

	repoHealthMetrics, err = manageInternalHealthMetrics(repoId, true)
	if err != nil {
		log.Printf("This is an arrow: %v", err)
	}

	repoHealthMetrics.Id = uuid.New()
	repoHealthMetrics.RepoId = repoId
	repoHealthMetrics.CreatedAt = util.TimeNow()
	repoHealthMetrics.ContributorCount = contributorCount
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
		//tmp := calculateRepoHealthValue(healthThreshold, &metrics)
		partialTmp := getPartialRepoHealthValues(healthThreshold, &metrics)
		tmp := calculateRepoHealthValue(*partialTmp)
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
	partialHealthValue := getPartialRepoHealthValues(healthThreshold, healthMetrics)
	healthValue := calculateRepoHealthValue(*partialHealthValue)
	return &RepoHealthValue{
		RepoId:      repoId,
		HealthValue: healthValue,
	}, nil
}

func getPartialRepoHealthValues(threshold *db.RepoHealthThreshold, metrics *db.RepoHealthMetrics) *PartialHealthValues {
	healthMetrics := []int{
		metrics.ContributorCount,
		metrics.CommitCount,
		metrics.SponsorCount,
		metrics.RepoStarCount,
		metrics.RepoMultiplierCount,
	}
	var partialHealthValues []float64

	healthThreshold := [][]int{
		{threshold.ThContributorCount.Lower, threshold.ThContributorCount.Upper},
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
		//healthValue += partialHealthValue
		partialHealthValues = append(partialHealthValues, partialHealthValue)
	}
	return &PartialHealthValues{
		ContributorValue:    partialHealthValues[0],
		CommitValue:         partialHealthValues[1],
		SponsorValue:        partialHealthValues[2],
		RepoStarValue:       partialHealthValues[3],
		RepoMultiplierValue: partialHealthValues[4],
	}
}

func calculateRepoHealthValue(partialHealthValues PartialHealthValues) float64 {
	healthValue := partialHealthValues.CommitValue + partialHealthValues.ContributorValue + partialHealthValues.SponsorValue + partialHealthValues.RepoStarValue + partialHealthValues.RepoMultiplierValue
	return math.Round(healthValue*100) / 100
}
