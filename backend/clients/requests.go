package clients

import (
	"backend/api"
	db "backend/db"
	"backend/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
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
	payoutUrl   string
	serverKey   string
	analysisUrl string
)

func Init(payoutUrl0 string, serverKey0 string, analysisUrl0 string) {
	payoutUrl = payoutUrl0
	serverKey = serverKey0
	analysisUrl = analysisUrl0
}

func PayoutRequest(userId uuid.UUID, amount *big.Int) (*api.PayoutResponse, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	preq := api.PayoutRequest2{
		userId,
		amount,
	}

	body, err := json.Marshal(preq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, payoutUrl+"/admin/sign", bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+serverKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var presp api.PayoutResponse
	err = json.NewDecoder(resp.Body).Decode(&presp)
	if err != nil {
		return nil, err
	}
	return &presp, nil
}

func AnalysisReq(repoId uuid.UUID, repoUrl string) error {
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
	req.Header.Add("Authorization", "Bearer "+serverKey)
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

/*
 *	==== CoinGecko ====
 */
type ExchangeRate struct {
	Ethereum struct {
		Usd decimal.Decimal `json:"usd"`
	} `json:"ethereum"`
}

// https://www.coingecko.com/en/api
func getPriceETH() (decimal.Decimal, error) {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", "https://api.coingecko.com/backend/v3/simple/price?ids=ethereum&vs_currencies=usd", nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("accept", "application/json")
	//curl -X GET "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd" -H  "accept: application/json"
	r, err := client.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer r.Body.Close()
	rate := ExchangeRate{}
	err = json.NewDecoder(r.Body).Decode(&rate)
	if err != nil {
		return decimal.Zero, err
	}
	return rate.Ethereum.Usd, nil
}

/*
 *	==== GitHub ====
 */
type RepoSearchResponse struct {
	TotalCount uint32       `json:"total_count"`
	Items      []RepoSearch `json:"items"`
}

type RepoSearch struct {
	Id          uint64      `json:"id,omitempty"`
	Url         string      `json:"html_url,omitempty"`
	GitUrl      string      `json:"clone_url,omitempty"`
	Name        string      `json:"full_name,omitempty"`
	Description string      `json:"description,omitempty"`
	Score       json.Number `json:"score,omitempty"`
}

func FetchGithubRepoSearch(q string) ([]RepoSearch, error) {
	log.Print("https://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	res, err := http.Get("https://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	if err != nil {
		log.Printf("Could not search for repos %v", err)
		return nil, err
	}
	var result RepoSearchResponse

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.TotalCount > 0 {
		return result.Items, nil
	}
	return []RepoSearch{}, nil
}

func fetchGithubRepoById(id uint32) (*RepoSearch, error) {
	res, err := http.Get("http://api.github.com/repositories/" + strconv.Itoa(int(id)))
	if err != nil {
		log.Printf("Could not fetch for repo details %v", err)
		return nil, err
	}

	var result RepoSearch
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("cant decode json %v", err)
		return nil, err
	}
	return &result, nil
}
