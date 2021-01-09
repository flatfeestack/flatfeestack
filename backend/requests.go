package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*
 *	==== Analysis Request  ====
 */
type AnalysisRequest struct {
	RepositoryUrl       string    `json:"repository_url"`
	DateFrom            time.Time `json:"since"`
	DateTo              time.Time `json:"until"`
	PlatformInformation bool      `json:"platform_information"`
	Branch              string    `json:"branch"`
}

type AnalysisResponse struct {
	RequestId uuid.UUID `json:"request_id"`
}

func analysisRequest(repoId uuid.UUID, repoUrl string) error {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	now := time.Now()
	req := AnalysisRequest{
		RepositoryUrl:       repoUrl,
		DateFrom:            now.AddDate(0, -3, 0),
		DateTo:              now,
		PlatformInformation: false,
		Branch:              "master",
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := client.Post(opts.AnalysisUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var resp AnalysisResponse
	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return err
	}

	return saveAnalysisRequest(resp.RequestId, repoId, req.DateFrom, req.DateTo, req.Branch)
}

/*
 *	==== CoinGecko ====
 */
type ExchangeRate struct {
	Ethereum struct {
		Usd decimal.Decimal `json:"usd"`
	} `json:"ethereum"`
}

//https://www.coingecko.com/en/api
func getPriceETH() (decimal.Decimal, error) {
	//https://stackoverflow.com/questions/16895294/how-to-set-timeout-for-http-get-requests-in-golang
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	//curl -X GET "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd" -H  "accept: application/json"
	r, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
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
	TotalCount uint32    `json:"total_count"`
	Items      []RepoDTO `json:"items"`
}

type RepoDTO struct {
	ID          uint32 `json:"id"`
	Url         string `json:"html_url"`
	Name        string `json:"full_name"`
	Description string `json:"description"`
}

func fetchGithubRepoSearch(q string) ([]RepoDTO, error) {
	log.Print("http://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	res, err := http.Get("http://api.github.com/search/repositories?q=" + url.QueryEscape(q))
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
	return []RepoDTO{}, nil
}

func fetchGithubRepoById(id uint32) (*RepoDTO, error) {
	res, err := http.Get("http://api.github.com/repositories/" + strconv.Itoa(int(id)))
	if err != nil {
		log.Printf("Could not fetch for repo details %v", err)
		return nil, err
	}

	var result RepoDTO
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("cant decode json %v", err)
		return nil, err
	}
	return &result, nil
}
