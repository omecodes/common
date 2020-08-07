package mailer

import (
	"crypto/tls"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/jcon"
	"github.com/xo/dburl"
	"gopkg.in/gomail.v2"
	"strconv"
)

type Mailer interface {
	Send(to, subject, contentType, content string, files ...string) error
}

func Get(cfg jcon.Map) (Mailer, error) {

	t, ok := cfg.GetString("type")
	if !ok {
		return nil, errors.New("unsupported mail type: " + t)
	}

	if t == "sendgrid" {
		m := &sendGrid{}
		m.host, ok = cfg.GetString("host")
		if ok {
			m.endpoint, ok = cfg.GetString("endpoint")
			if ok {
				m.key, ok = cfg.GetString("key")
			}
		}

		if !ok {
			return nil, errors.New("wrong sendgrid mailer config. Missing some items")
		}
		return m, nil
	}

	dm := &defaultMailer{}

	dm.server, ok = cfg.GetString("server")
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

	if t == "hog" {
		return &hogClient{
			defaultMailer: dm,
		}, nil
	}

	return dm, nil
}

func Parse(dsn string) (Mailer, error) {
	u, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}

	dm := &defaultMailer{}

	dm.user = u.User.Username()
	dm.password, _ = u.User.Password()
	dm.server = u.Host
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	dm.port = int32(port)
	return dm, nil
}

func sendToSMTPServer(server string, port int, user, password, to, subject, contentType, content string, files ...string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "zoenion.services@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody(contentType, content)

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
