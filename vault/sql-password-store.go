package vault

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/omecodes/common/codec"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/jcon"
	"github.com/omecodes/common/persist/dict"
)

type stringWrapper struct {
	Value string
}

type passwordStore struct {
	db dict.Dict
}

func (p *passwordStore) hexHashed(password string) string {
	hashedPasswordBytes := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hashedPasswordBytes[:])
}

func (p *passwordStore) Save(username, password string) error {
	return p.db.Save(username, &stringWrapper{Value: p.hexHashed(password)})
}

func (p *passwordStore) Update(username, oldPassword, newPassword string) error {
	ok, err := p.Verify(username, oldPassword)
	if err != nil {
		return err
	}

	if !ok {
		return errors.NotFound
	}

	return p.Save(username, newPassword)
}

func (p *passwordStore) Get(username string) (string, error) {
	var wrapper stringWrapper
	err := p.db.Read(username, &wrapper)
	return wrapper.Value, err
}

func (p *passwordStore) Delete(username string) error {
	return p.db.Delete(username)
}

func (p *passwordStore) Verify(username, password string) (bool, error) {
	hexPassword := p.hexHashed(password)
	savedPassword, err := p.Get(username)
	if err != nil {
		return false, err
	}
	return hexPassword == savedPassword, nil
}

func NewPasswordStore(cfg jcon.Map, tableNamePrefix string) (*passwordStore, error) {
	db, err := dict.New(cfg, tableNamePrefix, codec.Default)
	if err != nil {
		return nil, err
	}
	return &passwordStore{db: db}, nil
}

func NewSQLPasswordStore(db *sql.DB, tableNamePrefix string) (*passwordStore, error) {
	d, err := dict.NewSQL(db, tableNamePrefix, codec.Default)
	if err != nil {
		return nil, err
	}
	return &passwordStore{db: d}, nil
}
