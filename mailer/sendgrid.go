package mailer

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/zoenion/common/jcon"
)

type sendGrid struct {
	host     string
	endpoint string
	key      string
}

func (s *sendGrid) Send(to, subject, contentType string, content string, files ...string) error {
	request := sendgrid.GetRequest(s.key, s.endpoint, s.host)
	request.Method = "POST"

	body := jcon.Map{
		"personalizations": []jcon.Map{
			{"to": []jcon.Map{{"email": to}}, "subject": subject},
		},
		"from": jcon.Map{
			"email": "",
		},
		"content": []jcon.Map{{"type": contentType, "value": content}},
	}

	bodyBytes, err := body.Bytes()
	if err != nil {
		return err
	}

	request.Body = bodyBytes
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	fmt.Println(response.StatusCode)
	fmt.Println(response.Body)
	fmt.Println(response.Headers)
	return nil
}

func NewSendGrid(key, endpoint, host string) {}
