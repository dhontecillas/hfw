package mailer

import (
	"fmt"
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
