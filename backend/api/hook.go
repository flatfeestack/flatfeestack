package api

import (
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func AnalysisEngineHook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data WebhookCallback
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Errorf("Could not decode Webhook body: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	reqId, err := uuid.Parse(data.RequestId)
	if err != nil {
		log.Errorf("cannot parse request id: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rowsAffected := 0
	for _, v := range data.Result {
		err = db.InsertAnalysisResponse(reqId, v.Email, v.Names, v.Weight, utils.TimeNow())
		if err != nil {
			log.Errorf("insert error: %v", err)
			utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
			return
		}
		rowsAffected++
	}

	errA := db.UpdateAnalysisRequest(reqId, utils.TimeNow(), data.Error)
	if errA != nil {
		log.Warnf("cannot send to analyze engine %v", errA)
	}

	log.Printf("Inserted %v contributions into DB for request %v", rowsAffected, data.RequestId)
	w.WriteHeader(http.StatusOK)
}
