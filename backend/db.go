package main

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"io"
	"log"
	"math/big"
	"strconv"
	"time"
)

type User struct {
	Id             uuid.UUID `json:"id" sql:",type:uuid"`
	StripeId       *string   `json:"-"`
	PaymentCycleId uuid.UUID
	Balance        *int64  `json:"balance"`
	Email          *string `json:"email"`
	InviteEmail    *string `json:"invite_email"`
	Name           *string `json:"name"`
	Image          *string `json:"image"`
	PayoutETH      *string `json:"payout_eth"`
	PaymentMethod  *string `json:"payment_method"`
	Last4          *string `json:"last4"`
	Seats          int     `json:"seats"`
	Freq           int     `json:"freq"`
	Token          *string `json:"token"`
	Role           *string `json:"role"`
	CreatedAt      time.Time
}

type SponsorEvent struct {
	Id          uuid.UUID `json:"id"`
	Uid         uuid.UUID `json:"uid"`
	RepoId      uuid.UUID `json:"repo_id"`
	EventType   uint8     `json:"event_type"`
	SponsorAt   time.Time `json:"sponsor_at"`
	UnsponsorAt time.Time `json:"unsponsor_at"`
}

type Repo struct {
	Id          uuid.UUID         `json:"uuid"`
	OrigId      uint64            `json:"id"`
	Url         *string           `json:"html_url"`
	GitUrl      *string           `json:"clone_url"`
	Branch      *string           `json:"default_branch"`
	Name        *string           `json:"full_name"`
	Description *string           `json:"description"`
	Tags        map[string]string `json:"tags"`
	Score       uint32            `json:"score"`
	Source      *string           `json:"source"`
	CreatedAt   time.Time         `json:"created_at"`
}

type Payment struct {
	Id        uuid.UUID `json:"id"`
	Uid       uuid.UUID `json:"uid"`
	Amount    int64
	Seats     int64
	From      time.Time
	To        time.Time
	Sub       string
	CreatedAt time.Time
}

type UserAggBalance struct {
	PayoutEth          string      `json:"payout_eth"`
	Balance            int64       `json:"balance"`
	Emails             []string    `json:"email_list"`
	DailyUserPayoutIds []uuid.UUID `json:"daily_user_payout_id_list"`
	CreatedAt          time.Time
}

type PayoutsRequest struct {
	DailyUserPayoutId uuid.UUID `json:"daily-repo-balance-id"`
	BatchId           uuid.UUID `json:"batch-id"`
	ExchangeRate      big.Float
	CreatedAt         time.Time
}

type PayoutsResponse struct {
	BatchId    uuid.UUID
	TxHash     string
	Error      *string
	CreatedAt  time.Time
	PayoutWeis []PayoutWei
}

type GitEmail struct {
	Email       string     `json:"email"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type UserBalance struct {
	paymentCycleId uuid.UUID
	userId         uuid.UUID
	balance        int64
	day            time.Time
	balanceType    string
	createdAt      time.Time
}

func findUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow(`SELECT id, stripe_id, 
                                stripe_payment_method, stripe_balance, seats, freq,
                                stripe_last4, email, name, image, payout_eth, role 
                         FROM users WHERE email=$1`, email).
		Scan(&u.Id, &u.StripeId,
			&u.PaymentMethod, &u.Balance, &u.Seats, &u.Freq, &u.Last4, &u.Email, &u.Name,
			&u.Image, &u.PayoutETH, &u.Role)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func findUserById(uid uuid.UUID) (*User, error) {
	var u User
	err := db.
		QueryRow(`SELECT id, stripe_id, 
                                stripe_payment_method, stripe_balance, seats, freq, 
                                stripe_last4, email, name, image, payout_eth, role 
                         FROM users WHERE id=$1`, uid).
		Scan(&u.Id, &u.StripeId,
			&u.PaymentMethod, &u.Balance, &u.Seats, &u.Freq, &u.Last4, &u.Email, &u.Name,
			&u.Image, &u.PayoutETH, &u.Role)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

func insertUser(user *User, token string) error {
	stmt, err := db.Prepare("INSERT INTO users (id, email, stripe_id, payout_eth, token, created_at) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(user.Id, user.Email, user.StripeId, user.PayoutETH, token, user.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateUser(user *User) error {
	stmt, err := db.Prepare(`UPDATE users SET 
                                           stripe_id=$1, payout_eth=$2,  
                                           stripe_payment_method=$3, stripe_balance=$4, 
                                           stripe_last4=$5, seats=$6, freq=$7
                                    WHERE id=$8`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(user.StripeId, user.PayoutETH, user.PaymentMethod, user.Balance, user.Last4, user.Seats, user.Freq, user.Id)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateUserName(uid uuid.UUID, name string) error {
	stmt, err := db.Prepare("UPDATE users SET name=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(name, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateUserImage(uid uuid.UUID, data string) error {
	stmt, err := db.Prepare("UPDATE users SET image=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(data, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateUserMode(uid uuid.UUID, mode string) error {
	stmt, err := db.Prepare("UPDATE users SET role=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(mode, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//*********************************************************************************
//******************************* Sponsoring **************************************
//*********************************************************************************
func insertOrUpdateSponsor(event *SponsorEvent) (userError error, systemError error) {
	//first get last sponsored event to check if we need to insertOrUpdateSponsor or unsponsor
	//TODO: use mutex
	id, _, unsponsorAt, err := findLastEventSponsoredRepo(event.Uid, event.RepoId)
	if err != nil {
		return nil, err
	}

	if id == nil && event.EventType == Inactive {
		return fmt.Errorf("we want to unsponsor, but we are currently not sponsoring this repo"), nil
	}

	if id != nil && event.EventType == Inactive && unsponsorAt.Year() != 9999 {
		return fmt.Errorf("we want to unsponsor, but we already unsponsored it"), nil
	}

	if id != nil && event.EventType == Active && event.SponsorAt.Before(*unsponsorAt) {
		return fmt.Errorf("we want to insertOrUpdateSponsor, but we are already sponsoring this repo: "+
			"sponsor_at: %v, unsponsor_at: %v, %v", event.SponsorAt, unsponsorAt, event.SponsorAt.Before(*unsponsorAt)), nil
	}

	if event.EventType == Active {
		//insert
		stmt, err := db.Prepare("INSERT INTO sponsor_event (id, user_id, repo_id, sponsor_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return nil, fmt.Errorf("prepare INSERT INTO sponsor_event for %v statement event: %v", event, err)
		}
		defer closeAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.SponsorAt)
		if err != nil {
			return nil, err
		}
		return nil, handleErrMustInsertOne(res)
	} else if event.EventType == Inactive {
		//update
		stmt, err := db.Prepare("UPDATE sponsor_event SET unsponsor_at=$1 WHERE id=$2 AND unsponsor_at = to_date('9999', 'YYYY')")
		if err != nil {
			return nil, fmt.Errorf("prepare UPDATE sponsor_event for %v statement failed: %v", id, err)
		}
		defer closeAndLog(stmt)

		var res sql.Result
		res, err = stmt.Exec(event.UnsponsorAt, id)
		if err != nil {
			return nil, err
		}
		return nil, handleErrMustInsertOne(res)
	} else {
		return nil, fmt.Errorf("unknown event type %v", event.EventType)
	}
}

func findLastEventSponsoredRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
	var sponsorAt *time.Time
	var unsponsorAt *time.Time
	var id *uuid.UUID
	err := db.
		QueryRow(`SELECT id, sponsor_at, unsponsor_at
			      		 FROM sponsor_event 
						 WHERE user_id=$1 AND repo_id=$2 AND sponsor_at=
						     (SELECT max(sponsor_at) FROM sponsor_event WHERE user_id=$1 AND repo_id=$2)`,
			uid, rid).Scan(&id, &sponsorAt, &unsponsorAt)
	switch err {
	case sql.ErrNoRows:
		return nil, nil, nil, nil
	case nil:
		return id, sponsorAt, unsponsorAt, nil
	default:
		return nil, nil, nil, err
	}
}

// Repositories and Sponsors
func findSponsoredReposById(userId uuid.UUID) ([]Repo, error) {
	var repos []Repo
	s := `SELECT r.id, r.orig_id, r.url, r.git_url, r.branch, r.name, r.description, r.tags
            FROM sponsor_event s
            JOIN repo r ON s.repo_id=r.id 
			WHERE s.user_id=$1 AND s.unsponsor_at = to_date('9999', 'YYYY')`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var repo Repo
		var b []byte
		err = rows.Scan(&repo.Id, &repo.OrigId, &repo.Url, &repo.GitUrl, &repo.Branch, &repo.Name, &repo.Description, &b)
		if err != nil {
			return nil, err
		}
		d := gob.NewDecoder(bytes.NewReader(b))
		err = d.Decode(&repo.Tags)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

//*********************************************************************************
//******************************* Repository **************************************
//*********************************************************************************
func insertOrUpdateRepo(repo *Repo) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) 
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
									ON CONFLICT(url) DO UPDATE SET name=$4, description=$5 RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer closeAndLog(stmt)

	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err = e.Encode(repo.Tags)
	if err != nil {
		return nil, err
	}

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(repo.Id, repo.OrigId, repo.Url, repo.GitUrl, repo.Branch, repo.Name, repo.Description, b.Bytes(), repo.Score, repo.Source, repo.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil
}

func findRepoById(repoId uuid.UUID) (*Repo, error) {
	var r Repo
	var b []byte
	err := db.
		QueryRow("SELECT id, orig_id, url, git_url, branch, name, description, tags FROM repo WHERE id=$1", repoId).
		Scan(&r.Id, &r.OrigId, &r.Url, &r.GitUrl, &r.Branch, &r.Name, &r.Description, &b)

	d := gob.NewDecoder(bytes.NewReader(b))
	err = d.Decode(&r.Tags)
	if err != nil {
		return nil, err
	}

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &r, nil
	default:
		return nil, err
	}
}

//*********************************************************************************
//*******************************  Connected emails *******************************
//*********************************************************************************
func findGitEmailsByUserId(uid uuid.UUID) ([]GitEmail, error) {
	var gitEmails []GitEmail
	s := "SELECT email, confirmed_at, created_at FROM git_email WHERE user_id=$1"
	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var gitEmail GitEmail
		err = rows.Scan(&gitEmail.Email, &gitEmail.ConfirmedAt, &gitEmail.CreatedAt)
		if err != nil {
			return nil, err
		}
		gitEmails = append(gitEmails, gitEmail)
	}
	return gitEmails, nil
}

func insertGitEmail(id uuid.UUID, uid uuid.UUID, email string, token string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO git_email(id, user_id, email, token, created_at) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, uid, email, token, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func confirmGitEmail(email string, token string, now time.Time) error {
	stmt, err := db.Prepare("UPDATE git_email SET token=NULL, confirmed_at=$1 WHERE email=$2 AND token=$3")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(now, email, token)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func deleteGitEmail(uid uuid.UUID, email string) error {
	//TODO: don't delete, just mark as deleted
	stmt, err := db.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(email, uid)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//*********************************************************************************
//******************************* Analysis Requests *******************************
//*********************************************************************************
func insertAnalysisRequest(id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_request(id, repo_id, date_from, date_to, branch, created_at) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_request for %v statement event: %v", id, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id, repo_id, date_from, date_to, branch, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertAnalysisResponse(aid uuid.UUID, w *FlatFeeWeight, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_response(id, analysis_request_id, git_email, weight, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v/%v statement event: %v", aid, w, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uuid.New(), aid, w.Email, w.Weight, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//*********************************************************************************
//******************************* Payments ****************************************
//*********************************************************************************
func insertUserBalance(ub UserBalance) error {
	stmt, err := db.Prepare(`INSERT INTO user_balances(
                                           payment_cycle_id, user_id, balance, balance_type, day, created_at) 
                                    VALUES($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_balances for %v/%v statement event: %v", ub.userId, ub.paymentCycleId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(ub.paymentCycleId, ub.userId, ub.balance, ub.balanceType, ub.day, ub.createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func findSumUsersBalance(paymentCycleId uuid.UUID) (int64, error) {
	var sum int64
	err := db.
		QueryRow(`SELECT sum(balance) FROM user_balances
             WHERE u.payment_cycle_id = $1`, paymentCycleId).
		Scan(&sum)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return sum, nil
	default:
		return 0, err
	}
}

func findSumUserBalance(paymentCycleId uuid.UUID, userId uuid.UUID) (int64, error) {
	var sum int64
	err := db.
		QueryRow(`SELECT sum(balance) FROM user_balances WHERE payment_cycle_id = $1 and user_id = $2`, paymentCycleId, userId).
		Scan(&sum)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return sum, nil
	default:
		return 0, err
	}
}

//*********************************************************************************
//******************************* Payouts *****************************************
//*********************************************************************************
func insertPayoutsRequest(p *PayoutsRequest) error {
	stmt, err := db.Prepare(`
				INSERT INTO payouts_request(daily_user_payout_id, batch_id, exchange_rate, created_at) 
				VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_request for %v statement event: %v", p.DailyUserPayoutId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(p.DailyUserPayoutId, p.BatchId, p.ExchangeRate.String(), p.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertPayoutsResponse(p *PayoutsResponse) error {
	pid := uuid.New()
	stmt, err := db.Prepare(`
				INSERT INTO payouts_response(id, batch_id, tx_hash, error, created_at) 
				VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_response for %v statement event: %v", p.BatchId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(pid, p.BatchId, p.TxHash, p.Error, p.CreatedAt)
	if err != nil {
		return err
	}
	err = handleErrMustInsertOne(res)
	if err != nil {
		return err
	}
	for _, v := range p.PayoutWeis {
		err = insertPayoutsResponseDetails(pid, &v)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertPayoutsResponseDetails(pid uuid.UUID, pwei *PayoutWei) error {
	stmt, err := db.Prepare(`
				INSERT INTO payouts_response_details(payouts_response_id, address, balance_wei, created_at) 
				VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_response_details for %v statement event: %v", pid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(pid, pwei.Address, pwei.Balance.String(), timeNow())
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//*********************************************************************************
//******************************* Daily calculations ******************************
//*********************************************************************************

func runDailyRepoHours(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	//https://stackoverflow.com/questions/17833176/postgresql-days-months-years-between-two-dates
	stmt, err := db.Prepare(`INSERT INTO daily_repo_hours (user_id, repo_hours, day, created_at)
              SELECT s.user_id, 
                     SUM((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at))) / 3600)::int) as repo_hours, 
                     $1 as day, $3 as created_at
                FROM sponsor_event s 
                    JOIN users u ON u.id = s.user_id
                    LEFT JOIN user_balances ub ON u.payment_cycle_id = ub.payment_cycle_id
                WHERE u.sponsor_id IS NULL
                    AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
                HAVING SUM(ub.balance) >= ` + strconv.Itoa(mUSDPerDay) + `
                GROUP by s.user_id`)
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

func runDailyUserBalance(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	sUSDPerHour := strconv.Itoa(-mUSDPerHour) //we deduct
	stmt, err := db.Prepare(`INSERT INTO user_balances (user_id, balance, day, created_at)
		   SELECT user_id, 
				  SUM(((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at)))/3600)::bigint * ` + sUSDPerHour + `) / drh.repo_hours) + COALESCE((
		             SELECT dfl.balance 
		             FROM daily_future_leftover dfl 
		             WHERE dfl.repo_id = s.repo_id AND dfl.day = $1), 0) as balance, 
			      $1 as day, 
                  $3 as created_at
			 FROM sponsor_event s
			     JOIN users u ON u.id = s.user_id 
			     JOIN daily_repo_hours drh ON u.id = drh.user_id
                 LEFT JOIN user_balances ub ON u.payment_cycle_id = ub.payment_cycle_id
			 WHERE u.sponsor_id IS NULL AND drh.day=$1 
			     AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
             HAVING SUM(ub.balance) >= ` + strconv.Itoa(mUSDPerDay) + `
			 GROUP by s.user_id`)

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

//TODO: limit user to 10000 repos
//we can support up to 1000 (1h) - 27500 (24h) repos until the precision makes the distribution of 0
func runDailyRepoBalance(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
	sUSDPerHour := strconv.Itoa(mUSDPerHour)
	stmt, err := db.Prepare(`INSERT INTO daily_repo_balance (repo_id, balance, day, created_at)
		   SELECT repo_id, 
				  SUM(((EXTRACT(epoch from age(LEAST($2, s.unsponsor_at), GREATEST($1, s.sponsor_at)))/3600)::bigint * ` + sUSDPerHour + `) / drh.repo_hours) + COALESCE((
		             SELECT dfl.balance 
		             FROM daily_future_leftover dfl 
		             WHERE dfl.repo_id = s.repo_id AND dfl.day = $1), 0) as balance, 
			      $1 as day, 
                  $3 as created_at
			 FROM sponsor_event s
			     JOIN users u ON u.id = s.user_id 
			     JOIN daily_repo_hours drh ON u.id = drh.user_id
                 LEFT JOIN user_balances ub ON u.payment_cycle_id = ub.payment_cycle_id
			 WHERE u.sponsor_id IS NULL AND drh.day=$1 
			     AND NOT((s.sponsor_at<$1 AND s.unsponsor_at<$1) OR (s.sponsor_at>=$2 AND s.unsponsor_at>=$2))
             HAVING SUM(ub.balance) >= ` + strconv.Itoa(mUSDPerDay) + `
			 GROUP by s.repo_id`)

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

func runDailyEmailPayout(yesterdayStart time.Time, yesterdayStop time.Time, now time.Time) (int64, error) {
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
}

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
	stmt, err := db.Prepare(`INSERT INTO daily_user_payout (user_id, balance, day, created_at)
		SELECT g.user_id as user_id, 
		       FLOOR(SUM(drb.balance * res.weight / drw.weight)) as balance, 
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
        WHERE drw.day = $1 AND drb.day = $1
		GROUP BY g.user_id`)

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
	stmt, err := db.Prepare(`INSERT INTO daily_future_leftover (repo_id, balance, day, created_at)
		SELECT drb.repo_id, drb.balance, $2 as day, $3 as created_at
        FROM daily_repo_balance drb
            LEFT JOIN daily_repo_weight drw ON drb.repo_id = drw.repo_id
        WHERE drw.repo_id IS NULL AND drb.day = $1`)

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
	s := `SELECT r.id, r.url
            FROM repo r
                JOIN (SELECT req.id, req.repo_id, req.date_to FROM analysis_request req
                    JOIN (SELECT MAX(date_to) as date_to, repo_id FROM analysis_request	GROUP BY repo_id) 
                        AS tmp ON tmp.date_to = req.date_to AND tmp.repo_id = req.repo_id)
                    AS req ON req.repo_id = r.id
			WHERE DATE_PART('day', AGE(req.date_to, $1)) > $2`
	rows, err := db.Query(s, now, days)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var repos []Repo
	for rows.Next() {
		var repo Repo
		err = rows.Scan(&repo.Id, &repo.Url)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func getDailyPayouts(s string) ([]UserAggBalance, error) {
	var userAggBalances []UserAggBalance
	//select monthly payments, but only those that do not have a payout entry
	var query string
	switch s {
	case "pending":
		query = `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(dup.balance), ARRAY_AGG(dup.id) as id_list
				FROM daily_user_payout dup 
			    JOIN users u ON dup.user_id = u.id 
				LEFT JOIN payouts_request p ON p.daily_user_payout_id = dup.id
				WHERE p.id IS NULL
				GROUP BY u.payout_eth`
	case "paid":
		query = `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(dup.balance), ARRAY_AGG(dup.id) as id_list
				FROM daily_user_payout dup
			    JOIN users u ON dup.user_id = u.id 
				JOIN payouts_request preq ON preq.daily_user_payout_id = dup.id
				JOIN payouts_response pres ON pres.batch_id = preq.batch_id
                WHERE pres.error is NULL
				GROUP BY u.payout_eth`
	default: //limbo
		query = `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(dup.balance), ARRAY_AGG(dup.id) as id_list
				FROM daily_user_payout dup
			    JOIN users u ON dup.user_id = u.id 
				JOIN payouts_request preq ON preq.daily_user_payout_id = dup.id
				LEFT JOIN payouts_response pres ON pres.batch_id = preq.batch_id
				WHERE pres.id IS NULL OR pres.error is NOT NULL
				GROUP BY u.payout_eth`
	}
	rows, err := db.Query(query)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		defer closeAndLog(rows)
		for rows.Next() {
			var userAggBalance UserAggBalance
			err = rows.Scan(&userAggBalance.PayoutEth,
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

func handleErrMustInsertOne(res sql.Result) error {
	nr, err := res.RowsAffected()
	if nr == 0 || err != nil {
		return err
	}
	if nr != 1 {
		return fmt.Errorf("Only 1 row needs to be affacted, got %v", nr)
	}
	return nil
}

func handleErr(res sql.Result) (int64, error) {
	nr, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return nr, nil
}

// stringPointer connection with postgres db
func initDb() *sql.DB {
	// Open the connection
	db, err := sql.Open(opts.DBDriver, opts.DBPath)
	if err != nil {
		panic(err)
	}

	//we wait for ten seconds to connect
	err = db.Ping()
	now := time.Now()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(time.Now()) {
		// check the connection
		err = db.Ping()
		time.Sleep(time.Second)
	}
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected!")
	return db
}

func closeAndLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("could not close: %v", err)
	}
}
