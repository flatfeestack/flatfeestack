package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	ContributionsError = "Oops something went wrong with retrieving contributions. Please try again."
)

func ContributionsSend(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, false)
	if err != nil {
		log.Errorf("Could not find contributions: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, ContributionsError)
		return
	}
	util.WriteJson(w, cs)
}

func ContributionsRcv(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, true)
	if err != nil {
		log.Errorf("Could not find contributions: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, ContributionsError)
		return
	}
	util.WriteJson(w, cs)
}

func ContributionsSum2(w http.ResponseWriter, r *http.Request) {
	m := mux.Vars(r)
	u := m["uuid"]
	if u == "" {
		log.Errorf("UUID Parameter not set: %v", m)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		log.Errorf("Problem with parsing UUID: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	user, err := db.FindUserById(uu)
	if err != nil {
		log.Errorf("Could not find user: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong with retrieving the user. Please try again.")
		return
	}

	ContributionsSum(w, r, user)
}

func ContributionsSum(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		log.Errorf("Could not find sponsored repos by user: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}

	rbs := []db.RepoBalance{}
	for _, v := range repos {
		repoBalances, err := db.FindSumFutureBalanceByRepoId(v.Id)
		if err != nil {
			log.Errorf("Could not find sum of future balance by repo id: %v", err)
			util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
			return
		}
		rbs = append(rbs, db.RepoBalance{
			Repo:            v,
			CurrencyBalance: repoBalances,
		})
	}

	util.WriteJson(w, rbs)
}

func FakeContribution(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	var repoMap FakeRepoMapping
	err := json.NewDecoder(r.Body).Decode(&repoMap)
	if err != nil {
		log.Errorf("Error while decoding body: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	monthStart, err := time.Parse("2006-01-02 15:04", repoMap.StartData)
	if err != nil {
		log.Errorf("Error while parsing start date: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	monthStop, err := time.Parse("2006-01-02 15:04", repoMap.EndData)
	if err != nil {
		log.Errorf("Error while parsing end date: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	repoMap2, err := db.FindReposByName(repoMap.Name)
	if err != nil {
		log.Errorf("Error while retrieving repos by name: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
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

	err = db.InsertAnalysisRequest(a, util.TimeNow())
	if err != nil {
		log.Errorf("Error while inserting analysis request: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	for _, v := range repoMap.Weights {
		err = db.InsertAnalysisResponse(a.Id, v.Email, v.Names, v.Weight, util.TimeNow())
		if err != nil {
			log.Errorf("Error while inserting analysis response: %v", err)
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}
	return
}
