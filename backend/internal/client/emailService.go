package client

import (
	"backend/internal/db"
	"backend/pkg/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/template"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	KeyTopUpStripe         = "topup-stripe"
	KeyTopUpOther          = "topup-other"
	KeyTopUpUser1          = "topup-user1"
	KeyTopUpUser2          = "topup-user2"
	KeyMarketing           = "marketing"
	KeyStripeFailed        = "stripe-failed"
	KeyStripeAction        = "stripe-action"
	KeyStripeSuccess       = "stripe-success"
	KeyAddGit              = "add-git"
	KeyPaymentNowFinished  = "paymentnow-finished"
	KeyPaymentNowPartially = "paymentnow-partially"
	KeyPaymentNowRefunded  = "paymentnow-refunded"
	WaitToSendEmail        = 60 * 60 * 24 // for testing, the make it 7 days
)

var (
	queue                chan *SendEmailRequest
	EmailNotifications   = 0
	EmailNoNotifications = 0
	lastMailTo           = ""
)

type EmailClient struct {
	HTTPClient      *http.Client
	emailUrl        string
	emailFromName   string
	emailFrom       string
	emailToken      string
	env             string
	emailMarketing  string
	emailLinkPrefix string
}

func NewEmailClient(emailUrl string, emailFromName string, emailFrom string, emailToken string, env string, emailMarketing string, emailLinkPrefix string) *EmailClient {
	return &EmailClient{
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		}, emailUrl: emailUrl,
		emailFromName:   emailFromName,
		emailFrom:       emailFrom,
		emailToken:      emailToken,
		env:             env,
		emailMarketing:  emailMarketing,
		emailLinkPrefix: emailLinkPrefix,
	}
}

type WebhookResponse struct {
	ActuallyPaid     float64    `json:"actually_paid"`
	InvoiceId        int64      `json:"invoice_id"`
	OrderDescription *uuid.UUID `json:"order_description"`
	OrderId          *uuid.UUID `json:"order_id"`
	OutcomeAmount    float64    `json:"outcome_amount"`
	OutcomeCurrency  string     `json:"outcome_currency"`
	PayAddress       string     `json:"pay_address"`
	PayAmount        float64    `json:"pay_amount"`
	PayCurrency      string     `json:"pay_currency"`
	PaymentId        int64      `json:"payment_id"`
	PaymentStatus    string     `json:"payment_status"`
	PriceAmount      float64    `json:"price_amount"`
	PriceCurrency    string     `json:"price_currency"`
	PurchaseId       string     `json:"purchase_id"`
}

func init() {
	queue = make(chan *SendEmailRequest)
	go func() {
		for {
			select {
			case e := <-queue:
				err := SendEmail(*e)
				if err != nil {
					slog.Error("cannot send email",
						slog.Any("error", err))
				}
			}
		}
	}()

}

func sendEmail(e *SendEmailRequest) {
	queue <- e
}

func shouldSendEmail(uId *uuid.UUID, email string, key string) (bool, error) {
	c := 0
	var err error
	if uId != nil {
		c, err = db.CountEmailSentById(*uId, key)
	} else {
		c, err = db.CountEmailSentByEmail(email, key)
	}
	if err != nil {
		return false, err
	}

	if c > 0 {
		slog.Info("we already sent a notification",
			slog.String("email", email))
		EmailNoNotifications++
		return false, nil
	}
	id := uuid.New()
	err = db.InsertEmailSent(id, uId, email, key, util.TimeNow())
	if err != nil {
		return false, err
	}
	EmailNotifications++
	lastMailTo = email
	return true, nil
}

func (e *EmailClient) prepareSendEmail(
	uid *uuid.UUID,
	data map[string]string,
	templateKey string,
	defaultSubject string,
	defaultText string,
	lang string) error {

	sendgridRequest := PrepareEmail(data["mailTo"], data, templateKey, defaultSubject, defaultText, lang)

	shouldSend, err := shouldSendEmail(uid, data["email"], data["key"])
	if err != nil {
		return err
	}

	if shouldSend {
		slog.Debug("Sending",
			slog.String("key", data["key"]),
			slog.String("email", data["email"]),
			slog.String("mailTo", data["mailTo"]))
		lastMailTo = sendgridRequest.MailTo

		request := SendEmailRequest{
			SendgridRequest: sendgridRequest,
			Url:             e.emailUrl,
			EmailFromName:   e.emailFromName,
			EmailFrom:       e.emailFrom,
			EmailToken:      e.emailToken,
		}
		sendEmail(&request)

	} else {
		slog.Debug("Not sending",
			slog.String("key", data["key"]),
			slog.String("email", data["email"]),
			slog.String("mailTo", data["mailTo"]))
	}

	return nil
}

//******************** These are called by the application

func (e *EmailClient) SendMarketingEmail(email string, balanceMap map[string]*big.Int, repoNames []string) error {
	params := map[string]string{}
	params["email"] = email
	//don't spam in testing...
	if e.emailMarketing != "live" {
		email = e.emailMarketing
	}
	params["mailTo"] = email
	params["url"] = e.emailLinkPrefix
	params["lang"] = "en"
	weekly := int(util.TimeNow().Unix() / WaitToSendEmail)
	params["key"] = KeyMarketing + params["email"] + strconv.Itoa(weekly)

	return e.prepareSendEmail(
		nil,
		params,
		KeyMarketing,
		"[Marketing] Someone Likes Your Contribution for "+fmt.Sprint(repoNames),
		"Thanks for keep building and maintaining "+fmt.Sprint(repoNames)+". Someone sponsored you with "+
			util.PrintMap(balanceMap)+". \nGo to "+params["url"]+" and claim your support!",
		params["lang"])
}

func (e *EmailClient) SendStripeTopUp(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"

	key := KeyTopUpStripe
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c != nil {
		key += c.String()
	}
	params["key"] = key

	return e.prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpStripe,
		"We are about to top up your account",
		"Thanks for supporting with flatfeestack: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendTopUpSponsor(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"

	//we are sponsor, and the user beneficiaryEmail could not donate
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("cannot have date as nil %v", c)
	}
	params["key"] = c.String() + KeyTopUpOther

	return e.prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpOther,
		"Your invited users could not sponsor anymore",
		"Please add funds at: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendTopUpInvited(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"

	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("cannot have date as nil %v", c)
	}
	params["key"] = c.String() + KeyTopUpUser1

	return e.prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpUser1,
		u.Email+" (and you) are running low on funds",
		"Please add funds at: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendTopUpOther(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"

	key := KeyTopUpUser2
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c != nil {
		key += c.String()
	}
	params["key"] = key

	return e.prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpUser2,
		"You are running low on funding",
		"Please add funds at: "+params["url"],
		params["lang"])

}

func (e *EmailClient) SendAddGit(userId uuid.UUID, email string, addGitEmailToken string, lang string) error {
	var params = map[string]string{}
	params["token"] = addGitEmailToken
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/confirm/git-email/" + email + "/" + addGitEmailToken
	params["lang"] = lang
	params["key"] = KeyAddGit + email

	return e.prepareSendEmail(
		&userId,
		params,
		KeyAddGit,
		"Validate your git email",
		"Is this your email address? Please confirm: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendPaymentNowFinished(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowFinished + data.OrderId.String()

	return e.prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowFinished,
		"Payment successful",
		"Crypto payment successful. See your payment here: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendPaymentNowPartially(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowPartially + data.OrderId.String()

	defaultMessage := fmt.Sprintf("Only partial payment received (%v) of (%v), please send the rest (%v) to: ", data.ActuallyPaid, data.PayAmount, data.PayAmount-data.ActuallyPaid)
	return e.prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowPartially,
		"Partially paid",
		defaultMessage,
		params["lang"])
}

func (e *EmailClient) SendPaymentNowRefunded(userId uuid.UUID, status string, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowRefunded + status + "-" + externalId.String()

	defaultMessage := fmt.Sprintf("Payment %v, please check payment: %s", status, params["url"])
	return e.prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowRefunded,
		"Payment "+status,
		defaultMessage,
		params["lang"])
}

func (e *EmailClient) SendStripeSuccess(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeSuccess + externalId.String()

	return e.prepareSendEmail(
		&userId,
		params,
		KeyStripeSuccess,
		"Payment successful",
		"Payment successful. See your payment here: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendStripeAction(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeAction + externalId.String()

	return e.prepareSendEmail(
		&user.Id,
		params,
		KeyStripeAction,
		"Authentication requested",
		"Action is required, please go to the following site to continue: "+params["url"],
		params["lang"])
}

func (e *EmailClient) SendStripeFailed(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = e.emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeFailed + externalId.String()

	return e.prepareSendEmail(
		&user.Id,
		params,
		KeyStripeFailed,
		"Insufficient funds",
		"Your credit card transfer failed. If you have enough funds, please go to: "+params["url"],
		params["lang"])
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
