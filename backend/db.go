package main

//https://dataschool.com/how-to-teach-people-sql/sql-join-types-explained-visually/
import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"io"
	"math/big"
	"time"
)

type SponsorEvent struct {
	Id          uuid.UUID  `json:"id"`
	Uid         uuid.UUID  `json:"uid"`
	RepoId      uuid.UUID  `json:"repo_id"`
	EventType   uint8      `json:"event_type"`
	SponsorAt   time.Time  `json:"sponsor_at"`
	UnSponsorAt *time.Time `json:"un_sponsor_at"`
}

type Repos struct {
	Id       uuid.UUID           `json:"uuid"`
	Repo     []Repo              `json:"repos"`
	Balances map[string]*big.Int `json:"balances"`
}

type Repo struct {
	Id          uuid.UUID `json:"uuid"`
	Url         *string   `json:"url"`
	GitUrl      *string   `json:"gitUrl"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Score       uint32    `json:"score"`
	Source      *string   `json:"source"`
	Link        uuid.UUID `json:"link"`
	CreatedAt   time.Time
}

type PayoutRequest struct {
	UserId       uuid.UUID
	BatchId      uuid.UUID
	Currency     string
	ExchangeRate big.Float
	Tea          int64
	Address      string
	CreatedAt    time.Time
}

type PayoutsResponse struct {
	BatchId   uuid.UUID
	TxHash    string
	Error     *string
	CreatedAt time.Time
	Payouts   PayoutResponse
}

type GitEmail struct {
	Email       string     `json:"email"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type UserBalanceCore struct {
	UserId   uuid.UUID `json:"userId"`
	Balance  *big.Int  `json:"balance"`
	Currency string    `json:"currency"`
}

type UserBalance struct {
	UserId           uuid.UUID  `json:"userId"`
	Balance          *big.Int   `json:"balance"`
	Split            *big.Int   `json:"split"`
	PaymentCycleInId *uuid.UUID `json:"paymentCycleInId"`
	FromUserId       *uuid.UUID `json:"fromUserId"`
	BalanceType      string     `json:"balanceType"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type UserStatus struct {
	UserId   uuid.UUID `json:"userId"`
	Email    string    `json:"email"`
	Name     *string   `json:"name,omitempty"`
	DaysLeft int       `json:"daysLeft"`
}

type PaymentCycle struct {
	Id    uuid.UUID `json:"id"`
	Seats int64     `json:"seats"`
	Freq  int64     `json:"freq"`
}

type Contribution struct {
	RepoName         string    `json:"repoName"`
	RepoUrl          string    `json:"repoUrl"`
	SponsorName      *string   `json:"sponsorName,omitempty"`
	SponsorEmail     string    `json:"sponsorEmail"`
	ContributorName  *string   `json:"contributorName,omitempty"`
	ContributorEmail string    `json:"contributorEmail"`
	Balance          *big.Int  `json:"balance"`
	Currency         string    `json:"currency"`
	PaymentCycleInId uuid.UUID `json:"paymentCycleInId"`
	Day              time.Time `json:"day"`
}

type Wallet struct {
	Id       uuid.UUID `json:"id"`
	Currency string    `json:"currency"`
	Address  string    `json:"address"`
}

type PayoutInfo struct {
	Currency string   `json:"currency"`
	Amount   *big.Int `json:"amount"`
}

type Balance struct {
	Balance *big.Int
	Split   *big.Int
}

type AnalysisResponse struct {
	Id        uuid.UUID
	RequestId uuid.UUID `json:"request_id"`
	DateFrom  time.Time
	DateTo    time.Time
	GitEmail  string
	GitNames  []string
	Weight    float64
}

//*********************************************************************************
//********************************* Wallet ****************************************
//*********************************************************************************
func findActiveWalletsByUserId(uid uuid.UUID) ([]Wallet, error) {
	userWallets := []Wallet{}
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

func findAllWalletsByUserId(uid uuid.UUID) ([]Wallet, error) {
	userWallets := []Wallet{}
	s := "SELECT id, currency, address FROM wallet_address WHERE user_id=$1"
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

func insertWallet(uid uuid.UUID, currency string, address string, isDeleted bool) (*uuid.UUID, error) {
	stmt, err := db.Prepare("INSERT INTO wallet_address(user_id, currency, address, is_deleted) VALUES($1, $2, $3, $4) RETURNING id")
	if err != nil {
		return nil, fmt.Errorf("prepare INSERT INTO wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(uid, currency, address, isDeleted).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil
}

func updateWallet(uid uuid.UUID, isDeleted bool) error {
	stmt, err := db.Prepare("UPDATE wallet_address set is_deleted = $2 WHERE id=$1")
	if err != nil {
		return fmt.Errorf("prepare UPDATE wallet_address for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uid, isDeleted)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//*********************************************************************************
//******************************* Sponsoring **************************************
//*********************************************************************************
func insertOrUpdateSponsor(event *SponsorEvent) error {
	//first get last sponsored event to check if we need to insertOrUpdateSponsor or unsponsor
	//TODO: use mutex
	id, sponsorAt, unSponsorAt, err := findLastEventSponsoredRepo(event.Uid, event.RepoId)
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

func findLastEventSponsoredRepo(uid uuid.UUID, rid uuid.UUID) (*uuid.UUID, *time.Time, *time.Time, error) {
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

// Repositories and Sponsors

func findSponsoredReposById(userId uuid.UUID) (map[uuid.UUID]Repos, error) {
	//we want to send back an empty array, don't change
	s := `SELECT r.id, r.url, r.git_url, r.name, r.description, r.source, r.link
            FROM sponsor_event s
            INNER JOIN repo r ON s.repo_id=r.link
			WHERE s.user_id=$1 AND s.un_sponsor_at IS NULL`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)
	return scanRepo(rows)
}

func findContributions(contributorUserId uuid.UUID, myContribution bool) ([]Contribution, error) {
	cs := []Contribution{}
	subQuery := "d.user_sponsor_id"
	if myContribution {
		subQuery = "d.user_contributor_id"
	}
	s := `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.payment_cycle_in_id, d.day
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id 
			WHERE ` + subQuery + `=$1
            ORDER by d.day`
	rows, err := db.Query(s, contributorUserId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var c Contribution
		var b string
		err = rows.Scan(
			&c.RepoName,
			&c.RepoUrl,
			&c.SponsorName,
			&c.SponsorEmail,
			&c.ContributorName,
			&c.ContributorEmail,
			&b,
			&c.Currency,
			&c.PaymentCycleInId,
			&c.Day)

		if err != nil {
			return nil, err
		}

		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		c.Balance = b1
		cs = append(cs, c)
	}
	return cs, nil
}

//*********************************************************************************
//******************************* Repository **************************************
//*********************************************************************************
func insertOrUpdateRepo(repo *Repo) error {
	stmt, err := db.Prepare(`INSERT INTO repo (id, url, git_url, name, description, score, source, created_at, link) 
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
									ON CONFLICT(git_url) DO UPDATE SET git_url=$3 RETURNING id`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(repo.Id, repo.Url, repo.GitUrl, repo.Name, repo.Description, repo.Score, repo.Source, repo.CreatedAt, repo.Link).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	repo.Id = lastInsertId
	return nil
}

func insertOrUpdateRepoWithLink(repo *Repo) error {
	stmt, err := db.Prepare(`INSERT INTO repo (id, url, git_url, name, description, score, source, created_at, link) 
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
									ON CONFLICT(git_url) DO UPDATE SET link=$9, name=$4, description=$5`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(repo.Id, repo.Url, repo.GitUrl, repo.Name, repo.Description, repo.Score, repo.Source, repo.CreatedAt, repo.Link)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func updateRepoWhereLink(newLinkId uuid.UUID, oldLinkId *uuid.UUID) error {
	stmt, err := db.Prepare(`UPDATE repo SET link=$1 WHERE link=$2 OR id=$2`)
	if err != nil {
		return fmt.Errorf("prepare Update INTO repo for %v statement event: %v", newLinkId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(newLinkId, oldLinkId)
	if err != nil {
		return err
	}
	nr, err := handleErr(res)
	if err != nil {
		return err
	}
	log.Infof("affected %v rows", nr)
	return nil
}

func findRepoById(repoId uuid.UUID) (*Repo, error) {
	var r Repo
	err := db.
		QueryRow("SELECT id, url, git_url, name, description, source FROM repo WHERE id=$1", repoId).
		Scan(&r.Id, &r.Url, &r.GitUrl, &r.Name, &r.Description, &r.Source)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &r, nil
	default:
		return nil, err
	}
}

func findAllReposById(repoId uuid.UUID) (map[uuid.UUID]Repos, error) {
	rows, err := db.
		Query("SELECT id, url, git_url, name, description, source, link FROM repo WHERE id=$1 OR link=$1", repoId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)
	return scanRepo(rows)
}

func scanRepo(rows *sql.Rows) (map[uuid.UUID]Repos, error) {
	repoMap := map[uuid.UUID]Repos{}
	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.Id, &r.Url, &r.GitUrl, &r.Name, &r.Description, &r.Source, &r.Link)
		if err != nil {
			return nil, err
		}

		repos, ok := repoMap[r.Link]
		if !ok {
			repos.Id = r.Link
			repos.Repo = []Repo{r}
		} else {
			repos.Repo = append(repos.Repo, r)
		}
		repoMap[r.Link] = repos
	}
	return repoMap, nil
}

func findReposByName(name string) (map[uuid.UUID]Repos, error) {
	rows, err := db.Query("SELECT id, url, git_url, name, description, source, link FROM repo WHERE name=$1 ORDER BY link", name)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)
	return scanRepo(rows)
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
func insertAnalysisRequest(a AnalysisRequest, now time.Time) error {
	stmt, err := db.Prepare("INSERT INTO analysis_request(id, repo_id, date_from, date_to, git_urls, created_at) VALUES($1, $2, $3, $4, $5, $6)")
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_request for %v statement event: %v", a.RequestId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(a.RequestId, a.RepoId, a.DateFrom, a.DateTo, pq.Array(a.GitUrls), now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertAnalysisResponse(reqId uuid.UUID, gitEmail string, names []string, weight float64, now time.Time) error {
	stmt, err := db.Prepare(`INSERT INTO analysis_response(
                                     id, analysis_request_id, git_email, git_names, weight, created_at) 
									 VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO analysis_response for %v statement event: %v", reqId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(uuid.New(), reqId, gitEmail, pq.Array(names), weight, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

//https://stackoverflow.com/questions/3491329/group-by-with-maxdate
//https://pganalyze.com/docs/log-insights/app-errors/U115
func findLatestAnalysisRequest(repoId uuid.UUID) (*AnalysisRequest, error) {
	var a AnalysisRequest

	//https://stackoverflow.com/questions/47479973/golang-postgresql-array#47480256
	err := db.
		QueryRow(`SELECT id, repo_id, date_from, date_to, git_urls, received_at, error FROM (
                          SELECT id, repo_id, date_from, date_to, git_urls, received_at, error,
                            RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
                            FROM analysis_request WHERE repo_id=$1) AS x
                        WHERE dest_rank = 1`, repoId).
		Scan(&a.Id, &a.RepoId, &a.DateFrom, &a.DateTo, (*pq.StringArray)(&a.GitUrls), &a.ReceivedAt, &a.Error)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &a, nil
	default:
		return nil, err
	}
}

func findAllLatestAnalysisRequest(dateTo time.Time) ([]AnalysisRequest, error) {
	var as []AnalysisRequest

	rows, err := db.Query(`SELECT id, repo_id, date_from, date_to, git_urls, received_at, error FROM (
                          SELECT id, repo_id, date_from, date_to, git_urls, received_at, error,
                            RANK() OVER (PARTITION BY repo_id ORDER BY date_to DESC) dest_rank
                            FROM analysis_request) AS x
                        WHERE dest_rank = 1 AND date_to <= $1`, dateTo)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var a AnalysisRequest
		err = rows.Scan(&a.Id, &a.RepoId, &a.DateFrom, &a.DateTo, (*pq.StringArray)(&a.GitUrls), &a.ReceivedAt, &a.Error)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

type Contributions struct {
	DateFrom time.Time
	DateTo   time.Time
	GitEmail string
	GitNames []string
	Weight   float64
}

func findRepoContribution(repoId uuid.UUID) ([]Contributions, error) {
	var cs []Contributions

	rows, err := db.Query(`SELECT a.date_from, a.date_to, ar.git_email, ar.git_names, ar.weight
                        FROM analysis_request a
                        INNER JOIN analysis_response ar on a.id = ar.analysis_request_id
                        WHERE a.repo_id=$1 AND a.error IS NULL ORDER BY a.date_to, ar.weight DESC, ar.git_email`, repoId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var c Contributions
		err = rows.Scan(&c.DateFrom, &c.DateTo, &c.GitEmail, pq.Array(&c.GitNames), &c.Weight)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}

type Marketing struct {
	Email      string
	RepoIds    []uuid.UUID
	Balances   []string
	Currencies []string
}

func findMarketingEmails() ([]Marketing, error) {
	ms := []Marketing{}
	rows, err := db.Query(`SELECT u.email, array_agg(DISTINCT u.repo_id) as repo_ids, array_agg(u.balance) as balances,  array_agg(u.currency) as currencies
                        FROM unclaimed u
                        LEFT JOIN git_email g ON u.email = g.email 
                        WHERE g.email IS NULL 
                        GROUP BY u.email`)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var m Marketing
		err = rows.Scan(&m.Email, pq.Array(&m.RepoIds), pq.Array(&m.Balances), pq.Array(&m.Currencies))
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, nil
}

func updateAnalysisRequest(requestId uuid.UUID, now time.Time, errStr *string) error {
	stmt, err := db.Prepare(`UPDATE analysis_request set received_at = $2, error = $3 WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE analysis_request for statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(requestId, now, errStr)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
	return nil
}

func findAnalysisResults(reqId uuid.UUID) ([]AnalysisResponse, error) {
	var ars []AnalysisResponse

	rows, err := db.Query(`SELECT id, git_email, git_names, weight
                                 FROM analysis_response 
                                 WHERE analysis_request_id = $1`, reqId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var ar AnalysisResponse
		err = rows.Scan(&ar.Id, &ar.GitEmail, pq.Array(&ar.GitNames), &ar.Weight)
		if err != nil {
			return nil, err
		}
		ars = append(ars, ar)
	}
	return ars, nil
}

//*********************************************************************************
//******************************* Payments ****************************************
//*********************************************************************************
func insertUserBalance(ub UserBalance) error {
	stmt, err := db.Prepare(`INSERT INTO user_balances(
                                            payment_cycle_in_id, 
                          	                user_id,
                                            from_user_id,
                                            balance,
                          					split,
                                            balance_type, 
                          					currency,
                                            created_at) 
                                    VALUES($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_balances for %v/%v statement event: %v", ub.UserId, ub.PaymentCycleInId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := ub.Balance.String()
	s := ub.Split.String()
	res, err = stmt.Exec(ub.PaymentCycleInId, ub.UserId, ub.FromUserId, b, s, ub.BalanceType, ub.Currency, ub.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertNewPaymentCycleIn(seats int64, freq int64, createdAt time.Time) (*uuid.UUID, error) {
	stmt, err := db.Prepare(`INSERT INTO payment_cycle_in(seats, freq, created_at) 
                                    VALUES($1, $2, $3)  RETURNING id`)
	if err != nil {
		return nil, fmt.Errorf("prepareINSERT INTO payment_cycle_in statement event: %v", err)
	}
	defer closeAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(seats, freq, createdAt).Scan(&lastInsertId)
	if err != nil {
		return nil, err
	}
	return &lastInsertId, nil
}

func updateFreq(paymentCycleInId *uuid.UUID, freq int) error {
	stmt, err := db.Prepare(`UPDATE payment_cycle_in SET freq = $1 WHERE id=$2`)
	if err != nil {
		return fmt.Errorf("prepare UPDATE payment_cycle_in for %v statement event: %v", freq, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(freq, paymentCycleInId)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func findUserByGitEmail(gitEmail string) (*uuid.UUID, error) {
	var uid uuid.UUID
	err := db.
		QueryRow(`SELECT user_id FROM git_email WHERE email=$1`, gitEmail).
		Scan(&uid)
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &uid, nil
	default:
		return nil, err
	}
}

func insertFutureBalance(uid uuid.UUID, repoId uuid.UUID, paymentCycleInId *uuid.UUID, balance *big.Int,
	currency string, day time.Time, createdAt time.Time) error {

	stmt, err := db.Prepare(`
				INSERT INTO future_contribution(user_id, repo_id, payment_cycle_in_id, balance, 
				                               currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO future_contribution for %v statement event: %v", uid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := balance.String()
	res, err = stmt.Exec(uid, repoId, paymentCycleInId, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertUnclaimed(email string, rid uuid.UUID, balance *big.Int, currency string, day time.Time, createdAt time.Time) error {
	stmt, err := db.Prepare(`
				INSERT INTO unclaimed(email, repo_id, balance, currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6)
				ON CONFLICT(email, repo_id, currency) DO UPDATE SET balance=$3, day=$5, created_at=$6`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO unclaimed for %v statement event: %v", email, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := balance.String()
	res, err = stmt.Exec(email, rid, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertContribution(userSponsorId uuid.UUID, userContributorId uuid.UUID, repoId uuid.UUID, paymentCycleInId *uuid.UUID, payOutIdGit *uuid.UUID,
	balance *big.Int, currency string, day time.Time, createdAt time.Time) error {

	stmt, err := db.Prepare(`
				INSERT INTO daily_contribution(user_sponsor_id, user_contributor_id, repo_id, payment_cycle_in_id, payment_cycle_out_id, balance, 
				                               currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO daily_contribution for %v statement event: %v", userSponsorId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	b := balance.String()
	res, err = stmt.Exec(userSponsorId, userContributorId, repoId, paymentCycleInId, payOutIdGit, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)

	return nil
}

func findSumDailyContributors(userContributorId uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_contributor_id = $1
                     GROUP BY currency`, userContributorId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func findSumDailySponsors(userSponsorId uuid.UUID, paymentCycleInId uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_sponsor_id = $1 AND payment_cycle_in_id = $2
                     GROUP BY currency`, userSponsorId, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m1 := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m1[c] == nil {
			m1[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	rows, err = db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM future_contribution 
                     WHERE user_sponsor_id = $1 AND payment_cycle_in_id = $2
                     GROUP BY currency`, userSponsorId, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m2 := make(map[string]*Balance)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m2[c] == nil {
			m2[c] = &Balance{Balance: b1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	//TODO: integrate with loop above
	for k, _ := range m1 {
		if m2[k] != nil {
			m1[k].Balance = new(big.Int).Add(m2[k].Balance, m1[k].Balance)
		}
	}
	for k, _ := range m2 {
		if m1[k] == nil {
			m1[k] = m2[k]
		}
	}

	return m1, nil
}

func findSumDailyBalanceCurrency(paymentCycleInId *uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`	SELECT currency, COALESCE(sum(balance), 0)
        				FROM daily_contribution 
                        WHERE payment_cycle_in_id = $1
                        GROUP BY currency`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func findSumFutureBalanceByRepoId(repoId *uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM future_contribution 
                             WHERE repo_id = $1
                             GROUP BY currency`, repoId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func findSumFutureBalanceByCurrency(paymentCycleInId *uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM future_contribution 
                             WHERE payment_cycle_in_id = $1
                             GROUP BY currency`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*big.Int)
	for rows.Next() {
		var c, b string
		err = rows.Scan(&c, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = b1
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func findSumUserBalanceByCurrency(paymentCycleInId *uuid.UUID) (map[string]*Balance, error) {
	rows, err := db.
		Query(`SELECT currency, split, COALESCE(sum(balance), 0)
                             FROM user_balances 
                             WHERE payment_cycle_in_id = $1
                             GROUP BY currency, split`, paymentCycleInId)

	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	m := make(map[string]*Balance)
	for rows.Next() {
		var c, b, s string
		err = rows.Scan(&c, &s, &b)
		if err != nil {
			return nil, err
		}
		b1, ok := new(big.Int).SetString(b, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		s1, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("not a big.int %v", b1)
		}
		if m[c] == nil {
			m[c] = &Balance{Balance: b1, Split: s1}
		} else {
			return nil, fmt.Errorf("this is unexpected, we have duplicate! %v", c)
		}
	}

	return m, nil
}

func findUserBalances(userId uuid.UUID) ([]UserBalance, error) {
	s := `SELECT payment_cycle_in_id, user_id, balance, currency, balance_type, created_at FROM user_balances WHERE user_id = $1`
	rows, err := db.Query(s, userId)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		b := ""
		err = rows.Scan(&userBalance.PaymentCycleInId, &userBalance.UserId, &b, &userBalance.Currency, &userBalance.BalanceType, &userBalance.CreatedAt)
		userBalance.Balance, _ = new(big.Int).SetString(b, 10)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func findUserBalancesAndType(paymentCycleInId *uuid.UUID, balanceType string, currency string) ([]UserBalance, error) {
	s := `SELECT payment_cycle_in_id, user_id, balance, balance_type, created_at FROM user_balances WHERE payment_cycle_id = $1 and balance_type = $2 and currency = $3`
	rows, err := db.Query(s, paymentCycleInId, balanceType, currency)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	var userBalances []UserBalance
	for rows.Next() {
		var userBalance UserBalance
		err = rows.Scan(&userBalance.PaymentCycleInId, &userBalance.UserId, &userBalance.Balance, &userBalance.BalanceType, &userBalance.CreatedAt)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, userBalance)
	}
	return userBalances, nil
}

func findSponsoredUserBalances(userId uuid.UUID) ([]UserStatus, error) {
	s := `SELECT u.id, u.name, u.email
          FROM users u
          INNER JOIN payment_cycle_in p ON p.id = u.payment_cycle_in_id
          WHERE u.invited_id = $1`
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

func findPaymentCycle(paymentCycleInId *uuid.UUID) (*PaymentCycle, error) {
	var pc PaymentCycle
	err := db.
		QueryRow(`SELECT id, seats, freq FROM payment_cycle_in WHERE id=$1`, paymentCycleInId).
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

func findPaymentCycleLast(uid uuid.UUID) (PaymentCycle, error) {
	pc := PaymentCycle{}
	err := db.
		QueryRow(`SELECT p.id, p.seats, p.freq 
                        FROM payment_cycle_in p JOIN users u on p.id = u.payment_cycle_in_id
                        WHERE u.id=$1`, uid).
		Scan(&pc.Id, &pc.Seats, &pc.Freq)
	switch err {
	case sql.ErrNoRows:
		return pc, nil
	case nil:
		return pc, nil
	default:
		return pc, err
	}
}

//*********************************************************************************
//******************************* Payouts *****************************************
//*********************************************************************************
/*func insertPayoutsRequest(p *PayoutsRequest) error {
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
}*/

func insertPayoutRequest(p *PayoutRequest) error {
	stmt, err := db.Prepare(`
				INSERT INTO payout_request(user_id, batch_id, currency, exchange_rate, tea, address, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payouts_request for %v statement event: %v", p.UserId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(p.UserId, p.BatchId, p.Currency, p.ExchangeRate.String(), p.Tea, p.Address, p.CreatedAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func insertPayoutResponse(p *PayoutsResponse) error {
	pid := uuid.New()
	stmt, err := db.Prepare(`
				INSERT INTO payout_response(id, batch_id, tx_hash, error, created_at) 
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
	for _, v := range p.Payouts.Payout {
		if len(v.Meta) > 0 {
			for _, i := range v.Meta {
				v.NanoTea = i.Tea
				err = insertPayoutResponseDetails(pid, &v, i.Currency)
				if err != nil {
					return err
				}
			}
		} else {
			err = insertPayoutResponseDetails(pid, &v, p.Payouts.Currency)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func insertPayoutResponseDetails(pid uuid.UUID, payout *Payout, currency string) error {
	stmt, err := db.Prepare(`
				INSERT INTO payout_response_details(payout_response_id, currency, address, nano_tea, smart_contract_tea, created_at) 
				VALUES($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO payout_response_details for %v statement event: %v", pid, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(pid, currency, payout.Address, payout.NanoTea, payout.SmartContractTea.String(), timeNow())
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func getPendingDailyUserPayouts(uid uuid.UUID) ([]UserBalanceCore, error) {
	var ubs []UserBalanceCore
	s := `SELECT 
		 		payout.currency, 
		 		MAX(payout.balance) - COALESCE(SUM(payout_response.balance), 0) as balance
		 	FROM (	SELECT dup.user_id, SUM(dup.balance) as balance, dup.currency 
		 			FROM daily_user_payout dup 
		 			GROUP BY dup.user_id, dup.currency
		 		 ) as payout
		 	LEFT JOIN (	SELECT req.user_id, res_details.address, res_details.currency, MAX(res_details.nano_tea) as balance FROM payout_response_details res_details
		 				JOIN payout_response res on res_details.payout_response_id = res.id
		 				JOIN payout_request req on req.batch_id = res.batch_id and req.address = res_details.address
		 				GROUP BY req.user_id, res_details.address, res_details.currency 
		 			  ) as payout_response 
		 				on payout_response.user_id = payout.user_id 
		 				AND payout_response.currency = payout.currency 
			WHERE payout.user_id = $1
		 	GROUP BY payout.user_id, payout.currency`

	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var ub UserBalanceCore
		err = rows.Scan(&ub.Currency, &ub.Balance)
		if err != nil {
			return nil, err
		}
		ubs = append(ubs, ub)
	}
	return ubs, nil
}

func getTotalRealizedIncome(uid uuid.UUID) ([]UserBalanceCore, error) {
	var ubs []UserBalanceCore
	s := `	SELECT tmp.currency, SUM(tmp.balance) 
			FROM (	SELECT req.user_id, res_details.address, res_details.currency, MAX(res_details.nano_tea) as balance FROM payout_response_details res_details
					JOIN payout_response res on res_details.payout_response_id = res.id
					JOIN payout_request req on req.batch_id = res.batch_id and req.address = res_details.address
					where req.user_id = $1
					GROUP BY req.user_id, res_details.address, res_details.currency ) as tmp
			GROUP BY tmp.currency`

	rows, err := db.Query(s, uid)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var ub UserBalanceCore
		err = rows.Scan(&ub.Currency, &ub.Balance)
		if err != nil {
			return nil, err
		}
		ubs = append(ubs, ub)
	}
	return ubs, nil
}

func findPayoutInfos() ([]PayoutInfo, error) {
	var payoutInfos []PayoutInfo
	s := `	SELECT 
				dup.currency,
				CASE WHEN MAX(tmp2.balance) IS NULL THEN 
					SUM(dup.balance)
				ELSE
					SUM(dup.balance) - MAX(tmp2.balance)
				END AS balance
			FROM daily_user_payout dup 
			JOIN wallet_address wa ON wa.user_id = dup.user_id AND ((wa.currency = dup.currency) OR (dup.currency = 'USD' AND wa.currency = 'ETH'))
			LEFT JOIN ( SELECT tmp.currency , SUM(tmp.balance) AS balance
						FROM (	SELECT currency, MAX(nano_tea) AS balance 
								FROM payout_response_details GROUP BY address, currency
							 ) AS tmp GROUP BY tmp.currency
					  ) AS tmp2 ON tmp2.currency = dup.currency
			WHERE wa.is_deleted = false
			GROUP BY dup.currency`
	rows, err := db.Query(s)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(rows)

	for rows.Next() {
		var payoutInfo PayoutInfo
		err = rows.Scan(&payoutInfo.Currency, &payoutInfo.Amount)
		if err != nil {
			return nil, err
		}
		payoutInfos = append(payoutInfos, payoutInfo)
	}
	return payoutInfos, nil
}

func insertEmailSent(userId *uuid.UUID, email string, emailType string, now time.Time) error {
	stmt, err := db.Prepare(`
			INSERT INTO user_emails_sent(user_id, email, email_type, created_at) 
			VALUES($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO user_emails_sent for %v statement event: %v", userId, err)
	}
	defer closeAndLog(stmt)

	var res sql.Result
	res, err = stmt.Exec(userId, email, emailType, now)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func countEmailSentById(userId uuid.UUID, emailType string) (int, error) {
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

func countEmailSentByEmail(email string, emailType string) (int, error) {
	var c int
	err := db.
		QueryRow(`SELECT count(*) AS c FROM user_emails_sent WHERE email=$1 and email_type=$2`, email, emailType).
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
	now := timeNow()
	for err != nil && now.Add(time.Duration(10)*time.Second).After(timeNow()) {
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
