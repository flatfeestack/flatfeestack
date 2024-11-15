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

// tested
func InsertRepoHealthMetrics(repoHealthMetrics RepoHealthMetrics) error {
	stmt, err := DB.Prepare(`INSERT INTO repo_health_metrics 
		(id, created_at, repo_id, contributer_count, commit_count, sponsor_donation, repo_star_count, repo_multiplier_count, repo_weight) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo_health_metrics for %v statement event: %v", repoHealthMetrics, err)
	}
	defer CloseAndLog(stmt)

	res, err := stmt.Exec(repoHealthMetrics.Id, repoHealthMetrics.CreatedAt, repoHealthMetrics.RepoId, repoHealthMetrics.ContributerCount, repoHealthMetrics.CommitCount, repoHealthMetrics.SponsorCount, repoHealthMetrics.RepoStarCount, repoHealthMetrics.RepoMultiplierCount, repoHealthMetrics.RepoWeight)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

// tested
func UpdateRepoHealthMetrics(repoHealthMetrics RepoHealthMetrics) error {
	stmt, err := DB.Prepare("UPDATE repo_health_metrics SET contributer_count=$1, commit_count=$2, sponsor_donation=$3, repo_star_count=$4, repo_multiplier_count=$5, repo_weight=$6 WHERE id=$7")
	if err != nil {
		return fmt.Errorf("prepare UPDATE repo_health_metrics for %v statement failed: %v", repoHealthMetrics, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(repoHealthMetrics.ContributerCount, repoHealthMetrics.CommitCount, repoHealthMetrics.SponsorCount, repoHealthMetrics.RepoStarCount, repoHealthMetrics.RepoMultiplierCount, repoHealthMetrics.RepoWeight, repoHealthMetrics.Id)
	if err != nil {
		return err
	}

	return handleErrMustInsertOne(res)
}

// tested
func FindRepoHealthMetricsById(id uuid.UUID) (*RepoHealthMetrics, error) {
	var healthValue RepoHealthMetrics
	err := DB.
		QueryRow("SELECT id, repo_id, created_at, contributer_count, commit_count,  sponsor_donation, repo_star_count, repo_multiplier_count, repo_weight from repo_health_metrics WHERE id=$1", id).
		Scan(&healthValue.Id, &healthValue.RepoId, &healthValue.CreatedAt, &healthValue.ContributerCount, &healthValue.CommitCount, &healthValue.SponsorCount, &healthValue.RepoStarCount, &healthValue.RepoMultiplierCount, &healthValue.RepoWeight)
	if err != nil {
		return nil, err
	}
	fmt.Println(healthValue)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &healthValue, err
}

// done
func FindRepoHealthMetricsByRepoId(repoId uuid.UUID) ([]RepoHealthMetrics, error) {
	//var tv TrustValue
	rows, err := DB.
		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, repo_star_count, repo_multiplier_count, repo_weight from repo_health_metrics WHERE repo_id=$1 order by created_at desc", repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanTrustValue(rows)
}

// tested by integration testing
func scanTrustValue(rows *sql.Rows) ([]RepoHealthMetrics, error) {
	trustValues := []RepoHealthMetrics{}
	for rows.Next() {
		var repoHealthMetrics RepoHealthMetrics
		err := rows.Scan(&repoHealthMetrics.Id, &repoHealthMetrics.RepoId, &repoHealthMetrics.CreatedAt, &repoHealthMetrics.ContributerCount, &repoHealthMetrics.CommitCount, &repoHealthMetrics.SponsorCount, &repoHealthMetrics.RepoStarCount, &repoHealthMetrics.RepoMultiplierCount, &repoHealthMetrics.RepoWeight)
		if err != nil {
			return nil, err
		}
		trustValues = append(trustValues, repoHealthMetrics)
	}
	return trustValues, nil
}

// tested
func GetAllTrustValues() ([]RepoHealthMetrics, error) {
	rows, err := DB.
		Query("SELECT id, repo_id, created_at, contributer_count, commit_count, sponsor_donation, repo_star_count, repo_multiplier_count, repo_weight from repo_health_metrics order by created_at desc")
	if err != nil {
		return nil, err
	}

	return scanTrustValue(rows)
}
