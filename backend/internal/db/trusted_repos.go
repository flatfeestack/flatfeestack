package db

import (
	"database/sql"
	"fmt"
	"strings"
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

	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to untrust, but we are currently not trusting this repo")
	}

	if id != nil {
		if event.EventType == Inactive {
			if event.TrustAt != nil || event.UnTrustAt == nil { // Not tested
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
			if unTrustAt == nil { // not tested
				return fmt.Errorf("we want to trust, but we are already trusting this repo: "+
					"trust_at: %v, un_trust_at: %v", event.TrustAt, unTrustAt)
			} else {
				if event.TrustAt.Before(*trustAt) { // both not tested
					return fmt.Errorf("we want to trust, but we want to trust this repo in the past: "+
						"event.TrustAt: %v, trustAt: %v", event.TrustAt, trustAt)
				}
				if event.TrustAt.Before(*unTrustAt) {
					return fmt.Errorf("we want to trust, but we want to trust this repo in the past: "+
						"event.TrustAt: %v, unTrustAt: %v", event.TrustAt, unTrustAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}

	}

	// I don't see this tested anywhere?
	if event.EventType == Active {
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
	t := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source, t.trust_at
            FROM trust_event t
            INNER JOIN repo r ON t.repo_id=r.id
			WHERE t.un_trust_at IS NULL`
	rows, err := DB.Query(t)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepoWithTrustEvent(rows)
}

func GetTrustedReposFromList(rids []uuid.UUID) ([]uuid.UUID, error) {
	if len(rids) == 0 {
		return nil, nil
	}

	query := `
		SELECT repo_id
		FROM trust_event
		WHERE repo_id IN (` + GeneratePlaceholders(len(rids)) + `)`

	rows, err := DB.Query(query, ConvertToInterfaceSlice(rids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func GeneratePlaceholders(n int) string {
	var placeholders []string
	for i := 1; i <= n; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
	}
	return strings.Join(placeholders, ", ")
}

func ConvertToInterfaceSlice[T any](values []T) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

func CountReposForUsers(userIds []uuid.UUID, months int, isPostgres bool) (int, error) {
	var query string

	if len(userIds) == 0 || months < 0 {
		return 0, nil
	}

	if isPostgres {
		query = fmt.Sprintf(`
		SELECT COUNT(repo_id)
		FROM daily_contribution
		WHERE
			user_sponsor_id IN (`+GeneratePlaceholders(len(userIds))+`)
			AND created_at >= CURRENT_DATE - INTERVAL '%d month'`, months)
	} else {
		query = fmt.Sprintf(`
		SELECT COUNT(repo_id)
		FROM daily_contribution
		WHERE
			user_sponsor_id IN (`+GeneratePlaceholders(len(userIds))+`)
			AND created_at >= date('now', '-%d month')`, months)
	}

	var count int
	err := DB.QueryRow(query, ConvertToInterfaceSlice(userIds)...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetRepoWeight(repoId uuid.UUID, activeUserMinMonths int, latestRepoSponsoringMonths int, isPostgres bool) (float64, error) {
	emails, err := GetRepoEmails(repoId)
	if err != nil {
		return 0.0, err
	}

	userIds, err := FindUsersByGitEmails(emails)
	if err != nil {
		return 0.0, err
	}

	activeUsers, err := FilterActiveUsers(userIds, activeUserMinMonths, isPostgres)
	if err != nil {
		return 0.0, err
	}

	if len(activeUsers) == 0 {
		return 0.0, nil
	}

	trustedRepoCount, err := CountReposForUsers(activeUsers, latestRepoSponsoringMonths, isPostgres)
	if err != nil {
		return 0.0, err
	}

	return float64(trustedRepoCount), nil
}
