package api

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserById(t *testing.T) {
	t.Run("should return user", func(t *testing.T) {
		setup()
		defer teardown()

		userDetail := insertTestUser(t, "hello@world.com")
		request, _ := http.NewRequest(http.MethodPost, "/users/"+userDetail.Id.String(), nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"id": userDetail.Id.String(),
		}
		request = mux.SetURLVars(request, vars)

		GetUserById(response, request)

		assert.Equal(t, 200, response.Code)
		body, _ := io.ReadAll(response.Body)
		assert.Containsf(t, string(body), "\"image\":null}\n", "error message %s")
	})

	t.Run("returns 400 if id format is faulty", func(t *testing.T) {
		setup()
		defer teardown()

		request, _ := http.NewRequest(http.MethodPost, "/users/hello", nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"id": "hello",
		}
		request = mux.SetURLVars(request, vars)

		GetUserById(response, request)

		assert.Equal(t, 400, response.Code)
	})

	t.Run("returns 404 if user does not exist", func(t *testing.T) {
		setup()
		defer teardown()

		uuid := uuid.New()

		request, _ := http.NewRequest(http.MethodPost, "/users/"+uuid.String(), nil)
		response := httptest.NewRecorder()
		vars := map[string]string{
			"id": uuid.String(),
		}
		request = mux.SetURLVars(request, vars)

		GetUserById(response, request)

		assert.Equal(t, 404, response.Code)
	})
}
