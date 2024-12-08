package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var TABLENAMEFUTURECONTRIBUTION = map[bool]string{
	true:  "future_contribution",
	false: "daily_contribution",
}

type Contributions struct {
	DateFrom time.Time
	DateTo   time.Time
	GitEmail string
	GitNames []string
	Weight   float64
}

type Contribution struct {
	RepoName          string       `json:"repoName"`
	RepoUrl           string       `json:"repoUrl"`
	SponsorName       *string      `json:"sponsorName,omitempty"`
	SponsorEmail      string       `json:"sponsorEmail"`
	ContributorName   *string      `json:"contributorName,omitempty"`
	ContributorEmail  string       `json:"contributorEmail"`
	Balance           *big.Int     `json:"balance"`
	Currency          string       `json:"currency"`
	Day               time.Time    `json:"day"`
	ClaimedAt         JsonNullTime `json:"claimedAt,omitempty"`
	FoundationPayment bool         `json:"foundationPayment"`
}

func InsertContribution(userSponsorId uuid.UUID, userContributorId uuid.UUID, repoId uuid.UUID, balance *big.Int, currency string, day time.Time, createdAt time.Time, foudnationPayment bool) error {

	stmt, err := DB.Prepare(`
				INSERT INTO daily_contribution(id, user_sponsor_id, user_contributor_id, repo_id, 
				                               balance, currency, day, created_at, foundation_payment) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO daily_contribution for %v statement event: %v", userSponsorId, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	b := balance.String()
	id := uuid.New()
	res, err = stmt.Exec(id, userSponsorId, userContributorId, repoId, b, currency, day, createdAt, foudnationPayment)
	if err != nil {
		return err
	}
	return handleErrMustInsertOne(res)
}

func FindContributions(contributorUserId uuid.UUID, myContribution bool) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at, d.foundation_payment
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id WHERE d.user_sponsor_id = $1
            ORDER by d.day`
	if myContribution {
		s = `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at, d.foundation_payment
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
			&c.ContributorEmail, &b, &c.Currency, &c.Day, &c.ClaimedAt, &c.FoundationPayment)

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
	currency string, day time.Time, createdAt time.Time, foudnationPayment bool) error {

	stmt, err := DB.Prepare(`
				INSERT INTO future_contribution(id, user_sponsor_id, repo_id, balance, currency, day, created_at, foundation_payment) 
				VALUES($1, $2, $3, $4, $5, $6, $7, $8)`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO future_contribution for %v statement event: %v", uid, err)
	}
	defer CloseAndLog(stmt)

	var res sql.Result
	b := balance.String()
	id := uuid.New()
	res, err = stmt.Exec(id, uid, repoId, b, currency, day, createdAt, foudnationPayment)
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

func FindSumDailySponsorsFromFoundation(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM daily_contribution 
                     WHERE user_sponsor_id = $1
					 	AND foundation_payment
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

func FindSumDailySponsorsFromFoundationByCurrency(userId uuid.UUID, currency string) (*big.Int, error) {
	query := `
			SELECT COALESCE(sum(balance), 0)
               FROM daily_contribution 
               WHERE user_sponsor_id = $1
			   	AND foundation_payment
				AND currency = $3`

	var balanceSumInt int64

	err := DB.QueryRow(query, userId, currency).Scan(&balanceSumInt)

	if err != nil {
		if err == sql.ErrNoRows {
			return big.NewInt(0), nil
		}
		return nil, fmt.Errorf("this is an unexpected error: %v", err)
	}

	return big.NewInt(balanceSumInt), nil
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

func FindSumFutureSponsorsFromFoundation(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := DB.
		Query(`SELECT currency, COALESCE(sum(balance), 0)
        			 FROM future_contribution 
                     WHERE user_sponsor_id = $1
					 	AND foundation_payment
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

/*type FoundationSponsoring struct {
	FoundationID           string
	FoundationBalanceRepos []FoundationBalanceRepos
}

type FoundationBalanceRepos struct {
	Currency      uuid.UUID
	SponsorAmount big.Int
	RepoIds       []uuid.UUID
}*/

type UserDonationRepo struct {
	TrustedRepoSelected   []uuid.UUID
	UntrustedRepoSelected []uuid.UUID
	Currency              string
	SponsorAmount         big.Int
}

func GetUserDonationRepos(userId uuid.UUID, yesterdayStart time.Time, futureContribution bool) (map[uuid.UUID][]UserDonationRepo, error) {
	s := fmt.Sprintf(`
        SELECT
            dc.user_sponsor_id,
            dc.repo_id,
            dc.balance,
            dc.currency,
            COALESCE(te.trust_at IS NOT NULL AND te.un_trust_at IS NULL, FALSE) AS is_trusted
        FROM %s dc
        LEFT JOIN trust_event te ON te.repo_id = dc.repo_id
        WHERE dc.user_sponsor_id = $1
          AND dc.day = $2
          AND dc.repo_id IS NOT NULL
    `, TABLENAMEFUTURECONTRIBUTION[futureContribution])

	rows, err := DB.Query(s, userId, yesterdayStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := createUserDonationRepo(rows)
	if err != nil {
		return nil, fmt.Errorf("something went wrong %v", err)
	}

	return result, nil
}

func createUserDonationRepo(rows *sql.Rows) (map[uuid.UUID][]UserDonationRepo, error) {
	result := make(map[uuid.UUID][]UserDonationRepo)

	for rows.Next() {
		var userSponsorID uuid.UUID
		var repoID uuid.UUID
		var balance string
		var currency string
		var isTrusted bool

		err := rows.Scan(&userSponsorID, &repoID, &balance, &currency, &isTrusted)
		if err != nil {
			return nil, err
		}

		if repoID == uuid.Nil {
			continue
		}

		userDonationRepos, found := result[userSponsorID]
		if !found {
			userDonationRepos = []UserDonationRepo{}
		}

		sponsorAmount := new(big.Int)
		if balance != "" {
			sponsorAmount, _ = sponsorAmount.SetString(balance, 10)
		} else {
			sponsorAmount = big.NewInt(0)
		}

		var targetRepo *UserDonationRepo
		for i := range userDonationRepos {
			if userDonationRepos[i].Currency == currency {
				targetRepo = &userDonationRepos[i]
				break
			}
		}

		if targetRepo == nil {
			newRepo := UserDonationRepo{
				Currency:              currency,
				SponsorAmount:         *big.NewInt(0),
				TrustedRepoSelected:   []uuid.UUID{},
				UntrustedRepoSelected: []uuid.UUID{},
			}
			userDonationRepos = append(userDonationRepos, newRepo)
			targetRepo = &userDonationRepos[len(userDonationRepos)-1]
		}

		targetRepo.SponsorAmount.Add(&targetRepo.SponsorAmount, sponsorAmount)

		if isTrusted {
			if !containsUUID(targetRepo.TrustedRepoSelected, repoID) {
				targetRepo.TrustedRepoSelected = append(targetRepo.TrustedRepoSelected, repoID)
			}
		} else {
			if !containsUUID(targetRepo.UntrustedRepoSelected, repoID) {
				targetRepo.UntrustedRepoSelected = append(targetRepo.UntrustedRepoSelected, repoID)
			}
		}

		result[userSponsorID] = userDonationRepos
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func containsUUID(slice []uuid.UUID, id uuid.UUID) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}

/*func GetFoundationsFromDailyContributions(yesterdayStart time.Time) ([]FoundationSponsoring, error) {

	s := `SELECT
			me.user_id AS foundation_id,
			dc.currency,
			COALESCE(sum(dc.balance), 0) AS total_balance,
			me.repo_id
		  FROM daily_contribution dc
		  INNER JOIN multiplier_event me ON dc.repo_id = me.repo_id
	      WHERE me.un_multiplier_at IS NULL
			AND dc.day = $1
		  GROUP BY dc.currency, me.user_id`
	rows, err := DB.Query(s, yesterdayStart)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

}*/

func GetActiveSponsors(months int, isPostgres bool) ([]uuid.UUID, error) {
	var query string

	if isPostgres {
		query = fmt.Sprintf(`
            WITH trusted_repos AS (
                SELECT repo_id
                FROM trust_event
                WHERE un_trust_at IS NULL
            )
            SELECT DISTINCT dc.user_sponsor_id
            FROM daily_contribution dc
            JOIN trusted_repos tr ON dc.repo_id = tr.repo_id
            WHERE dc.created_at >= CURRENT_DATE - INTERVAL '%d month'`, months)
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
            WHERE dc.created_at >= date('now', '-%d month')`, months)
	}

	rows, err := DB.Query(query)

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

func FilterActiveUsers(userIds []uuid.UUID, months int, isPostgres bool) ([]uuid.UUID, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	var query string
	if isPostgres {
		query = fmt.Sprintf(`
			SELECT DISTINCT user_sponsor_id 
			FROM daily_contribution 
			WHERE user_sponsor_id IN (`+GeneratePlaceholders(len(userIds))+`) 
			AND created_at >= CURRENT_DATE - INTERVAL '%d month'`, months)
	} else {
		query = fmt.Sprintf(`
			SELECT DISTINCT user_sponsor_id 
			FROM daily_contribution 
			WHERE user_sponsor_id IN (`+GeneratePlaceholders(len(userIds))+`) 
			AND created_at >= date('now', '-%d month')`, months)
	}

	rows, err := DB.Query(query, ConvertToInterfaceSlice(userIds)...)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	var activeUsers []uuid.UUID
	for rows.Next() {
		var userId uuid.UUID
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		activeUsers = append(activeUsers, userId)
	}
	return activeUsers, nil
}
