package mailer

type hogClient struct {
	*defaultMailer
}

func (h *hogClient) Send(email *Email) error {
	return h.defaultMailer.Send(email)
}
