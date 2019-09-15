package configs

import (
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/database"
	"github.com/zoenion/common/network"
	"github.com/zoenion/common/prompt"
)

const (
	ConfDatabase = "database"
	ConfAdmin    = "admin"
	ConfMailer   = "mailer"
	ConfAccess   = "access"
)

type Access struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func (access *Access) Prompt() error {
	var err error

	access.Key, err = prompt.TextWithDefault("key", access.Key, false)
	if err != nil {
		return err
	}

	access.Secret, err = prompt.Password("secret")
	return err
}

type Databases map[string]conf.Map

func (dbs Databases) Prompt() error {

	prompt.Header("Databases")
	var err error

	for {
		choice, err := prompt.Selection("Set database config?", []string{"MySQL", "Mongo", "Redis", "No"})
		if err != nil {
			return err
		}

		if choice == "No" {
			break
		}

		name, err := prompt.Text("Config name", false)
		if err != nil {
			return err
		}

		oldDbCfg := dbs[name]

		var dbCfg conf.Map
		if choice == "MySQL" {
			dbCfg, err = MysqlPrompt(oldDbCfg)

		} else if choice == "Mongo" {
			dbCfg, err = MongoPrompt(oldDbCfg)

		} else {
			dbCfg, err = RedisPrompt(oldDbCfg)
		}

		if err != nil {
			return err
		}

		dbs[name] = dbCfg
	}
	return err
}

func MysqlPrompt(old conf.Map) (conf.Map, error) {
	var (
		oldHost, oldUser, oldName, oldCharset, oldWrapper string
	)

	if old != nil {
		oldHost, _ = old.GetString("host")
		oldUser, _ = old.GetString("user")
		oldName, _ = old.GetString("name")
		oldCharset, _ = old.GetString("charset")
		oldWrapper, _ = old.GetString("wrapper")
	} else {
		oldHost = "localhost:3306"
		oldCharset = "utf8"
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
	return cfg, nil
}

func MongoPrompt(old conf.Map) (conf.Map, error) {
	host, err := prompt.TextWithDefault("Host", "localhost:27017", false)
	if err != nil {
		return nil, err
	}
	user, err := prompt.Text("User", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	name, err := prompt.Text("Name", false)
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

func RedisPrompt(old conf.Map) (conf.Map, error) {
	host, err := prompt.TextWithDefault("Host", "localhost:6379", false)
	if err != nil {
		return nil, err
	}

	password, err := prompt.Password("Password")
	if err != nil {
		return nil, err
	}

	return conf.Map{
		"host":     host,
		"password": password,
	}, nil
}

func SQLitePrompt(old conf.Map) (conf.Map, error) {
	var dbFilepath string

	if old != nil {
		dbFilepath, _ = old.GetString("path")
	}

	p, err := prompt.TextWithDefault("store path", dbFilepath, false)
	if err != nil {
		return nil, err
	}

	cfg := database.SQLiteConfig(p)
	return cfg, nil
}

type Mailer struct {
	Server   string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (mailer *Mailer) Prompt() error {
	var err error

	mailer.Server, err = prompt.TextWithDefault("server", mailer.Server, false)
	if err != nil {
		return err
	}

	port, err := prompt.IntegerWithDefaultValue("Port", int64(mailer.Port))
	if err != nil {
		return nil
	}
	mailer.Port = int(port)

	mailer.User, err = prompt.TextWithDefault("User", mailer.User, false)
	if err != nil {
		return err
	}

	mailer.Password, err = prompt.Password("Password")
	if err != nil {
		return err
	}
	return err
}

func (mailer *Mailer) ToConf() conf.Map {
	cfg := conf.Map{}
	cfg["host"] = mailer.Server
	cfg["port"] = mailer.Port
	cfg["user"] = mailer.User
	cfg["password"] = mailer.Password
	return cfg
}

type Network struct {
	IP     string `json:"ip"`
	Domain string `json:"domain"`
}

func (n *Network) Prompt() error {
	var err error

	addrList := network.LocalAddresses()
	n.IP, err = prompt.Selection("ip", addrList)
	if err != nil {
		return err
	}
	n.Domain, err = prompt.TextWithDefault("domain", n.Domain, true)
	return err
}

type Security struct {
	Filename string
	Type     int
}

/*func PromptCertificate(ca bool, dir string, name string, password []byte) (*x509.Certificate, crypto.PrivateKey, error) {
	localAddresses := network.LocalAddresses()
	publicAddress := network.PublicAddresses()
	addresses := append(localAddresses, publicAddress...)

	ip, err := prompt.Selection("ip", addresses)
	if err != nil {
		return nil, nil, err
	}

	domain, err := prompt.Text("domain", true)
	if err != nil {
		return nil, nil, err
	}

	certFile := filepath.Join(dir, name+".crt")
	keyFile := filepath.Join(dir, name+".key")

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	var ips []net.IP
	var cert *x509.Certificate

	ips = append(ips, net.ParseIP(ip))

	if ca {
		cert, err = crypto2.GenerateCACertificate(&crypto2.Template{
			Name:             name,
			Domains:          []string{domain},
			IPs:              ips,
			Expiry:           time.Hour * 24 * 5 * 365, // about 5 years
			PublicKey:        privateKey.Public(),
			SignerPrivateKey: privateKey,
		})
		if err != nil {
			return nil, nil, err
		}
	} else {
		cert, err = crypto2.GenerateServiceCertificate(&crypto2.Template{
			Name:             name,
			Domains:          []string{domain},
			IPs:              ips,
			Expiry:           time.Hour * 24 * 5 * 365, // about 5 years
			PublicKey:        privateKey.Public(),
			SignerPrivateKey: privateKey,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	if err := crypto2.StoreCertificate(cert, certFile, os.ModePerm); err != nil {
		log.Println("could not save certificate")
	}

	if err := crypto2.StorePrivateKey(privateKey, password, keyFile); err != nil {
		log.Println("could not save private key")
	}

	return cert, privateKey, nil
}*/
