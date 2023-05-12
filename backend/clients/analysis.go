package clients

import (
	"backend/db"
	"backend/utils"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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

var (
	analysisUrl      string
	analysisPassword string
	analysisUsername string
)

func InitAnalyzer(analysisUrl0 string, analysisPassword0 string, analysisUsername0 string) {
	analysisUrl = analysisUrl0
	analysisPassword = analysisPassword0
	analysisUsername = analysisUsername0
}

func RequestAnalysis(repoId uuid.UUID, repoUrl string) error {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	now := utils.TimeNow()
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

	req, err := http.NewRequest(http.MethodPost, analysisUrl+"/analyze", bytes.NewBuffer(body))
	auth := analysisUsername + ":" + analysisPassword
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		e := err.Error()
		errA := db.UpdateAnalysisRequest(ar.Id, now, &e)
		if errA != nil {
			log.Warnf("cannot send to analyze engine %v", errA)
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the request %v received the status code %v", ar.Id, resp.StatusCode)
	}

	//just make sure we got the response
	var awr AnalysisResponse2
	err = json.NewDecoder(resp.Body).Decode(&awr)
	if err != nil {
		e := err.Error()
		errA := db.UpdateAnalysisRequest(ar.Id, utils.TimeNow(), &e)
		if errA != nil {
			log.Warnf("cannot send to analyze engine %v", errA)
		}
		return err
	}
	if awr.RequestId != ar.Id {
		return fmt.Errorf("we have a serious problem, request id does not match %v != %v", awr.RequestId, ar.Id)
	}

	return nil
}
