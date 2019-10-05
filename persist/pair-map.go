package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

type MappedPairs interface {
	Set(firstKey, secondKey string, val []byte) error
	Get(firstKey, secondKey string) ([]byte, error)
	GetAll(primaryKey string) (map[string][]byte, error)
	Delete(firstKey, secondKey string) error
	DeleteAll(firstKey string) error
	Clear() error
	Close() error
}

type sqlPairMap struct {
	dao.SQL
}

func (s *sqlPairMap) Set(firstKey, secondKey string, val []byte) error {
	if s.Exec("insert", firstKey, secondKey, val).Error != nil {
		return s.Exec("update", val, firstKey, secondKey).Error
	}
	return nil
}

func (s *sqlPairMap) Get(firstKey, secondKey string) ([]byte, error) {
	o, err := s.QueryOne("select", "scanner", firstKey, secondKey)
	if err != nil {
		return nil, err
	}
	return o.([]byte), nil
}

func (s *sqlPairMap) GetAll(primaryKey string) (map[string][]byte, error) {
	result := map[string][]byte{}
	c, err := s.Query("select_by_first_key", "scanner", primaryKey)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	for c.HasNext() {
		o, err := c.Next()
		if err != nil {
			return nil, err
		}
		p := o.(*pair)
		result[p.Key] = p.Value
	}
	return result, nil
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

func NewMappedPairs(dbConf conf.Map, prefix string) (MappedPairs, error) {
	d := new(sqlPairMap)
	d.SetTablePrefix(prefix).
		AddTableDefinition("mapped_pairs", "create table if not exists $prefix$_data (first_key varchar(255) not null, second_key varchar(255) not null, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_data values (?, ?, ?);").
		AddStatement("update", "update $prefix$_data set val=? where first_key=? and second_key=?;").
		AddStatement("select", "select val from $prefix$_data where first_key=? and second_key=?;").
		AddStatement("select_by_first_key", "select second_key, val from $prefix$_data where first_key=?;").
		AddStatement("delete", "delete from $prefix$_data where name=?;").
		AddStatement("delete_by_first_key", "delete from $prefix$_data where name=?;").
		AddStatement("clear", "delete from $prefix$_data;").
		RegisterScanner("nonce", dao.NewScannerFunc(scanBytes))
	err := d.Init(dbConf)
	if err != nil {
		return nil, err
	}

	err = d.AddUniqueIndex(dao.SQLIndex{Name: "unique_keys", Table: "$prefix$_data"}, false)
	return d, err
}
