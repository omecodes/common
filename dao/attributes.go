package dao

import (
	"context"
	"database/sql"

	"github.com/zoenion/micro"
	"github.com/zoenion/onion/app/common/proto"

	"fmt"
	"github.com/zoenion/micro/log"
)

var schema = []string{
	`create table if not exists $prefix$_attrs (
		app_name varchar(255) not null,
		attr_name varchar(255) not null,
		attr_value blob not null,
		attr_details blob
	);`,
}

var statements = map[string]string{
	"insert":     "insert into $prefix$_attrs values (?, ?, ?, ?);",
	"update":     "update $prefix$_attrs set attr_value=?,attr_details=? where app_name=? and attr_name=?;",
	"select":     "select attr_name, attr_value, attr_details from $prefix$_attrs where app_name=? and attr_name=?;",
	"select_all": "select attr_name, attr_value, attr_details from $prefix$_attrs where app_name=?;",
	"delete":     "delete from  $prefix$_attrs where app_name=? and attr_name=?;",
	"delete_all": "delete from  $prefix$_attrs where app_name=?;",
}

// AttributesDAO is an interface for attributes persistence
type AttributesDAO interface {
	GetAttributes(key string, names []string) (map[string]*proto.Attribute, error)
	Get(key string, name string) (*proto.Attribute, error)
	GetAll(key string) ([]*proto.Attribute, error)
	SetAttributes(key string, attrs ...*proto.Attribute) error
	Set(key string, attr *proto.Attribute) error
	Delete(key string, name string) error
	DeleteAttributes(key string, names []string) error
}

type sqlAttributesDAO struct {
	SQL
	ctx context.Context
}

// NewSQLiteAttributesDAO creates a sql attributes persistence
func NewSQLiteAttributesDAO(ctx context.Context, db *sql.DB, prefix string) (AttributesDAO, error) {
	dao := new(sqlAttributesDAO)
	dao.DB = db
	dao.ctx = ctx
	constraint := "create unique index if not exists `unique_attr_kn` on $prefix$_attrs (app_name, attr_name);"
	e := dao.Init(prefix, append(schema, constraint), statements)
	return dao, e
}

func NewMySQLAttributesDAO(ctx context.Context, db *sql.DB, prefix string) (AttributesDAO, error) {
	dao := new(sqlAttributesDAO)
	dao.DB = db
	dao.ctx = ctx
	e := dao.Init(prefix, schema, statements)
	if e != nil {
		return nil, e
	}

	if !dao.TableHasIndex(prefix+"_attrs", "unique_attr_kn") {
		err := dao.ExecuteSQL(fmt.Sprintf("create unique index `unique_attr_kn` on `%s_attrs` (app_name, attr_name)", prefix))
		if err != nil {
			return nil, err
		}
	}
	return dao, e
}

func (dao *sqlAttributesDAO) GetAttributes(key string, names []string) (map[string]*proto.Attribute, error) {
	result := map[string]*proto.Attribute{}
	for _, name := range names {
		attr, err := dao.Get(key, name)
		if err != nil {
			return nil, err
		}
		result[name] = attr
	}
	return result, nil
}
func (dao *sqlAttributesDAO) GetAll(key string) ([]*proto.Attribute, error) {
	attrs := []*proto.Attribute{}
	var err error

	scanner := func(rows *sql.Rows) {
		for rows.Next() {
			attr := new(proto.Attribute)
			err = rows.Scan(&(attr.Name), &(attr.Value), &(attr.Description))
			if err == nil {
				attrs = append(attrs, attr)
			} else {
				log.E(micro.GetServiceName(dao.ctx), err, "sql attr row scann failed")
			}
		}
	}
	if err := dao.Query("select_all", scanner, key); err != nil {
		return nil, err
	}
	return attrs, err
}
func (dao *sqlAttributesDAO) Get(key string, name string) (*proto.Attribute, error) {
	var attr *proto.Attribute
	var err error
	scanner := func(rows *sql.Rows) {
		for rows.Next() {
			attr = new(proto.Attribute)
			err = rows.Scan(&(attr.Name), &(attr.Value), &(attr.Description))
		}
	}
	if err := dao.Query("select", scanner, key, name); err != nil {
		return nil, err
	}
	return attr, err
}
func (dao *sqlAttributesDAO) SetAttributes(key string, attrs ...*proto.Attribute) error {
	for _, attr := range attrs {
		err := dao.Set(key, attr)
		if err != nil {
			return err
		}
	}
	return nil
}
func (dao *sqlAttributesDAO) Set(key string, attr *proto.Attribute) error {
	err := dao.Execute("insert", key, attr.Name, attr.Value, attr.Description)
	if err != nil {
		return dao.Execute("update", attr.Value, attr.Description, key, attr.Name)
	}
	return nil
}
func (dao *sqlAttributesDAO) Delete(key string, name string) error {
	return dao.Execute("delete", key, name)
}
func (dao *sqlAttributesDAO) DeleteAttributes(key string, names []string) error {
	if names == nil {
		err := dao.Execute("delete_all", key)
		if err != nil {
			return err
		}
		return nil
	}

	for _, name := range names {
		err := dao.Delete(key, name)
		if err != nil {
			return err
		}
	}
	return nil
}
