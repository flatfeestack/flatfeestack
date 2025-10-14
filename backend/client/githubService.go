package client

import (
	"encoding/json"
	"log/slog"
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
	urlEsc := "https://api.github.com/search/repositories?q=" + url.QueryEscape(q)
	slog.Info(urlEsc)
	res, err := gc.HTTPClient.Get(urlEsc)
	if err != nil {
		slog.Error("Could not search for repos",
			slog.Any("error", err))
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
	urlEsc := "http://api.github.com/repositories/" + strconv.Itoa(int(id))
	res, err := gc.HTTPClient.Get(urlEsc)
	if err != nil {
		slog.Error("Could not fetch for repo details",
			slog.Any("error", err))
		return nil, err
	}

	var result RepoSearch
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		slog.Error("cant decode json",
			slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}
