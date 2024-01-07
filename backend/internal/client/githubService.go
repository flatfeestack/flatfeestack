package client

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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

type GithubClient struct {
	HTTPClient *http.Client
}

func NewGithubClient() *GithubClient {
	return &GithubClient{
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (gc *GithubClient) FetchGithubRepoSearch(q string) ([]RepoSearch, error) {
	log.Print("https://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	res, err := gc.HTTPClient.Get("https://api.github.com/search/repositories?q=" + url.QueryEscape(q))
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

func (gc *GithubClient) fetchGithubRepoById(id uint32) (*RepoSearch, error) {
	res, err := gc.HTTPClient.Get("http://api.github.com/repositories/" + strconv.Itoa(int(id)))
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
