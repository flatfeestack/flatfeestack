package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type AnalysisRequest struct {
	Id         uuid.UUID
	RepoId     uuid.UUID
	DateFrom   time.Time
	DateTo     time.Time
	GitUrl     string
	ReceivedAt *time.Time
	Error      *string
}

type AnalysisResponse struct {
	Id        uuid.UUID
	RequestId uuid.UUID `json:"request_id"`
	DateFrom  time.Time
	DateTo    time.Time
	GitEmail  string
	GitNames  []string
	Weight    float64
}

func InsertAnalysisRequest(a AnalysisRequest, now time.Time) error {
	stmt, err := DB.Prepare("INSERT INTO analysis_request(id, repo_id, date_from, date_to, git_url, created_at) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_request for %v statement event: %v", a.Id, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(a.Id, a.RepoId, a.DateFrom, a.DateTo, a.GitUrl, now.Format("2006-01-02"))
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func InsertAnalysisResponse(reqId uuid.UUID, gitEmail string, names []string, weight float64, now time.Time) error {

	bytes, err := json.Marshal(names)
	if err != nil {
		return fmt.Errorf("cannot marshal %v", err)
	}

	stmt, err := DB.Prepare(`INSERT INTO analysis_response(
                                     id, analysis_request_id, git_email, git_names, weight, created_at) 
									 VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v statement event: %v", reqId, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uuid.New(), reqId, gitEmail, string(bytes), weight, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

// https://stackoverflow.com/questions/3491329/group-by-with-maxdate
// https://pganalyze.com/docs/log-insights/app-errors/U115
func FindLatestAnalysisRequest(repoId uuid.UUID) (*AnalysisRequest, error) {
	var a AnalysisRequest

	//https://stackoverflow.com/questions/47479973/golang-postgresql-array#47480256
	err := DB.
		QueryRow(`SELECT id, repo_id, date_from, date_to, git_url, received_at, error FROM (
                          SELECT id, repo_id, date_from, date_to, git_url, received_at, error,
                            RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
                            FROM analysis_request WHERE repo_id=$1) AS x
                        WHERE dest_rank = 1`, repoId).
		Scan(&a.Id, &a.RepoId, &a.DateFrom, &a.DateTo, &a.GitUrl, &a.ReceivedAt, &a.Error)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &a, nil
	default:
		return nil, err
	}
}

func FindAllLatestAnalysisRequest(dateTo time.Time) ([]AnalysisRequest, error) {
	var as []AnalysisRequest

	rows, err := DB.Query(`SELECT id, repo_id, date_from, date_to, git_url, received_at, error FROM (
                          SELECT id, repo_id, date_from, date_to, git_url, received_at, error,
                            RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
                            FROM analysis_request) AS x
                        WHERE dest_rank = 1 AND date_to <= $1
                        ORDER BY git_url`, dateTo.Format("2006-01-02"))

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var a AnalysisRequest
		err = rows.Scan(&a.Id, &a.RepoId, &a.DateFrom, &a.DateTo, &a.GitUrl, &a.ReceivedAt, &a.Error)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func UpdateAnalysisRequest(reqId uuid.UUID, now time.Time, errStr *string) error {
	stmt, err := DB.Prepare(`UPDATE analysis_request set received_at = $1, error = $2 WHERE id = $3`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE analysis_request for statement event: %v", err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, errStr, reqId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindAnalysisResults(reqId uuid.UUID) ([]AnalysisResponse, error) {
	var ars []AnalysisResponse

	rows, err := DB.Query(`SELECT id, git_email, git_names, weight
                                 FROM analysis_response 
                                 WHERE analysis_request_id = $1`, reqId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var ar AnalysisResponse
		var jsonNames string
		err = rows.Scan(&ar.Id, &ar.GitEmail, &jsonNames, &ar.Weight)
		var names []string
		if err := json.Unmarshal([]byte(jsonNames), &names); err != nil {
			return nil, err
		}
		ar.GitNames = names

		if err != nil {
			return nil, err
		}
		ars = append(ars, ar)
	}
	return ars, nil
}

func FindRepoContribution(repoId uuid.UUID) ([]Contributions, error) {
	var cs []Contributions

	rows, err := DB.Query(`SELECT areq.date_from, areq.date_to, ares.git_email, ares.git_names, ares.weight
                        FROM analysis_request areq
                        INNER JOIN analysis_response ares on areq.id = ares.analysis_request_id
                        WHERE areq.repo_id=$1 AND areq.error IS NULL 
                        ORDER BY areq.date_to, ares.weight DESC, ares.git_email`, repoId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var c Contributions
		var jsonNames string
		err = rows.Scan(&c.DateFrom, &c.DateTo, &c.GitEmail, &jsonNames, &c.Weight)

		var names []string
		if err := json.Unmarshal([]byte(jsonNames), &names); err != nil {
			return nil, err
		}
		c.GitNames = names

		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

func FindRepoContributors(repoId uuid.UUID) (int, error) {
	var c int
	err := DB.QueryRow(`SELECT count(distinct git_email) as c
                        FROM analysis_request areq
                        INNER JOIN analysis_response ares on areq.id = ares.analysis_request_id
                        WHERE areq.repo_id=$1 AND areq.error IS NULL`, repoId).Scan(&c)

	if err != nil {
		return 0, err
	}

	return c, nil
}
