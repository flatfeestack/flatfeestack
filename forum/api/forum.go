package api

import (
	"fmt"
	"net/http"
)

type Server struct{}

// Get all posts
// (GET /posts)
func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GetPosts called")
}

// Create a new post
// (POST /posts)
func (s *Server) PostPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "PostPosts called")
}

// Delete a Post
// (DELETE /posts/{postId})
func (s *Server) DeletePostsPostId(w http.ResponseWriter, r *http.Request, postId PostId) {
	fmt.Fprintf(w, "DeletePostsPostId called with postId=%s", postId)
}

// Get a specific post
// (GET /posts/{postId})
func (s *Server) GetPostsPostId(w http.ResponseWriter, r *http.Request, postId PostId) {
	fmt.Fprintf(w, "GetPostsPostId called with postId=%s", postId)
}

// Get all comments
// (GET /posts/{postId}/comments)
func (s *Server) GetPostsPostIdComments(w http.ResponseWriter, r *http.Request, postId PostId) {
	fmt.Fprintf(w, "GetPostsPostIdComments called with postId=%s", postId)
}

// Add a comment to a post
// (POST /posts/{postId}/comments)
func (s *Server) PostPostsPostIdComments(w http.ResponseWriter, r *http.Request, postId PostId) {
	fmt.Fprintf(w, "PostPostsPostIdComments called with postId=%s", postId)
}

// Update a post
// (PUT /posts/{postId}/comments)
func (s *Server) PutPostsPostIdComments(w http.ResponseWriter, r *http.Request, postId PostId) {
	fmt.Fprintf(w, "PutPostsPostIdComments called with postId=%s", postId)
}

// Delete a comment
// (DELETE /posts/{postId}/comments/{commentId})
func (s *Server) DeletePostsPostIdCommentsCommentId(w http.ResponseWriter, r *http.Request, postId PostId, commentId CommentId) {
	fmt.Fprintf(w, "DeletePostsPostIdCommentsCommentId called with postId=%s, commentId=%s", postId, commentId)
}

// Update a comment
// (PUT /posts/{postId}/comments/{commentId})
func (s *Server) PutPostsPostIdCommentsCommentId(w http.ResponseWriter, r *http.Request, postId PostId, commentId CommentId) {
	fmt.Fprintf(w, "PutPostsPostIdCommentsCommentId called with postId=%s, commentId=%s", postId, commentId)
}
