package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Contributions struct {
	DateFrom time.Time
	DateTo   time.Time
	GitEmail string
	GitNames []string
	Weight   float64
}

type Contribution struct {
	RepoName         string       `json:"repoName"`
	RepoUrl          string       `json:"repoUrl"`
	SponsorName      *string      `json:"sponsorName,omitempty"`
	SponsorEmail     string       `json:"sponsorEmail"`
	ContributorName  *string      `json:"contributorName,omitempty"`
	ContributorEmail string       `json:"contributorEmail"`
	Balance          *big.Int     `json:"balance"`
	Currency         string       `json:"currency"`
	Day              time.Time    `json:"day"`
	ClaimedAt        JsonNullTime `json:"claimedAt,omitempty"`
}

func InsertContribution(userSponsorId uuid.UUID, userContributorId uuid.UUID, repoId uuid.UUID, balance *big.Int, currency string, day time.Time, createdAt time.Time) error {

	stmt, err := DB.Prepare(`
				INSERT INTO daily_contribution(id, user_sponsor_id, user_contributor_id, repo_id, 
				                               balance, currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO daily_contribution for %v statement event: %v", userSponsorId, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	b := balance.String()
	id := uuid.New()
	res, err = stmt.Exec(id, userSponsorId, userContributorId, repoId, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindContributions(contributorUserId uuid.UUID, myContribution bool) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id WHERE d.user_sponsor_id = $1
            ORDER by d.day`
	if myContribution {
		s = `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id WHERE d.user_contributor_id = $1
            ORDER by d.day`
	}

	rows, err := DB.Query(s, contributorUserId)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var c Contribution
		var b string
		err = rows.Scan(
			&c.RepoName, &c.RepoUrl, &c.SponsorName, &c.SponsorEmail, &c.ContributorName,
			&c.ContributorEmail, &b, &c.Currency, &c.Day, &c.ClaimedAt)

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

func InsertFutureContribution(uid uuid.UUID, repoId uuid.UUID, balance *big.Int,
	currency string, day time.Time, createdAt time.Time) error {

	stmt, err := DB.Prepare(`
				INSERT INTO future_contribution(id, user_sponsor_id, repo_id, balance, currency, day, created_at) 
				VALUES($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO future_contribution for %v statement event: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	b := balance.String()
	id := uuid.New()
	res, err = stmt.Exec(id, uid, repoId, b, currency, day, createdAt)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindSumDailyContributors(userContributorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_contributor_id = $1
                     GROUP BY currency`, userContributorId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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

func FindSumDailySponsors(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_sponsor_id = $1
                     GROUP BY currency`, userSponsorId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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

func FindSumFutureSponsors(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM future_contribution 
                     WHERE user_sponsor_id = $1
                     GROUP BY currency`, userSponsorId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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

//TODO: integrate with loop above
/*for k, _ := range m1 {
	if m2[k] != nil {
		m1[k].Balance = new(big.Int).Add(m2[k].Balance, m1[k].Balance)
	}
}
for k, _ := range m2 {
	if m1[k] == nil {
		m1[k] = m2[k]
	}
}

return m1, nil*/

func FindSumFutureBalanceByRepoId(repoId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM future_contribution 
                             WHERE repo_id = $1
                             GROUP BY currency`, repoId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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

func FindSumDailyBalanceByRepoId(repoId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
                             FROM daily_contribution 
                             WHERE repo_id = $1
                             GROUP BY currency`, repoId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

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

func FindOwnContributionIds(contributorUserId uuid.UUID, currency string) ([]uuid.UUID, error) {
	contributionIds := []uuid.UUID{}

	s := `SELECT id FROM daily_contribution WHERE user_contributor_id = $1 AND currency = $2`
	rows, err := DB.Query(s, contributorUserId, currency)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)

		if err != nil {
			return nil, err
		}

		contributionIds = append(contributionIds, id)
	}
	return contributionIds, nil
}

func SumTotalEarnedAmountForContributionIds(contributionIds []uuid.UUID) (*big.Int, error) {
	var c string
	err := DB.
		QueryRow(`SELECT COALESCE(SUM(balance), 0) AS c FROM daily_contribution WHERE id = ANY($1)`, pq.Array(contributionIds)).
		Scan(&c)
	switch err {
	case sql.ErrNoRows:
		return big.NewInt(0), nil
	case nil:
		b1, ok := new(big.Int).SetString(c, 10)
		if !ok {
			return big.NewInt(0), fmt.Errorf("not a big.int %v", b1)
		}
		return b1, nil
	default:
		return big.NewInt(0), err
	}
}

func GetActiveSponsors(days int, isPostgres bool) ([]uuid.UUID, error) {
	var query string

	if isPostgres {
		query = `
            WITH trusted_repos AS (
                SELECT repo_id
                FROM trust_event
                WHERE un_trust_at IS NULL
            )
            SELECT DISTINCT dc.user_sponsor_id
            FROM daily_contribution dc
            JOIN trusted_repos tr ON dc.repo_id = tr.repo_id
            WHERE dc.created_at >= CURRENT_DATE - INTERVAL $1`
	} else {
		query = fmt.Sprintf(`
            WITH trusted_repos AS (
                SELECT repo_id
                FROM trust_event
                WHERE un_trust_at IS NULL
            )
            SELECT DISTINCT dc.user_sponsor_id
            FROM daily_contribution dc
            JOIN trusted_repos tr ON dc.repo_id = tr.repo_id
            WHERE dc.created_at >= date('now', '-%d days')`, days)
	}

	var rows *sql.Rows
	var err error
	if isPostgres {
		rows, err = DB.Query(query, fmt.Sprintf("%d days", days))
	} else {
		rows, err = DB.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var sponsors []uuid.UUID
	for rows.Next() {
		var sponsor uuid.UUID
		if err := rows.Scan(&sponsor); err != nil {
			return nil, err
		}
		sponsors = append(sponsors, sponsor)
	}

	return sponsors, nil
}

func IsActiveUser(userId uuid.UUID, isPostgres bool) (bool, error) {
	var count int
	var query string

	// Construct the query based on whether it's Postgres or SQLite
	if isPostgres {
		query = `
            SELECT COUNT(DISTINCT dc.user_sponsor_id) 
            FROM daily_contribution dc
            WHERE dc.user_sponsor_id = $1
            AND dc.created_at >= CURRENT_DATE - INTERVAL '1 month'`
	} else { // SQLite
		query = `
            SELECT COUNT(DISTINCT dc.user_sponsor_id) 
            FROM daily_contribution dc
            WHERE dc.user_sponsor_id = ? 
            AND dc.created_at >= date('now', '-1 month')`
	}

	// Execute the query
	err := DB.QueryRow(query, userId).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
