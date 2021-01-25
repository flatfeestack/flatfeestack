package main

import (
	"github.com/snabb/isoweek"
	"log"
	"strconv"
	"time"
)

var (
	mUSDPerHour = 27500 //2.75 cents - x 10'000
	sUSDPerHour = strconv.Itoa(mUSDPerHour)
)

type UserBalance struct {
	PayoutEth string `json:"payout_eth"`
	Balance   int64  `json:"balance"`
	Email     string `json:"email"`
}

func time2Day(year int, month time.Month, day int) *time.Time {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}

func dailyRunner(now time.Time) error {
	yesterdayStop := time2Day(now.Year(), now.Month(), now.Day()) //$2
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)             //$1

	//https://stackoverflow.com/questions/17833176/postgresql-days-months-years-between-two-dates
	stmt, err := db.Prepare(`INSERT INTO daily_repo_hours (user_id, repo_hours, day)
              SELECT s.user_id, 
                     SUM((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::int) as repo_hours, 
                     $1 as day
                FROM sponsor_event s 
                    JOIN users u ON u.id = s.user_id
                WHERE u.subscription_state='ACTIVE' 
                    AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
                GROUP by s.user_id`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}

	//TODO: limit user to 10000 repos
	//we can support up to 1000 (1h) - 27500 (24h) repos until the precision makes the distribution of 0
	stmt, err = db.Prepare(`INSERT INTO daily_repo_balance (repo_id, balance, day)
		   SELECT repo_id, 
				  SUM(((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::bigint * ` + sUSDPerHour + `) / d.repo_hours), 
			      $1 as day
			 FROM sponsor_event s
			     JOIN users u ON u.id = s.user_id 
			     JOIN daily_repo_hours d ON u.id = d.user_id
			 WHERE u.subscription_state='ACTIVE' AND d.day=$1 
			     AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
			 GROUP by s.repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}

	return nil
}

func weeklyRunner(now time.Time) error {
	dayStop := time2Day(now.Year(), now.Month(), now.Day()) //$2
	year, week := dayStop.ISOWeek()
	weekStop := isoweek.StartTime(year, week, time.UTC)
	weekStart := weekStop.AddDate(0, 0, -7) //$1

	stmt, err := db.Prepare(`INSERT INTO weekly_repo_balance (repo_id, balance, day)
		SELECT repo_id, SUM(balance) as balance, $1 as day
        FROM daily_repo_balance
        WHERE day >= $1 AND day < $2
        GROUP BY repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(weekStart, weekStop)
	if err != nil {
		return err
	}

	//to stay on the safe side and don't pay more than we have, round down
	stmt, err = db.Prepare(`INSERT INTO weekly_email_payout (email, balance, day)
		SELECT c.git_email as email, floor(SUM(c.weight * w.balance)) as balance, $1 as day
        FROM analysis_response c
            JOIN analysis_request a ON c.analysis_request_id = a.id
            JOIN weekly_repo_balance w ON w.repo_id = a.repo_id
        WHERE a.date_from = $1 AND a.date_to = $2 AND w.day = $1
		GROUP BY c.git_email`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(weekStart, weekStop)
	if err != nil {
		return err
	}

	var userBalances []UserBalance
	sql := "SELECT email, balance FROM weekly_email_payout WHERE day=$1"
	rows, err := db.Query(sql, weekStart)
	defer rows.Close()

	if err != nil {
		return err
	}

	for rows.Next() {
		var userBalance UserBalance
		err = rows.Scan(&userBalance.Email, &userBalance.Balance)
		if err != nil {
			return err
		}
		//TODO: get more infos to which repo the user contributed
		userBalances = append(userBalances, userBalance)
	}

	return campaignRequest(userBalances)
}

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
func monthlyRunner(now time.Time) error {
	monthStop := time2Day(now.Year(), now.Month(), 1) //$2
	monthStart := monthStop.AddDate(0, -1, 0)         //$1

	stmt, err := db.Prepare(`INSERT INTO monthly_repo_weight (repo_id, weight, day)
		SELECT req.repo_id, SUM(res.weight) as weight, $1 as day
        FROM analysis_response res
            JOIN analysis_request req ON res.analysis_request_id = req.id
			JOIN git_email g ON g.email = res.git_email
        WHERE req.date_from = $1 AND req.date_to = $2
		GROUP BY req.repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(monthStart, monthStop)
	if err != nil {
		return err
	}
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}

	stmt, err = db.Prepare(`INSERT INTO monthly_repo_balance (repo_id, balance, day)
		SELECT d.repo_id, 
		       SUM(balance) + COALESCE((
		           SELECT m.balance 
		           FROM monthly_future_leftover m 
		           WHERE m.repo_id = d.repo_id AND m.day = $1), 0) as balance, 
		       $1 as day
        FROM daily_repo_balance d
        WHERE d.day >= $1 AND d.day < $2
        GROUP BY d.repo_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err = stmt.Exec(monthStart, monthStop)
	if err != nil {
		return err
	}
	nr, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("affected: %v", nr)

	//this is the real payout
	stmt, err = db.Prepare(`INSERT INTO monthly_user_payout (user_id, balance, day)
		SELECT g.user_id, floor(SUM(mb.balance * res.weight / mw.weight)) as balance, $1 as day
        FROM analysis_response res
            JOIN analysis_request req ON res.analysis_request_id = req.id
            JOIN git_email g ON g.email = res.git_email
            JOIN monthly_repo_weight mw ON mw.repo_id = req.repo_id
            JOIN monthly_repo_balance mb ON mb.repo_id = req.repo_id
        WHERE req.date_from = $1 AND req.date_to = $2 AND mw.day = $1 AND mb.day = $1
		GROUP BY g.user_id`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err = stmt.Exec(monthStart, monthStop)
	if err != nil {
		return err
	}
	nr, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("affected: %v", nr)

	//shift leftover to separate table. Those repos which did not claim money go into separate
	stmt, err = db.Prepare(`INSERT INTO monthly_future_leftover (repo_id, balance, day)
		SELECT mb.repo_id, mb.balance, $2 as day
        FROM monthly_repo_balance mb
            LEFT JOIN monthly_repo_weight mw ON mb.repo_id = mw.repo_id
        WHERE mw.repo_id IS NULL AND mb.day = $1`)

	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err = stmt.Exec(monthStart, monthStop)
	if err != nil {
		return err
	}
	nr, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Printf("affected: %v", nr)
	//TODO: write email to admins if we have payouts going on
	return nil
}

func campaignRequest(balances []UserBalance) error {
	return nil
}
