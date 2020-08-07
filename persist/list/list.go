package list

import (
	"database/sql"
	"github.com/omecodes/common/codec"
	"github.com/omecodes/common/dao"
	"github.com/omecodes/common/jcon"
	filespb "github.com/omecodes/common/proto/files"
)

// State is a convenience for persistence list
type List interface {
	Append(o interface{}) error
	GetAt(index int64, o interface{}) error
	GetNextFromSeq(index int64, o interface{}) (int64, error)
	GetAllFromSeq(index int64) (Cursor, error)
	Delete(index int64) error
	MinIndex() (int64, error)
	MaxIndex() (int64, error)
	Count() (int64, error)
	Clear() error
	Close() error
}

type listDB struct {
	dao.SQL
	codec codec.Codec
}

func (l *listDB) Append(o interface{}) error {
	data, err := l.codec.Encode(o)
	if err != nil {
		return err
	}

	encoded := string(data)
	return l.Exec("insert", encoded).Error
}

func (l *listDB) GetAt(index int64, o interface{}) error {
	o, err := l.QueryOne("select", "encoded", index)
	if err != nil {
		return err
	}

	data := []byte(o.(string))
	return l.codec.Decode(data, o)
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

func (l *listDB) GetNextFromSeq(index int64, o interface{}) (int64, error) {
	i, err := l.QueryOne("select_from", "encoded", index)
	if err != nil {
		return 0, err
	}

	row := i.(*Row)
	err = l.codec.Decode([]byte(row.encoded), o)
	return row.index, err
}

func (l *listDB) GetAllFromSeq(index int64) (Cursor, error) {
	c, err := l.Query("select_from", "encoded", index)
	if err != nil {
		return nil, err
	}
	return newCursor(c, l.codec), nil
}

func (l *listDB) Delete(index int64) error {
	return l.Exec("delete_by_seq", index).Error
}

func (l *listDB) Clear() error {
	return l.Exec("clear").Error
}

func (l *listDB) Close() error {
	return l.DB.Close()
}

func (l *listDB) scanEventFromEncoded(row dao.Row) (interface{}, error) {
	var encoded string

	err := row.Scan(&encoded)
	if err != nil {
		return nil, err
	}

	var me filespb.SyncEvent
	err = codec.GobDecode([]byte(encoded), &me)

	return &me, err
}

func (l *listDB) scanInt(row dao.Row) (interface{}, error) {
	var v int64
	return v, row.Scan(&v)
}

func New(dbConf jcon.Map, prefix string, codec codec.Codec) (List, error) {
	d := new(listDB)
	d.SetTablePrefix(prefix).
		AddTableDefinition("map", "create table if not exists $prefix$_list (ind int not null primary key $auto_increment$, encoded longblob not null);").
		AddStatement("insert", "insert into $prefix$_list (encoded) values (?);").
		AddStatement("select", "select * from $prefix$_list where ind=?;").
		AddStatement("select_min_index", "select min(ind) from $prefix$_list;").
		AddStatement("select_max_index", "select max(ind) from $prefix$_list;").
		AddStatement("select_count", "select count(ind) from $prefix$_list;").
		AddStatement("select_from", "select * from $prefix$_list where ind>? order by ind;").
		AddStatement("delete_by_seq", "delete from $prefix$_list where ind=?;").
		AddStatement("clear", "delete from $prefix$_list;").
		RegisterScanner("encoded", dao.NewScannerFunc(scanRow)).
		RegisterScanner("index", dao.NewScannerFunc(scanInt))
	err := d.Init(dbConf)
	d.codec = codec
	return d, err
}

func NewSQL(dialect string, db *sql.DB, prefix string, codec codec.Codec) (List, error) {
	d := new(listDB)
	d.SetTablePrefix(prefix).
		AddTableDefinition("map", "create table if not exists $prefix$_list (ind int not null primary key $auto_increment$, encoded longblob not null);").
		AddStatement("insert", "insert into $prefix$_list (encoded) values (?);").
		AddStatement("select", "select * from $prefix$_list where ind=?;").
		AddStatement("select_min_index", "select min(ind) from $prefix$_list;").
		AddStatement("select_max_index", "select max(ind) from $prefix$_list;").
		AddStatement("select_count", "select count(ind) from $prefix$_list;").
		AddStatement("select_from", "select * from $prefix$_list where ind>? order by ind;").
		AddStatement("delete_by_seq", "delete from $prefix$_list where ind=?;").
		AddStatement("clear", "delete from $prefix$_list;").
		RegisterScanner("encoded", dao.NewScannerFunc(scanRow)).
		RegisterScanner("index", dao.NewScannerFunc(scanInt))
	err := d.InitWithSqlDB(dialect, db)
	d.codec = codec
	return d, err
}
