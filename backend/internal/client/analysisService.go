package client

import (
	"backend/internal/db"
	"backend/pkg/util"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type AnalysisRequest struct {
	Id         uuid.UUID `json:"reqId"`
	RepoId     uuid.UUID
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GitUrl     string    `json:"gitUrl"`
	ReceivedAt *time.Time
	Error      *string
}

type AnalysisResponse2 struct {
	RequestId uuid.UUID `json:"request_id"`
}

type AnalysisClient struct {
	HTTPClient       *http.Client
	analysisUrl      string
	analysisPassword string
	analysisUsername string
}

func NewAnalysisClient(analysisUrl string, analysisPassword string, analysisUsername string) *AnalysisClient {
	return &AnalysisClient{
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		}, analysisUrl: analysisUrl,
		analysisPassword: analysisPassword,
		analysisUsername: analysisUsername,
	}
}

func (a *AnalysisClient) RequestAnalysis(repoId uuid.UUID, repoUrl string) error {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	now := util.TimeNow()
	ar := db.AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repoId,
		DateFrom: now.AddDate(0, -3, 0),
		DateTo:   now,
		GitUrl:   repoUrl,
	}

	err := db.InsertAnalysisRequest(ar, now)
	if err != nil {
		return err
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, a.analysisUrl+"/analyze", bytes.NewBuffer(body))
	auth := a.analysisUsername + ":" + a.analysisPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		e := err.Error()
		errA := db.UpdateAnalysisRequest(ar.Id, now, &e)
		if errA != nil {
			slog.Warn("cannot send to analyze engine",
				slog.Any("error", err))
			return err
		}
		return nil
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Warn("Cannot close body",
				slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the request %v received the status code %v", ar.Id, resp.StatusCode)
	}

	//just make sure we got the response
	var awr AnalysisResponse2
	err = json.NewDecoder(resp.Body).Decode(&awr)
	if err != nil {
		e := err.Error()
		errA := db.UpdateAnalysisRequest(ar.Id, util.TimeNow(), &e)
		if errA != nil {
			slog.Warn("cannot send to analyze engine",
				slog.Any("error", err))
		}
		return err
	}
	if awr.RequestId != ar.Id {
		return fmt.Errorf("we have a serious problem, request id does not match %v != %v", awr.RequestId, ar.Id)
	}

	return nil
}
