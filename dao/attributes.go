package dao

import (
	"context"
	"database/sql"
	proto "github.com/zoenion/common/proto/attributes"
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
	return nil, nil
}

func NewMySQLAttributesDAO(ctx context.Context, db *sql.DB, prefix string) (AttributesDAO, error) {
	return nil, nil
}

func (dao *sqlAttributesDAO) GetAttributes(key string, names []string) (map[string]*proto.Attribute, error) {
	return nil, nil
}
func (dao *sqlAttributesDAO) GetAll(key string) ([]*proto.Attribute, error) {
	return nil, nil
}

func (dao *sqlAttributesDAO) Get(key string, name string) (*proto.Attribute, error) {
	return nil, nil
}

func (dao *sqlAttributesDAO) SetAttributes(key string, attrs ...*proto.Attribute) error {
	return nil
}

func (dao *sqlAttributesDAO) Set(key string, attr *proto.Attribute) error {

	return nil
}
func (dao *sqlAttributesDAO) Delete(key string, name string) error {
	return nil
}
func (dao *sqlAttributesDAO) DeleteAttributes(key string, names []string) error {
	return nil
}
