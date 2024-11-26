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
	if rows == nil {
		return healthMetrics, nil
	}

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

func GetInternalMetricsDummy() (*RepoHealthMetrics, error) {
	return &RepoHealthMetrics{
		SponsorCount:        rand.Intn(100) + 1,
		RepoStarCount:       rand.Intn(100) + 1,
		RepoMultiplierCount: rand.Intn(100) + 1,
	}, nil
}

func GetInternalMetrics(repoId uuid.UUID, isPostgres bool) (*RepoHealthMetrics, error) {
	var metrics RepoHealthMetrics
	var activeUserMinMonths = 1
	var latestRepoSponsoringMonths = 6

	activeSponsors, err := GetActiveSponsors(activeUserMinMonths, isPostgres)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active sponsors: %v", err)
	}

	// To add after pull request, because Multiplier is on different branch
	/*
		func GetMultiplierCount(repoId uuid.UUID, activeSponsors []uuid.UUID, isPostgres bool) (int, error) {
			if len(activeSponsors) == 0 {
				return 0, nil
			}

			var query string
			if isPostgres {
				query = `
					SELECT COUNT(DISTINCT user_id)
					FROM multiplier_event
					WHERE repo_id = $1 AND user_id = ANY($2) AND un_multiplier_at IS NULL`
			} else {
				query = `
					SELECT COUNT(DISTINCT user_id)
					FROM multiplier_event
					WHERE repo_id = ? AND user_id IN (?) AND un_multiplier_at IS NULL`
			}

			var args []interface{}
			if isPostgres {
				args = []interface{}{repoId, ConvertToInterfaceSlice(activeSponsors)}
			} else {
				args = append([]interface{}{repoId}, ConvertToInterfaceSlice(activeSponsors)...)
			}

			var count int
			err := DB.QueryRow(query, args...).Scan(&count)
			if err != nil {
				return 0, err
			}

			return count, nil
		}
	*/

	// multiplierCount, err := GetMultiplierCount(repoId, activeSponsors)
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		multiplierCount = 0
	// 	} else {
	// 		return nil, fmt.Errorf("failed to fetch multiplier count: %v", err)
	// 	}
	// }
	multiplierCount := 0
	metrics.RepoMultiplierCount = multiplierCount

	starCount, err := GetStarCount(repoId, activeSponsors, isPostgres)
	if err != nil {
		if err == sql.ErrNoRows {
			starCount = 0
		} else {
			return nil, fmt.Errorf("failed to fetch star count: %v", err)
		}
	}
	metrics.RepoStarCount = starCount

	repoWeight, err := GetRepoWeight(repoId, activeUserMinMonths, latestRepoSponsoringMonths, isPostgres)
	if err != nil {
		if err == sql.ErrNoRows {
			repoWeight = 0.0
		} else {
			return nil, fmt.Errorf("failed to calculate repo weight: %v", err)
		}
	}
	metrics.RepoWeight = repoWeight

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
