package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func jwtAuthAdmin(next func(w http.ResponseWriter, r *http.Request, email string), emails []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth0(r)
		if claims != nil && err != nil {
			slog.Error("Token expired", slog.String("claims.Subject", claims.Subject))
			WriteErrorf(w, http.StatusUnauthorized, GenericErrorMessage)
			return
		} else if claims == nil && err != nil {
			slog.Error("jwtAuthAdmin error", slog.Any("error", err))
			WriteErrorf(w, http.StatusBadRequest, NotAllowedToViewMessage)
			return
		}
		for _, email := range emails {
			if claims.Subject == email {
				slog.Info("Authenticated admin", slog.String("email", email))
				next(w, r, email)
				return
			}
		}
		slog.Error("No admin found", slog.String("claims.Subject", claims.Subject), slog.Any("emails", emails))
		WriteErrorf(w, http.StatusBadRequest, NotAllowedToViewMessage)
	}
}

func jwtAuth(next func(w http.ResponseWriter, r *http.Request, claims *jwt.Claims)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := jwtAuth0(r)
		if err != nil {
			keys := r.URL.Query()
			ru := keys.Get("redirect_uri")

			if claims != nil {
				if ru != "" {
					slog.Error("Token expired", slog.String("subject", claims.Subject))
					w.Header().Set("Location", ru)
					w.WriteHeader(http.StatusSeeOther)
				} else {
					slog.Error("Token expired", slog.String("subject", claims.Subject))
					WriteErrorf(w, http.StatusUnauthorized, GenericErrorMessage)
				}
			} else if claims == nil {
				if ru != "" {
					slog.Error("jwtAuth error", slog.Any("error", err))
					w.Header().Set("Location", ru)
					w.WriteHeader(http.StatusSeeOther)
				} else {
					slog.Error("jwtAuth error", slog.Any("error", err))
					WriteErrorf(w, http.StatusBadRequest, NotAllowedToViewMessage)
				}
			}
			return
		}
		next(w, r, claims)
	}
}

func jwtAuth0(r *http.Request) (*jwt.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return nil, fmt.Errorf("ERR-02, could not split token, auth header is: [%v]", authHeader)
	}
	bearerToken := split[1]

	tok, err := jwt.ParseSigned(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("ERR-03, could not parse token: %v", bearerToken[1])
	}

	claims := &jwt.Claims{}

	if tok.Headers[0].Algorithm == string(jose.RS256) {
		err = tok.Claims(privRSA.Public(), claims)
	} else if tok.Headers[0].Algorithm == string(jose.HS256) {
		err = tok.Claims(jwtKey, claims)
	} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
		err = tok.Claims(privEdDSA.Public(), claims)
	} else {
		return nil, fmt.Errorf("ERR-04, unknown algorithm: %v", tok.Headers[0].Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("ERR-05, could not parse claims: %v", bearerToken)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(timeNow()) {
		return claims, fmt.Errorf("ERR-06, unauthorized: %v", bearerToken)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("ERR-07, no subject: %v", claims)
	}
	return claims, nil
}

func checkRefreshToken(token string) (*RefreshClaims, error) {
	tok, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, fmt.Errorf("ERR-check-refresh-01, could not check sig %v", err)
	}
	refreshClaims := &RefreshClaims{}
	if tok.Headers[0].Algorithm == string(jose.RS256) {
		err := tok.Claims(privRSA.Public(), refreshClaims)
		if err != nil {
			return nil, fmt.Errorf("ERR-check-refresh-02, could not parse claims %v", err)
		}
	} else if tok.Headers[0].Algorithm == string(jose.HS256) {
		err := tok.Claims(jwtKey, refreshClaims)
		if err != nil {
			return nil, fmt.Errorf("ERR-check-refresh-03, could not parse claims %v", err)
		}
	} else if tok.Headers[0].Algorithm == string(jose.EdDSA) {
		err := tok.Claims(privEdDSA.Public(), refreshClaims)
		if err != nil {
			return nil, fmt.Errorf("ERR-check-refresh-04, could not parse claims %v", err)
		}
	} else {
		return nil, fmt.Errorf("ERR-check-refresh-05, could not parse claims, no algo found %v", tok.Headers[0].Algorithm)
	}
	t := time.Unix(refreshClaims.ExpiresAt, 0)
	if !t.After(timeNow()) {
		return nil, fmt.Errorf("ERR-check-refresh-06, expired %v", err)
	}
	return refreshClaims, nil
}

func encodeAccessToken(subject string, systemMeta map[string]interface{}) (string, error) {
	//if we have a system user, the system user only gets an access token that lives as long as the refresh token
	//normal users get both, access and refresh token
	var expiry *jwt.NumericDate
	if subject == "system" {
		expiry = jwt.NewNumericDate(timeNow().Add(refreshExp))
	} else {
		expiry = jwt.NewNumericDate(timeNow().Add(tokenExp))
	}

	tokenClaims := &jwt.Claims{
		Expiry:   expiry,
		Subject:  subject,
		IssuedAt: jwt.NewNumericDate(timeNow()),
	}

	var sig jose.Signer
	var err error
	if jwtKey != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: jwtKey}, (&jose.SignerOptions{}).WithType("JWT"))
	} else if privRSA != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privRSA}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", privRSAKid))
	} else if privEdDSA != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: *privEdDSA}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", privEdDSAKid))
	} else {
		return "", fmt.Errorf("JWT access token %v no key", tokenClaims.Subject)
	}

	if err != nil {
		return "", fmt.Errorf("JWT access token %v failed: %v", tokenClaims.Subject, err)
	}
	accessTokenString, err := jwt.Signed(sig).Claims(systemMeta).Claims(tokenClaims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("JWT access token %v failed: %v", tokenClaims.Subject, err)
	}
	if cfg.Dev != "" {
		slog.Debug("Access token", slog.String("accessToken", accessTokenString))
	}
	return accessTokenString, nil
}

/*
If the option ResetRefresh is set, then every time this function is called, which is
before the createRefreshToken, then the refresh token is renewed and the old one is
not valid anymore.

This function is also used in case of revoking a token, where a new token is created,
but not returned to the user, so the user has to login to get the refresh token
*/
func resetRefreshToken(oldToken string) (string, error) {
	newToken, err := genToken()
	if err != nil {
		return "", err
	}
	err = updateRefreshToken(oldToken, newToken)
	if err != nil {
		return "", err
	}
	return newToken, nil
}

func checkRefresh(email string, token string) (string, string, int64, error) {
	result, err := findAuthByEmail(email)
	if err != nil {
		return "", "", 0, fmt.Errorf("ERR-refresh-03, DB select, %v err %v", email, err)
	}

	if result.emailToken != nil {
		return "", "", 0, fmt.Errorf("ERR-refresh-04, user %v no email verified: %v", email, err)
	}

	if result.refreshToken == "" || token != result.refreshToken {
		if cfg.Dev != "" {
			slog.Warn("refresh token mismatch, not the same", slog.String("token", token), slog.String("result.refreshToken", result.refreshToken))
		}
		return "", "", 0, fmt.Errorf("ERR-refresh-05, refresh token mismatch")

	}
	return encodeTokens(result)
}

func encodeTokens(result *dbRes) (string, string, int64, error) {
	encodedAccessToken, err := encodeAccessTokens(result.email, result.metaSystem)
	if err != nil {
		return "", "", 0, err
	}
	encodedRefreshToken, expireAt, err := encodeRefreshTokens(result.email, result.refreshToken)
	if err != nil {
		return "", "", 0, err
	}
	return encodedAccessToken, encodedRefreshToken, expireAt, nil
}

func encodeAccessTokens(email string, metaSystem *string) (string, error) {
	jsonMapSystem, err := toJsonMap(metaSystem)
	if err != nil {
		return "", fmt.Errorf("cannot encode system meta in encodeTokens for %v, %v", email, err)
	}

	encodedAccessToken, err := encodeAccessToken(email, jsonMapSystem)
	if err != nil {
		return "", fmt.Errorf("ERR-refresh-06, cannot set access token for %v, %v", email, err)
	}

	return encodedAccessToken, nil
}

func encodeRefreshTokens(email string, refreshToken string) (string, int64, error) {
	rc := &RefreshClaims{}
	rc.ExpiresAt = timeNow().Add(refreshExp).Unix()
	rc.Subject = email
	rc.Token = refreshToken

	encodedRefreshToken, err := encodeAnyToken(rc)
	if err != nil {
		return "", 0, fmt.Errorf("ERR-refresh-08, cannot set refresh token for %v, %v", email, err)
	}
	return encodedRefreshToken, rc.ExpiresAt, nil
}

func encodeAnyToken(rc interface{}) (string, error) {
	var sig jose.Signer
	var err error
	if jwtKey != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: jwtKey}, (&jose.SignerOptions{}).WithType("JWT"))
	} else if privRSA != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privRSA}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", privRSAKid))
	} else if privEdDSA != nil {
		sig, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: *privEdDSA}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", privEdDSAKid))
	} else {
		return "", fmt.Errorf("JWT refresh token %v no key", rc)
	}

	if err != nil {
		return "", fmt.Errorf("JWT refresh token %v failed: %v", rc, err)
	}
	refreshToken, err := jwt.Signed(sig).Claims(rc).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("JWT serialize %v failed: %v", rc, err)
	}
	if cfg.Dev != "" {
		slog.Debug("Refresh token", slog.String("refreshToken", refreshToken))
	}
	return refreshToken, nil
}

func toJsonMap(jsonStr *string) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	if jsonStr != nil {
		err := json.Unmarshal([]byte(*jsonStr), &jsonMap)
		if err != nil {
			return nil, fmt.Errorf("ERR-refresh-06, cannot create json map %v", err)
		}
	}
	return jsonMap, nil
}
