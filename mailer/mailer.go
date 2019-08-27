package mailer

import (
	"crypto/tls"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(to string, subject string, plain string, html string, files ...string) error
}

type defaultMailer struct {
	server, user, password string
	port                   int32
}

func NewMailer(cfg conf.Map) (*defaultMailer, error) {
	dm := &defaultMailer{}
	var ok bool

	dm.server, ok = cfg.GetString("host")
	if !ok {
		return nil, errors.BadInput
	}
	dm.port, ok = cfg.GetInt32("port")
	if !ok {
		return nil, errors.BadInput
	}

	dm.user, ok = cfg.GetString("user")
	if !ok {
		return nil, errors.BadInput
	}

	dm.password, ok = cfg.GetString("password")
	if !ok {
		return nil, errors.BadInput
	}
	return dm, nil
}

func (m *defaultMailer) Send(to string, subject string, plain string, html string, files ...string) error {
	return sendMail(m.server, int(m.port), m.user, m.password, to, subject, plain, html, files...)
}

func sendMail(server string, port int, user string, password string, to string, subject string, plain string, html string, files ...string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "zoenion.services@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if len(html) > 0 {
		m.SetBody("text/html", html)
	} else {
		m.SetBody("text/plain", plain)
	}

	for i := range files {
		m.Attach(files[i])
	}

	d := gomail.NewDialer(server, port, user, password)
	d.TLSConfig = &tls.Config{
		ServerName: server,
	}

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
