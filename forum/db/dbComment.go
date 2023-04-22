package database

import (
	"forum/globals"
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
	PostID    string     `db:"post_id"`
}

func GetAllComments(postId uuid.UUID) ([]DbComment, error) {

	var comments []DbComment
	rows, err := globals.DB.Query(`SELECT id, author, content, created_at, updated_at, post_id FROM comment WHERE post_id = $1`, postId)
	if err != nil {
		return nil, err
	}

	defer closeAndLog(rows)

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
