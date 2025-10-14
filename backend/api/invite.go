package api

import (
	"backend/db"
	"backend/util"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
)

// ********************************************************************
const (
	EmailEscapeError   = "Oops something went wrong with escaping the email. Please try again."
	InviteDeleteError  = "Oops something went wrong with deleting the invitation. Please try again."
	InviteConfirmError = "Oops something went wrong with confirming the invite. Please try again."
)

func Invitations(w http.ResponseWriter, _ *http.Request, user *db.UserDetail) {
	invites, err := db.FindInvitationsByAnyEmail(user.Email)
	if err != nil {
		slog.Error("Failed to find invitations by email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops there was a problem with retrieving the invitations. Please try again.")
		return
	}

	oauthEnc, err := json.Marshal(invites)
	if err != nil {
		slog.Error("Cannot verify refresh token",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	w.Write(oauthEnc)
}

func InviteByDelete(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	//delete the invite from me of other users
	emailEsc := r.PathValue("email")
	email, err := url.QueryUnescape(emailEsc)
	if err != nil {
		slog.Error("Query unescape invite-by email",
			slog.String("email", emailEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}
	err = db.DeleteInvite(email, user.Email)
	if err != nil {
		slog.Error("Failed to delete invitation",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, InviteDeleteError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func InviteMyDelete(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	emailEsc := r.PathValue("email")
	email, err := url.QueryUnescape(emailEsc)
	if err != nil {
		slog.Error("Query unescape invite-my email",
			slog.String("email", emailEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}
	err = db.DeleteInvite(user.Email, email)
	if err != nil {
		slog.Error("Insert user failed",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, InviteDeleteError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ConfirmInvite(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	emailEsc := r.PathValue("email")
	email, err := url.QueryUnescape(emailEsc)
	if err != nil {
		slog.Error("Query unescape invite-my email",
			slog.String("email", emailEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}

	err = db.UpdateConfirmInviteAt(email, user.Email, util.TimeNow())
	if err != nil {
		slog.Error("Cannot confirm invite",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}

	//either we already confirmed, or we just did so
	sponsor, err := db.FindUserByEmail(email)
	if err != nil {
		slog.Error("Cannot find user by email",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}

	err = db.UpdateUserInviteId(user.Id, sponsor.Id)
	if err != nil {
		slog.Error("Cannot update user invite",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}
}

func InviteOther(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	emailEsc := r.PathValue("email")
	email, err := url.QueryUnescape(emailEsc)
	if err != nil {
		slog.Error("Query unescape invite-my email",
			slog.String("email", emailEsc),
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}

	if email == user.Email {
		slog.Error("User tried to invite themselves, not possible",
			slog.String("email", user.Email))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong. You aren't able to invite yourself.")
		return
	}

	err = validateEmail(email)
	if err != nil {
		slog.Error("Email address not valid",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong. Please check the entered email address and try again.")
		return
	}
	inviteId := uuid.New()
	err = db.InsertInvite(inviteId, user.Email, email, util.TimeNow())
	if err != nil {
		slog.Error("Insert invite failed",
			slog.Any("error", err))
		util.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong while creating the invitation. Please try again.")
		return
	}
}

func validateEmail(email string) error {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(email) > 254 || !rxEmail.MatchString(email) {
		return fmt.Errorf("[%s] is not a valid email address", email)
	}
	return nil
}
