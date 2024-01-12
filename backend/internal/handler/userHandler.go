package handler

import (
	"backend/internal/db"
	"backend/pkg/util"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	prefix string
	mux    *http.ServeMux
}

const (
	GenericErrorMessage            = "Oops something went wrong. Please try again."
	RepositoryNotFoundErrorMessage = "Oops something went wrong with retrieving the repositories. Please try again."
	NotAllowedToViewMessage        = "Oops you are not allowed to view this resource."
)

func RegisterUserHandler(mux *http.ServeMux) *UserHandler {

	mux.HandleFunc("GET /users/{id}", GetUserById)

	return &UserHandler{
		mux: mux,
	}
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	convertedUserId, err := uuid.Parse(userId)

	if err != nil {
		slog.Error("Invalid user ID",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	}

	user, err := db.FindPublicUserById(convertedUserId)

	if user == nil {
		slog.Error("User not found",
			slog.String("userId", userId))
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		slog.Error("Could not fetch user",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, user)
	}

	return
}
