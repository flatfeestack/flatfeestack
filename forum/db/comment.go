package database

import (
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

type DbComment struct {
	ID        uuid.UUID  `db:"id"`
	Author    uuid.UUID  `db:"author"`
	Content   string     `db:"content"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	PostID    uuid.UUID  `db:"post_id"`
}

func GetAllComments(postId uuid.UUID) ([]DbComment, error) {

	var comments []DbComment
	rows, err := dbLib.DB.Query(`SELECT id, author, content, created_at, updated_at, post_id FROM comment WHERE post_id = $1`, postId)
	if err != nil {
		return nil, err
	}

	defer dbLib.CloseAndLog(rows)

	for rows.Next() {
		var comment DbComment
		err = rows.Scan(&comment.ID, &comment.Author, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt, &comment.PostID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func InsertComment(postId uuid.UUID, author uuid.UUID, content string) (*DbComment, error) {
	stmt, err := dbLib.DB.Prepare(`INSERT INTO comment (author, content, post_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	var comment DbComment
	err = stmt.QueryRow(author, content, postId).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return nil, err
	}

	comment.Author = author
	comment.Content = content
	comment.PostID = postId

	return &comment, nil
}

func DeleteComment(commentId uuid.UUID) error {
	stmt, err := dbLib.DB.Prepare(`DELETE FROM comment WHERE id = $1`)
	if err != nil {
		return err
	}
	defer dbLib.CloseAndLog(stmt)

	res, err := stmt.Exec(commentId)
	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nr != 1 {
		return fmt.Errorf("comment with id %v does not exist", commentId)
	}
	return nil
}

func UpdateComment(commentId uuid.UUID, content string) (*DbComment, error) {
	stmt, err := dbLib.DB.Prepare(`UPDATE comment SET content = $1, updated_at = $2 WHERE id = $3 RETURNING id, author, created_at, updated_at, post_id`)
	if err != nil {
		return nil, err
	}
	defer dbLib.CloseAndLog(stmt)

	var comment DbComment
	err = stmt.QueryRow(content, time.Now(), commentId).Scan(&comment.ID, &comment.Author, &comment.CreatedAt, &comment.UpdatedAt, &comment.PostID)
	if err != nil {
		return nil, err
	}

	comment.Content = content

	return &comment, nil
}

func CheckIfCommentExists(commentId uuid.UUID) (bool, error) {
	var exists bool
	err := dbLib.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM comment WHERE id = $1)`, commentId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetCommentAuthor(commentId uuid.UUID) uuid.UUID {
	var author uuid.UUID
	err := dbLib.DB.QueryRow(`SELECT author FROM comment WHERE id = $1`, commentId).Scan(&author)
	if err != nil {
		return uuid.Nil
	}
	return author
}
