package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type BaseHandler struct {
	userRepo             UserRepository
	repoRepo             RepoRepository
	sponsorEventRepo     SponsorEventRepository
	dailyRepoBalanceRepo DailyRepoBalanceRepository
}

// NewBaseHandler returns a new BaseHandler
func NewBaseHandler(userRepo UserRepository,
	repoRepo RepoRepository,
	sponsorEventRepo SponsorEventRepository,
	dailyRepoBalanceRepo DailyRepoBalanceRepository) *BaseHandler {
	return &BaseHandler{
		userRepo:             userRepo,
		repoRepo:             repoRepo,
		sponsorEventRepo:     sponsorEventRepo,
		dailyRepoBalanceRepo: dailyRepoBalanceRepo,
	}
}

// HttpResponse format
type HttpResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewHttpErrorResponse(message string) HttpResponse {
	return HttpResponse{Success: false, Message: message}
}

/*
 *	==== USER ====
 */

type CreateUserResponse struct {
	HttpResponse
	Data User `json:"data,omitempty"`
}
type CreateUserDTO struct {
	Email    string `json:"email" example:"info@flatfeestack"`
	Username string `json:"username" example:"flatfee"`
}

// @Summary Create new user
// @Tags Users
// @Param user body CreateUserDTO true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} CreateUserResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users [post]
func (h *BaseHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var user User
	jsonErr := json.NewDecoder(r.Body).Decode(&user)
	if jsonErr != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// generate uuuid
	uid, uuidErr := uuid.NewRandom()
	if uuidErr != nil {
		res := NewHttpErrorResponse("Unable to create uuid.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	user.ID = uid.String()
	// call insert user function and pass the user
	dbErr := h.userRepo.Save(&user)

	if dbErr != nil {
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type GetUserByIDResponse struct {
	HttpResponse
	Data User `json:"data,omitempty"`
}

// @Summary Get User by ID
// @Description Get details of all users
// @Tags Users
// @Param id path string true "User ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetUserByIDResponse
// @Failure 404 {object} HttpResponse
// @Router /api/users/{id} [get]
func (h *BaseHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	if !IsValidUUID(id) {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Not a valid user id")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	user, err := h.userRepo.FindByID(id)
	if err != nil {
		w.WriteHeader(404)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	res := HttpResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

// @Summary Get User by sub in token
// @Description Get details of all users
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} GetUserByIDResponse
// @Failure 404 {object} HttpResponse
// @Router /api/users/me [get]
func (h *BaseHandler) GetMyUser(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r)
	if err != nil {
		res := NewHttpErrorResponse("Not a valid user")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	res := HttpResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

/*
 *	==== REPO ====
 */
type CreateRepoResponse struct {
	HttpResponse
	Data Repo `json:"data,omitempty"`
}
type CreateRepoDTO struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

// @Summary Create new repo
// @Tags Repos
// @Param repo body CreateRepoDTO true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} CreateRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/repos [post]
func (h *BaseHandler) CreateRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var repo Repo
	jsonErr := json.NewDecoder(r.Body).Decode(&repo)
	if jsonErr != nil {
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	// generate uuuid
	uid, uuidErr := uuid.NewRandom()
	if uuidErr != nil {
		res := NewHttpErrorResponse("Unable to create uuid.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	repo.ID = uid.String()
	// call insert user function and pass the user
	dbErr := h.repoRepo.Save(&repo)

	if dbErr != nil {
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    repo,
		Message: "Repo created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type GetRepoByIDResponse struct {
	HttpResponse
	Data Repo `json:"data,omitempty"`
}

// @Summary Get Repo By ID
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetRepoByIDResponse
// @Failure 404 {object} HttpResponse
// @Router /api/repos/{id} [get]
func (h *BaseHandler) GetRepoByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	repo, err := h.repoRepo.FindByID(id)
	if err != nil {
		log.Println(err)
		//w.WriteHeader(404)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	res := HttpResponse{
		Success: true,
		Data:    repo,
		Message: "",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

/*
 *	==== SPONSOR EVENT ====
 */

type SponsorRepoDTO struct {
	Uid string `json:"uid"`
}
type SponsorRepoResponse struct {
	HttpResponse
	Data SponsorEvent `json:"data,omitempty"`
}

// @Summary Sponsor a repo
// @Tags Repos
// @Param body body SponsorRepoDTO true "Request Body"
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} SponsorRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/repos/{id}/sponsor [post]
func (h *BaseHandler) SponsorRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var dto SponsorRepoDTO
	jsonErr := json.NewDecoder(r.Body).Decode(&dto)
	params := mux.Vars(r)
	repoId := params["id"]
	if jsonErr != nil {
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	event, dbErr := h.sponsorEventRepo.Sponsor(repoId, dto.Uid)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    event,
		Message: "Sponsored repo successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

// @Summary Unsponsor a repo
// @Tags Repos
// @Param body body SponsorRepoDTO true "Request Body"
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} SponsorRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/repos/{id}/unsponsor [post]
func (h *BaseHandler) UnsponsorRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// create an empty user of type User
	var dto SponsorRepoDTO
	jsonErr := json.NewDecoder(r.Body).Decode(&dto)
	params := mux.Vars(r)
	repoId := params["id"]
	if jsonErr != nil {
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	event, dbErr := h.sponsorEventRepo.Unsponsor(repoId, dto.Uid)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    event,
		Message: "Unsponsored repo successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type GetSponsoredReposResponse struct {
	HttpResponse
	Data []Repo `json:"data,omitempty"`
}

// @Summary List sponsored Repos of a user
// @Tags Users
// @Param id path string true "User ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetSponsoredReposResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users/{id}/sponsored [get]
func (h *BaseHandler) GetSponsoredRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	uid := params["id"]
	repos, dbErr := h.sponsorEventRepo.GetSponsoredRepos(uid)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not read from database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    repos,
		Message: "Retrieved sponsored repos successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type CalculateDailyRepoBalanceByUserResponse struct {
	HttpResponse
	Data []DailyRepoBalance `json:"data,omitempty"`
}

// @Summary Calculate Daily Repo Balance for user
// @Tags Users
// @Param id path string true "User ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} CalculateDailyRepoBalanceByUserResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users/{id}/sponsored/calculateDaily [post]
func (h *BaseHandler) CalculateDailyRepoBalanceByUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)
	uid := params["id"]
	sponsoredRepos, dbErr := h.sponsorEventRepo.GetSponsoredRepos(uid)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not read from database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	repoBalance, dbErr1 := h.dailyRepoBalanceRepo.CalculateDailyByUser(uid, sponsoredRepos, 100)
	if dbErr1 != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not read from database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    repoBalance,
		Message: "Retrieved sponsored repos successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

/*
 *	==== Helper ====
 */

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func getUserFromContext(r *http.Request) (user *User,err error) {
	ctx := r.Context()
	user, ok := ctx.Value(authMiddlewareKey("user")).(*User)
	if !ok {
		return user, errors.New("could not get usre")
	}
	return user, nil
}