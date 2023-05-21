package api

import (
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
		log.Errorf("ERR-invite-06, failed to find invitations by email: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, "Oops there was a problem with retrieving the invitations. Please try again.")
		return
	}

	oauthEnc, err := json.Marshal(invites)
	if err != nil {
		log.Errorf("ERR-oauth-08, cannot verify refresh token %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	w.Write(oauthEnc)
}

func InviteByDelete(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	//delete the invite from me of other users
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		log.Errorf("ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		utils.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}
	err = db.DeleteInvite(email, user.Email)
	if err != nil {
		log.Errorf("ERR-invite-06, failed to delete invitation: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, InviteDeleteError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func InviteMyDelete(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		log.Errorf("ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		utils.WriteErrorf(w, http.StatusBadRequest, EmailEscapeError)
		return
	}
	err = db.DeleteInvite(user.Email, email)
	if err != nil {
		log.Errorf("ERR-invite-06, insert user failed: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, InviteDeleteError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ConfirmInvite(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	m := mux.Vars(r)
	email := m["email"]

	err := db.UpdateConfirmInviteAt(email, user.Email, utils.TimeNow())
	if err != nil {
		log.Errorf("cannot confirm invite: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}

	//either we already confirmed, or we just did so
	sponsor, err := db.FindUserByEmail(email)
	if err != nil {
		log.Errorf("cannot find user by email: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}

	err = db.UpdateUserInviteId(user.Id, sponsor.Id)
	if err != nil {
		log.Errorf("cannot update user invite: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, InviteConfirmError)
		return
	}
}

func InviteOther(w http.ResponseWriter, r *http.Request, user *db.UserDetail) {
	m := mux.Vars(r)
	email := m["email"]

	err := validateEmail(email)
	if err != nil {
		log.Errorf("email address not valid %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong. Please check the entered email address and try again.")
		return
	}
	inviteId := uuid.New()
	err = db.InsertInvite(inviteId, user.Email, email, utils.TimeNow())
	if err != nil {
		log.Errorf("ERR-invite-06, insert invite failed: %v", err)
		utils.WriteErrorf(w, http.StatusBadRequest, "Oops something went wrong while creating the invitation. Please try again.")
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
