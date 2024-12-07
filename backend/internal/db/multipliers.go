package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MultiplierEvent struct {
	Id             uuid.UUID  `json:"id"`
	Uid            uuid.UUID  `json:"uid"`
	RepoId         uuid.UUID  `json:"repo_id"`
	EventType      uint8      `json:"event_type"`
	MultiplierAt   *time.Time `json:"multiplier_at"`
	UnMultiplierAt *time.Time `json:"un_multiplier_at"`
}

func InsertOrUpdateMultiplierRepo(event *MultiplierEvent) error {
	//first get last multiplier event to check if we need to InsertOrUpdateMultiplierRepo or unset the multiplier
	//TODO: use mutex
	id, multiplierAt, unMultiplierAt, err := FindLastEventMultiplierRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	//no event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to unset multiplier, but we are currently not having a multilier on this repo")
	}

	//event found
	if id != nil {
		if event.EventType == Inactive {
			if event.MultiplierAt != nil || event.UnMultiplierAt == nil {
				return fmt.Errorf("to unset multiplier, we must set the unset multiplier_at, but not the multiplier_at: "+
					"event.MultiplierAt: %v, event.UnMultiplierAt: %v", event.MultiplierAt, event.UnMultiplierAt)
			}
			if unMultiplierAt != nil {
				return fmt.Errorf("we want to unset multiplier, but we already unset multipliered it: unMultiplierAt: %v", unMultiplierAt)
			}
			if unMultiplierAt == nil && event.UnMultiplierAt.Before(*multiplierAt) {
				return fmt.Errorf("we want to unset multiplier, but the unset multiplier date is before the multiplier date: multiplierAt: "+
					"%v, event.UnMultiplierAt: %v", multiplierAt, event.UnMultiplierAt)
			}
		} else if event.EventType == Active {
			if event.MultiplierAt == nil || event.UnMultiplierAt != nil {
				return fmt.Errorf("to set multiplier, we must set the multiplier_at, but not the unset multiplier_at: "+
					"event.MultiplierAt: %v, event.UnMultiplierAt: %v", event.MultiplierAt, event.UnMultiplierAt)
			}
			if unMultiplierAt == nil {
				return fmt.Errorf("we want to set multiplier, but we are already having a multilier on this repo: "+
					"multiplier_at: %v, un_multiplier_at: %v", event.MultiplierAt, unMultiplierAt)
			} else {
				if event.MultiplierAt.Before(*multiplierAt) {
					return fmt.Errorf("we want to set multiplier, but we want want to set a multiplier on this repo in the past: "+
						"event.MultiplierAt: %v, multiplierAt: %v", event.MultiplierAt, multiplierAt)
				}
				if event.MultiplierAt.Before(*unMultiplierAt) {
					return fmt.Errorf("we want to set multiplier, but we want want to set a multiplier on this repo in the past: "+
						"event.MultiplierAt: %v, unMultiplierAt: %v", event.MultiplierAt, unMultiplierAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}

	}

	if event.EventType == Active {
		//insert
		stmt, err := DB.Prepare("INSERT INTO multiplier_event (id, user_id, repo_id, multiplier_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return fmt.Errorf("prepare INSERT INTO multiplier_event for %v statement event: %v", event, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.MultiplierAt)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else if event.EventType == Inactive {
		//update
		stmt, err := DB.Prepare("UPDATE multiplier_event SET un_multiplier_at=$1 WHERE id=$2 AND un_multiplier_at IS NULL")
		if err != nil {
			return fmt.Errorf("prepare UPDATE multiplier_event for %v statement failed: %v", id, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.UnMultiplierAt, id)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func FindLastEventMultiplierRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var multiplierAt *time.Time
	var unMultiplierAt *time.Time
	var id *uuid.UUID
	err := DB.
		QueryRow(`SELECT id, multiplier_at, un_multiplier_at
			      		FROM multiplier_event 
						WHERE user_id=$1 AND repo_id=$2 
						ORDER by multiplier_at DESC LIMIT 1`,
			uid, rid).Scan(&id, &multiplierAt, &unMultiplierAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return id, multiplierAt, unMultiplierAt, nil
	default:
		return nil, nil, nil, err
	}
}

func FindMultiplierRepoByUserId(userId uuid.UUID) ([]Repo, error) {
	s := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
            FROM multiplier_event m
            INNER JOIN repo r ON m.repo_id=r.id
			WHERE m.user_id=$1 AND m.un_multiplier_at IS NULL`
	rows, err := DB.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepo(rows)
}

func GetFoundationsSupportingRepo(rid uuid.UUID) ([]Foundation, error) {
	s := `SELECT u.id, u.multiplier_daily_limit
            FROM multiplier_event m
			INNER JOIN users u ON m.user_id = u.id
			WHERE m.repo_id=$1 AND u.multiplier AND m.un_multiplier_at IS NULL`
	rows, err := DB.Query(s, rid)
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

func GetAllFoundationsSupportingRepos(rids []uuid.UUID) ([]Foundation, int, error) {
	if len(rids) == 0 {
		return nil, 0, nil
	}

	query := `
		SELECT u.id, u.multiplier_daily_limit, m.repo_id
		FROM multiplier_event m
		INNER JOIN users u ON m.user_id = u.id
		WHERE m.repo_id IN (` + GeneratePlaceholders(len(rids)) + `)
		AND u.multiplier
		AND m.un_multiplier_at IS NULL`

	rows, err := DB.Query(query, ConvertToInterfaceSlice(rids)...)
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
