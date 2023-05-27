package main

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	dbLib "github.com/flatfeestack/go-lib/database"
	mail "github.com/flatfeestack/go-lib/email"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-jose/go-jose/v3"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/text/language"
)

type Timewarp struct {
	Offset int `json:"offset"`
}

type EmailToken struct {
	Email      string `json:"email"`
	EmailToken string `json:"emailToken"`
}

type scryptParam struct {
	n   int
	r   int
	p   int
	len int
}

const (
	flowPassword   = "pwd"
	flowCode       = "code"
	flowInvitation = "inv"
)

var (
	m       = map[uint8]scryptParam{0: {16384, 8, 1, 32}}
	matcher = language.NewMatcher([]language.Tag{
		language.English,
		language.German,
	})
)

const (
	KeyInviteOld = "invite-old"
	KeyInvite    = "invite-new"
	KeyLogin     = "login"
	KeyReset     = "reset"
	KeySignup    = "signup"
)

const (
	GenericErrorMessage             = "Oops something went wrong. Please try again."
	CannotVerifyRefreshTokenMessage = "Cannot verify refresh token"
	BasicAuthFailedMessage          = "Basic auth failed"
)

func confirmEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	email := vars["email"]

	err := updateEmailToken(email, token)
	if err != nil {
		log.Errorf("ERR-confirm-email-01, update email token for %v failed, token %v: %v", email, token, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "Update Email Token failed. Please try again.")
		return
	}

	result, err := findAuthByEmail(email)
	if err != nil {
		log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	if result.flowType == flowCode {
		keys := r.URL.Query()
		uri := keys.Get("redirect_uri")
		w.Header().Set("Location", uri+"&email="+url.QueryEscape(email))
		w.WriteHeader(http.StatusSeeOther)
	} else {
		writeOAuth(w, result)
	}
}

func writeOAuth(w http.ResponseWriter, result *dbRes) {
	encodedAccessToken, encodedRefreshToken, expiresAt, err := encodeTokens(result)
	if err != nil {
		log.Errorf("cannot encode tokens %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", GenericErrorMessage)
		return
	}

	oauth := OAuth{
		AccessToken:  encodedAccessToken,
		TokenType:    "Bearer",
		RefreshToken: encodedRefreshToken,
		Expires:      strconv.FormatInt(expiresAt, 10),
	}
	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		log.Errorf("ERR-oauth-08, cannot verify refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
		return
	}
	w.Write(oauthEnc)
}

func confirmEmailPost(w http.ResponseWriter, r *http.Request) {
	var et EmailToken
	err := json.NewDecoder(r.Body).Decode(&et)
	if err != nil {
		log.Errorf("ERR-signup-01, cannot parse JSON credentials %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = updateEmailToken(et.Email, et.EmailToken)
	if err != nil {
		//the token can be only updated once. Otherwise, anyone with the link can always login. Thus, if the email
		//leaks, the account is compromised. Thus, disallow this.
		log.Errorf("ERR-confirm-email-01, update email token for %v failed, token %v: %v", et.Email, et.EmailToken, err)
		writeErr(w, http.StatusForbidden, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(et.Email)
	if err != nil {
		log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", et.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func invite(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	vars := mux.Vars(r)
	email := vars["email"]

	params := map[string]string{}
	if r.Body != nil && r.Body != http.NoBody {
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			log.Errorf("cannot decode invite %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
	}
	params["lang"] = lang(r)

	err := validateEmail(email)
	if err != nil {
		log.Errorf("email in invite is wrong %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "Invalid email. Please check and try again.")
		return
	}

	u, err := findAuthByEmail(email)

	if err != nil && err != sql.ErrNoRows {
		log.Errorf("find email %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	if u != nil {
		//user already exists, send email to direct him to the invitations
		params["url"] = opts.EmailLinkPrefix + "/user/invitations"
		sendgridRequest := mail.PrepareEmail(email, params,
			KeyInviteOld, "You have been invited by "+claims.Subject,
			"Click on this link to see your invitation: "+params["url"],
			params["lang"])
		go func() {
			request := mail.SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             opts.EmailUrl,
				EmailFromName:   opts.EmailFromName,
				EmailFrom:       opts.EmailFrom,
				EmailToken:      opts.EmailToken,
			}
			err = mail.SendEmail(request)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()

		return
	}

	emailToken, err := genToken()
	if err != nil {
		log.Errorf("cannot generate rnd token for %v, err %v", email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	refreshToken, err := genToken()
	if err != nil {
		log.Errorf("cannot generate rnd refresh token for %v, err %v", email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	params["token"] = emailToken
	params["email"] = email

	//TODO: better check if user is already in DB
	err = insertUser(email, nil, emailToken, refreshToken, flowInvitation, timeNow())
	if err != nil {
		log.Printf("could not insert user %v", err)
		params["url"] = opts.EmailLinkPrefix + "/login"

		sendgridRequest := mail.PrepareEmail(email, params,
			KeyLogin, "You have been invited again by "+claims.Subject,
			"Click on this link to login: "+params["url"],
			params["lang"])

		go func() {
			request := mail.SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             opts.EmailUrl,
				EmailFromName:   opts.EmailFromName,
				EmailFrom:       opts.EmailFrom,
				EmailToken:      opts.EmailToken,
			}
			err = mail.SendEmail(request)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()

		//do not write error, we do not want the user to know that this user does not exist (privacy)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		params["url"] = opts.EmailLinkPrefix + "/confirm/invite/" + url.QueryEscape(email) + "/" + emailToken + "/" + claims.Subject

		sendgridRequest := mail.PrepareEmail(email, params,
			KeyInvite, "You have been invited by "+claims.Subject,
			"Click on this link to create your account: "+params["url"],
			params["lang"])

		go func() {
			request := mail.SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             opts.EmailUrl,
				EmailFromName:   opts.EmailFromName,
				EmailFrom:       opts.EmailFrom,
				EmailToken:      opts.EmailToken,
			}
			err = mail.SendEmail(request)
			if err != nil {
				log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
			}
		}()

		if opts.Env == "dev" || opts.Env == "local" {
			w.Write([]byte(`{"url":"` + params["url"] + `"}`))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func confirmInvite(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		log.Errorf("ERR-signup-01, cannot parse JSON credentials %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	newPw, err := newPw(cred.Password, 0)
	if err != nil {
		log.Errorf("ERR-signup-05, key %v error: %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	err = updatePasswordInvite(cred.Email, cred.EmailToken, newPw)
	if err != nil {
		log.Errorf("ERR-confirm-reset-email-07, update user failed: %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "Update user failed.")
		return
	}

	result, err := findAuthByEmail(cred.Email)
	if err != nil {
		log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func signup(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		log.Errorf("ERR-signup-01, cannot parse JSON credentials %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = validateEmail(cred.Email)
	if err != nil {
		log.Errorf("ERR-signup-02, email is wrong %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = validatePassword(cred.Password)
	if err != nil {
		log.Errorf("ERR-signup-03, password is wrong %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	emailToken, err := genToken()
	if err != nil {
		log.Errorf("ERR-signup-04, RND %v err %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	//https://security.stackexchange.com/questions/11221/how-big-should-salt-be
	calcPw, err := newPw(cred.Password, 0)
	if err != nil {
		log.Errorf("ERR-signup-05, key %v error: %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	refreshToken, err := genToken()
	if err != nil {
		log.Errorf("ERR-signup-06, key %v error: %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	flowType := flowPassword
	urlParams := ""
	if cred.RedirectUri != "" {
		urlParams = "?redirect_uri=" + url.QueryEscape(cred.RedirectUri)
		i := strings.Index(cred.RedirectUri, "?")
		if i < len(cred.RedirectUri) && i > 0 {
			m, err := url.ParseQuery(cred.RedirectUri[strings.Index(cred.RedirectUri, "?")+1:])
			if err != nil {
				log.Errorf("ERR-signup-07, insert user failed: %v", err)
				writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
				return
			}
			if m.Get("code_challenge") != "" {
				flowType = flowCode
			}
		}
	}

	//check if user exists than was not activated yet. In that case, resend the email and don't try to insert
	//the user, as this would fail due to constraints

	err = insertUser(cred.Email, calcPw, emailToken, refreshToken, flowType, timeNow())
	if err != nil {
		log.Errorf("ERR-signup-07, insert user failed: %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	params := map[string]string{}
	params["token"] = emailToken
	params["email"] = cred.Email
	params["url"] = opts.EmailLinkPrefix + "/confirm/signup/" + url.QueryEscape(cred.Email) + "/" + emailToken + urlParams
	params["lang"] = lang(r)

	sendgridRequest := mail.PrepareEmail(cred.Email, params,
		KeySignup, "Validate your email",
		"Click on this link: "+params["url"],
		params["lang"])

	go func() {
		request := mail.SendEmailRequest{
			SendgridRequest: sendgridRequest,
			Url:             opts.EmailUrl,
			EmailFromName:   opts.EmailFromName,
			EmailFrom:       opts.EmailFrom,
			EmailToken:      opts.EmailToken,
		}
		err = mail.SendEmail(request)
		if err != nil {
			log.Printf("ERR-signup-07, send email failed: %v, %v\n", opts.EmailUrl, err)
		}
	}()

	if opts.Env == "dev" || opts.Env == "local" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"url":"` + params["url"] + `"}`))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func lang(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, accept)
	b, _ := tag.Base()
	return b.String()
}

func login(w http.ResponseWriter, r *http.Request) {
	var cred Credentials

	//https://medium.com/@xoen/golang-read-from-an-io-readwriter-without-loosing-its-content-2c6911805361
	var bodyCopy []byte
	var err error
	if r.Body != nil {
		bodyCopy, err = io.ReadAll(r.Body)
		if err != nil {
			log.Errorf("ERR-login-01, cannot parse POST data %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
	err = json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
		err = r.ParseForm()
		if err != nil {
			log.Errorf("ERR-login-01, cannot parse POST data %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		err = schema.NewDecoder().Decode(&cred, r.PostForm)
		if err != nil {
			log.Errorf("ERR-login-02, cannot populate POST data %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
	}

	result, errString, err := checkEmailPassword(cred.Email, cred.Password)
	if err != nil {
		log.Errorf("ERR-login-02 %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_client", errString, GenericErrorMessage)
		return
	}

	//SMS logic
	if result.totp != nil && result.sms != nil && result.smsVerified != nil {
		totp := newTOTP(*result.totp)
		token := totp.Now()
		if cred.TOTP == "" {
			url := strings.Replace(opts.UrlSms, "{sms}", *result.sms, 1)
			url = strings.Replace(url, "{token}", token, 1)
			err = sendSMS(url)
			if err != nil {
				log.Errorf("ERR-login-07, send sms failed %v error: %v", cred.Email, err)
				writeErr(w, http.StatusUnauthorized, "invalid_request", "blocked", "Send SMS failed")
				return
			}
			log.Errorf("ERR-login-08, waiting for sms verification: %v", cred.Email)
			writeErr(w, http.StatusLocked, "invalid_client", "blocked", GenericErrorMessage)
			return
		} else if token != cred.TOTP {
			log.Errorf("ERR-login-09, sms wrong token, %v err %v", cred.Email, err)
			writeErr(w, http.StatusForbidden, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
	}

	//TOTP logic
	if result.totp != nil && result.totpVerified != nil {
		totp := newTOTP(*result.totp)
		token := totp.Now()
		if token != cred.TOTP {
			log.Errorf("ERR-login-10, totp wrong token, %v err %v", cred.Email, err)
			writeErr(w, http.StatusForbidden, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
	}

	if cred.CodeCodeChallengeMethod != "" {
		//return the code flow
		handleCode(w, cred.Email, cred.CodeChallenge, cred.CodeCodeChallengeMethod, cred.RedirectUri, cred.RedirectAs201)
		return
	}

	refreshToken, err := resetRefreshToken(result.refreshToken)
	if err != nil {
		log.Errorf("ERR-login-15, cannot reset refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", "Cannot reset refresh token")
		return
	}
	encodedAccessToken, encodedRefreshToken, expiresAt, err := checkRefresh(cred.Email, refreshToken)
	if err != nil {
		log.Errorf("ERR-login-16, cannot verify refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
		return
	}

	oauth := OAuth{AccessToken: encodedAccessToken, TokenType: "Bearer", RefreshToken: encodedRefreshToken, Expires: strconv.FormatInt(expiresAt, 10)}
	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		log.Errorf("ERR-login-17, cannot encode refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", "Cannot encode refresh token")
		return
	}

	w.Write(oauthEnc)
}

func handleCode(w http.ResponseWriter, email string, codeChallenge string, codeChallengeMethod string, redirectUri string, redirectAs201 bool) {
	encoded, _, err := encodeCodeToken(email, codeChallenge, codeChallengeMethod)
	if err != nil {
		log.Errorf("ERR-login-14, cannot set refresh token for %v, %v", email, err)
		writeErr(w, http.StatusInternalServerError, "invalid_request", "blocked", "Cannot set refresh token")
		return
	}
	w.Header().Set("Location", redirectUri+"?code="+encoded)
	if redirectAs201 {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusSeeOther)
	}
}

func displayEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		email = fmt.Sprintf("email decoding error %v", err)
		log.Printf(email)
	}
	action, err := url.QueryUnescape(vars["action"])
	if err != nil {
		action = fmt.Sprintf("action decoding error %v", err)
		log.Printf(action)
	}

	if action == "signup" {
		fmt.Printf("go to URL: http://%s/confirm/signup/%s/%s\n", r.Host, email, token)
	} else if action == "reset" {
		fmt.Printf("go to URL: http://%s/confirm/reset/%s/%s\n", r.Host, email, token)
	}

	w.WriteHeader(http.StatusOK)
}

func displaySMS(_ http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	sms, err := url.QueryUnescape(vars["sms"])
	if err != nil {
		log.Printf("decoding error %v", err)
	}
	fmt.Printf("Send token [%s] to NR %s\n", token, sms)
}

func resetEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		log.Errorf("ERR-confirm-reset-email-01, query unescape email %v err: %v", vars["email"], err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	forgetEmailToken, err := genToken()
	if err != nil {
		log.Errorf("ERR-reset-email-02, RND %v err %v", email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = updateEmailForgotToken(email, forgetEmailToken)
	if err != nil {
		log.Errorf("ERR-reset-email-03, update token for %v failed, token %v: %v", email, forgetEmailToken, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	params := map[string]string{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Printf("No or wrong json, ignoring [%v]", err)
	}

	params["email"] = email
	params["url"] = opts.EmailLinkPrefix + "/confirm/reset/" + email + "/" + forgetEmailToken
	params["lang"] = lang(r)

	sendgridRequest := mail.PrepareEmail(email, params,
		KeyReset, "Reset your email",
		"Click on this link: "+params["url"],
		params["lang"])

	go func() {
		request := mail.SendEmailRequest{
			SendgridRequest: sendgridRequest,
			Url:             opts.EmailUrl,
			EmailFromName:   opts.EmailFromName,
			EmailFrom:       opts.EmailFrom,
			EmailToken:      opts.EmailToken,
		}
		err = mail.SendEmail(request)
		if err != nil {
			log.Printf("ERR-reset-email-04, send email failed: %v", opts.EmailUrl)
		}
	}()

	w.WriteHeader(http.StatusOK)
}

func newPw(password string, version uint8) ([]byte, error) {
	salt, err := genRnd(16) //salt is always 128bit
	if err != nil {
		return nil, err
	}

	calcPw, err := scrypt.Key([]byte(password), salt, m[version].n, m[version].r, m[version].p, m[version].len)
	if err != nil {
		return nil, err
	}

	ret := []byte{version}
	ret = append(ret, salt...)
	ret = append(ret, calcPw...)
	return ret, nil
}

func checkPw(checkPw string, encodedPw []byte) ([]byte, []byte, error) {
	key := encodedPw[0]
	salt := encodedPw[1:17] //salt is always 128bit
	storedPw := encodedPw[17 : 17+m[key].len]
	calcPw, err := scrypt.Key([]byte(checkPw), salt, m[key].n, m[key].r, m[key].p, m[key].len)
	return storedPw, calcPw, err
}

func confirmReset(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		log.Errorf("ERR-signup-01, cannot parse JSON credentials %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	newPw, err := newPw(cred.Password, 0)
	if err != nil {
		log.Errorf("ERR-signup-05, key %v error: %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	err = updatePasswordForgot(cred.Email, cred.EmailToken, newPw)
	if err != nil {
		log.Errorf("ERR-confirm-reset-email-07, update user failed: %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "Update user failed.")
		return
	}

	result, err := findAuthByEmail(cred.Email)
	if err != nil {
		log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", cred.Email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func setupTOTP(w http.ResponseWriter, _ *http.Request, claims *TokenClaims) {
	secret, err := genToken()
	if err != nil {
		log.Errorf("ERR-setup-totp-01, RND %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = updateTOTP(claims.Subject, secret)
	if err != nil {
		log.Errorf("ERR-setup-totp-02, update failed %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	totp := newTOTP(secret)
	p := ProvisioningUri{}
	p.Uri = totp.ProvisioningUri(claims.Subject, opts.Issuer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func confirmTOTP(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	vars := mux.Vars(r)
	token, err := url.QueryUnescape(vars["token"])
	if err != nil {
		log.Errorf("ERR-confirm-totp-01, query unescape token %v err: %v", vars["token"], err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(claims.Subject)
	if err != nil {
		log.Errorf("ERR-confirm-totp-02, DB select, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	totp := newTOTP(*result.totp)
	if token != totp.Now() {
		log.Errorf("ERR-confirm-totp-03, token different, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	err = updateTOTPVerified(claims.Subject, timeNow())
	if err != nil {
		log.Errorf("ERR-confirm-totp-04, DB select, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func setupSMS(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	vars := mux.Vars(r)
	sms, err := url.QueryUnescape(vars["sms"])
	if err != nil {
		log.Errorf("ERR-setup-sms-01, query unescape sms %v err: %v", vars["sms"], err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	secret, err := genToken()
	if err != nil {
		log.Errorf("ERR-setup-sms-02, RND %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	err = updateSMS(claims.Subject, secret, sms)
	if err != nil {
		log.Errorf("ERR-setup-sms-03, updateSMS failed %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	totp := newTOTP(secret)

	url := strings.Replace(opts.UrlSms, "{sms}", sms, 1)
	url = strings.Replace(url, "{token}", totp.Now(), 1)

	err = sendSMS(url)
	if err != nil {
		log.Errorf("ERR-setup-sms-04, send SMS failed %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func confirmSMS(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	vars := mux.Vars(r)
	token, err := url.QueryUnescape(vars["token"])
	if err != nil {
		log.Errorf("ERR-confirm-sms-01, query unescape token %v err: %v", vars["token"], err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(claims.Subject)
	if err != nil {
		log.Errorf("ERR-confirm-sms-02, DB select, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	totp := newTOTP(*result.totp)
	if token != totp.Now() {
		log.Errorf("ERR-confirm-sms-03, token different, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusUnauthorized, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	err = updateSMSVerified(claims.Subject, timeNow())
	if err != nil {
		log.Errorf("ERR-confirm-sms-04, update sms failed, %v err %v", claims.Subject, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func readiness(w http.ResponseWriter, _ *http.Request) {
	err := dbLib.DB.Ping()
	if err != nil {
		log.Printf(fmt.Sprintf("not ready: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func jwkFunc(w http.ResponseWriter, _ *http.Request) {
	json := []byte(`{"keys":[`)
	if privRSA != nil {
		k := jose.JSONWebKey{Key: privRSA.Public()}
		kid, err := k.Thumbprint(crypto.SHA256)
		if err != nil {
			log.Errorf("ERR-jwk-1, %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		k.KeyID = hex.EncodeToString(kid)
		mj, err := k.MarshalJSON()
		if err != nil {
			log.Errorf("ERR-jwk-2, %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		json = append(json, mj...)
	}
	if privEdDSA != nil {
		k := jose.JSONWebKey{Key: privEdDSA.Public()}
		mj, err := k.MarshalJSON()
		if err != nil {
			log.Errorf("ERR-jwk-3, %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		json = append(json, []byte(`,`)...)
		json = append(json, mj...)
	}
	json = append(json, []byte(`]}`)...)

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

//****************** OAuth

func refresh(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-type")
	var refreshToken string
	var err error
	if strings.Contains(contentType, "application/json") {
		refreshToken, err = paramJson("refresh_token", r)
	} else {
		refreshToken, err = param("refresh_token", r)
	}
	if err != nil {
		log.Errorf("ERR-oauth-01, basic auth failed")
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
		return
	}
	if refreshToken == "" {
		log.Errorf("ERR-oauth-02, no refresh token")
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", "No refresh token")
		return
	}

	refreshClaims, err := checkRefreshToken(refreshToken)
	if err != nil {
		log.Errorf("ERR-oauth-03, cannot verify refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
		return
	}

	encodedAccessToken, encodedRefreshToken, expiresAt, err := checkRefresh(refreshClaims.Subject, refreshClaims.Token)
	if err != nil {
		log.Errorf("ERR-oauth-03, cannot verify refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
		return
	}

	oauth := OAuth{AccessToken: encodedAccessToken, TokenType: "Bearer", RefreshToken: encodedRefreshToken, Expires: strconv.FormatInt(expiresAt, 10)}
	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		log.Errorf("ERR-oauth-04, cannot verify refresh token %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
		return
	}
	w.Write(oauthEnc)
}

func oauth(w http.ResponseWriter, r *http.Request) {
	grantType, err := param("grant_type", r)
	if err != nil {
		log.Errorf("ERR-oauth-01, basic auth failed")
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
		return
	}

	switch grantType {
	case "refresh_token":
		refresh(w, r)
	case "client_credentials":
		user, err := basicAuth(r)
		if err != nil {
			log.Errorf("Basic auth failed: %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
			return
		}

		encodedAccessToken, err := encodeAccessToken(user, opts.Scope, opts.Audience, opts.Issuer, nil)
		if err != nil {
			log.Errorf("Basic auth failed: %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
			return
		}

		oauth := OAuthSystem{
			AccessToken: encodedAccessToken,
			TokenType:   "Bearer",
		}
		oauthEnc, err := json.Marshal(oauth)
		if err != nil {
			log.Errorf("ERR-oauth-08, cannot verify refresh token %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_grant", "blocked", CannotVerifyRefreshTokenMessage)
			return
		}
		w.Write(oauthEnc)

	case "authorization_code":
		code, err := param("code", r)
		if err != nil {
			log.Errorf("ERR-oauth-01, basic auth failed")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
			return
		}
		codeVerifier, err := param("code_verifier", r)
		if err != nil {
			log.Errorf("ERR-oauth-01, basic auth failed")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", BasicAuthFailedMessage)
			return
		}
		//https://tools.ietf.org/html/rfc7636#section-4.1 length must be <= 43 <= 128
		if len(codeVerifier) < 43 {
			log.Errorf("ERR-oauth-01, min 43 chars")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		if len(codeVerifier) > 128 {
			log.Errorf("ERR-oauth-01, max 128 chars")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		cc, err := checkCodeToken(code)
		if err != nil {
			log.Errorf("ERR-oauth-04, code check failed: %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		if cc.CodeCodeChallengeMethod == "S256" {
			h := sha256.Sum256([]byte(codeVerifier))
			s := base64.RawURLEncoding.EncodeToString(h[:])
			if cc.CodeChallenge != s {
				log.Errorf("ERR-oauth-04, auth challenge failed")
				writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
				return
			}
		} else {
			log.Errorf("ERR-oauth-04, only S256 supported")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}

		result, err := findAuthByEmail(cc.Subject)
		if err != nil {
			log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", cc.Subject, err)
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}

		writeOAuth(w, result)
	case "password":
		if !opts.PasswordFlow {
			log.Errorf("ERR-oauth-05a, no username")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		email, err := param("username", r)
		if err != nil {
			log.Errorf("ERR-oauth-05a, no username")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		password, err := param("password", r)
		if err != nil {
			log.Errorf("ERR-oauth-05b, no password")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		scope, err := param("scope", r)
		if err != nil {
			log.Errorf("ERR-oauth-05c, no scope")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}
		if email == "" || password == "" || scope == "" {
			log.Errorf("ERR-oauth-05, username, password, or scope empty")
			writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
			return
		}

		result, errString, err := checkEmailPassword(email, password)
		if err != nil {
			log.Errorf("ERR-oauth-06 %v", err)
			writeErr(w, http.StatusBadRequest, "invalid_grant", errString, GenericErrorMessage)
			return
		}

		writeOAuth(w, result)
	default:
		log.Errorf("ERR-oauth-09, unsupported grant type")
		writeErr(w, http.StatusBadRequest, "unsupported_grant_type", "blocked", GenericErrorMessage)
	}
}

// https://tools.ietf.org/html/rfc6749#section-1.3.1
// https://developer.okta.com/blog/2019/08/22/okta-authjs-pkce
func authorize(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	rt := keys.Get("response_type")
	email := keys.Get("email")
	if rt == flowCode && email != "" {
		handleCode(w, keys.Get("email"),
			keys.Get("code_challenge"),
			keys.Get("code_challenge_method"),
			keys.Get("redirect_uri"), false)
	} else {
		http.ServeFile(w, r, "login.html")
	}
}

func revoke(w http.ResponseWriter, r *http.Request) {
	tokenHint, err := param("token_type_hint", r)
	if err != nil {
		log.Errorf("ERR-oauth-07, unsupported grant type")
		writeErr(w, http.StatusBadRequest, "unsupported_grant_type1", "blocked", GenericErrorMessage)
		return
	}
	if tokenHint == "refresh_token" {
		oldToken, err := param("token", r)
		if err != nil {
			log.Errorf("ERR-oauth-07, unsupported grant type")
			writeErr(w, http.StatusBadRequest, "unsupported_grant_type1", "blocked", GenericErrorMessage)
			return
		}
		if oldToken == "" {
			log.Errorf("ERR-oauth-07, unsupported grant type")
			writeErr(w, http.StatusBadRequest, "unsupported_grant_type1", "blocked", GenericErrorMessage)
			return
		}
		_, err = resetRefreshToken(oldToken)
		if err != nil {
			log.Errorf("ERR-oauth-07, unsupported grant type")
			writeErr(w, http.StatusBadRequest, "unsupported_grant_type2", "blocked", GenericErrorMessage)
			return
		}
	} else {
		log.Errorf("ERR-oauth-07, unsupported grant type")
		writeErr(w, http.StatusBadRequest, "unsupported_grant_type", "blocked", GenericErrorMessage)
		return
	}
}

func logout(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	keys := r.URL.Query()
	redirectUri := keys.Get("redirect_uri")

	result, err := findAuthByEmail(claims.Subject)
	if err != nil {
		log.Errorf("ERR-oauth-06 %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", GenericErrorMessage)
		return
	}

	refreshToken := result.refreshToken
	_, err = resetRefreshToken(refreshToken)
	if err != nil {
		log.Errorf("ERR-oauth-07, unsupported grant type: %v", err)
		writeErr(w, http.StatusBadRequest, "unsupported_grant_type", "blocked", GenericErrorMessage)
		return
	}

	if redirectUri != "" {
		w.Header().Set("Location", redirectUri)
		w.WriteHeader(http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func serverTime(w http.ResponseWriter, r *http.Request, email string) {
	currentTime := timeNow()
	writeJsonStr(w, `{"time":"`+currentTime.Format("2006-01-02 15:04:05")+`","offset":`+strconv.Itoa(secondsAdd)+`}`)
}

func timeWarp(w http.ResponseWriter, r *http.Request, adminEmail string) {
	m := mux.Vars(r)
	h := m["hours"]
	if h == "" {
		log.Errorf("ERR-timewarp-01 %v", m)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", GenericErrorMessage)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		log.Errorf("ERR-timewarp-02 %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", GenericErrorMessage)
		return
	}

	seconds := hours * 60 * 60
	secondsAdd += seconds
	log.Printf("time warp: %v", timeNow())

	//since we warp, the token will be invalid
	result, err := findAuthByEmail(adminEmail)
	if err != nil {
		log.Errorf("ERR-timeWarp, findAuthByEmail for %v failed, %v", adminEmail, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	writeOAuth(w, result)
}

func asUser(w http.ResponseWriter, r *http.Request, _ string) {
	m := mux.Vars(r)
	email := m["email"]
	result, err := findAuthByEmail(email)
	if err != nil {
		log.Errorf("ERR-writeOAuth, findAuthByEmail for %v failed, %v", email, err)
		writeErr(w, http.StatusBadRequest, "invalid_request", "blocked", GenericErrorMessage)
		return
	}
	writeOAuth(w, result)
}

func deleteUser(w http.ResponseWriter, r *http.Request, admin string) {
	m := mux.Vars(r)
	email := m["email"]
	err := deleteDbUser(email)
	if err != nil {
		log.Errorf("could not delete user %v, requested by %s", err, admin)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", "Could not delete user")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateUser(w http.ResponseWriter, r *http.Request, admin string) {
	//now we update the meta data that comes as system meta data. Thus we trust the system to provide the correct metadata, not the user
	m := mux.Vars(r)
	email := m["email"]
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("could not update user %v, requested by %s", err, admin)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", "Could not update user")
		return
	}
	if !json.Valid(b) {
		log.Errorf("invalid json [%s], requested by %s", string(b), admin)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", "Invalid JSON")
		return
	}
	err = updateSystemMeta(email, string(b))
	if err != nil {
		log.Errorf("could not update system meta %v, requested by %s", err, admin)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", GenericErrorMessage)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJsonStr(w http.ResponseWriter, obj string) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(obj))
	if err != nil {
		log.Errorf("Could write json: %v", err)
		writeErr(w, http.StatusBadRequest, "invalid_grant", "not-found", GenericErrorMessage)
	}
}
