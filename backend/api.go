package main

import (
	"bytes"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spaolacci/murmur3"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"golang.org/x/text/language"
)

const (
	maxTopContributors = 20
)

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.German,
})

type Timewarp struct {
	Offset int `json:"offset"`
}

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

type GitUrl struct {
	GitUrl string `json:"gitUrl"`
}

type WebhookCallback struct {
	RequestId string          `json:"requestId"`
	Error     *string         `json:"error"`
	Result    []FlatFeeWeight `json:"result"`
}

type FakeRepoMapping struct {
	StartData string          `json:"startDate"`
	EndData   string          `json:"endDate"`
	Name      string          `json:"name"`
	Url       string          `json:"url"`
	Weights   []FlatFeeWeight `json:"weights"`
}

type FlatFeeWeight struct {
	Names  []string `json:"names"`
	Email  string   `json:"email"`
	Weight float64  `json:"weight"`
}

type Plan struct {
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Freq        int64   `json:"freq"`
	Description string  `json:"desc"`
	Disclaimer  string  `json:"disclaimer"`
	FeePrm      int64   `json:"feePrm"`
}

type Currencies struct {
	Name       string `json:"name"`
	Short      string `json:"short"`
	Smallest   string `json:"smallest"`
	FactorPow  int64  `json:"factorPow"`
	IsCrypto   bool   `json:"isCrypto"`
	PayoutName string
}

var supportedCurrencies = map[string]Currencies{
	"ETH": {
		Name:       "Ethereum",
		Short:      "ETH",
		Smallest:   "wei",
		FactorPow:  18,
		IsCrypto:   true,
		PayoutName: "eth",
	},
	"GAS": {
		Name:       "Neo Gas",
		Short:      "GAS",
		Smallest:   "mGAS",
		FactorPow:  8,
		IsCrypto:   true,
		PayoutName: "neo",
	},
	"USD": {
		Name:       "US Dollar",
		Short:      "USD",
		Smallest:   "mUSD",
		FactorPow:  6,
		IsCrypto:   false,
		PayoutName: "usdc",
	},
}

type PayoutMeta struct {
	Currency string
	Tea      int64
}

type PayoutToService struct {
	Address      string       `json:"address"`
	ExchangeRate big.Float    `json:"exchange_rate_USD_ETH"`
	Tea          int64        `json:"nano_tea"`
	Meta         []PayoutMeta `json:"meta"`
}

type NameWeight struct {
	Names  []string
	Weight float64
}

var plans = []Plan{
	{
		Title:       "Yearly",
		Price:       125.47, //365 * 330000 / 1-(0.04)
		Freq:        365,
		FeePrm:      40,
		Description: "You can help your sponsored projects on a yearly basis with a flat fee of <b>125.47 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 4%",
	},
	{
		Title:       "5 Years",
		Price:       624.09, //1825 * 330000 / 1-(0.035)
		Freq:        1825,
		FeePrm:      35,
		Description: "You want to support Open Source software for 5 years with a flat fee of <b>624.09 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 3.5%",
	},
	{
		Title:       "Beta",
		Price:       0.66,
		Freq:        2,
		Description: "Beta testing: <b>" + big.NewFloat(0.66).String() + " USD</b>",
		Disclaimer:  "",
	},
}

const (
	usdFactor = 1_000_000 // 10^6
)

func getFactorInt(currency string) (int64, error) {
	for supportedCurrency, cryptoCurrency := range supportedCurrencies {
		if supportedCurrency == strings.ToUpper(currency) {
			return IntPow(10, cryptoCurrency.FactorPow), nil
		}
	}
	return 0, fmt.Errorf("currency not found, %v", currency)
}

func getFactor(currency string) (*big.Int, error) {
	for supportedCurrency, cryptoCurrency := range supportedCurrencies {
		if supportedCurrency == strings.ToUpper(currency) {
			return new(big.Int).Exp(big.NewInt(10), big.NewInt(cryptoCurrency.FactorPow), nil), nil
		}
	}
	return nil, fmt.Errorf("currency not found, %v", currency)
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
		writeErrorf(w, http.StatusInternalServerError, "Could not find git emails %v", err)
		return
	}
	writeJson(w, emails)
}

func confirmConnectedEmails(w http.ResponseWriter, r *http.Request) {
	var emailToken EmailToken
	err := json.NewDecoder(r.Body).Decode(&emailToken)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = confirmGitEmail(emailToken.Email, emailToken.Token, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Invalid email: %v", err)
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
		writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	rnd, err := genRnd(20)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-reset-email-02, err %v", err)
		return
	}
	addGitEmailToken := base32.StdEncoding.EncodeToString(rnd)
	err = insertGitEmail(user.Id, body.Email, &addGitEmailToken, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not save email: %v", err)
		return
	}

	email := url.QueryEscape(body.Email)
	sendAddGit(email, addGitEmailToken, lang(r))
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
		writeErrorf(w, http.StatusBadRequest, "Invalid email: %v", err)
		return
	}
}

func deleteMethod(w http.ResponseWriter, r *http.Request, user *User) {
	user.PaymentMethod = nil
	user.Last4 = nil
	err := updateUser(user)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could update user: %v", err)
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
		writeErrorf(w, http.StatusInternalServerError, "Could update method: %v", err)
		return
	}

	user.Last4 = &pm.Card.Last4
	err = updateUser(user)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could update user: %v", err)
		return
	}

	writeJson(w, user)
}

func updateName(w http.ResponseWriter, r *http.Request, user *User) {
	params := mux.Vars(r)
	a := params["name"]
	err := updateUserName(user.Id, a)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not save name: %v", err)
		return
	}
}

func updateImage(w http.ResponseWriter, r *http.Request, user *User) {
	var img ImageRequest
	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = updateUserImage(user.Id, img.Image)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not save name: %v", err)
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
		writeErrorf(w, http.StatusInternalServerError, "Could not get repos: %v", err)
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
	log.Infof("query %v", q)
	if q == "" {
		writeErrorf(w, http.StatusBadRequest, "Empty search")
		return
	}

	var repos []Repo

	name := isValidUrl(q)

	if name != nil {
		repoId := uuid.New()
		repo := &Repo{
			Id:          repoId,
			Url:         stringPointer(q),
			GitUrl:      stringPointer(q),
			Name:        name,
			Description: stringPointer("n/a"),
			Score:       0,
			Source:      stringPointer("user-url"),
			CreatedAt:   timeNow(),
		}
		insertOrUpdateRepo(repo)
		repos = append(repos, *repo)
	}

	ghRepos, err := fetchGithubRepoSearch(q)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
		return
	}

	//write those to the DB...
	for _, v := range ghRepos {
		repoId := uuid.New()
		nr, err := v.Score.Float64()
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
			return
		}
		repo := &Repo{
			Id:          repoId,
			Url:         stringPointer(v.Url),
			GitUrl:      stringPointer(v.GitUrl),
			Name:        stringPointer(v.Name),
			Description: stringPointer(v.Description),
			Score:       uint32(nr),
			Source:      stringPointer("github"),
			CreatedAt:   timeNow(),
		}
		insertOrUpdateRepo(repo)
		repos = append(repos, *repo)
	}

	writeJson(w, repos)
}

func searchRepoNames(w http.ResponseWriter, r *http.Request, _ *User) {
	q := r.URL.Query().Get("q")
	log.Infof("query %v", q)
	if q == "" {
		writeErrorf(w, http.StatusBadRequest, "Empty search")
		return
	}
	repos, err := findReposByName(q)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
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
		writeErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}

	repo, err := findRepoById(id)
	if repo == nil {
		writeErrorf(w, http.StatusNotFound, "Could not find repo with id %v", id)
		return
	}
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not fetch DB %v", err)
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
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	tagRepo0(w, user, repoId, Active)
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
		writeErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	tagRepo0(w, user, repoId, Inactive)
}

func graph(w http.ResponseWriter, r *http.Request, _ *User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	contributions, err := findRepoContribution(repoId)

	offsetString := params["offset"]
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}

	data := Data{}
	data.Total = len(contributions)

	perDay := make(map[string]*Dataset)
	previousDay := time.Time{}
	days := 0
	nrDay := 0

	for _, v := range contributions {
		if v.DateTo != previousDay {
			data.Labels = append(data.Labels, v.DateTo.Format("02.01.2006"))
			days++
			nrDay = 0
			previousDay = v.DateTo
		}
		nrDay++
		if nrDay-offset < 0 || nrDay-offset > maxTopContributors {
			continue
		}

		d := perDay[v.GitEmail]
		if d == nil {
			d = &Dataset{}
			d.Fill = false
			names, err := json.Marshal(v.GitNames)
			if err != nil {
				continue
			}
			d.Label = v.GitEmail + ";" + string(names)
			d.BackgroundColor = getColor1(v.GitEmail)
			d.BorderColor = getColor1(v.GitEmail)
			d.PointBorderWidth = 3
			perDay[v.GitEmail] = d
		}
		d.Data = append(d.Data, v.Weight)
	}

	m := make([]Dataset, 0, len(perDay))
	for _, val := range perDay {
		m = append(m, *val)
	}
	data.Days = days
	data.Datasets = m

	writeJson(w, data)
}

func myHash(s string) float64 {
	i := murmur3.Sum32([]byte(s))
	const maxUint32 = ^uint32(0)
	return float64(i) / float64(maxUint32)
}

func getColor1(input string) string {
	a := strconv.Itoa(int(12 * (30 * myHash(input+"a"))))
	b := strconv.Itoa(int(35 + 10*(5*myHash(input+"b"))))
	c := strconv.Itoa(int(25 + 10*(5*myHash(input+"c"))))
	return "hsl(" + a + "," + b + "%," + c + "%)"
}

func tagRepo0(w http.ResponseWriter, user *User, repoId uuid.UUID, newEventType uint8) {
	repo, err := findRepoById(repoId)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not save to DB: %v", err)
		return
	}

	now := timeNow()
	event := SponsorEvent{
		Id:          uuid.New(),
		Uid:         user.Id,
		RepoId:      repo.Id,
		EventType:   newEventType,
		SponsorAt:   now,
		UnSponsorAt: &now,
	}
	err = insertOrUpdateSponsor(&event)
	if err != nil {
		writeErrorf(w, http.StatusInternalServerError, "Could not save to DB: %v", err)
		return
	}

	//no need for transaction here, repoId is very static
	log.Printf("repoId %v", repo.Id)

	if newEventType == Active {
		ar, err := findLatestAnalysisRequest(repo.Id)
		if err != nil {
			log.Warningf("could not find latest analysis request: %v", err)
		}
		if ar == nil {
			err = analysisRequest(repo.Id, *repo.GitUrl)
			if err != nil {
				log.Warningf("Could not submit analysis request %v\n", err)
			}
		}
	}
	if newEventType == Inactive {
		//TODO
		//check if others are using it, otherwise disable fetching the metrics
	}

	writeJson(w, repo)
}

func extractGitUrls(repos []Repo) []string {
	gitUrls := []string{}
	for _, v := range repos {
		if v.GitUrl != nil {
			gitUrls = append(gitUrls, *v.GitUrl)
		}
	}
	return gitUrls
}

func analysisEngineHook(w http.ResponseWriter, r *http.Request, email string) {
	w.Header().Set("Content-Type", "application/json")
	var data WebhookCallback
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	reqId, err := uuid.Parse(data.RequestId)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "cannot parse request id: %v", err)
		return
	}

	rowsAffected := 0
	for _, v := range data.Result {
		err = insertAnalysisResponse(reqId, v.Email, v.Names, v.Weight, timeNow())
		if err != nil {
			writeErrorf(w, http.StatusInternalServerError, "insert error: %v", err)
			return
		}
		rowsAffected++
	}

	errA := updateAnalysisRequest(reqId, timeNow(), data.Error)
	if errA != nil {
		log.Warnf("cannot send to analyze engine %v", errA)
	}

	log.Printf("Inserted %v contributions into DB for request %v", rowsAffected, data.RequestId)
	w.WriteHeader(http.StatusOK)
}

func serverTime(w http.ResponseWriter, r *http.Request, email string) {
	currentTime := timeNow()
	writeJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(secondsAdd)+`}`)
}

func users(w http.ResponseWriter, r *http.Request, _ string) {
	u, err := findAllEmails()
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could fetch users: %v", err)
		return
	}
	writeJson(w, u)
}

func config(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(plans)
	supportedCurrencies, err := json.Marshal(supportedCurrencies)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	writeJsonStr(w, `{
			"stripePublicApi":"`+opts.StripeAPIPublicKey+`", 
			"wsBaseUrl":"`+opts.WebSocketBaseUrl+`",
            "plans": `+string(b)+`,
			"env":"`+opts.Env+`",
			"supportedCurrencies":`+string(supportedCurrencies)+`
			}`)
}

func fakeUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("fake user")
	m := mux.Vars(r)
	n := m["email"]

	uid := uuid.New()
	payOutI := uuid.New()

	u := User{
		Email:             n,
		Id:                uid,
		PaymentCycleOutId: payOutI,
		CreatedAt:         timeNow(),
	}

	err := insertUser(&u)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	err = insertGitEmail(uid, n, nil, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}
}

func fakeContribution(w http.ResponseWriter, r *http.Request, email string) {
	var repoMap FakeRepoMapping
	err := json.NewDecoder(r.Body).Decode(&repoMap)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode fakeContribution body: %v", err)
		return
	}

	monthStart, err := time.Parse("2006-01-02 15:04", repoMap.StartData)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	monthStop, err := time.Parse("2006-01-02 15:04", repoMap.EndData)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	repoMap2, err := findReposByName(repoMap.Name)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	var repos Repo
	for _, v := range repoMap2 {
		repos = v
	}

	a := AnalysisRequest{
		RequestId: uuid.New(),
		RepoId:    repos.Id,
		DateFrom:  monthStart,
		DateTo:    monthStop,
		GitUrl:    "test",
	}

	err = insertAnalysisRequest(a, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	for _, v := range repoMap.Weights {
		err = insertAnalysisResponse(a.RequestId, v.Email, v.Names, v.Weight, timeNow())
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
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
	if err != nil || u == nil {
		writeErrorf(w, http.StatusBadRequest, "Unable to find user from given e-mail address: %v", err)
		return
	}

	seats, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not determine amount of seats: %v", err)
		return
	}

	paymentCycleInId, err := insertNewPaymentCycleIn(365, seats, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not create new payment cycle: %v", err)
		return
	}

	balance := big.NewInt(120000000)

	ubNew := UserBalance{
		PaymentCycleInId: paymentCycleInId,
		UserId:           u.Id,
		Balance:          balance,
		BalanceType:      "PAY",
		CreatedAt:        timeNow(),
		Split:            new(big.Int).Div(balance, big.NewInt(365*seats)),
		Currency:         "USD",
	}

	err = insertUserBalance(ubNew)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not create new user balance: %v", err)
		return
	}

	err = updatePaymentCycleInId(u.Id, paymentCycleInId)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could not update payment cycle: %v", err)
		return
	}
	return
}

func timeWarp(w http.ResponseWriter, r *http.Request, _ string) {
	m := mux.Vars(r)
	h := m["hours"]
	if h == "" {
		writeErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	seconds := hours * 60 * 60
	secondsAdd += seconds
	log.Printf("time warp: %v", timeNow())
}

func crontester(w http.ResponseWriter, r *http.Request, _ string) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return
	}

	yesterdayStop, _ := time.Parse(time.RFC3339, data["yesterdayStop"])
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	log.Printf("Start daily runner from %v to %v", yesterdayStart, yesterdayStop)

	/*nr, err := runDailyUserBalance(yesterdayStart, yesterdayStop, timeNow())
	if err != nil {
		return
	}
	log.Printf("Daily User Balance inserted %v entries", nr)

	nr, err = runDailyDaysLeftDailyPayment()
	if err != nil {
		return
	}
	log.Printf("Daily Days Left Daily Payment updated %v entries", nr)

	nr, err = runDailyDaysLeftPaymentCycle()
	if err != nil {
		return
	}
	log.Printf("Daily Days Left Payment Cycle updated %v entries", nr)

	nr, err = runDailyRepoBalance(yesterdayStart, yesterdayStop, timeNow())
	if err != nil {
		return
	}
	log.Printf("Daily Repo Balance inserted %v entries", nr)

	nr, err = runDailyRepoWeight(yesterdayStart, yesterdayStop, timeNow())
	if err != nil {
		return
	}
	log.Printf("Daily Repo Weight inserted %v entries", nr)

	nr, err = runDailyUserPayout(yesterdayStart, yesterdayStop, timeNow())
	if err != nil {
		return
	}
	log.Printf("Daily User Payout inserted %v entries", nr)

	nr, err = runDailyFutureLeftover(yesterdayStart, yesterdayStop, timeNow())
	if err != nil {
		return
	}
	log.Printf("Daily Leftover inserted %v entries", nr)*/
}

func contributionsSend(w http.ResponseWriter, _ *http.Request, user *User) {
	cs, err := findContributions(user.Id, false)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	writeJson(w, cs)
}

func contributionsRcv(w http.ResponseWriter, _ *http.Request, user *User) {
	cs, err := findContributions(user.Id, true)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	writeJson(w, cs)
}

func userSummary2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		writeErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := findUserById(uu)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user2 := User{
		Id:    user.Id,
		Name:  user.Name,
		Image: user.Image,
	}
	writeJson(w, user2)
}

func contributionsSum2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		writeErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := findUserById(uu)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	contributionsSum(w, r, user)
}

func contributionsSum(w http.ResponseWriter, _ *http.Request, user *User) {
	repos, err := findSponsoredReposById(user.Id)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	rbs := []RepoBalance{}
	for _, v := range repos {
		repoBalances, err := findSumFutureBalanceByRepoId(&v.Id)
		if err != nil {
			writeErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		rbs = append(rbs, RepoBalance{
			Repo:            v,
			CurrencyBalance: repoBalances,
		})
	}

	writeJson(w, rbs)
}

func requestPayout(w http.ResponseWriter, r *http.Request, user *User) {
	m := mux.Vars(r)
	targetCurrency := m["targetCurrency"]

	currencyMetadata, ok := supportedCurrencies[targetCurrency]
	if !ok {
		writeErrorf(w, http.StatusBadRequest, "Unsupported currency requested")
		return
	}

	ownContributionIds, err := findOwnContributionIds(user.Id, targetCurrency)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Unable to retrieve contributions: %v", err)
		return
	}

	totalEarnedAmount, err := sumTotalEarnedAmountForContributionIds(ownContributionIds)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Unable to retrieve already earned amount in target currency: %v", err)
		return
	}

	if targetCurrency == "USD" {
		// For USDC, 10^18 units are one dollar
		// See explorer https://explorer.bitquery.io/bsc/token/0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d
		// And https://docs.openzeppelin.com/contracts/4.x/erc20#a-note-on-decimals
		// FlatFeeStack already calculates in micro dollars, so we need to blow up the value a bit
		usdcDecimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18-currencyMetadata.FactorPow), nil)

		// Divide the float by the power of 10
		totalEarnedAmount.Mul(totalEarnedAmount, usdcDecimals)
	}

	signature, err := payoutRequest(user.Id, totalEarnedAmount, currencyMetadata.PayoutName)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Error when generating signature: %v", err)
		return
	}

	err = markContributionAsClaimed(ownContributionIds)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "Error when marking contributions as claimed: %v", err)
		return
	} else {
		writeJson(w, signature)
	}
}

func lang(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	return b.String()
}

func parseTemplate(filename string, other map[string]string) string {
	textMessage := ""
	tmplPlain, err := template.ParseFiles("mail-templates/" + filename)
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
