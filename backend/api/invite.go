package api

import (
	db "backend/db"
	"backend/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"regexp"
)

//********************************************************************

func Invitations(w http.ResponseWriter, _ *http.Request, user *db.User) {
	invites, err := db.FindInvitationsByAnyEmail(user.Email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}

	oauthEnc, err := json.Marshal(invites)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-oauth-08, cannot verify refresh token %v", err)
		return
	}
	w.Write(oauthEnc)
}

func InviteByDelete(w http.ResponseWriter, r *http.Request, user *db.User) {
	//delete the invite from me of other users
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = db.DeleteInvite(email, user.Email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func InviteMyDelete(w http.ResponseWriter, r *http.Request, user *db.User) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = db.DeleteInvite(user.Email, email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ConfirmInvite(w http.ResponseWriter, r *http.Request, user *db.User) {
	m := mux.Vars(r)
	email := m["email"]

	err := db.UpdateConfirmInviteAt(email, user.Email, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}

	//either we already confirmed, or we just did so
	sponsor, err := db.FindUserByEmail(email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}

	err = db.UpdateUserInviteId(user.Id, sponsor.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}
}

func InviteOther(w http.ResponseWriter, r *http.Request, user *db.User) {
	m := mux.Vars(r)
	email := m["email"]

	err := validateEmail(email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "email address not valid %v", err)
		return
	}
	inviteId := uuid.New()
	err = db.InsertInvite(inviteId, user.Email, email, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
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
