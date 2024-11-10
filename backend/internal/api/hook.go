package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func AnalysisEngineHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data WebhookCallback
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		slog.Error("Could not decode Webhook body",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	fmt.Println(data.ContribCommit)

	reqId, err := uuid.Parse(data.RequestId)
	if err != nil {
		slog.Error("Cannot parse request id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rowsAffected := 0
	for _, v := range data.Result {
		err = db.InsertAnalysisResponse(reqId, v.Email, v.Names, v.Weight, util.TimeNow())
		if err != nil {
			slog.Error("Insert problem",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
			return
		}
		rowsAffected++
	}

	errA := db.UpdateAnalysisRequest(reqId, util.TimeNow(), data.Error)
	if errA != nil {
		slog.Warn("Update problem",
			slog.Any("error", errA))
	}

	/*
		Add magical trust value update

		TODO: Mino add additional functionality for the other three values
		errTV := foobar.MinosMagicalFunction(data.ContribCommit)
	*/

	trustValueMetrics := db.TrustValueMetrics{
		RepoId:                data.ContribCommit.RepoId,
		ContributerCount:      data.ContribCommit.ContributerCount,
		CommitCount:           data.ContribCommit.CommitCount,
		SponsorCount:          0,
		SponsorStarMultiplier: 0,
		RepoSponsorDonated:    0,
	}

	errTV := db.InsertTrustValue(trustValueMetrics)
	if errTV != nil {
		slog.Warn("Update problem into trustValueMetrics",
			slog.Any("error", errTV))
	}

	slog.Info("Analysis stats",
		slog.Int("rowsAffected", rowsAffected),
		slog.String("requestId", data.RequestId))
	w.WriteHeader(http.StatusOK)
}
