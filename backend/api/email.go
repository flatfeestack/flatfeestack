package api

import (
	"backend/client"
	"backend/db"
	"backend/util"
	"encoding/base32"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
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
		slog.Error("Could not find git emails",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with retrieving the email addresses. Please try again.")
		return
	}
	util.WriteJson(w, emails)
}

func ConfirmConnectedEmails(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	var emailToken EmailToken
	err := json.NewDecoder(r.Body).Decode(&emailToken)
	if err != nil {
		slog.Error("Could not decode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = db.ConfirmGitEmail(emailToken.Email, emailToken.Token, util.TimeNow())
	if err != nil {
		slog.Error("Invalid email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong with confirming the git email. Please try again.")
		return
	}
}

func (e *EmailHandler) AddGitEmail(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	var body GitEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		slog.Error("Could not decode json",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rnd, err := util.GenRnd(20)
	if err != nil {
		slog.Error("Random number error",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	addGitEmailToken := base32.StdEncoding.EncodeToString(rnd)
	id := uuid.New()

	c, err := db.CountExistingOrConfirmedGitEmail(user.Id, body.Email)
	if err != nil {
		slog.Error("Could not save email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}
	if c > 0 {
		slog.Warn("Could not save email, either user has entered already or is confirmed",
			slog.Int("count", c))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}

	err = db.InsertGitEmail(id, user.Id, body.Email, &addGitEmailToken, util.TimeNow())
	if err != nil {
		slog.Error("Could not save email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops could not save email: Git Email already in use")
		return
	}

	err = e.e.SendAddGit(user.Id, body.Email, addGitEmailToken, lang(r))
	if err != nil {
		slog.Error("Could not send add git email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusInternalServerError, "Oops something went wrong with sending the email. Please try again.")
	}
}

func RemoveGitEmail(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	email := r.PathValue("email")

	err := db.DeleteGitEmail(user.Id, email)
	if err != nil {
		slog.Error("Could not remove email, invalid email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops could not remove email. Please try again.")
		return
	}
	err = db.DeleteGitEmailFromUserEmailsSent(user.Id, email)
	if err != nil {
		slog.Error("Could not remove user emails sent entry",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops could not remove email. Please try again.")
		return
	}
}
