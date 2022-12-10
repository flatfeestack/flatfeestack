package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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
	queue                chan *EmailRequest
	emailNotifications   = 0
	emailNoNotifications = 0
	lastMailTo           = ""
)

func init() {
	queue = make(chan *EmailRequest)
	go func() {
		for {
			select {
			case e := <-queue:
				sendEmailQueue(e)
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
		c, err = countEmailSentById(*uId, key)
	} else {
		c, err = countEmailSentByEmail(email, key)
	}
	if err != nil {
		return false, err
	}

	if c > 0 {
		log.Printf("we already sent a notification %v", email)
		emailNoNotifications++
		return false, nil
	}
	err = insertEmailSent(uId, email, key, timeNow())
	if err != nil {
		return false, err
	}
	emailNotifications++
	lastMailTo = email
	return true, nil
}

func sendEmailQueue(e *EmailRequest) error {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var jsonData []byte
	var err error
	if strings.Contains(opts.EmailUrl, "sendgrid") {
		sendGridReq := NewSingleEmailPlainText(
			NewEmail(opts.EmailFromName, opts.EmailFrom),
			e.Subject,
			NewEmail("", e.MailTo),
			e.TextMessage)
		jsonData, err = json.Marshal(sendGridReq)
	} else {
		jsonData, err = json.Marshal(e)
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", opts.EmailUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+opts.EmailToken)
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

	subject := parseTemplate("template-subject-"+templateKey+"_"+lang+".tmpl", data)
	if subject == "" {
		subject = defaultSubject
	}
	textMessage := parseTemplate("template-plain-"+templateKey+"_"+lang+".tmpl", data)
	if textMessage == "" {
		textMessage = defaultText
	}
	htmlMessage := parseTemplate("template-html-"+templateKey+"_"+lang+".tmpl", data)

	e := EmailRequest{
		MailTo:      data["mailTo"],
		Subject:     subject,
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
		if opts.Env != "test" {
			sendEmail(&e)
		}
	} else {
		log.Debugf("not sending %v email to %v/%v", data["key"], data["email"], data["mailTo"])
	}

	return nil
}

//******************** These are called by the application

func sendMarketingEmail(email string, balanceMap map[string]*big.Int, repoNames []string) error {
	var other = map[string]string{}
	other["email"] = email
	//dont spam in testing...
	if opts.EmailMarketing != "live" {
		email = opts.EmailMarketing
	}
	other["mailTo"] = email
	other["url"] = opts.EmailLinkPrefix
	other["lang"] = "en"
	weekly := int(timeNow().Unix() / WaitToSendEmail)
	other["key"] = KeyMarketing + other["email"] + strconv.Itoa(weekly)

	return prepareEmail(
		nil,
		other,
		KeyMarketing,
		"[Marketing] Someone Likes Your Contribution for "+fmt.Sprint(repoNames),
		"Thanks for keep building and maintaining "+fmt.Sprint(repoNames)+". Someone sponsored you with "+
			printMap(balanceMap)+". \nGo to "+other["url"]+" and claim your support!",
		other["lang"])
}

func sendStripeTopUp(u User) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	key := KeyTopUpStripe
	if u.PaymentCycleInId != nil {
		key += u.PaymentCycleInId.String()
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

func sendTopUpSponsor(u User) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"

	//we are sponser, and the user beneficiaryEmail could not donate
	other["key"] = u.PaymentCycleInId.String() + KeyTopUpOther
	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpOther,
		"Your invited users could not sponsor anymore",
		"Please add funds at: "+other["url"],
		other["lang"])
}

func sendTopUpInvited(u User) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = u.PaymentCycleInId.String() + KeyTopUpUser1
	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpUser1,
		u.Email+" (and you) are running low on funds",
		"Please add funds at: "+other["url"],
		other["lang"])
}

func sendTopUpOther(u User) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = u.PaymentCycleInId.String() + KeyTopUpUser2
	return prepareEmail(
		&u.Id,
		other,
		KeyTopUpUser2,
		"You are running low on funding",
		"Please add funds at: "+other["url"],
		other["lang"])

}

func sendAddGit(email string, addGitEmailToken string, lang string) error {
	var other = map[string]string{}
	other["token"] = addGitEmailToken
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/confirm/git-email/" + email + "/" + addGitEmailToken
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

func sendPaymentNowFinished(user *User, data WebhookResponse) error {
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
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

func sendPaymentNowPartially(user User, data WebhookResponse) error {
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
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

func sendPaymentNowRefunded(user User, data WebhookResponse, status string) error {
	email := user.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyPaymentNowRefunded + status + "-" + data.OrderId.String()

	defaultMessage := fmt.Sprintf("Payment %v, please check payment: %s", data.PaymentStatus, other["url"])
	return prepareEmail(
		&user.Id,
		other,
		KeyPaymentNowRefunded,
		"Payment "+data.PaymentStatus,
		defaultMessage,
		other["lang"])
}

func sendStripeSuccess(u User, newPaymentCycleInId *uuid.UUID) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeSuccess + newPaymentCycleInId.String()

	return prepareEmail(
		&u.Id,
		other,
		KeyStripeSuccess,
		"Payment successful",
		"Payment successful. See your payment here: "+other["url"],
		other["lang"])
}

func sendStripeAction(u User, newPaymentCycleInId *uuid.UUID) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeAction + newPaymentCycleInId.String()

	return prepareEmail(
		&u.Id,
		other,
		KeyStripeAction,
		"Authentication requested",
		"Action is required, please go to the following site to continue: "+other["url"],
		other["lang"])
}

func sendStripeFailed(u User, newPaymentCycleInId *uuid.UUID) error {
	email := u.Email
	var other = map[string]string{}
	other["mailTo"] = email
	other["email"] = email
	other["url"] = opts.EmailLinkPrefix + "/user/payments"
	other["lang"] = "en"
	other["key"] = KeyStripeFailed + newPaymentCycleInId.String()

	return prepareEmail(
		&u.Id,
		other,
		KeyStripeFailed,
		"Insufficient funds",
		"Your credit card transfer failed. If you have enough funds, please go to: "+other["url"],
		other["lang"])
}
