package api

import (
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

const (
	ContributionsError = "Oops something went wrong with retrieving contributions. Please try again."
)

func ContributionsSend(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, false)
	if err != nil {
		slog.Error("Could not find contributions",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, ContributionsError)
		return
	}
	util.WriteJson(w, cs)
}

func ContributionsRcv(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	cs, err := db.FindContributions(user.Id, true)
	if err != nil {
		slog.Error("Could not find contributions",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, ContributionsError)
		return
	}
	util.WriteJson(w, cs)
}

func ContributionsSum2(w http.ResponseWriter, r *http.Request) {
	u := r.PathValue("uuid")
	if u == "" {
		slog.Error("UUID Parameter not set")
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	uu, err := uuid.Parse(u)
	if err != nil {
		slog.Error("Problem with parsing UUID",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	user, err := db.FindUserById(uu)
	if err != nil {
		slog.Error("Could not find user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong with retrieving the user. Please try again.")
		return
	}

	ContributionsSum(w, r, user)
}

func ContributionsSum(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		slog.Error("Could not find sponsored repos by user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}

	rbs := []db.RepoBalance{}
	for _, v := range repos {
		repoBalances, err := db.FindSumFutureBalanceByRepoId(v.Id)
		if err != nil {
			slog.Error("Could not find sum of future balance by repo id",
				slog.Any("error", err))
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
		slog.Error("Error while decoding body",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	monthStart, err := time.Parse("2006-01-02 15:04", repoMap.StartData)
	if err != nil {
		slog.Error("Error while parsing start date",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	monthStop, err := time.Parse("2006-01-02 15:04", repoMap.EndData)
	if err != nil {
		slog.Error("Error while parsing end date",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	repoMap2, err := db.FindReposByName(repoMap.Name)
	if err != nil {
		slog.Error("Error while retrieving repos by name",
			slog.Any("error", err))
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
		slog.Error("Error while inserting analysis request",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	for _, v := range repoMap.Weights {
		err = db.InsertAnalysisResponse(a.Id, v.Email, v.Names, v.Weight, util.TimeNow())
		if err != nil {
			slog.Error("Error while inserting analysis response",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}
	return
}
