package database

import (
	"forum/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestInsertPost(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should insert a post successfully", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		assert.NotEqual(t, uuid.Nil, post.Id, "Post ID should not be nil")
		assert.Equal(t, author, post.Author, "Incorrect author")
		assert.Equal(t, title, post.Title, "Incorrect title")
		assert.Equal(t, content, post.Content, "Incorrect content")
		assert.True(t, post.Open, "Open should be true by default")
		assert.Nil(t, post.UpdatedAt, "Updated at should be nil")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})
}

func TestGetAllPosts(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should retrieve all posts successfully", func(t *testing.T) {
		author := uuid.New()
		title1 := "Test Post 1"
		content1 := "This is test post 1"
		_, err := InsertPost(author, title1, content1)
		assert.NoError(t, err, "Failed to insert post")

		title2 := "Test Post 2"
		content2 := "This is test post 2"
		_, err = InsertPost(author, title2, content2)
		assert.NoError(t, err, "Failed to insert post")

		posts, err := GetAllPosts(nil)
		assert.NoError(t, err, "Failed to get all posts")
		assert.Len(t, posts, 2, "Incorrect number of posts")

		for _, post := range posts {
			assert.NotEqual(t, uuid.Nil, post.Id, "Post ID should not be nil")
			assert.Equal(t, author, post.Author, "Incorrect author")
			assert.True(t, post.Open, "Open should be true by default")
			assert.Nil(t, post.UpdatedAt, "Updated at should be nil")
		}

		// Clean up the posts
		for _, post := range posts {
			err = DeletePost(post.Id)
			assert.NoError(t, err, "Failed to delete post")
		}
	})
}

func TestGetPostById(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should retrieve a post by ID successfully", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert a post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		// Retrieve the post by ID
		retrievedPost, err := GetPostById(post.Id)
		assert.NoError(t, err, "Failed to retrieve post by ID")
		assert.NotNil(t, retrievedPost, "Retrieved post should not be nil")
		assert.Equal(t, post.Id, retrievedPost.Id, "Incorrect post ID")
		assert.Equal(t, author, retrievedPost.Author, "Incorrect author")
		assert.Equal(t, title, retrievedPost.Title, "Incorrect title")
		assert.Equal(t, content, retrievedPost.Content, "Incorrect content")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})

	t.Run("should return nil for non-existent post ID", func(t *testing.T) {
		nonExistentID := uuid.New()

		// Retrieve a non-existent post by ID
		retrievedPost, err := GetPostById(nonExistentID)
		assert.NoError(t, err, "Failed to retrieve post by ID")
		assert.Nil(t, retrievedPost, "Retrieved post should be nil")
	})
}

func TestUpdatePostByPostID(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should update a post successfully", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert a post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		newTitle := "Updated Post"
		newContent := "This is an updated post"

		// Update the post
		updatedPost, err := UpdatePostByPostID(post.Id, newTitle, newContent)
		assert.NoError(t, err, "Failed to update post")
		assert.NotNil(t, updatedPost, "Updated post should not be nil")
		assert.Equal(t, post.Id, updatedPost.Id, "Incorrect post ID")
		assert.Equal(t, author, updatedPost.Author, "Incorrect author")
		assert.Equal(t, newTitle, updatedPost.Title, "Incorrect title")
		assert.Equal(t, newContent, updatedPost.Content, "Incorrect content")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})

	t.Run("should return nil for non-existent post ID", func(t *testing.T) {
		nonExistentID := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Update a non-existent post
		updatedPost, err := UpdatePostByPostID(nonExistentID, title, content)
		assert.Error(t, err, "Expected an error when updating a non-existent post")
		assert.Nil(t, updatedPost, "Updated post should be nil")
	})
}

func TestAddProposalIdToPost(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should add a proposal ID to a post successfully", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert a post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		proposalID := big.NewInt(123)

		// Add a proposal ID to the post
		updatedPost, err := AddProposalIdToPost(post.Id, proposalID)
		assert.NoError(t, err, "Failed to add proposal ID to post")
		assert.NotNil(t, updatedPost, "Updated post should not be nil")
		assert.Equal(t, post.Id, updatedPost.Id, "Incorrect post ID")
		assert.Equal(t, author, updatedPost.Author, "Incorrect author")
		assert.Equal(t, title, updatedPost.Title, "Incorrect title")
		assert.Equal(t, content, updatedPost.Content, "Incorrect content")
		assert.Contains(t, updatedPost.ProposalIds, proposalID.String(), "Proposal ID not added to post")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})
}

func TestCheckIfPostIsClosed(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should return true for a closed post", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert a closed post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")
		err = ClosePost(post.Id)
		assert.NoError(t, err, "Failed to close post")

		// Check if the post is closed
		isClosed, err := CheckIfPostIsClosed(post.Id)
		assert.NoError(t, err, "Failed to check if post is closed")
		assert.True(t, isClosed, "Post should be closed")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})

	t.Run("should return false for an open post", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert an open post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		// Check if the post is closed
		isClosed, err := CheckIfPostIsClosed(post.Id)
		assert.NoError(t, err, "Failed to check if post is closed")
		assert.False(t, isClosed, "Post should not be closed")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})
}

func TestCheckIfPostExists(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()
	t.Run("should return true for an existing post", func(t *testing.T) {
		author := uuid.New()
		title := "Test Post"
		content := "This is a test post"

		// Insert a post
		post, err := InsertPost(author, title, content)
		assert.NoError(t, err, "Failed to insert post")

		// Check if the post exists
		exists, err := CheckIfPostExists(post.Id)
		assert.NoError(t, err, "Failed to check if post exists")
		assert.True(t, exists, "Post should exist")

		// Clean up the post
		err = DeletePost(post.Id)
		assert.NoError(t, err, "Failed to delete post")
	})

	t.Run("should return false for a non-existing post", func(t *testing.T) {
		// Generate a random post ID
		postID := uuid.New()

		// Check if the post exists
		exists, err := CheckIfPostExists(postID)
		assert.NoError(t, err, "Failed to check if post exists")
		assert.False(t, exists, "Post should not exist")
	})
}
