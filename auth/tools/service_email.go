package tools

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"dsoob/backend/include"
)

type LocalsEmailVerify struct {
	Token string
}
type LocalsLoginForgotPassword struct {
	Token string
}
type LocalsLoginNewLocation struct {
	Token          string
	Timestamp      string
	IpAddress      string
	DeviceBrowser  string
	DeviceLocation string
}
type LocalsLoginNewDevice struct {
	Timestamp      string
	IpAddress      string
	DeviceBrowser  string
	DeviceLocation string
}
type LocalsLoginPasscode struct {
	Code     string
	Lifetime string
}
type LocalsNotifyUserDeleted struct {
	Reason string
}
type LocalsNotifyUserEmailModified struct{}
type LocalsNotifyUserPasswordModified struct{}

var (
	EmailVerify                     = setupEmailTemplate[LocalsEmailVerify]( /*---------------*/ "EMAIL_VERIFY", "Verify your Email Address")
	EmailLoginForgotPassword        = setupEmailTemplate[LocalsLoginForgotPassword]( /*-------*/ "LOGIN_FORGOT_PASSWORD", "Forgot Your Password?")
	EmailLoginNewLocation           = setupEmailTemplate[LocalsLoginNewLocation]( /*----------*/ "LOGIN_NEW_LOCATION", "Allow Login from a New Location")
	EmailLoginNewDevice             = setupEmailTemplate[LocalsLoginNewDevice]( /*------------*/ "LOGIN_NEW_DEVICE", "Login from a New Device")
	EmailLoginPasscode              = setupEmailTemplate[LocalsLoginPasscode]( /*-------------*/ "LOGIN_PASSCODE", "Your One Time Passcode")
	EmailNotifyUserDeleted          = setupEmailTemplate[LocalsNotifyUserDeleted]( /*---------*/ "NOTIFY_USER_DELETED", "Account Deleted")
	EmailNotifyUserEmailModified    = setupEmailTemplate[LocalsNotifyUserEmailModified]( /*---*/ "NOTIFY_USER_EMAIL_MODIFIED", "Your Account Password has Changed")
	EmailNotifyUserPasswordModified = setupEmailTemplate[LocalsNotifyUserPasswordModified]( /**/ "NOTIFY_USER_PASS_MODIFIED", "Your Account Email has Changed")
)

func setupEmailTemplate[L any](filename, subjectLine string) func(toAddress string, locals L) {

	template, err := template.ParseFS(
		include.Templates, "templates/"+filename+".txt")
	if err != nil {
		panic("cannot parse template: " + err.Error())
	}

	return func(toAddress string, locals L) {

		// Render Content
		var content bytes.Buffer
		literals := map[string]any{
			"Host": SITE_NAME,
			"Data": locals,
		}
		if err := template.Execute(&content, literals); err != nil {
			LoggerEmail.Data(ERROR, "Render Failed", map[string]any{
				"address":  toAddress,
				"template": filename,
				"locals":   locals,
				"error":    err,
			})
			return
		}

		// Send Email
		boundary := "bunny-bunny-bunny-bunny-bunny"
		envelope := bytes.Buffer{}

		fmt.Fprintf(&envelope, "From: %s\r\n", EMAIL_SMTP_ADDRESS)
		fmt.Fprintf(&envelope, "To: %s\r\n", toAddress)
		fmt.Fprintf(&envelope, "Subject: %s\r\n", subjectLine)
		fmt.Fprintf(&envelope, "MIME-Version: 1.0\r\n")
		fmt.Fprintf(&envelope, "Content-Type: multipart/alternative; boundary=%s\r\n", boundary)
		fmt.Fprintf(&envelope, "\r\n--%s\r\n", boundary)
		fmt.Fprintf(&envelope, "Content-Type: text/plain; charset=utf-8\r\n\r\n")
		fmt.Fprintf(&envelope, "%s\r\n", content.String())
		fmt.Fprintf(&envelope, "--%s--\r\n", boundary)

		err := smtp.SendMail(EMAIL_SMTP_HOST, nil, EMAIL_SMTP_ADDRESS, []string{toAddress}, envelope.Bytes())
		errObject := map[string]any{}
		if err != nil {
			errObject["message"] = err.Error()
			errObject["error"] = err
		}

		LoggerEmail.Data(INFO, "Email Info", map[string]any{
			"template": map[string]any{
				"filename": filename,
				"literals": literals,
			},
			"destination": map[string]any{
				"email_address": toAddress,
			},
			"error": errObject,
		})

	}
}
