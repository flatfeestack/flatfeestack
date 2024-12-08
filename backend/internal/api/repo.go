package api

import (
	"backend/internal/client"
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type RepoHandler struct {
	c *client.AnalysisClient
	g *client.GithubClient
}

func NewRepoHandler(c *client.AnalysisClient, g *client.GithubClient) *RepoHandler {
	return &RepoHandler{c, g}
}

// Data wraps the "data" JSON
type Data struct {
	Days     int       `json:"days"`
	Total    int       `json:"total"`
	Datasets []Dataset `json:"datasets"`
	Labels   []string  `json:"labels"`
}

type Dataset struct {
	Label string    `json:"label,omitempty"`
	Data  []float64 `json:"data"`
	Fill  bool      `json:"fill,omitempty"`
	//https://www.chartjs.org/docs/latest/configuration/elements.html
	BackgroundColor  string `json:"backgroundColor"`
	BorderColor      string `json:"borderColor"`
	PointBorderWidth int    `json:"pointBorderWidth"`
}

const (
	SearchErrorMessage = "Empty search. Please enter a search term and try again."
)

func GetRepoByID(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id ",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	repo, err := db.FindRepoById(id)
	if repo == nil {
		slog.Error("Could not find repo with id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusNotFound, RepositoryNotFoundErrorMessage)
		return
	}
	if err != nil {
		slog.Error("Could not fetch DB",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}
	util.WriteJson(w, repo)
}

func (rs *RepoHandler) TagRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.tagRepo0(w, user, repoId, db.Active)
}

func (rs *RepoHandler) UnTagRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.tagRepo0(w, user, repoId, db.Inactive)
}

func (rs *RepoHandler) SetMultiplierRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.multiplierRepo0(w, user, repoId, db.Active)
}

func (rs *RepoHandler) UnsetMultiplierRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.multiplierRepo0(w, user, repoId, db.Inactive)
}

func (rs *RepoHandler) TrustRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)

	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.trustRepo0(w, user, repoId, db.Active)
}

func (rs *RepoHandler) UnTrustRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	rs.trustRepo0(w, user, repoId, db.Inactive)
}

func Graph(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	contributions, _ := db.FindRepoContribution(repoId)
	contributors, _ := db.FindRepoContributors(repoId)

	offsetString := r.PathValue("offset")
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	data := Data{}
	data.Total = contributors

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
			d.BackgroundColor = util.GetColor1(v.GitEmail)
			d.BorderColor = util.GetColor1(v.GitEmail)
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

	util.WriteJson(w, data)
}

func (rs *RepoHandler) tagRepo0(w http.ResponseWriter, user *db.UserDetail, repoId uuid.UUID, newEventType uint8) {
	repo, err := db.FindRepoById(repoId)
	if err != nil {
		slog.Error("Could not find repo",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	now := util.TimeNow()
	event := db.SponsorEvent{
		Id:          uuid.New(),
		Uid:         user.Id,
		RepoId:      repo.Id,
		EventType:   newEventType,
		SponsorAt:   &now,
		UnSponsorAt: &now,
	}

	if newEventType == db.Active {
		event.UnSponsorAt = nil
	} else {
		event.SponsorAt = nil
	}

	err = db.InsertOrUpdateSponsor(&event)
	if err != nil {
		slog.Error("Could not save to DB",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	//no need for transaction here, repoId is very static
	if newEventType == db.Active {
		ar, err := db.FindLatestAnalysisRequest(repo.Id)
		if err != nil {
			slog.Warn("could not find latest analysis request",
				slog.Any("error", err))
		}
		if ar == nil {
			err = rs.c.RequestAnalysis(repo.Id, *repo.GitUrl)
			if err != nil {
				slog.Warn("Could not submit analysis request",
					slog.Any("error", err))
			}
		}
	}
	if newEventType == db.Inactive {
		//TODO
		//check if others are using it, otherwise disable fetching the metrics
	}

	util.WriteJson(w, repo)
}

func (rs *RepoHandler) multiplierRepo0(w http.ResponseWriter, user *db.UserDetail, repoId uuid.UUID, newEventType uint8) {
	repo, err := db.FindRepoById(repoId)
	if err != nil {
		slog.Error("Could not find repo",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	now := util.TimeNow()
	event := db.MultiplierEvent{
		Id:             uuid.New(),
		Uid:            user.Id,
		RepoId:         repo.Id,
		EventType:      newEventType,
		MultiplierAt:   &now,
		UnMultiplierAt: &now,
	}

	if newEventType == db.Active {
		event.UnMultiplierAt = nil
	} else {
		event.MultiplierAt = nil
	}

	err = db.InsertOrUpdateMultiplierRepo(&event)
	if err != nil {
		slog.Error("Could not save to DB",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	util.WriteJson(w, repo)
}

func (rs *RepoHandler) trustRepo0(w http.ResponseWriter, user *db.UserDetail, repoId uuid.UUID, newEventType uint8) {
	repo, err := db.FindRepoById(repoId)
	if err != nil {
		slog.Error("Could not find repo",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	now := util.TimeNow()
	event := db.TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: newEventType,
		TrustAt:   &now,
		UnTrustAt: &now,
	}

	if newEventType == db.Active {
		event.UnTrustAt = nil
	} else {
		event.TrustAt = nil
	}

	err = db.InsertOrUpdateTrustRepo(&event)
	if err != nil {
		slog.Error("Could not save to DB",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	if newEventType == db.Active {
		repoWithTrustDate, err := db.FindRepoWithTrustDateById(repoId)
		if err != nil {
			slog.Error("Could not find repo",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
			return
		}

		healthValue, err := getRepoHealthValue(repoWithTrustDate.Id)
		if err != nil {
			repoWithTrustDate.HealthValue = 0.0
		} else {
			repoWithTrustDate.HealthValue = healthValue.HealthValue
		}

		util.WriteJson(w, repoWithTrustDate)
	} else {
		util.WriteJson(w, repo)
	}

}

func GetSponsoredRepos(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		slog.Error("Could not get repos",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	util.WriteJson(w, repos)
}

func GetMultipliedRepos(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	repos, err := db.FindMultiplierRepoByUserId(user.Id)
	if err != nil {
		slog.Error("Could not get repos",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	util.WriteJson(w, repos)
}

func GetTrustedRepos(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	repos, err := db.FindTrustedRepos()
	if err != nil {
		slog.Error("Could not get repos",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	for i, repo := range repos {
		value, err := getRepoHealthValue(repo.Id)
		if err != nil {
			repos[i].HealthValue = 0
			repos[i].Analyzed = false
		} else {
			repos[i].HealthValue = value.HealthValue
			repos[i].Analyzed = true
		}
	}

	util.WriteJson(w, repos)
}

func SearchRepoNames(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	q := r.URL.Query().Get("q")
	if q == "" {
		util.WriteErrorf(w, http.StatusBadRequest, SearchErrorMessage)
		return
	}
	repos, err := db.FindReposByName(q)
	if err != nil {
		slog.Error("Could not fetch repos",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}
	util.WriteJson(w, repos)
}

func (rh *RepoHandler) SearchRepoGitHub(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	q := r.URL.Query().Get("q")
	if q == "" {
		util.WriteErrorf(w, http.StatusBadRequest, SearchErrorMessage)
		return
	}

	var repos []db.Repo

	name := util.IsValidUrl(q)

	if name != nil {
		repoId := uuid.New()
		repo := &db.Repo{
			Id:          repoId,
			Url:         util.StringPointer(q),
			GitUrl:      util.StringPointer(q),
			Name:        name,
			Description: util.StringPointer("n/a"),
			Score:       0,
			Source:      util.StringPointer("user-url"),
			CreatedAt:   util.TimeNow(),
		}
		err := db.InsertOrUpdateRepo(repo)
		if err != nil {
			slog.Error("Error while insert/update repo",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		repos = append(repos, *repo)
	}

	ghRepos, err := rh.g.FetchGithubRepoSearch(q)
	if err != nil {
		slog.Error("Could not fetch repos",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}

	//write those to the DB...
	for _, v := range ghRepos {
		repoId := uuid.New()
		nr, err := v.Score.Float64()
		if err != nil {
			slog.Error("Could not create score",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		repo := &db.Repo{
			Id:          repoId,
			Url:         util.StringPointer(v.Url),
			GitUrl:      util.StringPointer(v.GitUrl),
			Name:        util.StringPointer(v.Name),
			Description: util.StringPointer(v.Description),
			Score:       uint32(nr),
			Source:      util.StringPointer("github"),
			CreatedAt:   util.TimeNow(),
		}
		err = db.InsertOrUpdateRepo(repo)
		if err != nil {
			slog.Error("Error while insert/update repo",
				slog.Any("error", err))
			util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		healthValue, err := getRepoHealthValue(repo.Id)
		if err != nil {
			repo.HealthValue = 0.0
			repo.Analyzed = false
		} else {
			repo.HealthValue = healthValue.HealthValue
			repo.Analyzed = true
		}

		repos = append(repos, *repo)
	}

	util.WriteJson(w, repos)
}

func (rs *RepoHandler) ForceAnalyzeRepo(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	idStr := r.PathValue("id")
	repoId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Not a valid id",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	repo, err := db.FindRepoById(repoId)
	if err != nil {
		slog.Error("Could not find repo",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	lastAnalysis, err := db.FindRepoHealthMetricsByRepoId(repoId)
	if err != nil || lastAnalysis == nil {
		slog.Error("Could not find repo metrics. Requesting analysis.",
			slog.Any("error", err))
		err = rs.c.RequestAnalysis(repo.Id, *repo.GitUrl)
		if err != nil {
			slog.Warn("Could not submit analysis request",
				slog.Any("error", err))
		}
		return
	}

	if time.Since(lastAnalysis.CreatedAt) > time.Hour {
		err = rs.c.RequestAnalysis(repo.Id, *repo.GitUrl)
		if err != nil {
			slog.Warn("Could not submit analysis request",
				slog.Any("error", err))
		}
	} else {
		util.WriteErrorf(w, http.StatusInternalServerError, ForcingRepoAnalysisTooSoon)
	}
}
