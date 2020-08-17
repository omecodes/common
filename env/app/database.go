package app

import (
	"database/sql"
	"fmt"
	"github.com/omecodes/common/utils/jcon"
	"time"

	"github.com/boltdb/bolt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/omecodes/common/errors"
	"gopkg.in/mgo.v2"
	"strings"
)

func Connect(c jcon.Map) (string, interface{}, error) {
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

	/* case "redis":
	port, ok := c.GetInt32("port")
	if !ok {
		port = 27017
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c["host"], port),
		Password: "",
		DB:       0,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return "", nil, err
	}
	return t, client, nil */

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
			if err != nil {
				return driver, s, err
			}
			_, _ = s.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c["name"]))
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
		return "", nil, errors.NotImplemented
	}
}

func Create(c jcon.Map) error {
	driver := c["driver"]
	if driver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/?charset=%s&parseTime=True&loc=Local",
			c["user"],
			c["password"],
			c["host"],
			c["charset"],
		)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c["name"]))
		return err
	}
	return nil
}

func SQLiteConfig(filename string) jcon.Map {
	return jcon.Map{
		"type":   "sqlite",
		"driver": "sqlite3",
		"path":   filename,
	}
}
