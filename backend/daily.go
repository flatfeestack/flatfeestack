package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
func runDailyRepoHours(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	//https://stackoverflow.com/questions/17833176/postgresql-days-months-years-between-two-dates
	stmt, err := db.Prepare(`INSERT INTO daily_repo_hours (user_id, repo_hours, day, created_at)
              SELECT s.user_id, 
                     SUM((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::bigint) as repo_hours, 
                     $1 as day, $3 as created_at
                FROM sponsor_event s 
                    INNER JOIN users u ON u.id = s.user_id
                    INNER JOIN payment_cycle pc ON u.payment_cycle_id = pc.id
                WHERE NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
                    AND pc.days_left > 0
                    AND u.role = 'USR'
                GROUP BY s.user_id`)
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

//This inserts a balance deduction for the user, than has the role "USR", has funds, and at least 1h of supported
//repo. The balance is negative, thus deducted.
//
//Running this twice wont work as we have a unique index on: user_id, day, balance_type
// ToDo: documentation
func runDailyUserBalance(yesterdayStart time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO user_balances (payment_cycle_id, user_id, balance, balance_type, currency, day, created_at)
		   select * from test(
		       $1, 
		       $2) 
		       f(payment_cycle_id uuid, user_id uuid, balance bigint, balance_type text, currency VARCHAR(16), day date, created_at TIMESTAMP with time zone)
`)
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

// ToDo: documentation
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

//Here we update the days left of the user. This calculates as the remaining balance divided by mUSDPerDay
//
//Running this twice is ok, as it will give a more accurate state
// TODO: update doc
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
//Here we calculate how much balance a repository gets. The calculation is based on the daily_repo_hours. So if
//a user has 3 repos with 72 repo hours, and supports repo X, then we calculate how much repo X gets from that user,
//which is 24h (the user supported for 24h) x 24/72 = 8h, which is 1/3 of his repo hours.
//
//Running this twice does not work as we have a unique index on: repo_id, day
//TODO: update doc
func runDailyRepoBalance(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_repo_balance (repo_id, balance, day, currency, created_at)
		   	SELECT
			    s.repo_id,
		   	    SUM((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at)))/3600)::bigint * q.dailyAmount / drh.repo_hours) + COALESCE((
		   	         SELECT dfl.balance 
		             FROM daily_future_leftover dfl 
		             WHERE dfl.repo_id = s.repo_id AND dfl.day = $1), 0) as balance, 
			    $1 as day,
			    q.currency,
				$3 as created_at
			FROM (
			         SELECT ub.user_id,-(min(ub.balance)) as dailyAmount, min(ub.currency) as currency from user_balances ub
			         JOIN sponsor_event s ON s.user_id = ub.user_id
			         WHERE ub.day = $1
			           AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
			           AND ub.balance_type = 'DAY'
			         GROUP BY ub.user_id
			     ) AS q
			         INNER JOIN sponsor_event s ON s.user_id = q.user_id
			         INNER JOIN users u ON u.id = s.user_id
			         INNER JOIN daily_repo_hours drh ON u.id = drh.user_id
			WHERE drh.day=$1 
				AND u.role = 'USR' -- Maybe don't needed because we check balance_type = DAY and only USR should be able to make sponsoring
			GROUP BY s.repo_id, q.currency`)
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

/*func runDailyEmailPayout(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_email_payout (email, balance, day, created_at)
		SELECT res.git_email as email,
		       FLOOR(SUM(res.weight * drb.balance)) as balance,
		       $1 as day,
		       $3 as created_at
        FROM analysis_response res
            JOIN (SELECT req.id, req.repo_id FROM analysis_request req
                JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	WHERE date_to <= $2 GROUP BY repo_id)
                    AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                AS req ON res.analysis_request_id = req.id
            JOIN daily_repo_balance drb ON drb.repo_id = req.repo_id
        WHERE drb.day = $1
		GROUP BY res.git_email`)

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
		WHERE g.token IS NULL
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

//ToDo: needs to be better testet, but needs more testdata and users
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
// Todo: update doc
func runDailyFutureLeftover(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO daily_future_leftover (repo_id, balance, currency, day, created_at)
		SELECT ar.repo_id, drb.balance, drb.currency, $2 as day, $3 as created_at
		FROM analysis_request ar 
		join analysis_response ar2 on ar.id = ar2.analysis_request_id
		join daily_repo_balance drb on drb.repo_id = ar.repo_id
		where ar2.git_email not in (select g.email from git_email g join daily_user_payout dup on dup.user_id = g.user_id where dup.day = $1)
		group by ar.repo_id, drb.currency, drb.balance`)

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

// ToDo: reminder for each currency if needed
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

//ToDo: marketing for each currency if needed
func runDailyMarketing(yesterdayStart time.Time) ([]Contribution, error) {
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
}

//ToDo: do we need this daily payout anymore if we make everything over batch job?
func getDailyPayouts(s string) ([]UserAggBalance, error) {
	day := timeDay(-60, timeNow()) //day -2 month
	var userAggBalances []UserAggBalance
	//select monthly payments, but only those that do not have a payout entry
	var query string
	var rows *sql.Rows
	var err error
	switch s {
	case "pending":
		query = `SELECT u.id, u.payout_eth, ARRAY_AGG(DISTINCT(u.email)) AS email_list, SUM(dup.balance), ARRAY_AGG(dup.id) AS id_list
				FROM daily_user_payout dup 
			    JOIN users u ON dup.user_id = u.id 
				LEFT JOIN payouts_request p ON p.daily_user_payout_id = dup.id
				WHERE p.id IS NULL AND dup.day < $1
				GROUP BY u.id, u.payout_eth`
		rows, err = db.Query(query, day)
	case "paid":
		query = `SELECT u.id, u.payout_eth, ARRAY_AGG(DISTINCT(u.email)) AS email_list, SUM(dup.balance), ARRAY_AGG(dup.id) AS id_list
				FROM daily_user_payout dup
			    JOIN users u ON dup.user_id = u.id 
				JOIN payouts_request preq ON preq.daily_user_payout_id = dup.id
				JOIN payouts_response pres ON pres.batch_id = preq.batch_id
                WHERE pres.error is NULL
				GROUP BY u.id, u.payout_eth`
		rows, err = db.Query(query)
	default: //limbo
		query = `SELECT u.id, u.payout_eth, ARRAY_AGG(DISTINCT(u.email)) AS email_list, SUM(dup.balance), ARRAY_AGG(dup.id) AS id_list
				FROM daily_user_payout dup
			    JOIN users u ON dup.user_id = u.id 
				JOIN payouts_request preq ON preq.daily_user_payout_id = dup.id
				LEFT JOIN payouts_response pres ON pres.batch_id = preq.batch_id
				WHERE pres.id IS NULL OR pres.error is NOT NULL
				GROUP BY u.id, u.payout_eth`
		rows, err = db.Query(query)
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer closeAndLog(rows)
		for rows.Next() {
			var userAggBalance UserAggBalance
			err = rows.Scan(
				&userAggBalance.UserId,
				&userAggBalance.PayoutEth,
				pq.Array(&userAggBalance.Emails),
				&userAggBalance.Balance,
				pq.Array(&userAggBalance.DailyUserPayoutIds))
			if err != nil {
				return nil, err
			}
			userAggBalances = append(userAggBalances, userAggBalance)
		}
		return userAggBalances, nil
	default:
		return nil, err
	}
}

// TODO: check case what the balance should be (for registered and not registered contributors)
func runDailyUserContribution(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
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
}

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
	s := `SELECT dup.user_id, wa.address, SUM(dup.balance), dup.currency FROM daily_user_payout dup 
		  JOIN wallet_address wa ON wa.user_id = dup.user_id AND wa.currency = dup.currency
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
