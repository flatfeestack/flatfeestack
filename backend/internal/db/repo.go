package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Repo struct {
	Id          uuid.UUID `json:"uuid"`
	Url         *string   `json:"url"`
	GitUrl      *string   `json:"gitUrl"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Score       uint32    `json:"score"`
	Source      *string   `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
}

func InsertOrUpdateRepo(repo *Repo) error {
	stmt, err := DB.Prepare(`INSERT INTO repo (id, url, git_url, name, description, score, source, created_at) 
								   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
								   ON CONFLICT(git_url) DO UPDATE SET git_url=$3 RETURNING id`)
	if err != nil {
		return fmt.Errorf("prepare INSERT INTO repo for %v statement event: %v", repo, err)
	}
	defer CloseAndLog(stmt)

	var lastInsertId uuid.UUID
	err = stmt.QueryRow(repo.Id, repo.Url, repo.GitUrl, repo.Name, repo.Description, repo.Score, repo.Source, repo.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	repo.Id = lastInsertId
	return nil
}

func FindRepoById(repoId uuid.UUID) (*Repo, error) {
	var r Repo
	err := DB.
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

func FindReposByName(name string) ([]Repo, error) {
	rows, err := DB.Query("SELECT id, url, git_url, name, description, source FROM repo WHERE name=$1", name)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepo(rows)
}

func scanRepo(rows *sql.Rows) ([]Repo, error) {
	repos := []Repo{}
	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.Id, &r.Url, &r.GitUrl, &r.Name, &r.Description, &r.Source)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}
