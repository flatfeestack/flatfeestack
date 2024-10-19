package api

import (
	"backend/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserById(t *testing.T) {
	t.Run("should return user", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		userDetail := insertTestUser(t, "hello@world.com")
		request, _ := http.NewRequest(http.MethodPost, "/users/"+userDetail.Id.String(), nil)
		//it does not go via router, so add it manually
		request.SetPathValue("id", userDetail.Id.String())
		response := httptest.NewRecorder()

		GetUserById(response, request)
		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		assert.Containsf(t, string(body), "\"image\":null}\n", "error message %s")
	})

	t.Run("returns 400 if id format is faulty", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		request, _ := http.NewRequest(http.MethodPost, "/users/hello", nil)
		//it does not go via router, so add it manually
		request.SetPathValue("id", "hello")
		response := httptest.NewRecorder()

		GetUserById(response, request)
		assert.Equal(t, 400, response.Code)
	})

	t.Run("returns 404 if user does not exist", func(t *testing.T) {
		db.SetupTestData()
		defer db.TeardownTestData()

		uuid := uuid.New()
		request, _ := http.NewRequest(http.MethodPost, "/users/"+uuid.String(), nil)
		//it does not go via router, so add it manually
		request.SetPathValue("id", uuid.String())
		response := httptest.NewRecorder()

		GetUserById(response, request)
		assert.Equal(t, 404, response.Code)
	})
}
