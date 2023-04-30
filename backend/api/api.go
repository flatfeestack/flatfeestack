package api

import (
	"backend/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"math/big"
	"net/http"
	"strconv"
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
		Title:       "Yearly",
		Price:       125.47, //365 * 330000 / 1-(0.04)
		PriceBase:   125468750,
		Freq:        365,
		FeePrm:      40,
		Description: "You can help your sponsored projects on a yearly basis with a flat fee of <b>125.47 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 4%",
	},
	{
		Title:       "5 Years",
		Price:       624.09, //1825 * 330000 / 1-(0.035)
		PriceBase:   624093265,
		Freq:        1825,
		FeePrm:      35,
		Description: "You want to support Open Source software for 5 years with a flat fee of <b>624.09 USD</b>",
		Disclaimer:  "Stripe charges 2.9% + 0.3 USD per transaction, with the bank transaction fee, we deduct in total 3.5%",
	},
	{
		Title:       "Beta",
		Price:       0.66,
		PriceBase:   660000,
		Freq:        2,
		Description: "Beta testing: <b>" + big.NewFloat(0.66).String() + " USD</b>",
		Disclaimer:  "",
	},
}

var (
	stripeAPIPublicKey string
	env                string
)

func InitApi(stripeAPIPublicKey0 string, env0 string) {
	stripeAPIPublicKey = stripeAPIPublicKey0
	env = env0
}

func ServerTime(w http.ResponseWriter, _ *http.Request, _ string) {
	currentTime := utils.TimeNow()
	utils.WriteJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(utils.SecondsAdd)+`}`)
}

func Config(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(Plans)
	supportedCurrencies, err := json.Marshal(utils.SupportedCurrencies)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could write json: %v", err)
		return
	}

	utils.WriteJsonStr(w, `{
			"stripePublicApi":"`+stripeAPIPublicKey+`",
            "plans": `+string(b)+`,
			"env":"`+env+`",
			"supportedCurrencies":`+string(supportedCurrencies)+`
			}`)
}

func TimeWarp(w http.ResponseWriter, r *http.Request, _ string) {
	m := mux.Vars(r)
	h := m["hours"]
	if h == "" {
		utils.WriteErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	seconds := hours * 60 * 60
	utils.SecondsAdd += seconds
	log.Printf("time warp: %v", utils.TimeNow())
}

func lang(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	return b.String()
}
