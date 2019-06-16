package mail

import (
	"crypto/tls"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"gopkg.in/gomail.v2"
	"log"
)

func Send(cfg conf.Map, to string, subject string, html string, plain string) error {
	server, ok := cfg.GetString("server")
	if !ok {
		return errors.Detailed(errors.BadRequest, "missing 'server' entry in configs")
	}
	port, ok := cfg.GetInt32("port")
	if !ok {
		return errors.Detailed(errors.BadRequest, "missing 'port' entry in configs")
	}

	user, ok := cfg.GetString("user")
	if !ok {
		return errors.Detailed(errors.BadRequest, "missing 'user' entry in configs")
	}
	password, ok := cfg.GetString("password")
	if !ok {
		return errors.Detailed(errors.BadRequest, "missing 'password' entry in configs")
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
		log.Println("email", err.Error())
		return err
	}
	return nil
}
