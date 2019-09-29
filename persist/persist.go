package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

func scanBytes(row dao.Row) (interface{}, error) {
	var bytes []byte
	return bytes, row.Scan(&bytes)
}

// Dict is a convenience for persistence dict
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
	o, err := d.QueryOne("select", "bytes", key)
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

func NewDBDict(dbConf conf.Map, prefix string) (Dict, error) {
	d := new(dictDB)
	d.SetTablePrefix(prefix).
		AddTableDefinition("map", "create table if not exists $prefix$_dict (name varchar(255) not null primary key, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_dict values (?, ?);").
		AddStatement("update", "update $prefix$_dict set val=? where name=?;").
		AddStatement("select", "select val from $prefix$_dict where name=?;").
		AddStatement("delete", "delete from $prefix$_dict where name=?;").
		AddStatement("clear", "delete from $prefix$_dict;").
		RegisterScanner("nonce", dao.NewScannerFunc(scanBytes))
	err := d.Init(dbConf)
	return d, err
}

// List is a convenience for persistence list
type Sequence interface {
	Append(data []byte) error
	Get(index int64) ([]byte, error)
	GetAfter(index int64) (dao.Cursor, error)
	Clear() error
	Close() error
}

type listDB struct {
	dao.SQL
}

func (l *listDB) Append(data []byte) error {
	return l.Exec("insert", data).Error
}

func (l *listDB) Get(index int64) ([]byte, error) {
	o, err := l.QueryOne("select", "bytes", index)
	if err != nil {
		return nil, err
	}
	return o.([]byte), err
}

func (l *listDB) GetAfter(index int64) (dao.Cursor, error) {
	return l.Query("select_from", "bytes", index)
}

func (l *listDB) Clear() error {
	return l.Exec("clear").Error
}

func (l *listDB) Close() error {
	return l.DB.Close()
}

func NewDBList(dbConf conf.Map, prefix string) (Sequence, error) {
	d := new(listDB)
	d.SetTablePrefix(prefix).
		AddTableDefinition("map", "create table if not exists $prefix$_list (ind int not null primary key $auto_increment$, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_list (val) values (?);").
		AddStatement("select", "select val from $prefix$_list where ind=?;").
		AddStatement("select_from", "select val from $prefix$_list where ind>?;").
		AddStatement("clear", "delete from $prefix$_list;").
		RegisterScanner("bytes", dao.NewScannerFunc(scanBytes))
	err := d.Init(dbConf)
	return d, err
}
