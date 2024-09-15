package mailer

import (
	"os"
	"testing"
)

func TestSendMail(t *testing.T) {
	sender, err := NewSender(os.Getenv("email"), os.Getenv("AppPass"), "smtp.mail.ru", 587)
	if err != nil {
		t.Error(err)
		return
	}
	err = sender.Send(&EmailInput{
		To:      os.Getenv("email"),
		Subject: "TEST Message",
		Body:    "TEST MESSGE",
	})
	if err != nil {
		t.Error(err)
		return
	}
}
