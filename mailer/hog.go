package mailer

type hogClient struct {
	*defaultMailer
}

func (h *hogClient) Send(to, subject, contentType, content string, files ...string) error {
	return h.defaultMailer.Send(to, subject, contentType, content, files...)
}
