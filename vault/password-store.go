package vault

type PasswordStore interface {
	Save(username, password string) error
	Update(username, odlPassword, newPassword string) error
	Get(username string) (string, error)
	Delete(username string) error
	Verify(username, password string) (bool, error)
}
