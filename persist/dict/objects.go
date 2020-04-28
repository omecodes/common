package dict

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
	"github.com/zoenion/common/errors"
)

type Dict interface {
	Save(key string, o interface{}) error
	Read(key string, o interface{}) error
	Contains(key string) (bool, error)
	Delete(key string) error
	List() (Cursor, error)
	Clear() error
	Close() error
}

type sqlObjects struct {
	dao.SQL
	objectCodec codec.Codec
}

func (s *sqlObjects) Save(key string, o interface{}) error {
	data, err := s.objectCodec.Encode(o)
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

	r := o.(*Row)
	return s.objectCodec.Decode([]byte(r.encoded), object)
}

func (s *sqlObjects) Contains(key string) (bool, error) {
	res, err := s.QueryOne("contains", "", key)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return res.(bool), nil
}

func (s *sqlObjects) Delete(key string) error {
	return s.Exec("delete", key).Error
}

func (s *sqlObjects) List() (Cursor, error) {
	c, err := s.Query("select_all", "scanner")
	if err != nil {
		return nil, err
	}
	return newCursor(c, s.objectCodec), nil
}

func (s *sqlObjects) Clear() error {
	return s.Exec("clear").Error
}

func (s *sqlObjects) Close() error {
	return s.DB.Close()
}

func NewSQL(cfg conf.Map, tablePrefix string, codec codec.Codec) (Dict, error) {
	d := new(sqlObjects)
	d.SetTablePrefix(tablePrefix).
		AddTableDefinition("map", "create table if not exists $prefix$_mapping (name varchar(255) not null primary key, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where name=?;").
		AddStatement("select", "select * from $prefix$_mapping where name=?;").
		AddStatement("select_all", "select * from $prefix$_mapping;").
		AddStatement("contains", "select 1 from $prefix$_mapping where name=?;").
		AddStatement("delete", "delete from $prefix$_mapping where name=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(scanRow)).
		RegisterScanner("bool", dao.NewScannerFunc(scanBool))

	err := d.Init(cfg)
	d.objectCodec = codec
	return d, err
}
