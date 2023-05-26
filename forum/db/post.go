package database

import (
	"database/sql"
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

type DbPost struct {
	Id         uuid.UUID
	Author     uuid.UUID
	Content    string
	CreatedAt  time.Time
	Open       bool
	Title      string
	UpdatedAt  *time.Time
	ProposalId *big.Int
}

func GetAllPosts(open *bool) ([]DbPost, error) {
	var posts []DbPost
	var rows *sql.Rows
	var err error

	query := `SELECT id, author, content, created_at, open, title, updated_at, proposal_id FROM post`

	if open != nil {
		rows, err = dbLib.DB.Query(query+" WHERE open = $1", open)
	} else {
		rows, err = dbLib.DB.Query(query)
	}
	if err != nil {
		return nil, err
	}

	defer dbLib.CloseAndLog(rows)

	for rows.Next() {
		var post DbPost
		var proposalId sql.NullString

		err = rows.Scan(&post.Id, &post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt, &post.ProposalId)
		if err != nil {
			return nil, err
		}

		if proposalId.Valid {
			post.ProposalId, _ = new(big.Int).SetString(proposalId.String, 10)
		}

		posts = append(posts, post)
	}
	return posts, nil
}

func InsertPost(author uuid.UUID, title string, content string) (*DbPost, error) {
	stmt, err := dbLib.DB.Prepare(`INSERT INTO post (author, content, title) VALUES ($1, $2, $3) RETURNING id, created_at, "open", updated_at`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	var post DbPost
	err = stmt.QueryRow(author, content, title).Scan(&post.Id, &post.CreatedAt, &post.Open, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	post.Author = author
	post.Content = content
	post.Title = title

	return &post, nil
}

func DeletePost(id uuid.UUID) error {
	stmt, err := dbLib.DB.Prepare(`DELETE FROM post WHERE id = $1`)
	if err != nil {
		return err
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(id)
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nr != 1 {
		return fmt.Errorf("post with id %v does not exist", id)
	}
	return nil
}

func ClosePost(postId uuid.UUID) error {
	stmt, err := dbLib.DB.Prepare(`UPDATE post SET "open" = false WHERE id = $1`)
	if err != nil {
		return err
	}
	defer dbLib.CloseAndLog(stmt)

	_, err = stmt.Exec(postId)
	if err != nil {
		return err
	}
	return nil
}

func GetPostById(id uuid.UUID) (*DbPost, error) {
	var post DbPost
	var proposalId sql.NullString

	stmt, err := dbLib.DB.Prepare(`SELECT id, author, content, created_at, open, title, updated_at, proposal_id FROM post where id = $1`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	err = stmt.QueryRow(id).Scan(&post.Id, &post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt, &proposalId)
	if err != nil {
		return nil, err
	}

	if proposalId.Valid {
		post.ProposalId, _ = new(big.Int).SetString(proposalId.String, 10)
	}

	return &post, nil
}

func UpdatePostByPostID(postId uuid.UUID, title string, content string) (*DbPost, error) {
	stmt, err := dbLib.DB.Prepare(`UPDATE post SET title=$1, content = $2, updated_at = $3 WHERE id = $4 RETURNING id, author, content, created_at, "open" ,title, updated_at`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	var post DbPost
	err = stmt.QueryRow(title, content, time.Now(), postId).Scan(&post.Id, &post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func GetPostAuthorId(postId uuid.UUID) uuid.UUID {
	var authorId = uuid.Nil
	err := dbLib.DB.QueryRow(`SELECT author FROM post WHERE id = $1`, postId).Scan(&authorId)
	if err != nil {
		return uuid.Nil
	}
	return authorId
}

func CheckIfPostExists(postId uuid.UUID) (bool, error) {
	var exists bool
	err := dbLib.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM post WHERE id = $1)`, postId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CheckIfPostIsClosed(postId uuid.UUID) (bool, error) {
	var closed bool
	err := dbLib.DB.QueryRow(`SELECT "open" FROM post WHERE id = $1`, postId).Scan(&closed)
	if err != nil {
		return false, err
	}
	return !closed, nil
}

func AddProposalIdToPost(postId uuid.UUID, proposalId *big.Int) (*DbPost, error) {
	stmt, err := dbLib.DB.Prepare(`UPDATE post SET proposal_id=$1 WHERE id = $2 RETURNING id, author, content, created_at, open, title, updated_at, proposal_id`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	var post DbPost

	err = stmt.QueryRow(proposalId.String(), postId).Scan(&post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	post.Id = postId
	post.ProposalId = proposalId

	return &post, nil
}
