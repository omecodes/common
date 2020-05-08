package app

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/database"
	"github.com/zoenion/common/errors"
	"github.com/zoenion/common/network"
	"github.com/zoenion/common/prompt"
)

type configItem struct {
	description string
	configType  ConfigType
	entries     []string
	values      []string
}

type ConfigType int

const (
	ConfigMailer ConfigType = iota + 1
	ConfigAccess
	ConfigAdminsCredentials
	ConfigSecrets
	ConfigCredentialsTable
	ConfigDirs
	ConfigNetwork
	ConfigMySQLDatabase
	ConfigSQLiteDatabase
	ConfigRedisDatabase
	ConfigMongoDatabase
)

func (ci ConfigType) String() string {
	switch ci {

	case ConfigAccess:
		return "access"

	case ConfigMailer:
		return "mailer"

	case ConfigAdminsCredentials:
		return "admins"

	case ConfigDirs:
		return "dirs"

	case ConfigSecrets:
		return "secrets"

	case ConfigCredentialsTable:
		return "credentials"

	case ConfigNetwork:
		return "network"

	case ConfigSQLiteDatabase:
		return "databases/file"

	case ConfigMySQLDatabase:
		return "databases/mysql"

	case ConfigRedisDatabase:
		return "databases/redis"

	case ConfigMongoDatabase:
		return "databases/mongo"

	default:
		return ""
	}
}

func (ci configItem) create(description string, defaults conf.Map) (conf.Map, error) {
	switch ci.configType {

	case ConfigAccess:
		return configureAccess(description, defaults)

	case ConfigMailer:
		return configureMailer(description, defaults)

	case ConfigAdminsCredentials:
		return configureAdminsCredentials(description, defaults)

	case ConfigDirs:
		return configureDirs(description, defaults, ci.entries...)

	case ConfigCredentialsTable:
		return configureCredentialsTable(description, defaults)

	case ConfigNetwork:
		return configureNetwork(description, defaults)

	case ConfigMySQLDatabase:
		return configureMySQLDatabase(description, defaults)

	case ConfigSQLiteDatabase:
		return configureSQLiteDatabase(description, defaults)

	case ConfigRedisDatabase:
		return configureRedisDatabase(description, defaults)

	case ConfigMongoDatabase:
		return configureMongoDatabase(description, defaults)

	case ConfigSecrets:
		return configureSecrets(description, defaults)

	default:
		return nil, errors.NotSupported
	}
}

func configureAccess(description string, defaults conf.Map) (conf.Map, error) {
	header("Access", description)

	cfg := conf.Map{}

	name, err := prompt.Text("Key", false)
	if err != nil {
		return nil, err
	}

	secret, err := prompt.Password("Secret")
	if err != nil {
		return nil, err
	}

	cfg["key"] = name
	cfg["secret"] = secret
	return cfg, nil
}

func configureSecrets(description string, defaults conf.Map) (conf.Map, error) {
	header("Secrets", description)

	var err error
	count := 0
	cfg := conf.Map{}

	for {
		if count > 0 {
			selection, err := prompt.Selection("add another secret?", "yes", "no")
			if err != nil {
				return nil, err
			}
			if selection != "yes" {
				break
			}
		}
		count++

		name, err := prompt.Text("Name", false)
		if err != nil {
			return nil, err
		}

		secret, err := prompt.Password("Secret")
		if err != nil {
			return nil, err
		}
		cfg[name] = secret
		fmt.Println()
	}
	return cfg, err
}

func configureCredentialsTable(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Credentials", description)

	var err error
	key, _ := defaults.GetString("subject")

	if key != "" {
		key, err = prompt.TextWithDefault("subject", key, false)
	} else {
		key, err = prompt.Text("subject", false)
	}

	secret, err := prompt.Password("password")
	if err != nil {
		return nil, err
	}
	return conf.Map{"subject": key, "password": secret}, nil
}

func configureMailer(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Mailer", description)

	var err error
	server, _ := defaults.GetString("server")
	port, _ := defaults.GetInt64("port")
	user, _ := defaults.GetString("user")

	server, err = prompt.TextWithDefault("server", server, false)
	if err != nil {
		return nil, err
	}

	if port > 0 {
		port, err = prompt.IntegerWithDefaultValue("port", port)
		if err != nil {
			return nil, err
		}
	} else {
		port, err = prompt.Integer("port")
		if err != nil {
			return nil, err
		}
	}

	user, err = prompt.TextWithDefault("user", user, false)
	if err != nil {
		return nil, err
	}

	var password string
	password, err = prompt.Password("password")
	if err != nil {
		return nil, err
	}
	return conf.Map{
		"server":   server,
		"port":     port,
		"user":     user,
		"password": password,
	}, nil
}

func configureAdminsCredentials(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Admins credentials", description)
	var err error
	count := 0
	cfg := conf.Map{}

	for {
		if count > 0 {
			selection, err := prompt.Selection("add another admin user?", "yes", "no")
			if err != nil {
				return nil, err
			}

			if selection != "yes" {
				break
			}
		}
		count++

		user, err := prompt.Text("username", false)
		if err != nil {
			return nil, err
		}

		password, err := prompt.Password("password")
		if err != nil {
			return nil, err
		}

		data := sha256.Sum256([]byte(password))
		cfg[user] = base64.StdEncoding.EncodeToString(data[:])
		fmt.Println()
	}
	return cfg, err
}

func configureDirs(description string, defaults conf.Map, names ...string) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}

	header("Directories", description)
	var err error
	count := 0
	cfg := conf.Map{}

	for _, name := range names {
		count++

		dirPath, _ := defaults.GetString(name)
		if dirPath == "" {
			dirPath, err = prompt.Text(name, false)
		} else {
			dirPath, err = prompt.TextWithDefault(name, dirPath, false)
		}

		if err != nil {
			return nil, err
		}
		cfg[name] = dirPath
		fmt.Println()
	}

	return cfg, err
}

func configureNetwork(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Network", description)

	var err error
	domain, _ := defaults.GetString("domain")
	internalIP, _ := defaults.GetString("internal_ip")
	externalIP, _ := defaults.GetString("external_ip")

	addrList := network.LocalAddresses()
	internalIP, err = prompt.Selection("internal IP", addrList...)
	if err != nil {
		return nil, err
	}

	domain, err = prompt.TextWithDefault("domain", domain, true)
	if err != nil {
		return nil, err
	}

	selection, err := prompt.Selection("Computer has external IP different from bind ip?", "yes", "no")
	if err != nil {
		return nil, err
	}

	if selection == "yes" {
		externalIP, err = prompt.TextWithDefault("external IP", externalIP, false)
	}

	return conf.Map{
		"domain":      domain,
		"internal_ip": internalIP,
		"external_ip": externalIP,
	}, err
}

func configureMySQLDatabase(description string, defaults conf.Map) (conf.Map, error) {
	header("MySQL DB", description)
	var (
		oldHost, oldUser, oldName, oldCharset, oldWrapper string
	)

	if defaults != nil {
		oldHost, _ = defaults.GetString("host")
		oldUser, _ = defaults.GetString("user")
		oldName, _ = defaults.GetString("name")
		oldCharset, _ = defaults.GetString("charset")
		oldWrapper, _ = defaults.GetString("wrapper")
	} else {
		oldHost = "localhost:3306"
		oldCharset = "utf8"
		oldUser = "root"
	}

	host, err := prompt.TextWithDefault("Host", oldHost, false)
	if err != nil {
		return nil, err
	}

	user, err := prompt.TextWithDefault("User", oldUser, false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	name, err := prompt.TextWithDefault("Name", oldName, false)
	if err != nil {
		return nil, err
	}

	encoding, err := prompt.TextWithDefault("Charset", oldCharset, false)
	if err != nil {
		return nil, err
	}

	wrapper, _ := prompt.TextWithDefault("Wrapper", oldWrapper, true)

	cfg := conf.Map{
		"type":     "sql",
		"driver":   "mysql",
		"host":     host,
		"user":     user,
		"password": password,
		"name":     name,
		"charset":  encoding,
	}
	hasWrapper := len(wrapper) > 0
	cfg["wrapped"] = hasWrapper
	if hasWrapper {
		cfg["wrapper"] = wrapper
	}

	return cfg, database.Create(cfg)
}

func configureSQLiteDatabase(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("SQLite DB", description)

	filename, _ := defaults.GetString("path")
	var err error
	if filename != "" {
		filename, err = prompt.TextWithDefault("path", filename, false)
	} else {
		filename, err = prompt.Text("path", false)
	}

	if err != nil {
		return nil, err
	}

	return conf.Map{
		"driver": "sqlite3",
		"type":   "sqlite",
		"path":   filename,
	}, nil
}

func configureRedisDatabase(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Redis DB", description)

	host, err := prompt.TextWithDefault("host", "localhost:6379", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.TextWithDefault("password", "", false)
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"password": password,
	}, nil
}

func configureMongoDatabase(description string, defaults conf.Map) (conf.Map, error) {
	if defaults == nil {
		defaults = conf.Map{}
	}
	header("Mongo DB", description)

	host, err := prompt.TextWithDefault("host", "localhost:27017", false)
	if err != nil {
		return nil, err
	}
	user, err := prompt.TextWithDefault("user", "", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.TextWithDefault("password", "", false)
	if err != nil {
		return nil, err
	}

	name, err := prompt.TextWithDefault("name", "", false)
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"user":     user,
		"password": password,
		"name":     name,
	}, nil
}

func header(title, description string) {
	fmt.Println()
	fmt.Println()
	fmt.Println(title)
	fmt.Println()

	if description != "" {
		fmt.Println(description)
		fmt.Println()
	}
}
