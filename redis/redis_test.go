package redis_test

import (
	"fmt"
	"testing"

	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/redis"
)

func TestRedis(t *testing.T) {

	config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"})

	redisHost := config.GetConfig().TEST.Redis.Host
	redisPort := config.GetConfig().TEST.Redis.Port
	redisPassword := config.GetConfig().TEST.Redis.Password

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	db := redis.NewDB(redisAddr, redisPassword)
	db.MustOpen()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
