package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	"gopkg.in/mgo.v2"
	"strings"
)

func GetDB(c conf.Map) (string, interface{}, error) {
	t, ok := c.GetString("type")
	if !ok {
		return "", nil, errors.NotFound
	}
	switch t {
	case "mongo":
		port, ok := c.GetInt32("port")
		if !ok {
			port = 27017
		}
		host := fmt.Sprintf("%s:%d", c["host"], port)
		info := mgo.DialInfo{
			Addrs:    []string{host},
			Timeout:  5 * time.Second,
			Database: c["name"].(string),
			Username: c["user"].(string),
			Password: c["password"].(string),
		}
		s, e := mgo.DialWithInfo(&info)
		if e != nil {
			return t, nil, e
		}
		s.SetMode(mgo.Monotonic, true)
		return t, s.DB(c["name"].(string)), nil

	case "redis":
		port, ok := c.GetInt32("port")
		if !ok {
			port = 27017
		}

		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", c["host"], port),
			Password: "",
			DB:       0,
		})
		_, err := client.Ping().Result()
		if err != nil {
			return "", nil, err
		}
		return t, client, nil

	case "sql":
		var host string
		var port1 string
		driver := c["driver"].(string)
		port, ok := c.GetInt32("port")
		if !ok { //pas de port
			host = c["host"].(string)
		} else {
			host = fmt.Sprintf(port1, port)
			host = strings.Split(c["host"].(string), ":")[0] + port1
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local",
			c["user"],
			c["password"],
			host,
			c["name"],
			c["charset"],
		)
		if c["wrapped"].(bool) {
			wrapper, ok := c.GetString("wrapper")
			if !ok {
				return "", nil, errors.NotFound
			}
			switch wrapper {
			case "gorm":
				g, err := gorm.Open(c["driver"].(string), dsn)
				return "gorm", g, err
			default:
				return "", nil, errors.NotSupported
			}
		} else {
			s, err := sql.Open(c["driver"].(string), dsn)
			return driver, s, err
		}

	case "sqlite":
		driver := c["driver"].(string)
		path := c["path"].(string)
		s, err := sql.Open(driver, path)
		return driver, s, err

	case "bolt":
		path := c["path"].(string)
		b, err := bolt.Open(path, 0600, nil)
		return t, b, err
	default:
		return "", nil, errors.HttpNotImplemented
	}
}

func GetMysql(c conf.Map) (*sql.DB, error) {
	_, dbi, err := GetDB(c)
	if err != nil {
		return nil, err
	}

	db, ok := dbi.(*sql.DB)
	if !ok {
		return nil, errors.NotFound
	}
	return db, nil
}

// SQLite creates an instance of SQLite database
func SQLite(driver string, path string) (*sql.DB, error) {
	return sql.Open(driver, path)
}

func Bolt(path string) (*bolt.DB, error) {
	return bolt.Open(path, 0755, nil)
}

func SQLiteConfig(filename string) conf.Map {
	return conf.Map{
		"type":   "sqlite",
		"driver": "sqlite3",
		"path":   filename,
	}
}
