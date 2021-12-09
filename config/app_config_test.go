package config_test

import (
	"reflect"
	"testing"

	"github.com/music-gang/music-gang-api/config"
)

func TestAppConfig_LoadConfig(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		if err := config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"}); err != nil {
			t.Fatal(err)
		}

		if reflect.DeepEqual(config.GetConfig(), config.Config{}) {
			t.Fatal("Config is empty")
		}
	})

	t.Run("InvalidPathErr", func(t *testing.T) {

		if err := config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "/some-wrong-path/config.yaml"}); err == nil {
			t.Fatal("Expected error")
		}
	})
}
