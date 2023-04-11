package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func InsertAnalysisRequest(a AnalysisRequest, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_request(id, repo_id, date_from, date_to, git_url, created_at) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_request for %v statement event: %v", a.Id, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(a.Id, a.RepoId, a.DateFrom, a.DateTo, a.GitUrl, now.Format("2006-02-01"))
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

	stmt, err := db.Prepare(`INSERT INTO analysis_response(
                                     id, analysis_request_id, git_email, git_names, weight, created_at) 
									 VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v statement event: %v", reqId, err)
	}
	defer closeAndLog(stmt)

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
	err := db.
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

	rows, err := db.Query(`SELECT id, repo_id, date_from, date_to, git_url, received_at, error FROM (
                          SELECT id, repo_id, date_from, date_to, git_url, received_at, error,
                            RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
                            FROM analysis_request) AS x
                        WHERE dest_rank = 1 AND date_to <= $1
                        ORDER BY git_url`, dateTo.Format("2006-02-01"))

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

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
	//stmt, err := db.Prepare(`UPDATE analysis_request set received_at = $2, error = $3 WHERE id = $1`)
	stmt, err := db.Prepare(`UPDATE analysis_request set received_at = $1, error = $2 WHERE id = $3`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE analysis_request for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, errStr, reqId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindAnalysisResults(reqId uuid.UUID) ([]AnalysisResponse, error) {
	var ars []AnalysisResponse

	rows, err := db.Query(`SELECT id, git_email, git_names, weight
                                 FROM analysis_response 
                                 WHERE analysis_request_id = $1`, reqId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

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
