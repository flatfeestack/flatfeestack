package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)
type RepoSearchResponse struct{
	TotalCount int8 `json:"total_count"`
	Items []Repo `json:"items"`
}
func FetchGithubRepoSearch(q string) ([]Repo, error){
	log.Print("http://api.github.com/search/repositories?q="+url.QueryEscape(q))
	res, err := http.Get("http://api.github.com/search/repositories?q="+url.QueryEscape(q))
	if err != nil {
		log.Printf("Could not search for repos %v",err)
		return nil,err
	}
	var result RepoSearchResponse

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("cant decode json %v",err)
		return nil, err
	}
	if result.TotalCount > 0 {
		log.Printf("%v", result.Items[0])
		return result.Items,nil
	}
	return []Repo{}, nil
}



// @Summary Search for Repos on github
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200 {object} []Repo
// @Failure 404 {object} HttpResponse
// @Router /api/repos/search [get]
func SearchRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := r.URL.Query()
	q := v.Get("q")
	log.Printf("query %v",q)
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
