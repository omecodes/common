package mapped

import (
	"github.com/zoenion/common/codec"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

type Doubled interface {
	Set(firstKey, secondKey string, val string) error
	Get(firstKey, secondKey string, o interface{}) error
	GetForFirst(firstKey string) (MapCursor, error)
	GetAll() (Cursor, error)
	Delete(firstKey, secondKey string) error
	DeleteAll(firstKey string) error
	Clear() error
	Close() error
}

type sqlPairMap struct {
	dao.SQL
	codec codec.Codec
}

func (s *sqlPairMap) Set(firstKey, secondKey string, val string) error {
	if s.Exec("insert", firstKey, secondKey, val).Error != nil {
		return s.Exec("update", val, firstKey, secondKey).Error
	}
	return nil
}

func (s *sqlPairMap) Get(firstKey, secondKey string, o interface{}) error {
	i, err := s.QueryOne("select", "scanner", firstKey, secondKey)
	if err != nil {
		return err
	}

	row := i.(*Row)
	return s.codec.Decode([]byte(row.encoded), o)
}

func (s *sqlPairMap) GetForFirst(primaryKey string) (MapCursor, error) {
	c, err := s.Query("select_by_first_key", "pair_scanner", primaryKey)
	if err != nil {
		return nil, err
	}
	return newMapCursor(c, s.codec), nil
}

func (s *sqlPairMap) GetAll() (Cursor, error) {
	c, err := s.Query("select", "triplet_scanner")
	if err != nil {
		return nil, err
	}
	return newCursor(c, s.codec), nil
}

func (s *sqlPairMap) Delete(firstKey, secondKey string) error {
	return s.Exec("delete", firstKey, secondKey).Error
}

func (s *sqlPairMap) DeleteAll(firstKey string) error {
	return s.Exec("delete_by_first_key", firstKey).Error
}

func (s *sqlPairMap) Clear() error {
	return s.Exec("clear").Error
}

func (s *sqlPairMap) Close() error {
	return s.DB.Close()
}

func NewSQL(dbConf conf.Map, prefix string, codec codec.Codec) (Doubled, error) {
	d := new(sqlPairMap)
	d.SetTablePrefix(prefix).
		AddTableDefinition("mapped_pairs", "create table if not exists $prefix$_mapping (first_key varchar(255) not null, second_key varchar(255) not null, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where first_key=? and second_key=?;").
		AddStatement("select", "select * from $prefix$_mapping where first_key=? and second_key=?;").
		AddStatement("select_by_first_key", "select second_key, val from $prefix$_mapping where first_key=?;").
		AddStatement("select_all", "select * from $prefix$_mapping;").
		AddStatement("delete", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_first_key", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(sqlScanRow)).
		RegisterScanner("pair_scanner", dao.NewScannerFunc(sqlScanPair))

	err := d.Init(dbConf)
	if err != nil {
		return nil, err
	}
	d.codec = codec

	err = d.AddUniqueIndex(dao.SQLIndex{Name: "unique_keys", Table: "$prefix$_mapping", Fields: []string{"first_key", "second_key"}}, false)
	return d, err
}
