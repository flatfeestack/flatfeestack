package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Threshold struct {
	Upper int `json:"upper"`
	Lower int `json:"lower"`
}

type RepoHealthThreshold struct {
	Id                 uuid.UUID  `db:"id"`
	CreatedAt          time.Time  `db:"created_at"`
	ThContributerCount *Threshold `db:"th_contributer_count" validate:"required"`
	ThCommitCount      *Threshold `db:"th_commit_count" validate:"required"`
	ThSponsorDonation  *Threshold `db:"th_sponsor_donation" validate:"required"`
	ThRepoStarCount    *Threshold `db:"th_repo_star_count" validate:"required"`
	ThRepoMultiplier   *Threshold `db:"th_repo_multiplier" validate:"required"`
}

/*
Marshal not tested for the reason taht RepoHealthThreshold already restricted
to types that can't trigger an error.
*/
func InsertRepoHealthThreshold(threshold RepoHealthThreshold) error {
	if threshold.ThCommitCount == nil || threshold.ThContributerCount == nil || threshold.ThRepoMultiplier == nil || threshold.ThRepoStarCount == nil || threshold.ThSponsorDonation == nil {
		return fmt.Errorf("Threshold values can't be empty, aborting")
	}

	contributerJSON, err := json.Marshal(threshold.ThContributerCount)
	if err != nil {
		return fmt.Errorf("error marshaling contributer threshold: %w", err)
	}

	commitJSON, err := json.Marshal(threshold.ThCommitCount)
	if err != nil {
		return fmt.Errorf("error marshaling commit threshold: %w", err)
	}

	sponsorJSON, err := json.Marshal(threshold.ThSponsorDonation)
	if err != nil {
		return fmt.Errorf("error marshaling sponsor threshold: %w", err)
	}

	starJSON, err := json.Marshal(threshold.ThRepoStarCount)
	if err != nil {
		return fmt.Errorf("error marshaling star threshold: %w", err)
	}

	multiplierJSON, err := json.Marshal(threshold.ThRepoMultiplier)
	if err != nil {
		return fmt.Errorf("error marshaling multiplier threshold: %w", err)
	}

	query := `
		INSERT INTO 
			repo_health_threshold (
      	id,
      	created_at,
      	th_contributer_count,
      	th_commit_count,
      	th_sponsor_donation,
      	th_repo_star_count,
      	th_repo_multiplier)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7)`

	res, err := DB.Exec(query,
		threshold.Id,
		threshold.CreatedAt,
		contributerJSON,
		commitJSON,
		sponsorJSON,
		starJSON,
		multiplierJSON,
	)

	if err != nil {
		return fmt.Errorf("error inserting repo health threshold: %w", err)
	}

	return handleErrMustInsertOne(res)
}

/*
No negative testing as a base value is loaded when the database is initialized
*/
func GetFirstRepoHealthThreshold() (*RepoHealthThreshold, error) {
	query := `
		SELECT 
			id,                                          	
			created_at,                                  	
			th_contributer_count,                        	
			th_commit_count,                             	
			th_sponsor_donation,                         	
			th_repo_star_count,                          	
			th_repo_multiplier                           	
		FROM                                          	
			repo_health_threshold                        	
		ORDER BY                                      	
			created_at ASC
		LIMIT 1`

	result, err := executeRepoThresholdQuery(query)
	if err != nil {
		return nil, err
	}
	return &result[0], nil
}

func GetLatestThresholds() (*RepoHealthThreshold, error) {
	query := `
		SELECT 
			id,
			created_at,
			th_contributer_count,
			th_commit_count,
			th_sponsor_donation,
			th_repo_star_count,
			th_repo_multiplier
		FROM 
			repo_health_threshold
		ORDER BY 
			created_at DESC
		LIMIT 1`

	result, err := executeRepoThresholdQuery(query)
	if err != nil {
		return nil, err
	}
	return &result[0], nil
}

func GetRepoThresholdHistory() ([]RepoHealthThreshold, error) {
	query := `
    SELECT 
      id,
      created_at,
      th_contributer_count,
      th_commit_count,
      th_sponsor_donation,
      th_repo_star_count,
      th_repo_multiplier
    FROM 
			repo_health_threshold`

	return executeRepoThresholdQuery(query)
}

/*
Both functions from here on are tested via other autoamted acceptance tests
*/
func executeRepoThresholdQuery(query string) ([]RepoHealthThreshold, error) {
	var repoHealthThresholds []RepoHealthThreshold

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var time time.Time
		var contributerRawJSON, commitRawJSON, sponsorRawJSON, starRawJSON, multiplierRawJSON []byte
		err := rows.Scan(
			&id,
			&time,
			&contributerRawJSON,
			&commitRawJSON,
			&sponsorRawJSON,
			&starRawJSON,
			&multiplierRawJSON,
		)
		if err != nil {
			return nil, err
		}

		row := byteToJson(id, time, contributerRawJSON, commitRawJSON, sponsorRawJSON, starRawJSON, multiplierRawJSON)
		repoHealthThresholds = append(repoHealthThresholds, *row)
	}

	return repoHealthThresholds, nil
}

func byteToJson(id uuid.UUID, time time.Time, contrib, commit, sponsor, star, multi []byte) *RepoHealthThreshold {
	var contributerJSON, commitJSON, sponsorJSON, starJSON, multiplierJSON Threshold

	json.Unmarshal([]byte(contrib), &contributerJSON)
	json.Unmarshal([]byte(commit), &commitJSON)
	json.Unmarshal([]byte(sponsor), &sponsorJSON)
	json.Unmarshal([]byte(star), &starJSON)
	json.Unmarshal([]byte(multi), &multiplierJSON)

	row := RepoHealthThreshold{
		Id:                 id,
		CreatedAt:          time,
		ThContributerCount: &contributerJSON,
		ThCommitCount:      &commitJSON,
		ThSponsorDonation:  &sponsorJSON,
		ThRepoStarCount:    &starJSON,
		ThRepoMultiplier:   &multiplierJSON,
	}

	return &row
}
