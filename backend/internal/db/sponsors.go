package db

import (
	"backend/pkg/util"
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

func InsertOrUpdateSponsor(event *SponsorEvent) error {
	//first get last sponsored event to check if we need to insertOrUpdateSponsor or unsponsor
	//TODO: use mutex
	id, sponsorAt, unSponsorAt, err := FindLastEventSponsoredRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	//no event found
	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to unsponsor, but we are currently not sponsoring this repo")
	}

	//event found
	if id != nil {
		if event.EventType == Inactive {
			if event.SponsorAt != nil || event.UnSponsorAt == nil {
				return fmt.Errorf("to unsponser, we must set the unsponser_at, but not the sponser_at: "+
					"event.SponsorAt: %v, event.UnSponsorAt: %v", event.SponsorAt, event.UnSponsorAt)
			}
			if unSponsorAt != nil {
				return fmt.Errorf("we want to unsponsor, but we already unsponsored it: unSponsorAt: %v", unSponsorAt)
			}
			if unSponsorAt == nil && event.UnSponsorAt.Before(*sponsorAt) {
				return fmt.Errorf("we want to unsponsor, but the unsponsor date is before the sponsor date: sponsorAt: "+
					"%v, event.UnSponsorAt: %v", sponsorAt, event.UnSponsorAt)
			}
		} else if event.EventType == Active {
			if event.SponsorAt == nil || event.UnSponsorAt != nil {
				return fmt.Errorf("to sponser, we must set the sponsor_at, but not the unsponser_at: "+
					"event.SponsorAt: %v, event.UnSponsorAt: %v", event.SponsorAt, event.UnSponsorAt)
			}
			if unSponsorAt == nil {
				return fmt.Errorf("we want to sponsor, but we are already sponsoring this repo: "+
					"sponsor_at: %v, un_sponsor_at: %v", event.SponsorAt, unSponsorAt)
			} else {
				if event.SponsorAt.Before(*sponsorAt) {
					return fmt.Errorf("we want to sponsor, but we want want to sponser this repo in the past: "+
						"event.SponsorAt: %v, sponsorAt: %v", event.SponsorAt, sponsorAt)
				}
				if event.SponsorAt.Before(*unSponsorAt) {
					return fmt.Errorf("we want to sponsor, but we want want to sponser this repo in the past: "+
						"event.SponsorAt: %v, unSponsorAt: %v", event.SponsorAt, unSponsorAt)
				}
			}
		} else {
			return fmt.Errorf("unknown event type %v", event.EventType)
		}

	}

	if event.EventType == Active {
		//insert
		stmt, err := DB.Prepare("INSERT INTO sponsor_event (id, user_id, repo_id, sponsor_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return fmt.Errorf("prepare INSERT INTO sponsor_event for %v statement event: %v", event, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.SponsorAt)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else if event.EventType == Inactive {
		//update
		stmt, err := DB.Prepare("UPDATE sponsor_event SET un_sponsor_at=$1 WHERE id=$2 AND un_sponsor_at IS NULL")
		if err != nil {
			return fmt.Errorf("prepare UPDATE sponsor_event for %v statement failed: %v", id, err)
		}
		defer CloseAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.UnSponsorAt, id)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else {
		return fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func FindLastEventSponsoredRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var sponsorAt *time.Time
	var unSponsorAt *time.Time
	var id *uuid.UUID
	err := DB.
		QueryRow(`SELECT id, sponsor_at, un_sponsor_at
			      		FROM sponsor_event 
						WHERE user_id=$1 AND repo_id=$2 
						ORDER by sponsor_at DESC LIMIT 1`,
			uid, rid).Scan(&id, &sponsorAt, &unSponsorAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return id, sponsorAt, unSponsorAt, nil
	default:
		return nil, nil, nil, err
	}
}

func FindSponsoredReposByUserId(userId uuid.UUID) ([]Repo, error) {
	//we want to send back an empty array, don't change
	s := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
            FROM sponsor_event s
            INNER JOIN repo r ON s.repo_id=r.id
			WHERE s.user_id=$1 AND s.un_sponsor_at IS NULL`
	rows, err := DB.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepo(rows)
}

type SponsorResult struct {
	UserId  uuid.UUID
	RepoIds []uuid.UUID
}

func FindSponsorsBetween(start time.Time, stop time.Time) ([]SponsorResult, error) {
	rows, err := DB.Query(`		
			SELECT user_id, repo_id
			FROM sponsor_event
			WHERE sponsor_at < $1 AND (un_sponsor_at IS NULL OR un_sponsor_at >= $2)
			GROUP BY user_id, repo_id
			ORDER BY user_id`, start, stop)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	sponsorResults := []SponsorResult{}
	var userIdOld uuid.UUID
	var userId uuid.UUID
	sponsorResult := SponsorResult{}
	for rows.Next() {
		var repoId uuid.UUID
		err = rows.Scan(&userId, &repoId)

		if userId != userIdOld && !util.IsUUIDZero(userIdOld) {
			sponsorResult.UserId = userIdOld
			sponsorResults = append(sponsorResults, sponsorResult)

			sponsorResult = SponsorResult{}
			userIdOld = userId
		}

		if err != nil {
			return nil, err
		}
		sponsorResult.RepoIds = append(sponsorResult.RepoIds, repoId)
		if util.IsUUIDZero(userIdOld) {
			userIdOld = userId
		}
	}

	sponsorResult.UserId = userId
	sponsorResults = append(sponsorResults, sponsorResult)

	return sponsorResults, nil
}

func GetStarCount(repoId uuid.UUID, activeSponsors []uuid.UUID, isPostgres bool) (int, error) {
	if len(activeSponsors) == 0 {
		return 0, nil
	}

	var query string
	if isPostgres {
		query = `
			SELECT COUNT(DISTINCT user_id)
			FROM sponsor_event
			WHERE repo_id = $1 AND user_id = ANY($2) AND un_sponsor_at IS NULL`
	} else {
		query = `
			SELECT COUNT(DISTINCT user_id)
			FROM sponsor_event
			WHERE repo_id = ? AND user_id IN (` + GeneratePlaceholders(len(activeSponsors)) + `) AND un_sponsor_at IS NULL`
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
