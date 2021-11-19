package main

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

//*********************************************************************************
//******************************* Daily calculations ******************************
//*********************************************************************************

//Here we calculate the total time (repo hours) a user has supported per day. If the user supported
//2 repositories for 24h, then the repo hour is 48h. If the user supported 3 repos for 2h each, then
//the repo hour for the user at this day is 6h. The result is stored in daily_repo_hours.
//
//Only users with the role "USR" and who have balance left are considered. If a user supports at least
//1h then the full day (mUSDPerDay) should be deducted.
//
//Running this twice won't work as we have a unique index on: user_id, day

//This inserts a balance deduction for the user, that has the role "USR", has funds, and at least 24h of supported
//repo. The balance is negative, thus deducted.
//
//Running this twice won't work as we have a unique index on: payment_cycle_id, user_id, balance_type, day
func runDailyUserBalance(yesterdayStart time.Time, yesterdayEnd time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO user_balances (payment_cycle_id, user_id, balance, balance_type, currency, day, created_at)
		   select * from updateDailyUserBalance($1, $2, $3) 
		       f(payment_cycle_id uuid, user_id uuid, balance bigint, balance_type text, currency VARCHAR(16), day date, created_at TIMESTAMP with time zone)
`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayEnd, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

// Here we update the days left for each currency of the user. This is done by balance_sum_of_currency / daily_payment_of_currency
//Running this twice is ok, as it will give a more accurate state
func runDailyDaysLeftDailyPayment() (int64, error) {
	stmt, err := db.Prepare(`
           	UPDATE daily_payment SET days_left = q.sum / q.amount
			FROM (
				SELECT ub.payment_cycle_id, ub.currency, SUM(ub.balance) AS sum, MIN(dp.amount) as amount
				FROM user_balances ub 
				JOIN daily_payment dp ON dp.payment_cycle_id = ub.payment_cycle_id AND dp.currency = ub.currency
				GROUP BY ub.payment_cycle_id, ub.currency 
			) AS q
			WHERE daily_payment.payment_cycle_id = q.payment_cycle_id AND daily_payment.currency = q.currency`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec()
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

// Because we have multiple currencies we need to update the sum of the days_left by currency and ad it to the payment_cycle
//Running this twice is ok, as it will give a more accurate state
func runDailyDaysLeftPaymentCycle() (int64, error) {
	stmt, err := db.Prepare(`
        	UPDATE payment_cycle SET days_left = q.days_left
			FROM (
					SELECT dp.payment_cycle_id, SUM(dp.days_left) AS days_left FROM daily_payment dp GROUP BY dp.payment_cycle_id  
			) AS q
			WHERE payment_cycle.id = q.payment_cycle_id`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec()
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

//TODO: limit user to 10000 repos
//we can support up to 1000 (1h) - 27500 (24h) repos until the precision makes the distribution of 0
//
//Here we calculate how much balance a repository gets. Each repo which is longer supported than 24h will get the same portion.
//So if the user has supported 3 repos, each repo gets 1/3.
//
//Running this twice does not work as we have a unique index on: repo_id, day
func runDailyRepoBalance(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_repo_balance (repo_id, balance, day, currency, created_at)
					SELECT
						s.repo_id, 
						SUM((q.balance / q.sponsored_repos)) + COALESCE((
					    	SELECT dfl.balance 
					    	FROM daily_future_leftover dfl 
					    	WHERE dfl.repo_id = s.repo_id AND dfl.day = $4 AND dfl.currency = q.currency), 0) as balance, 
						$1 AS DAY,
						q.currency,
						$3 AS created_at
					FROM (
						SELECT ub.user_id, COUNT(*) AS sponsored_repos, -(MIN(ub.balance)) AS balance, ub.currency AS currency FROM user_balances ub
						JOIN sponsor_event s ON s.user_id = ub.user_id
						WHERE ub.balance_type = 'DAY'
							AND ub.day = $1
							AND (EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at)))/3600)::bigInt >= 24
						GROUP BY ub.user_id, ub.currency
					) AS q
					INNER JOIN sponsor_event s ON s.user_id = q.user_id
					INNER JOIN users u ON u.id = s.user_id
					WHERE u.role = 'USR'
					GROUP BY s.repo_id, q.currency`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	// $4 to add leftover from the day before yesterday to yesterday
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now, yesterdayStart.AddDate(0, 0, -1))
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

/*func runDailyEmailPayout(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_email_payout (email, balance, currency, day, created_at)
		SELECT res.git_email as email,
		       FLOOR(SUM(res.weight * drb.balance)) as balance,
		       drb.currency
		       $1 as day,
		       $3 as created_at
        FROM analysis_response res
            JOIN (SELECT req.id, req.repo_id FROM analysis_request req
                JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	WHERE date_to <= $2 GROUP BY repo_id)
                    AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                AS req ON res.analysis_request_id = req.id
            JOIN daily_repo_balance drb ON drb.repo_id = req.repo_id
        WHERE drb.day = $1
		GROUP BY res.git_email, drb.currency`)

	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}*/

func runDailyRepoWeight(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_repo_weight (repo_id, weight, day, created_at)
		SELECT req.repo_id as repo_id, 
		       SUM(res.weight) as weight,
		       $1 as day, 
		       $3 as created_at
        FROM analysis_response res
            JOIN (SELECT req.id, req.repo_id FROM analysis_request req
                JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	WHERE date_to <= $2 GROUP BY repo_id) 
                    AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                AS req ON res.analysis_request_id = req.id
			JOIN git_email g ON g.email = res.git_email
		WHERE g.token IS NOT NULL
        GROUP BY req.repo_id`)

	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

func runDailyUserPayout(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_user_payout (user_id, balance, currency, day, created_at)
		SELECT g.user_id as user_id, 
		       FLOOR(SUM(drb.balance * res.weight / drw.weight)) as balance, 
		       drb.currency,
		       $1 as day, 
		       $3 as created_at
        FROM analysis_response res
            JOIN (SELECT req.id, req.repo_id FROM analysis_request req
                JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	WHERE date_to <= $2 GROUP BY repo_id) 
                    AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                AS req ON res.analysis_request_id = req.id
            JOIN git_email g ON g.email = res.git_email
            JOIN daily_repo_weight drw ON drw.repo_id = req.repo_id
            JOIN daily_repo_balance drb ON drb.repo_id = req.repo_id
        WHERE drw.day = $1 AND drb.day = $1 AND g.token IS NOT NULL -- IS NULL changed to IS NOT NULL verify if correct
		GROUP BY g.user_id, drb.currency`)

	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

//if a repo gets funds, but no user is in our system, it goes to the leftover table and can be claimed later on
//by the first user that registers in our system.
func runDailyFutureLeftover(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_future_leftover (repo_id, balance, currency, day, created_at)
		SELECT drb.repo_id, drb.balance, drb.currency, $1 as day, $2 as created_at
        FROM daily_repo_balance drb
        LEFT JOIN daily_repo_weight drw ON drb.repo_id = drw.repo_id
        WHERE drw.repo_id IS NULL AND drb.day = $1`)

	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}

//return repos where the data_to is older than 5 days. This are repos where we can run the analysis again.
func runDailyAnalysisCheck(now time.Time, days int) ([]Repo, error) {
	s := `SELECT r.id, r.url, r.branch
            FROM repo r
                JOIN (SELECT req.id, req.repo_id, req.date_to FROM analysis_request req
                    JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	GROUP BY repo_id) 
                        AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                    AS req ON req.repo_id = r.id
			WHERE DATE_PART('day', AGE(req.date_to, $1)) < $2`
	rows, err := db.Query(s, now, days)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	repos := []Repo{}
	for rows.Next() {
		var repo Repo
		err = rows.Scan(&repo.Id, &repo.Url, &repo.Branch)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func runDailyTopupReminderUser() ([]User, error) {
	s := `SELECT u.id, u.sponsor_id, u.email, u.payment_cycle_id, u.stripe_id, u.stripe_payment_method
            FROM users u
                INNER JOIN payment_cycle pc ON u.payment_cycle_id = pc.id
			WHERE u.role='USR' AND pc.days_left <= 1 AND pc.freq != 0 AND (u.stripe_payment_method IS NOT NULL OR u.sponsor_id IS NOT NULL)`

	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	repos := []User{}
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.SponsorId, &user.Email, &user.PaymentCycleId, &user.StripeId, &user.PaymentMethod)
		if err != nil {
			return nil, err
		}
		repos = append(repos, user)
	}
	return repos, nil
}

// TODO: Thomas Bocek has to look into this one
/*func runDailyMarketing(yesterdayStart time.Time) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT STRING_AGG(r.name, ','), d.contributor_email, SUM(d.balance) as balance
            FROM daily_user_contribution d
            	INNER JOIN repo r ON d.repo_id=r.id
			WHERE d.day = $1 AND d.contributor_user_id IS NULL
			GROUP BY d.contributor_email`
	rows, err := db.Query(s, yesterdayStart)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var c Contribution
		err = rows.Scan(
			&c.RepoName,
			&c.ContributorEmail,
			&c.Balance,
			yesterdayStart)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}*/

// TODO: Thomas Bocek has to look into this one
/*func runDailyUserContribution(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_user_contribution(
                                    user_id,
                                    repo_id,
                                    contributor_email,
									contributor_name,
                                    contributor_weight,
                                    contributor_user_id,
                                    balance,
                                    balance_repo,
                                    day,
                                    created_at)
		   SELECT s.user_id as user_id,
		          s.repo_id as repo_id,
		          res.git_email as contributor_email,
		          string_agg(res.git_name, ',') as contributor_name,
	              AVG(res.weight) as contributor_weight,
	              g.user_id as contributor_user_id,
		          CASE WHEN g.user_id IS NULL THEN
		              	  FLOOR(SUM(drb.balance * res.weight / (drw.weight + res.weight)))
		              ELSE
		                  FLOOR(SUM(drb.balance * res.weight / drw.weight))
		              END as balance,
		          SUM(drb.balance) as balance_repo,
			      $1 as day,
                  $3 as created_at
			 FROM sponsor_event s
			     INNER JOIN daily_repo_balance drb ON drb.repo_id = s.repo_id
			     LEFT JOIN daily_repo_weight drw ON drw.repo_id = s.repo_id
                 LEFT JOIN (SELECT req.id, req.repo_id FROM analysis_request req
                     INNER JOIN (SELECT MAX(date_to) as date_to, ARRAY_AGG(date_from) as dates_from, repo_id
                                 FROM analysis_request
                                 WHERE date_to <= $2
                                 GROUP BY repo_id)
                         AS tmp ON tmp.date_to = req.date_to AND tmp.dates_from[1] = req.date_from AND tmp.repo_id = req.repo_id)
                     AS req ON s.repo_id = req.repo_id
			     LEFT JOIN analysis_response res ON res.analysis_request_id = req.id
                 LEFT JOIN git_email g ON g.email = res.git_email
			 WHERE drb.day=$1 AND drw.day=$1
			     AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
             GROUP BY s.user_id, s.repo_id, res.git_email, g.user_id`)
	if err != nil {
		return 0, err
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(yesterdayStart, yesterdayStop, now)
	if err != nil {
		return 0, err
	}
	return handleErr(res)
}*/

//*********************************************************************************
//**************************** Monthly Batchjob Payout ****************************
//*********************************************************************************

type PayoutCrypto struct {
	UserId   uuid.UUID
	Address  string
	Tea      int64
	Currency string
}

func monthlyBatchJobPayout() ([]PayoutCrypto, error) {
	s := `SELECT dup.user_id, wa.address, SUM(dup.balance), dup.currency 
		  FROM daily_user_payout dup 
		  JOIN wallet_address wa ON wa.user_id = dup.user_id AND ((wa.currency = dup.currency) OR (dup.currency = 'USD' AND wa.currency = 'ETH'))
		  WHERE wa.is_deleted = false
		  GROUP BY dup.user_id, dup.currency, wa.address`
	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var payoutsCrypto []PayoutCrypto
	for rows.Next() {
		var payoutCrypto PayoutCrypto
		err = rows.Scan(&payoutCrypto.UserId, &payoutCrypto.Address, &payoutCrypto.Tea, &payoutCrypto.Currency)
		if err != nil {
			return nil, err
		}
		payoutsCrypto = append(payoutsCrypto, payoutCrypto)
	}
	return payoutsCrypto, nil
}
