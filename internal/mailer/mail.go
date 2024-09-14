package mailer

import (
	"fmt"
	"log"

	"github.com/go-gomail/gomail"
)

type Sender struct {
	from string
	pass string
	host string
	port int
}

type EmailInput struct {
	To      string
	Subject string
	Body    string
}

func NewSender(from, pass, host string, port int) (*Sender, error) {
	return &Sender{from: from, pass: pass, host: host, port: port}, nil
}

func (s *Sender) Send(input *EmailInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", input.To)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	dialer := gomail.NewDialer(s.host, s.port, s.from, s.pass)
	log.Printf("Attempting to send email %s:%d", s.host, s.port)
	if err := dialer.DialAndSend(msg); err != nil {
		log.Printf("Error sending email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Println("Message sent successfully")

	return nil
}
