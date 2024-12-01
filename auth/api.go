package main

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"golang.org/x/text/language"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	matcher = language.NewMatcher([]language.Tag{
		language.English,
		language.German,
	})
)

const (
	GenericErrorMessage             = "Oops something went wrong. Please try again."
	CannotVerifyRefreshTokenMessage = "Cannot verify refresh token"
	NotAllowedToViewMessage         = "Oops you are not allowed to view this resource."
)

func logout(w http.ResponseWriter, r *http.Request, claims *jwt.Claims) {
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

func refresh(w http.ResponseWriter, r *http.Request) {
	var refresh Refresh
	err := json.NewDecoder(r.Body).Decode(&refresh)
	if err != nil {
		slog.Error("cannot parse JSON credentials", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	refreshClaims, err := checkRefreshToken(refresh.RefreshToken)
	if err != nil {
		slog.Error("cannot checkRefreshToken", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}

	if refresh.Email != "" && refreshClaims.Subject != refresh.Email {
		slog.Error("wrong email for refresh token")
		WriteErrorf(w, http.StatusNotFound, CannotVerifyRefreshTokenMessage)
		return
	}

	encodedAccessToken, encodedRefreshToken, expiresAt, err := checkRefresh(refreshClaims.Subject, refreshClaims.Token)
	if err != nil {
		slog.Error("cannot do refresh", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}

	oAuth := &OAuth{
		AccessToken:  encodedAccessToken,
		RefreshToken: encodedRefreshToken,
		ExpiresAt:    expiresAt}

	oauthEnc, err := json.Marshal(oAuth)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}
	writeJsonBytes(w, oauthEnc)
}

func confirm(w http.ResponseWriter, r *http.Request) {
	vars := r.PathValue("email")
	email, err := url.QueryUnescape(vars)
	if err != nil {
		slog.Error("cannot QueryUnescape email", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	vars = r.PathValue("emailToken")
	emailToken, err := url.QueryUnescape(vars)
	if err != nil {
		slog.Error("cannot QueryUnescape emailToken", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rc, err := checkRefreshToken(emailToken)
	if err != nil {
		slog.Error("cannot checkRefreshToken emailToken", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}
	err = updateEmailToken(rc.Subject, rc.Token)
	if err != nil {
		slog.Error("update email token failed",
			slog.String("email", email),
			slog.String("token", emailToken),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusForbidden, GenericErrorMessage)
		return
	}

	result, err := findAuthByEmail(email)
	if err != nil {
		slog.Error("findAuthByEmail failed",
			slog.String("subject", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	writeOAuth(w, result)
}

func login(w http.ResponseWriter, r *http.Request) {
	vars := r.PathValue("email")
	email, err := url.QueryUnescape(vars)
	if err != nil {
		slog.Error("cannot QueryUnescape email", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	err = validateEmail(email)
	if err != nil {
		slog.Error("cannot validate email", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	//special case to skip email verification
	for _, noAuthUser := range noAuthUsers {
		if email == noAuthUser {
			res, err := findAuthByEmail(email)
			if err != nil {
				slog.Error("cannot validate email", slog.Any("error", err))
				WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
				return
			}
			writeOAuth(w, res)
			return
		}
	}

	emailTokenRnd, err := genToken()
	if err != nil {
		slog.Error("RND error email",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	rc := &RefreshClaims{
		ExpiresAt: timeNow().Add(time.Minute * 30).Unix(),
		Subject:   email,
		Token:     emailTokenRnd,
	}
	emailToken, err := encodeAnyToken(rc)
	if err != nil {
		slog.Error("encodeAnyToken error",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	refreshToken, err := genToken()
	if err != nil {
		slog.Error("RND error auth",
			slog.String("email", email),
			slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	//only inserts refreshToken if the user does not exist, otherwise it would logout all other sessions
	err = insertOrUpdateUser(email, emailTokenRnd, refreshToken, timeNow())
	if err != nil {
		slog.Error("insert user failed", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
		return
	}

	params := map[string]string{}
	params["token"] = emailToken
	params["email"] = email
	params["url"] = cfg.EmailLinkPrefix + "/login-confirm?email=" + url.QueryEscape(email) + "&emailToken=" + emailToken
	params["lang"] = lang(r)

	sendgridRequest := PrepareEmail(email, params,
		"Validate your email",
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

func serverTime(w http.ResponseWriter, r *http.Request, email string) {
	currentTime := timeNow()
	time := `{"time":"` + currentTime.Format("2006-01-02 15:04:05") + `","offset":` + strconv.Itoa(secondsAdd) + `}`
	writeJsonBytes(w, []byte(time))
}

func timeWarp(w http.ResponseWriter, r *http.Request, adminEmail string) {
	h := r.PathValue("hours")
	if h == "" {
		slog.Error("timewarp parameter", slog.Any("message", h))
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
		ExpiresAt:    expiresAt,
	}

	oauthEnc, err := json.Marshal(oauth)
	if err != nil {
		slog.Error("cannot verify refresh token", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, CannotVerifyRefreshTokenMessage)
		return
	}
	writeJsonBytes(w, oauthEnc)
}
