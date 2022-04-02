package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type WebhookRequest struct {
	RequestId uuid.UUID `json:"reqId"`
	DateFrom  time.Time `json:"dateFrom"`
	DateTo    time.Time `json:"dateTo"`
	GitUrls   []string  `json:"gitUrls"`
}

type WebhookResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

type WebhookCallback struct {
	RequestId uuid.UUID       `json:"request_id"`
	Error     string          `json:"error,omitempty"`
	Result    []FlatFeeWeight `json:"result"`
}

type FlatFeeWeight struct {
	Names  []string `json:"names"`
	Email  string   `json:"email"`
	Weight float64  `json:"weight"`
}

func analyzeRepository(w http.ResponseWriter, r *http.Request) {
	var request WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
	}

	log.Infof("analyze repo: %v", request)

	if len(request.GitUrls) == 0 {
		makeHttpStatusErr(w, "no required repository_url provided", http.StatusBadRequest)
	}

	err = json.NewEncoder(w).Encode(WebhookResponse{RequestId: request.RequestId})
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
	}

	go analyzeForWebhookInBackground(request)
	log.Debugf("is analyzing")

}

func analyzeForWebhookInBackground(request WebhookRequest) {
	log.Debugf("\n\n---> webhook request for repository %s\n", request.GitUrls)
	log.Debugf("Request id: %s\n", request.RequestId)

	contributionMap, err := analyzeRepositoryFromString(request.DateFrom, request.DateTo, request.GitUrls...)
	if err != nil {
		callbackToWebhook(WebhookCallback{RequestId: request.RequestId, Error: "analyzeRepositoryFromString: " + err.Error()})
		return
	}

	weightsMap, err := weightContributions(contributionMap)
	if err != nil {
		callbackToWebhook(WebhookCallback{RequestId: request.RequestId, Error: "weightContributions: " + err.Error()})
		return
	}

	callbackToWebhook(WebhookCallback{
		RequestId: request.RequestId,
		Result:    weightsMap,
	})

	log.Debugf("Finished request %s\n", request.RequestId)
}

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	w.WriteHeader(httpStatusError)
}

func callbackToWebhook(body WebhookCallback) {
	if body.Error == "" {
		log.Printf("About to returt the following data: %v", body.Result)
	}

	reqBody, _ := json.Marshal(body)
	log.Printf("Call to %s with success %v", os.Getenv("WEBHOOK_CALLBACK_URL"), body.Error == "")

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
