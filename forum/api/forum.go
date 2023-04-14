package api

import (
	"context"
	"fmt"
	database "forum/db"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sync"
)

type StrictServerImpl struct {
	Lock sync.Mutex
}

var _ StrictServerInterface = (*StrictServerImpl)(nil)

func NewStrictServerImpl() *StrictServerImpl {
	// Initialize and return a new instance of StrictServerImpl
	return &StrictServerImpl{}
}

func (s *StrictServerImpl) GetMetrics(ctx context.Context, request GetMetricsRequestObject) (GetMetricsResponseObject, error) {
	return GetMetrics200Response{}, nil
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
	newPost, err := database.InsertPost(id, request.Body.Title, request.Body.Content)
	if err != nil {
		log.Error(err)
		return PostPosts500Response{}, nil
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
	s.Lock.Lock()
	defer s.Lock.Unlock()
	err := database.DeletePost(request.PostId)
	if err != nil {
		log.Error(err)
		errMsg := err.Error()
		return DeletePostsPostId204JSONResponse{NoContentJSONResponse{Info: &errMsg}}, nil
	}
	return DeletePostsPostId200Response{}, nil
}

func (s *StrictServerImpl) GetPostsPostId(ctx context.Context, request GetPostsPostIdRequestObject) (GetPostsPostIdResponseObject, error) {
	dbPost, err := database.GetPostById(request.PostId)
	if err != nil {
		log.Error(err)
		return GetPostsPostId500Response{}, nil
	}
	if dbPost == nil {
		return GetPostsPostId404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}
	return GetPostsPostId200JSONResponse{
		Author:    dbPost.Author,
		Content:   dbPost.Content,
		CreatedAt: dbPost.CreatedAt,
		Id:        dbPost.Id,
		Open:      dbPost.Open,
		Title:     dbPost.Title,
		UpdatedAt: dbPost.UpdatedAt,
	}, nil
}

func (s *StrictServerImpl) PutPostsPostId(ctx context.Context, request PutPostsPostIdRequestObject) (PutPostsPostIdResponseObject, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	exists, err := database.CheckIfPostExists(request.PostId)
	if err != nil {
		log.Error(err)
		return PutPostsPostId500Response{}, nil
	}
	if !exists {
		return PutPostsPostId404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}
	closed, err := database.CheckIfPostIsClosed(request.PostId)
	if err != nil {
		log.Error(err)
		return PutPostsPostId500Response{}, nil
	}
	if closed {
		return PutPostsPostId403JSONResponse{ForbiddenJSONResponse{Error: fmt.Sprintf("this post is closed: %v", request.PostId)}}, nil
	}
	id := getCurrentUserId(ctx)
	authorId := database.GetPostAuthorId(request.PostId)
	if id == uuid.Nil || authorId == uuid.Nil || id != authorId {
		log.Errorf("user %v tried to update post: %v but post was written by %v", id, authorId, request.PostId)
		return PutPostsPostId403JSONResponse{ForbiddenJSONResponse{Error: fmt.Sprintf("you not author of this post: %v", request.PostId)}}, nil
	}
	updatedPost, err := database.UpdatePostByPostID(request.PostId, request.Body.Title, request.Body.Content)
	if err != nil {
		log.Error(err)
		return PutPostsPostId500Response{}, nil
	}
	return PutPostsPostId200JSONResponse{
		Id:        updatedPost.Id,
		Author:    updatedPost.Author,
		Title:     updatedPost.Title,
		Content:   updatedPost.Content,
		CreatedAt: updatedPost.CreatedAt,
		UpdatedAt: updatedPost.UpdatedAt,
		Open:      updatedPost.Open,
	}, nil
}

func (s *StrictServerImpl) PutPostsPostIdClose(ctx context.Context, request PutPostsPostIdCloseRequestObject) (PutPostsPostIdCloseResponseObject, error) {
	exists, err := database.CheckIfPostExists(request.PostId)
	if err != nil {
		log.Error(err)
		return PutPostsPostIdClose500Response{}, nil
	}
	if !exists {
		return PutPostsPostIdClose404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}

	if !isCurrentUserAdmin(ctx) {
		id := getCurrentUserId(ctx)
		authorId := database.GetPostAuthorId(request.PostId)
		if id == uuid.Nil || authorId == uuid.Nil || id != authorId {
			log.Errorf("user %v tried to close post: %v but post was written by %v", id, request.PostId, authorId)
			return PutPostsPostIdClose403JSONResponse{ForbiddenJSONResponse{Error: fmt.Sprintf("you not author of this post: %v", request.PostId)}}, nil
		}
	}

	err = database.ClosePost(request.PostId)
	if err != nil {
		log.Error(err)
		return PutPostsPostIdClose500Response{}, nil
	}
	return PutPostsPostIdClose200Response{}, nil
}

func (s *StrictServerImpl) GetPostsPostIdComments(ctx context.Context, request GetPostsPostIdCommentsRequestObject) (GetPostsPostIdCommentsResponseObject, error) {
	exists, err := database.CheckIfPostExists(request.PostId)
	if err != nil {
		log.Error(err)
		return GetPosts500Response{}, nil
	}
	if !exists {
		return GetPostsPostIdComments404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}

	dbComments, err := database.GetAllComments(request.PostId)
	if err != nil {
		log.Error(err)
		return GetPosts500Response{}, nil
	}
	if dbComments == nil {
		return GetPostsPostIdComments204JSONResponse{}, nil
	}
	var response GetPostsPostIdComments200JSONResponse
	for _, dbComment := range dbComments {
		comment := mapDbCommentToApiComment(dbComment)
		response = append(response, comment)
	}
	return response, nil
}

func (s *StrictServerImpl) PostPostsPostIdComments(ctx context.Context, request PostPostsPostIdCommentsRequestObject) (PostPostsPostIdCommentsResponseObject, error) {
	exists, err := database.CheckIfPostExists(request.PostId)
	if err != nil {
		log.Error(err)
		return PostPostsPostIdComments500Response{}, nil
	}
	if !exists {
		return PostPostsPostIdComments404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}
	id := getCurrentUserId(ctx)
	comment, err := database.InsertComment(request.PostId, id, request.Body.Content)
	if err != nil {
		log.Error(err)
		return PostPostsPostIdComments500Response{}, nil
	}

	return PostPostsPostIdComments201JSONResponse{
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		Id:        comment.ID,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}

func (s *StrictServerImpl) DeletePostsPostIdCommentsCommentId(ctx context.Context, request DeletePostsPostIdCommentsCommentIdRequestObject) (DeletePostsPostIdCommentsCommentIdResponseObject, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	err := database.DeleteComment(request.CommentId)
	if err != nil {
		log.Error(err)
		return DeletePostsPostIdCommentsCommentId404JSONResponse{NotFoundJSONResponse{Error: err.Error()}}, nil
	}
	return DeletePostsPostIdCommentsCommentId200Response{}, nil
}

func (s *StrictServerImpl) PutPostsPostIdCommentsCommentId(ctx context.Context, request PutPostsPostIdCommentsCommentIdRequestObject) (PutPostsPostIdCommentsCommentIdResponseObject, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	postExists, err := database.CheckIfPostExists(request.PostId)
	if !postExists {
		return PutPostsPostIdCommentsCommentId404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("post with id %v does not exist", request.PostId)}}, nil
	}
	commentExists, err := database.CheckIfCommentExists(request.CommentId)
	if !commentExists {
		return PutPostsPostIdCommentsCommentId404JSONResponse{NotFoundJSONResponse{Error: fmt.Sprintf("comment with id %v does not exist", request.CommentId)}}, nil
	}
	if err != nil {
		log.Error(err)
		return PutPostsPostIdCommentsCommentId500Response{}, nil
	}

	closed, err := database.CheckIfPostIsClosed(request.PostId)
	if err != nil {
		log.Error(err)
		return PutPostsPostId500Response{}, nil
	}
	if closed {
		return PutPostsPostIdCommentsCommentId403JSONResponse{ForbiddenJSONResponse{Error: fmt.Sprintf("the post to this comment is closed: %v", request.PostId)}}, nil
	}

	id := getCurrentUserId(ctx)
	authorId := database.GetCommentAuthor(request.CommentId)
	if id == uuid.Nil || authorId == uuid.Nil || id != authorId {
		log.Errorf("user %v tried to update comment: %v but comment was written by %v", id, request.CommentId, authorId)
		return PutPostsPostIdCommentsCommentId403JSONResponse{ForbiddenJSONResponse{Error: fmt.Sprintf("you not author of this comment: %v", request.CommentId)}}, nil
	}

	comment, err := database.UpdateComment(request.CommentId, request.Body.Content)
	if err != nil {
		log.Error(err)
		return PutPostsPostIdCommentsCommentId500Response{}, nil
	}
	return PutPostsPostIdCommentsCommentId200JSONResponse{
		Id:        comment.ID,
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}

func getCurrentUserId(ctx context.Context) uuid.UUID {
	user, ok := ctx.Value("currentUser").(*database.DbUser)
	if !ok {
		log.Error("value is not a *database.DbUser")
		return uuid.Nil
	}
	return user.Id
}

func isCurrentUserAdmin(ctx context.Context) bool {
	user, ok := ctx.Value("currentUser").(*database.DbUser)
	if !ok {
		log.Error("value is not a *database.DbUser")
		return false
	}
	return user.Role == "Admin"
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

func mapDbCommentToApiComment(dbComment database.DbComment) Comment {
	comment := Comment{
		Author:    dbComment.Author,
		Content:   dbComment.Content,
		CreatedAt: dbComment.CreatedAt,
		Id:        dbComment.ID,
		UpdatedAt: dbComment.UpdatedAt,
	}
	return comment
}
