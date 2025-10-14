package api

import (
	"backend/db"
	"backend/util"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func GetLatestThresholds(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	res, err := db.GetLatestThresholds()
	if res == nil {

	} else if err != nil {
		slog.Error("No Repo Health Value Thresholds available %s")
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else {
		util.WriteJson(w, res)
	}
}

func GetThresholdHistory(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	res, err := db.GetRepoThresholdHistory()
	if res == nil {

	} else if err != nil {
		slog.Error("No Repo Health Value Thresholds available %s")
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else {
		util.WriteJson(w, res)
	}
}

func SetNewThresholds(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	var newThresholds db.RepoHealthThreshold

	if err := json.NewDecoder(r.Body).Decode(&newThresholds); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		util.WriteErrorf(w, http.StatusBadRequest, "Invalid request format, Thresholds must be integers")
		return
	}
	defer r.Body.Close()

	err := validateThresholds(&newThresholds)
	if err != nil {
		slog.Error("Invalid threshold values", "error", err)
		util.WriteErrorf(w, http.StatusBadRequest, err.Error())
		return
	}

	newThresholds.Id = uuid.New()
	newThresholds.CreatedAt = time.Now()

	// db call
	err = db.InsertRepoHealthThreshold(newThresholds)
	if err != nil {
		slog.Error("failed to insert new threshold", "error", err)
		util.WriteErrorf(w, http.StatusBadRequest, "error while inserting new threshold")
		return
	}

	util.WriteJson(w, newThresholds)
}

func validateThresholds(t *db.RepoHealthThreshold) error {
	if t.ThContributorCount == nil ||
		t.ThCommitCount == nil ||
		t.ThSponsorDonation == nil ||
		t.ThRepoStarCount == nil ||
		t.ThRepoMultiplier == nil ||
		t.ThActiveFFSUserCount == nil {
		return fmt.Errorf("all threshold fields are required")
	}

	thresholds := map[string]*db.Threshold{
		"contributor_count":     t.ThContributorCount,
		"commit_count":          t.ThCommitCount,
		"sponsor_donation":      t.ThSponsorDonation,
		"repo_star_count":       t.ThRepoStarCount,
		"repo_multiplier":       t.ThRepoMultiplier,
		"active_ffs_user_count": t.ThActiveFFSUserCount,
	}

	for name, th := range thresholds {
		if th.Lower < 0 {
			return fmt.Errorf("%s lower threshold cannot be negative", name)
		}
		if th.Upper < 0 {
			return fmt.Errorf("%s upper threshold cannot be negative", name)
		}

		if th.Upper < th.Lower {
			return fmt.Errorf("%s upper threshold cannot be less than lower threshold", name)
		}
	}

	return nil
}
