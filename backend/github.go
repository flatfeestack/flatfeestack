package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

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

// @Summary Search for Repos on github
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200 {object} []RepoDTO
// @Failure 400 {object}
// @Router /api/repos/search [get]
func searchRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := r.URL.Query()
	q := v.Get("q")
	log.Printf("query %v", q)
	repos, err := fetchGithubRepoSearch(q)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
		return
	}
	// send the HttpResponse
	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode repos: %v", err)
		return
	}
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
