package main

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"math/big"
	"strings"
	"time"
)

type User struct {
	Id             uuid.UUID  `json:"id" sql:",type:uuid"`
	SponsorId      *uuid.UUID `json:"sponsor_id" sql:",type:uuid"`
	InvitedEmail   *string    `json:"invited_email"`
	StripeId       *string    `json:"-"`
	PaymentCycleId uuid.UUID
	Email          string  `json:"email"`
	Name           *string `json:"name"`
	Image          *string `json:"image"`
	PayoutETH      *string `json:"payout_eth"`
	PaymentMethod  *string `json:"payment_method"`
	Last4          *string `json:"last4"`
	Token          *string `json:"token"`
	Role           *string `json:"role"`
	CreatedAt      time.Time
	Claims         *TokenClaims
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

type UserAggBalance struct {
	UserId             uuid.UUID   `json:"userId"`
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

// PayoutRequestDB New for Crypto, USD pay in should be migrated
type PayoutRequestDB struct {
	UserId       uuid.UUID
	BatchId      uuid.UUID
	Currency     string
	ExchangeRate big.Float
	Tea          int64
	Address      string
	CreatedAt    time.Time
}

type PayoutsResponse struct {
	BatchId    uuid.UUID
	TxHash     string
	Error      *string
	CreatedAt  time.Time
	PayoutWeis []PayoutWei
}

// PayoutsResponseDB New for Crypto, USD pay in should be migrated
type PayoutResponseDB struct {
	BatchId   uuid.UUID
	TxHash    string
	Error     *string
	CreatedAt time.Time
	Payouts   PayoutResponseNew
}

type GitEmail struct {
	Email       string     `json:"email"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type UserBalanceCore struct {
	UserId  uuid.UUID `json:"userId"`
	Balance int64     `json:"balance"`
}

type UserBalance struct {
	UserId         uuid.UUID  `json:"userId"`
	Balance        int64      `json:"balance"`
	PaymentCycleId uuid.UUID  `json:"paymentCycleId"`
	FromUserId     *uuid.UUID `json:"fromUserId"`
	BalanceType    string     `json:"balanceType"`
	Currency       string     `json:"currency"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type UserStatus struct {
	UserId   uuid.UUID `json:"userId"`
	Email    string    `json:"email"`
	Name     *string   `json:"name,omitempty"`
	DaysLeft int       `json:"daysLeft"`
}

type PaymentCycle struct {
	Id       uuid.UUID `json:"id"`
	Seats    int       `json:"seats"`
	Freq     int       `json:"freq"`
	DaysLeft int       `json:"daysLeft"`
}

type Contribution struct {
	UserEmail         string    `json:"userEmail"`
	UserName          string    `json:"userName"`
	RepoName          string    `json:"repoName"`
	ContributorEmail  *string   `json:"contributorEmail"`
	ContributorWeight *float64  `json:"contributorWeight"`
	FlatFeeStackUser  bool      `json:"isFlatFeeStackUser"`
	Balance           *int64    `json:"balance"`
	BalanceRepo       int64     `json:"balanceRepo"`
	Day               time.Time `json:"day"`
}

type Wallet struct {
	Id       uuid.UUID `json:"id"`
	Currency string    `json:"currency"`
	Address  string    `json:"address"`
}

func findAllUsers() ([]User, error) {
	users := []User{}
	s := `SELECT email from users`
	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func findUserByEmail(email string) (*User, error) {
	var u User
	err := db.
		QueryRow(`SELECT id, stripe_id, invited_email, stripe_payment_method, payment_cycle_id,
                                stripe_last4, email, name, image, payout_eth, role 
                         FROM users WHERE email=$1`, email).
		Scan(&u.Id, &u.StripeId, &u.InvitedEmail,
			&u.PaymentMethod, &u.PaymentCycleId, &u.Last4, &u.Email, &u.Name,
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
		QueryRow(`SELECT id, stripe_id, invited_email, stripe_payment_method, payment_cycle_id,
                                stripe_last4, email, name, image, payout_eth, role 
                         FROM users WHERE id=$1`, uid).
		Scan(&u.Id, &u.StripeId, &u.InvitedEmail,
			&u.PaymentMethod, &u.PaymentCycleId, &u.Last4, &u.Email, &u.Name,
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
                                           stripe_payment_method=$3, 
                                           stripe_last4=$4
                                    WHERE id=$5`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", user, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(user.StripeId, user.PayoutETH, user.PaymentMethod, user.Last4, user.Id)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updatePaymentCycleId(uid uuid.UUID, paymentCycleId *uuid.UUID, sponsorId *uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE users SET payment_cycle_id=$1, sponsor_id = $2 WHERE id=$3")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(paymentCycleId, sponsorId, uid)
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

func updateDbSeats(paymentCycleId uuid.UUID, seats int) error {
	stmt, err := db.Prepare("UPDATE payment_cycle SET seats=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle for %v statement failed: %v", paymentCycleId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(seats, paymentCycleId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateInvitedEmail(invitedEmail *string, userId uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE users SET invited_email=$1 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(invitedEmail, userId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func findWalletsByUserId(uid uuid.UUID) ([]Wallet, error) {
	var userWallets []Wallet
	s := "SELECT id, currency, address FROM wallet_address WHERE user_id=$1 AND is_deleted = false"
	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var userWallet Wallet
		err = rows.Scan(&userWallet.Id, &userWallet.Currency, &userWallet.Address)
		if err != nil {
			return nil, err
		}
		userWallets = append(userWallets, userWallet)
	}
	return userWallets, nil
}

func insertWallet(uid uuid.UUID, currency string, address string, isDeleted bool) error {
	stmt, err := db.Prepare("INSERT INTO wallet_address(user_id, currency, address, is_deleted) VALUES($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid, strings.ToUpper(currency), address, isDeleted)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func deleteWallet(id uuid.UUID) error {
	stmt, err := db.Prepare("UPDATE wallet_address set is_deleted = true WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(id)
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

func findSponsoredReposByOrgId(orgEmail string) ([]Repo, error) {
	repos := []Repo{}
	s := `SELECT r.id, r.orig_id, r.url, r.git_url, r.branch, r.name, r.description, r.tags, COUNT(r.id) as score
            FROM sponsor_event s
            INNER JOIN repo r ON s.repo_id=r.id 
            INNER JOIN users u ON s.user_id=u.id
			WHERE u.invited_email=$1 AND s.unsponsor_at = to_date('9999', 'YYYY')
			GROUP BY r.id ORDER BY score DESC`
	rows, err := db.Query(s, orgEmail)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var repo Repo
		var b []byte
		err = rows.Scan(&repo.Id, &repo.OrigId, &repo.Url, &repo.GitUrl, &repo.Branch, &repo.Name, &repo.Description, &b, &repo.Score)
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

func findSponsoredReposById(userId uuid.UUID) ([]Repo, error) {
	//we want to send back an empty array, don't change
	repos := []Repo{}
	s := `SELECT r.id, r.orig_id, r.url, r.git_url, r.branch, r.name, r.description, r.tags
            FROM sponsor_event s
            INNER JOIN repo r ON s.repo_id=r.id 
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

func findMyContributions(contributorUserId uuid.UUID) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT u.name, r.name, d.contributor_email, d.contributor_weight, d.contributor_user_id, d.balance, d.balance_repo, d.day
            FROM daily_user_contribution d
                INNER JOIN users u ON d.user_id = u.id
                INNER JOIN repo r ON d.repo_id=r.id 
			WHERE d.contributor_user_id=$1`
	rows, err := db.Query(s, contributorUserId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var c Contribution
		var cUuid *uuid.UUID
		err = rows.Scan(
			&c.UserName,
			&c.RepoName,
			&c.ContributorEmail,
			&c.ContributorWeight,
			&cUuid,
			&c.Balance,
			&c.BalanceRepo,
			&c.Day)
		if cUuid == nil {
			c.FlatFeeStackUser = false
		} else {
			c.FlatFeeStackUser = true
		}
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

func findUserContributions(userId uuid.UUID) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT r.name, d.contributor_email, d.contributor_weight, d.contributor_user_id, d.balance, d.balance_repo, d.day
            FROM daily_user_contribution d
            JOIN repo r ON d.repo_id=r.id 
			WHERE d.user_id=$1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var c Contribution
		var cUuid *uuid.UUID
		err = rows.Scan(
			&c.RepoName,
			&c.ContributorEmail,
			&c.ContributorWeight,
			&cUuid,
			&c.Balance,
			&c.BalanceRepo,
			&c.Day)
		if cUuid == nil {
			c.FlatFeeStackUser = false
		} else {
			c.FlatFeeStackUser = true
		}
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

//*********************************************************************************
//******************************* Repository **************************************
//*********************************************************************************
func insertOrUpdateRepo(repo *Repo) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) 
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
									ON CONFLICT(url) DO UPDATE SET name=$6, description=$7 RETURNING id`)
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

func findRepoByName(name string) (*Repo, error) {
	var r Repo
	var b []byte
	err := db.
		QueryRow("SELECT id, orig_id, url, git_url, branch, name, description, tags FROM repo WHERE name=$1", name).
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

func insertGitEmail(uid uuid.UUID, email string, token *string, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO git_email(user_id, email, token, created_at) VALUES($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO git_email for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid, email, token, now)
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
	// yes, 'EXCLUDED' is weird...
	// read stuff here before freaking out: https://www.postgresqltutorial.com/postgresql-upsert/
	stmt, err := db.Prepare(`INSERT INTO analysis_response(id, analysis_request_id, git_email, git_name, weight, created_at) 
									VALUES ($1, $2, $3, $4, $5, $6)
									ON CONFLICT(analysis_request_id, git_email) DO UPDATE SET weight=EXCLUDED.weight + analysis_response.weight`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v/%v statement event: %v", aid, w, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uuid.New(), aid, w.Email, w.Name, w.Weight, now)
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
                                            payment_cycle_id, 
                          	                user_id,
                                            from_user_id,
                                            balance, 
                                            balance_type, 
                          					currency,
                                            created_at) 
                                    VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_balances for %v/%v statement event: %v", ub.UserId, ub.PaymentCycleId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(ub.PaymentCycleId, ub.UserId, ub.FromUserId, ub.Balance, ub.BalanceType, strings.ToUpper(ub.Currency), ub.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertNewPaymentCycle(uid uuid.UUID, daysLeft int, seats int, freq int, createdAt time.Time) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO payment_cycle(user_id, days_left, seats, freq, created_at) 
                                    VALUES($1, $2, $3, $4, $5)  RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("prepareINSERT INTO payment_cycle for %v statement event: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(uid, daysLeft, seats, freq, createdAt).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil

	var res sql.Result
	res, err = stmt.Exec(uid, createdAt)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, handleErrMustInsertOne(res)
}

func insertNewInvoice(invoiceDb InvoiceDB) error {
	stmt, err := db.Prepare(`INSERT INTO invoice (nowpayments_invoice_id, 
					 payment_cycle_id, 
					 price_amount, 
					 price_currency, 
					 pay_currency, 
					 created_at,
                     last_update,
                     payment_status,
                     freq, invoice_url) 
					 values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`)

	if err != nil {
		return fmt.Errorf("prepareINSERT INTO payment_cycle for %v statement event: %v", 123, err) //error anpassen
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(invoiceDb.NowpaymentsInvoiceId,
		invoiceDb.PaymentCycleId,
		invoiceDb.PriceAmount,
		strings.ToUpper(invoiceDb.PriceCurrency), strings.ToUpper(invoiceDb.PayCurrency),
		invoiceDb.CreatedAt, invoiceDb.LastUpdate, strings.ToUpper(invoiceDb.PaymentStatus),
		invoiceDb.Freq, invoiceDb.InvoiceUrl.String).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	return nil

	var res sql.Result
	res, err = stmt.Exec("uid", " createdAt") //wieso braucht es hier noch das Exec??
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateInvoice(invoice InvoiceDB) error {
	stmt, err := db.Prepare(`UPDATE invoice SET 
                                           payment_cycle_id = $1, 
								payment_id = $2,            
								price_amount = $3,          
								price_currency = $4,        
								pay_amount = $5,            
								pay_currency = $6,          
								actually_paid = $7,         
								outcome_amount = $8,        
								outcome_currency = $9,      
								payment_status = $10,        
								created_at = $11,            
								last_update = $12,
                   				invoice_url = $13
                                    WHERE nowpayments_invoice_id=$14`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE users for %v statement failed: %v", "user", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(
		invoice.PaymentCycleId, invoice.PaymentId.Int64, invoice.PriceAmount, strings.ToUpper(invoice.PriceCurrency),
		invoice.PayAmount.Int64, strings.ToUpper(invoice.PayCurrency), invoice.ActuallyPaid.Int64, invoice.OutcomeAmount.Int64, strings.ToUpper(invoice.OutcomeCurrency.String), strings.ToUpper(invoice.PaymentStatus),
		invoice.CreatedAt, invoice.LastUpdate, invoice.InvoiceUrl.String, invoice.NowpaymentsInvoiceId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func getInvoice(id int64) (*InvoiceDB, error) {
	var invoice InvoiceDB
	err := db.
		QueryRow(`SELECT  nowpayments_invoice_id,
								payment_cycle_id, 
								payment_id,            
								price_amount,          
								price_currency,        
								pay_amount,            
								pay_currency,          
								actually_paid,         
								outcome_amount,        
								outcome_currency,      
								payment_status,
       							freq,
								created_at,            
								last_update                FROM invoice WHERE nowpayments_invoice_id=$1`, id).
		Scan(&invoice.NowpaymentsInvoiceId, &invoice.PaymentCycleId, &invoice.PaymentId, &invoice.PriceAmount, &invoice.PriceCurrency,
			&invoice.PayAmount, &invoice.PayCurrency, &invoice.ActuallyPaid, &invoice.OutcomeAmount, &invoice.OutcomeCurrency, &invoice.PaymentStatus,
			&invoice.Freq, &invoice.CreatedAt, &invoice.LastUpdate)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &invoice, nil
	default:
		return nil, err
	}

}

func findUserIdByInvoice(id int64) (uuid.UUID, error) {
	var userId uuid.UUID

	err := db.
		QueryRow(`SELECT payment_cycle.user_id FROM invoice 
    join payment_cycle on invoice.payment_cycle_id = payment_cycle.id where invoice.nowpayments_invoice_id = $1`, id).Scan(&userId)

	switch err {
	case sql.ErrNoRows:
		return uuid.UUID{}, nil
	case nil:
		return userId, nil
	default:
		return uuid.UUID{}, err
	}
}

type DailyPayment struct {
	PaymentCycleId uuid.UUID
	Currency       string
	Amount         int64
	DaysLeft       int
	LastUpdate     time.Time
}

func insertDailyPayment(dailyPayment DailyPayment) error {
	stmt, err := db.Prepare(`
					INSERT INTO daily_payment (payment_cycle_id, currency, amount, days_left, last_update) 
					values($1, $2, $3, $4, $5)`)

	if err != nil {
		return fmt.Errorf("prepare INSERT INTO daily_payment statement event: %v", err)
	}

	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(dailyPayment.PaymentCycleId, strings.ToUpper(dailyPayment.Currency), dailyPayment.Amount, dailyPayment.DaysLeft, dailyPayment.LastUpdate)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateDailyPayment(dailyPayment DailyPayment) error {
	stmt, err := db.Prepare(`UPDATE daily_payment SET amount = $1, days_left = $2, last_update = $3 WHERE payment_cycle_id=$4 AND currency = $5`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(dailyPayment.Amount, dailyPayment.DaysLeft, dailyPayment.LastUpdate, dailyPayment.PaymentCycleId, dailyPayment.Currency)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func findDaysLeftForCurrency(paymentCycleId uuid.UUID, currency string) (int, error) {
	var daysLeft int
	err := db.
		QueryRow(`select days_left from daily_payment where payment_cycle_id = $1 and currency = $2 order by last_update desc limit 1`, paymentCycleId, currency).
		Scan(&daysLeft)
	if err != nil {
		return 0, err
	}

	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return daysLeft, nil
	default:
		return 0, err
	}
}

func findDailyPaymentByPaymentCycleId(paymentCycleId uuid.UUID) ([]DailyPayment, error) {
	s := `select distinct payment_cycle_id, currency, amount, days_left, last_update from daily_payment where payment_cycle_id = $1`
	rows, err := db.Query(s, paymentCycleId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var dailyPayments []DailyPayment
	for rows.Next() {
		var dailyPayment DailyPayment
		err = rows.Scan(&dailyPayment.PaymentCycleId, &dailyPayment.Currency, &dailyPayment.Amount, &dailyPayment.DaysLeft, &dailyPayment.LastUpdate)
		if err != nil {
			return nil, err
		}
		dailyPayments = append(dailyPayments, dailyPayment)
	}
	return dailyPayments, nil
}

func updateFreq(paymentCycleId uuid.UUID, freq int) error {
	stmt, err := db.Prepare(`UPDATE payment_cycle SET freq = $1 WHERE id=$2`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle for %v statement event: %v", freq, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(freq, paymentCycleId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updatePaymentCycleDaysLeft(paymentCycleId uuid.UUID, daysLeft int64) error {
	stmt, err := db.Prepare(`UPDATE payment_cycle SET days_left = $1 WHERE id=$2`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle for %v statement event: %v", daysLeft, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(daysLeft, paymentCycleId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func findSumUserBalances(userId uuid.UUID, paymentCycleId uuid.UUID) ([]UserBalance, error) {
	s := `SELECT currency, COALESCE(sum(balance), 0)
          FROM user_balances 
          WHERE user_id = $1 AND payment_cycle_id = $2
          GROUP BY currency`

	rows, err := db.Query(s, userId, paymentCycleId)

	if err != nil {
		return nil, err
	}

	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		err = rows.Scan(&userBalance.Currency, &userBalance.Balance)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func findSumUserBalanceByCurrency(userId uuid.UUID, paymentCycleId uuid.UUID, currency string) (int64, error) {
	var sum int64
	var err error
	err = db.
		QueryRow(`SELECT COALESCE(sum(balance), 0)
                             FROM user_balances 
                             WHERE user_id = $1 AND payment_cycle_id = $2 AND currency = $3`, userId, paymentCycleId, currency).
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

func findUserBalances(userId uuid.UUID) ([]UserBalance, error) {
	s := `SELECT payment_cycle_id, user_id, balance, balance_type, created_at FROM user_balances WHERE user_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		err = rows.Scan(&userBalance.PaymentCycleId, &userBalance.UserId, &userBalance.Balance, &userBalance.BalanceType, &userBalance.CreatedAt)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func findUserBalancesAndType(paymentCycleId uuid.UUID, balanceType string, currency string) ([]UserBalance, error) {
	s := `SELECT payment_cycle_id, user_id, balance, balance_type, created_at FROM user_balances WHERE payment_cycle_id = $1 and balance_type = $2 and currency = $3`
	rows, err := db.Query(s, paymentCycleId, balanceType, currency)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		err = rows.Scan(&userBalance.PaymentCycleId, &userBalance.UserId, &userBalance.Balance, &userBalance.BalanceType, &userBalance.CreatedAt)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func findSponsoredUserBalances(userId uuid.UUID) ([]UserStatus, error) {
	s := `SELECT u.id, u.name, u.email, p.days_left
          FROM users u
          INNER JOIN payment_cycle p ON p.id = u.payment_cycle_id
          WHERE u.sponsor_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userStatus []UserStatus
	for rows.Next() {
		var userState UserStatus
		err = rows.Scan(&userState.UserId, &userState.Name, &userState.Email, &userState.DaysLeft)
		if err != nil {
			return nil, err
		}
		userStatus = append(userStatus, userState)
	}
	return userStatus, nil
}

func findAllCurreniesFromUserBalance(paymentCycleId uuid.UUID) ([]string, error) {
	s := `select distinct currency from user_balances where payment_cycle_id = $1`
	rows, err := db.Query(s, paymentCycleId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var currencies []string
	for rows.Next() {
		var currency string
		err = rows.Scan(&currency)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}
	return currencies, nil
}

func findPaymentCycle(pcid uuid.UUID) (*PaymentCycle, error) {
	var pc PaymentCycle
	err := db.
		QueryRow(`SELECT id, seats, freq FROM payment_cycle WHERE id=$1`, pcid).
		Scan(&pc.Id, &pc.Seats, &pc.Freq)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &pc, nil
	default:
		return nil, err
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

func insertPayoutRequest(p *PayoutRequestDB) error {
	stmt, err := db.Prepare(`
				INSERT INTO payout_request(user_id, batch_id, currency, exchange_rate, tea, address, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_request for %v statement event: %v", p.UserId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(p.UserId, p.BatchId, strings.ToUpper(p.Currency), p.ExchangeRate.String(), p.Tea, p.Address, p.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertPayoutResponse(p *PayoutResponseDB) error {
	pid := uuid.New()
	stmt, err := db.Prepare(`
				INSERT INTO payout_response(id, tx_hash, batch_id, error, created_at) 
				VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_response for %v statement event: %v", p.BatchId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(pid, p.TxHash, p.BatchId, p.Error, p.CreatedAt)
	if err != nil {
		return err
	}
	err = handleErrMustInsertOne(res)
	if err != nil {
		return err
	}
	for _, v := range p.Payouts.Payout {
		err = insertPayoutResponseDetails(pid, &v, p.Payouts.Currency)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertPayoutResponseDetails(pid uuid.UUID, payout *Payout, currency string) error {
	stmt, err := db.Prepare(`
				INSERT INTO payout_response_details(payout_response_id, currency, address, balance, created_at) 
				VALUES($1, $2, $3, $4, $5)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payout_response_details for %v statement event: %v", pid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(pid, strings.ToUpper(currency), payout.Address, payout.Balance, timeNow())
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func getPendingDailyUserPayouts(uid uuid.UUID, day time.Time) (*UserBalanceCore, error) {
	day = timeDay(-60, day) //day -2 month
	var ub UserBalanceCore
	err := db.
		QueryRow(`SELECT COALESCE(SUM(balance),0) as balance from daily_user_payout where user_id = $1 AND day >= $2`, uid, day).
		Scan(&ub.Balance)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		ub.UserId = uid
		return &ub, nil
	default:
		return nil, err
	}
}

func insertEmailSent(userId uuid.UUID, emailType string, now time.Time) error {
	stmt, err := db.Prepare(`
			INSERT INTO user_emails_sent(user_id, email_type, created_at) 
			VALUES($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_emails_sent for %v statement event: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(userId, emailType, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func countEmailSent(userId uuid.UUID, emailType string) (int, error) {
	var c int
	err := db.
		QueryRow(`SELECT count(*) AS c FROM user_emails_sent WHERE user_id=$1 and email_type=$2`, userId, emailType).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return c, nil
	default:
		return 0, err
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

func isUUIDZero(id uuid.UUID) bool {
	for x := 0; x < 16; x++ {
		if id[x] != 0 {
			return false
		}
	}
	return true
}
