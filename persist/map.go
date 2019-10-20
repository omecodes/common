package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

// Dict is a convenience for persistence dict
type Dict interface {
	Set(key string, data string) error
	Get(key string) (string, error)
	GetAll() (dao.Cursor, error)
	Del(key string) error
	Clear() error
	Close() error
}

type dictDB struct {
	dao.SQL
}

func (d *dictDB) Set(key string, data string) error {
	err := d.Exec("insert", key, data).Error
	if err != nil {
		err = d.Exec("update", data, key).Error
	}
	return err
}

func (d *dictDB) Get(key string) (string, error) {
	o, err := d.QueryOne("select", "data_scanner", key)
	if err != nil {
		return "", err
	}
	return o.(string), nil
}

func (d *dictDB) GetAll() (dao.Cursor, error) {
	return d.Query("select_all", "pair_scanner")
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
		AddTableDefinition("map", "create table if not exists $prefix$_mapping (name varchar(255) not null primary key, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where name=?;").
		AddStatement("select", "select val from $prefix$_mapping where name=?;").
		AddStatement("delete", "delete from $prefix$_mapping where name=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(scanData)).
		RegisterScanner("pair_scanner", dao.NewScannerFunc(scanPair)).
		RegisterScanner("data_scanner", dao.NewScannerFunc(scanData))
	err := d.Init(dbConf)
	return d, err
}
