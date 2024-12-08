package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type AnalysisRequest struct {
	Id         uuid.UUID
	RepoId     uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
	GitUrl     string
	ReceivedAt *time.Time
	Error      *string
}

type AnalysisResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

type AnalysisCallback struct {
	RequestId uuid.UUID       `json:"requestId"`
	Error     string          `json:"error,omitempty"`
	Result    []FlatFeeWeight `json:"result"`
	RepoId    uuid.UUID       `json:"repoid"`
}

type FlatFeeWeight struct {
	Names       []string `json:"names"`
	Email       string   `json:"email"`
	Weight      float64  `json:"weight"`
	CommitCount int      `json:"commitcount"`
}

func analyze(w http.ResponseWriter, r *http.Request) {
	var request AnalysisRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
	}

	slog.Info("analyze repo: %v", request)

	err = json.NewEncoder(w).Encode(AnalysisResponse{RequestId: request.Id})
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
	}

	go analyzeBackground(request)
	slog.Debug("is analyzing")
}

func analyzeBackground(request AnalysisRequest) {
	slog.Debug("---> webhook request for repository %s\n", request.GitUrl)
	slog.Debug("Request id: %s\n", request.Id)

	contributionMap, err := analyzeRepository(request.DateFrom, request.DateTo, request.GitUrl)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "analyzeRepositoryFromString: " + err.Error()}, cfg.BackendCallbackUrl)
		return
	}

	weightsMap, err := weightContributions(contributionMap)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "weightContributions: " + err.Error()}, cfg.BackendCallbackUrl)
		return
	}

	callbackToWebhook(AnalysisCallback{
		RequestId: request.Id,
		Result:    weightsMap,
		RepoId:    request.RepoId,
	}, cfg.BackendCallbackUrl)

	slog.Debug("Finished request %s\n", request.Id)
}

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	slog.Error(errString)
	w.WriteHeader(httpStatusError)
}

func callbackToWebhook(body AnalysisCallback, url string) {
	if body.Error == "" {
		log.Printf("About to return the following data: %v", body.Result)
	}

	reqBody, _ := json.Marshal(body)
	log.Printf("Call to %s with success %v", cfg.BackendCallbackUrl, body.Error == "")

	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Could not create a HTTP request to call the webhook %v", err)
		return
	}

	auth := cfg.BackendUsername + ":" + cfg.BackendPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		log.Printf("Could not call webhook %v", err)
		return
	}

	defer resp.Body.Close()
	slog.Debug("return value: %v", resp.StatusCode)
}
