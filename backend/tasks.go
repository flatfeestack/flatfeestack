package main

import (
	"github.com/google/uuid"
	"log"
	"strconv"
	"time"
)

var (
	mUSDPerHour = 27500 //2.75 cents - x 10'000
	sUSDPerHour = strconv.Itoa(mUSDPerHour)
)

func time2Day(now time.Time) *time.Time {
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return &t
}

func time2Month(now time.Time) *time.Time {
	t := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return &t
}

func dailyRunner(now time.Time) error {
	yesterdayStop := time2Day(now)                    //$2
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1) //$1

	//https://stackoverflow.com/questions/17833176/postgresql-days-months-years-between-two-dates
	stmt, err := db.Prepare(`INSERT INTO daily_repo_hours (user_id, repo_hours, day, created_at)
              SELECT s.user_id, 
                     SUM((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::int) as repo_hours, 
                     $1 as day, $3 as created_at
                FROM sponsor_event s 
                    JOIN users u ON u.id = s.user_id
                WHERE u.subscription_state='ACTIVE' 
                    AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
                GROUP by s.user_id`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}

	//TODO: limit user to 10000 repos
	//we can support up to 1000 (1h) - 27500 (24h) repos until the precision makes the distribution of 0

	stmt, err = db.Prepare(`INSERT INTO daily_repo_balance (repo_id, balance, day, created_at)
		   SELECT repo_id, 
				  SUM(((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::bigint * ` + sUSDPerHour + `) / drh.repo_hours) + COALESCE((
		             SELECT dfl.balance 
		             FROM daily_future_leftover dfl 
		             WHERE dfl.repo_id = s.repo_id AND dfl.day = $1), 0), 
			      $1 as day, 
                  $3 as created_at
			 FROM sponsor_event s
			     JOIN users u ON u.id = s.user_id 
			     JOIN daily_repo_hours drh ON u.id = drh.user_id
			 WHERE u.subscription_state='ACTIVE' AND drh.day=$1 
			     AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
			 GROUP by s.repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}

	//daily runner for knowing who could potentially get something
	stmt, err = db.Prepare(`INSERT INTO daily_email_payout (email, balance, day, created_at)
		SELECT res.git_email as email, 
		       FLOOR(SUM(res.weight * drb.balance)) as balance, 
		       $1 as day, 
		       $2 as created_at
        FROM analysis_response res
            JOIN (SELECT id, MAX(date_to) as date_to, repo_id FROM analysis_request GROUP BY id, date_to, repo_id) 
                as req ON res.analysis_request_id = req.id
            JOIN daily_repo_balance drb ON drb.repo_id = req.repo_id
        WHERE drb.day = $1
		GROUP BY res.git_email`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(yesterdayStart, now)
	if err != nil {
		return err
	}

	stmt, err = db.Prepare(`INSERT INTO daily_repo_weight (repo_id, weight, day, created_at)
		SELECT req.repo_id as repo_id, 
		       SUM(res.weight) as weight,
		       $1 as day, 
		       $2 as created_at
        FROM analysis_response res
            JOIN (SELECT id, MAX(date_to) as date_to, repo_id FROM analysis_request GROUP BY id, date_to, repo_id)
                as req ON res.analysis_request_id = req.id
			JOIN git_email g ON g.email = res.git_email
		GROUP BY req.repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(yesterdayStart, now)
	if err != nil {
		return err
	}
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}

	stmt, err = db.Prepare(`INSERT INTO daily_user_payout (user_id, balance, day, created_at)
		SELECT g.user_id as user_id, 
		       FLOOR(SUM(drb.balance * res.weight / drw.weight)) as balance, 
		       $1 as day, 
		       $2 as created_at
        FROM analysis_response res
            JOIN (SELECT id, MAX(date_to) as date_to, repo_id FROM analysis_request GROUP BY id, date_to, repo_id)
                as req ON res.analysis_request_id = req.id
            JOIN git_email g ON g.email = res.git_email
            JOIN daily_repo_weight drw ON drw.repo_id = req.repo_id
            JOIN daily_repo_balance drb ON drb.repo_id = req.repo_id
        WHERE drw.day = $1 AND drb.day = $1
		GROUP BY g.user_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err = stmt.Exec(yesterdayStart, now)
	if err != nil {
		return err
	}
	nr, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("affected: %v", nr)

	//shift leftover to separate table. Those repos which did not claim money go into separate
	stmt, err = db.Prepare(`INSERT INTO daily_future_leftover (repo_id, balance, day, created_at)
		SELECT drb.repo_id, drb.balance, $2 as day, $3 as created_at
        FROM daily_repo_balance drb
            LEFT JOIN daily_repo_weight drw ON drb.repo_id = drw.repo_id
        WHERE drw.repo_id IS NULL AND drb.day = $1`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return err
	}
	nr, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("affected: %v", nr)

	return nil
}

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
func weeklyRunner(now time.Time) error {
	sql := `SELECT drw.repo_id, r.url
            FROM analysis_response res
                JOIN (SELECT id, MAX(date_to) as date_to, repo_id FROM analysis_request GROUP BY id, date_to, repo_id)
                    as req ON res.analysis_request_id = req.id
		        JOIN daily_repo_weight drw ON drw.repo_id=req.repo_id
			    JOIN repo r ON drw.repo_id = r.id 
			WHERE DATE_PART('day', AGE(req.date_to, $1)) > 5`
	rows, err := db.Query(sql, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var repoId uuid.UUID
		var url string
		err = rows.Scan(&repoId, &url)
		if err != nil {
			return err
		}
		err = analysisRequest(repoId, url)
		if err != nil {
			return err
		}
	}
	return nil
}
