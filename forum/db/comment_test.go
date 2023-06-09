package database

import (
	"forum/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createPost(t *testing.T) uuid.UUID {
	author := uuid.New()
	title := "Test Post"
	content := "This is a test post."

	post, err := InsertPost(author, title, content)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	return post.Id
}

func TestGetAllComments(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should return all comments for a given post ID", func(t *testing.T) {
		postID := createPost(t)

		// Insert some comments for the post
		comment1, err := InsertComment(postID, uuid.New(), "Comment 1")
		assert.NoError(t, err, "Failed to insert comment")
		comment2, err := InsertComment(postID, uuid.New(), "Comment 2")
		assert.NoError(t, err, "Failed to insert comment")

		// Retrieve all comments for the post
		comments, err := GetAllComments(postID)
		assert.NoError(t, err, "Failed to get comments")
		assert.Len(t, comments, 2, "Incorrect number of comments")

		// Check the content of each comment
		assert.Equal(t, comment1.ID, comments[0].ID, "Incorrect comment ID")
		assert.Equal(t, comment1.Author, comments[0].Author, "Incorrect comment author")
		assert.Equal(t, comment1.Content, comments[0].Content, "Incorrect comment content")
		assert.Equal(t, comment1.CreatedAt, comments[0].CreatedAt, "Incorrect comment created time")
		assert.Equal(t, comment1.UpdatedAt, comments[0].UpdatedAt, "Incorrect comment updated time")
		assert.Equal(t, comment1.PostID, comments[0].PostID, "Incorrect comment post ID")

		assert.Equal(t, comment2.ID, comments[1].ID, "Incorrect comment ID")
		assert.Equal(t, comment2.Author, comments[1].Author, "Incorrect comment author")
		assert.Equal(t, comment2.Content, comments[1].Content, "Incorrect comment content")
		assert.Equal(t, comment2.CreatedAt, comments[1].CreatedAt, "Incorrect comment created time")
		assert.Equal(t, comment2.UpdatedAt, comments[1].UpdatedAt, "Incorrect comment updated time")
		assert.Equal(t, comment2.PostID, comments[1].PostID, "Incorrect comment post ID")

		// Clean up the comments
		err = DeleteComment(comment1.ID)
		assert.NoError(t, err, "Failed to delete comment")
		err = DeleteComment(comment2.ID)
		assert.NoError(t, err, "Failed to delete comment")
	})
}

func TestInsertComment(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should insert a new comment successfully", func(t *testing.T) {
		postID := createPost(t)
		authorID := uuid.New()
		content := "New comment"

		// Insert a new comment
		comment, err := InsertComment(postID, authorID, content)
		assert.NoError(t, err, "Failed to insert comment")
		assert.NotNil(t, comment, "Inserted comment should not be nil")
		assert.NotEqual(t, uuid.Nil, comment.ID, "Comment ID should not be nil")
		assert.Equal(t, postID, comment.PostID, "Incorrect comment post ID")
		assert.Equal(t, authorID, comment.Author, "Incorrect comment author")
		assert.Equal(t, content, comment.Content, "Incorrect comment content")
		assert.WithinDuration(t, time.Now(), comment.CreatedAt, 1*time.Second, "Incorrect comment created time")
		assert.Nil(t, comment.UpdatedAt, "Updated at should be nil for a new comment")

		// Clean up the comment
		err = DeleteComment(comment.ID)
		assert.NoError(t, err, "Failed to delete comment")
	})
}

func TestDeleteComment(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should delete an existing comment", func(t *testing.T) {
		postID := createPost(t)

		// Insert a comment to be deleted
		comment, err := InsertComment(postID, uuid.New(), "Comment to delete")
		assert.NoError(t, err, "Failed to insert comment")

		// Delete the comment
		err = DeleteComment(comment.ID)
		assert.NoError(t, err, "Failed to delete comment")

		// Check if the comment still exists
		exists, err := CheckIfCommentExists(comment.ID)
		assert.NoError(t, err, "Failed to check if comment exists")
		assert.False(t, exists, "Comment should not exist")
	})

	t.Run("should return an error for a non-existing comment", func(t *testing.T) {
		// Generate a random comment ID
		commentID := uuid.New()

		// Delete the comment
		err := DeleteComment(commentID)
		assert.Error(t, err, "Expected an error when deleting non-existing comment")
	})
}

func TestUpdateComment(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should update an existing comment", func(t *testing.T) {
		postID := createPost(t)
		content := "Updated comment"

		// Insert a comment to be updated
		comment, err := InsertComment(postID, uuid.New(), "Comment to update")
		assert.NoError(t, err, "Failed to insert comment")

		// Update the comment
		updatedComment, err := UpdateComment(comment.ID, content)
		assert.NoError(t, err, "Failed to update comment")
		assert.NotNil(t, updatedComment, "Updated comment should not be nil")
		assert.Equal(t, comment.ID, updatedComment.ID, "Incorrect comment ID")
		assert.Equal(t, content, updatedComment.Content, "Incorrect comment content")
		assert.WithinDuration(t, time.Now(), (*updatedComment.UpdatedAt).UTC(), 1*time.Second, "Incorrect comment updated time")

		// Clean up the comment
		err = DeleteComment(comment.ID)
		assert.NoError(t, err, "Failed to delete comment")
	})

	t.Run("should return an error for a non-existing comment", func(t *testing.T) {
		// Generate a random comment ID
		commentID := uuid.New()
		content := "Updated comment"

		// Update the comment
		updatedComment, err := UpdateComment(commentID, content)
		assert.Error(t, err, "Expected an error when updating non-existing comment")
		assert.Nil(t, updatedComment, "Updated comment should be nil")
	})
}

func TestCheckIfCommentExists(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should return true for an existing comment", func(t *testing.T) {
		postID := createPost(t)

		// Insert a comment to check
		comment, err := InsertComment(postID, uuid.New(), "Comment to check")
		assert.NoError(t, err, "Failed to insert comment")

		// Check if the comment exists
		exists, err := CheckIfCommentExists(comment.ID)
		assert.NoError(t, err, "Failed to check if comment exists")
		assert.True(t, exists, "Comment should exist")

		// Clean up the comment
		err = DeleteComment(comment.ID)
		assert.NoError(t, err, "Failed to delete comment")
	})

	t.Run("should return false for a non-existing comment", func(t *testing.T) {
		// Generate a random comment ID
		commentID := uuid.New()

		// Check if the comment exists
		exists, err := CheckIfCommentExists(commentID)
		assert.NoError(t, err, "Failed to check if comment exists")
		assert.False(t, exists, "Comment should not exist")
	})
}

func TestGetCommentAuthor(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	t.Run("should return the author ID for an existing comment", func(t *testing.T) {
		postID := createPost(t)
		authorID := uuid.New()

		// Insert a comment with the author ID
		comment, err := InsertComment(postID, authorID, "Comment with author")
		assert.NoError(t, err, "Failed to insert comment")

		// Get the comment author
		author := GetCommentAuthor(comment.ID)
		assert.Equal(t, authorID, author, "Incorrect comment author")

		// Clean up the comment
		err = DeleteComment(comment.ID)
		assert.NoError(t, err, "Failed to delete comment")
	})

	t.Run("should return uuid.Nil for a non-existing comment", func(t *testing.T) {
		// Generate a random comment ID
		commentID := uuid.New()

		// Get the comment author
		author := GetCommentAuthor(commentID)
		assert.Equal(t, uuid.Nil, author, "Comment author should be uuid.Nil")
	})
}
