package api_test

import (
	"context"
	"forum/api"
	"github.com/DATA-DOG/go-sqlmock"
	dbLib "github.com/flatfeestack/go-lib/database"
	"testing"
	"time"
)

// TestGetPosts tests the GetPosts function
func TestGetPosts(t *testing.T) {
	// TODO: Has to be improved
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	dbLib.DB = db
	defer db.Close()

	mock.ExpectQuery(`SELECT id, author, content, created_at, "open" ,title, updated_at FROM post`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "author", "content", "created_at", "open", "title", "updated_at"}).
			AddRow("8bef1c41-7625-482e-8589-25cfb31b14a4", "0798e80e-8be1-4ac5-887c-1395ed841ebe", "Test content", time.Now(), true, "Test title", nil),
	)

	request := api.GetPostsRequestObject{}
	server := &api.StrictServerImpl{}

	response, err := server.GetPosts(context.Background(), request)

	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	_, ok := response.(api.GetPosts200JSONResponse)
	if !ok {
		t.Fatalf("Expected response of type GetPosts200JSONResponse, but got: %T", response)
	}
}
