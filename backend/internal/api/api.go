package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

const (
	maxTopContributors = 20
)

const (
	GenericErrorMessage            = "Oops something went wrong. Please try again."
	RepositoryNotFoundErrorMessage = "Oops something went wrong with retrieving the repositories. Please try again."
	NotAllowedToViewMessage        = "Oops you are not allowed to view this resource."
	ForcingRepoAnalysisTooSoon     = "Ooops you are trying to re-analyse too soon."
)

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.German,
})

type Timewarp struct {
	Offset int `json:"offset"`
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
	RepoId    uuid.UUID       `json:"repoid"`
}

type FakeRepoMapping struct {
	StartData string          `json:"startDate"`
	EndData   string          `json:"endDate"`
	Name      string          `json:"name"`
	Url       string          `json:"url"`
	Weights   []FlatFeeWeight `json:"weights"`
}

type FlatFeeWeight struct {
	Names       []string `json:"names"`
	Email       string   `json:"email"`
	Weight      float64  `json:"weight"`
	CommitCount int      `json:"commitcount"`
}

type ContribCommitCount struct {
	RepoId           uuid.UUID `json:"repoid"`
	ContributorCount int       `json:"contributorcount"`
	CommitCount      int       `json:"commitcount"`
}

type Plan struct {
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	PriceBase   int64
	Freq        int64  `json:"freq"`
	Description string `json:"desc"`
	Disclaimer  string `json:"disclaimer"`
	FeePrm      int64  `json:"feePrm"`
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

var Plans = []Plan{
	{
		Title:       "Monthly",
		Price:       10.31, // 30 * 0.33 = 9.9, 9.9 = 10.31 - 0.04(10.31)
		PriceBase:   10310000,
		Freq:        30,
		FeePrm:      40,
		Description: "You can help your sponsored projects on a monthly basis with a flat fee of <b>10.31 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 4%",
	},
	{
		Title:       "Yearly",
		Price:       125.47, // 365 * 0.33 = 120.45, 120.45 = 125.47 - 0.04(125.47)
		PriceBase:   125470000,
		Freq:        365,
		FeePrm:      40,
		Description: "You can help your sponsored projects on a yearly basis with a flat fee of <b>125.47 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 4%",
	},
	{
		Title:       "5 Years",
		Price:       624.09, // 1825 * 0.33 = 602.25, 602.25 = 624.09 - 0.035(624.09)
		PriceBase:   624090000,
		Freq:        1825,
		FeePrm:      35,
		Description: "You want to support open-source software for 5 years with a flat fee of <b>624.09 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 3.5%",
	},
}

type ApiHandler struct {
	stripeAPIPublicKey string
	env                string
}

func NewApiHandler(stripeAPIPublicKey string, env string) *ApiHandler {
	return &ApiHandler{stripeAPIPublicKey, env}
}

func ServerTime(w http.ResponseWriter, _ *http.Request, u *db.UserDetail) {
	currentTime := util.TimeNow()
	util.WriteJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(util.SecondsAdd())+`}`)
}

func (h *ApiHandler) Config(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(Plans)
	supportedCurrencies, err := json.Marshal(util.SupportedCurrencies)
	if err != nil {
		slog.Error("Error while writing json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	util.WriteJsonStr(w, `{
			"stripePublicApi":"`+h.stripeAPIPublicKey+`",
            "plans": `+string(b)+`,
			"env":"`+h.env+`",
			"supportedCurrencies":`+string(supportedCurrencies)+`
			}`)
}

func TimeWarp(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	h := r.PathValue("hours")
	if h == "" {
		slog.Error("Parameter hours not set")
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		slog.Error("Error while parsing hours to int",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	seconds := hours * 60 * 60
	util.AddTimeNowSeconds(seconds)
	slog.Info("Time warp",
		slog.Any("time", util.TimeNow()))
}

func lang(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	return b.String()
}
