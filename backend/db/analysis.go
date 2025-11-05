package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (db *DB) InsertAnalysisRequest(a AnalysisRequest, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO analysis_request(id, repo_id, date_from, date_to, git_url, created_at) 
		 VALUES($1, $2, $3, $4, $5, $6)`,
		a.Id, a.RepoId, a.DateFrom, a.DateTo, a.GitUrl, now)
	return err
}

func (db *DB) InsertAnalysisResponse(reqId uuid.UUID, repoId uuid.UUID, gitEmail string, names []string, weight float64, now time.Time) error {
	namesJSON, err := json.Marshal(names)
	if err != nil {
		return fmt.Errorf("cannot marshal names: %w", err)
	}

	_, err = db.Exec(
		`INSERT INTO analysis_response(id, analysis_request_id, repo_id, git_email, git_names, weight, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), reqId, repoId, gitEmail, namesJSON, weight, now)
	return err
}

func (db *DB) FindLatestAnalysisRequest(repoId uuid.UUID) (*AnalysisRequest, error) {
	var a AnalysisRequest

	err := db.QueryRow(
		`SELECT id, repo_id, date_from, date_to, git_url, received_at, error 
		 FROM (
			 SELECT id, repo_id, date_from, date_to, git_url, received_at, error,
				 RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
			 FROM analysis_request WHERE repo_id=$1
		 ) AS x
		 WHERE dest_rank = 1`, 
		repoId).
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

func (db *DB) FindAllLatestAnalysisRequest(dateTo time.Time) ([]AnalysisRequest, error) {
	var as []AnalysisRequest

	rows, err := db.Query(
		`SELECT id, repo_id, date_from, date_to, git_url, received_at, error 
		 FROM (
			 SELECT id, repo_id, date_from, date_to, git_url, received_at, error,
				 RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
			 FROM analysis_request
		 ) AS x
		 WHERE dest_rank = 1 AND date_to <= $1
		 ORDER BY git_url`, 
		dateTo)

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

func (db *DB) UpdateAnalysisRequest(reqId uuid.UUID, now time.Time, errStr *string) error {
	_, err := db.Exec(
		`UPDATE analysis_request SET received_at = $1, error = $2 WHERE id = $3`,
		now, errStr, reqId)
	return err
}

func (db *DB) FindAnalysisResults(reqId uuid.UUID) ([]AnalysisResponse, error) {
	var ars []AnalysisResponse

	rows, err := db.Query(
		`SELECT id, git_email, git_names, weight
		 FROM analysis_response 
		 WHERE analysis_request_id = $1`, 
		reqId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var ar AnalysisResponse
		var jsonNames string
		err = rows.Scan(&ar.Id, &ar.GitEmail, &jsonNames, &ar.Weight)
		if err != nil {
			return nil, err
		}

		var names []string
		if err := json.Unmarshal([]byte(jsonNames), &names); err != nil {
			return nil, fmt.Errorf("unmarshal git_names: %w", err)
		}
		ar.GitNames = names

		ars = append(ars, ar)
	}
	return ars, nil
}

func (db *DB) FindRepoContribution(repoId uuid.UUID) ([]Contributions, error) {
	var cs []Contributions

	rows, err := db.Query(
		`SELECT areq.date_from, areq.date_to, ares.git_email, ares.git_names, ares.weight
		 FROM analysis_request areq
		 INNER JOIN analysis_response ares ON areq.id = ares.analysis_request_id
		 WHERE areq.repo_id=$1 AND areq.error IS NULL 
		 ORDER BY areq.date_to, ares.weight DESC, ares.git_email`, 
		repoId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var c Contributions
		var jsonNames string
		err = rows.Scan(&c.DateFrom, &c.DateTo, &c.GitEmail, &jsonNames, &c.Weight)
		if err != nil {
			return nil, err
		}

		var names []string
		if err := json.Unmarshal([]byte(jsonNames), &names); err != nil {
			return nil, fmt.Errorf("unmarshal git_names: %w", err)
		}
		c.GitNames = names

		cs = append(cs, c)
	}
	return cs, nil
}

func (db *DB) FindRepoContributors(repoId uuid.UUID) (int, error) {
	var c int
	err := db.QueryRow(
		`SELECT count(distinct git_email) as c
		 FROM analysis_request areq
		 INNER JOIN analysis_response ares ON areq.id = ares.analysis_request_id
		 WHERE areq.repo_id=$1 AND areq.error IS NULL`, 
		repoId).Scan(&c)

	if err != nil {
		return 0, err
	}

	return c, nil
}