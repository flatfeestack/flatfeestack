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

func GetRepoByID(w http.ResponseWriter, r *http.Request, _ *db.User) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}

	repo, err := db.FindRepoById(id)
	if repo == nil {
		utils.WriteErrorf(w, http.StatusNotFound, "Could not find repo with id %v", id)
		return
	}
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not fetch DB %v", err)
		return
	}
	utils.WriteJson(w, repo)
}

func TagRepo(w http.ResponseWriter, r *http.Request, user *db.User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	tagRepo0(w, user, repoId, db.Active)
}

func UnTagRepo(w http.ResponseWriter, r *http.Request, user *db.User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	tagRepo0(w, user, repoId, db.Inactive)
}

func Graph(w http.ResponseWriter, r *http.Request, _ *db.User) {
	params := mux.Vars(r)
	repoId, err := uuid.Parse(params["id"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
		return
	}
	contributions, err := db.FindRepoContribution(repoId)

	offsetString := params["offset"]
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Not a valid id %v", err)
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

func tagRepo0(w http.ResponseWriter, user *db.User, repoId uuid.UUID, newEventType uint8) {
	repo, err := db.FindRepoById(repoId)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save to DB: %v", err)
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
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save to DB: %v", err)
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
			err = clnt.AnalysisReq(repo.Id, *repo.GitUrl)
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

func GetSponsoredRepos(w http.ResponseWriter, r *http.Request, user *db.User) {
	repos, err := db.FindSponsoredReposByUserId(user.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not get repos: %v", err)
		return
	}

	utils.WriteJson(w, repos)
}

func SearchRepoNames(w http.ResponseWriter, r *http.Request, _ *db.User) {
	q := r.URL.Query().Get("q")
	log.Infof("query %v", q)
	if q == "" {
		utils.WriteErrorf(w, http.StatusBadRequest, "Empty search")
		return
	}
	repos, err := db.FindReposByName(q)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
		return
	}
	utils.WriteJson(w, repos)
}

func SearchRepoGitHub(w http.ResponseWriter, r *http.Request, _ *db.User) {
	q := r.URL.Query().Get("q")
	log.Infof("query %v", q)
	if q == "" {
		utils.WriteErrorf(w, http.StatusBadRequest, "Empty search")
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
		db.InsertOrUpdateRepo(repo)
		repos = append(repos, *repo)
	}

	ghRepos, err := clnt.FetchGithubRepoSearch(q)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
		return
	}

	//write those to the DB...
	for _, v := range ghRepos {
		repoId := uuid.New()
		nr, err := v.Score.Float64()
		if err != nil {
			utils.WriteErrorf(w, http.StatusBadRequest, "Could not fetch repos: %v", err)
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
		db.InsertOrUpdateRepo(repo)
		repos = append(repos, *repo)
	}

	utils.WriteJson(w, repos)
}
