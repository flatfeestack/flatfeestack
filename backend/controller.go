package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

type GitEmailRequest struct {
	Email string `json:"email"`
}

type WebhookCallback struct {
	RequestId string          `json:"request_id"`
	Success   bool            `json:"success"`
	Error     string          `json:"error"`
	Result    []FlatFeeWeight `json:"result"`
}

type FlatFeeWeight struct {
	Contributor Contributor `json:"contributor"`
	Weight      float64     `json:"weight"`
}

type Contributor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

/*
 *	==== USER ====
 */


// @Summary Get User by sub in token
// @Description Get details of all users
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} User
// @Failure 403
// @Router /backend/users/me [get]
func getMyUser(w http.ResponseWriter, _ *http.Request, user *User) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

// @Summary Get connected Git Email addresses
// @Description Get details of all users
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} []string
// @Failure 403
// @Failure 500
// @Router /backend/users/me/connectedEmails [get]
func getMyConnectedEmails(w http.ResponseWriter, _ *http.Request, user *User) {
	emails, err := findGitEmails(user.Id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not find git emails %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(emails)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

// @Summary Add new git email
// @Tags Users
// @Param repo body GitEmailRequest true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} GitEmailRequest
// @Failure 403
// @Failure 400
// @Router /backend/users/me/connectedEmails [post]
func addGitEmail(w http.ResponseWriter, r *http.Request, user *User) {
	var body GitEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	//TODO: send email to user and add email after verification
	err = saveGitEmail(uuid.New(), user.Id, body.Email)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save email: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary Delete git email
// @Tags Users
// @Param email path string true "Git email"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/users/me/connectedEmails [delete]
func removeGitEmail(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	email := params["email"]

	//TODO: send email to user and add email after verification
	err := deleteGitEmail(user.Id, email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid email: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updatePayout(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["address"]
	user.PayoutETH = &a
	err := updateUser(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save payout address: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary List sponsored Repos of a user
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/users/sponsored [get]
func getSponsoredRepos(w http.ResponseWriter, r *http.Request, user *User) {
	repos, err := getSponsoredReposById(user.Id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not get repos: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

/*
 *	==== Repo ====
 */


// @Summary Search for Repos on github
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/search [get]
func searchRepoGitHub(w http.ResponseWriter, r *http.Request, _ *User) {
	params := mux.Vars(r)
	q := params["query"]
	log.Printf("query %v", q)
	repos, err := fetchGithubRepoSearch(q)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

// @Summary Get Repo By ID
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 404
// @Router /backend/repos/{id} [get]
func getRepoByID(w http.ResponseWriter, r *http.Request, _ *User) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}

	repo, err := findRepoByID(id)
	if repo == nil {
		writeErr(w, http.StatusNotFound, "Could not find repo with id %v", id)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not fetch DB %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repo)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

/*
 *	==== SPONSOR EVENT ====
 */

func sponsorRepoGitHub(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	s := params["id"]
	id, err := strconv.Atoi(s)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}

	repoDto, err := fetchGithubRepoById(uint32(id))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Repo not found %v", err)
		return
	}

	repo := Repo{
		Id:          uuid.New(),
		OrigId:      uint32(id),
		Url:         &repoDto.Url,
		Name:        &repoDto.Name,
		Description: &repoDto.Description,
	}

	var repoId *uuid.UUID
	repoId, err = saveRepo(&repo)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could store in DB: %v", err)
		return
	}
	sponsorRepo0(w, user, *repoId, SPONSOR)
}

// @Summary Sponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/{id}/sponsor [post]
func sponsorRepo(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	sponsorRepo0(w, user, id, SPONSOR)
}

// @Summary Unsponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/{id}/unsponsor [post]
func unsponsorRepo(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	sponsorRepo0(w, user, repoId, UNSPONSOR)
}
func sponsorRepo0(w http.ResponseWriter, user *User, repoId uuid.UUID, newEventType uint8) {
	now := time.Now()
	event := SponsorEvent{
		Id:          uuid.New(),
		Uid:         user.Id,
		RepoId:      repoId,
		EventType:   newEventType,
		SponsorAt:   now,
		UnsponsorAt: now,
	}
	err := sponsor(&event)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save to DB %v", err)
		return
	}

	//no need for transaction here, repoId is very static
	log.Printf("repoId %v", repoId)
	var repo *Repo
	repo, err = findRepoByID(repoId)
	if repo == nil {
		writeErr(w, http.StatusNotFound, "Could not find repo with id %v", repoId)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not fetch DB %v", err)
		return
	}

	err = analysisRequest(repo.Id, *repo.Url)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not submit analysis request %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repo)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

// @Summary Get exchange requests
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 404
// @Router /backend/exchanges [get]
func getExchanges(w http.ResponseWriter, _ *http.Request, _ *User) {
	price, err := getUpdateExchanges("ETH")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not get exchange Rate %v", err)
		return
	}

	e := ExchangeRate{}
	e.Ethereum.Usd=price
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(e)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

func analysisEngineHook(w http.ResponseWriter, r *http.Request, email string) {
	w.Header().Set("Content-Type", "application/json")
	var data WebhookCallback
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	rid, err := uuid.Parse(data.RequestId)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "cannot parse request id: %v", err)
		return
	}
	rowsAffected := 0
	for _, wh := range data.Result {
		err = saveAnalysisResponse(rid, &wh)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "insert error: %v", err)
			return
		}
		rowsAffected ++
	}
	log.Printf("Inserted %v contributions into DB for request %v", rowsAffected, data.RequestId)
	w.WriteHeader(http.StatusOK)
}
