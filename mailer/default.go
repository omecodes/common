package mailer

type defaultMailer struct {
	from                   string
	server, user, password string
	port                   int32
}

func (m *defaultMailer) Send(email *Email) error {
	return sendToSMTPServer(
		m.server,
		int(m.port),
		m.user,
		m.password,
		m.from, email.To.Email, email.Subject, "", email.Html, email.Files...)
}
