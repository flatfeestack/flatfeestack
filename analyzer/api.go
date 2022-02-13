package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type FlatFeeWeight struct {
	Names  []string `json:"names"`
	Email  string   `json:"email"`
	Weight float64  `json:"weight"`
}

type WebhookRequest struct {
	RequestId uuid.UUID `json:"reqId"`
	DateFrom  time.Time `json:"dateFrom"`
	DateTo    time.Time `json:"dateTo"`
	GitUrl    string    `json:"gitUrl"`
	Branch    string    `json:"branch"`
}

type WebhookResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

type WebhookCallback struct {
	RequestId uuid.UUID       `json:"request_id"`
	Success   bool            `json:"success"`
	Error     string          `json:"error"`
	Result    []FlatFeeWeight `json:"result"`
}

func analyzeRepository(w http.ResponseWriter, r *http.Request) {
	var request WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
	}

	log.Debugf("analyze repo: %v", request)

	if len(request.GitUrl) == 0 {
		makeHttpStatusErr(w, "no required repository_url provided", http.StatusBadRequest)
	}

	var branch string
	if len(request.Branch) > 0 {
		branch = request.Branch
	} else {
		branch = opts.GitDefaultBranch
	}

	err = json.NewEncoder(w).Encode(WebhookResponse{RequestId: request.RequestId})
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
	}

	go analyzeForWebhookInBackground(request, branch)
	log.Debugf("is analyzing")

}

func analyzeForWebhookInBackground(request WebhookRequest, branch string) {
	log.Debugf("\n\n---> webhook request for repository %s on branch %s \n", request.GitUrl, branch)
	log.Debugf("Request id: %s\n", request.RequestId)

	contributionMap, err := analyzeRepositoryFromString(request.GitUrl, request.DateFrom, request.DateTo, branch)
	if err != nil {
		callbackToWebhook(WebhookCallback{RequestId: request.RequestId, Success: false, Error: err.Error()})
		return
	}

	weightsMap, err := weightContributions(contributionMap)
	if err != nil {
		callbackToWebhook(WebhookCallback{RequestId: request.RequestId, Success: false, Error: err.Error()})
		return
	}

	var contributionWeights []FlatFeeWeight
	for _, v := range weightsMap {
		contributionWeights = append(contributionWeights, v)
	}

	callbackToWebhook(WebhookCallback{
		RequestId: request.RequestId,
		Success:   true,
		Result:    contributionWeights,
	})

	log.Debugf("Finished request %s\n", request.RequestId)
}

// getRepositoryFromRequest extracts the repository from the route parameters
func getRepositoryFromRequest(r *http.Request) (string, error) {
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		return "", errors.New("repository not found")
	}
	return repositoryUrl[0], nil
}

// getTimeRange returns the time range in the format since, until, error from the request with time in rfc3339 format
func getTimeRange(r *http.Request) (time.Time, time.Time, error) {
	var err error

	// convert since RFC3339 into golang time
	commitsSinceString := r.URL.Query()["since"]
	var commitsSince time.Time
	if len(commitsSinceString) > 0 {
		commitsSince, err = convertTimestampStringToTime(commitsSinceString[0])
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// convert until RFC3339 into golang time
	commitsUntilString := r.URL.Query()["until"]
	var commitsUntil time.Time
	if len(commitsUntilString) > 0 {
		commitsUntil, err = convertTimestampStringToTime(commitsUntilString[0])
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return commitsSince, commitsUntil, nil
}

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	w.WriteHeader(httpStatusError)
}

func callbackToWebhook(body WebhookCallback) {
	if body.Success {
		log.Printf("About to returt the following data: %v", body.Result)
	}

	reqBody, _ := json.Marshal(body)
	log.Printf("Call to %s with success %v", os.Getenv("WEBHOOK_CALLBACK_URL"), body.Success)

	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("POST", opts.CallbackUrl, bytes.NewBuffer(reqBody))
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
}

// getBranchToAnalyze extracts from the route parameters and env variables the correct branch to analyze
func getBranchToAnalyze(r *http.Request) string {
	branchUrlParam := r.URL.Query()["branch"]
	// check whether the param was set. If it was return this branch name, else return the default one
	if len(branchUrlParam) > 0 {
		return branchUrlParam[0]
	} else {
		return opts.GitDefaultBranch
	}
}

// convertTimestampStringToTime is a timestamp converter to the time interpretation of go
func convertTimestampStringToTime(rfc3339time string) (time.Time, error) {
	commitsSinceTime, err := time.Parse(time.RFC3339, rfc3339time)
	if err != nil {
		return time.Time{}, err
	}
	return commitsSinceTime, nil
}
