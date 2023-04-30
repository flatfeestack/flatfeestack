package api

import (
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func ContributionsSend(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, false)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	utils.WriteJson(w, cs)
}

func ContributionsRcv(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, true)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}
	utils.WriteJson(w, cs)
}

func ContributionsSum2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		utils.WriteErrorf(w, http.StatusBadRequest, "Parameter hours not set: %v", m)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	user, err := db.FindUserById(uu)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	ContributionsSum(w, r, user)
}

func ContributionsSum(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
		return
	}

	rbs := []db.RepoBalance{}
	for _, v := range repos {
		repoBalances, err := db.FindSumFutureBalanceByRepoId(v.Id)
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Could statusSponsoredUsers: %v", err)
			return
		}
		rbs = append(rbs, db.RepoBalance{
			Repo:            v,
			CurrencyBalance: repoBalances,
		})
	}

	utils.WriteJson(w, rbs)
}

func FakeContribution(w http.ResponseWriter, r *http.Request, email string) {
	var repoMap FakeRepoMapping
	err := json.NewDecoder(r.Body).Decode(&repoMap)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode fakeContribution body: %v", err)
		return
	}

	monthStart, err := time.Parse("2006-01-02 15:04", repoMap.StartData)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}
	monthStop, err := time.Parse("2006-01-02 15:04", repoMap.EndData)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	repoMap2, err := db.FindReposByName(repoMap.Name)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	var repos db.Repo
	for _, v := range repoMap2 {
		repos = v
	}

	a := db.AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repos.Id,
		DateFrom: monthStart,
		DateTo:   monthStop,
		GitUrl:   "test",
	}

	err = db.InsertAnalysisRequest(a, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
		return
	}

	for _, v := range repoMap.Weights {
		err = db.InsertAnalysisResponse(a.Id, v.Email, v.Names, v.Weight, utils.TimeNow())
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode Webhook body: %v", err)
			return
		}
	}
	return
}
