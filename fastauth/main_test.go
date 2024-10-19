package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xlzd/gotp"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	testDBPath    = "/tmp/fa.db"
	testDBDriver  = "sqlite3"
	testDBScripts = "rmdb.sql:init.sql"
	testDomain    = "localhost"
	testPort      = 8082
	testUrl       = "http://" + testDomain + ":8082"
)

var (
	testParams = []string{"-issuer=FFFS",
		"-port=" + strconv.Itoa(testPort),
		"-db-path=" + testDBPath,
		"-db-driver=" + testDBDriver,
		"-db-scripts=" + testDBScripts,
		"-email-url=" + testUrl + "/send/email/{email}/{token}",
		"-dev=true"}
)

/*
curl -v "http://localhost:8080/signup"   -X POST   -d "{\"email\":\"tom\",\"password\":\"test\"}"   -H "Content-Type: application/json"
*/
func TestSignup(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tom@test.ch", "testtest")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSignupWrongEmail(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tomtest.ch", "testtest")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	assert.True(t, strings.Index(bodyString, "Oops something went wrong. Please try again.") > 0)
}

func TestSignupTwiceWorking(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tom@test.ch", "testtest")
	resp = doSignup(router, "tom@test.ch", "testtest")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSignupTwiceNotWorking(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tom@test.ch", "testtest")
	token := token("tom@test.ch")
	resp = doConfirm(router, "tom@test.ch", token)
	resp = doSignup(router, "tom@test.ch", "testtest")

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	log.Println(bodyString)
	assert.True(t, strings.Index(bodyString, "Oops something went wrong. Please try again.") > 0)
}

func TestConfirm(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tom@test.ch", "testtest")
	assert.Equal(t, 200, resp.StatusCode)

	token := token("tom@test.ch")
	resp = doConfirm(router, "tom@test.ch", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogin(t *testing.T) {
	router := mainTest(testParams...)
	resp := doSignup(router, "tom@test.ch", "testtest")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	token := token("tom@test.ch")
	resp = doConfirm(router, "tom@test.ch", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest", "", "")
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
}

func TestLoginFalse(t *testing.T) {
	router := mainTest(testParams...)
	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest", "", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest2", "", "0123456789012345678901234567890123456789012")
	bodyBytes, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.True(t, strings.Index(string(bodyBytes), "Oops something went wrong. Please try again.") > 0)

	resp = doLogin(router, "tom@test.ch2", "testtest", "", "0123456789012345678901234567890123456789012")
	bodyBytes, _ = io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.True(t, strings.Index(string(bodyBytes), "Oops something went wrong. Please try again.") > 0)
}

func TestRefresh(t *testing.T) {
	tmp := append(testParams, "-expire-refresh=10")
	router := mainTest(tmp...)
	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	oauth := OAuth{}
	json.NewDecoder(resp.Body).Decode(&oauth)
	assert.NotEqual(t, "", oauth.AccessToken)
}

func TestReset(t *testing.T) {
	tmp := append(testParams, "-expire-refresh=1")
	router := mainTest(tmp...)

	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doReset(router, "tom@test.ch")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	token, _ := getForgotEmailToken("tom@test.ch")

	resp = doConfirmReset(router, "tom@test.ch", token, "testtest2")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest2", "", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
}

func TestResetFailed(t *testing.T) {
	tmp := append(testParams, "-expire-refresh=1")
	router := mainTest(tmp...)
	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doReset(router, "tom@test.ch")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	token, _ := getForgotEmailToken("tom@test.ch")

	resp = doConfirmReset(router, "tom@test.ch", token, "testtest2")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest", "", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTOTP(t *testing.T) {
	router := mainTest(testParams...)
	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	oauth := OAuth{}
	json.NewDecoder(resp.Body).Decode(&oauth)

	resp = doTOTP(router, oauth.AccessToken)
	p := ProvisioningUri{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &p)

	u, err := url.Parse(p.Uri)
	require.Nil(t, err)

	secret := u.Query().Get("secret")
	require.NotNil(t, secret)

	totp := newTOTP(secret)
	conf := totp.Now()

	resp = doTOTPConfirm(router, conf, oauth.AccessToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginTOTP(t *testing.T) {
	router := mainTest(testParams...)
	resp := doAll(router, "tom@test.ch", "testtest", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	oauth := OAuth{}
	json.NewDecoder(resp.Body).Decode(&oauth)

	totp := doAllTOTP(router, oauth.AccessToken)

	resp = doLogin(router, "tom@test.ch", "testtest", totp.Now(), "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)

	resp = doLogin(router, "tom@test.ch", "testtest", "", "0123456789012345678901234567890123456789012")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func doTOTP(router *http.ServeMux, token string) *http.Response {
	req, _ := http.NewRequest("POST", testUrl+"/setup/totp", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doTOTPConfirm(router *http.ServeMux, conf string, token string) *http.Response {
	req, _ := http.NewRequest("POST", testUrl+"/confirm/totp/"+conf, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doAllTOTP(router *http.ServeMux, token string) *gotp.TOTP {
	resp := doTOTP(router, token)
	p := ProvisioningUri{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &p)

	u, err := url.Parse(p.Uri)
	if err != nil {
		panic(err)
	}

	secret := u.Query().Get("secret")
	totp := newTOTP(secret)
	conf := totp.Now()
	resp = doTOTPConfirm(router, conf, token)
	return totp
}

func doConfirmReset(router *http.ServeMux, email string, token string, password string) *http.Response {
	data := Credentials{
		Email:      email,
		Password:   password,
		EmailToken: token,
	}
	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)
	req, _ := http.NewRequest("POST", testUrl+"/confirm/reset", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doReset(router *http.ServeMux, email string) *http.Response {
	req, _ := http.NewRequest("POST", testUrl+"/reset/"+email, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doAll(router *http.ServeMux, email string, pass string, secret string) *http.Response {
	resp := doSignup(router, email, pass)
	token := token(email)
	resp = doConfirm(router, email, token)
	resp = doLogin(router, email, pass, "", secret)
	code := resp.Header.Get("Location")[6:]
	resp = doCode(router, code, secret)
	return resp
}

func doCode(router *http.ServeMux, codeToken string, codeVerifier string) *http.Response {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", codeToken)
	data.Set("code_verifier", codeVerifier)
	req, _ := http.NewRequest("POST", testUrl+"/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doRefresh(router *http.ServeMux, refreshToken string) *http.Response {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	req, _ := http.NewRequest("POST", testUrl+"/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doLogin(router *http.ServeMux, email string, pass string, totp string, secret string) *http.Response {
	h := sha256.Sum256([]byte(secret))
	data := Credentials{
		Email:                   email,
		Password:                pass,
		TOTP:                    totp,
		CodeChallenge:           base64.RawURLEncoding.EncodeToString(h[:]),
		CodeCodeChallengeMethod: "S256",
	}

	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)
	req, _ := http.NewRequest(http.MethodPost, testUrl+"/login", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doSignup(router *http.ServeMux, email string, pass string) *http.Response {
	data := Credentials{
		Email:    email,
		Password: pass,
	}
	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", testUrl+"/signup", body)
	if err != nil {
		log.Printf("request failed %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func doConfirm(router *http.ServeMux, email string, token string) *http.Response {

	req, err := http.NewRequest("GET", testUrl+"/confirm/signup/"+email+"/"+token, nil)
	if err != nil {
		log.Printf("request failed %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}

func token(email string) string {
	r, _ := getEmailToken(email)
	return r
}

func getEmailToken(email string) (string, error) {
	var emailToken string
	err := DB.QueryRow("SELECT email_token from auth where email = $1", email).Scan(&emailToken)
	if err != nil {
		return "", err
	}
	return emailToken, nil
}

func getForgotEmailToken(email string) (string, error) {
	var forgetEmailToken string
	err := DB.QueryRow("SELECT forget_email_token from auth where email = $1", email).Scan(&forgetEmailToken)
	if err != nil {
		return "", err
	}
	return forgetEmailToken, nil
}

func TestSecret(t *testing.T) {
	h := sha256.Sum256([]byte("test"))
	s := base64.RawURLEncoding.EncodeToString(h[:])
	assert.Equal(t, "n4bQgYhMfWWaL-qgxVrQFaO_TxsrC4Is0V1sFbDwCgg", s)
}

func mainTest(args ...string) *http.ServeMux {
	oldArgs := os.Args
	os.Args = []string{oldArgs[0]}
	os.Args = append(os.Args, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) //flags are now reset

	opts = NewOpts()
	var err error
	err = InitDb(opts.DBDriver, opts.DBPath, opts.DBScripts)
	if err != nil {
		log.Fatal(err)
	}
	addInitialUsers()

	return setupRouter()
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
