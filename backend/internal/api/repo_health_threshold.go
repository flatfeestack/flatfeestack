package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"log/slog"
	"net/http"
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

func GetThresholdHistory(w http.ResponseWriter, r *http.Request) {
	res, err := db.GetRepoThresholdHistory()
	if res == nil {

	} else if err != nil {
		slog.Error("No Repo Health Value Thresholds available %s")
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else {
		util.WriteJson(w, res)
	}
}
