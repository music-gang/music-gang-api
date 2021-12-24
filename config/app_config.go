package config

import (
	"fmt"
	"os"

	"github.com/music-gang/music-gang-api/app/util"
	"gopkg.in/yaml.v2"
)

const (
	configFile = "config.yaml"
)

// cfg is the config loaded from LoadConfigWithOptions
var cfg Config

type AuthConfig struct {
	// ClientID is the application's ID.
	ClientID string `yaml:"client_id"`

	// ClientSecret is the application's secret.
	ClientSecret string `yaml:"client_secret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint struct {
		AuthURL  string `yaml:"auth_url"`
		TokenURL string `yaml:"token_url"`

		// AuthStyle optionally specifies how the endpoint wants the
		// client ID & client secret sent. The zero value means to
		// auto-detect.
		AuthStyle int `yaml:"auth_style"`
	} `yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `yaml:"redirect_url"`

	// Scope specifies optional requested permissions.
	Scopes []string `yaml:"scopes"`
}

// AuthListConfig contains the list of auth configs
type AuthListConfig struct {
	// Local is the local auth config
	Local AuthConfig `yaml:"local"`

	// Github is the github auth config
	Github AuthConfig `yaml:"github"`
}

type DatabaseConfig struct {
	// Database is the name of the database to connect to.
	Database string `yaml:"database"`

	// User is the database user to sign in as.
	User string `yaml:"user"`

	// Password is the user's password.
	Password string `yaml:"password"`

	// Host is the host to connect to. Values that start with / are for unix domain sockets.
	Host string `yaml:"host"`

	// Port is the port to connect to.
	Port int `yaml:"port"`
}

// HTTPConfig contains the http config
type HTTPConfig struct {
	Domain string `yaml:"domain"`
	Addr   string `yaml:"addr"`
}

// JWTConfig contains the jwt config
type JWTConfig struct {
	Secret           string `yaml:"secret"`
	ExpiresIn        int    `yaml:"expiresIn"`
	RefreshExpiresIn int    `yaml:"refreshExpiresIn"`
}

// RedisConfig contains the redis config
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type DatabaseListConfig struct {
	// Postgres is the Postgres database configuration
	Postgres DatabaseConfig `yaml:"postgres"`
}

type AppConfig struct {
	// HTTP is the http config
	HTTP HTTPConfig `yaml:"http"`

	// JWT is the jwt config
	JWT JWTConfig `yaml:"jwt"`

	// Databases contains the databases configuration
	Databases DatabaseListConfig `yaml:"databases"`

	// Auths contains the auths configuration
	Auths AuthListConfig `yaml:"auths"`

	// Redis contains the redis configuration
	Redis RedisConfig `yaml:"redis"`
}

// Config - Configuration
type Config struct {
	// APP contains the application configuration
	APP AppConfig `yaml:"app"`

	// TEST contains the test configuration
	TEST AppConfig `yaml:"test"`
}

type LoadOptions struct {
	// ConfigFilePath is the path to the config file
	ConfigFilePath string
}

// BuildDSNFromDatabaseConfigForPostgres returns a DSN string for Postgres
func BuildDSNFromDatabaseConfigForPostgres(dbConfig DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&TimeZone=UTC", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
}

// GetConfig returns the config
func GetConfig() Config {
	return cfg
}

// LoadConfig loads config
func LoadConfig() error {
	return LoadConfigWithOptions(LoadOptions{})
}

// LoadConfigWithOptions loads config with options
func LoadConfigWithOptions(opt LoadOptions) error {

	configFilePath := getConfigFilePath()

	if opt.ConfigFilePath != "" {
		configFilePath = opt.ConfigFilePath
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	if err := decoder.Decode(&cfg); err != nil {
		return err
	}

	return nil
}

// getConfigFilePath returns the config file path
func getConfigFilePath() string {
	wd := util.GetWd()
	return wd + "/" + configFile
}
