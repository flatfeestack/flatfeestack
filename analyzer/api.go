package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	RequestId     uuid.UUID          `json:"requestId"`
	Error         string             `json:"error,omitempty"`
	Result        []FlatFeeWeight    `json:"result"`
	ContribCommit ContribCommitCount `json:contribcommit`
}

type FlatFeeWeight struct {
	Names  []string `json:"names"`
	Email  string   `json:"email"`
	Weight float64  `json:"weight"`
}

func analyze(w http.ResponseWriter, r *http.Request) {
	var request AnalysisRequest
	fmt.Println("---------------------------")
	fmt.Println(r.Body)
	fmt.Println("---------------------------")
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
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "analyzeRepositoryFromString: " + err.Error()}, opts.BackendCallbackUrl)
		return
	}

	trustValueContributersCommits, err := getTotalContributersCommits(contributionMap)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "getTotalContributersCommits: " + err.Error()}, opts.BackendCallbackUrl)
		return
	}

	weightsMap, err := weightContributions(contributionMap)
	if err != nil {
		callbackToWebhook(AnalysisCallback{RequestId: request.Id, Error: "weightContributions: " + err.Error()}, opts.BackendCallbackUrl)
		return
	}

	callbackToWebhook(AnalysisCallback{
		RequestId:     request.Id,
		Result:        weightsMap,
		ContribCommit: *trustValueContributersCommits,
	}, opts.BackendCallbackUrl)

	log.Debugf("Finished request %s\n", request.Id)
}

func getTotalContributersCommits(contributionMap map[string]Contribution) (*ContribCommitCount, error) {
	return nil, fmt.Errorf("unimplemented")
}

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	log.Error(errString)
	w.WriteHeader(httpStatusError)
}

func callbackToWebhook(body AnalysisCallback, url string) {
	if body.Error == "" {
		log.Printf("About to return the following data: %v", body.Result)
	}

	reqBody, _ := json.Marshal(body)
	log.Printf("Call to %s with success %v", opts.BackendCallbackUrl, body.Error == "")

	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Could not create a HTTP request to call the webhook %v", err)
		return
	}

	auth := opts.BackendUsername + ":" + opts.BackendPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		log.Printf("Could not call webhook %v", err)
		return
	}

	defer resp.Body.Close()
	log.Debugf("return value: %v", resp.StatusCode)
}
