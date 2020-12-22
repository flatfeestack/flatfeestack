package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type RepoSearchResponse struct {
	TotalCount int32     `json:"total_count"`
	Items      []RepoDTO `json:"items"`
}

func FetchGithubRepoSearch(q string) ([]RepoDTO, error) {
	log.Print("http://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	res, err := http.Get("http://api.github.com/search/repositories?q=" + url.QueryEscape(q))
	if err != nil {
		log.Printf("Could not search for repos %v", err)
		return nil, err
	}
	var result RepoSearchResponse

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("cant decode json %v", err)
		return nil, err
	}
	if result.TotalCount > 0 {
		log.Printf("%v", result.Items[0])
		return result.Items, nil
	}
	return []RepoDTO{}, nil
}

func FetchGithubRepoById(id int) (*Repo, error) {
	res, err := http.Get("http://api.github.com/repositories/" + url.QueryEscape(fmt.Sprint(id)))
	if err != nil {
		log.Printf("Could not fetch for repo details %v", err)
		return nil, err
	}

	var result Repo
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("cant decode json %v", err)
		return nil, err
	}
	return &result, nil
}

type SearchRepoResponse struct {
	HttpResponse
	Data []RepoDTO `json:"data,omitempty"`
}

// @Summary Search for Repos on github
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200 {object} SearchRepoResponse
// @Failure 404 {object} HttpResponse
// @Router /api/repos/search [get]
func SearchRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := r.URL.Query()
	q := v.Get("q")
	log.Printf("query %v", q)
	repos, err := FetchGithubRepoSearch(q)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Could not fetch repos")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	res := HttpResponse{
		Success: true,
		Data:    repos,
		Message: "",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}
