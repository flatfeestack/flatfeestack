package main

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-jose/go-jose/v3"
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
	NotAllowedToViewMessage         = "Oops you are not allowed to view this resource."
)

func confirmEmail(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	email := r.PathValue("email")

	err := updateEmailToken(email, token)
	if err != nil {
		slog.Error("Update email token failed",
			slog.String("email", email),
			slog.String("token", token),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Update Email Token failed. Please try again.")
		return
	}

	result, err := findAuthByEmail(email)
	if err != nil {
		slog.Error("FindAuthByEmail failed",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
		slog.Error("cannot encode tokens", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	oauth := OAuth{
		AccessToken:  encodedAccessToken,
		RefreshToken: encodedRefreshToken,
		ExpiresAt:    strconv.FormatInt(expiresAt, 10),
	}

	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}
	writeJsonBytes(w, oauthEnc)
}

func confirmEmailPost(w http.ResponseWriter, r *http.Request) {
	var et EmailToken
	err := json.NewDecoder(r.Body).Decode(&et)
	if err != nil {
		slog.Error("cannot parse JSON credentials", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = updateEmailToken(et.Email, et.EmailToken)
	if err != nil {
		//the token can be only updated once. Otherwise, anyone with the link can always login. Thus, if the email
		//leaks, the account is compromised. Thus, disallow this.
		slog.Error("update email token failed",
			slog.String("email", et.Email),
			slog.String("token", et.EmailToken),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusForbidden, GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(et.Email)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("email", et.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func invite(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	email := r.PathValue("email")

	params := map[string]string{}
	if r.Body != nil && r.Body != http.NoBody {
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			slog.Error("cannot decode invite", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}
	params["lang"] = lang(r)

	err := validateEmail(email)
	if err != nil {
		slog.Error("email in invite is wrong", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Invalid email. Please check and try again.")
		return
	}

	u, err := findAuthByEmail(email)

	if err != nil && err != sql.ErrNoRows {
		slog.Error("find email failed", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	if u != nil {
		//user already exists, send email to direct him to the invitations
		params["url"] = cfg.EmailLinkPrefix + "/user/invitations"
		sendgridRequest := PrepareEmail(email, params,
			KeyInviteOld, "You have been invited by "+claims.Subject,
			"Click on this link to see your invitation: "+params["url"],
			params["lang"])
		go func() {
			request := SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             cfg.EmailUrl,
				EmailFromName:   cfg.EmailFromName,
				EmailFrom:       cfg.EmailFrom,
				EmailToken:      cfg.EmailToken,
			}
			err = SendEmail(request)
			if err != nil {
				slog.Info("send email failed",
					slog.String("emailUrl", cfg.EmailUrl),
					slog.Any("error", err))
			}
		}()

		return
	}

	emailToken, err := genToken()
	if err != nil {
		slog.Error("cannot generate rnd token",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	refreshToken, err := genToken()
	if err != nil {
		slog.Error("cannot generate rnd refresh token",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	params["token"] = emailToken
	params["email"] = email

	err = insertUser(email, nil, emailToken, refreshToken, flowInvitation, timeNow())
	if err != nil {
		slog.Info("could not insert user", slog.Any("error", err))
		params["url"] = cfg.EmailLinkPrefix + "/login"

		sendgridRequest := PrepareEmail(email, params,
			KeyLogin, "You have been invited again by "+claims.Subject,
			"Click on this link to login: "+params["url"],
			params["lang"])

		go func() {
			request := SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             cfg.EmailUrl,
				EmailFromName:   cfg.EmailFromName,
				EmailFrom:       cfg.EmailFrom,
				EmailToken:      cfg.EmailToken,
			}
			err = SendEmail(request)
			if err != nil {
				slog.Info("send email failed",
					slog.String("emailUrl", cfg.EmailUrl),
					slog.Any("error", err))
			}
		}()

		//do not write error, we do not want the user to know that this user does not exist (privacy)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		params["url"] = cfg.EmailLinkPrefix + "/confirm/invite/" + url.QueryEscape(email) + "/" + emailToken + "/" + claims.Subject

		sendgridRequest := PrepareEmail(email, params,
			KeyInvite, "You have been invited by "+claims.Subject,
			"Click on this link to create your account: "+params["url"],
			params["lang"])

		go func() {
			request := SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             cfg.EmailUrl,
				EmailFromName:   cfg.EmailFromName,
				EmailFrom:       cfg.EmailFrom,
				EmailToken:      cfg.EmailToken,
			}
			err = SendEmail(request)
			if err != nil {
				slog.Info("send email failed",
					slog.String("emailUrl", cfg.EmailUrl),
					slog.Any("error", err))
			}
		}()

		if cfg.Env == "dev" || cfg.Env == "local" {
			dev := `{"url":"` + params["url"] + `"}`
			writeJsonBytes(w, []byte(dev))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func confirmInvite(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		slog.Error("cannot parse JSON credentials", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	newPw, err := newPw(cred.Password, 0)
	if err != nil {
		slog.Error("key error",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	err = updatePasswordInvite(cred.Email, cred.EmailToken, newPw)
	if err != nil {
		slog.Error("update user failed", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Update user failed.")
		return
	}

	result, err := findAuthByEmail(cred.Email)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func signup(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		slog.Error("cannot parse JSON credentials", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = validateEmail(cred.Email)
	if err != nil {
		slog.Error("email is wrong", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = validatePassword(cred.Password)
	if err != nil {
		slog.Error("password is wrong", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	emailToken, err := genToken()
	if err != nil {
		slog.Error("RND error",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	//https://security.stackexchange.com/questions/11221/how-big-should-salt-be
	calcPw, err := newPw(cred.Password, 0)
	if err != nil {
		slog.Error("key error",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	refreshToken, err := genToken()
	if err != nil {
		slog.Error("key error",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
				slog.Error("insert user failed", slog.Any("error", err))
				WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
		slog.Error("insert user failed", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	params := map[string]string{}
	params["token"] = emailToken
	params["email"] = cred.Email
	params["url"] = cfg.EmailLinkPrefix + "/confirm/signup/" + url.QueryEscape(cred.Email) + "/" + emailToken + urlParams
	params["lang"] = lang(r)

	sendgridRequest := PrepareEmail(cred.Email, params,
		KeySignup, "Validate your email",
		"Click on this link: "+params["url"],
		params["lang"])

	go func() {
		request := SendEmailRequest{
			SendgridRequest: sendgridRequest,
			Url:             cfg.EmailUrl,
			EmailFromName:   cfg.EmailFromName,
			EmailFrom:       cfg.EmailFrom,
			EmailToken:      cfg.EmailToken,
		}
		err = SendEmail(request)
		if err != nil {
			slog.Info("send email failed",
				slog.String("emailUrl", cfg.EmailUrl),
				slog.Any("error", err))
		}
	}()

	if cfg.Env == "dev" || cfg.Env == "local" {
		dev := `{"url":"` + params["url"] + `"}`
		writeJsonBytes(w, []byte(dev))
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
			slog.Error("cannot parse POST data", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
	err = json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
		err = r.ParseForm()
		if err != nil {
			slog.Error("cannot parse POST data", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		cred.Email = r.PostForm.Get("email")
		cred.Password = r.PostForm.Get("password")
	}

	result, errString, err := checkEmailPassword(cred.Email, cred.Password)
	if err != nil {
		slog.Error("login error",
			slog.Any("error", err),
			slog.String("errorString", errString))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	//TOTP logic
	if result.totp != nil && result.totpVerified != nil {
		totp := newTOTP(*result.totp)
		token := totp.Now()
		if token != cred.TOTP {
			slog.Error("totp wrong token",
				slog.String("email", cred.Email),
				slog.Any("error", err))
			WriteErrorf(w, http.StatusForbidden, GenericErrorMessage)
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
		slog.Error("cannot reset refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Cannot reset refresh token")
		return
	}
	encodedAccessToken, encodedRefreshToken, expiresAt, err := checkRefresh(cred.Email, refreshToken)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}

	oauth := OAuth{
		AccessToken:  encodedAccessToken,
		RefreshToken: encodedRefreshToken,
		ExpiresAt:    strconv.FormatInt(expiresAt, 10)}

	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		slog.Error("cannot encode refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Cannot encode refresh token")
		return
	}

	writeJsonBytes(w, oauthEnc)
}

func handleCode(w http.ResponseWriter, email string, codeChallenge string, codeChallengeMethod string, redirectUri string, redirectAs201 bool) {
	encoded, _, err := encodeCodeToken(email, codeChallenge, codeChallengeMethod)
	if err != nil {
		slog.Error("cannot set refresh token",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusInternalServerError, "Cannot set refresh token")
		return
	}
	w.Header().Set("Location", redirectUri+"?code="+encoded)
	if redirectAs201 {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusSeeOther)
	}
}

func resetEmail(w http.ResponseWriter, r *http.Request) {
	vars := r.PathValue("email")
	email, err := url.QueryUnescape(vars)
	if err != nil {
		slog.Error("query unescape email failed",
			slog.String("email", vars),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	forgetEmailToken, err := genToken()
	if err != nil {
		slog.Error("RND error",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = updateEmailForgotToken(email, forgetEmailToken)
	if err != nil {
		slog.Error("update token failed",
			slog.String("email", email),
			slog.String("token", forgetEmailToken),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	params := map[string]string{}
	params["email"] = email
	params["url"] = cfg.EmailLinkPrefix + "/confirm/reset/" + email + "/" + forgetEmailToken
	params["lang"] = lang(r)

	sendgridRequest := PrepareEmail(email, params,
		KeyReset, "Reset your email",
		"Click on this link: "+params["url"],
		params["lang"])

	go func() {
		request := SendEmailRequest{
			SendgridRequest: sendgridRequest,
			Url:             cfg.EmailUrl,
			EmailFromName:   cfg.EmailFromName,
			EmailFrom:       cfg.EmailFrom,
			EmailToken:      cfg.EmailToken,
		}
		err = SendEmail(request)
		if err != nil {
			slog.Info("send email failed",
				slog.String("emailUrl", cfg.EmailUrl))
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
		slog.Error("cannot parse JSON credentials", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	newPw, err := newPw(cred.Password, 0)
	if err != nil {
		slog.Error("key error",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	err = updatePasswordForgot(cred.Email, cred.EmailToken, newPw)
	if err != nil {
		slog.Error("update user failed", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Update user failed.")
		return
	}

	result, err := findAuthByEmail(cred.Email)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("email", cred.Email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func setupTOTP(w http.ResponseWriter, _ *http.Request, claims *TokenClaims) {
	secret, err := genToken()
	if err != nil {
		slog.Error("RND error",
			slog.String("subject", claims.Subject),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = updateTOTP(claims.Subject, secret)
	if err != nil {
		slog.Error("update failed",
			slog.String("subject", claims.Subject),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	totp := newTOTP(secret)
	p := ProvisioningUri{}
	p.Uri = totp.ProvisioningUri(claims.Subject, cfg.Issuer)

	w.Header().Set("Content-Type", "application/json")

	pStr, err := json.Marshal(p)
	if err != nil {
		slog.Error("cannot encode refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, "Cannot encode refresh token")
		return
	}
	writeJsonBytes(w, pStr)
}

func confirmTOTP(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	token := r.PathValue("token")
	token, err := url.QueryUnescape(token)
	if err != nil {
		slog.Error("query unescape token failed",
			slog.String("token", token),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(claims.Subject)
	if err != nil {
		slog.Error("DB select failed",
			slog.String("subject", claims.Subject),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	totp := newTOTP(*result.totp)
	if token != totp.Now() {
		slog.Error("token different",
			slog.String("subject", claims.Subject),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	err = updateTOTPVerified(claims.Subject, timeNow())
	if err != nil {
		slog.Error("DB select failed",
			slog.String("subject", claims.Subject),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func readiness(w http.ResponseWriter, _ *http.Request) {
	err := DB.Ping()
	if err != nil {
		slog.Warn("not ready", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func jwkFunc(w http.ResponseWriter, _ *http.Request) {
	j := []byte(`{"keys":[`)
	if privRSA != nil {
		k := jose.JSONWebKey{Key: privRSA.Public()}
		kid, err := k.Thumbprint(crypto.SHA256)
		if err != nil {
			slog.Error("jwk thumb", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		k.KeyID = hex.EncodeToString(kid)
		mj, err := k.MarshalJSON()
		if err != nil {
			slog.Error("jwk marshal", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		j = append(j, mj...)
	}
	if privEdDSA != nil {
		k := jose.JSONWebKey{Key: privEdDSA.Public()}
		mj, err := k.MarshalJSON()
		if err != nil {
			slog.Error("jwk marshal key", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		j = append(j, []byte(`,`)...)
		j = append(j, mj...)
	}
	j = append(j, []byte(`]}`)...)

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	writeJsonBytes(w, j)
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
		slog.Error("basic auth failed")
		WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
		return
	}
	if refreshToken == "" {
		slog.Error("no refresh token")
		WriteErrorf(w, http.StatusBadRequest, "No refresh token")
		return
	}

	refreshClaims, err := checkRefreshToken(refreshToken)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}

	encodedAccessToken, encodedRefreshToken, expiresAt, err := checkRefresh(refreshClaims.Subject, refreshClaims.Token)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}

	oauth := OAuth{
		AccessToken:  encodedAccessToken,
		RefreshToken: encodedRefreshToken,
		ExpiresAt:    strconv.FormatInt(expiresAt, 10)}

	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}
	writeJsonBytes(w, oauthEnc)
}

func oauth(w http.ResponseWriter, r *http.Request) {
	grantType, err := param("grant_type", r)
	if err != nil {
		slog.Error("basic auth failed")
		WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
		return
	}

	switch grantType {
	case "refresh_token":
		refresh(w, r)
	case "client_credentials":
		user, err := basicAuth(r)
		if err != nil {
			slog.Error("Basic auth failed", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
			return
		}

		encodedAccessToken, err := encodeAccessToken(user, cfg.Scope, cfg.Audience, cfg.Issuer, nil)
		if err != nil {
			slog.Error("Basic auth failed", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
			return
		}

		oauth := OAuthSystem{
			AccessToken: encodedAccessToken,
		}
		oauthEnc, err := json.Marshal(oauth)
		if err != nil {
			slog.Error("cannot verify refresh token", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
			return
		}
		writeJsonBytes(w, oauthEnc)

	case "authorization_code":
		code, err := param("code", r)
		if err != nil {
			slog.Error("basic auth failed")
			WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
			return
		}
		codeVerifier, err := param("code_verifier", r)
		if err != nil {
			slog.Error("basic auth failed")
			WriteErrorf(w, http.StatusBadRequest, BasicAuthFailedMessage)
			return
		}
		//https://tools.ietf.org/html/rfc7636#section-4.1 length must be <= 43 <= 128
		if len(codeVerifier) < 43 {
			slog.Error("minimum 43 characters required")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		if len(codeVerifier) > 128 {
			slog.Error("maximum 128 characters allowed")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		cc, err := checkCodeToken(code)
		if err != nil {
			slog.Error("code check failed", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		if cc.CodeCodeChallengeMethod == "S256" {
			h := sha256.Sum256([]byte(codeVerifier))
			s := base64.RawURLEncoding.EncodeToString(h[:])
			if cc.CodeChallenge != s {
				slog.Error("auth challenge failed")
				WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
				return
			}
		} else {
			slog.Error("only S256 supported")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}

		result, err := findAuthByEmail(cc.Subject)
		if err != nil {
			slog.Error("findAuthByEmail failed",
				slog.String("subject", cc.Subject),
				slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}

		writeOAuth(w, result)
	case "password":
		if !cfg.PasswordFlow {
			slog.Error("no username")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		email, err := param("username", r)
		if err != nil {
			slog.Error("no username")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		password, err := param("password", r)
		if err != nil {
			slog.Error("no password")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		scope, err := param("scope", r)
		if err != nil {
			slog.Error("no scope")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		if email == "" || password == "" || scope == "" {
			slog.Error("ERR-oauth-05, username, password, or scope empty")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}

		result, errString, err := checkEmailPassword(email, password)
		if err != nil {
			slog.Error("ERR-oauth-06",
				slog.Any("error", err),
				slog.String("errorString", errString))
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}

		writeOAuth(w, result)
	default:
		slog.Error("unsupported grant type")
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
		slog.Error("unsupported grant type")
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	if tokenHint == "refresh_token" {
		oldToken, err := param("token", r)
		if err != nil {
			slog.Error("unsupported grant type")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		if oldToken == "" {
			slog.Error("unsupported grant type")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
		_, err = resetRefreshToken(oldToken)
		if err != nil {
			slog.Error("unsupported grant type")
			WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
			return
		}
	} else {
		slog.Error("unsupported grant type")
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
}

func logout(w http.ResponseWriter, r *http.Request, claims *TokenClaims) {
	keys := r.URL.Query()
	redirectUri := keys.Get("redirect_uri")

	result, err := findAuthByEmail(claims.Subject)
	if err != nil {
		slog.Error("logout", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	refreshToken := result.refreshToken
	_, err = resetRefreshToken(refreshToken)
	if err != nil {
		slog.Error("unsupported grant type", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
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
	time := `{"time":"` + currentTime.Format("2006-01-02 15:04:05") + `","offset":` + strconv.Itoa(secondsAdd) + `}`
	writeJsonBytes(w, []byte(time))
}

func timeWarp(w http.ResponseWriter, r *http.Request, adminEmail string) {
	h := r.PathValue("hours")
	if h == "" {
		slog.Error("timewarp parameter", slog.Any("message", m))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	hours, err := strconv.Atoi(h)
	if err != nil {
		slog.Error("timewarp parse", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	seconds := hours * 60 * 60
	secondsAdd += seconds
	slog.Info("time warp", slog.String("time", timeNow().String()))

	//since we warp, the token will be invalid
	result, err := findAuthByEmail(adminEmail)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("adminEmail", adminEmail),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	writeOAuth(w, result)
}

func asUser(w http.ResponseWriter, r *http.Request, _ string) {
	email := r.PathValue("email")
	result, err := findAuthByEmail(email)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	writeOAuth(w, result)
}

func deleteUser(w http.ResponseWriter, r *http.Request, admin string) {
	email := r.PathValue("email")
	err := deleteDbUser(email)
	if err != nil {
		slog.Error("could not delete user",
			slog.Any("error", err),
			slog.String("requestedBy", admin))

		WriteErrorf(w, http.StatusBadRequest, "Could not delete user")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateUser(w http.ResponseWriter, r *http.Request, admin string) {
	//now we update the meta data that comes as system meta data. Thus we trust the system to provide the correct metadata, not the user
	email := r.PathValue("email")
	b, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("could not update user",
			slog.Any("error", err),
			slog.String("requestedBy", admin))
		WriteErrorf(w, http.StatusBadRequest, "Could not update user")
		return
	}
	if !json.Valid(b) {
		slog.Error("invalid json",
			slog.String("json", string(b)),
			slog.String("requestedBy", admin))
		WriteErrorf(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	err = updateSystemMeta(email, string(b))
	if err != nil {
		slog.Error("could not update system meta",
			slog.Any("error", err),
			slog.String("requestedBy", admin))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
