package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TrustEvent struct {
	Id        uuid.UUID  `json:"id"`
	Uid       uuid.UUID  `json:"uid"`
	RepoId    uuid.UUID  `json:"repo_id"`
	EventType uint8      `json:"event_type"`
	TrustAt   *time.Time `json:"trust_at"`
	UnTrustAt *time.Time `json:"un_trust_at"`
}

func InsertOrUpdateTrustRepo(event *TrustEvent) error {
	//first get last trusted event to check if we need to InsertOrUpdateTrustRepo or untrust
	//TODO: use mutex
	id, trustAt, unTrustAt, err := FindLastEventTrustedRepo(event.RepoId)
	if err != nil {
		return err
	}

	//no event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to untrust, but we are currently not trusting this repo")
	}

	//event found
	if id != nil {
		if event.EventType == Inactive {
			if event.TrustAt != nil || event.UnTrustAt == nil {
				return fmt.Errorf("to untrust, we must set the untrust_at, but not the trust_at: "+
					"event.TrustAt: %v, event.UnTrustAt: %v", event.TrustAt, event.UnTrustAt)
			}
			if unTrustAt != nil {
				return fmt.Errorf("we want to untrust, but we already untrusted it: unTrustAt: %v", unTrustAt)
			}
			if unTrustAt == nil && event.UnTrustAt.Before(*trustAt) {
				return fmt.Errorf("we want to untrust, but the untrust date is before the trust date: trustAt: "+
					"%v, event.UnTrustAt: %v", trustAt, event.UnTrustAt)
			}
		} else if event.EventType == Active {
			if event.TrustAt == nil || event.UnTrustAt != nil {
				return fmt.Errorf("to trust, we must set the trust_at, but not the untrust_at: "+
					"event.TrustAt: %v, event.UnTrustAt: %v", event.TrustAt, event.UnTrustAt)
			}
			if unTrustAt == nil {
				return fmt.Errorf("we want to trust, but we are already trusting this repo: "+
					"trust_at: %v, un_trust_at: %v", event.TrustAt, unTrustAt)
			} else {
				if event.TrustAt.Before(*trustAt) {
					return fmt.Errorf("we want to trust, but we want want to trust this repo in the past: "+
						"event.TrustAt: %v, trustAt: %v", event.TrustAt, trustAt)
				}
				if event.TrustAt.Before(*unTrustAt) {
					return fmt.Errorf("we want to trust, but we want want to trust this repo in the past: "+
						"event.TrustAt: %v, unTrustAt: %v", event.TrustAt, unTrustAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}

	}

	if event.EventType == Active {
		//insert
		stmt, err := DB.Prepare("INSERT INTO trust_event (id, user_id, repo_id, trust_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return fmt.Errorf("prepare INSERT INTO trust_event for %v statement event: %v", event, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.TrustAt)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else if event.EventType == Inactive {
		//update
		stmt, err := DB.Prepare("UPDATE trust_event SET un_trust_at=$1 WHERE id=$2 AND un_trust_at IS NULL")
		if err != nil {
			return fmt.Errorf("prepare UPDATE trust_event for %v statement failed: %v", id, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.UnTrustAt, id)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func FindLastEventTrustedRepo(rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var trustAt *time.Time
	var unTrustAt *time.Time
	var id *uuid.UUID
	err := DB.
		QueryRow(`SELECT id, trust_at, un_trust_at
			      		FROM trust_event 
						WHERE repo_id=$1 
						ORDER by trust_at DESC LIMIT 1`,
			rid).Scan(&id, &trustAt, &unTrustAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return id, trustAt, unTrustAt, nil
	default:
		return nil, nil, nil, err
	}
}

func FindTrustedRepos() ([]Repo, error) {
	//we want to send back an empty array, don't change
	t := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
            FROM trust_event t
            INNER JOIN repo r ON t.repo_id=r.id
			WHERE t.un_trust_at IS NULL`
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
//func FindTrustedReposBetween(start time.Time, stop time.Time) ([]TrustResult, error) {
//	rows, err := DB.Query(`
//			SELECT user_id, repo_id
//			FROM trust_event
//			WHERE trust_at < $1 AND (un_trust_at IS NULL OR un_trust_at >= $2)
//			GROUP BY user_id, repo_id
//			ORDER BY user_id`, start, stop)
//	if err != nil {
//		return nil, err
//	}
//	defer CloseAndLog(rows)
//
//	trustResults := []TrustResult{}
//	var userIdOld uuid.UUID
//	var userId uuid.UUID
//	trustResult := TrustResult{}
//	for rows.Next() {
//		var repoId uuid.UUID
//		err = rows.Scan(&userId, &repoId)
//
//		if userId != userIdOld && !util.IsUUIDZero(userIdOld) {
//			trustResult.UserId = userIdOld
//			trustResults = append(trustResults, trustResult)
//
//			trustResult = TrustResult{}
//			userIdOld = userId
//		}
//
//		if err != nil {
//			return nil, err
//		}
//		trustResult.RepoIds = append(trustResult.RepoIds, repoId)
//		if util.IsUUIDZero(userIdOld) {
//			userIdOld = userId
//		}
//	}
//
//	trustResult.UserId = userId
//	trustResults = append(trustResults, trustResult)
//
//	return trustResults, nil
//}
