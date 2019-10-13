package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

type MappedPairs interface {
	Set(firstKey, secondKey string, val string) error
	Get(firstKey, secondKey string) (string, error)
	GetForFirst(firstKey string) (map[string]string, error)
	GetAll() (dao.Cursor, error)
	Delete(firstKey, secondKey string) error
	DeleteAll(firstKey string) error
	Clear() error
	Close() error
}

type sqlPairMap struct {
	dao.SQL
}

func (s *sqlPairMap) Set(firstKey, secondKey string, val string) error {
	if s.Exec("insert", firstKey, secondKey, val).Error != nil {
		return s.Exec("update", val, firstKey, secondKey).Error
	}
	return nil
}

func (s *sqlPairMap) Get(firstKey, secondKey string) (string, error) {
	o, err := s.QueryOne("select", "scanner", firstKey, secondKey)
	if err != nil {
		return "", err
	}
	return o.(string), nil
}

func (s *sqlPairMap) GetForFirst(primaryKey string) (map[string]string, error) {
	result := map[string]string{}
	c, err := s.Query("select_by_first_key", "pair_scanner", primaryKey)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	for c.HasNext() {
		o, err := c.Next()
		if err != nil {
			return nil, err
		}
		p := o.(*Pair)
		result[p.First] = p.Second
	}
	return result, nil
}

func (s *sqlPairMap) GetAll() (dao.Cursor, error) {
	return s.Query("select", "triplet_scanner")
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
		AddTableDefinition("mapped_pairs", "create table if not exists $prefix$_mapping (first_key varchar(255) not null, second_key varchar(255) not null, val longblob not null);").
		AddStatement("insert", "insert into $prefix$_mapping values (?, ?, ?);").
		AddStatement("update", "update $prefix$_mapping set val=? where first_key=? and second_key=?;").
		AddStatement("select", "select val from $prefix$_mapping where first_key=? and second_key=?;").
		AddStatement("select_by_first_key", "select second_key, val from $prefix$_mapping where first_key=?;").
		AddStatement("select_all", "select * from $prefix$_mapping;").
		AddStatement("delete", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("delete_by_first_key", "delete from $prefix$_mapping where first_key=?;").
		AddStatement("clear", "delete from $prefix$_mapping;").
		RegisterScanner("scanner", dao.NewScannerFunc(scanData)).
		RegisterScanner("pair_scanner", dao.NewScannerFunc(scanPair)).
		RegisterScanner("triplet_scanner", dao.NewScannerFunc(scanTriplet))

	err := d.Init(dbConf)
	if err != nil {
		return nil, err
	}

	err = d.AddUniqueIndex(dao.SQLIndex{Name: "unique_keys", Table: "$prefix$_mapping", Fields: []string{"first_key", "second_key"}}, false)
	return d, err
}
