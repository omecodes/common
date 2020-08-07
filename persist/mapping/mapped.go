package mapping

import (
	"database/sql"
	"github.com/omecodes/common/codec"
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/jcon"
)

type DoubleMap interface {
	Set(firstKey, secondKey string, o interface{}) error
	Get(firstKey, secondKey string, o interface{}) error
	GetForFirst(firstKey string) (MapCursor, error)
	GetForSecond(secondKey string) (MapCursor, error)
	GetAll() (Cursor, error)
	Delete(firstKey, secondKey string) error
	DeleteAllMatchingFirstKey(firstKey string) error
	DeleteAllMatchingSecondKey(secondKey string) error
	Clear() error
	Close() error
}

type sqlPairMap struct {
	dao.SQL
	codec codec.Codec
}

func (s *sqlPairMap) Set(firstKey, secondKey string, o interface{}) error {
	val, err := s.codec.Encode(o)
	if err != nil {
		return err
	}

	if s.Exec("insert", firstKey, secondKey, string(val)).Error != nil {
		return s.Exec("update", string(val), firstKey, secondKey).Error
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

func (s *sqlPairMap) GetForFirst(firstKey string) (MapCursor, error) {
	c, err := s.Query("select_by_first_key", "pair_scanner", firstKey)
	if err != nil {
		return nil, err
	}
	return newMapCursor(c, s.codec), nil
}

func (s *sqlPairMap) GetForSecond(secondKey string) (MapCursor, error) {
	c, err := s.Query("select_by_second_key", "pair_scanner", secondKey)
	if err != nil {
		return nil, err
	}
	return newMapCursor(c, s.codec), nil
}

func (s *sqlPairMap) GetAll() (Cursor, error) {
	c, err := s.Query("select_all", "scanner")
	if err != nil {
		return nil, err
	}
	return newCursor(c, s.codec), nil
}

func (s *sqlPairMap) Delete(firstKey, secondKey string) error {
	return s.Exec("delete", firstKey, secondKey).Error
}

func (s *sqlPairMap) DeleteAllMatchingFirstKey(firstKey string) error {
	return s.Exec("delete_by_first_key", firstKey).Error
}

func (s *sqlPairMap) DeleteAllMatchingSecondKey(secondKey string) error {
	return s.Exec("delete_by_first_key", secondKey).Error
}

func (s *sqlPairMap) Clear() error {
	return s.Exec("clear").Error
}

func (s *sqlPairMap) Close() error {
	return s.DB.Close()
}

func New(dbConf jcon.Map, prefix string, codec codec.Codec) (DoubleMap, error) {
	d := new(sqlPairMap)
	d.SetTablePrefix(prefix).
		AddTableDefinition("mapped_pairs", "create table if not exists $prefix$_mapping (first_key varchar(255) not null, second_key varchar(255) not null, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where first_key=? and second_key=?;").
		AddStatement("select", "select * from $prefix$_mapping where first_key=? and second_key=?;").
		AddStatement("select_by_first_key", "select second_key, val from $prefix$_mapping where first_key=?;").
		AddStatement("select_by_second_key", "select first_key, val from $prefix$_mapping where second_key=?;").
		AddStatement("select_all", "select * from $prefix$_mapping;").
		AddStatement("delete", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_first_key", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_second_key", "delete from $prefix$_mapping where second_key=?;").
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

func NewSQL(dialect string, db *sql.DB, prefix string, cdc codec.Codec) (DoubleMap, error) {
	d := new(sqlPairMap)
	d.SetTablePrefix(prefix).
		AddTableDefinition("mapped_pairs", "create table if not exists $prefix$_mapping (first_key varchar(255) not null, second_key varchar(255) not null, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where first_key=? and second_key=?;").
		AddStatement("select", "select * from $prefix$_mapping where first_key=? and second_key=?;").
		AddStatement("select_by_first_key", "select second_key, val from $prefix$_mapping where first_key=?;").
		AddStatement("select_by_second_key", "select first_key, val from $prefix$_mapping where second_key=?;").
		AddStatement("select_all", "select * from $prefix$_mapping;").
		AddStatement("delete", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_first_key", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_second_key", "delete from $prefix$_mapping where second_key=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(sqlScanRow)).
		RegisterScanner("pair_scanner", dao.NewScannerFunc(sqlScanPair))

	err := d.InitWithSqlDB(dialect, db)
	if err != nil {
		return nil, err
	}
	d.codec = cdc

	err = d.AddUniqueIndex(dao.SQLIndex{Name: "unique_keys", Table: "$prefix$_mapping", Fields: []string{"first_key", "second_key"}}, false)
	return d, err
}
