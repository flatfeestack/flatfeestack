package clients

import (
	db "backend/db"
	"backend/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
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
	queue                chan *EmailRequest
	EmailNotifications   = 0
	EmailNoNotifications = 0
	lastMailTo           = ""
)

var (
	emailUrl        string
	emailFromName   string
	emailFrom       string
	emailToken      string
	env             string
	emailMarketing  string
	emailLinkPrefix string
)

type EmailRequest struct {
	MailTo      string `json:"mail_to,omitempty"`
	Subject     string `json:"subject"`
	TextMessage string `json:"text_message"`
	HtmlMessage string `json:"html_message"`
}

func InitEmail(emailUrl0 string, emailFromName0 string, emailFrom0 string, emailToken0 string, env0 string, emailMarketing0 string, emailLinkPrefix0 string) {
	emailUrl = emailUrl0
	emailFromName = emailFromName0
	emailFrom = emailFrom0
	emailToken = emailToken0
	env = env0
	emailMarketing = emailMarketing0
	emailLinkPrefix = emailLinkPrefix0
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
	queue = make(chan *EmailRequest)
	go func() {
		for {
			select {
			case e := <-queue:
				err := sendEmailQueue(e)
				if err != nil {
					log.Errorf("cannot send email %v", err)
				}
			}
		}
	}()

}

func sendEmail(e *EmailRequest) {
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
		log.Printf("we already sent a notification %v", email)
		EmailNoNotifications++
		return false, nil
	}
	id := uuid.New()
	err = db.InsertEmailSent(id, uId, email, key, utils.TimeNow())
	if err != nil {
		return false, err
	}
	EmailNotifications++
	lastMailTo = email
	return true, nil
}

func sendEmailQueue(e *EmailRequest) error {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var jsonData []byte
	var err error
	if strings.Contains(emailUrl, "sendgrid") {
		sendGridReq := mail.NewSingleEmail(
			mail.NewEmail(emailFromName, emailFrom),
			e.Subject,
			mail.NewEmail("", e.MailTo),
			e.TextMessage,
			e.HtmlMessage)
		jsonData, err = json.Marshal(sendGridReq)
	} else {
		jsonData, err = json.Marshal(e)
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", emailUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+emailToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("could not send the email: %v %v", resp.Status, resp.StatusCode)
	}
	return nil
}

func prepareEmail(
	uid *uuid.UUID,
	data map[string]string,
	templateKey string,
	defaultSubject string,
	defaultText string,
	lang string) error {

	textMessage := utils.ParseTemplate("plain/"+lang+"/"+templateKey+".txt", data)
	if textMessage == "" {
		textMessage = defaultText
	}
	headerTemplate := utils.ParseTemplate("html/"+lang+"/header.html", data)
	footerTemplate := utils.ParseTemplate("html/"+lang+"/footer.html", data)
	htmlBody := utils.ParseTemplate("html/"+lang+"/"+templateKey+".html", data)
	htmlMessage := headerTemplate + htmlBody + footerTemplate

	e := EmailRequest{
		MailTo:      data["mailTo"],
		Subject:     defaultSubject,
		TextMessage: textMessage,
		HtmlMessage: htmlMessage,
	}

	b, err := shouldSendEmail(uid, data["email"], data["key"])
	if err != nil {
		return err
	}

	if b {
		log.Debugf("sending %v email to %v/%v", data["key"], data["email"], data["mailTo"])
		lastMailTo = e.MailTo
		if env != "local" {
			sendEmail(&e)
		}
	} else {
		log.Debugf("not sending %v email to %v/%v", data["key"], data["email"], data["mailTo"])
	}

	return nil
}

//******************** These are called by the application

func SendMarketingEmail(email string, balanceMap map[string]*big.Int, repoNames []string) error {
	var other = map[string]string{}
	other["email"] = email
	//dont spam in testing...
	if emailMarketing != "live" {
		email = emailMarketing
	}
	other["mailTo"] = email
	other["url"] = emailLinkPrefix
	other["lang"] = "en"
	weekly := int(utils.TimeNow().Unix() / WaitToSendEmail)
	other["key"] = KeyMarketing + other["email"] + strconv.Itoa(weekly)

	return prepareEmail(
		nil,
		other,
		KeyMarketing,
		"[Marketing] Someone Likes Your Contribution for "+fmt.Sprint(repoNames),
		"Thanks for keep building and maintaining "+fmt.Sprint(repoNames)+". Someone sponsored you with "+
			utils.PrintMap(balanceMap)+". \nGo to "+other["url"]+" and claim your support!",
		other["lang"])
}

func SendStripeTopUp(u db.UserDetail) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"

	key := KeyTopUpStripe
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c != nil {
		key += c.String()
	}
	other["key"] = key

	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpStripe,
		"We are about to top up your account",
		"Thanks for supporting with flatfeestack: "+other["url"],
		other["lang"])
}

func SendTopUpSponsor(u db.UserDetail) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"

	//we are sponsor, and the user beneficiaryEmail could not donate
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("cannot have date as nil %v", c)
	}
	other["key"] = c.String() + KeyTopUpOther

	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpOther,
		"Your invited users could not sponsor anymore",
		"Please add funds at: "+other["url"],
		other["lang"])
}

func SendTopUpInvited(u db.UserDetail) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"

	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("cannot have date as nil %v", c)
	}
	other["key"] = c.String() + KeyTopUpUser1

	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpUser1,
		u.Email+" (and you) are running low on funds",
		"Please add funds at: "+other["url"],
		other["lang"])
}

func SendTopUpOther(u db.UserDetail) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"

	key := KeyTopUpUser2
	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c != nil {
		key += c.String()
	}
	other["key"] = key

	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpUser2,
		"You are running low on funding",
		"Please add funds at: "+other["url"],
		other["lang"])

}

func SendAddGit(email string, addGitEmailToken string, lang string) error {
	var other = map[string]string{}
	other["token"] = addGitEmailToken
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/confirm/git-email/" + email + "/" + addGitEmailToken
	other["lang"] = lang
	other["key"] = KeyAddGit + email

	return prepareEmail(
		nil,
		other,
		KeyAddGit,
		"Validate your git email",
		"Is this your email address? Please confirm: "+other["url"],
		other["lang"])
}

func SendPaymentNowFinished(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyPaymentNowFinished + data.OrderId.String()

	return prepareEmail(
		&user.Id,
		other,
		KeyPaymentNowFinished,
		"Payment successful",
		"Crypto payment successful. See your payment here: "+other["url"],
		other["lang"])
}

func SendPaymentNowPartially(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyPaymentNowPartially + data.OrderId.String()

	defaultMessage := fmt.Sprintf("Only partial payment received (%v) of (%v), please send the rest (%v) to: ", data.ActuallyPaid, data.PayAmount, data.PayAmount-data.ActuallyPaid)
	return prepareEmail(
		&user.Id,
		other,
		KeyPaymentNowPartially,
		"Partially paid",
		defaultMessage,
		other["lang"])
}

func SendPaymentNowRefunded(userId uuid.UUID, status string, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyPaymentNowRefunded + status + "-" + externalId.String()

	defaultMessage := fmt.Sprintf("Payment %v, please check payment: %s", status, other["url"])
	return prepareEmail(
		&user.Id,
		other,
		KeyPaymentNowRefunded,
		"Payment "+status,
		defaultMessage,
		other["lang"])
}

func SendStripeSuccess(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeSuccess + externalId.String()

	return prepareEmail(
		&userId,
		other,
		KeyStripeSuccess,
		"Payment successful",
		"Payment successful. See your payment here: "+other["url"],
		other["lang"])
}

func SendStripeAction(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeAction + externalId.String()

	return prepareEmail(
		&user.Id,
		other,
		KeyStripeAction,
		"Authentication requested",
		"Action is required, please go to the following site to continue: "+other["url"],
		other["lang"])
}

func SendStripeFailed(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = emailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeFailed + externalId.String()

	return prepareEmail(
		&user.Id,
		other,
		KeyStripeFailed,
		"Insufficient funds",
		"Your credit card transfer failed. If you have enough funds, please go to: "+other["url"],
		other["lang"])
}
