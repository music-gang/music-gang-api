package config

import (
	"os"

	"github.com/music-gang/music-gang-api/app/util"
	"gopkg.in/yaml.v2"
)

func init() {
	if err := loadConfig(); err != nil {
		panic(err)
	}
}

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
	SQLServer DatabaseConfig `yaml:"sql_server"`
}

type AppConfig struct {
	Databases DatabaseListConfig `yaml:"databases"`
}

// Config - Configuration
type Config struct {
	APP AppConfig `yaml:"app"`
}

// GetConfig - Restituisce la configurazione dell'applicazione
func GetConfig() Config {
	return cfg
}

// getConfigFilePath - Restituisce il percorso del file di configurazione
func getConfigFilePath() string {
	wd := util.GetWd()
	return wd + "/" + configFile
}

// MARK: Unexported funcs

// loadConfig - Si occupa di caricare la configurazione dell'applicazione
func loadConfig() error {

	file, err := os.Open(getConfigFilePath())
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
