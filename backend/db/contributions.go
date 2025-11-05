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

type ContributionDetail struct {
	Balance   *big.Int  `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserDonationRepo struct {
	TrustedRepoSelected   []uuid.UUID
	UntrustedRepoSelected []uuid.UUID
	Currency              string
	SponsorAmount         big.Int
}

func (db *DB) InsertContribution(userSponsorId uuid.UUID, userContributorId uuid.UUID, repoId uuid.UUID, balance *big.Int, currency string, day time.Time, createdAt time.Time, foundationPayment bool) error {
	_, err := db.Exec(
		`INSERT INTO daily_contribution(id, user_sponsor_id, user_contributor_id, repo_id, 
		                                balance, currency, day, created_at, foundation_payment) 
		 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.New(), userSponsorId, userContributorId, repoId, balance.String(), currency, day, createdAt, foundationPayment)
	return err
}

func (db *DB) FindContributions(contributorUserId uuid.UUID, myContribution bool) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at, d.foundation_payment
          FROM daily_contribution d
              INNER JOIN users sp ON d.user_sponsor_id = sp.id
              INNER JOIN users co ON d.user_contributor_id = co.id
              INNER JOIN repo r ON d.repo_id = r.id 
          WHERE d.user_sponsor_id = $1
          ORDER BY d.day`
	
	if myContribution {
		s = `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.day, d.claimed_at, d.foundation_payment
             FROM daily_contribution d
                 INNER JOIN users sp ON d.user_sponsor_id = sp.id
                 INNER JOIN users co ON d.user_contributor_id = co.id
                 INNER JOIN repo r ON d.repo_id = r.id 
             WHERE d.user_contributor_id = $1
             ORDER BY d.day`
	}

	rows, err := db.Query(s, contributorUserId)
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

func (db *DB) InsertFutureContribution(uid uuid.UUID, repoId uuid.UUID, balance *big.Int,
	currency string, day time.Time, createdAt time.Time, foundationPayment bool) error {
	_, err := db.Exec(
		`INSERT INTO future_contribution(id, user_sponsor_id, repo_id, balance, currency, day, created_at, foundation_payment) 
		 VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), uid, repoId, balance.String(), currency, day, createdAt, foundationPayment)
	return err
}

func (db *DB) InsertOrUpdateFutureContribution(uid uuid.UUID, repoId uuid.UUID, balance *big.Int,
	currency string, day time.Time, createdAt time.Time, foundationPayment bool) error {
	_, err := db.Exec(
		`INSERT INTO future_contribution(
		     id, user_sponsor_id, repo_id, balance, currency, day, created_at, foundation_payment
		 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_sponsor_id, repo_id, currency, day) 
		 DO UPDATE SET 
		     balance = future_contribution.balance + EXCLUDED.balance,
		     created_at = EXCLUDED.created_at,
		     foundation_payment = EXCLUDED.foundation_payment`,
		uuid.New(), uid, repoId, balance.String(), currency, day, createdAt, foundationPayment)
	return err
}

func (db *DB) FindSumDailyContributors(userContributorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM daily_contribution 
         WHERE user_contributor_id = $1
         GROUP BY currency`, 
		userContributorId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindSumDailySponsors(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM daily_contribution 
         WHERE user_sponsor_id = $1 AND foundation_payment = FALSE
         GROUP BY currency`, 
		userSponsorId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindSumDailySponsorsFromFoundation(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM daily_contribution 
         WHERE user_sponsor_id = $1 AND foundation_payment = TRUE
         GROUP BY currency`, 
		userSponsorId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindSumDailySponsorsFromFoundationByCurrency(userId uuid.UUID, currency string) (*big.Int, error) {
	var balanceStr string
	err := db.QueryRow(
		`SELECT COALESCE(sum(balance), 0)
         FROM daily_contribution 
         WHERE user_sponsor_id = $1 AND foundation_payment = TRUE AND currency = $2`,
		userId, currency).Scan(&balanceStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid balance: %v", balanceStr)
	}

	return balance, nil
}

func (db *DB) FindContributionsGroupedByCurrencyAndRepo(userSponsorId uuid.UUID) (map[string]map[uuid.UUID]ContributionDetail, error) {
	rows, err := db.Query(`
        SELECT currency, repo_id, COALESCE(SUM(balance), 0), MIN(created_at) AS latest_created_at
        FROM (
            SELECT currency, repo_id, balance, created_at
            FROM daily_contribution
            WHERE user_sponsor_id = $1 AND foundation_payment = FALSE

            UNION ALL

            SELECT currency, repo_id, balance, created_at
            FROM future_contribution
            WHERE user_sponsor_id = $1 AND foundation_payment = FALSE
        ) combined_contributions
        GROUP BY currency, repo_id
		ORDER BY latest_created_at ASC
    `, userSponsorId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	contributions := make(map[string]map[uuid.UUID]ContributionDetail)
	for rows.Next() {
		var currency string
		var repoID uuid.UUID
		var balanceStr string
		var createdAt time.Time

		err = rows.Scan(&currency, &repoID, &balanceStr, &createdAt)
		if err != nil {
			return nil, err
		}

		balance, ok := new(big.Int).SetString(balanceStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid balance format: %v", balanceStr)
		}

		if contributions[currency] == nil {
			contributions[currency] = make(map[uuid.UUID]ContributionDetail)
		}

		if _, exists := contributions[currency][repoID]; exists {
			return nil, fmt.Errorf("unexpected duplicate entry for currency: %v, repo_id: %v", currency, repoID)
		}

		contributions[currency][repoID] = ContributionDetail{
			Balance:   balance,
			CreatedAt: createdAt,
		}
	}

	return contributions, nil
}

func (db *DB) FindSumFutureSponsors(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM future_contribution 
         WHERE user_sponsor_id = $1 AND foundation_payment = FALSE
         GROUP BY currency`, 
		userSponsorId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindSumFutureSponsorsFromFoundation(userSponsorId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM future_contribution 
         WHERE user_sponsor_id = $1 AND foundation_payment = TRUE
         GROUP BY currency`, 
		userSponsorId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindFoundationContributionsGroupedByCurrencyAndRepo(userSponsorId uuid.UUID) (map[string]map[uuid.UUID]ContributionDetail, error) {
	rows, err := db.Query(`
        SELECT currency, repo_id, COALESCE(SUM(balance), 0), MIN(created_at) AS latest_created_at
        FROM (
            SELECT currency, repo_id, balance, created_at
            FROM daily_contribution
            WHERE user_sponsor_id = $1 AND foundation_payment = TRUE

            UNION ALL

            SELECT currency, repo_id, balance, created_at
            FROM future_contribution
            WHERE user_sponsor_id = $1 AND foundation_payment = TRUE
        ) combined_contributions
        GROUP BY currency, repo_id
		ORDER BY latest_created_at ASC
    `, userSponsorId)

	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)

	contributions := make(map[string]map[uuid.UUID]ContributionDetail)
	for rows.Next() {
		var currency string
		var repoID uuid.UUID
		var balanceStr string
		var createdAt time.Time

		err = rows.Scan(&currency, &repoID, &balanceStr, &createdAt)
		if err != nil {
			return nil, err
		}

		balance, ok := new(big.Int).SetString(balanceStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid balance format: %v", balanceStr)
		}

		if contributions[currency] == nil {
			contributions[currency] = make(map[uuid.UUID]ContributionDetail)
		}

		if _, exists := contributions[currency][repoID]; exists {
			return nil, fmt.Errorf("unexpected duplicate entry for currency: %v, repo_id: %v", currency, repoID)
		}

		contributions[currency][repoID] = ContributionDetail{
			Balance:   balance,
			CreatedAt: createdAt,
		}
	}

	return contributions, nil
}

func (db *DB) FindSumFutureBalanceByRepoId(repoId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM future_contribution 
         WHERE repo_id = $1
         GROUP BY currency`, 
		repoId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindSumDailyBalanceByRepoId(repoId uuid.UUID) (map[string]*big.Int, error) {
	rows, err := db.Query(
		`SELECT currency, COALESCE(sum(balance), 0)
         FROM daily_contribution 
         WHERE repo_id = $1
         GROUP BY currency`, 
		repoId)

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
			return nil, fmt.Errorf("unexpected duplicate currency: %v", c)
		}
	}

	return m, nil
}

func (db *DB) FindOwnContributionIds(contributorUserId uuid.UUID, currency string) ([]uuid.UUID, error) {
	contributionIds := []uuid.UUID{}

	rows, err := db.Query(
		`SELECT id FROM daily_contribution WHERE user_contributor_id = $1 AND currency = $2`,
		contributorUserId, currency)
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

func (db *DB) SumTotalEarnedAmountForContributionIds(contributionIds []uuid.UUID) (*big.Int, error) {
	var c string
	err := db.QueryRow(
		`SELECT COALESCE(SUM(balance), 0) AS c FROM daily_contribution WHERE id = ANY($1)`,
		pq.Array(contributionIds)).Scan(&c)
	
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

func (db *DB) GetUserDonationRepos(userId uuid.UUID, yesterdayStart time.Time, futureContribution bool) (map[uuid.UUID][]UserDonationRepo, error) {
	var s string
	if futureContribution {
		s = `
			SELECT
				dc.user_sponsor_id,
				dc.repo_id,
				dc.balance,
				dc.currency,
				COALESCE(te.trust_at IS NOT NULL AND te.un_trust_at IS NULL, FALSE) AS is_trusted
			FROM future_contribution dc
			LEFT JOIN trust_event te ON te.repo_id = dc.repo_id
			WHERE dc.user_sponsor_id = $1
			  AND dc.day = $2
			  AND dc.repo_id IS NOT NULL
		`
	} else {
		s = `
			SELECT
				dc.user_sponsor_id,
				dc.repo_id,
				dc.balance,
				dc.currency,
				COALESCE(te.trust_at IS NOT NULL AND te.un_trust_at IS NULL, FALSE) AS is_trusted
			FROM daily_contribution dc
			LEFT JOIN trust_event te ON te.repo_id = dc.repo_id
			WHERE dc.user_sponsor_id = $1
			  AND dc.day = $2
			  AND dc.repo_id IS NOT NULL
		`
	}

	rows, err := db.Query(s, userId, yesterdayStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := createUserDonationRepo(rows)
	if err != nil {
		return nil, fmt.Errorf("createUserDonationRepo failed: %w", err)
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

func (db *DB) GetActiveSponsors(months int) ([]uuid.UUID, error) {
	query := fmt.Sprintf(`
        WITH trusted_repos AS (
            SELECT repo_id
            FROM trust_event
            WHERE un_trust_at IS NULL
        )
        SELECT DISTINCT dc.user_sponsor_id
        FROM daily_contribution dc
        JOIN trusted_repos tr ON dc.repo_id = tr.repo_id
        WHERE dc.created_at >= CURRENT_DATE - INTERVAL '%d month'`, months)

	rows, err := db.Query(query)
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

func (db *DB) FilterActiveUsers(userIds []uuid.UUID, months int) ([]uuid.UUID, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT user_sponsor_id 
		FROM daily_contribution 
		WHERE user_sponsor_id = ANY($1)
		AND created_at >= CURRENT_DATE - INTERVAL '%d month'`, months)

	rows, err := db.Query(query, pq.Array(userIds))
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