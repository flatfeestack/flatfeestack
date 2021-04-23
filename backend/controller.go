package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/alecthomas/template"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"golang.org/x/text/language"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.German,
})

type EmailRequest struct {
	MailTo      string `json:"mail_to,omitempty"`
	Subject     string `json:"subject"`
	TextMessage string `json:"text_message"`
	HtmlMessage string `json:"html_message"`
}

type ImageRequest struct {
	Image string `json:"image"`
}

type GitEmailRequest struct {
	Email string `json:"email"`
}

type EmailToken struct {
	Email string `json:"email"`
	Token string `json:"token"`
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

func getMyUser(w http.ResponseWriter, _ *http.Request, user *User) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func getPaymentCycle(w http.ResponseWriter, _ *http.Request, user *User) {
	w.Header().Set("Content-Type", "application/json")
	pc, err := findPaymentCycle(user.PaymentCycleId)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
	err = json.NewEncoder(w).Encode(pc)
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

func confirmConnectedEmails(w http.ResponseWriter, r *http.Request) {
	var emailToken EmailToken
	err := json.NewDecoder(r.Body).Decode(&emailToken)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = confirmGitEmail(emailToken.Email, emailToken.Token, timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Invalid email: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
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
	rnd, err := genRnd(16)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "ERR-reset-email-02, RND %v err %v", err)
		return
	}
	addGitEmailToken := base32.StdEncoding.EncodeToString(rnd)
	err = insertGitEmail(uuid.New(), user.Id, body.Email, addGitEmailToken, timeNow())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save email: %v", err)
		return
	}

	var other = map[string]string{}
	other["token"] = addGitEmailToken

	subject := parseTemplate("template-subject-signup_"+lang(r)+".tmpl", other)
	if subject == "" {
		subject = "Validate your email"
	}
	textMessage := parseTemplate("template-plain-signup_"+lang(r)+".tmpl", other)
	if textMessage == "" {
		textMessage = "Is this your email address? " + opts.EmailLinkPrefix + "/confirm/git-email/" + url.QueryEscape(body.Email) + "/" + addGitEmailToken
	}
	htmlMessage := parseTemplate("template-html-signup_"+lang(r)+".tmpl", other)

	e := EmailRequest{
		MailTo:      url.QueryEscape(body.Email),
		Subject:     subject,
		TextMessage: textMessage,
		HtmlMessage: htmlMessage,
	}

	url := strings.Replace(opts.EmailUrl, "{email}", url.QueryEscape(body.Email), 1)
	url = strings.Replace(url, "{token}", addGitEmailToken, 1)

	go func() {
		err = sendEmail(url, e)
		if err != nil {
			log.Printf("ERR-signup-07, send email failed: %v, %v\n", url, err)
		}
	}()

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

func deleteMethod(w http.ResponseWriter, r *http.Request, user *User) {

}

func updateMethod(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["method"]
	user.PaymentMethod = &a

	pm, err := paymentmethod.Get(
		*user.PaymentMethod,
		nil,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save payout address: %v", err)
		return
	}

	user.Last4 = &pm.Card.Last4
	err = updateUser(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save payout address: %v", err)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func updateName(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["name"]
	err := updateUserName(user.Id, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateImage(w http.ResponseWriter, r *http.Request, user *User) {
	var img ImageRequest
	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = updateUserImage(user.Id, img.Image)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateMode(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["mode"]
	if a != "USR" && a != "ORG" {
		writeErr(w, http.StatusInternalServerError, "Can only change between USR/ORG, input: %s", a)
		return
	}
	err := updateUserMode(user.Id, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
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

// @Summary Tag a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/{id}/insertOrUpdateTag [post]
func tagRepo(w http.ResponseWriter, r *http.Request, user *User) {
	var repo RepoDTO
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	sc, err := repo.Score.Int64()
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json/int: %v", err)
		return
	}
	rp := &Repo{
		Id:          uuid.New(),
		OrigId:      repo.Id,
		Url:         &repo.Url,
		GitUrl:      &repo.GitUrl,
		Branch:      &repo.Branch,
		Name:        &repo.Name,
		Description: &repo.Description,
		Tags:        nil,
		Score:       uint32(sc),
		Source:      stringPointer("github"),
		CreatedAt:   timeNow(),
	}

	repoId, err := insertOrUpdateRepo(rp)
	tagRepo0(w, user, *repoId, Active)
}

// @Summary Unsponsor a repo
// @Tags Repos
// @Param id path string true "Repo ID"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/repos/{id}/unsponsor [post]
func unTagRepo(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	tagRepo0(w, user, repoId, Inactive)
}
func tagRepo0(w http.ResponseWriter, user *User, repoId uuid.UUID, newEventType uint8) {
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
	go func() {
		if newEventType == Active {
			err = analysisRequest(repo.Id, *repo.GitUrl, *repo.Branch)
			if err != nil {
				log.Printf("Could not submit analysis request %v\n", err)
			}
		}
	}()

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
	uid1, rid1, err := fakeRepoUser("tom."+randomdata.SillyName()+"@bocek.ch", repo, repo, fakePubKey1)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data1: %v", err)
		return
	}

	repo = randomdata.SillyName()
	uid2, rid2, err := fakeRepoUser("tom."+randomdata.SillyName()+"@bocek.ch", repo, repo, fakePubKey2)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}

	repo = randomdata.SillyName()
	uid3, _, err := fakeRepoUser("tom."+randomdata.SillyName()+"@bocek.ch", repo, repo, fakePubKey2)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data3: %v", err)
		return
	}

	err = fakePayment(uid1, mUSDPerDay*90)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create fake payment 1: %v", err)
		return
	}
	err = fakePayment(uid2, mUSDPerDay*90)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create fake payment 2: %v", err)
		return
	}
	err = fakePayment(uid3, mUSDPerDay*90)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create fake payment 3: %v", err)
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
	err = insertGitEmail(uuid.New(), *uid1, "tom@tom.tom", "A", timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}
	err = insertGitEmail(uuid.New(), *uid2, "sam@sam.sam", "B", timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}

	repoMap := map[uuid.UUID][]FlatFeeWeight{}
	repoMap[*rid1] = []FlatFeeWeight{{Email: "tom@tom.tom", Weight: 0.5}}
	repoMap[*rid2] = []FlatFeeWeight{{Email: "sam@sam.sam", Weight: 0.6},
		{Email: "max@max.max", Weight: 0.1}}
	err = fakeContribution(repoMap)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could create random data2: %v", err)
		return
	}
}

func fakePayment(uid1 *uuid.UUID, amount int64) error {
	paymentCycleId, err := insertNewPaymentCycle(*uid1, 90, 1, 90, timeNow())
	if err != nil {
		return err
	}

	ubNew := UserBalance{
		PaymentCycleId: *paymentCycleId,
		UserId:         *uid1,
		Balance:        amount,
		Day:            timeNow(),
		BalanceType:    "PAY",
		CreatedAt:      timeNow(),
	}

	err = insertUserBalance(ubNew)
	if err != nil {
		return err
	}

	err = updatePaymentCycleId(*uid1, paymentCycleId)
	if err != nil {
		return err
	}
	return nil
}

func fakeContribution(repoMap map[uuid.UUID][]FlatFeeWeight) error {
	monthStart := timeNow().AddDate(0, -1, 0) //$2
	monthStop := monthStart.AddDate(0, 1, 0)  //$1

	for k, v := range repoMap {
		aid := uuid.New()
		err := insertAnalysisRequest(aid, k, monthStart, monthStop, "master", timeNow())
		if err != nil {
			return err
		}
		for _, v2 := range v {
			err = insertAnalysisResponse(aid, &v2, timeNow())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func fakeRepoUser(email string, repoUrl string, repoName string, payoutEth string) (*uuid.UUID, *uuid.UUID, error) {

	u := User{
		Email:     stringPointer(email),
		Id:        uuid.New(),
		PayoutETH: &payoutEth,
		CreatedAt: timeNow(),
	}

	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer(repoUrl),
		Name:        stringPointer(repoName),
		Description: stringPointer("desc"),
		GitUrl:      stringPointer("git-url" + randomdata.SillyName()),
		Branch:      stringPointer("branch"),
		Source:      stringPointer("gitlab"),
		CreatedAt:   timeNow(),
	}
	err := insertUser(&u, "A")
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

func genRnd(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func sendEmail(url string, e EmailRequest) error {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var jsonData []byte
	var err error
	if strings.Contains(url, "sendgrid") {
		sendGridReq := NewSingleEmailPlainText(
			NewEmail("", "info@flatfeestack.io"),
			e.Subject,
			NewEmail("", e.MailTo),
			e.TextMessage)
		jsonData, err = json.Marshal(sendGridReq)
	} else {
		jsonData, err = json.Marshal(e)
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+opts.EmailToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not update DB as status from email server: %v %v", resp.Status, resp.StatusCode)
	}
	return nil
}

func lang(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	return b.String()
}

func parseTemplate(filename string, other map[string]string) string {
	textMessage := ""
	tmplPlain, err := template.ParseFiles(filename)
	if err == nil {
		var buf bytes.Buffer
		err = tmplPlain.Execute(&buf, other)
		if err == nil {
			textMessage = buf.String()
		} else {
			log.Printf("cannot execute template file [%v], err: %v", filename, err)
		}
	} else {
		log.Printf("cannot prepare file template file [%v], err: %v", filename, err)
	}
	return textMessage
}
