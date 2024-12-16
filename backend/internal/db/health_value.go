package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

type RepoHealthMetrics struct {
	Id                  uuid.UUID `json:"id"`
	RepoId              uuid.UUID `json:"repoid"`
	CreatedAt           time.Time `json:"createdat"`
	ContributorCount    int       `json:"contributorcount"`
	CommitCount         int       `json:"commitcount"`
	SponsorCount        int       `json:"sponsorcount"`
	RepoStarCount       int       `json:"repostarcount"`
	RepoMultiplierCount int       `json:"repomultipliercount"`
	ActiveFFSUserCount  int       `json:"activeffsusercount"`
}

type InternalHealthMetrics struct {
	SponsorCount        int `json:"sponsorcount"`
	RepoStarCount       int `json:"repostarcount"`
	RepoMultiplierCount int `json:"repomultipliercount"`
	ActiveFFSUserCount  int `json:"reposponsordonated"`
}

func InsertRepoHealthMetrics(repoHealthMetrics RepoHealthMetrics) error {
	stmt, err := DB.Prepare(`
		INSERT INTO 
			repo_health_metrics (
				id,
				created_at,
				repo_id,
				contributor_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				active_ffs_user_count)
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

	res, err := stmt.Exec(repoHealthMetrics.Id, repoHealthMetrics.CreatedAt, repoHealthMetrics.RepoId, repoHealthMetrics.ContributorCount, repoHealthMetrics.CommitCount, repoHealthMetrics.SponsorCount, repoHealthMetrics.RepoStarCount, repoHealthMetrics.RepoMultiplierCount, repoHealthMetrics.ActiveFFSUserCount)
	if err != nil {
		return fmt.Errorf("error occured trying to insert: %v", err)
	}

	return handleErrMustInsertOne(res)
}

// tested
func UpdateRepoHealthMetrics(repoHealthMetrics RepoHealthMetrics) error {
	stmt, err := DB.Prepare(`
	UPDATE 
		repo_health_metrics 
	SET 
		contributor_count=$1,
		commit_count=$2,
		sponsor_donation=$3,
		repo_star_count=$4,
		repo_multiplier_count=$5,
		active_ffs_user_count=$6 
	WHERE 
		id=$7`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE repo_health_metrics for %v statement failed: %v", repoHealthMetrics, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(repoHealthMetrics.ContributorCount, repoHealthMetrics.CommitCount, repoHealthMetrics.SponsorCount, repoHealthMetrics.RepoStarCount, repoHealthMetrics.RepoMultiplierCount, repoHealthMetrics.ActiveFFSUserCount, repoHealthMetrics.Id)
	if err != nil {
		return fmt.Errorf("something went wrong updating the health value: %v", err)
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
				contributor_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				active_ffs_user_count 
			FROM 
				repo_health_metrics 
			WHERE 
				id=$1`, id).
		Scan(
			&healthValue.Id,
			&healthValue.RepoId,
			&healthValue.CreatedAt,
			&healthValue.ContributorCount,
			&healthValue.CommitCount,
			&healthValue.SponsorCount,
			&healthValue.RepoStarCount,
			&healthValue.RepoMultiplierCount,
			&healthValue.ActiveFFSUserCount)
	if err != nil {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &healthValue, err
}

// done
func FindRepoHealthMetricsByRepoId(repoId uuid.UUID) (*RepoHealthMetrics, error) {
	//var tv TrustValue
	if repoId == uuid.Nil {
		return nil, fmt.Errorf("repoId is empty")
	}
	rows, err := DB.
		Query(`
			SELECT 
				id,
				repo_id,
				created_at,
				contributor_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				active_ffs_user_count 
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

	if len(result) == 0 {
		return nil, nil
	}
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &result[0], nil
	default:
		return nil, err
	}
}

func FindRepoHealthMetricsByRepoIdHistory(repoId uuid.UUID) ([]RepoHealthMetrics, error) {
	rows, err := DB.
		Query(`
			SELECT 
				id,
				repo_id,
				created_at,
				contributor_count,
				commit_count,
				sponsor_donation,
				repo_star_count,
				repo_multiplier_count,
				active_ffs_user_count 
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
	if rows == nil {
		return healthMetrics, nil
	}

	for rows.Next() {
		var repoHealthMetrics RepoHealthMetrics
		err := rows.Scan(&repoHealthMetrics.Id, &repoHealthMetrics.RepoId, &repoHealthMetrics.CreatedAt, &repoHealthMetrics.ContributorCount, &repoHealthMetrics.CommitCount, &repoHealthMetrics.SponsorCount, &repoHealthMetrics.RepoStarCount, &repoHealthMetrics.RepoMultiplierCount, &repoHealthMetrics.ActiveFFSUserCount)
		if err != nil {
			return nil, err
		}
		healthMetrics = append(healthMetrics, repoHealthMetrics)
	}

	return healthMetrics, nil
}

func GetAllRepoHealthMetrics() ([]RepoHealthMetrics, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			repo_id,
			created_at,
			contributor_count,
			commit_count,
			sponsor_donation,
			repo_star_count,
			repo_multiplier_count,
			active_ffs_user_count
		FROM
			repo_health_metrics
		ORDER BY 
			created_at desc`)
	if err != nil {
		return nil, err
	}

	return scanRepoHealthMetrics(rows)
}

func GetInternalMetricsDummy() (*RepoHealthMetrics, error) {
	return &RepoHealthMetrics{
		SponsorCount:        rand.Intn(100) + 1,
		RepoStarCount:       rand.Intn(100) + 1,
		RepoMultiplierCount: rand.Intn(100) + 1,
		ActiveFFSUserCount:  rand.Intn(100) + 1,
	}, nil
}

func GetInternalMetrics(repoId uuid.UUID, isPostgres bool) (*RepoHealthMetrics, error) {
	var metrics RepoHealthMetrics
	var activeUserMinMonths = 3
	var latestRepoSponsoringMonths = 6

	activeSponsors, err := GetActiveSponsors(activeUserMinMonths, isPostgres)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active sponsors: %v", err)
	}

	multiplierCount, err := GetMultiplierCount(repoId, activeSponsors, isPostgres)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("no multiplier count available",
				slog.Any("error", err))
			multiplierCount = 0
		} else {
			slog.Error("couldn't query Multiplier Count",
				slog.Any("error", err))
			return nil, fmt.Errorf("failed to fetch multiplier count: %v", err)
		}
	}
	metrics.RepoMultiplierCount = multiplierCount

	starCount, err := GetStarCount(repoId, activeSponsors, isPostgres)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("no star count available",
				slog.Any("error", err))
			starCount = 0
		} else {
			slog.Error("couldn't query Star Count",
				slog.Any("error", err))
			return nil, fmt.Errorf("failed to fetch star count: %v", err)
		}
	}
	metrics.RepoStarCount = starCount

	activeFFSUserCount, err := GetActiveFFSUserCount(repoId, activeUserMinMonths, latestRepoSponsoringMonths, isPostgres)
	if err != nil {
		if err == sql.ErrNoRows {
			activeFFSUserCount = 0
		} else {
			slog.Error("couldn't query repo weight",
				slog.Any("error", err))
			return nil, fmt.Errorf("failed to calculate active ffs user count: %v", err)
		}
	}

	metrics.ActiveFFSUserCount = activeFFSUserCount

	err = DB.QueryRow(
		`SELECT COUNT(DISTINCT user_sponsor_id)
		 FROM daily_contribution
		 WHERE repo_id = $1
		 GROUP BY repo_id`, repoId).Scan(&metrics.SponsorCount)
	if err != nil {
		if err == sql.ErrNoRows {
			metrics.SponsorCount = 0
		} else {
			return nil, err
		}
	}

	return &metrics, nil
}
