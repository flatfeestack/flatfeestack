package database

import (
	"fmt"
	"forum/globals"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

type DbPost struct {
	Id        uuid.UUID
	Author    uuid.UUID
	Content   string
	CreatedAt time.Time
	Open      bool
	Title     string
	UpdatedAt *time.Time
}

func GetAllPosts() ([]DbPost, error) {
	var posts []DbPost
	rows, err := globals.DB.Query(`SELECT id, author, content, created_at, "open" ,title, updated_at FROM post`)
	if err != nil {
		return nil, err
	}

	defer closeAndLog(rows)

	for rows.Next() {
		var post DbPost
		err = rows.Scan(&post.Id, &post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func InsertPost(author uuid.UUID, title string, content string) (*DbPost, error) {
	stmt, err := globals.DB.Prepare(`INSERT INTO post (author, content, title) VALUES ($1, $2, $3) RETURNING id, created_at, "open" ,updated_at`)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(stmt)

	var post DbPost
	stmt.QueryRow(author, content, title).Scan(&post.Id, &post.CreatedAt, &post.Open, &post.UpdatedAt)

	post.Author = author
	post.Content = content
	post.Title = title

	return &post, nil
}

func DeletePost(id uuid.UUID) error {
	stmt, err := globals.DB.Prepare(`DELETE FROM post WHERE id = $1`)
	if err != nil {
		return err
	}
	defer closeAndLog(stmt)

	res, err := stmt.Exec(id)
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nr != 1 {
		return fmt.Errorf("id %v does not exist", id)
	}
	return nil
}

func GetPostById(id uuid.UUID) (*DbPost, error) {
	var post DbPost
	row, err := globals.DB.Query(`SELECT id, author, content, created_at, "open" ,title, updated_at FROM post where id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer closeAndLog(row)

	err = row.Scan(&post.Id, &post.Author, &post.Content, &post.CreatedAt, &post.Open, &post.Title, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func CheckIfPostExists(postId uuid.UUID) (bool, error) {
	var exists bool
	err := globals.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM post WHERE id = $1)`, postId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
