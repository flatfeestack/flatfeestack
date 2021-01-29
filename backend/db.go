package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"math/big"
	"time"
)

type User struct {
	Id                uuid.UUID `json:"id" sql:",type:uuid"`
	StripeId          *string   `json:"-"`
	Email             *string   `json:"email"`
	Subscription      *string   `json:"subscription"`
	SubscriptionState *string   `json:"subscription_state"`
	PayoutETH         *string   `json:"payout_eth"`
	Role              *string   `json:"role"`
	CreatedAt         time.Time
}

type SponsorEvent struct {
	Id          uuid.UUID `json:"id"`
	Uid         uuid.UUID `json:"uid"`
	RepoId      uuid.UUID `json:"repo_id"`
	EventType   uint8     `json:"event_type"`
	SponsorAt   time.Time `json:"created_at"`
	UnsponsorAt time.Time `json:"created_at"`
}

type Repo struct {
	Id          uuid.UUID `json:"id"`
	OrigId      uint32
	Url         *string `json:"html_url"`
	Name        *string `json:"full_name"`
	Description *string `json:"description"`
	CreatedAt   time.Time
}

type Payment struct {
	Id        uuid.UUID `json:"id"`
	Uid       uuid.UUID `json:"uid"`
	Amount    int64
	From      time.Time
	To        time.Time
	Sub       string
	CreatedAt time.Time
}

type UserAggBalance struct {
	PayoutEth            string      `json:"payout_eth"`
	Balance              int64       `json:"balance"`
	Emails               []string    `json:"email_list"`
	MonthlyUserPayoutIds []uuid.UUID `json:"monthly_user_payout_id_list"`
	CreatedAt            time.Time
}

type PayoutsRequest struct {
	MonthlyUserPayoutId uuid.UUID `json:"monthly-repo-balance-id"`
	BatchId             uuid.UUID `json:"batch-id"`
	ExchangeRate        big.Float
	CreatedAt           time.Time
}

type PayoutsResponse struct {
	BatchId    uuid.UUID
	TxHash     string
	Error      *string
	CreatedAt  time.Time
	PayoutWeis []PayoutWei
}

// FindByID returns a single user
func findUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow("SELECT id, stripe_id, email, subscription, subscription_state, payout_eth FROM users WHERE email=$1", email).
		Scan(&u.Id, &u.StripeId, &u.Email, &u.Subscription, &u.SubscriptionState, &u.PayoutETH)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

// FindByID returns a single user
func findUserByID(uid uuid.UUID) (*User, error) {
	var u User
	err := db.
		QueryRow("SELECT id, stripe_id, email, subscription, subscription_state, payout_eth FROM users WHERE id=$1", uid).
		Scan(&u.Id, &u.StripeId, &u.Email, &u.Subscription, &u.SubscriptionState, &u.PayoutETH)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &u, nil
	default:
		return nil, err
	}
}

// Save inserts a user into the database
func saveUser(user *User) error {
	stmt, err := db.Prepare("INSERT INTO users (id, email, stripe_id, payout_eth, subscription_state, created_at) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO users for %v statement failed: %v", user, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(user.Id, user.Email, user.StripeId, user.PayoutETH, user.SubscriptionState, user.CreatedAt)
	return handleErr(res, err, "INSERT INTO users", user)
}

func updateUser(user *User) error {
	stmt, err := db.Prepare("UPDATE users SET email=$1, stripe_id=$2, subscription=$3, subscription_state=$4, payout_eth=$5 WHERE id=$6")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(user.Email, user.StripeId, user.Subscription, user.SubscriptionState, user.PayoutETH, user.Id)
	return handleErr(res, err, "UPDATE users", user)
}

//sponsor events
func sponsor(event *SponsorEvent) (userError error, systemError error) {
	//first get last sponsored event to check if we need to sponsor or unsponsor
	//TODO: use mutex
	id, _, unsponsorAt, err := lastEventSponsoredRepo(event.Uid, event.RepoId)
	if err != nil {
		return nil, err
	}

	if id == nil && event.EventType == UNSPONSOR {
		return fmt.Errorf("we want to unsponsor, but we are currently not sponsoring this repo"), nil
	}

	if id != nil && event.EventType == UNSPONSOR && unsponsorAt.Year() != 9999 {
		return fmt.Errorf("we want to unsponsor, but we already unsponsored it"), nil
	}

	if id != nil && event.EventType == SPONSOR && event.SponsorAt.Before(*unsponsorAt) {
		return fmt.Errorf("we want to sponsor, but we are already sponsoring this repo: "+
			"sponsor_at: %v, unsponsor_at: %v, %v", event.SponsorAt, unsponsorAt, event.SponsorAt.Before(*unsponsorAt)), nil
	}

	if event.EventType == SPONSOR {
		//insert
		stmt, err := db.Prepare("INSERT INTO sponsor_event (id, user_id, repo_id, sponsor_at) VALUES ($1, $2, $3, $4)")
		if err != nil {
			return nil, fmt.Errorf("prepare INSERT INTO sponsor_event for %v statement event: %v", event, err)
		}
		defer stmt.Close()

		var res sql.Result
		res, err = stmt.Exec(event.Id, event.Uid, event.RepoId, event.SponsorAt)
		return nil, handleErr(res, err, "INSERT INTO sponsor_event", event)
	} else if event.EventType == UNSPONSOR {
		//update
		stmt, err := db.Prepare("UPDATE sponsor_event SET unsponsor_at=$1 WHERE id=$2 AND unsponsor_at = to_date('9999', 'YYYY')")
		if err != nil {
			return nil, fmt.Errorf("prepare UPDATE sponsor_event for %v statement failed: %v", id, err)
		}
		defer stmt.Close()

		var res sql.Result
		res, err = stmt.Exec(event.UnsponsorAt, id)
		return nil, handleErr(res, err, "UPDATE sponsor_event", id)
	} else {
		return nil, fmt.Errorf("unknown event type %v", event.EventType)
	}
}

// Repositories
func getSponsoredReposById(userId uuid.UUID) ([]Repo, error) {
	var repos []Repo
	sql := `SELECT r.id, r.orig_id, r.url, name, description 
            FROM sponsor_event s
            JOIN repo r ON s.repo_id=r.id 
			WHERE s.user_id=$1 AND s.unsponsor_at = to_date('9999', 'YYYY')`
	rows, err := db.Query(sql, userId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var repo Repo
		err = rows.Scan(&repo.Id, &repo.OrigId, &repo.Url, &repo.Name, &repo.Description)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func lastEventSponsoredRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
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

func saveRepo(repo *Repo) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO repo (id, orig_id, url, name, description, created_at) 
									VALUES ($1, $2, $3, $4, $5, $6)
									ON CONFLICT(url) DO UPDATE SET name=$4, description=$5`)
	if err != nil {
		return nil, fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(repo.Id, repo.OrigId, repo.Url, repo.Name, repo.Description, repo.CreatedAt)
	err = handleErr(res, err, "INSERT INTO repo", repo)
	if err != nil {
		return nil, fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}

	//"RETURNING" clause does not work with SQLite, so getting back with a select
	//its not transactional, but it does not matter, once written, the url and id do not change
	//TODO: now we use postgres exclusively, do returning again
	var id uuid.UUID
	err = db.
		QueryRow("SELECT id FROM repo WHERE repo.Url=$1", repo.Url).
		Scan(&id)
	switch err {
	case nil:
		return &id, nil
	default:
		return nil, err
	}
}

func findRepoByID(rid uuid.UUID) (*Repo, error) {
	var r Repo
	err := db.
		QueryRow("SELECT id, orig_id, url, name, description FROM repo WHERE id=$1", rid).
		Scan(&r.Id, &r.OrigId, &r.Name, &r.Url, &r.Description)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &r, nil
	default:
		return nil, err
	}
}

//connected emails
func findGitEmails(uid uuid.UUID) ([]string, error) {
	var emails []string
	sql := "SELECT email FROM git_email WHERE user_id=$1"
	rows, err := db.Query(sql, uid)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var email string
		err = rows.Scan(&email)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func saveGitEmail(id uuid.UUID, uid uuid.UUID, email string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO git_email(id, user_id, email, created_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(id, uid, email, now)
	return handleErr(res, err, "INSERT INTO git_email", email)
}

func deleteGitEmail(uid uuid.UUID, email string) error {
	//TODO: don't delete, just mark as deleted
	stmt, err := db.Prepare("DELETE FROM git_email WHERE email=$1 AND user_id=$2")
	if err != nil {
		return fmt.Errorf("prepare DELETE FROM git_email for %v statement event: %v", email, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(email, uid)
	return handleErr(res, err, "DELETE FROM git_email", email)
}

func saveAnalysisRequest(id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_request(id, repo_id, date_from, date_to, branch, created_at) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_request for %v statement event: %v", id, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(id, repo_id, date_from, date_to, branch, now)
	return handleErr(res, err, "INSERT INTO exchange", id)
}

func saveAnalysisResponse(aid uuid.UUID, w *FlatFeeWeight, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_response(id, analysis_request_id, git_email, weight, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v/%v statement event: %v", aid, w, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(uuid.New(), aid, w.Email, w.Weight, now)
	return handleErr(res, err, "INSERT INTO analysis_response", aid)
}

func savePayment(p *Payment) error {
	stmt, err := db.Prepare("INSERT INTO payments(id, user_id, date_from, date_to, sub, amount) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payments for %v statement event: %v", p.Id, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(p.Id, p.Uid, p.From, p.To, p.Sub, p.Amount)
	return handleErr(res, err, "INSERT INTO payments", p.Id)
}

func savePayoutsRequest(p *PayoutsRequest) error {
	stmt, err := db.Prepare(`
				INSERT INTO payouts_request(monthly_user_payout_id, batch_id, exchange_rate, created_at) 
				VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_request for %v statement event: %v", p.MonthlyUserPayoutId, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(p.MonthlyUserPayoutId, p.BatchId, p.ExchangeRate.String(), p.CreatedAt)
	return handleErr(res, err, "INSERT INTO payouts_request", p.MonthlyUserPayoutId)
}

func savePayoutsResponse(p *PayoutsResponse) error {
	pid := uuid.New()
	stmt, err := db.Prepare(`
				INSERT INTO payouts_response(id, batch_id, tx_hash, created_at) 
				VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_response for %v statement event: %v", p.BatchId, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(pid, p.BatchId, p.TxHash, p.CreatedAt)
	err = handleErr(res, err, "INSERT INTO payouts_response", p.BatchId)
	if err != nil {
		return err
	}
	for _, v := range p.PayoutWeis {
		err = savePayoutsResponseDetails(pid, &v)
		if err != nil {
			return err
		}
	}
	return nil
}

func savePayoutsResponseDetails(pid uuid.UUID, pwei *PayoutWei) error {
	stmt, err := db.Prepare(`
				INSERT INTO payouts_response_details(payouts_response_id, address, balance_wei, created_at) 
				VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_response_details for %v statement event: %v", pid, err)
	}
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(pid, pwei.Address, pwei.Balance.String(), timeNow())
	return handleErr(res, err, "INSERT INTO payouts_response_details", pid)
}

func getPendingPayouts() ([]UserAggBalance, error) {
	var userAggBalances []UserAggBalance
	//select monthly payments, but only those that do not have a payout entry
	query := `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(m.balance), ARRAY_AGG(m.id) as id_list
				FROM monthly_user_payout m 
			    JOIN users u ON m.user_id = u.id 
				LEFT JOIN payouts_request p ON p.monthly_user_payout_id = m.id
				WHERE p.id IS NULL
				GROUP BY u.payout_eth
				HAVING SUM(m.balance) > 10000`

	rows, err := db.Query(query)
	defer rows.Close()

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		for rows.Next() {
			var userAggBalance UserAggBalance
			err = rows.Scan(&userAggBalance.PayoutEth,
				pq.Array(&userAggBalance.Emails),
				&userAggBalance.Balance,
				pq.Array(&userAggBalance.MonthlyUserPayoutIds))
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

func getPaidPayouts() ([]UserAggBalance, error) {
	var userAggBalances []UserAggBalance
	//select monthly payments, but only those that do not have a payout entry
	query := `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(m.balance), ARRAY_AGG(m.id) as id_list
				FROM monthly_user_payout m 
			    JOIN users u ON m.user_id = u.id 
				JOIN payouts_request preq ON preq.monthly_user_payout_id = m.id
				JOIN payouts_response pres ON pres.batch_id = preq.batch_id
				GROUP BY u.payout_eth`

	rows, err := db.Query(query)
	defer rows.Close()

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		for rows.Next() {
			var userAggBalance UserAggBalance
			err = rows.Scan(&userAggBalance.PayoutEth,
				pq.Array(&userAggBalance.Emails),
				&userAggBalance.Balance,
				pq.Array(&userAggBalance.MonthlyUserPayoutIds))
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

func getLimboPayouts() ([]UserAggBalance, error) {
	var userAggBalances []UserAggBalance
	//select monthly payments, but only those that do not have a payout entry
	query := `SELECT u.payout_eth, ARRAY_AGG(u.email) email_list, SUM(m.balance), ARRAY_AGG(m.id) as id_list
				FROM monthly_user_payout m 
			    JOIN users u ON m.user_id = u.id 
				JOIN payouts_request preq ON preq.monthly_user_payout_id = m.id
				LEFT JOIN payouts_response pres ON pres.batch_id = preq.batch_id
				WHERE pres.id IS NULL
				GROUP BY u.payout_eth`

	rows, err := db.Query(query)
	defer rows.Close()

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		for rows.Next() {
			var userAggBalance UserAggBalance
			err = rows.Scan(&userAggBalance.PayoutEth,
				pq.Array(&userAggBalance.Emails),
				&userAggBalance.Balance,
				pq.Array(&userAggBalance.MonthlyUserPayoutIds))
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

func handleErr(res sql.Result, err error, info string, value interface{}) error {
	if err != nil {
		return fmt.Errorf("%v query %v failed: %v", info, value, err)
	}
	var nr int64
	nr, err = res.RowsAffected()
	if nr == 0 || err != nil {
		return fmt.Errorf("%v %v rows %v, affected or err: %v", info, nr, value, err)
	}
	return nil
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
