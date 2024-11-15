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
	//first get last multiplier event to check if we need to InsertOrUpdateTrustRepo or unset the multiplier
	//TODO: use mutex
	id, multiplierAt, unMultiplierAt, err := FindLastEventMultiplierRepo(event.RepoId)
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

func FindLastEventMultiplierRepo(rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var multiplierAt *time.Time
	var unMultiplierAt *time.Time
	var id *uuid.UUID
	err := DB.
		QueryRow(`SELECT id, multiplier_at, un_multiplier_at
			      		FROM multiplier_event 
						WHERE repo_id=$1 
						ORDER by multiplier_at DESC LIMIT 1`,
			rid).Scan(&id, &multiplierAt, &unMultiplierAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return id, multiplierAt, unMultiplierAt, nil
	default:
		return nil, nil, nil, err
	}
}

func FindMultipliedRepos() ([]Repo, error) {
	//we want to send back an empty array, don't change
	t := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
            FROM multiplier_event t
            INNER JOIN repo r ON t.repo_id=r.id
			WHERE t.un_multiplier_at IS NULL`
	rows, err := DB.Query(t)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepo(rows)
}

//type TrustResult struct {
//	UserId  uuid.UUID
//	RepoIds []uuid.UUID
//}

// can be used to provide filter function in frontend
//func FindMultipliedReposBetween(start time.Time, stop time.Time) ([]TrustResult, error) {
//	rows, err := DB.Query(`
//			SELECT user_id, repo_id
//			FROM multiplier_event
//			WHERE multiplier_at < $1 AND (un_multiplier_at IS NULL OR un_multiplier_at >= $2)
//			GROUP BY user_id, repo_id
//			ORDER BY user_id`, start, stop)
//	if err != nil {
//		return nil, err
//	}
//	defer CloseAndLog(rows)
//
//	multiplierResults := []TrustResult{}
//	var userIdOld uuid.UUID
//	var userId uuid.UUID
//	multiplierResult := TrustResult{}
//	for rows.Next() {
//		var repoId uuid.UUID
//		err = rows.Scan(&userId, &repoId)
//
//		if userId != userIdOld && !util.IsUUIDZero(userIdOld) {
//			multiplierResult.UserId = userIdOld
//			multiplierResults = append(multiplierResults, multiplierResult)
//
//			multiplierResult = TrustResult{}
//			userIdOld = userId
//		}
//
//		if err != nil {
//			return nil, err
//		}
//		multiplierResult.RepoIds = append(multiplierResult.RepoIds, repoId)
//		if util.IsUUIDZero(userIdOld) {
//			userIdOld = userId
//		}
//	}
//
//	multiplierResult.UserId = userId
//	multiplierResults = append(multiplierResults, multiplierResult)
//
//	return multiplierResults, nil
//}
