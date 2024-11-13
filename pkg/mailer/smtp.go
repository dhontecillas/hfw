package mailer

import (
	"fmt"
	"net/smtp"
	"time"
)

const mailMsgTemplate string = "To: %s\r\n" +
	"From: %s\r\n" +
	"Date: %s\r\n" +
	"Message-Id: <%s>\r\n" +
	"Subject: %s\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Transfer-Encoding: 8bit;\r\n" +
	"Content-Type: multipart/mixed; boundary=%s\r\n\r\n" +
	"--%s\r\n" +
	"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
	"%s\r\n" +
	"--%s\r\n" +
	"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
	"%s\r\n" +
	"--%s--\r\n"

// ComposeSMTPMsg builds the content to be sent through an SMTP server
func ComposeSMTPMsg(e Email) string {
	now := time.Now()
	frontier := fmt.Sprintf("multipart-separator-%d", now.Unix())
	msgID := fmt.Sprintf("%d%s", now.Unix(), e.From.Address)
	return fmt.Sprintf(mailMsgTemplate, e.To, e.From, now.Format(time.RFC1123), msgID,
		e.Subject, frontier, frontier,
		e.Text, frontier, e.HTML, frontier)
}

// SMTPConfig has the configurations to use the SMTP service
type SMTPConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	SenderName    string
	SenderAddress string
}

type smtpMailer struct {
	conf SMTPConfig
}

// NewSMTPMailer instantiates a new SMTP mailer.
func NewSMTPMailer(conf SMTPConfig) (Mailer, error) {
	if len(conf.User) == 0 {
		return nil, fmt.Errorf("missing SMTP user")
	}
	if len(conf.Password) == 0 {
		return nil, fmt.Errorf("missing SMTP password")
	}
	// for server and port, we can fall back to defaults:
	if len(conf.Host) == 0 {
		return nil, fmt.Errorf("missing Host")
	}
	if conf.Port == 0 {
		conf.Port = 587
	}
	if len(conf.SenderAddress) == 0 {
		return nil, fmt.Errorf("missing sender Address")
	}
	if len(conf.SenderName) == 0 {
		conf.SenderName = conf.SenderAddress
	}
	return &smtpMailer{
		conf: conf,
	}, nil
}

// Sends an email to SMTP service
func (m *smtpMailer) Send(e Email) error {
	address := fmt.Sprintf("%s:%d", m.conf.Host, m.conf.Port)
	auth := smtp.PlainAuth("", m.conf.User, m.conf.Password, m.conf.Host)
	msg := ComposeSMTPMsg(e)
	err := smtp.SendMail(address, auth, e.From.Address, []string{e.To.Address}, []byte(msg))
	return err
}

// Sender returns the default sender address and name
func (m *smtpMailer) Sender() (string, string) {
	return m.conf.SenderAddress, m.conf.SenderName
}
