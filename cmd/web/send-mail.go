// Command web implements an asynchronous mail-sending loop for the application.
// It consumes messages from the global app.MailChan and delivers them using
// a local SMTP server via the go-simple-mail library.
package main

import (
	"fmt"
	"log"
	"os"
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

// sendMsg builds and sends a single email message through a local SMTP server.
//
// Parameters:
//   - m: MailData containing sender, recipient, subject, body, and optional
//     template name.
//
// Behavior:
//   - Connects to an SMTP server at localhost:1025 with 10-second timeouts.
//   - Uses m.Template if provided, reading the template file from
//     ./email-templates/ and replacing the [%body%] placeholder with m.Content.
//   - Falls back to sending m.Content directly if no template is specified.
//   - Logs any connection or send errors to app.ErrorLog and the standard logger.
//
// Notes:
//   - This implementation assumes a development mail catcher (e.g., MailHog)
//     listening on localhost:1025. Adjust host/port and security settings for
//     production use.
func sendMsg(m models.MailData) {
	// Configure SMTP client (development defaults).
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	// Establish SMTP connection.
	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	// Construct email message.
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	// Select body: raw content or template substitution.
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// Read the specified template file and replace placeholder.
		data, err := os.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}
		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	// Attempt to send the message.
	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
