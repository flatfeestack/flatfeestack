package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repo struct {
	Id          uuid.UUID `json:"id"`
	Url         *string   `json:"url"`
	GitUrl      *string   `json:"gitUrl"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Source      *string   `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (db *DB) InsertOrUpdateRepo(repo *Repo) error {
	var lastInsertId uuid.UUID
	err := db.QueryRow(`
		INSERT INTO repo (id, url, git_url, name, description, source, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(git_url) DO UPDATE SET 
			url = EXCLUDED.url,
			name = EXCLUDED.name,
			description = EXCLUDED.description
		RETURNING id`,
		repo.Id, repo.Url, repo.GitUrl, repo.Name, repo.Description, repo.Source, repo.CreatedAt).
		Scan(&lastInsertId)
	if err != nil {
		return err
	}
	repo.Id = lastInsertId
	return nil
}

func (db *DB) FindRepoById(repoId uuid.UUID) (*Repo, error) {
	var r Repo
	err := db.QueryRow(`
		SELECT id, url, git_url, name, description, source, created_at
		FROM repo 
		WHERE id = $1`, repoId).
		Scan(&r.Id, &r.Url, &r.GitUrl, &r.Name, &r.Description, &r.Source, &r.CreatedAt)
	
	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &r, nil
	default:
		return nil, err
	}
}

func (db *DB) FindReposByName(name string) ([]Repo, error) {
	rows, err := db.Query(`
		SELECT id, url, git_url, name, description, source, created_at
		FROM repo 
		WHERE name = $1`, name)
	if err != nil {
		return nil, err
	}
	defer CloseAndLog(rows)
	return scanRepos(rows)
}

func scanRepos(rows *sql.Rows) ([]Repo, error) {
	var repos []Repo
	for rows.Next() {
		var r Repo
		err := rows.Scan(&r.Id, &r.Url, &r.GitUrl, &r.Name, &r.Description, &r.Source, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}