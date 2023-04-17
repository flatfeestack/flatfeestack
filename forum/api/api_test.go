package api_test

import (
	"forum/api"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPosts(t *testing.T) {
	// Create a new instance of Server
	s := &api.Server{}

	// Create a mock HTTP request and response
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	w := httptest.NewRecorder()

	// Call the GetPosts method with the mock request and response
	s.GetPosts(w, req)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	expectedBody := "GetPosts called"
	if body := w.Body.String(); body != expectedBody {
		t.Errorf("Expected response body %q, but got %q", expectedBody, body)
	}
}
