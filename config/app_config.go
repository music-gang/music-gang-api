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

var cfg Config

type DatabaseConfig struct {
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
}

type DatabaseListConfig struct {
	Postgres DatabaseConfig `yaml:"postgres"`
}

type AppConfig struct {
	Databases DatabaseListConfig `yaml:"databases"`
}

// Config - Configuration
type Config struct {
	APP  AppConfig `yaml:"app"`
	TEST AppConfig `yaml:"test"`
}

type LoadOptions struct {
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

// getConfigFilePath - Restituisce il percorso del file di configurazione
func getConfigFilePath() string {
	wd := util.GetWd()
	return wd + "/" + configFile
}
