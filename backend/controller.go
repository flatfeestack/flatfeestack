package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

/*
 *	==== USER ====
 */

type GitEmailRequest struct {
	Email string `json:"email"`
}

// @Summary Get User by sub in token
// @Description Get details of all users
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} User
// @Failure 403
// @Router /api/users/me [get]
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
// @Router /api/users/me/connectedEmails [get]
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
// @Router /api/users/me/connectedEmails [post]
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
// @Router /api/users/me/connectedEmails [delete]
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
// @Router /api/users/sponsored [get]
func getSponsoredRepos(w http.ResponseWriter, r *http.Request, user *User) {
	repos, err := getSponsoredReposById(user.Id, SPONSOR)
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

// @Summary Get User by ID
// @Description Get details of all users
// @Tags Users
// @Param id path string true "User ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 404
// @Router /api/users/{id} [get]
func getUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	uid, err := uuid.Parse(id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid user id: %v", err)
		return
	}
	var user *User
	user, err = findUserByID(uid)
	if user == nil {
		writeErr(w, http.StatusNotFound, "User with id %v not found", uid)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not query DB: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
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
// @Router /api/repos/search [get]
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
// @Router /api/repos/{id} [get]
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
// @Router /api/repos/{id}/sponsor [post]
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
// @Router /api/repos/{id}/unsponsor [post]
func unsponsorRepo(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	sponsorRepo0(w, user, id, UNSPONSOR)
}
func sponsorRepo0(w http.ResponseWriter, user *User, id uuid.UUID, newEventType uint8) {
	eventType, err := lastEventSponsoredRepo(user.Id, id)
	if eventType == newEventType {
		writeErr(w, http.StatusNotModified, "We already have the current following event type %v", eventType)
		return
	}

	fmt.Printf("id %v", id)
	var repo *Repo
	repo, err = findRepoByID(id)
	if repo == nil {
		writeErr(w, http.StatusNotFound, "Could not find repo with id %v", id)
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

	event := SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    id,
		EventType: newEventType,
		CreatedAt: time.Now(),
	}
	err = sponsor(&event)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save to DB %v", err)
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
// @Router /api/exchanges [get]
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
