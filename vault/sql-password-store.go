package vault

import (
	"database/sql"
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

func (p *passwordStore) Save(username, password string) error {
	return p.db.Save(username, &stringWrapper{Value: password})
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
	savedPassword, err := p.Get(username)
	if err != nil {
		return false, err
	}
	return password == savedPassword, nil
}

func NewPasswordStore(cfg jcon.Map, tableNamePrefix string) (*passwordStore, error) {
	db, err := dict.New(cfg, tableNamePrefix, codec.Default)
	if err != nil {
		return nil, err
	}
	return &passwordStore{db: db}, nil
}

func NewMySQLPasswordStore(db *sql.DB, tableNamePrefix string) (*passwordStore, error) {
	d, err := dict.NewSQL("mysql", db, tableNamePrefix, codec.Default)
	if err != nil {
		return nil, err
	}
	return &passwordStore{db: d}, nil
}
