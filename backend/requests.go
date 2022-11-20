package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*
 *	==== Analysis Request  ====
 */
type AnalysisRequest struct {
	Id         uuid.UUID
	RequestId  uuid.UUID `json:"reqId"`
	RepoId     uuid.UUID
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GitUrls    []string  `json:"gitUrls"`
	ReceivedAt *time.Time
	Error      *string
}

type AnalysisWebhookResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

type Payout struct {
	Address          string       `json:"address"`
	NanoTea          int64        `json:"nano_tea"`
	SmartContractTea big.Int      `json:"smart_contract_tea"`
	Meta             []PayoutMeta `json:"meta"`
}

type PayoutResponse struct {
	TxHash   string   `json:"tx_hash"`
	Currency string   `json:"currency"`
	Payout   []Payout `json:"payout_cryptos"`
}

func payoutRequest(pts []PayoutToService, currency string) (*PayoutResponse, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	body, err := json.Marshal(pts)
	if err != nil {
		return nil, err
	}

	r, err := client.Post(opts.PayoutUrl+"/pay-crypto/"+string(bytes.ToLower([]byte(currency))), "application/json", bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var resp PayoutResponse
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func analysisRequest(repoId uuid.UUID, repoUrls []string) error {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	now := timeNow()
	req := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    repoId,
		DateFrom:  now.AddDate(0, -3, 0),
		DateTo:    now,
		GitUrls:   repoUrls,
	}

	err := insertAnalysisRequest(req, now)
	if err != nil {
		return err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := client.Post(opts.AnalysisUrl+"/analyze", "application/json", bytes.NewBuffer(body))
	if err != nil {
		e := err.Error()
		errA := updateAnalysisRequest(req.RequestId, now, &e)
		if errA != nil {
			log.Warnf("cannot send to analyze engine %v", errA)
		}
		return err
	}
	defer r.Body.Close()

	//just make sure we got the response
	var awr AnalysisWebhookResponse
	err = json.NewDecoder(r.Body).Decode(&awr)
	if err != nil {
		e := err.Error()
		errA := updateAnalysisRequest(req.RequestId, timeNow(), &e)
		if errA != nil {
			log.Warnf("cannot send to analyze engine %v", errA)
		}
		log.Warnf("cannot send to analyze engine %v", err)
	}
	if awr.RequestId != req.RequestId {
		log.Errorf("We have a serious problem, request id does not match %v != %v", awr.RequestId, req.RequestId)
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

func fetchGithubRepoSearch(q string) ([]RepoSearch, error) {
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
