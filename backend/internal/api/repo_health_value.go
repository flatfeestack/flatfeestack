package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"time"

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
	ActiveFFSUserValue  float64   `json:"activeffsuservalue"`
}

func GetRepoHealthValueByRepoId(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	repoId := uuid.MustParse(r.PathValue("id"))
	healthValue, err := getRepoHealthValue(repoId)

	if healthValue == nil {
		slog.Error("Health Value not found %s",
			slog.String("id", repoId.String()))
		util.WriteJson(w, &RepoHealthValue{RepoId: repoId, HealthValue: 0})
	} else if err != nil {
		slog.Error("Could not fetch health value",
			slog.Any("error", err))
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
		util.WriteErrorf(w, http.StatusNotFound, NoRepoMetricsAvailable)
	} else if err != nil {
		slog.Error("Could not fetch repo metrics",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		parsed, _ := time.Parse("2006-01-02 15:04:05", repoMetrics.CreatedAt.Format("2006-01-02 15:04:05"))
		repoMetrics.CreatedAt = parsed

		util.WriteJson(w, repoMetrics)
	}

}

func GetPartialHealthValuesById(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	repoId := uuid.MustParse(r.PathValue("id"))
	partialHealthValues, err := getPartialRepoHealthValues(repoId)

	if partialHealthValues == nil {
		slog.Error("repo metrics not found %s",
			slog.String("id", repoId.String()))
		util.WriteErrorf(w, http.StatusNotFound, NoPartialHealthValuesAvailable)
	} else if err != nil {
		slog.Error("Could not fetch repo metrics",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, partialHealthValues)
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
	var repoHealthMetrics *db.RepoHealthMetrics
	var err error

	for _, email := range data {
		contributorCount++
		commitCount += email.CommitCount
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

	slog.Info("inserting new repo health metrics",
		slog.Any("Repo: %v", repoHealthMetrics))

	err = db.InsertRepoHealthMetrics(*repoHealthMetrics)
	if err != nil {
		return err
	}

	return nil
}

func manageInternalHealthMetrics(repoId uuid.UUID, isPostgres bool) (*db.RepoHealthMetrics, error) {
	internalHealthMetric, err := db.GetInternalMetrics(repoId, isPostgres)
	if err != nil {
		slog.Warn("issues getting internalHealthMetrics",
			slog.Any("error", err))
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
		partialTmp := calculatePartialHealthValues(db.DefaultMetricWeight, healthThreshold, &metrics)
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
	partialHealthValue, err := getPartialRepoHealthValues(repoId)
	if err != nil {
		return returnZeroHealthValue(repoId), fmt.Errorf("couldn't get partial health values for repo with id %v: %v", repoId, err)
	}

	healthValue := calculateRepoHealthValue(*partialHealthValue)
	return &RepoHealthValue{
		RepoId:      repoId,
		HealthValue: healthValue,
	}, nil
}

func getPartialRepoHealthValues(repoId uuid.UUID) (*PartialHealthValues, error) {
	healthMetrics, err := db.FindRepoHealthMetricsByRepoId(repoId)
	if err != nil {
		return nil, fmt.Errorf("couldn't get repo health metrics: %v", err)
	}

	if healthMetrics == nil {
		return nil, fmt.Errorf("no health metrics founds to corresponding repoId, %v", repoId)
	}

	healthThreshold, err := db.GetLatestThresholds()
	if err != nil {
		return nil, fmt.Errorf("couldn't get latest threshold values: %v", err)
	}

	return calculatePartialHealthValues(db.DefaultMetricWeight, healthThreshold, healthMetrics), nil
}

func calculatePartialHealthValues(weights *db.MetricWeight, threshold *db.RepoHealthThreshold, metrics *db.RepoHealthMetrics) *PartialHealthValues {
	return &PartialHealthValues{
		ContributorValue:    calcValue(metrics.ContributorCount, threshold.ThContributorCount.Lower, threshold.ThContributorCount.Upper, weights.WeightContributorCount),
		CommitValue:         calcValue(metrics.CommitCount, threshold.ThCommitCount.Lower, threshold.ThCommitCount.Upper, weights.WeightCommitCount),
		SponsorValue:        calcValue(metrics.SponsorCount, threshold.ThSponsorDonation.Lower, threshold.ThSponsorDonation.Upper, weights.WeightSponsorDonation),
		RepoStarValue:       calcValue(metrics.RepoStarCount, threshold.ThRepoStarCount.Lower, threshold.ThRepoStarCount.Upper, weights.WeightRepoStarCount),
		RepoMultiplierValue: calcValue(metrics.RepoMultiplierCount, threshold.ThRepoMultiplier.Lower, threshold.ThRepoMultiplier.Upper, weights.WeightRepoMultiplier),
		ActiveFFSUserValue:  calcValue(metrics.RepoWeight, threshold.ThActiveFFSUserCount.Lower, threshold.ThActiveFFSUserCount.Upper, weights.WeightActiveFFSUserCount),
	}
}

func calcValue(metric, lower, upper, weight int) float64 {
	if metric < lower {
		return 0.0
	} else if metric > upper {
		return float64(weight)
	} else {
		thresholdDifference := upper - lower + 1
		normalizedCurrentMetric := metric - lower + 1
		return math.Round(100*(float64(weight)/float64(thresholdDifference)*float64(normalizedCurrentMetric))) / 100
	}
}

func calculateRepoHealthValue(partialHealthValues PartialHealthValues) float64 {
	healthValue := partialHealthValues.CommitValue + partialHealthValues.ContributorValue + partialHealthValues.SponsorValue + partialHealthValues.RepoStarValue + partialHealthValues.RepoMultiplierValue + partialHealthValues.ActiveFFSUserValue
	return math.Round(healthValue*100) / 100
}

func returnZeroHealthValue(repoId uuid.UUID) *RepoHealthValue {
	return &RepoHealthValue{RepoId: repoId, HealthValue: 0}
}
