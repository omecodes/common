package persist

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

type Objects interface {
	Save(key string, o interface{}) error
	Read(key string, o interface{}) error
	Delete(key string) error
	List() (dao.Cursor, error)
	DecoderFunc() func(data []byte, o interface{}) error
	Clear() error
	Close() error
}

type sqlObjects struct {
	dao.SQL
}

func (s *sqlObjects) Save(key string, o interface{}) error {
	data, err := codec.GSONEncode(o)
	if err != nil {
		return err
	}
	err = s.Exec("insert", key, string(data)).Error
	if err != nil {
		err = s.Exec("update", string(data), key).Error
	}
	return err
}

func (s *sqlObjects) Read(key string, object interface{}) error {
	o, err := s.QueryOne("select", "scanner", key)
	if err != nil {
		return err
	}
	return codec.GSONDecode([]byte(o.(string)), object)
}

func (s *sqlObjects) Delete(key string) error {
	return s.Exec("delete", key).Error
}

func (s *sqlObjects) List() (dao.Cursor, error) {
	return s.Query("select_all", "scanner")
}

func (s *sqlObjects) DecoderFunc() func(data []byte, o interface{}) error {
	return codec.GSONDecode
}

func (s *sqlObjects) Clear() error {
	return s.Exec("clear").Error
}

func (s *sqlObjects) Close() error {
	return s.DB.Close()
}

func NewSQLObjectsDB(cfg conf.Map, tablePrefix string) (Objects, error) {
	d := new(sqlObjects)
	d.SetTablePrefix(tablePrefix).
		AddTableDefinition("map", "create table if not exists $prefix$_mapping (name varchar(255) not null primary key, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where name=?;").
		AddStatement("select", "select val from $prefix$_mapping where name=?;").
		AddStatement("select_all", "select val from $prefix$_mapping;").
		AddStatement("delete", "delete from $prefix$_mapping where name=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(scanData)).
		RegisterScanner("pair_scanner", dao.NewScannerFunc(scanPair)).
		RegisterScanner("data_scanner", dao.NewScannerFunc(scanData))
	err := d.Init(cfg)
	return d, err
}
