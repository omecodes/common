package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

// List is a convenience for persistence list
type Sequence interface {
	Append(rawStr string) error
	Get(index int64) (string, error)
	MinIndex() (int64, error)
	MaxIndex() (int64, error)
	Count() (int64, error)
	GetAfter(index int64) (dao.Cursor, error)
	Clear() error
	Close() error
}

type listDB struct {
	dao.SQL
}

func (l *listDB) Append(rawStr string) error {
	return l.Exec("insert", rawStr).Error
}

func (l *listDB) Get(index int64) (string, error) {
	o, err := l.QueryOne("select", "scanner", index)
	if err != nil {
		return "", err
	}
	return o.(string), err
}

func (l *listDB) MinIndex() (int64, error) {
	res, err := l.QueryOne("select_min_index", "index")
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

func (l *listDB) MaxIndex() (int64, error) {
	res, err := l.QueryOne("select_min_index", "index")
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

func (l *listDB) Count() (int64, error) {
	res, err := l.QueryOne("select_count", "index")
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

func (l *listDB) GetAfter(index int64) (dao.Cursor, error) {
	return l.Query("select_from", "scanner", index)
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
		AddStatement("select_min_index", "select min(ind) from $prefix$_list;").
		AddStatement("select_max_index", "select max(ind) from $prefix$_list;").
		AddStatement("select_count", "select count(ind) from $prefix$_list;").
		AddStatement("select_from", "select val from $prefix$_list where ind>?;").
		AddStatement("clear", "delete from $prefix$_list;").
		RegisterScanner("scanner", dao.NewScannerFunc(scanData)).
		RegisterScanner("index", dao.NewScannerFunc(scanInt))
	err := d.Init(dbConf)
	return d, err
}
