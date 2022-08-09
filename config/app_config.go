package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

// cfg is the config loaded from init func
var cfg Config

func init() {

	if os.Getenv("PORT") != "" {
		// Check if PORT env variable is injected by Heroku
		os.Setenv("MG_HTTP_ADDR", fmt.Sprintf(":%s", os.Getenv("PORT")))
	}

	if os.Getenv("DATABASE_URL") != "" {
		// split the DATABASE_URL into single elements
		// and set the environment variables accordingly
		// e.g. DATABASE_URL=postgres://user:pass@host:port/dbname
		// will set the following environment variables:
		// MG_PG_DATABASE=user
		// MG_PG_USER=pass
		// MG_PG_PASSWORD=host
		// MG_PG_HOST=port
		// MG_PG_PORT=dbname
		// MG_PG_SSLMODE=disable

		url := os.Getenv("DATABASE_URL")
		url = strings.TrimPrefix(url, "postgres://")

		urlParts := strings.Split(url, "@")
		if len(urlParts) != 2 {
			log.Fatal("DATABASE_URL is not valid")
		}

		userPass := strings.Split(urlParts[0], ":")
		if len(userPass) != 2 {
			log.Fatal("DATABASE_URL is not valid")
		}

		hostPortDbName := strings.Split(urlParts[1], "/")
		if len(hostPortDbName) != 2 {
			log.Fatal("DATABASE_URL is not valid")
		}

		os.Setenv("MG_PG_DATABASE", hostPortDbName[len(hostPortDbName)-1])
		os.Setenv("MG_PG_USER", userPass[0])
		os.Setenv("MG_PG_PASSWORD", userPass[1])
		os.Setenv("MG_PG_HOST", strings.Split(hostPortDbName[0], ":")[0])
		os.Setenv("MG_PG_PORT", strings.Split(hostPortDbName[0], ":")[1])
		os.Setenv("MG_PG_SSL_MODE", "require")
	}

	if os.Getenv("REDIS_URL") != "" {
		// split the REDIS_URL into single elements
		// and set the environment variables accordingly
		// e.g. REDIS_URL=redis://:pass@host:port
		// will set the following environment variables:
		// MG_REDIS_HOST=host
		// MG_REDIS_PORT=port
		// MG_REDIS_PASSWORD=pass

		url := os.Getenv("REDIS_URL")
		url = strings.TrimPrefix(url, "redis://")

		urlParts := strings.Split(url, "@")
		if len(urlParts) != 2 {
			log.Fatal("REDIS_URL is not valid")
		}

		userPass := strings.Split(urlParts[0], ":")
		if len(userPass) != 2 {
			log.Fatal("DATABASE_URL is not valid")
		}

		hostPort := strings.Split(urlParts[1], ":")
		if len(hostPort) != 2 {
			log.Fatal("REDIS_URL is not valid")
		}

		os.Setenv("MG_REDIS_HOST", hostPort[0])
		os.Setenv("MG_REDIS_PORT", hostPort[1])
		os.Setenv("MG_REDIS_PASSWORD", userPass[1])
	}

	if err := env.Parse(&cfg, env.Options{
		Prefix: "MG_",
	}); err != nil {
		panic(err)
	}
}

type AuthConfig struct {
	// ClientID is the application's ID.
	ClientID string `env:"CLIENT_ID"`

	// ClientSecret is the application's secret.
	ClientSecret string `env:"CLIENT_SECRET"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint struct {
		AuthURL  string `env:"AUTH_URL"`
		TokenURL string `env:"TOKEN_URL"`

		// AuthStyle optionally specifies how the endpoint wants the
		// client ID & client secret sent. The zero value means to
		// auto-detect.
		AuthStyle int `env:"AUTH_STYLE"`
	}

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `env:"REDIRECT_URL"`

	// Scope specifies optional requested permissions.
	Scopes []string `env:"SCOPES" envSeparator:","`
}

// AuthListConfig contains the list of auth configs
type AuthListConfig struct {
	// Local is the local auth config
	Local AuthConfig `envPrefix:"LOCAL_"`

	// Github is the github auth config
	Github AuthConfig `envPrefix:"GITHUB_"`
}

type DatabaseConfig struct {
	// Database is the name of the database to connect to.
	Database string `env:"DATABASE" envDefault:"musicgang"`

	// User is the database user to sign in as.
	User string `env:"USER" default:"postgres"`

	// Password is the user's password.
	Password string `env:"PASSWORD" envDefault:"admin"`

	// Host is the host to connect to. Values that start with / are for unix domain sockets.
	Host string `env:"HOST" envDefault:"localhost"`

	// Port is the port to connect to.
	Port int `env:"PORT" envDefault:"5432"`

	// SSLMode specifies whether to use SSL.
	SSLMode string `env:"SSL_MODE" envDefault:"disable"`
}

// HTTPConfig contains the http config
type HTTPConfig struct {
	Domain string `env:"DOMAIN"`
	Addr   string `env:"ADDR" envDefault:":8888"`
}

// JWTConfig contains the jwt config
type JWTConfig struct {
	Secret           string `env:"SECRET"`
	ExpiresIn        int    `env:"EXPIRES_IN" envDefault:"60"`
	RefreshExpiresIn int    `env:"REFRESH_EXPIRES_IN" envDefault:"20160"`
}

// RedisConfig contains the redis config
type RedisConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD" envDefault:""`
}

// DatabaseListConfig contains the list of database configs
type DatabaseListConfig struct {
	// Postgres is the Postgres database configuration
	Postgres DatabaseConfig `envPrefix:"PG_"`
}

// VmConfig contains the vm config
type VmConfig struct {
	MaxFuelTank      string `env:"MAX_FUEL_TANK" envDefault:"100 vKFuel"`
	MaxExecutionTime string `env:"MAX_EXECUTION_TIME" envDefault:"10s"`
	RefuelAmount     string `env:"REFUEL_AMOUNT" envDefault:""`
	RefuelRate       string `env:"REFUEL_RATE" envDefault:"400ms"`
}

type AppConfig struct {
	// HTTP is the http config
	HTTP HTTPConfig `envPrefix:"HTTP_"`

	// JWT is the jwt config
	JWT JWTConfig `envPrefix:"JWT_"`

	// Databases contains the databases configuration
	Databases DatabaseListConfig

	// Auths contains the auths configuration
	Auths AuthListConfig `envPrefix:"AUTH_"`

	// Redis contains the redis configuration
	Redis RedisConfig `envPrefix:"REDIS_"`

	// Vm contains the vm configuration
	Vm VmConfig `envPrefix:"VM_"`
}

// Config - Configuration
type Config struct {
	// APP contains the application configuration
	APP AppConfig
}

// BuildDSNFromDatabaseConfigForPostgres returns a DSN string for Postgres
func BuildDSNFromDatabaseConfigForPostgres(dbConfig DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=UTC", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.SSLMode)
}

// GetConfig returns the config
func GetConfig() Config {
	return cfg
}
