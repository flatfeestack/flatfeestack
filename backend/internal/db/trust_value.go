package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TrustValue struct {
	Id               uuid.UUID `json:"uuid"`
	RepoId           uuid.UUID `json:"uuid"`
	CreatedAt        time.Time `json:"createdAt"`
	ContributerCount float32   `json:"contributerCount"`
	CommitCount      float32   `json:"commitCount"`
	MetricOne        float32   `json:"metricOne"`
	MetricTwo        float32   `json:"metricTwo"`
	MetricThree      float32   `json:"metricThree"`
}

func InsertOrUpdateTrustValue(repo *Repo) error {
	stmt, err := DB.Prepare(`INSERT INTO repo (id, repo_id, created_at, contributer_count, commit_count, metric_3, metric_4, metric_5)
								   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
								   RETURNING id`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer CloseAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(repo.Id, repo.Url, repo.GitUrl, repo.Name, repo.Description, repo.Score, repo.Source, repo.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	repo.Id = lastInsertId
	return nil
}

func FindTrustValueById(id uuid.UUID) (*TrustValue, error) {
	var tv TrustValue
	err := DB.
		QueryRow("SELECT id, repo_id, created_at, contributer_count, commit_count, metric_3, metric_4, metric_5 from trust_value WHERE id=$1", id).
		Scan(&tv.Id, &tv.RepoId, &tv.CreatedAt, &tv.ContributerCount, &tv.CommitCount, &tv.MetricOne, &tv.MetricTwo, &tv.MetricThree)
	if err != nil {
		return nil, err
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &tv, nil
	default:
		return nil, err
	}
}

func FindTrustValueByRepoId(repoId uuid.UUID) ([]TrustValue, error) {
	//var tv TrustValue
	rows, err := DB.
		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, metric_3, metric_4, metric_5 from trust_value WHERE repo_id=$1 order by created_at desc limit 1", repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanTrustValue(rows)
}

func scanTrustValue(rows *sql.Rows) ([]TrustValue, error) {
	trustValues := []TrustValue{}
	for rows.Next() {
		var tv TrustValue
		err := rows.Scan(&tv.Id, &tv.RepoId, &tv.CreatedAt, &tv.ContributerCount, &tv.CommitCount, &tv.MetricOne, &tv.MetricTwo, &tv.MetricThree)
		if err != nil {
			return nil, err
		}
		trustValues = append(trustValues, tv)
	}
	return trustValues, nil
}
