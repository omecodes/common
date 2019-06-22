package dao

import (
	"database/sql"
	"fmt"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/database"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/log"
	"strings"
)

const (
	StatementKeyHasIndex = "has_index"

	VarPrefix        = "$prefix$"
	VarAutoIncrement = "$auto_increment$"

	ScannerIndex = "scanner_index"
)

type SQLv2RowScanner interface {
	ScanRow(rows *sql.Rows) (interface{}, error)
}

type SQlv2Result struct {
	Error        error
	LastInserted int64
	AffectedRows int64
}

type SQLv2 struct {
	DB                   *sql.DB
	dialect              string
	compiledStatements   map[string]*sql.Stmt
	vars                 map[string]string
	tableDefs            map[string]string
	registeredStatements map[string]string
	scanners             map[string]SQLv2RowScanner
}

func (dao *SQLv2) Init(cfg conf.Map) error {
	d, dbi, err := database.GetDB(cfg)
	if err != nil {
		return err
	}

	switch d {
	case "mysql", "sqlite3":
		dao.dialect = d
		dao.DB = dbi.(*sql.DB)
		if d == "mysql" {
			dao.SetVariable(VarAutoIncrement, "AUTO_INCREMENT")
		} else {
			dao.SetVariable(VarAutoIncrement, "AUTOINCREMENT")
		}
	default:
		return errors.New("database dialect is not supported")
	}

	dao.AddStatement(StatementKeyHasIndex, ` SELECT COUNT(*)
	FROM information_schema.statistics
	WHERE TABLE_SCHEMA = DATABASE()
  	AND TABLE_NAME = ? AND INDEX_NAME LIKE ?;`).
		RegisterScanner(StatementKeyHasIndex, NewScannerFunc(dao.scanIndex))

	if dao.tableDefs != nil && len(dao.tableDefs) > 0 {
		for _, schema := range dao.tableDefs {
			for name, value := range dao.vars {
				schema = strings.Replace(schema, name, value, -1)
			}
			_, err := dao.DB.Exec(schema)
			if err != nil {
				return err
			}
		}
	}

	if dao.registeredStatements != nil && len(dao.registeredStatements) > 0 {
		for name, stmt := range dao.registeredStatements {
			for name, value := range dao.vars {
				stmt = strings.Replace(stmt, name, value, -1)
			}

			stmt, err := dao.DB.Prepare(stmt)
			if err != nil {
				log.Ef("SQLv2", err, "failed to compile '%s' statement: %s", name, stmt)
				return err
			}
			dao.compiledStatements[name] = stmt
		}
	}
	return nil
}

func (dao *SQLv2) SetVariable(name string, value string) *SQLv2 {
	if dao.vars == nil {
		dao.vars = map[string]string{}
	}
	dao.vars[name] = value
	return dao
}

func (dao *SQLv2) SetTablePrefix(prefix string) *SQLv2 {
	if dao.vars == nil {
		dao.vars = map[string]string{}
	}
	dao.vars[VarPrefix] = prefix
	return dao
}

func (dao *SQLv2) SetTableDefinition(name string, schema string) *SQLv2 {
	if dao.tableDefs == nil {
		dao.tableDefs = map[string]string{}
	}
	dao.tableDefs[name] = schema
	return dao
}

func (dao *SQLv2) AddStatement(name string, statementStr string) *SQLv2 {
	if dao.registeredStatements == nil {
		dao.registeredStatements = map[string]string{}
	}
	dao.registeredStatements[name] = statementStr
	return dao
}

func (dao *SQLv2) RegisterScanner(name string, scanner SQLv2RowScanner) *SQLv2 {
	if dao.scanners == nil {
		dao.scanners = map[string]SQLv2RowScanner{}
	}
	dao.scanners[name] = scanner
	return dao
}

func (dao *SQLv2) TableHasIndex(table string, indexName string) (bool, error) {
	cursor, err := dao.Query(StatementKeyHasIndex, ScannerIndex, table, indexName)
	if err != nil {
		return false, err
	}

	defer func() {
		_ = cursor.Close()
	}()

	if cursor.HasNext() {
		c, err := cursor.Next()
		if err != nil {
			return false, err
		}
		return c.(int) > 0, nil
	}

	return false, nil
}

func (dao *SQLv2) RawQuery(query string, scannerName string, params ...interface{}) (DBCursor, error) {
	rows, err := dao.DB.Query(query)
	if err != nil {
		return nil, err
	}
	scanner, err := dao.findScanner(scannerName)
	if err != nil {
		return nil, err
	}
	return NewSQLDBCursor(rows, scanner), nil
}

func (dao *SQLv2) RawExec(rawQuery string) *SQlv2Result {
	var r sql.Result
	result := &SQlv2Result{}

	r, result.Error = dao.DB.Exec(rawQuery)
	if result.Error == nil {
		result.LastInserted, _ = r.LastInsertId()
		result.AffectedRows, _ = r.RowsAffected()
	}
	return result
}

func (dao *SQLv2) Query(stmt string, scannerName string, params ...interface{}) (DBCursor, error) {
	st := dao.getStatement(stmt)
	if st == nil {
		return nil, fmt.Errorf("statement `%s` does not exist", stmt)
	}

	rows, err := st.Query(params...)
	if err != nil {
		return nil, err
	}

	scanner, err := dao.findScanner(scannerName)
	if err != nil {
		return nil, err
	}

	cursor := NewSQLDBCursor(rows, scanner)
	return cursor, nil
}

func (dao *SQLv2) Exec(stmt string, params ...interface{}) *SQlv2Result {
	result := &SQlv2Result{}
	var (
		st *sql.Stmt
		r  sql.Result
	)

	st, result.Error = dao.findCompileStatement(stmt)
	if result.Error != nil {
		return result
	}

	r, result.Error = st.Exec(params...)
	if result.Error == nil {
		result.LastInserted, _ = r.LastInsertId()
		result.AffectedRows, _ = r.RowsAffected()
	}
	return result
}

func (dao *SQLv2) scanIndex(rows *sql.Rows) (interface{}, error) {
	var count int
	err := rows.Scan(&count)
	if err != nil {
		count = 0
	}
	return count, nil
}

func (dao *SQLv2) findCompileStatement(name string) (*sql.Stmt, error) {
	if dao.compiledStatements == nil {
		return nil, errors.Detailed(errors.NotFound, fmt.Sprintf("no compiled statement with name '%s' found", name))
	}

	if compiledStmt, found := dao.compiledStatements[name]; found {
		return compiledStmt, nil
	}
	return nil, errors.NotFound
}

func (dao *SQLv2) findScanner(name string) (SQLv2RowScanner, error) {
	scanner, found := dao.scanners[name]
	if !found {
		return nil, errors.Detailed(errors.NotFound, fmt.Sprintf("no scanner with name '%s' found", name))
	}
	return scanner, nil
}

func (dao *SQLv2) getStatement(name string) *sql.Stmt {
	if dao.compiledStatements == nil {
		return nil
	}
	s, found := dao.compiledStatements[name]
	if !found {
		return nil
	}
	return s
}

type DBCursor interface {
	Next() (interface{}, error)
	HasNext() bool
	Close() error
}

type scannerFunc struct {
	f func(rows *sql.Rows) (interface{}, error)
}

func (sf *scannerFunc) ScanRow(rows *sql.Rows) (interface{}, error) {
	return sf.f(rows)
}

func NewScannerFunc(f func(rows *sql.Rows) (interface{}, error)) SQLv2RowScanner {
	return &scannerFunc{
		f: f,
	}
}

// SQLCursor
type SQLCursor struct {
	err  error
	scan SQLv2RowScanner
	rows *sql.Rows
}

func NewSQLDBCursor(rows *sql.Rows, scanner SQLv2RowScanner) DBCursor {
	return &SQLCursor{
		scan: scanner,
		rows: rows,
	}
}

func (c *SQLCursor) Close() error {
	return c.rows.Close()
}

func (c *SQLCursor) HasNext() bool {
	return c.rows.Next()
}

func (c *SQLCursor) Next() (interface{}, error) {
	return c.scan.ScanRow(c.rows)
}
