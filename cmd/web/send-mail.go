// Command web implements an asynchronous mail-sending loop for the application.
// It consumes messages from the global app.MailChan and delivers them using
// a local SMTP server via the go-simple-mail library.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bensabler/milos-residence/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

// listenForMail starts a background goroutine that continuously reads messages
// from app.MailChan and dispatches them using sendMsg.
//
// Behavior:
//   - Blocks on app.MailChan, ensuring backpressure when the channel is full.
//   - Each received MailData is handed to sendMsg for SMTP delivery.
//
// Usage:
//
//	// During startup after app.MailChan is created:
//	go listenForMail()
func listenForMail() {
	go func() {
		for {
			// Pull the next queued email and send it.
			msg := <-app.MailChan
			sendMsg(msg)
		}
	}()
}

// sendMsg builds and sends a single email message through an SMTP server.
//
// Parameters:
//   - m: models.MailData containing sender, recipient, subject, message body,
//     and an optional template name.
//
// Behavior:
//   - Resolves SMTP host and port from environment variables MAIL_HOST and
//     MAIL_PORT, defaulting to "localhost" and "1025" when unset.
//   - Configures a go-simple-mail SMTP client with 10-second connect/send
//     timeouts and no persistent connections (KeepAlive=false).
//   - Establishes a connection to the SMTP server.
//   - Constructs a new email message and sets From, To, and Subject headers.
//   - If m.Template is empty, sets the raw HTML body to m.Content.
//   - If m.Template is provided, reads the template file from
//     ./email-templates/, replaces the [%body%] placeholder with m.Content,
//     and uses the resulting HTML as the body.
//   - Attempts to send the email, logging any connection or send errors to
//     errorLog and the standard logger.
//
// Notes:
//   - Designed for development and testing with MailHog or a similar SMTP
//     catcher. Adjust host, port, and security settings for production use.
//
// Usage:
//   sendMsg(models.MailData{From: "noreply@example.com", To: "user@example.com",
//       Subject: "Welcome!", Content: "<p>Hello!</p>"})
func sendMsg(m models.MailData) {
	// Resolve SMTP host and port from environment variables or use defaults.
	host := os.Getenv("MAIL_HOST")
	if host == "" {
		host = "localhost"
	}
	portStr := os.Getenv("MAIL_PORT")
	if portStr == "" {
		portStr = "1025"
	}
	port, _ := strconv.Atoi(portStr)

	// Configure the SMTP client with development-friendly defaults.
	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Attempt to establish a connection to the SMTP server.
	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	// Create the email message and set standard headers.
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	// Determine body source: direct content or template substitution.
	if m.Template == "" {
		// No template provided; send raw HTML content.
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// Template specified; read file and replace [%body%] placeholder.
		data, err := os.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}
		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	// Attempt to send the email and log the outcome.
	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
