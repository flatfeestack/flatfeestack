package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RepoHealthMetrics struct {
	Id                  uuid.UUID `json:"id"`
	RepoId              uuid.UUID `json:"repoid"`
	CreatedAt           time.Time `json:"createdat"`
	ContributerCount    int       `json:"contributercount"`
	CommitCount         int       `json:"commitcount"`
	SponsorCount        int       `json:"sponsorcount"`
	RepoStarCount       int       `json:"repostarcount"`
	RepoMultiplierCount int       `json:"repomultipliercount"`
	RepoWeight          float64   `json:"reposponsordonated"`
}

// This works, do you understand, Mino?
func InsertTrustValue(trustValueMetric RepoHealthMetrics) error {
	stmt, err := DB.Prepare("INSERT INTO trust_value_metrics (repo_id, contributer_count, commit_count, sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO trust_value_metrics for %v statement event: %v", trustValueMetric, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(trustValueMetric.Id, trustValueMetric.RepoId, trustValueMetric.ContributerCount, trustValueMetric.CommitCount, trustValueMetric.SponsorCount, trustValueMetric.RepoStarCount, trustValueMetric.RepoMultiplierCount, trustValueMetric.RepoWeight)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func UpdateTrustValue(trustValueMetric RepoHealthMetrics) error {
	stmt, err := DB.Prepare("UPDATE trust_value_metrics SET repo_id=$1, contributer_count=$2, commit_count=$3, sponsor_donation=$4, sponsor_star_multiplier=$5, repo_sponsor_donated=$6) WHERE id=$7")
	if err != nil {
		return fmt.Errorf("prepare UPDATE trust_value_metrics for %v statement failed: %v", trustValueMetric, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(trustValueMetric.Id, trustValueMetric.RepoId, trustValueMetric.ContributerCount, trustValueMetric.CommitCount, trustValueMetric.SponsorCount, trustValueMetric.RepoStarCount, trustValueMetric.RepoMultiplierCount, trustValueMetric.RepoWeight)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

func FindTrustValueById(id uuid.UUID) (*RepoHealthMetrics, error) {
	var trustValue RepoHealthMetrics
	err := DB.
		QueryRow("SELECT id, repo_id, created_at, contributer_count, commit_count,  sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value WHERE id=$1", id).
		Scan(&trustValue.Id, &trustValue.RepoId, &trustValue.CreatedAt, &trustValue.ContributerCount, &trustValue.CommitCount, &trustValue.SponsorCount, &trustValue.RepoStarCount, &trustValue.RepoMultiplierCount)
	if err != nil {
		return nil, err
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &trustValue, nil
	default:
		return nil, err
	}
}

func FindTrustValueByRepoId(repoId uuid.UUID) ([]RepoHealthMetrics, error) {
	//var tv TrustValue
	rows, err := DB.
		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value WHERE repo_id=$1 order by created_at desc limit 1", repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanTrustValue(rows)
}

func scanTrustValue(rows *sql.Rows) ([]RepoHealthMetrics, error) {
	trustValues := []RepoHealthMetrics{}
	for rows.Next() {
		var tv RepoHealthMetrics
		err := rows.Scan(&tv.Id, &tv.RepoId, &tv.CreatedAt, &tv.ContributerCount, &tv.CommitCount, &tv.SponsorCount, &tv.RepoStarCount, &tv.RepoMultiplierCount, &tv.RepoWeight)
		if err != nil {
			return nil, err
		}
		trustValues = append(trustValues, tv)
	}
	return trustValues, nil
}

func GetAllTrustValues() ([]RepoHealthMetrics, error) {
	rows, err := DB.
		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, sponsor_star_multiplier, repo_sponsor_donated from trust_value order by created_at desc")
	if err != nil {
		return nil, err
	}

	return scanTrustValue(rows)
}
