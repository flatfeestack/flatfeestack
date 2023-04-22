package api

import (
	"context"
	database "forum/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type StrictServerImpl struct {
	// Implement any necessary dependencies or data stores here
}

func NewStrictServerImpl() *StrictServerImpl {
	// Initialize and return a new instance of StrictServerImpl
	return &StrictServerImpl{}
}

func (s *StrictServerImpl) GetPosts(ctx context.Context, request GetPostsRequestObject) (GetPostsResponseObject, error) {
	posts, err := database.GetAllPosts()
	if err != nil {
		log.Error(err)
		return GetPosts500Response{}, nil
	}
	if posts == nil {
		return GetPosts204JSONResponse{}, nil
	}
	var response GetPosts200JSONResponse
	for _, dbPost := range posts {
		post := mapDbPostToPost(dbPost)
		response = append(response, post)
	}
	return response, nil
}

func (s *StrictServerImpl) PostPosts(ctx context.Context, request PostPostsRequestObject) (PostPostsResponseObject, error) {
	id := getCurrentUserId(ctx)
	newPost, err := database.InsertPost(*id, request.Body.Title, request.Body.Content)
	if err != nil {
		return PostPosts500Response{}, err
	}
	return PostPosts201JSONResponse{
		Author:    newPost.Author,
		Content:   newPost.Content,
		CreatedAt: newPost.CreatedAt,
		Id:        newPost.Id,
		Open:      newPost.Open,
		Title:     newPost.Title,
		UpdatedAt: newPost.UpdatedAt,
	}, nil
}

func (s *StrictServerImpl) DeletePostsPostId(ctx context.Context, request DeletePostsPostIdRequestObject) (DeletePostsPostIdResponseObject, error) {
	err := database.DeletePost(request.PostId)
	if err != nil {
		errMsg := err.Error()
		return DeletePostsPostId204JSONResponse{NoContentJSONResponse{Info: &errMsg}}, err
	}
	return DeletePostsPostId200Response{}, nil
}

func (s *StrictServerImpl) GetPostsPostId(ctx context.Context, request GetPostsPostIdRequestObject) (GetPostsPostIdResponseObject, error) {
	// Implementation of GetPostsPostId method
	return nil, nil
}

func (s *StrictServerImpl) GetPostsPostIdComments(ctx context.Context, request GetPostsPostIdCommentsRequestObject) (GetPostsPostIdCommentsResponseObject, error) {
	// Implementation of GetPostsPostIdComments method
	return nil, nil
}

func (s *StrictServerImpl) PostPostsPostIdComments(ctx context.Context, request PostPostsPostIdCommentsRequestObject) (PostPostsPostIdCommentsResponseObject, error) {
	// Implementation of PostPostsPostIdComments method
	return nil, nil
}

func (s *StrictServerImpl) PutPostsPostIdComments(ctx context.Context, request PutPostsPostIdCommentsRequestObject) (PutPostsPostIdCommentsResponseObject, error) {
	// Implementation of PutPostsPostIdComments method
	return nil, nil
}

func (s *StrictServerImpl) DeletePostsPostIdCommentsCommentId(ctx context.Context, request DeletePostsPostIdCommentsCommentIdRequestObject) (DeletePostsPostIdCommentsCommentIdResponseObject, error) {
	// Implementation of DeletePostsPostIdCommentsCommentId method
	return nil, nil
}

func (s *StrictServerImpl) PutPostsPostIdCommentsCommentId(ctx context.Context, request PutPostsPostIdCommentsCommentIdRequestObject) (PutPostsPostIdCommentsCommentIdResponseObject, error) {
	// Implementation of PutPostsPostIdCommentsCommentId method
	return nil, nil
}

func getCurrentUserId(ctx context.Context) *uuid.UUID {
	user, ok := ctx.Value("currentUser").(*database.DbUser)
	if !ok {
		log.Error("value is not a *database.DbUser")
		return nil
	}
	return &user.Id
}

// Function to map DbPost to Post
func mapDbPostToPost(dbPost database.DbPost) Post {
	return Post{
		Author:    dbPost.Author,
		Content:   dbPost.Content,
		CreatedAt: dbPost.CreatedAt,
		Id:        dbPost.Id,
		Open:      dbPost.Open,
		Title:     dbPost.Title,
		UpdatedAt: dbPost.UpdatedAt,
	}
}
