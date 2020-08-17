package mailer

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGrid struct {
	host     string
	endpoint string
	key      string
	from     string
	fromName string
}

func (s *sendGrid) Send(email *Email) error {
	from := mail.NewEmail(s.fromName, s.from)
	receiver := mail.NewEmail(email.To.Name, email.To.Email)
	message := mail.NewSingleEmail(from, email.Subject, receiver, email.Plain, email.Html)
	client := sendgrid.NewSendClient(s.key)
	response, err := client.Send(message)

	if err != nil {
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return nil
}
