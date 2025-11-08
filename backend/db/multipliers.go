package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type MultiplierEvent struct {
	Id             uuid.UUID  `json:"id"`
	Uid            uuid.UUID  `json:"uid"`
	RepoId         uuid.UUID  `json:"repo_id"`
	EventType      uint8      `json:"event_type"`
	MultiplierAt   *time.Time `json:"multiplier_at"`
	UnMultiplierAt *time.Time `json:"un_multiplier_at"`
}

func (db *DB) InsertOrUpdateMultiplierRepo(event *MultiplierEvent) error {
	// First get last multiplier event to check if we need to insert or update
	id, multiplierAt, unMultiplierAt, err := db.FindLastEventMultiplierRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	// No event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("cannot unset multiplier: no active multiplier on this repo")
	}

	// Event found - validate
	if id != nil {
		if event.EventType == Inactive {
			if event.MultiplierAt != nil || event.UnMultiplierAt == nil {
				return fmt.Errorf("to unset multiplier, must set un_multiplier_at but not multiplier_at: "+
					"event.MultiplierAt: %v, event.UnMultiplierAt: %v", event.MultiplierAt, event.UnMultiplierAt)
			}
			if unMultiplierAt != nil {
				return fmt.Errorf("multiplier already unset: unMultiplierAt: %v", unMultiplierAt)
			}
			if unMultiplierAt == nil && event.UnMultiplierAt.Before(*multiplierAt) {
				return fmt.Errorf("un_multiplier_at cannot be before multiplier_at: multiplierAt: %v, event.UnMultiplierAt: %v",
					multiplierAt, event.UnMultiplierAt)
			}
		} else if event.EventType == Active {
			if event.MultiplierAt == nil || event.UnMultiplierAt != nil {
				return fmt.Errorf("to set multiplier, must set multiplier_at but not un_multiplier_at: "+
					"event.MultiplierAt: %v, event.UnMultiplierAt: %v", event.MultiplierAt, event.UnMultiplierAt)
			}
			if unMultiplierAt == nil {
				return fmt.Errorf("multiplier already active on this repo: multiplier_at: %v, un_multiplier_at: %v",
					event.MultiplierAt, unMultiplierAt)
			} else {
				if event.MultiplierAt.Before(*multiplierAt) {
					return fmt.Errorf("cannot set multiplier in the past: event.MultiplierAt: %v, multiplierAt: %v",
						event.MultiplierAt, multiplierAt)
				}
				if event.MultiplierAt.Before(*unMultiplierAt) {
					return fmt.Errorf("cannot set multiplier before un_multiplier_at: event.MultiplierAt: %v, unMultiplierAt: %v",
						event.MultiplierAt, unMultiplierAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}
	}

	if event.EventType == Active {
		_, err := db.Exec(
			`INSERT INTO multiplier_event (id, user_id, repo_id, multiplier_at) VALUES ($1, $2, $3, $4)`,
			event.Id, event.Uid, event.RepoId, event.MultiplierAt)
		return err
	} else if event.EventType == Inactive {
		_, err := db.Exec(
			`UPDATE multiplier_event SET un_multiplier_at=$1 WHERE id=$2 AND un_multiplier_at IS NULL`,
			event.UnMultiplierAt, id)
		return err
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func (db *DB) FindLastEventMultiplierRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var multiplierAt *time.Time
	var unMultiplierAt *time.Time
	var id uuid.UUID
	
	err := db.QueryRow(`
		SELECT id, multiplier_at, un_multiplier_at
		FROM multiplier_event 
		WHERE user_id=$1 AND repo_id=$2 
		ORDER BY multiplier_at DESC 
		LIMIT 1`, uid, rid).Scan(&id, &multiplierAt, &unMultiplierAt)
	
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return &id, multiplierAt, unMultiplierAt, nil
	default:
		return nil, nil, nil, err
	}
}

func (db *DB) FindMultiplierRepoByUserId(userId uuid.UUID) ([]Repo, error) {
	rows, err := db.Query(`
		SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
		FROM multiplier_event m
		INNER JOIN repo r ON m.repo_id=r.id
		WHERE m.user_id=$1 AND m.un_multiplier_at IS NULL`, userId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepos(rows)
}

func (db *DB) GetFoundationsSupportingRepo(rid uuid.UUID) ([]Foundation, error) {
	rows, err := db.Query(`
		SELECT u.id, u.multiplier_daily_limit
		FROM multiplier_event m
		INNER JOIN users u ON m.user_id = u.id
		WHERE m.repo_id=$1 AND u.multiplier AND m.un_multiplier_at IS NULL`, rid)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var foundations []Foundation
	for rows.Next() {
		var foundation Foundation
		err = rows.Scan(&foundation.Id, &foundation.MultiplierDailyLimit)
		if err != nil {
			return nil, err
		}
		foundations = append(foundations, foundation)
	}
	return foundations, nil
}

func (db *DB) GetValidatedFoundationsSupportingRepo(rid uuid.UUID, currency string, yesterdayStart time.Time) ([]Foundation, error) {
	rows, err := db.Query(`
		SELECT u.id, u.multiplier_daily_limit
		FROM multiplier_event m
		INNER JOIN users u ON m.user_id = u.id
		WHERE m.repo_id=$1 AND u.multiplier AND m.un_multiplier_at IS NULL`, rid)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var foundations []Foundation
	for rows.Next() {
		var foundation Foundation
		err = rows.Scan(&foundation.Id, &foundation.MultiplierDailyLimit)
		if err != nil {
			return nil, err
		}
		
		firstCheck, err := db.CheckDailyLimitStillAdheredTo(&foundation, big.NewInt(0), currency, yesterdayStart)
		if err != nil {
			return nil, err
		}

		secondCheck, err := db.CheckFondsAmountEnough(&foundation, big.NewInt(0), currency)
		if err != nil {
			return nil, err
		}

		if firstCheck.Cmp(big.NewInt(-1)) == 0 || secondCheck.Cmp(big.NewInt(-1)) == 0 {
			continue
		}

		foundations = append(foundations, foundation)
	}
	return foundations, nil
}

func (db *DB) GetAllFoundationsSupportingRepos(rids []uuid.UUID) ([]Foundation, int, error) {
	if len(rids) == 0 {
		return nil, 0, nil
	}

	rows, err := db.Query(`
		SELECT u.id, u.multiplier_daily_limit, m.repo_id
		FROM multiplier_event m
		INNER JOIN users u ON m.user_id = u.id
		WHERE m.repo_id = ANY($1) AND u.multiplier AND m.un_multiplier_at IS NULL`,
		pq.Array(rids))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	foundationMap := make(map[uuid.UUID]*Foundation)
	totalRepoCount := 0

	for rows.Next() {
		var foundationId uuid.UUID
		var repoId uuid.UUID
		var multiplierLimit int

		err = rows.Scan(&foundationId, &multiplierLimit, &repoId)
		if err != nil {
			return nil, 0, err
		}

		totalRepoCount++

		if foundation, exists := foundationMap[foundationId]; exists {
			if !contains(foundation.RepoIds, repoId) {
				foundation.RepoIds = append(foundation.RepoIds, repoId)
			}
		} else {
			foundationMap[foundationId] = &Foundation{
				Id:                   foundationId,
				MultiplierDailyLimit: multiplierLimit,
				RepoIds:              []uuid.UUID{repoId},
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	foundations := make([]Foundation, 0, len(foundationMap))
	for _, foundation := range foundationMap {
		foundations = append(foundations, *foundation)
	}

	return foundations, totalRepoCount, nil
}

func contains(slice []uuid.UUID, item uuid.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (db *DB) GetMultiplierCount(repoId uuid.UUID, activeSponsors []uuid.UUID) (int, error) {
	if len(activeSponsors) == 0 {
		return 0, nil
	}

	var count int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT user_id)
		FROM multiplier_event
		WHERE repo_id = $1 AND user_id = ANY($2) AND un_multiplier_at IS NULL`,
		repoId, pq.Array(activeSponsors)).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}