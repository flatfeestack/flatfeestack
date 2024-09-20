package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
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

	slog.Info("Analysis stats",
		slog.Int("rowsAffected", rowsAffected),
		slog.String("requestId", data.RequestId))
	w.WriteHeader(http.StatusOK)
}
