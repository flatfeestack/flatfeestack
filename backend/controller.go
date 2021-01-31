package main

import (
	"encoding/json"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"math/big"
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
	Email  string  `json:"email"`
	Weight float64 `json:"weight"`
}

const (
	fakePubKey1  = "0x985B60456DF6db6952644Ee0C70dfa9146e4E12C"
	fakePrivKey1 = "0xc76d23e248188840aacec04183d94cde00ce1b591a2e6610b034094f7aef5ecf"
	//check with
	//curl --data '{"method":"eth_call","params":[{"to": "0x731a10897d267e19b34503ad902d0a29173ba4b1", "data":"0x70a08231000000000000000000000000005759e3FDE48688AAB1d6E7B434D46F2A9E9c50"}],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST localhost:8545
	fakePubKey2  = "0x005759e3FDE48688AAB1d6E7B434D46F2A9E9c50"
	fakePrivKey2 = "0xd8ac01d26dc438ba2ba99529ffd46fc1e5e924ade931a256a255dc36762deab0"
)

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
	emails, err := findGitEmailsByUserId(user.Id)
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
	err = insertGitEmail(uuid.New(), user.Id, body.Email, timeNow())
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
	repos, err := findSponsoredReposById(user.Id)
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
	q := r.URL.Query().Get("q")
	log.Printf("query %v", q)
	if q == "" {
		writeErr(w, http.StatusBadRequest, "Empty search")
		return
	}
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

	repo, err := findRepoById(id)
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
 *	==== Active EVENT ====
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
	repoId, err = insertOrUpdateRepo(&repo)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could store in DB: %v", err)
		return
	}
	sponsorRepo0(w, user, *repoId, Active)
}

// @Summary Sponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/{id}/insertOrUpdateSponsor [post]
func sponsorRepo(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	sponsorRepo0(w, user, id, Active)
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
	sponsorRepo0(w, user, repoId, Inactive)
}
func sponsorRepo0(w http.ResponseWriter, user *User, repoId uuid.UUID, newEventType uint8) {
	now := timeNow()
	event := SponsorEvent{
		Id:          uuid.New(),
		Uid:         user.Id,
		RepoId:      repoId,
		EventType:   newEventType,
		SponsorAt:   now,
		UnsponsorAt: now,
	}
	userErr, err := insertOrUpdateSponsor(&event)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save to DB: %v", err)
		return
	}
	if userErr != nil {
		writeErr(w, http.StatusConflict, "User error: %v", userErr)
		return
	}

	//no need for transaction here, repoId is very static
	log.Printf("repoId %v", repoId)
	var repo *Repo
	repo, err = findRepoById(repoId)
	if repo == nil {
		writeErr(w, http.StatusNotFound, "Could not find repo with id %v", repoId)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not fetch DB %v", err)
		return
	}
	// TODO: only if repo is sponsored for the first time
	if newEventType == Active {
		err = analysisRequest(repo.Id, *repo.Url)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not submit analysis request %v", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repo)
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
		err = insertAnalysisResponse(rid, &wh, timeNow())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "insert error: %v", err)
			return
		}
		rowsAffected++
	}
	log.Printf("Inserted %v contributions into DB for request %v", rowsAffected, data.RequestId)
	w.WriteHeader(http.StatusOK)
}

func getPayouts(w http.ResponseWriter, r *http.Request, email string) {
	m := mux.Vars(r)
	h := m["type"]
	if h == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	userAggBalances, err := getDailyPayouts(h)

	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(userAggBalances)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}
}

type PayoutToService struct {
	Address      string    `json:"address"`
	Balance      int64     `json:"balance_micro_USD"`
	ExchangeRate big.Float `json:"exchange_rate_USD_ETH"`
}

type UserAggBalanceExchangeRate struct {
	UserAggBalancs []UserAggBalance `json:"user_agg_balance"`
	ExchangeRate   big.Float        `json:"exchange_rate"`
}

func payout(w http.ResponseWriter, r *http.Request, email string) {
	userAggBalances, err := getDailyPayouts("pending")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could encode json: %v", err)
		return
	}

	m := mux.Vars(r)
	h := m["exchangeRate"]
	if h == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}
	e, _, err := big.ParseFloat(h, 10, 128, big.ToZero)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	var pts []PayoutToService
	batchId := uuid.New()
	for _, ub := range userAggBalances {
		//TODO: do one SQL insert instead of many small ones
		for _, mid := range ub.DailyUserPayoutIds {
			p := PayoutsRequest{
				DailyUserPayoutId: mid,
				BatchId:           batchId,
				ExchangeRate:      *e,
				CreatedAt:         timeNow(),
			}
			err = insertPayoutsRequest(&p)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "Could not send payout0: %v", err)
				return
			}
		}

		pt := PayoutToService{
			Address:      ub.PayoutEth,
			Balance:      ub.Balance,
			ExchangeRate: *e,
		}
		pts = append(pts, pt)

		if len(pts) >= 50 {
			err = payout0(pts, batchId)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "Could not send payout1: %v", err)
				return
			}

			//clear vars
			batchId = uuid.New()
			pts = []PayoutToService{}
		}
	}
	//save remaining batch
	err = payout0(pts, batchId)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not send payout2: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func payout0(pts []PayoutToService, batchId uuid.UUID) error {
	res, err := payoutRequest(pts)
	if err != nil {
		err1 := err.Error()
		err2 := insertPayoutsResponse(&PayoutsResponse{
			BatchId:   batchId,
			Error:     &err1,
			CreatedAt: timeNow(),
		})
		return fmt.Errorf("error %v/%v", err, err2)
	}
	return insertPayoutsResponse(&PayoutsResponse{
		BatchId:    batchId,
		Error:      nil,
		CreatedAt:  timeNow(),
		PayoutWeis: res.PayoutWeis,
	})
}

func serverTime(w http.ResponseWriter, r *http.Request, email string) {
	currentTime := timeNow()
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{"time":"` + currentTime.Format("2006-01-02 15:04:05") + `"}`))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}
}

func fakeUser(w http.ResponseWriter, r *http.Request, email string) {
	repo := randomdata.SillyName()
	uid1, rid1, err := fakeRepoUser(randomdata.Email(), repo, repo, fakePubKey1)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data1: %v", err)
		return
	}

	repo = randomdata.SillyName()
	uid2, rid2, err := fakeRepoUser(randomdata.Email(), repo, repo, fakePubKey2)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}

	s1 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *uid1,
		RepoId:    *rid1,
		EventType: Active,
		SponsorAt: timeNow(),
	}
	err1, err2 := insertOrUpdateSponsor(&s1)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "Could create sponsor1: %v, %v", err1, err2)
		return
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *uid1,
		RepoId:      *rid1,
		EventType:   Inactive,
		UnsponsorAt: timeNow().Add(time.Duration(24) * time.Hour),
	}
	err1, err2 = insertOrUpdateSponsor(&s2)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "Could create sponsor1: %v, %v", err1, err2)
		return
	}

	s3 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *uid2,
		RepoId:    *rid2,
		EventType: Active,
		SponsorAt: timeNow(),
	}
	err1, err2 = insertOrUpdateSponsor(&s3)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "Could create sponsor1: %v, %v", err1, err2)
		return
	}

	s4 := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *uid1,
		RepoId:    *rid2,
		EventType: Active,
		SponsorAt: timeNow(),
	}
	err1, err2 = insertOrUpdateSponsor(&s4)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "Could create sponsor1: %v, %v", err1, err2)
		return
	}

	//fake contribution
	err = insertGitEmail(uuid.New(), *uid1, "tom@tom.tom", timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}
	err = insertGitEmail(uuid.New(), *uid2, "sam@sam.sam", timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}

	err = fakeContribution(rid1, rid2)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}
}

func fakeContribution(rid1 *uuid.UUID, rid2 *uuid.UUID) error {
	monthStart := timeNow().AddDate(0, -1, 0)                      //$2
	monthStop := monthStart.AddDate(0, int(monthStart.Month()), 0) //$1

	aid1 := uuid.New()
	aid2 := uuid.New()
	err := insertAnalysisRequest(aid1, *rid1, monthStart, monthStop, "master", timeNow())
	if err != nil {
		return err
	}
	err = insertAnalysisRequest(aid2, *rid2, monthStart, monthStop, "master", timeNow())
	if err != nil {
		return err
	}

	err = insertAnalysisResponse(aid1, &FlatFeeWeight{
		Email:  "tom@tom.tom",
		Weight: 0.55,
	}, timeNow())
	if err != nil {
		return err
	}
	err = insertAnalysisResponse(aid2, &FlatFeeWeight{
		Email:  "sam@sam.sam",
		Weight: 0.6,
	}, timeNow())
	if err != nil {
		return err
	}
	return nil
}

func fakeRepoUser(email string, repoUrl string, repoName string, payoutEth string) (*uuid.UUID, *uuid.UUID, error) {
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		Email:             &email,
		Subscription:      stringPointer("sub"),
		SubscriptionState: stringPointer("Active"),
		PayoutETH:         &payoutEth,
		CreatedAt:         timeNow(),
	}

	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer(repoUrl),
		Name:        stringPointer(repoName),
		Description: stringPointer("desc"),
		CreatedAt:   timeNow(),
	}
	err := insertUser(&u)
	if err != nil {
		return nil, nil, err
	}
	id, err := insertOrUpdateRepo(&r)
	if err != nil {
		return nil, nil, err
	}

	return &u.Id, id, nil
}

func timeWarp(w http.ResponseWriter, r *http.Request, _ string) {
	m := mux.Vars(r)
	h := m["hours"]
	if h == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	hoursAdd += hours
	log.Printf("time warp: %v", timeNow())
	w.WriteHeader(http.StatusOK)
}
