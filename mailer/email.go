package mailer

type User struct {
	Name  string
	Email string
}

type Email struct {
	Subject string
	From    User
	To      User
	Plain   string
	Html    string
	Files   []string
}
