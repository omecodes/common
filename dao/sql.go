package dao

import (
	"database/sql"
	"fmt"
	"github.com/omecodes/common/database"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/jcon"
	"github.com/omecodes/common/log"
	"strings"
	"sync"
)

const (
	MySQLIndexScanner  = "mysql_index_scanner"
	SQLiteIndexScanner = "sqlite_index_scanner"

	VarPrefix        = "$prefix$"
	VarAutoIncrement = "$auto_increment$"
	VarLocate        = "$locate$"

	// ScannerIndex = "scanner_index"
)

type SQlv2Result struct {
	Error        error
	LastInserted int64
	AffectedRows int64
}

type SQLIndex struct {
	Name   string
	Table  string
	Fields []string
}

type SQL struct {
	DB                         *sql.DB
	mux                        *sync.RWMutex
	dialect                    string
	isSQLite                   bool
	compiledStatements         map[string]*sql.Stmt
	vars                       map[string]string
	tableDefs                  map[string]string
	registeredStatements       map[string]string
	registeredSQLiteStatements map[string]string
	registeredMySQLStatements  map[string]string
	migrationScripts           []string
	scanners                   map[string]RowScannerV2
	initDone                   bool
}

func (dao *SQL) Init(cfg jcon.Map) error {
	d, dbi, err := database.Connect(cfg)
	if err != nil {
		return err
	}

	switch d {
	case "mysql", "sqlite3":
		dao.dialect = d
		db := dbi.(*sql.DB)
		if d == "mysql" {
			return dao.InitWithMySQLDB(db)
		} else {
			return dao.InitSQLite(db)
		}
	default:
		return errors.New("database dialect is not supported")
	}
}

func (dao *SQL) InitWithSqlDB(dialect string, db *sql.DB) error {
	if dialect == "mysql" {
		return dao.InitWithMySQLDB(db)
	}

	if dialect == "sqlite3" {
		return dao.InitSQLite(db)
	}

	return errors.NotSupported
}

func (dao *SQL) InitWithMySQLDB(db *sql.DB) error {
	dao.DB = db
	dao.SetVariable(VarLocate, "locate")
	dao.SetVariable(VarAutoIncrement, "AUTO_INCREMENT")
	return dao.init()
}

func (dao *SQL) InitSQLite(db *sql.DB) error {
	dao.DB = db
	dao.isSQLite = true
	dao.SetVariable(VarLocate, "instr")
	dao.SetVariable(VarAutoIncrement, "AUTOINCREMENT")
	if _, err := dao.DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		log.Error("failed to enable foreign key feature", err)
	}
	dao.mux = new(sync.RWMutex)
	return dao.init()
}

func (dao *SQL) init() error {
	dao.RegisterScanner(MySQLIndexScanner, NewScannerFunc(dao.mysqlIndexScan))
	dao.RegisterScanner(SQLiteIndexScanner, NewScannerFunc(dao.sqliteIndexScan))

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

	var specificStatements map[string]string
	if dao.isSQLite && dao.registeredSQLiteStatements != nil {
		specificStatements = dao.registeredSQLiteStatements
	} else {
		specificStatements = dao.registeredMySQLStatements
	}

	if specificStatements != nil {
		if dao.registeredStatements == nil {
			dao.registeredSQLiteStatements = map[string]string{}
		}
		for name, stmt := range specificStatements {
			dao.registeredStatements[name] = stmt
		}
	}

	if dao.registeredStatements != nil && len(dao.registeredStatements) > 0 {
		dao.compiledStatements = map[string]*sql.Stmt{}
		for name, stmt := range dao.registeredStatements {
			for name, value := range dao.vars {
				stmt = strings.Replace(stmt, name, value, -1)
			}

			compiledStmt, err := dao.DB.Prepare(stmt)
			if err != nil {
				log.Error("failed to compile statement", err, log.Field("name", name), log.Field("sql", stmt))
				return err
			}
			dao.compiledStatements[name] = compiledStmt
		}
	}

	dao.initDone = true
	return nil
}

func (dao *SQL) Migrate() error {
	if !dao.initDone {
		return errors.New("initWithBox method must be called once before calling TableHasIndex method")
	}
	for _, ms := range dao.migrationScripts {
		for name, value := range dao.vars {
			ms = strings.Replace(ms, name, value, -1)
		}

		_, err := dao.DB.Exec(ms)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dao *SQL) IsSQLite() bool {
	return dao.isSQLite
}

func (dao *SQL) SetVariable(name string, value string) *SQL {
	if dao.vars == nil {
		dao.vars = map[string]string{}
	}
	dao.vars[name] = value
	return dao
}

func (dao *SQL) SetTablePrefix(prefix string) *SQL {
	if dao.vars == nil {
		dao.vars = map[string]string{}
	}
	dao.vars[VarPrefix] = prefix
	return dao
}

func (dao *SQL) AddMigrationScript(s string) *SQL {
	dao.migrationScripts = append(dao.migrationScripts, s)
	return dao
}

func (dao *SQL) AddUniqueIndex(index SQLIndex, forceUpdate bool) error {
	if !dao.initDone {
		return errors.New("initWithBox method must be called once before calling TableHasIndex method")
	}

	for varName, value := range dao.vars {
		index.Table = strings.Replace(index.Table, varName, value, -1)
	}
	hasIndex, err := dao.TableHasIndex(index)
	if err != nil {
		return err
	}

	var result *SQlv2Result
	if hasIndex && forceUpdate {
		var dropIndexSQL string
		if dao.dialect == "mysql" {
			dropIndexSQL = fmt.Sprintf("drop index %s on %s", index.Name, index.Table)
		} else {
			dropIndexSQL = fmt.Sprintf("drop index if exists %s", index.Name)
		}

		result = dao.RawExec(dropIndexSQL)
		if result.Error != nil {
			return result.Error
		}
	}

	if !hasIndex || forceUpdate {
		var createIndexSQL string
		if dao.dialect == "mysql" {
			createIndexSQL = fmt.Sprintf("create unique index %s on %s(%s)", index.Name, index.Table, strings.Join(index.Fields, ","))
		} else {
			createIndexSQL = fmt.Sprintf("create unique index if not exists %s on %s(%s)", index.Name, index.Table, strings.Join(index.Fields, ","))
		}

		result = dao.RawExec(createIndexSQL)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (dao *SQL) AddTableDefinition(name string, schema string) *SQL {
	if dao.tableDefs == nil {
		dao.tableDefs = map[string]string{}
	}
	dao.tableDefs[name] = schema
	return dao
}

func (dao *SQL) AddStatement(name string, statementStr string) *SQL {
	if dao.registeredStatements == nil {
		dao.registeredStatements = map[string]string{}
	}
	dao.registeredStatements[name] = statementStr
	return dao
}

func (dao *SQL) AddSQLiteStatement(name string, statementStr string) *SQL {
	if dao.registeredSQLiteStatements == nil {
		dao.registeredSQLiteStatements = map[string]string{}
	}
	dao.registeredSQLiteStatements[name] = statementStr
	return dao
}

func (dao *SQL) AddMySQLStatement(name string, statementStr string) *SQL {
	if dao.registeredMySQLStatements == nil {
		dao.registeredMySQLStatements = map[string]string{}
	}
	dao.registeredMySQLStatements[name] = statementStr
	return dao
}

func (dao *SQL) RegisterScanner(name string, scanner RowScannerV2) *SQL {
	if dao.scanners == nil {
		dao.scanners = map[string]RowScannerV2{}
	}
	dao.scanners[name] = scanner
	return dao
}

func (dao *SQL) TableHasIndex(index SQLIndex) (bool, error) {
	if !dao.initDone {
		return false, errors.New("initWithBox method must be called once before calling TableHasIndex method")
	}

	var (
		scannerName string
		rawQuery    string
	)
	if dao.dialect == "mysql" {
		rawQuery = fmt.Sprintf("SHOW INDEX FROM %s", index.Table)
		scannerName = MySQLIndexScanner
	} else {
		rawQuery = fmt.Sprintf("PRAGMA INDEX_LIST('%s')", index.Table)
		scannerName = SQLiteIndexScanner
	}

	cursor, err := dao.RawQuery(rawQuery, scannerName)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = cursor.Close()
	}()

	for cursor.HasNext() {
		ind, err := cursor.Next()
		if err != nil {
			return false, err
		}

		rowIndex := ind.(SQLIndex)
		if rowIndex.Name == index.Name {
			return true, nil
		}
	}
	return false, nil
}

func (dao *SQL) RawQuery(query string, scannerName string, params ...interface{}) (DBCursor, error) {
	for name, value := range dao.vars {
		query = strings.Replace(query, name, value, -1)
	}
	rows, err := dao.DB.Query(query, params...)
	if err != nil {
		return nil, err
	}
	scanner, err := dao.findScanner(scannerName)
	if err != nil {
		return nil, err
	}
	return NewSQLDBCursor(rows, scanner), nil
}

func (dao *SQL) RawQueryOne(query string, scannerName string, params ...interface{}) (interface{}, error) {
	for name, value := range dao.vars {
		query = strings.Replace(query, name, value, -1)
	}

	rows, err := dao.DB.Query(query, params...)
	if err != nil {
		return nil, err
	}
	scanner, err := dao.findScanner(scannerName)
	if err != nil {
		return nil, err
	}

	cursor := NewSQLDBCursor(rows, scanner)
	defer func() {
		_ = cursor.Close()
	}()

	if !cursor.HasNext() {
		return nil, errors.NotFound
	}
	return cursor.Next()
}

func (dao *SQL) RawExec(rawQuery string) *SQlv2Result {
	dao.wLock()
	defer dao.wUnlock()
	var r sql.Result
	result := &SQlv2Result{}
	for name, value := range dao.vars {
		rawQuery = strings.Replace(rawQuery, name, value, -1)
	}
	r, result.Error = dao.DB.Exec(rawQuery)
	if result.Error == nil && dao.dialect != "sqlite3" {
		result.LastInserted, _ = r.LastInsertId()
		result.AffectedRows, _ = r.RowsAffected()
	}
	return result
}

func (dao *SQL) Query(stmt string, scannerName string, params ...interface{}) (DBCursor, error) {
	dao.rLock()
	defer dao.rUnLock()

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

func (dao *SQL) QueryOne(stmt string, scannerName string, params ...interface{}) (interface{}, error) {
	dao.rLock()
	defer dao.rUnLock()

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
	defer func() {
		_ = cursor.Close()
	}()

	if !cursor.HasNext() {
		return nil, errors.NotFound
	}
	return cursor.Next()
}

func (dao *SQL) Exec(stmt string, params ...interface{}) *SQlv2Result {
	dao.wLock()
	defer dao.wUnlock()

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

func (dao *SQL) sqliteIndexScan(row Row) (interface{}, error) {
	dao.rLock()
	defer dao.rUnLock()

	var index SQLIndex
	m, err := dao.rowToMap(row.(*sql.Rows))
	if err != nil {
		return nil, err
	}

	var ok bool
	index.Name, ok = m["name"].(string)
	if !ok {
		return nil, errors.New("SQLite index scanned row 'Name' field type mismatches")
	}
	return index, nil
}

func (dao *SQL) mysqlIndexScan(row Row) (interface{}, error) {
	var index SQLIndex
	m, err := dao.rowToMap(row.(*sql.Rows))
	if err != nil {
		return nil, err
	}

	index.Name = fmt.Sprintf("%s", m["Key_name"])
	if index.Name == "" {
		return nil, errors.New("mysql index scanned row 'Name' field type mismatches")
	}
	index.Table = fmt.Sprintf("%s", m["Table"])
	if index.Table == "" {
		return nil, errors.New("mysql index scanned row 'Table' field type mismatches")
	}
	return index, nil
}

func (dao *SQL) rowToMap(rows *sql.Rows) (map[string]interface{}, error) {
	cols, _ := rows.Columns()
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	// Scan the result into the column pointers...
	if err := rows.Scan(columnPointers...); err != nil {
		return nil, err
	}

	// Create our map, and retrieve the value for each column from the pointers slice,
	// storing it in the map with the name of the column as the key.
	m := make(map[string]interface{})
	for i, colName := range cols {
		val := columnPointers[i].(*interface{})
		m[colName] = *val
	}
	return m, nil

}

func (dao *SQL) findCompileStatement(name string) (*sql.Stmt, error) {
	if dao.compiledStatements == nil {
		return nil, errors.NotFound
	}

	if compiledStmt, found := dao.compiledStatements[name]; found {
		return compiledStmt, nil
	}
	return nil, errors.NotFound
}

func (dao *SQL) findScanner(name string) (RowScannerV2, error) {
	scanner, found := dao.scanners[name]
	if !found {
		return nil, errors.New("no scanner found")
	}
	return scanner, nil
}

func (dao *SQL) getStatement(name string) *sql.Stmt {
	if dao.compiledStatements == nil {
		return nil
	}
	s, found := dao.compiledStatements[name]
	if !found {
		return nil
	}
	return s
}

func (dao *SQL) rLock() {
	if dao.mux != nil {
		dao.mux.RLock()
	}
}

func (dao *SQL) wLock() {
	if dao.mux != nil {
		dao.mux.Lock()
	}
}

func (dao *SQL) rUnLock() {
	if dao.mux != nil {
		dao.mux.RUnlock()
	}
}

func (dao *SQL) wUnlock() {
	if dao.mux != nil {
		dao.mux.Unlock()
	}
}

type DBCursor interface {
	Next() (interface{}, error)
	HasNext() bool
	Close() error
}

type scannerFunc struct {
	f func(rows Row) (interface{}, error)
}

func (sf *scannerFunc) ScanRow(row Row) (interface{}, error) {
	return sf.f(row)
}

func NewScannerFunc(f func(row Row) (interface{}, error)) RowScannerV2 {
	return &scannerFunc{
		f: f,
	}
}

// SQLCursor
type SQLCursor struct {
	err  error
	scan RowScannerV2
	rows *sql.Rows
}

func NewSQLDBCursor(rows *sql.Rows, scanner RowScannerV2) DBCursor {
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
