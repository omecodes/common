package vault

type PasswordStore interface {
	Save(username, password string) error
	Get(username string) (string, error)
	Delete(username string) error
}
