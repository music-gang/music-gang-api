package common

import (
	"errors"
	"os"
)

const (
	CustomLogsDirPath = "custom/logs"
)

// FileExists - Restituisce se eiste il file al path passato
func FileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// GetWd - Restituisce il path assoluto verso la working directory
func GetWd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

// CreateIfNotExistsFolder - Crea la cartella passata se non esiste
func CreateIfNotExistsFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
