package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/rand"
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

// tested
func InsertRepoHealthMetrics(repoHealthMetrics RepoHealthMetrics) error {
	stmt, err := DB.Prepare(`
		INSERT INTO 
			repo_health_metrics (
				id,
				created_at,
				repo_id,
				contributer_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				repo_weight)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9)`)
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
	stmt, err := DB.Prepare(`
	UPDATE 
		repo_health_metrics 
	SET 
		contributer_count=$1,
		commit_count=$2,
		sponsor_donation=$3,
		repo_star_count=$4,
		repo_multiplier_count=$5,
		repo_weight=$6 
	WHERE 
		id=$7`)
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
		QueryRow(`
			SELECT id,
				repo_id,
				created_at,
				contributer_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				repo_weight 
			FROM 
				repo_health_metrics 
			WHERE 
				id=$1`, id).
		Scan(
			&healthValue.Id,
			&healthValue.RepoId,
			&healthValue.CreatedAt,
			&healthValue.ContributerCount,
			&healthValue.CommitCount,
			&healthValue.SponsorCount,
			&healthValue.RepoStarCount,
			&healthValue.RepoMultiplierCount,
			&healthValue.RepoWeight)
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
func FindRepoHealthMetricsByRepoId(repoId uuid.UUID) (*RepoHealthMetrics, error) {
	//var tv TrustValue
	rows, err := DB.
		Query(`
			SELECT 
				id,
				repo_id,
				created_at,
				contributer_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				repo_weight 
			FROM 
				repo_health_metrics 
			WHERE 
				repo_id=$1 
			ORDER BY 
				created_at DESC
			LIMIT 1`, repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	result, err := scanRepoHealthMetrics(rows)
	if err != nil {
		return nil, err
	}
	return &result[0], nil
}

func FindRepoHealthMetricsByRepoIdHistory(repoId uuid.UUID) ([]RepoHealthMetrics, error) {
	//var tv TrustValue
	rows, err := DB.
		Query(`
			SELECT 
				id,
				repo_id,
				created_at,
				contributer_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				repo_weight 
			FROM 
				repo_health_metrics 
			WHERE 
				repo_id=$1 
			ORDER BY 
				created_at DESC`, repoId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepoHealthMetrics(rows)
}

// tested by integration testing
func scanRepoHealthMetrics(rows *sql.Rows) ([]RepoHealthMetrics, error) {
	healthMetrics := []RepoHealthMetrics{}
	for rows.Next() {
		var repoHealthMetrics RepoHealthMetrics
		err := rows.Scan(&repoHealthMetrics.Id, &repoHealthMetrics.RepoId, &repoHealthMetrics.CreatedAt, &repoHealthMetrics.ContributerCount, &repoHealthMetrics.CommitCount, &repoHealthMetrics.SponsorCount, &repoHealthMetrics.RepoStarCount, &repoHealthMetrics.RepoMultiplierCount, &repoHealthMetrics.RepoWeight)
		if err != nil {
			return nil, err
		}
		healthMetrics = append(healthMetrics, repoHealthMetrics)
	}
	return healthMetrics, nil
}

// tested
func GetAllRepoHealthMetrics() ([]RepoHealthMetrics, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			repo_id,
			created_at,
			contributer_count,
			commit_count,
			sponsor_donation,
			repo_star_count,
			repo_multiplier_count,
			repo_weight
		FROM
			repo_health_metrics
		ORDER BY 
			created_at desc`)
	if err != nil {
		return nil, err
	}

	return scanRepoHealthMetrics(rows)
}

/*
Currently GetLatestThresholds doesn't work reliably
Values are missing
The query statements aren't working
Thus this dummy function will replace GetLatestThresholds for now
*/
func GetInternalMetricsDummy() (*RepoHealthMetrics, error) {
	return &RepoHealthMetrics{
		SponsorCount:        rand.Intn(100) + 1,
		RepoStarCount:       rand.Intn(100) + 1,
		RepoMultiplierCount: rand.Intn(100) + 1,
	}, nil
}

func GetInternalMetrics(repoId uuid.UUID, isPostgres bool) (*RepoHealthMetrics, error) {
	var internalHealthMetric RepoHealthMetrics
	var queryForMetricRepoStarMultiplier string
	var queryForMetricRepoWeight string
	err := DB.QueryRow(
		`SELECT
			DISTINCT user_sponsor_id AS total_sponsor_count
		FROM 
			daily_contribution
		WHERE 
			repo_id = $1
		GROUP BY 
			repo_id`, repoId).Scan(&internalHealthMetric.SponsorCount)
	if err != nil {
		if err == sql.ErrNoRows {
			internalHealthMetric.SponsorCount = 0
		} else {
			return nil, err
		}
	}

	if isPostgres { // For Postgres
		queryForMetricRepoStarMultiplier = `
		WITH active_sponsors AS (
			SELECT DISTINCT user_sponsor_id AS user_id
			FROM daily_contribution
			WHERE created_at >= CURRENT_DATE - INTERVAL '1 month'
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
				se.un_sponsor_at IS NULL
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
				me.un_multiplier_at IS NULL
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
			COALESCE(mr.repo_id, sr.repo_id) = $1;
		`
	} else { // For SQLite
		queryForMetricRepoStarMultiplier = `
		WITH active_sponsors AS (
			SELECT DISTINCT user_sponsor_id AS user_id
			FROM daily_contribution
			WHERE created_at >= date('now', '-1 month')
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
				se.un_sponsor_at IS NULL
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
				me.un_multiplier_at IS NULL
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
			(SELECT * FROM multiplied_repos WHERE repo_id = $1) mr
		LEFT JOIN
			(SELECT * FROM starred_repos WHERE repo_id = $1) sr ON mr.repo_id = sr.repo_id
		UNION
		SELECT
			COALESCE(mr.multiplier_count, 0) AS multiplier_count,
			COALESCE(sr.star_count, 0) AS star_count
		FROM
			(SELECT * FROM starred_repos WHERE repo_id = $1) sr
		LEFT JOIN
			(SELECT * FROM multiplied_repos WHERE repo_id = $1) mr ON sr.repo_id = mr.repo_id;
		`
	}

	err = DB.QueryRow(queryForMetricRepoStarMultiplier, repoId).
		Scan(&internalHealthMetric.RepoMultiplierCount, &internalHealthMetric.RepoStarCount)
	if err != nil {
		if err == sql.ErrNoRows {
			internalHealthMetric.RepoMultiplierCount = 0
			internalHealthMetric.RepoStarCount = 0
		} else {
			return nil, err
		}
	}

	if isPostgres { // Postgres
		queryForMetricRepoWeight = `
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
			dc.created_at >= CURRENT_DATE - INTERVAL '1 month'`
	} else { // SQLite query
		queryForMetricRepoWeight = `
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
			dc.created_at >= date('now', '-1 month')`
	}

	err = DB.QueryRow(queryForMetricRepoWeight, repoId).Scan(&internalHealthMetric.RepoWeight)
	if err != nil {
		if err == sql.ErrNoRows {
			internalHealthMetric.RepoWeight = 0.0
		} else {
			return nil, err
		}
	}

	return &internalHealthMetric, err
}
