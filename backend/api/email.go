package api

import (
	clnt "backend/clients"
	db "backend/db"
	"backend/utils"
	"encoding/base32"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

func GetMyConnectedEmails(w http.ResponseWriter, _ *http.Request, user *db.User) {
	emails, err := db.FindGitEmailsByUserId(user.Id)
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not find git emails %v", err)
		return
	}
	utils.WriteJson(w, emails)
}

func ConfirmConnectedEmails(w http.ResponseWriter, r *http.Request) {
	var emailToken EmailToken
	err := json.NewDecoder(r.Body).Decode(&emailToken)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	err = db.ConfirmGitEmail(emailToken.Email, emailToken.Token, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Invalid email: %v", err)
		return
	}
}

func AddGitEmail(w http.ResponseWriter, r *http.Request, user *db.User) {
	var body GitEmailRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	rnd, err := utils.GenRnd(20)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "ERR-reset-email-02, err %v", err)
		return
	}
	addGitEmailToken := base32.StdEncoding.EncodeToString(rnd)
	err = db.InsertGitEmail(user.Id, body.Email, &addGitEmailToken, utils.TimeNow())
	if err != nil {
		utils.WriteErrorf(w, http.StatusInternalServerError, "Could not save email: %v", err)
		return
	}

	email := url.QueryEscape(body.Email)
	clnt.SendAddGit(email, addGitEmailToken, lang(r))
}

func RemoveGitEmail(w http.ResponseWriter, r *http.Request, user *db.User) {
	params := mux.Vars(r)
	email := params["email"]

	err := db.DeleteGitEmail(user.Id, email)
	if err != nil {
		utils.WriteErrorf(w, http.StatusBadRequest, "Invalid email: %v", err)
		return
	}
}
