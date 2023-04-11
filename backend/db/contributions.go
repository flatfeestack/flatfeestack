package db

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

type Contributions struct {
	DateFrom time.Time
	DateTo   time.Time
	GitEmail string
	GitNames []string
	Weight   float64
}

func FindContributions(contributorUserId uuid.UUID, myContribution bool) ([]Contribution, error) {
	cs := []Contribution{}
	s := `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.payment_cycle_in_id, d.day
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id WHERE d.user_sponsor_id = $1
            ORDER by d.day`
	if myContribution {
		s = `SELECT r.name, r.url, sp.name, sp.email, co.name, co.email, 
                 d.balance, d.currency, d.payment_cycle_in_id, d.day
            FROM daily_contribution d
                INNER JOIN users sp ON d.user_sponsor_id = sp.id
                INNER JOIN users co ON d.user_contributor_id = co.id
                INNER JOIN repo r ON d.repo_id = r.id WHERE d.user_contributor_id = $1
            ORDER by d.day`
	}

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

func FindRepoContribution(repoId uuid.UUID) ([]Contributions, error) {
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
		var jsonNames string
		err = rows.Scan(&c.DateFrom, &c.DateTo, &c.GitEmail, &jsonNames, &c.Weight)

		var names []string
		if err := json.Unmarshal([]byte(jsonNames), &names); err != nil {
			return nil, err
		}
		c.GitNames = names

		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, nil
}
