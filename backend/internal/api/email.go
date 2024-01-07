package api

import (
	"backend/internal/client"
	"backend/internal/db"
	"backend/pkg/util"
	"encoding/base32"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type EmailHandler struct {
	e *client.EmailClient
}

func NewEmailHandler(e *client.EmailClient) *EmailHandler {
	return &EmailHandler{e}
}

func GetMyConnectedEmails(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	emails, err := db.FindGitEmailsByUserId(user.Id)
	if err != nil {
		log.Errorf("Could not find git emails %v", err)
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with retrieving the email addresses. Please try again.")
		return
	}
	util.WriteJson(w, emails)
}

func ConfirmConnectedEmails(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	var emailToken EmailToken
	err := json.NewDecoder(r.Body).Decode(&emailToken)
	if err != nil {
		log.Errorf("Could not decode json: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = db.ConfirmGitEmail(emailToken.Email, emailToken.Token, util.TimeNow())
	if err != nil {
		log.Errorf("Invalid email: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong with confirming the git email. Please try again.")
		return
	}
}

func (e *EmailHandler) AddGitEmail(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	var body GitEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Errorf("Could not decode json: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rnd, err := util.GenRnd(20)
	if err != nil {
		log.Errorf("ERR-reset-email-02, err %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	addGitEmailToken := base32.StdEncoding.EncodeToString(rnd)
	id := uuid.New()

	c, err := db.CountExistingOrConfirmedGitEmail(user.Id, body.Email)
	if err != nil {
		log.Errorf("Could not save email: %v", err)
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}
	if c > 0 {
		log.Errorf("Could not save email, either user has entered already or is confirmed, count: %v", c)
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}

	err = db.InsertGitEmail(id, user.Id, body.Email, &addGitEmailToken, util.TimeNow())
	if err != nil {
		log.Errorf("Could not save email: %v", err)
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}

	err = e.e.SendAddGit(user.Id, body.Email, addGitEmailToken, lang(r))
	if err != nil {
		log.Errorf("Could not send add git email: %v", err)
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with sending the email. Please try again.")
	}
}

func RemoveGitEmail(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	params := mux.Vars(r)
	email := params["email"]

	err := db.DeleteGitEmail(user.Id, email)
	if err != nil {
		log.Errorf("Could not remove email, Invalid email: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, "Oops could not remove email. Please try again.")
		return
	}
	err = db.DeleteGitEmailFromUserEmailsSent(user.Id, email)
	if err != nil {
		log.Errorf("Could not remove user emails sent entry: %v", err)
		util.WriteErrorf(w, http.StatusBadRequest, "Oops could not remove email. Please try again.")
		return
	}
}
