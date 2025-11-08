package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SponsorEvent struct {
	Id          uuid.UUID  `json:"id"`
	Uid         uuid.UUID  `json:"uid"`
	RepoId      uuid.UUID  `json:"repo_id"`
	EventType   uint8      `json:"event_type"`
	SponsorAt   *time.Time `json:"sponsor_at"`
	UnSponsorAt *time.Time `json:"un_sponsor_at"`
}

func (db *DB) InsertOrUpdateSponsor(event *SponsorEvent) error {
	// First get last sponsored event to check if we need to insert or update
	id, sponsorAt, unSponsorAt, err := db.FindLastEventSponsoredRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	// No event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("cannot unsponsor: not currently sponsoring this repo")
	}

	// Event found - validate
	if id != nil {
		if event.EventType == Inactive {
			if event.SponsorAt != nil || event.UnSponsorAt == nil {
				return fmt.Errorf("to unsponsor, must set un_sponsor_at but not sponsor_at: "+
					"event.SponsorAt: %v, event.UnSponsorAt: %v", event.SponsorAt, event.UnSponsorAt)
			}
			if unSponsorAt != nil {
				return fmt.Errorf("already unsponsored: unSponsorAt: %v", unSponsorAt)
			}
			if unSponsorAt == nil && event.UnSponsorAt.Before(*sponsorAt) {
				return fmt.Errorf("un_sponsor_at cannot be before sponsor_at: sponsorAt: %v, event.UnSponsorAt: %v",
					sponsorAt, event.UnSponsorAt)
			}
		} else if event.EventType == Active {
			if event.SponsorAt == nil || event.UnSponsorAt != nil {
				return fmt.Errorf("to sponsor, must set sponsor_at but not un_sponsor_at: "+
					"event.SponsorAt: %v, event.UnSponsorAt: %v", event.SponsorAt, event.UnSponsorAt)
			}
			if unSponsorAt == nil {
				return fmt.Errorf("already sponsoring this repo: sponsor_at: %v, un_sponsor_at: %v",
					event.SponsorAt, unSponsorAt)
			} else {
				if event.SponsorAt.Before(*sponsorAt) {
					return fmt.Errorf("cannot sponsor in the past: event.SponsorAt: %v, sponsorAt: %v",
						event.SponsorAt, sponsorAt)
				}
				if event.SponsorAt.Before(*unSponsorAt) {
					return fmt.Errorf("cannot sponsor before un_sponsor_at: event.SponsorAt: %v, unSponsorAt: %v",
						event.SponsorAt, unSponsorAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}
	}

	if event.EventType == Active {
		_, err := db.Exec(
			`INSERT INTO sponsor_event (id, user_id, repo_id, sponsor_at) VALUES ($1, $2, $3, $4)`,
			event.Id, event.Uid, event.RepoId, event.SponsorAt)
		return err
	} else if event.EventType == Inactive {
		_, err := db.Exec(
			`UPDATE sponsor_event SET un_sponsor_at=$1 WHERE id=$2 AND un_sponsor_at IS NULL`,
			event.UnSponsorAt, id)
		return err
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func (db *DB) FindLastEventSponsoredRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var sponsorAt *time.Time
	var unSponsorAt *time.Time
	var id uuid.UUID
	
	err := db.QueryRow(`
		SELECT id, sponsor_at, un_sponsor_at
		FROM sponsor_event 
		WHERE user_id=$1 AND repo_id=$2 
		ORDER BY sponsor_at DESC 
		LIMIT 1`, uid, rid).Scan(&id, &sponsorAt, &unSponsorAt)
	
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return &id, sponsorAt, unSponsorAt, nil
	default:
		return nil, nil, nil, err
	}
}

func (db *DB) FindSponsoredReposByUserId(userId uuid.UUID) ([]Repo, error) {
	rows, err := db.Query(`
		SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
		FROM sponsor_event s
		INNER JOIN repo r ON s.repo_id=r.id
		WHERE s.user_id=$1 AND s.un_sponsor_at IS NULL`, userId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepos(rows)
}

type SponsorResult struct {
	UserId  uuid.UUID
	RepoIds []uuid.UUID
}

func (db *DB) FindSponsorsBetween(start time.Time, stop time.Time) ([]SponsorResult, error) {
	rows, err := db.Query(`		
		SELECT user_id, repo_id
		FROM sponsor_event
		WHERE sponsor_at < $1 AND (un_sponsor_at IS NULL OR un_sponsor_at >= $2)
		GROUP BY user_id, repo_id
		ORDER BY user_id`, start, stop)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var sponsorResults []SponsorResult
	var userIdOld uuid.UUID
	var userId uuid.UUID
	sponsorResult := SponsorResult{}
	
	for rows.Next() {
		var repoId uuid.UUID
		err = rows.Scan(&userId, &repoId)
		if err != nil {
			return nil, err
		}

		if userId != userIdOld && !IsUUIDZero(userIdOld) {
			sponsorResult.UserId = userIdOld
			sponsorResults = append(sponsorResults, sponsorResult)
			sponsorResult = SponsorResult{}
			userIdOld = userId
		}

		sponsorResult.RepoIds = append(sponsorResult.RepoIds, repoId)
		if IsUUIDZero(userIdOld) {
			userIdOld = userId
		}
	}

	if !IsUUIDZero(userId) {
		sponsorResult.UserId = userId
		sponsorResults = append(sponsorResults, sponsorResult)
	}

	return sponsorResults, nil
}

func IsUUIDZero(id uuid.UUID) bool {
	for x := 0; x < 16; x++ {
		if id[x] != 0 {
			return false
		}
	}
	return true
}