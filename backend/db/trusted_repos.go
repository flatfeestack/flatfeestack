package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TrustEvent struct {
	Id        uuid.UUID  `json:"id"`
	Uid       uuid.UUID  `json:"uid"`
	RepoId    uuid.UUID  `json:"repo_id"`
	EventType uint8      `json:"event_type"`
	TrustAt   *time.Time `json:"trust_at"`
	UnTrustAt *time.Time `json:"un_trust_at"`
}

func (db *DB) InsertOrUpdateTrust(event *TrustEvent) error {
	// First get last trust event to check if we need to insert or update
	id, trustAt, unTrustAt, err := db.FindLastEventTrustedRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	// No event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("cannot untrust: not currently trusting this repo")
	}

	// Event found - validate
	if id != nil {
		if event.EventType == Inactive {
			if event.TrustAt != nil || event.UnTrustAt == nil {
				return fmt.Errorf("to untrust, must set un_trust_at but not trust_at: "+
					"event.TrustAt: %v, event.UnTrustAt: %v", event.TrustAt, event.UnTrustAt)
			}
			if unTrustAt != nil {
				return fmt.Errorf("already untrusted: unTrustAt: %v", unTrustAt)
			}
			if unTrustAt == nil && event.UnTrustAt.Before(*trustAt) {
				return fmt.Errorf("un_trust_at cannot be before trust_at: trustAt: %v, event.UnTrustAt: %v",
					trustAt, event.UnTrustAt)
			}
		} else if event.EventType == Active {
			if event.TrustAt == nil || event.UnTrustAt != nil {
				return fmt.Errorf("to trust, must set trust_at but not un_trust_at: "+
					"event.TrustAt: %v, event.UnTrustAt: %v", event.TrustAt, event.UnTrustAt)
			}
			if unTrustAt == nil {
				return fmt.Errorf("already trusting this repo: trust_at: %v, un_trust_at: %v",
					event.TrustAt, unTrustAt)
			} else {
				if event.TrustAt.Before(*trustAt) {
					return fmt.Errorf("cannot trust in the past: event.TrustAt: %v, trustAt: %v",
						event.TrustAt, trustAt)
				}
				if event.TrustAt.Before(*unTrustAt) {
					return fmt.Errorf("cannot trust before un_trust_at: event.TrustAt: %v, unTrustAt: %v",
						event.TrustAt, unTrustAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}
	}

	if event.EventType == Active {
		_, err := db.Exec(
			`INSERT INTO trust_event (id, user_id, repo_id, trust_at) VALUES ($1, $2, $3, $4)`,
			event.Id, event.Uid, event.RepoId, event.TrustAt)
		return err
	} else if event.EventType == Inactive {
		_, err := db.Exec(
			`UPDATE trust_event SET un_trust_at=$1 WHERE id=$2 AND un_trust_at IS NULL`,
			event.UnTrustAt, id)
		return err
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func (db *DB) FindLastEventTrustedRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var trustAt *time.Time
	var unTrustAt *time.Time
	var id uuid.UUID
	
	err := db.QueryRow(`
		SELECT id, trust_at, un_trust_at
		FROM trust_event 
		WHERE user_id=$1 AND repo_id=$2 
		ORDER BY trust_at DESC 
		LIMIT 1`, uid, rid).Scan(&id, &trustAt, &unTrustAt)
	
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return &id, trustAt, unTrustAt, nil
	default:
		return nil, nil, nil, err
	}
}

func (db *DB) FindTrustedReposByUserId(userId uuid.UUID) ([]Repo, error) {
	rows, err := db.Query(`
		SELECT r.id, r.url, r.git_url, r.name, r.description, r.source, r.created_at
		FROM trust_event t
		INNER JOIN repo r ON t.repo_id=r.id
		WHERE t.user_id=$1 AND t.un_trust_at IS NULL`, userId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepos(rows)
}

func (db *DB) GetTrustedReposFromList(rids []uuid.UUID) ([]uuid.UUID, error) {
	if len(rids) == 0 {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT repo_id
		FROM trust_event
		WHERE repo_id = ANY($1) AND un_trust_at IS NULL`,
		pq.Array(rids))
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var trustedRepos []uuid.UUID
	for rows.Next() {
		var repoID uuid.UUID
		if err := rows.Scan(&repoID); err != nil {
			return nil, err
		}
		trustedRepos = append(trustedRepos, repoID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return trustedRepos, nil
}