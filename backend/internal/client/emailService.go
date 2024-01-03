package client

import (
	"backend/db"
	"backend/pkg/util"
	"fmt"
	mail "github.com/flatfeestack/go-lib/email"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
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
	queue                chan *mail.SendEmailRequest
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
	queue = make(chan *mail.SendEmailRequest)
	go func() {
		for {
			select {
			case e := <-queue:
				err := mail.SendEmail(*e)
				if err != nil {
					log.Errorf("cannot send email %v", err)
				}
			}
		}
	}()

}

func sendEmail(e *mail.SendEmailRequest) {
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
	err = db.InsertEmailSent(id, uId, email, key, util.TimeNow())
	if err != nil {
		return false, err
	}
	EmailNotifications++
	lastMailTo = email
	return true, nil
}

func prepareSendEmail(
	uid *uuid.UUID,
	data map[string]string,
	templateKey string,
	defaultSubject string,
	defaultText string,
	lang string) error {

	sendgridRequest := mail.PrepareEmail(data["mailTo"], data, templateKey, defaultSubject, defaultText, lang)

	shouldSend, err := shouldSendEmail(uid, data["email"], data["key"])
	if err != nil {
		return err
	}

	if shouldSend {
		log.Debugf("sending %v email to %v/%v", data["key"], data["email"], data["mailTo"])
		lastMailTo = sendgridRequest.MailTo
		if env != "local" {
			request := mail.SendEmailRequest{
				SendgridRequest: sendgridRequest,
				Url:             emailUrl,
				EmailFromName:   emailFromName,
				EmailFrom:       emailFrom,
				EmailToken:      emailToken,
			}
			sendEmail(&request)
		}
	} else {
		log.Debugf("not sending %v email to %v/%v", data["key"], data["email"], data["mailTo"])
	}

	return nil
}

//******************** These are called by the application

func SendMarketingEmail(email string, balanceMap map[string]*big.Int, repoNames []string) error {
	params := map[string]string{}
	params["email"] = email
	//dont spam in testing...
	if emailMarketing != "live" {
		email = emailMarketing
	}
	params["mailTo"] = email
	params["url"] = emailLinkPrefix
	params["lang"] = "en"
	weekly := int(util.TimeNow().Unix() / WaitToSendEmail)
	params["key"] = KeyMarketing + params["email"] + strconv.Itoa(weekly)

	return prepareSendEmail(
		nil,
		params,
		KeyMarketing,
		"[Marketing] Someone Likes Your Contribution for "+fmt.Sprint(repoNames),
		"Thanks for keep building and maintaining "+fmt.Sprint(repoNames)+". Someone sponsored you with "+
			util.PrintMap(balanceMap)+". \nGo to "+params["url"]+" and claim your support!",
		params["lang"])
}

func SendStripeTopUp(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
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

	return prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpStripe,
		"We are about to top up your account",
		"Thanks for supporting with flatfeestack: "+params["url"],
		params["lang"])
}

func SendTopUpSponsor(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
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

	return prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpOther,
		"Your invited users could not sponsor anymore",
		"Please add funds at: "+params["url"],
		params["lang"])
}

func SendTopUpInvited(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"

	_, _, _, c, err := db.FindLatestDailyPayment(u.Id, "USD")
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("cannot have date as nil %v", c)
	}
	params["key"] = c.String() + KeyTopUpUser1

	return prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpUser1,
		u.Email+" (and you) are running low on funds",
		"Please add funds at: "+params["url"],
		params["lang"])
}

func SendTopUpOther(u db.UserDetail) error {
	email := u.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
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

	return prepareSendEmail(
		&u.Id,
		params,
		KeyTopUpUser2,
		"You are running low on funding",
		"Please add funds at: "+params["url"],
		params["lang"])

}

func SendAddGit(userId uuid.UUID, email string, addGitEmailToken string, lang string) error {
	var params = map[string]string{}
	params["token"] = addGitEmailToken
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/confirm/git-email/" + email + "/" + addGitEmailToken
	params["lang"] = lang
	params["key"] = KeyAddGit + email

	return prepareSendEmail(
		&userId,
		params,
		KeyAddGit,
		"Validate your git email",
		"Is this your email address? Please confirm: "+params["url"],
		params["lang"])
}

func SendPaymentNowFinished(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowFinished + data.OrderId.String()

	return prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowFinished,
		"Payment successful",
		"Crypto payment successful. See your payment here: "+params["url"],
		params["lang"])
}

func SendPaymentNowPartially(userId uuid.UUID, data WebhookResponse) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowPartially + data.OrderId.String()

	defaultMessage := fmt.Sprintf("Only partial payment received (%v) of (%v), please send the rest (%v) to: ", data.ActuallyPaid, data.PayAmount, data.PayAmount-data.ActuallyPaid)
	return prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowPartially,
		"Partially paid",
		defaultMessage,
		params["lang"])
}

func SendPaymentNowRefunded(userId uuid.UUID, status string, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyPaymentNowRefunded + status + "-" + externalId.String()

	defaultMessage := fmt.Sprintf("Payment %v, please check payment: %s", status, params["url"])
	return prepareSendEmail(
		&user.Id,
		params,
		KeyPaymentNowRefunded,
		"Payment "+status,
		defaultMessage,
		params["lang"])
}

func SendStripeSuccess(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeSuccess + externalId.String()

	return prepareSendEmail(
		&userId,
		params,
		KeyStripeSuccess,
		"Payment successful",
		"Payment successful. See your payment here: "+params["url"],
		params["lang"])
}

func SendStripeAction(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeAction + externalId.String()

	return prepareSendEmail(
		&user.Id,
		params,
		KeyStripeAction,
		"Authentication requested",
		"Action is required, please go to the following site to continue: "+params["url"],
		params["lang"])
}

func SendStripeFailed(userId uuid.UUID, externalId uuid.UUID) error {
	user, err := db.FindUserById(userId)
	if err != nil {
		return err
	}
	email := user.Email
	var params = map[string]string{}
	params["mailTo"] = email
	params["email"] = email
	params["url"] = emailLinkPrefix + "/user/payments"
	params["lang"] = "en"
	params["key"] = KeyStripeFailed + externalId.String()

	return prepareSendEmail(
		&user.Id,
		params,
		KeyStripeFailed,
		"Insufficient funds",
		"Your credit card transfer failed. If you have enough funds, please go to: "+params["url"],
		params["lang"])
}
