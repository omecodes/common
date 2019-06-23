package mail

import (
	"crypto/tls"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/log"
	"gopkg.in/gomail.v2"
)

func Send(cfg conf.Map, to string, subject string, html string, plain string) error {
	server, ok := cfg.GetString("server")
	if !ok {
		log.E("Mailer", errors.BadInput, "missing 'server' entry in configs")
		return errors.BadInput
	}
	port, ok := cfg.GetInt32("port")
	if !ok {
		log.E("Mailer", errors.BadInput, "missing 'port' entry in configs")
		return errors.BadInput
	}

	user, ok := cfg.GetString("user")
	if !ok {
		log.E("Mailer", errors.BadInput, "missing 'user' entry in configs")
		return errors.BadInput
	}
	password, ok := cfg.GetString("password")
	if !ok {
		log.E("Mailer", errors.BadInput, "missing 'password' entry in configs")
		return errors.BadInput
	}
	return SendMail(server, int(port), user, password, to, subject, html, plain)
}

func SendMail(server string, port int, user string, password string, to string, subject string, html string, plain string, files ...string) error {
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
