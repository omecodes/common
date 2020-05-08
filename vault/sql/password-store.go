package sql

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/persist/dict"
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

func (p *passwordStore) Get(username string) (string, error) {
	var wrapper stringWrapper
	err := p.db.Read(username, &wrapper)
	return wrapper.Value, err
}

func (p *passwordStore) Delete(username string) error {
	return p.db.Delete(username)
}

func NewPasswordStore(cfg conf.Map, tableNamePrefix string) (*passwordStore, error) {
	db, err := dict.NewSQL(cfg, tableNamePrefix, codec.Default)
	if err != nil {
		return nil, err
	}
	return &passwordStore{db: db}, nil
}
