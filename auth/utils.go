package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/xlzd/gotp"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func genRnd(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func genToken() (string, error) {
	rn, err := genRnd(20)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(rn), nil
}

func validateEmail(email string) error {
	rxEmail := regexp.MustCompile(`(?i)^[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)
	if len(email) > 254 || !rxEmail.MatchString(email) {
		return fmt.Errorf("[%s] is not a valid email address", email)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password is less than 8 characters")
	}
	return nil
}

func newTOTP(secret string) *gotp.TOTP {
	hasher := &gotp.Hasher{
		HashName: "sha256",
		Digest:   sha256.New,
	}
	return gotp.NewTOTP(secret, 6, 30, hasher)
}

func basicAuth(r *http.Request) (string, error) {
	if cfg.OAuthUser != "" || cfg.OAuthPass != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != cfg.OAuthUser || pass != cfg.OAuthPass {
			clientId, err := param("client_id", r)
			if err != nil {
				return "", fmt.Errorf("no client_id %v", err)
			}
			clientSecret, err := param("client_secret", r)
			if err != nil {
				return "", fmt.Errorf("no client_secret %v", err)
			}
			if clientId != cfg.OAuthUser || clientSecret != cfg.OAuthPass {
				return "", fmt.Errorf("no match, user/pass %v", err)
			}
		}
		return user, nil
	}
	return "", fmt.Errorf("no user/pass set")
}

func param(name string, r *http.Request) (string, error) {
	n1 := r.PathValue(name)
	n2, err := url.QueryUnescape(r.URL.Query().Get(name))
	if err != nil {
		return "", err
	}
	err = r.ParseForm()
	if err != nil {
		return "", err
	}
	n3 := r.FormValue(name)

	if n1 == "" {
		if n2 == "" {
			return n3, nil
		}
		return n2, nil
	}
	return n1, nil
}

func paramJson(name string, r *http.Request) (string, error) {
	var objmap map[string]json.RawMessage
	err := json.NewDecoder(r.Body).Decode(&objmap)
	if err != nil {
		return "", err
	}
	var s string
	err = json.Unmarshal(objmap[name], &s)
	if err != nil {
		return "", err
	}
	return s, nil
}

func timeNow() time.Time {
	if debug {
		return time.Now().Add(time.Duration(secondsAdd) * time.Second).UTC()
	} else {
		return time.Now().UTC()
	}
}

func writeJsonBytes(w http.ResponseWriter, obj []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(obj)
	if err != nil {
		slog.Error("Could write json", slog.Any("error", err))
		WriteErrorf(w, http.StatusBadRequest, GenericErrorMessage)
	}
}

func WriteErrorf(w http.ResponseWriter, statusCode int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	slog.Error("error while trying to encode", slog.String("msg", msg))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(statusCode)

	msgEnc, err := json.Marshal(msg)
	if err != nil {
		slog.Error("error while trying to encode", slog.String("msg", msg), slog.Any("error", err))
		return
	}
	_, err = w.Write([]byte(`{"error":` + string(msgEnc) + `}`))
	if err != nil {
		slog.Error("Something went wrong while writing error message", slog.Any("error", err))
		return
	}
}

func LookupEnv(key string, defaultValues ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	for _, v := range defaultValues {
		if v != "" {
			err := os.Setenv(key, v)
			if err != nil {
				slog.Debug("Could not set env variable", slog.String("key", key), slog.String("value", v), slog.Any("error", err))
				return ""
			}
			return v
		}
	}
	return ""
}

func LookupEnvInt(key string, defaultValues ...int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			slog.Debug("LookupEnvInt[%s]: %v", slog.String(key, key), slog.Any("error", err))
			return 0
		}
		return v
	}
	for _, v := range defaultValues {
		if v != 0 {
			err := os.Setenv(key, strconv.Itoa(v))
			if err != nil {
				slog.Debug("Could not set env variable", slog.String("key", key), slog.Int("value", v), slog.Any("error", err))
				return 0
			}
			return v
		}
	}
	return 0
}

type SendEmailRequest struct {
	SendgridRequest SendgridRequest
	Url             string
	EmailFromName   string
	EmailFrom       string
	EmailToken      string
}

type SendgridRequest struct {
	MailTo      string `json:"mail_to,omitempty"`
	Subject     string `json:"subject"`
	TextMessage string `json:"text_message"`
	HtmlMessage string `json:"html_message"`
}

func SendEmail(sendEmailRequest SendEmailRequest) error {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var jsonData []byte
	var err error
	if strings.Contains(sendEmailRequest.Url, "sendgrid") {
		sendGridReq := mail.NewSingleEmail(
			mail.NewEmail(sendEmailRequest.EmailFromName, sendEmailRequest.EmailFrom),
			sendEmailRequest.SendgridRequest.Subject,
			mail.NewEmail("", sendEmailRequest.SendgridRequest.MailTo),
			sendEmailRequest.SendgridRequest.TextMessage,
			sendEmailRequest.SendgridRequest.HtmlMessage)
		jsonData, err = json.Marshal(sendGridReq)
	} else {
		jsonData, err = json.Marshal(sendEmailRequest.SendgridRequest)
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", sendEmailRequest.Url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+sendEmailRequest.EmailToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("could not send email: %v %v", resp.Status, resp.StatusCode)
	}
	return nil
}

func PrepareEmail(
	mailTo string,
	data map[string]string,
	templateKey string,
	defaultSubject string,
	defaultText string,
	lang string) SendgridRequest {
	textMessage := parseTemplate("plain/"+lang+"/"+templateKey+".txt", data)
	if textMessage == "" {
		textMessage = defaultText
	}

	headerTemplate := parseTemplate("html/"+lang+"/header.html", data)
	footerTemplate := parseTemplate("html/"+lang+"/footer.html", data)
	htmlBody := parseTemplate("html/"+lang+"/"+templateKey+".html", data)
	htmlMessage := headerTemplate + htmlBody + footerTemplate

	return SendgridRequest{
		MailTo:      mailTo,
		Subject:     defaultSubject,
		TextMessage: textMessage,
		HtmlMessage: htmlMessage,
	}
}

func parseTemplate(filename string, other map[string]string) string {
	textMessage := ""
	tmplPlain, err := template.ParseFiles("mail-templates/" + filename)
	if err == nil {
		var buf bytes.Buffer
		err = tmplPlain.Execute(&buf, other)
		if err == nil {
			textMessage = buf.String()
		} else {
			slog.Warn("cannot execute template file", slog.String("filename", filename), slog.Any("error", err))
		}
	} else {
		slog.Warn("cannot prepare file template file", slog.String("filename", filename), slog.Any("error", err))
	}
	return textMessage
}
