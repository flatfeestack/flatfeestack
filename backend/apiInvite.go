package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

//********************************************************************

func invitations(w http.ResponseWriter, _ *http.Request, user *User) {
	invites, err := findInvitationsByAnyEmail(user.Email)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}

	oauthEnc, err := json.Marshal(invites)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-oauth-08, cannot verify refresh token %v", err)
		return
	}
	w.Write(oauthEnc)
}

func inviteByDelete(w http.ResponseWriter, r *http.Request, user *User) {
	//delete the invite from me of other users
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = deleteInvite(email, user.Email)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func inviteMyDelete(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = deleteInvite(user.Email, email)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func confirmInvite(w http.ResponseWriter, r *http.Request, user *User) {
	m := mux.Vars(r)
	email := m["email"]

	err := updateConfirmInviteAt(email, user.Email, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}

	//either we already confirmed, or we just did so
	sponsor, err := findUserByEmail(email)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}

	err = updateUserInviteId(user.Id, sponsor.Id)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "cannot confirm invite: %v", err)
		return
	}
}

func inviteOther(w http.ResponseWriter, r *http.Request, user *User) {
	m := mux.Vars(r)
	email := m["email"]

	err := validateEmail(email)
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "email address not valid %v", err)
		return
	}

	err = insertInvite(user.Email, email, timeNow())
	if err != nil {
		writeErrorf(w, http.StatusBadRequest, "ERR-invite-06, insert user failed: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
