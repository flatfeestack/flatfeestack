package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
)

//********************************************************************

func invitations(w http.ResponseWriter, _ *http.Request, user *User) {
	invites, err := findInvitationsByEmail(user.Email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-invite-06, insert user failed: %v", err)
		return
	}

	oauthEnc, err := json.Marshal(invites)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", "ERR-oauth-08, cannot verify refresh token %v", err)
		return
	}
	w.Write(oauthEnc)
}

func inviteOtherDelete(w http.ResponseWriter, r *http.Request, user *User) {
	//delete the invite from me of other users
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = deleteInvite(email, user.Email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func inviteMyDelete(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		return
	}
	err = deleteInvite(user.Email, email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-invite-06, insert user failed: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func confirmInvite(w http.ResponseWriter, r *http.Request, user *User) {
	m := mux.Vars(r)
	email := m["email"]

	err := updateConfirmInviteAt(email, user.Email, timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-confirm-invite-03, update user failed: %v", err)
		return
	}
}

func inviteOther(w http.ResponseWriter, r *http.Request, user *User) {
	m := mux.Vars(r)
	email := m["email"]
	freqStr := m["freq"]
	freq, err := strconv.Atoi(freqStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "freq is not a number %v", err)
		return
	}

	err = validateEmail(email)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "email address not valid %v", err)
		return
	}

	err = insertInvite(user.Email, email, int64(freq), timeNow())
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "ERR-invite-06, insert user failed: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
