package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func sendEmail(url string, e EmailRequest) error {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var jsonData []byte
	var err error
	if strings.Contains(url, "sendgrid") {
		sendGridReq := mail.NewSingleEmail(
			mail.NewEmail(opts.EmailFromName, opts.EmailFrom),
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("could not send email: %v %v", resp.Status, resp.StatusCode)
	}
	return nil
}

func prepareEmail(mailTo string, data map[string]interface{}, templateName string, defaultSubject string,
	defaultText string, lang string) EmailRequest {
	subject := parseTemplate("subject/"+lang+templateName+".txt", data)
	if subject == "" {
		subject = defaultSubject
	}
	textMessage := parseTemplate("plain/"+lang+templateName+".txt", data)
	if textMessage == "" {
		textMessage = defaultText
	}

	headerTemplate := parseTemplate("html/"+lang+"/header.html", data)
	footerTemplate := parseTemplate("html/"+lang+"/footer.html", data)
	htmlBody := parseTemplate("html/"+lang+templateName+".html", data)
	htmlMessage := headerTemplate + htmlBody + footerTemplate

	return EmailRequest{
		MailTo:      mailTo,
		Subject:     subject,
		TextMessage: textMessage,
		HtmlMessage: htmlMessage,
	}
}
