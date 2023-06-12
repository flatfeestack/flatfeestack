package api

import (
	"context"
	"fmt"
	database "forum/db"
	"forum/types"
	"forum/utils"
	dbLib "github.com/flatfeestack/go-lib/database"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	container, err := utils.InitDatabase()
	if err != nil {
		log.Error(err)
		panic(err)
	}

	// Run tests
	code := m.Run()

	err = dbLib.DB.Close()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	defer container.Terminate(ctx)

	os.Exit(code)
}

func TestGetPosts(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should return all posts with open discussions", func(t *testing.T) {
		// Prepare the test data
		open := true
		request := GetPostsRequestObject{
			Params: GetPostsParams{
				Open: &open,
			},
		}

		// Create some example posts
		post1, _ := database.InsertPost(uuid.New(), "Post 1", "Content of post 1")
		post2, _ := database.InsertPost(uuid.New(), "Post 2", "Content of post 2")

		// Invoke the function
		response, err := s.GetPosts(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get posts")
		assert.IsType(t, GetPosts200JSONResponse{}, response, "Unexpected response type")

		posts := response.(GetPosts200JSONResponse)
		assert.Len(t, posts, 2, "Incorrect number of posts")

		// Verify the content of each post
		assert.Equal(t, post1.Id.String(), posts[0].Id.String(), "Incorrect post ID")
		assert.Equal(t, post1.Author.String(), posts[0].Author.String(), "Incorrect post author")
		assert.Equal(t, "Post 1", posts[0].Title, "Incorrect post title")
		assert.Equal(t, "Content of post 1", posts[0].Content, "Incorrect post content")
		assert.True(t, posts[0].Open, "Post should be open")

		assert.Equal(t, post2.Id.String(), posts[1].Id.String(), "Incorrect post ID")
		assert.Equal(t, post2.Author.String(), posts[1].Author.String(), "Incorrect post author")
		assert.Equal(t, "Post 2", posts[1].Title, "Incorrect post title")
		assert.Equal(t, "Content of post 2", posts[1].Content, "Incorrect post content")
		assert.True(t, posts[1].Open, "Post should be open")
	})

	t.Run("should return 204 No Content response when no posts found", func(t *testing.T) {
		// Prepare the test data
		open := false
		request := GetPostsRequestObject{
			Params: GetPostsParams{
				Open: &open,
			},
		}

		// Invoke the function
		response, err := s.GetPosts(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get posts")
		assert.IsType(t, GetPosts204JSONResponse{}, response, "Unexpected response type")
	})
}

func TestPostPosts(t *testing.T) {
	utils.Setup()
	defer utils.Teardown()

	s := NewStrictServerImpl()

	t.Run("should create a new post successfully", func(t *testing.T) {
		// Prepare the test data
		ctx := context.Background()
		userID := uuid.New()
		ctx = setCurrentUserID(ctx, userID, "User")
		request := PostPostsRequestObject{
			Body: &PostPostsJSONRequestBody{
				Title:   "New Post",
				Content: "This is the content of the new post.",
			},
		}

		// Invoke the function
		response, err := s.PostPosts(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to create post")
		assert.IsType(t, PostPosts201JSONResponse{}, response, "Unexpected response type")

		newPost := response.(PostPosts201JSONResponse)
		assert.Equal(t, userID.String(), newPost.Author.String(), "Incorrect post author")
		assert.Equal(t, "New Post", newPost.Title, "Incorrect post title")
		assert.Equal(t, "This is the content of the new post.", newPost.Content, "Incorrect post content")
		assert.True(t, newPost.Open, "Post should be open")
	})

	t.Run("should return a 400 Bad Request response for inappropriate language", func(t *testing.T) {
		// Prepare the test data
		ctx := context.Background()
		request := PostPostsRequestObject{
			Body: &PostPostsJSONRequestBody{
				Title:   "Fuck",
				Content: "This is bad language.",
			},
		}

		// Invoke the function
		response, err := s.PostPosts(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to create post")
		assert.IsType(t, PostPosts400JSONResponse{}, response, "Unexpected response type")

		badRequest := response.(PostPosts400JSONResponse)
		assert.Equal(t, "Please use appropriate language!", badRequest.Error, "Incorrect error message")
	})
}

func TestDeletePostsPostId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should delete the post and return 200 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Create the request object
		request := DeletePostsPostIdRequestObject{
			PostId: post.Id,
		}

		// Invoke the DeletePostsPostId method
		response, err := s.DeletePostsPostId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to delete post")
		assert.IsType(t, DeletePostsPostId200Response{}, response, "Unexpected response type")
	})

	t.Run("should handle error during post deletion and return 204 response", func(t *testing.T) {
		// Create the request object
		request := DeletePostsPostIdRequestObject{
			PostId: uuid.New(),
		}
		// Invoke the DeletePostsPostId method
		response, err := s.DeletePostsPostId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to delete post")
		assert.IsType(t, DeletePostsPostId204JSONResponse{}, response, "Unexpected response type")
	})
}

func TestGetPostsPostId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should return the post with 200 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Create the request object
		request := GetPostsPostIdRequestObject{
			PostId: post.Id,
		}

		// Invoke the GetPostsPostId method
		response, err := s.GetPostsPostId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post")
		assert.IsType(t, GetPostsPostId200JSONResponse{}, response, "Unexpected response type")

		postResponse := response.(GetPostsPostId200JSONResponse)
		assert.Equal(t, post.Author, postResponse.Author, "Incorrect post author")
		assert.Equal(t, post.Content, postResponse.Content, "Incorrect post content")
		assert.Equal(t, post.CreatedAt, postResponse.CreatedAt, "Incorrect post created time")
		assert.Equal(t, post.Id, postResponse.Id, "Incorrect post ID")
		assert.Equal(t, post.Open, postResponse.Open, "Incorrect post open status")
		assert.Equal(t, post.Title, postResponse.Title, "Incorrect post title")
		assert.Equal(t, post.UpdatedAt, postResponse.UpdatedAt, "Incorrect post updated time")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := GetPostsPostIdRequestObject{
			PostId: uuid.New(),
		}

		// Invoke the GetPostsPostId method
		response, err := s.GetPostsPostId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post")
		assert.IsType(t, GetPostsPostId404JSONResponse{}, response, "Unexpected response type")
	})
}

func TestPutPostsPostId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should update the post and return 200 response", func(t *testing.T) {
		authorId := uuid.New()
		// Insert a mock post
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PutPostsPostIdRequestObject{
			PostId: post.Id,
			Body: &PutPostsPostIdJSONRequestBody{
				Title:   "Updated Post Title",
				Content: "Updated post content",
			},
		}

		// Invoke the PutPostsPostId method
		ctx := setCurrentUserID(context.Background(), authorId, "User")
		response, err := s.PutPostsPostId(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to update post")
		assert.IsType(t, PutPostsPostId200JSONResponse{}, response, "Unexpected response type")

		updatedPost := response.(PutPostsPostId200JSONResponse)
		assert.Equal(t, post.Author, updatedPost.Author, "Incorrect post author")
		assert.Equal(t, request.Body.Title, updatedPost.Title, "Incorrect post title")
		assert.Equal(t, request.Body.Content, updatedPost.Content, "Incorrect post content")
		assert.Equal(t, post.CreatedAt, updatedPost.CreatedAt, "Incorrect post created time")
		assert.Equal(t, post.Id, updatedPost.Id, "Incorrect post ID")
		assert.Equal(t, post.Open, updatedPost.Open, "Incorrect post open status")
		assert.NotNil(t, updatedPost.UpdatedAt, "Updated post should have non-nil UpdatedAt value")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := PutPostsPostIdRequestObject{
			PostId: uuid.New(),
			Body: &PutPostsPostIdJSONRequestBody{
				Title:   "Updated Post Title",
				Content: "Updated post content",
			},
		}

		// Invoke the PutPostsPostId method
		response, err := s.PutPostsPostId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to update post")
		assert.IsType(t, PutPostsPostId404JSONResponse{}, response, "Unexpected response type")
	})

	t.Run("should return 403 response if user is not the author of the post", func(t *testing.T) {
		authorId := uuid.New()
		// Insert a mock post
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PutPostsPostIdRequestObject{
			PostId: post.Id,
			Body: &PutPostsPostIdJSONRequestBody{
				Title:   "Updated Post Title",
				Content: "Updated post content",
			},
		}

		// Set a different user ID in the context
		otherUserId := uuid.New()
		ctx := setCurrentUserID(context.Background(), otherUserId, "User")

		// Invoke the PutPostsPostId method
		response, err := s.PutPostsPostId(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to update post")
		assert.IsType(t, PutPostsPostId403JSONResponse{}, response, "Unexpected response type")

		forbiddenResponse := response.(PutPostsPostId403JSONResponse)
		assert.Equal(t, fmt.Sprintf("you not author of this post: %v", request.PostId), forbiddenResponse.Error, "Incorrect error message")
	})
}

func TestPutPostsPostIdClose(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should close the post and return 200 response", func(t *testing.T) {
		authorId := uuid.New()
		// Insert a mock post
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PutPostsPostIdCloseRequestObject{
			PostId: post.Id,
		}

		// Set the user as an admin in the context
		ctx := setCurrentUserID(context.Background(), authorId, "User")

		// Invoke the PutPostsPostIdClose method
		response, err := s.PutPostsPostIdClose(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to close post")
		assert.IsType(t, PutPostsPostIdClose200Response{}, response, "Unexpected response type")
	})

	t.Run("admin should be able to close post", func(t *testing.T) {
		authorId := uuid.New()
		// Insert a mock post
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PutPostsPostIdCloseRequestObject{
			PostId: post.Id,
		}

		// Set the user as an admin in the context
		ctx := setCurrentUserID(context.Background(), uuid.New(), "Admin")

		// Invoke the PutPostsPostIdClose method
		response, err := s.PutPostsPostIdClose(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to close post")
		assert.IsType(t, PutPostsPostIdClose200Response{}, response, "Unexpected response type")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := PutPostsPostIdCloseRequestObject{
			PostId: uuid.New(),
		}

		// Invoke the PutPostsPostIdClose method
		response, err := s.PutPostsPostIdClose(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to close post")
		assert.IsType(t, PutPostsPostIdClose404JSONResponse{}, response, "Unexpected response type")
	})

	t.Run("should handle error when user is not the author and return 403 response", func(t *testing.T) {
		authorId := uuid.New()
		// Insert a mock post
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PutPostsPostIdCloseRequestObject{
			PostId: post.Id,
		}

		// Set a different user ID in the context
		otherUserId := uuid.New()
		ctx := setCurrentUserID(context.Background(), otherUserId, "User")

		// Invoke the PutPostsPostIdClose method
		response, err := s.PutPostsPostIdClose(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to close post")
		assert.IsType(t, PutPostsPostIdClose403JSONResponse{}, response, "Unexpected response type")

		forbiddenResponse := response.(PutPostsPostIdClose403JSONResponse)
		assert.Equal(t, fmt.Sprintf("you not author of this post: %v", request.PostId), forbiddenResponse.Error, "Incorrect error message")
	})
}

func TestGetPostsPostIdComments(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should return the comments for an existing post", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Insert mock comments for the post
		comment1 := insertMockComment(post.Id, uuid.New(), "Comment 1")
		comment2 := insertMockComment(post.Id, uuid.New(), "Comment 2")

		// Create the request object
		request := GetPostsPostIdCommentsRequestObject{
			PostId: post.Id,
		}

		// Invoke the GetPostsPostIdComments method
		response, err := s.GetPostsPostIdComments(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post comments")
		assert.IsType(t, GetPostsPostIdComments200JSONResponse{}, response, "Unexpected response type")

		comments := response.(GetPostsPostIdComments200JSONResponse)
		assert.Equal(t, 2, len(comments), "Incorrect number of comments")

		assertCommentEquals(t, comment1, comments[0])
		assertCommentEquals(t, comment2, comments[1])
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := GetPostsPostIdCommentsRequestObject{
			PostId: uuid.New(),
		}

		// Invoke the GetPostsPostIdComments method
		response, err := s.GetPostsPostIdComments(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post comments")
		assert.IsType(t, GetPostsPostIdComments404JSONResponse{}, response, "Unexpected response type")
	})
}

func TestPostPostsPostIdComments(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should create a new comment for an existing post", func(t *testing.T) {
		// Insert a mock post
		authorId := uuid.New()
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Create the request object
		request := PostPostsPostIdCommentsRequestObject{
			PostId: post.Id,
			Body: &PostPostsPostIdCommentsJSONRequestBody{
				Content: "New comment content",
			},
		}

		// Invoke the PostPostsPostIdComments method
		ctx := setCurrentUserID(context.Background(), authorId, "User")
		response, err := s.PostPostsPostIdComments(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to create comment")
		assert.IsType(t, PostPostsPostIdComments201JSONResponse{}, response, "Unexpected response type")

		comment := response.(PostPostsPostIdComments201JSONResponse)
		assert.Equal(t, request.Body.Content, comment.Content, "Incorrect comment content")
		assert.Equal(t, authorId, comment.Author, "Incorrect comment content")
		assert.NotNil(t, comment.Id, "Comment ID should not be nil")
		assert.Nil(t, comment.UpdatedAt, "Comment UpdatedAt should be nil")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := PostPostsPostIdCommentsRequestObject{
			PostId: uuid.New(),
			Body: &PostPostsPostIdCommentsJSONRequestBody{
				Content: "New comment content",
			},
		}

		// Invoke the PostPostsPostIdComments method
		response, err := s.PostPostsPostIdComments(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to create comment")
		assert.IsType(t, PostPostsPostIdComments404JSONResponse{}, response, "Unexpected response type")
	})

	t.Run("should handle inappropriate content and return 400 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Create the request object with inappropriate content
		request := PostPostsPostIdCommentsRequestObject{
			PostId: post.Id,
			Body: &PostPostsPostIdCommentsJSONRequestBody{
				Content: "Fuck",
			},
		}

		// Invoke the PostPostsPostIdComments method
		response, err := s.PostPostsPostIdComments(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to create comment")
		assert.IsType(t, PostPostsPostIdComments400JSONResponse{}, response, "Unexpected response type")
	})
}

func TestDeletePostsPostIdCommentsCommentId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should delete an existing comment and return 200 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Insert a mock comment for the post
		comment := insertMockComment(post.Id, uuid.New(), "Test Comment")

		// Create the request object
		request := DeletePostsPostIdCommentsCommentIdRequestObject{
			PostId:    post.Id,
			CommentId: comment.ID,
		}

		// Invoke the DeletePostsPostIdCommentsCommentId method
		response, err := s.DeletePostsPostIdCommentsCommentId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to delete comment")
		assert.IsType(t, DeletePostsPostIdCommentsCommentId200Response{}, response, "Unexpected response type")
	})

	t.Run("should handle error when comment does not exist and return 404 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Create the request object with a non-existent comment ID
		request := DeletePostsPostIdCommentsCommentIdRequestObject{
			PostId:    post.Id,
			CommentId: uuid.New(),
		}

		// Invoke the DeletePostsPostIdCommentsCommentId method
		response, err := s.DeletePostsPostIdCommentsCommentId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to delete comment")
		assert.IsType(t, DeletePostsPostIdCommentsCommentId404JSONResponse{}, response, "Unexpected response type")
	})
}

func TestPutPostsPostIdCommentsCommentId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should update an existing comment and return 200 response", func(t *testing.T) {
		// Insert a mock post
		authorId := uuid.New()
		post := insertMockPost(authorId, "Test Post", "This is a test post")

		// Insert a mock comment for the post
		comment := insertMockComment(post.Id, authorId, "Test Comment")

		// Create the request object
		request := PutPostsPostIdCommentsCommentIdRequestObject{
			PostId:    post.Id,
			CommentId: comment.ID,
			Body: &PutPostsPostIdCommentsCommentIdJSONRequestBody{
				Content: "Updated Comment",
			},
		}

		// Invoke the PutPostsPostIdCommentsCommentId method
		ctx := setCurrentUserID(context.Background(), authorId, "User")
		response, err := s.PutPostsPostIdCommentsCommentId(ctx, request)

		// Assertions
		assert.NoError(t, err, "Failed to update comment")
		assert.IsType(t, PutPostsPostIdCommentsCommentId200JSONResponse{}, response, "Unexpected response type")
		assert.Equal(t, request.Body.Content, response.(PutPostsPostIdCommentsCommentId200JSONResponse).Content, "Comment content was not updated")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent post ID
		request := PutPostsPostIdCommentsCommentIdRequestObject{
			PostId:    uuid.New(),
			CommentId: uuid.New(),
			Body: &PutPostsPostIdCommentsCommentIdJSONRequestBody{
				Content: "Updated Comment",
			},
		}

		// Invoke the PutPostsPostIdCommentsCommentId method
		response, err := s.PutPostsPostIdCommentsCommentId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to update comment")
		assert.IsType(t, PutPostsPostIdCommentsCommentId404JSONResponse{}, response, "Unexpected response type")
	})

	t.Run("should handle error when comment does not exist and return 404 response", func(t *testing.T) {
		// Insert a mock post
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")

		// Create the request object with a non-existent comment ID
		request := PutPostsPostIdCommentsCommentIdRequestObject{
			PostId:    post.Id,
			CommentId: uuid.New(),
			Body: &PutPostsPostIdCommentsCommentIdJSONRequestBody{
				Content: "Updated Comment",
			},
		}

		// Invoke the PutPostsPostIdCommentsCommentId method
		response, err := s.PutPostsPostIdCommentsCommentId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to update comment")
		assert.IsType(t, PutPostsPostIdCommentsCommentId404JSONResponse{}, response, "Unexpected response type")
	})
}

func TestGetPostsByProposalIdProposalId(t *testing.T) {
	utils.Setup()
	s := NewStrictServerImpl()
	defer utils.Teardown()

	t.Run("should return the post for an existing proposal ID", func(t *testing.T) {
		// Insert a mock post with a proposal ID
		proposalID := big.NewInt(123)
		post := insertMockPost(uuid.New(), "Test Post", "This is a test post")
		addMockProposalIdToPost(post.Id, proposalID)

		// Create the request object
		request := GetPostsByProposalIdProposalIdRequestObject{
			ProposalId: proposalID.String(),
		}

		// Invoke the GetPostsByProposalIdProposalId method
		response, err := s.GetPostsByProposalIdProposalId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post by proposal ID")
		assert.IsType(t, GetPostsByProposalIdProposalId200JSONResponse{}, response, "Unexpected response type")
		assert.Equal(t, post.Author, response.(GetPostsByProposalIdProposalId200JSONResponse).Author, "Incorrect author")
		assert.Equal(t, post.Content, response.(GetPostsByProposalIdProposalId200JSONResponse).Content, "Incorrect content")
		assert.Equal(t, post.CreatedAt, response.(GetPostsByProposalIdProposalId200JSONResponse).CreatedAt, "Incorrect created time")
		assert.Equal(t, post.Id, response.(GetPostsByProposalIdProposalId200JSONResponse).Id, "Incorrect ID")
		assert.Equal(t, post.Open, response.(GetPostsByProposalIdProposalId200JSONResponse).Open, "Incorrect open status")
		assert.Equal(t, post.Title, response.(GetPostsByProposalIdProposalId200JSONResponse).Title, "Incorrect title")
		assert.Equal(t, post.UpdatedAt, response.(GetPostsByProposalIdProposalId200JSONResponse).UpdatedAt, "Incorrect updated time")
		assert.Equal(t, []string{proposalID.String()}, response.(GetPostsByProposalIdProposalId200JSONResponse).ProposalIds, "Incorrect proposal IDs")
	})

	t.Run("should handle error when post does not exist and return 404 response", func(t *testing.T) {
		// Create the request object with a non-existent proposal ID
		request := GetPostsByProposalIdProposalIdRequestObject{
			ProposalId: big.NewInt(456).String(),
		}

		// Invoke the GetPostsByProposalIdProposalId method
		response, err := s.GetPostsByProposalIdProposalId(context.Background(), request)

		// Assertions
		assert.NoError(t, err, "Failed to get post by proposal ID")
		assert.IsType(t, GetPostsByProposalIdProposalId404JSONResponse{}, response, "Unexpected response type")
	})
}

func assertCommentEquals(t *testing.T, expected *database.DbComment, actual Comment) {
	assert.Equal(t, expected.Author, actual.Author, "Incorrect comment author")
	assert.Equal(t, expected.Content, actual.Content, "Incorrect comment content")
	assert.Equal(t, expected.CreatedAt, actual.CreatedAt, "Incorrect comment created time")
	assert.Equal(t, expected.ID, actual.Id, "Incorrect comment ID")
	assert.Equal(t, expected.UpdatedAt, actual.UpdatedAt, "Incorrect comment updated time")
}

func insertMockComment(postId uuid.UUID, authorId uuid.UUID, content string) *database.DbComment {
	comment, err := database.InsertComment(postId, authorId, content)
	if err != nil {
		return nil
	}
	return comment
}

func insertMockPost(authorId uuid.UUID, title, content string) *database.DbPost {
	post, err := database.InsertPost(authorId, title, content)
	if err != nil {
		return nil
	}
	return post
}

func addMockProposalIdToPost(postId uuid.UUID, proposalId *big.Int) *database.DbPost {
	post, err := database.AddProposalIdToPost(postId, proposalId)
	if err != nil {
		return nil
	}
	return post
}

func setCurrentUserID(ctx context.Context, userID uuid.UUID, role string) context.Context {
	user := &types.User{
		Id:   userID,
		Role: role,
	}
	return context.WithValue(ctx, "currentUser", user)
}
