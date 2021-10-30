package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
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

type RepoMapping struct {
	StartData string          `json:"startDate"`
	EndData   string          `json:"endDate"`
	Name      string          `json:"name"`
	Weights   []FlatFeeWeight `json:"weights"`
}

type FlatFeeWeight struct {
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type Plan struct {
	Title       string    `json:"title"`
	Price       big.Float `json:"price"`
	Freq        int       `json:"freq"`
	Description string    `json:"desc"`
	Disclaimer  string    `json:"disclaimer"`
	FeePrm      int       `json:"feePrm"`
}

var plans = []Plan{}

func init() {
	py := new(big.Float)

	//365 * 330000 / 1-(0.04)
	py.SetString("125.47")
	plan := Plan{
		Title:       "Yearly",
		Price:       *py,
		Freq:        365,
		FeePrm:      40,
		Description: "By paying yearly <b>" + py.String() + " USD</b>, you help us to keep payment processing costs low and more money will reach your sponsored projects",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 4%",
	}
	plans = append(plans, plan)

	//91 * 330000 / 1-(0.05)
	py = new(big.Float)
	py.SetString("31.61")
	plan = Plan{
		Title:       "Quarterly",
		Price:       *py,
		Freq:        91,
		FeePrm:      50,
		Description: "You want to support Open Source software with a quarterly flat fee of <b>" + py.String() + " USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 5%",
	}
	plans = append(plans, plan)

	py = new(big.Float)
	py.SetString("0.66")
	plan = Plan{
		Title:       "Beta",
		Price:       *py,
		Freq:        2,
		Description: "Beta testing: <b>" + py.String() + " USD</b>",
		Disclaimer:  "",
	}
	plans = append(plans, plan)
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
	writeJson(w, user)
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
	writeJson(w, emails)
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
	err = insertGitEmail(user.Id, body.Email, &addGitEmailToken, timeNow())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save email: %v", err)
		return
	}

	email := url.QueryEscape(body.Email)
	var other = map[string]string{}
	other["token"] = addGitEmailToken
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/confirm/git-email/" + email + "/" + addGitEmailToken
	other["lang"] = lang(r)

	defaultMessage := "Is this your email address? Please confirm: " + other["url"]
	e := prepareEmail(body.Email, other,
		"template-subject-addgitemail_", "Validate your git email",
		"template-plain-addgitemail_", defaultMessage,
		"template-html-addgitemail_", other["lang"])

	go func() {
		insertEmailSent(user.Id, "gitemail-"+email, timeNow())
		err = sendEmail(opts.EmailUrl, e)
		if err != nil {
			log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
		}
	}()
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
}

func deleteMethod(w http.ResponseWriter, r *http.Request, user *User) {
	user.PaymentMethod = nil
	user.Last4 = nil
	err := updateUser(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could update user: %v", err)
		return
	}
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
		writeErr(w, http.StatusInternalServerError, "Could update method: %v", err)
		return
	}

	user.Last4 = &pm.Card.Last4
	err = updateUser(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could update user: %v", err)
		return
	}

	writeJson(w, user)
}

func updateName(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["name"]
	err := updateUserName(user.Id, a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
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
}

func updateSeats(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["seats"]
	seats, err := strconv.Atoi(a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
	err = updateDbSeats(user.PaymentCycleId, seats)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
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
	writeJson(w, repos)
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
	writeJson(w, repos)
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
	writeJson(w, repo)
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
	var repo RepoSearch
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
	writeJson(w, repo)
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
	log.Printf("papout")
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

		//only transfer if amount is larger than 25$
		if ub.Balance < 25*1000000 {
			continue
		}
		u2, err := findUserById(ub.UserId)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "find user: %v", err)
			return
		}
		u3 := UserBalance{
			PaymentCycleId: u2.PaymentCycleId,
			FromUserId:     nil,
			BalanceType:    "INCOME_REQUEST",
			CreatedAt:      timeNow(),
			UserId:         ub.UserId,
			Balance:        ub.Balance,
		}
		err = insertUserBalance(u3)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "user balance: %v", err)
			return
		}

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
	if pts != nil {
		//save remaining batch
		err = payout0(pts, batchId)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could not send payout2: %v", err)
			return
		}
	}
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
	writeJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`"}`)
}

func users(w http.ResponseWriter, r *http.Request, email string) {
	u, err := findAllUsers()
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could fetch users: %v", err)
		return
	}
	writeJson(w, u)
}

func config(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(plans)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	writeJsonStr(w, `{
			"stripePublicApi":"`+opts.StripeAPIPublicKey+`", 
			"wsBaseUrl":"`+opts.WebSocketBaseUrl+`",
			"restTimeout":"`+strconv.Itoa(opts.RestTimeout)+`",
            "plans": `+string(b)+`,
			"env":"`+opts.Env+`",
			"contractAddr":"`+opts.ContractAddr+`"}`)
}

func fakeUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("fake user")
	m := mux.Vars(r)
	n := m["email"]

	uid := uuid.New()

	u := User{
		Email:     n,
		Id:        uid,
		PayoutETH: stringPointer(fakePubKey1),
		CreatedAt: timeNow(),
	}

	err := insertUser(&u, "A")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	err = insertGitEmail(uid, n, nil, timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}
}

func fakeContribution(w http.ResponseWriter, r *http.Request, email string) {
	var repoMap RepoMapping
	err := json.NewDecoder(r.Body).Decode(&repoMap)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode fakeContribution body: %v", err)
		return
	}

	monthStart, err := time.Parse("2006-01-02 15:04", repoMap.StartData)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	monthStop, err := time.Parse("2006-01-02 15:04", repoMap.EndData)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	repo, err := findRepoByName(repoMap.Name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	aid := uuid.New()
	err = insertAnalysisRequest(aid, repo.Id, monthStart, monthStop, "master", timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	for _, v := range repoMap.Weights {

		err = insertAnalysisResponse(aid, &v, timeNow())
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
			return
		}

	}
	return
}

func fakePayment(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("fake payment")
	m := mux.Vars(r)
	n := m["email"]
	s := m["seats"]

	u, err := findUserByEmail(n)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	seats, err := strconv.Atoi(s)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	// ToDo: daysLeft removed, verify if needed to add daily_payment logic
	paymentCycleId, err := insertNewPaymentCycle(u.Id, seats, 90, timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	ubNew := UserBalance{
		PaymentCycleId: *paymentCycleId,
		UserId:         u.Id,
		Balance:        2970,
		BalanceType:    "PAY",
		CreatedAt:      timeNow(),
	}

	err = insertUserBalance(ubNew)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	err = updatePaymentCycleId(u.Id, paymentCycleId, nil)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	return
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
}

func contributionsSend(w http.ResponseWriter, _ *http.Request, user *User) {
	cs, err := findUserContributions(user.Id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	writeJson(w, cs)
}

func contributionsRcv(w http.ResponseWriter, _ *http.Request, user *User) {
	cs, err := findMyContributions(user.Id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	if len(cs) == 0 {

	}

	writeJson(w, cs)
}

func userSummary2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := findUserById(uu)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user2 := User{
		Id:    user.Id,
		Name:  user.Name,
		Role:  user.Role,
		Image: user.Image,
	}
	writeJson(w, user2)
}

func contributionsSum2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		writeErr(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := findUserById(uu)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	if user.Role != nil && *user.Role == "USR" {
		r, err := findSponsoredReposById(user.Id)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		writeJson(w, r)
	} else {
		r, err := findSponsoredReposByOrgId(user.Email)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		writeJson(w, r)
	}
}

func contributionsSum(w http.ResponseWriter, _ *http.Request, user *User) {
	if user.Role != nil && *user.Role == "USR" {
		r, err := findSponsoredReposById(user.Id)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		writeJson(w, r)
	} else {
		r, err := findSponsoredReposByOrgId(user.Email)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		writeJson(w, r)
	}
}

func pendingDailyUserPayouts(w http.ResponseWriter, _ *http.Request, user *User) {
	ub, err := getPendingDailyUserPayouts(user.Id, timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	writeJson(w, ub)
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
