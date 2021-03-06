package redis_test

import (
	"fmt"
	"testing"

	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/redis"
)

func TestRedis(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		db := MustOpenDB(t)
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TimeoutPing", func(t *testing.T) {

		redisHost := config.GetConfig().APP.Redis.Host
		redisPort := config.GetConfig().APP.Redis.Port
		redisPassword := config.GetConfig().APP.Redis.Password

		redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

		db := redis.NewDB(redisAddr, redisPassword)

		db.Cancel()

		if db.Open() == nil {
			t.Fatal("Expected error, but got nil")
		}
	})

	t.Run("CloseNotOpened", func(t *testing.T) {

		redisHost := config.GetConfig().APP.Redis.Host
		redisPort := config.GetConfig().APP.Redis.Port
		redisPassword := config.GetConfig().APP.Redis.Password

		redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

		db := redis.NewDB(redisAddr, redisPassword)

		db.Close()

	})
}

func MustOpenDB(tb testing.TB) *redis.DB {

	tb.Helper()

	redisHost := config.GetConfig().APP.Redis.Host
	redisPort := config.GetConfig().APP.Redis.Port
	redisPassword := config.GetConfig().APP.Redis.Password

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	db := redis.NewDB(redisAddr, redisPassword)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}

	return db
}
