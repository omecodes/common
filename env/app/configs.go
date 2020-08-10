package app

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/omecodes/common/database"
	"github.com/omecodes/common/errors"
	"github.com/omecodes/common/netx"
	"github.com/omecodes/common/security/crypto/key"
	"github.com/omecodes/common/utils/jcon"
	"github.com/omecodes/common/utils/prompt"
	"strings"
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
	ConfigSecretKeys
	ConfigNetwork
	ConfigMySQLDatabase
	ConfigSQLiteDatabase
	ConfigRedisDatabase
	ConfigMongoDatabase
	ConfigOauth2Providers
	ConfigOme
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

	case ConfigSecretKeys:
		return "secret-keys"

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

	case ConfigOauth2Providers:
		return "oauth2-providers"

	case ConfigOme:
		return "ome"

	default:
		return ""
	}
}

func (ci configItem) create(description string, defaults jcon.Map) (jcon.Map, error) {
	switch ci.configType {

	case ConfigAccess:
		return configureAccess(description, defaults)

	case ConfigMailer:
		return configureMailer(description, defaults)

	case ConfigAdminsCredentials:
		return configureAdminsCredentials(description, defaults)

	case ConfigDirs:
		return configureDirs(description, defaults, ci.entries...)

	case ConfigSecretKeys:
		return configureSecretKeys(description, defaults, ci.entries...)

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

	case ConfigOauth2Providers:
		return configureOauth2Providers(description, defaults)

	case ConfigOme:
		return configureOme(description, defaults)

	default:
		return nil, errors.NotSupported
	}
}

func configureAccess(description string, defaults jcon.Map) (jcon.Map, error) {
	header("Access", description)

	cfg := jcon.Map{}

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

func configureSecrets(description string, defaults jcon.Map) (jcon.Map, error) {
	header("Secrets", description)

	var err error
	count := 0
	if defaults == nil {
		defaults = jcon.Map{}
	}

	cfg := jcon.Map{}

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

		oldValue, _ := defaults.GetString(name)
		secret, err := prompt.TextWithDefault("Secret", oldValue, false)
		if err != nil {
			return nil, err
		}
		cfg[name] = secret
		fmt.Println()
	}
	return cfg, err
}

func configureCredentialsTable(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
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
	return jcon.Map{"subject": key, "password": secret}, nil
}

func configureMailer(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
	}
	header("Mailer", description)
	selected, err := prompt.Selection("", "MailHog (for tests)", "SMTP client (default)", "sendGrid")
	if err != nil {
		return nil, err
	}
	switch selected {
	case "SMTP client (default)":
		return configureSMTP(defaults)

	case "sendGrid":
		return configureSendGridMailer(defaults)

	default:
		return configureHogMailer()
	}
}

func configureHogMailer() (jcon.Map, error) {
	return jcon.Map{
		"type":     "hog",
		"server":   "127.0.0.1",
		"port":     1025,
		"user":     "Ome",
		"password": "Ome",
	}, nil
}

func configureSMTP(defaults jcon.Map) (jcon.Map, error) {
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
	return jcon.Map{
		"type":     "smtp",
		"server":   server,
		"port":     port,
		"user":     user,
		"password": password,
	}, nil
}

func configureSendGridMailer(defaults jcon.Map) (jcon.Map, error) {
	var err error

	server, _ := defaults.GetString("host")
	if server == "" {
		server = "https://api.sendgrid.com"
	}
	endpoint, _ := defaults.GetString("endpoints")
	if endpoint == "" {
		endpoint = "/v3/mail/send"
	}
	key, _ := defaults.GetString("key")

	server, err = prompt.TextWithDefault("server", server, false)
	if err != nil {
		return nil, err
	}

	endpoint, err = prompt.TextWithDefault("endpoint", endpoint, false)
	if err != nil {
		return nil, err
	}

	key, err = prompt.TextWithDefault("key", key, false)
	if err != nil {
		return nil, err
	}

	return jcon.Map{
		"type":     "sendgrid",
		"host":     server,
		"endpoint": endpoint,
		"key":      key,
	}, nil
}

func configureAdminsCredentials(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
	}
	header("Admins credentials", description)
	var err error
	count := 0
	cfg := jcon.Map{}

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

func configureDirs(description string, defaults jcon.Map, names ...string) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
	}

	header("Directories", description)
	var err error
	count := 0
	cfg := jcon.Map{}

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

func configureSecretKeys(description string, defaults jcon.Map, names ...string) (jcon.Map, error) {
	header("Secret keys", description)

	var err error
	if defaults == nil {
		defaults = jcon.Map{}
	}

	cfg := jcon.Map{}

	passPhrase, err := prompt.Password("pass phrase")
	if err != nil {
		return nil, err
	}

	passPhraseConfirm, err := prompt.Password("confirm")
	if err != nil {
		return nil, err
	}

	if passPhrase != passPhraseConfirm {
		return nil, errors.New("pass phrases differ")
	}

	for _, name := range names {
		_, info, err := key.Generate(passPhrase, 32)
		if err != nil {
			return nil, err
		}
		cfg[name] = info
		// fmt.Println()
	}
	return cfg, err
}

func configureNetwork(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
	}
	header("Network", description)

	var err error
	domain, _ := defaults.GetString("domain")
	internalIP, _ := defaults.GetString("internal_ip")
	externalIP, _ := defaults.GetString("external_ip")

	addrList := netx.LocalAddresses()
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

	return jcon.Map{
		"domain":      domain,
		"internal_ip": internalIP,
		"external_ip": externalIP,
	}, err
}

func configureMySQLDatabase(description string, defaults jcon.Map) (jcon.Map, error) {
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

	cfg := jcon.Map{
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

func configureSQLiteDatabase(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
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

	return jcon.Map{
		"driver": "sqlite3",
		"type":   "sqlite",
		"path":   filename,
	}, nil
}

func configureRedisDatabase(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
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

	return jcon.Map{
		"host":     host,
		"password": password,
	}, nil
}

func configureMongoDatabase(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
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

	return jcon.Map{
		"host":     host,
		"user":     user,
		"password": password,
		"name":     name,
	}, nil
}

func configureOauth2Providers(description string, defaults jcon.Map) (jcon.Map, error) {
	if defaults == nil {
		defaults = jcon.Map{}
	}
	header("OAUTH2 Providers", description)
	var err error
	count := 0
	cfg := jcon.Map{}

	for {
		if count > 0 {
			selection, err := prompt.Selection("add another provider?", "yes", "no")
			if err != nil {
				return nil, err
			}

			if selection != "yes" {
				break
			}
		}
		count++

		name, err := prompt.Text("name", false)
		if err != nil {
			return nil, err
		}

		defaultsValues := defaults.GetConf(name)
		if defaultsValues == nil {
			defaultsValues = jcon.Map{}
		}

		oldLabel, _ := defaultsValues.GetString("info/label")
		label, err := prompt.TextWithDefault("display name", oldLabel, false)
		if err != nil {
			return nil, err
		}

		oldLogoURL, _ := defaultsValues.GetString("info/log_url")
		logoURL, err := prompt.TextWithDefault("logo URL", oldLogoURL, false)
		if err != nil {
			return nil, err
		}

		oldServerURL, _ := defaultsValues.GetString("config/server_url")
		serverURL, err := prompt.TextWithDefault("serverURL", oldServerURL, false)
		if err != nil {
			return nil, err
		}

		oldClientID, _ := defaultsValues.GetString("config/client_id")
		clientID, err := prompt.TextWithDefault("client ID", oldClientID, false)
		if err != nil {
			return nil, err
		}

		oldSecret, _ := defaultsValues.GetString("config/secret")
		clientSecret, err := prompt.TextWithDefault("client secret", oldSecret, false)
		if err != nil {
			return nil, err
		}

		cfg[name] = jcon.Map{
			"config": jcon.Map{
				"server_url": serverURL,
				"client_id":  clientID,
				"secret":     clientSecret,
			},
			"info": jcon.Map{
				"label":    label,
				"logo_url": logoURL,
			},
		}
		fmt.Println()
	}
	return cfg, err
}

func configureOme(description string, defaults jcon.Map) (jcon.Map, error) {
	header("Ome Authority", description)
	if defaults == nil {
		defaults = jcon.Map{}
	}

	url, _ := defaults.GetString("url")
	url, err := prompt.TextWithDefault("server url", url, false)
	if err != nil {
		return nil, err
	}

	accessKey, _ := defaults.GetString("oauth2/client_id")
	accessKey, err = prompt.TextWithDefault("client ID", accessKey, false)
	if err != nil {
		return nil, err
	}

	secret, err := prompt.Password("secret")
	if err != nil {
		return nil, err
	}

	tokenEndpoint, _ := defaults.GetString("oauth2/token_endpoint")
	if tokenEndpoint == "" {
		tokenEndpoint = "/token"
	}
	tokenEndpoint, err = prompt.TextWithDefault("token endpoint", tokenEndpoint, false)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(tokenEndpoint, "/") {
		tokenEndpoint = "/" + tokenEndpoint
	}

	authorizeEndpoint, _ := defaults.GetString("oauth2/authorize_endpoint")
	if authorizeEndpoint == "" {
		authorizeEndpoint = "/authorize"
	}
	authorizeEndpoint, err = prompt.TextWithDefault("authorize endpoint", authorizeEndpoint, false)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(authorizeEndpoint, "/") {
		authorizeEndpoint = "/" + authorizeEndpoint
	}

	return jcon.Map{
		"oauth2": jcon.Map{
			"server_url":         url,
			"client_id":          accessKey,
			"secret":             secret,
			"token_endpoint":     tokenEndpoint,
			"authorize_endpoint": authorizeEndpoint,
		},
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
