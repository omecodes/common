package data

import (
	"github.com/zoenion/common/dao"
	"github.com/zoenion/common/database"
)

type Dict interface {
	Set(key string, val []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Clear() error
	Close() error
}

type dictDB struct {
	dao.SQL
}

func (d *dictDB) Set(key string, data []byte) error {
	err := d.Exec("insert", key, data).Error
	if err != nil {
		err = d.Exec("update", data, key).Error
	}
	return err
}

func (d *dictDB) Get(key string) ([]byte, error) {
	o, err := d.QueryOne("select", "nonce", key)
	if err != nil {
		return nil, err
	}
	return o.([]byte), nil
}

func (d *dictDB) Del(key string) error {
	return d.Exec("delete", key).Error
}

func (d *dictDB) Clear() error {
	return d.Exec("clear").Error
}

func (d *dictDB) Close() error {
	return d.DB.Close()
}

func (d *dictDB) scanNonce(row dao.Row) (interface{}, error) {
	var bytes []byte
	return bytes, row.Scan(&bytes)
}

func NewDictDB(filename string) (Dict, error) {
	d := new(dictDB)
	d.AddTableDefinition("map", "create table if not exists dict (key varchar(255) not null primary key, val long blob not null);").
		AddStatement("insert", "insert into dict values (?, ?);").
		AddStatement("update", "update dict set val=? where key=?;").
		AddStatement("select", "select val from dict where key=?;").
		AddStatement("delete", "delete from dict where key=?;").
		AddStatement("clear", "delete from dict;").
		RegisterScanner("nonce", dao.NewScannerFunc(d.scanNonce))
	err := d.Init(database.SQLiteConfig(filename))
	return d, err
}
