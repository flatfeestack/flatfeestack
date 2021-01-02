package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v72"
	"log"
	"net/http"
	"strconv"
	"time"
)

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
	Data UserDTO `json:"data,omitempty"`
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
func CreateUser(w http.ResponseWriter, r *http.Request) {
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
	user.Id = uid
	// call insert user function and pass the user
	dbErr := SaveUser(&user)

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
	Data UserDTO `json:"data,omitempty"`
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
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	if !IsValidUUID(id) {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Not a valid user id")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(403)
		return
	}
	user, err := FindUserByID(uid)
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
func GetMyUser(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(404)
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

type GetMyConnectedEmailsResponse struct {
	HttpResponse
	Data []string `json:"data,omitempty"`
}

// @Summary Get connected Git Email addresses
// @Description Get details of all users
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} GetMyConnectedEmailsResponse
// @Failure 404 {object} HttpResponse
// @Router /api/users/me/connectedEmails [get]
func GetMyConnectedEmails(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(404)
		res := NewHttpErrorResponse("Not a valid user")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	emails, err := FindConnectedEmails(user.Id)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Not a valid user")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	res := HttpResponse{
		Success: true,
		Data:    emails,
		Message: "User created successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type GitEmailRequest struct {
	Email string `json:"email"`
}

// @Summary Add new git email
// @Tags Users
// @Param repo body GitEmailRequest true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} CreateRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users/me/connectedEmails [post]
func AddGitEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(403)
		res := NewHttpErrorResponse("Unauthorized")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// create an empty user of type User
	var body GitEmailRequest
	jsonErr := json.NewDecoder(r.Body).Decode(&body)
	if jsonErr != nil {
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	//TODO: send email to user and add email after verification

	dbErr := InsertConnectedEmail(user.Id, body.Email)

	if dbErr != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    nil,
		Message: "Added email successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

// @Summary Delete git email
// @Tags Users
// @Param email path string true "Git email"
// @Accept  json
// @Produce  json
// @Success 200 {object} CreateRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users/me/connectedEmails [delete]
func RemoveGitEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(403)
		res := NewHttpErrorResponse("Unauthorized")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	params := mux.Vars(r)
	email := params["email"]
	if err != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Invalid email")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	//TODO: send email to user and add email after verification

	dbErr := DeleteConnectedEmail(user.Id, email)

	if dbErr != nil {
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    nil,
		Message: "Removed email successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type PayoutAddressRequest struct {
	ChainId string `json:"chain_id"`
	Address string `json:"address"`
}

// @Summary update payout address
// @Tags Users
// @Param user body PayoutAddressRequest true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} interface{}
// @Failure 404 {object} HttpResponse
// @Router /api/users/me/payout [put]
func PutPayoutAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	var body PayoutAddressRequest
	jsonErr := json.NewDecoder(r.Body).Decode(&body)
	if jsonErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not decode body")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	address := PayoutAddress{
		Uid:     user.Id,
		ChainId: body.ChainId,
		Address: body.Address,
	}

	dbErr := InsertOrUpdatePayoutAddress(address)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not update in database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	res := HttpResponse{
		Success: true,
		Message: "Updated exchange successfully",
	}
	_ = json.NewEncoder(w).Encode(res)
}

// @Summary get payout addresses
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} interface{}
// @Failure 404 {object} HttpResponse
// @Router /api/users/me/payout [get]
func GetPayoutAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	addresses, dbErr := SelectPayoutAddressesByUid(user.Id)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not query database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	res := HttpResponse{
		Success: true,
		Data:    addresses,
		Message: "Updated exchange successfully",
	}
	_ = json.NewEncoder(w).Encode(res)
}

/*
 *	==== REPO ====
 */
type CreateRepoResponse struct {
	HttpResponse
	Data RepoDTO `json:"data,omitempty"`
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
func CreateRepo(w http.ResponseWriter, r *http.Request) {
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
	// call insert user function and pass the user
	dbErr := SaveRepo(&repo)

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
	Data RepoDTO `json:"data,omitempty"`
}

// @Summary Get Repo By ID
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetRepoByIDResponse
// @Failure 404 {object} HttpResponse
// @Router /api/repos/{id} [get]
func GetRepoByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Invalid repoId")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	repo, err := FindRepoByID(id)
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
	Data RepoDTO `json:"data,omitempty"`
}

// @Summary Sponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} SponsorRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/repos/{id}/sponsor [post]
func SponsorRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	params := mux.Vars(r)
	repoId, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Invalid repoId")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	repo, e := FindRepoByID(repoId)
	if e != nil {
		// If repo does not exists, create the repo
		repo, err = FetchGithubRepoById(repoId)
		if err != nil {
			w.WriteHeader(400)
			res := NewHttpErrorResponse("Could not find matching Github Repo")
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		err = SaveRepo(repo)
		if err != nil {
			w.WriteHeader(500)
			res := NewHttpErrorResponse("Could not save repo")
			_ = json.NewEncoder(w).Encode(res)
			return
		}
		// Trigger first analysis of the repo

		webhookRequest := WebhookRequest{
			RepositoryUrl:       repo.Url,
			Since:               time.Now().AddDate(0, -3, 0).Format(time.RFC3339),
			Until:               time.Now().Format(time.RFC3339),
			PlatformInformation: false,
			Branch:              "master",
		}
		body, jsonErr := json.Marshal(webhookRequest)
		if jsonErr != nil {
			log.Printf("Could not marshal json body: %v", jsonErr)
		}

		res, err := http.Post("http://analysis-engine:8080/webhook", "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Could not send POST request to analyze repository %v", err)
		} else {
			var webhookResponse WebhookResponse
			err = json.NewDecoder(res.Body).Decode(&webhookResponse)
			// insert request ID in DB
			sqlInsert := `INSERT INTO "analysis_request" ("id","from", "to", "repo_id", "branch") VALUES ($1, $2, $3, $4, $5)`
			_, insertErr := db.Exec(sqlInsert, webhookResponse.RequestId, webhookRequest.Since, webhookRequest.Until, repo.ID, webhookRequest.Branch)
			if insertErr != nil {
				log.Printf("Could not insert Analysis Request into DB: %v", insertErr)
			}
			log.Printf("Requested analysis for repo %v (requestID:%v)", repo.ID, webhookResponse.RequestId)
		}
	}

	_, dbErr := Sponsor(repoId, user.Id)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not write to database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    repo,
		Message: "Sponsored repo successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

type UnsponsorRepoResponse struct {
	HttpResponse
	Data SponsorEvent `json:"data,omitempty"`
}

// @Summary Unsponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} UnsponsorRepoResponse
// @Failure 400 {object} HttpResponse
// @Router /api/repos/{id}/unsponsor [post]
func UnsponsorRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	params := mux.Vars(r)
	repoId, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Invalid repoId")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	event, dbErr := Unsponsor(repoId, user.Id)

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
	Data []RepoDTO `json:"data,omitempty"`
}

// @Summary List sponsored Repos of a user
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} GetSponsoredReposResponse
// @Failure 400 {object} HttpResponse
// @Router /api/users/sponsored [get]
func GetSponsoredRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}
	repos, dbErr := GetSponsoredReposById(user.Id)

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

/*
	==== Payment
*/

type PostSubscriptionBody struct {
	Plan          string `json:"plan"`
	PaymentMethod string `json:"paymentMethod"`
}

type PostSubscriptionResponse struct {
	HttpResponse
	Data stripe.Subscription `json:"data,omitempty"`
}

// @Summary Create a subscription
// @Tags Repos
// @Param body body PostSubscriptionBody true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} interface{}
// @Failure 400 {object} HttpResponse
// @Router /api/payments/subscriptions [post]
func PostSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	user, err := getUserFromContext(r)
	if err != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Internal server error")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	var body PostSubscriptionBody
	jsonErr := json.NewDecoder(r.Body).Decode(&body)
	if jsonErr != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Unable to decode the request body.")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	s, err := CreateSubscription(*user, body.Plan, body.PaymentMethod)
	if err != nil {
		// Make error more specifix
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Something went wrong")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    s,
		Message: "Created subscription",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

/*
	==== Exchange ====
*/

type GetExchangesResponse struct {
	HttpResponse
	Data []ExchangeEntry `json:"data,omitempty"`
}

// @Summary Get exchange requests
// @Tags Repos
// @Param q query string true "Search String"
// @Accept  json
// @Produce  json
// @Success 200 {object} interface{}
// @Failure 404 {object} HttpResponse
// @Router /api/exchanges [get]
func GetExchanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	exchanges, dbErr := SelectExchanges()

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not read from database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Data:    exchanges,
		Message: "Retrieved exchanges successfully",
	}
	// send the HttpResponse
	_ = json.NewEncoder(w).Encode(res)
}

// @Summary update exchange with price and date
// @Tags Repos
// @Param user body ExchangeEntryUpdate true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200 {object} interface{}
// @Failure 404 {object} HttpResponse
// @Router /api/exchanges/{id} [put]
func PutExchange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	exchangeID, err := strconv.Atoi(params["id"])
	if err != nil {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("Invalid exchangeID")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	var body ExchangeEntryUpdate

	jsonErr := json.NewDecoder(r.Body).Decode(&body)
	if jsonErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not decode body")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	if exchangeID != body.ID {
		w.WriteHeader(400)
		res := NewHttpErrorResponse("ID must be identical in body and url param")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	dbErr := UpdateExchange(body)

	if dbErr != nil {
		w.WriteHeader(500)
		res := NewHttpErrorResponse("Could not update in database")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// format a HttpResponse object
	res := HttpResponse{
		Success: true,
		Message: "Updated exchange successfully",
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

func getUserFromContext(r *http.Request) (user *User, err error) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(*User)
	if !ok {
		fmt.Printf("Could not get user from token %v", ok)
		return user, errors.New("could not get user")
	}
	return user, nil
}
