package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func InsertOrUpdateSponsor(event *SponsorEvent) error {
	//first get last sponsored event to check if we need to insertOrUpdateSponsor or unsponsor
	//TODO: use mutex
	id, sponsorAt, unSponsorAt, err := FindLastEventSponsoredRepo(event.Uid, event.RepoId)
	if err != nil {
		return err
	}

	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to unsponsor, but we are currently not sponsoring this repo")
	}

	if id != nil && event.EventType == Inactive && unSponsorAt != nil {
		return fmt.Errorf("we want to unsponsor, but we already unsponsored it")
	}

	if id != nil && event.EventType == Active && (unSponsorAt == nil || event.SponsorAt.Before(*unSponsorAt)) {
		return fmt.Errorf("we want to insertOrUpdateSponsor, but we are already sponsoring this repo: "+
			"sponsor_at: %v, un_sponsor_at: %v", event.SponsorAt, unSponsorAt)
	}

	if id != nil && event.EventType == Active && !sponsorAt.Before(event.SponsorAt) {
		return fmt.Errorf("we want to insertOrUpdateSponsor, but we want to sponsor at an earlier time: "+
			"sponsor_at: %v, sponsor_at(db): %v, un_sponsor_at: %v, %v", event.SponsorAt, sponsorAt, unSponsorAt, event.SponsorAt.Before(*unSponsorAt))
	}

	if event.EventType == Active {
		//insert
		stmt, err := db.Prepare("INSERT INTO sponsor_event (id, user_id, repo_id, sponsor_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return fmt.Errorf("prepare INSERT INTO sponsor_event for %v statement event: %v", event, err)
		}
		defer closeAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.SponsorAt)
		if err != nil {
			return err
		}
		return handleErrMustInsertOne(res)
	} else if event.EventType == Inactive {
		//update
		stmt, err := db.Prepare("UPDATE sponsor_event SET un_sponsor_at=$1 WHERE id=$2 AND un_sponsor_at IS NULL")
		if err != nil {
			return fmt.Errorf("prepare UPDATE sponsor_event for %v statement failed: %v", id, err)
		}
		defer closeAndLog(stmt)

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
	err := db.
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

func FindSponsoredReposById(userId uuid.UUID) ([]Repo, error) {
	//we want to send back an empty array, don't change
	s := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source
            FROM sponsor_event s
            INNER JOIN repo r ON s.repo_id=r.id
			WHERE s.user_id=$1 AND s.un_sponsor_at IS NULL`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)
	return ScanRepo(rows)
}

type SponsorResult struct {
	UserId  uuid.UUID
	RepoIds []uuid.UUID
}

func FindSponsorsBetween(start time.Time, stop time.Time) ([]SponsorResult, error) {
	rows, err := db.Query(`		
			SELECT user_id, `+agg+`(repo_id)
			FROM sponsor_event
			WHERE sponsor_at < $1 AND (un_sponsor_at IS NULL OR un_sponsor_at >= $2)
			GROUP BY user_id`, start.Format("2006-02-01"), stop.Format("2006-02-01"))
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	sponsorResults := []SponsorResult{}
	for rows.Next() {
		var sponsorResult SponsorResult
		var repoIdsJSON string
		err = rows.Scan(&sponsorResult.UserId, &repoIdsJSON)
		if err != nil {
			return nil, err
		}
		var repoIds []uuid.UUID
		err = json.Unmarshal([]byte(repoIdsJSON), &repoIds)
		if err != nil {
			return nil, err
		}
		sponsorResult.RepoIds = repoIds

		sponsorResults = append(sponsorResults, sponsorResult)
	}
	return sponsorResults, nil
}
