package persist

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/dao"
)

// List is a convenience for persistence list
type Sequence interface {
	Append(data []byte) error
	Get(index int64) (string, error)
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

func (l *listDB) Get(index int64) (string, error) {
	o, err := l.QueryOne("select", "bytes", index)
	if err != nil {
		return "", err
	}
	return o.(string), err
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
		RegisterScanner("scanner", dao.NewScannerFunc(scanData))
	err := d.Init(dbConf)
	return d, err
}
