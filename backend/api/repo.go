package api

import (
	clnt "backend/clients"
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

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
	SearchErrorMessage             = "Empty search. Please enter a search term and try again."
	RepositoryNotFoundErrorMessage = "Oops something went wrong with finding the repositories. Please try again."
)

func GetRepoByID(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		log.Errorf("Not a valid id %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	repo, err := db.FindRepoById(id)
	if repo == nil {
		log.Errorf("Could not find repo with id %v", id)
		utils.WriteErrorf(w, http.StatusNotFound, RepositoryNotFoundErrorMessage)
		return
	}
	if err != nil {
		log.Errorf("Could not fetch DB %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}
	utils.WriteJson(w, repo)
}

func TagRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		log.Errorf("Not a valid id %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	tagRepo0(w, user, repoId, db.Active)
}

func UnTagRepo(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		log.Errorf("Not a valid id %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	tagRepo0(w, user, repoId, db.Inactive)
}

func Graph(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		log.Errorf("Not a valid id %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	contributions, err := db.FindRepoContribution(repoId)

	offsetString := params["offset"]
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		log.Errorf("Not a valid id %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
			d.BackgroundColor = utils.GetColor1(v.GitEmail)
			d.BorderColor = utils.GetColor1(v.GitEmail)
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

	utils.WriteJson(w, data)
}

func tagRepo0(w http.ResponseWriter, user *db.UserDetail, repoId uuid.UUID, newEventType uint8) {
	repo, err := db.FindRepoById(repoId)
	if err != nil {
		log.Errorf("Could not find repo: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	now := utils.TimeNow()
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
		log.Errorf("Could not save to DB: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
		return
	}

	//no need for transaction here, repoId is very static
	log.Printf("repoId %v", repo.Id)

	if newEventType == db.Active {
		ar, err := db.FindLatestAnalysisRequest(repo.Id)
		if err != nil {
			log.Warningf("could not find latest analysis request: %v", err)
		}
		if ar == nil {
			err = clnt.RequestAnalysis(repo.Id, *repo.GitUrl)
			if err != nil {
				log.Warningf("Could not submit analysis request %v\n", err)
			}
		}
	}
	if newEventType == db.Inactive {
		//TODO
		//check if others are using it, otherwise disable fetching the metrics
	}

	utils.WriteJson(w, repo)
}

func GetSponsoredRepos(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		log.Errorf("Could not get repos: %v", err)
		utils.WriteErrorf(w, http.StatusInternalServerError, RepositoryNotFoundErrorMessage)
		return
	}

	utils.WriteJson(w, repos)
}

func SearchRepoNames(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	q := r.URL.Query().Get("q")
	log.Infof("query %v", q)
	if q == "" {
		log.Errorf("Empty search")
		utils.WriteErrorf(w, http.StatusBadRequest, SearchErrorMessage)
		return
	}
	repos, err := db.FindReposByName(q)
	if err != nil {
		log.Errorf("Could not fetch repos: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}
	utils.WriteJson(w, repos)
}

func SearchRepoGitHub(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	q := r.URL.Query().Get("q")
	log.Infof("query %v", q)
	if q == "" {
		log.Errorf("Empty search")
		utils.WriteErrorf(w, http.StatusBadRequest, SearchErrorMessage)
		return
	}

	var repos []db.Repo

	name := utils.IsValidUrl(q)

	if name != nil {
		repoId := uuid.New()
		repo := &db.Repo{
			Id:          repoId,
			Url:         utils.StringPointer(q),
			GitUrl:      utils.StringPointer(q),
			Name:        name,
			Description: utils.StringPointer("n/a"),
			Score:       0,
			Source:      utils.StringPointer("user-url"),
			CreatedAt:   utils.TimeNow(),
		}
		err := db.InsertOrUpdateRepo(repo)
		if err != nil {
			log.Errorf("Error while insert/update repo: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		repos = append(repos, *repo)
	}

	ghRepos, err := clnt.FetchGithubRepoSearch(q)
	if err != nil {
		log.Errorf("Could not fetch repos: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, RepositoryNotFoundErrorMessage)
		return
	}

	//write those to the DB...
	for _, v := range ghRepos {
		repoId := uuid.New()
		nr, err := v.Score.Float64()
		if err != nil {
			log.Errorf("Could not create score: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		repo := &db.Repo{
			Id:          repoId,
			Url:         utils.StringPointer(v.Url),
			GitUrl:      utils.StringPointer(v.GitUrl),
			Name:        utils.StringPointer(v.Name),
			Description: utils.StringPointer(v.Description),
			Score:       uint32(nr),
			Source:      utils.StringPointer("github"),
			CreatedAt:   utils.TimeNow(),
		}
		err = db.InsertOrUpdateRepo(repo)
		if err != nil {
			log.Errorf("Error while insert/update repo: %v", err)
			utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		repos = append(repos, *repo)
	}

	utils.WriteJson(w, repos)
}
