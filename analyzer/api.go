package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
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
}

type FlatFeeWeight struct {
	Names  []string `json:"names"`
	Email  string   `json:"email"`
	Weight float64  `json:"weight"`
}

func analyze(w http.ResponseWriter, r *http.Request) {
	var request AnalysisRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
	}

	log.Infof("analyze repo: %v", request)

	err = json.NewEncoder(w).Encode(AnalysisResponse{RequestId: request.Id})
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
	}

	go analyzeBackground(request)
	log.Debugf("is analyzing")

}

func analyzeBackground(request AnalysisRequest) {
	log.Debugf("---> webhook request for repository %s\n", request.GitUrl)
	log.Debugf("Request id: %s\n", request.Id)

	contributionMap, err := analyzeRepository(request.DateFrom, request.DateTo, request.GitUrl)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "analyzeRepositoryFromString: " + err.Error()}, opts.CallbackUrl)
		return
	}

	weightsMap, err := weightContributions(contributionMap)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "weightContributions: " + err.Error()}, opts.CallbackUrl)
		return
	}

	callbackToWebhook(AnalysisCallback{
		RequestId: request.Id,
		Result:    weightsMap,
	}, opts.CallbackUrl)

	log.Debugf("Finished request %s\n", request.Id)
}

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	w.WriteHeader(httpStatusError)
}

func callbackToWebhook(body AnalysisCallback, url string) {
	if body.Error == "" {
		log.Printf("About to returt the following data: %v", body.Result)
	}

	reqBody, _ := json.Marshal(body)
	log.Printf("Call to %s with success %v", os.Getenv("WEBHOOK_CALLBACK_URL"), body.Error == "")

	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Could not create a HTTP request to call the webhook %v", err)
		return
	}

	if len(opts.BackendToken) > 0 {
		req.Header.Add("Authorization", "Bearer "+opts.BackendToken)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		log.Printf("Could not call webhook %v", err)
		return
	}

	defer resp.Body.Close()
	log.Debugf("return value: %v", resp.StatusCode)
}

func writeErr(w http.ResponseWriter, code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Warnf(msg)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	if debug {
		w.Write([]byte(`{"error":"` + msg + `"}`))
	}
}
