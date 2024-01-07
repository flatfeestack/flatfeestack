package handler

import (
	"backend/internal/db"
	"backend/pkg/util"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
	params := mux.Vars(r)
	userId := params["id"]
	convertedUserId, err := uuid.Parse(userId)

	if err != nil {
		log.Errorf("Invalid user ID: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	}

	user, err := db.FindPublicUserById(convertedUserId)

	if user == nil {
		log.Errorf("User not found %s", userId)
		util.WriteErrorf(w, http.StatusNotFound, GenericErrorMessage)
	} else if err != nil {
		log.Errorf("Could not fetch user: %v", err)
		util.WriteErrorf(w, http.StatusInternalServerError, GenericErrorMessage)
	} else {
		util.WriteJson(w, user)
	}

	return
}
