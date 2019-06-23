package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/log"
)

type SQLRowScanner func(rows *sql.Rows)

type SQL struct {
	DB   *sql.DB
	stmt map[string]*sql.Stmt
	vars map[string]string
}

func (dao *SQL) Set(name string, value string) {
	if dao.vars == nil {
		dao.vars = map[string]string{}
	}
	dao.vars[name] = value
}

func (dao *SQL) TableHasIndex(table string, indexName string) bool {
	var count = 0
	err := dao.Query("has_index", func(rows *sql.Rows) {
		if rows.Next() {
			err := rows.Scan(&count)
			if err != nil {
				count = 0
			}
		}
		if err := rows.Close(); err != nil {
			log.E("dao.index.check", err, "cursor close() caused error")
		}
	}, table, indexName)
	if err != nil {
		log.E("dao.index.check", err)
	}
	return count > 0
}

func (dao *SQL) ExecuteSQL(query string) error {
	_, err := dao.DB.Exec(query)
	return err
}

func (dao *SQL) QuerySQL(query string, scanner SQLRowScanner, params ...interface{}) error {
	rows, err := dao.DB.Query(query)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	if scanner != nil {
		scanner(rows)
	}
	return rows.Err()
}

func (dao *SQL) Init(tablePrefix string, schema []string, statements map[string]string) error {
	dao.stmt = map[string]*sql.Stmt{}

	query := ` SELECT COUNT(*)
	FROM information_schema.statistics
	WHERE TABLE_SCHEMA = DATABASE()
  	AND TABLE_NAME = ? AND INDEX_NAME LIKE ?;`

	stmt, err := dao.DB.Prepare(query)
	if err == nil {
		dao.stmt["has_index"] = stmt
	}

	for i := range schema {
		str := strings.Replace(schema[i], "$prefix$", tablePrefix, -1)
		for name, value := range dao.vars {
			str = strings.Replace(str, name, value, -1)
		}
		_, err := dao.DB.Exec(str)
		if err != nil {
			return err
		}
	}

	for key, strStmt := range statements {
		stmt, err := dao.DB.Prepare(strings.Replace(strStmt, "$prefix$", tablePrefix, -1))
		if err != nil {
			log.E("attributes-DAO", err, key+":"+strStmt)
			return err
		}
		dao.stmt[key] = stmt
	}

	return nil
}

func (dao *SQL) getStatement(name string) *sql.Stmt {
	if dao.stmt == nil {
		return nil
	}
	s, found := dao.stmt[name]
	if !found {
		return nil
	}
	return s
}

func (dao *SQL) Query(stmt string, scanner SQLRowScanner, params ...interface{}) error {
	st := dao.getStatement(stmt)
	if st == nil {
		if dao.stmt == nil {
			return errors.Detailed(errors.BadInput, "database misconfigured")
		}
		return fmt.Errorf("statement `%s` does not exist", stmt)
	}
	rows, err := st.Query(params...)
	if err != nil {
		return err
	}
	scanner(rows)
	err = rows.Err()
	_ = rows.Close()
	return err
}

func (dao *SQL) Execute(stmt string, params ...interface{}) error {
	st := dao.getStatement(stmt)
	if st == nil {
		return errors.Detailed(errors.HttpNotFound, fmt.Sprintf("statement `%s` does not exist", stmt))
	}
	_, err := st.Exec(params...)
	return err
}
