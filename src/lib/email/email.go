package lib_email

import (
	lib_error "app/src/lib/error"
	"bytes"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"strings"
)

// Error codes for Email Utility
var ErrorCode = struct {
	Unknown         int
	AuthError       int
	SendError       int
	EmptyReceiver   int
	EmptySubject    int
	EmptyBody       int
	InvalidSettings int
}{
	Unknown:         0,
	AuthError:       1,
	SendError:       2,
	EmptyReceiver:   3,
	EmptySubject:    4,
	EmptyBody:       5,
	InvalidSettings: 6,
}

// Settings holds SMTP Server settings
type Settings struct {
	Username string
	Password string
	Host     string
	Port     string
}

// Map of SMTP Server setting configurations
var SMTPSettings = map[string]*Settings{
	"test": {
		Username: "tauriktester@gmail.com",
		Password: "WtN9gDQKWfcrXFg",
		Host:     "smtp.gmail.com",
		Port:     "587",
	},
	"sale": {
		Username: "tokenlinksupport@platinum-egg.com",
		Password: "faJSBWeN",
		Host:     "smtp.gmail.com",
		Port:     "587",
	},
}

// EmailUtil holds information for Auth and SMTP Server settings
type EmailUtil struct {
	Auth     smtp.Auth
	Settings *Settings
}

// EmailRequest holds the information for preparing and sending the email
type EmailRequest struct {
	to       []string
	subject  string
	body     string
	host     string
	username string
	auth     *smtp.Auth
}

// EmailUtil used to create new EmailRequest in order to send emails
func NewEmailUtil(user string) (*EmailUtil, error) {
	// Get Auth and SMTP Server settings and return a new EmailUtil
	auth, settings, err := setupClient(user)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	emailUtil := &EmailUtil{
		Auth:     auth,
		Settings: settings,
	}
	return emailUtil, nil
}

func IsValidEmail(address string) bool {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return false
	}
	containsBannedChar := strings.Contains(address, "+")
	if containsBannedChar {
		return false
	}
	return true
}

// function for retrieving the SMTP Server settings and Auth for EmailUtil struct
func setupClient(user string) (smtp.Auth, *Settings, error) {
	// Get the settings that correspond to the user arg
	var settings *Settings = nil
	if set, exists := SMTPSettings[user]; exists {
		settings = set
	} else {
		return nil, nil, lib_error.NewAppError(ErrorCode.InvalidSettings, "SMTP Server Settings not found")
	}
	// Get auth, if auth returns nil the authentication failed
	auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
	if auth == nil {
		return nil, nil, lib_error.NewAppError(ErrorCode.AuthError, "Auth failed")
	}
	return auth, settings, nil
}

// Returns a new EmailRequest.
//Email Request holds methods and information to parse template files for the email body and to send the email itself
func (self *EmailUtil) NewEmailRequest(to []string, subject, body string) (*EmailRequest, error) {
	req := &EmailRequest{
		to:       to,
		subject:  subject,
		body:     body,
		host:     fmt.Sprintf("%s:%s", self.Settings.Host, self.Settings.Port),
		username: self.Settings.Username,
		auth:     &self.Auth,
	}
	return req, nil
}

// This parses the template file, assigns the data to the template variables,
// and sets the string representation of the parsed template to the email body
func (self *EmailRequest) ParseTemplate(templateName string, data interface{}) error {
	// get the template from the html file
	template, err := template.ParseFiles(templateName)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// template.Execute needs a writer so we declare a bytes.Buffer for it
	buf := new(bytes.Buffer)
	err = template.Execute(buf, data)
	if err != nil {
		return lib_error.WrapError(err)
	}
	// set the string representation of the parsed template to body
	self.body = buf.String()
	return nil
}

// This prepares and sends the email from EmailRequest
func (self *EmailRequest) SendMail() error {
	// Check for empty to/subject/body
	if len(self.to) == 0 {
		return lib_error.NewAppError(ErrorCode.EmptyReceiver, "Not enough recipients")
	}
	if self.subject == "" {
		return lib_error.NewAppError(ErrorCode.EmptySubject, "Subject is empty")
	}
	if self.body == "" {
		return lib_error.NewAppError(ErrorCode.EmptyBody, "Body is empty")
	}
	// Join the to addresses. This is for the To: field in the email header.
	tos := strings.Join(self.to, ",")
	// Set the mime type for the email header. This is needed for displaying html
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	//  Format the email according to RFC 822-style. CRLF is required.
	msgFormatted := fmt.Sprintf("To:%s\r\nSubject:%s\r\n%s\r\n\r\n%s\r\n", tos, self.subject, mime, self.body)
	msg := []byte(msgFormatted)
	// Send the email
	err := smtp.SendMail(self.host, *self.auth, self.username, self.to, msg)
	if err != nil {
		return lib_error.NewAppError(ErrorCode.SendError, err.Error())
	}
	return nil
}
