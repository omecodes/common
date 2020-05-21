package mailer

type defaultMailer struct {
	server, user, password string
	port                   int32
}

func (m *defaultMailer) Send(to, subject, contentType, content string, files ...string) error {
	return sendToSMTPServer(m.server, int(m.port), m.user, m.password, to, subject, contentType, content, files...)
}
