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

type InternalHealthMetrics struct {
	SponsorCount        int     `json:"sponsorcount"`
	RepoStarCount       int     `json:"repostarcount"`
	RepoMultiplierCount int     `json:"repomultipliercount"`
	RepoWeight          float64 `json:"reposponsordonated"`
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

func GetInternalMetrics(repoId uuid.UUID) (*InternalHealthMetrics, error) {
	var internalHealthMetric InternalHealthMetrics
	rowsMetricSponsorDonation, err := DB.Query(
		`SELECT
			COUNT(DISTINCT user_sponsor_id) AS total_sponsor_count
		FROM 
			daily_contribution
		WHERE 
			repo_id = $1
		GROUP BY 
			repo_id`, repoId)
	if err != nil {
		return nil, err
	}

	err = rowsMetricSponsorDonation.Scan(&internalHealthMetric.SponsorCount)
	if err != nil {
		return nil, err
	}

	rowsMetricRepoStarMultiplier, err := DB.Query(
		`WITH active_sponsors AS (
			SELECT DISTINCT user_sponsor_id AS user_id
			FROM daily_contribution
			WHERE created_at  >= CURRENT_DATE - INTERVAL '1 month'
		),
		latest_sponsorships AS (
			SELECT 
				se.repo_id,
				se.user_id
			FROM 
				sponsor_event se
			JOIN 
				active_sponsors s ON se.user_id = s.user_id
			WHERE 
				AND se.un_sponsored_at IS NULL
		),
		latest_multipliers AS (
			SELECT 
				me.repo_id,
				me.user_id
			FROM 
				multiplier_event me
			JOIN 
				latest_sponsorships ls ON me.repo_id = ls.repo_id AND me.user_id = ls.user_id
			WHERE 
				AND me.un_multiplier_at IS NULL
		),
		multiplied_repos AS (
			SELECT 
				repo_id,
				COUNT(DISTINCT user_id) AS multiplier_count
			FROM 
				latest_multipliers
			GROUP BY 
				repo_id
		),
		starred_repos AS (
			SELECT 
				repo_id,
				COUNT(DISTINCT user_id) AS star_count
			FROM 
				latest_sponsorships
			GROUP BY 
				repo_id
		)
		SELECT 
			COALESCE(mr.multiplier_count, 0) AS multiplier_count,
			COALESCE(sr.star_count, 0) AS star_count
		FROM 
			multiplied_repos mr
		FULL OUTER JOIN 
			starred_repos sr ON mr.repo_id = sr.repo_id
		WHERE 
			COALESCE(mr.repo_id, sr.repo_id) = $1`, repoId)
	if err != nil {
		return nil, err
	}

	err = rowsMetricRepoStarMultiplier.Scan(&internalHealthMetric.RepoMultiplierCount, &internalHealthMetric.RepoStarCount)
	if err != nil {
		return nil, err
	}

	rowsMetricRepoWeight, err := DB.Query(`
		WITH user_repos AS (
			SELECT 
				r.id AS repo_id,
				ar.git_email AS repo_email
			FROM 
				repo r
			LEFT JOIN 
				analysis_response ar ON r.id = ar.repo_id
			WHERE 
				r.id = $1
		),
		user_projects AS (
			SELECT 
				ge.user_id,
				ur.repo_id
			FROM 
				git_email ge
			JOIN 
				user_repos ur ON ge.email = ur.repo_email
		),
		trusted_repos AS (
			SELECT 
				r.id AS repo_id
			FROM 
				trust_event t
			INNER JOIN 
				repo r ON t.repo_id = r.id
			WHERE 
				t.un_trust_at IS NULL
		)
		SELECT
			COUNT(DISTINCT tr.repo_id) AS trusted_project_count
		FROM
			user_projects up
		LEFT JOIN
			daily_contribution dc ON up.user_id = dc.user_sponsor_id AND up.repo_id = dc.repo_id
		LEFT JOIN
			trusted_repos tr ON tr.repo_id = dc.repo_id
		WHERE
			dc.created_at >= CURRENT_DATE - INTERVAL '1 month'`, repoId)
	if err != nil {
		return nil, err
	}

	err = rowsMetricRepoWeight.Scan(&internalHealthMetric.RepoWeight)
	if err != nil {
		return nil, err
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &internalHealthMetric, nil
	default:
		return nil, err
	}
}
