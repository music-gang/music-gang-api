package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/auth"
	"github.com/music-gang/music-gang-api/auth/jwt"
	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/http"
	applog "github.com/music-gang/music-gang-api/log"
	"github.com/music-gang/music-gang-api/postgres"
	"github.com/music-gang/music-gang-api/redis"
)

var (
	Commit string
)

func init() {
	if err := config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "../config.yaml"}); err != nil {
		if err = config.LoadConfigWithOptions(config.LoadOptions{ConfigFilePath: "config.yaml"}); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
}

func main() {

	app.Commit = Commit

	// Create a context that is cancelled when the program is terminated
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	m := NewMain()

	if err := m.Run(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()

	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	Postgres *postgres.DB

	Redis *redis.DB

	HTTPServerAPI *http.ServerAPI
}

// NewMain returns a new instance of Main
func NewMain() *Main {

	redisHost := config.GetConfig().APP.Redis.Host
	redisPort := config.GetConfig().APP.Redis.Port
	redisPassword := config.GetConfig().APP.Redis.Password

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	return &Main{
		Postgres:      postgres.NewDB(config.BuildDSNFromDatabaseConfigForPostgres(config.GetConfig().APP.Databases.Postgres)),
		Redis:         redis.NewDB(redisAddr, redisPassword),
		HTTPServerAPI: http.NewServerAPI(),
	}
}

// Close closes the main application
func (m *Main) Close() error {

	if m.HTTPServerAPI != nil {
		if err := m.HTTPServerAPI.Close(); err != nil {
			return err
		}
	}

	if m.Postgres != nil {
		if err := m.Postgres.Close(); err != nil {
			return err
		}
	}

	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Run starts the main application
func (m *Main) Run(ctx context.Context) error {

	if err := m.Postgres.Open(); err != nil {
		return err
	}

	if err := m.Redis.Open(); err != nil {
		return err
	}

	postgresAuthService := postgres.NewAuthService(m.Postgres)
	postgresUserService := postgres.NewUserService(m.Postgres)

	authService := auth.NewAuth(postgresAuthService, postgresUserService, config.GetConfig().APP.Auths)

	jwtService := jwt.NewJWTService()
	jwtService.Secret = config.GetConfig().APP.JWT.Secret
	jwtService.JWTBlacklistService = redis.NewJWTBlacklistService(m.Redis)

	m.HTTPServerAPI.Addr = config.GetConfig().APP.HTTP.Addr
	m.HTTPServerAPI.Domain = config.GetConfig().APP.HTTP.Domain
	m.HTTPServerAPI.UserService = postgresUserService
	m.HTTPServerAPI.AuthService = authService
	m.HTTPServerAPI.JWTService = jwtService

	logService := &applog.Logger{}
	logService.AddBackend(applog.NewStdOutputLogger())

	m.HTTPServerAPI.LogService = logService

	if err := m.HTTPServerAPI.Open(); err != nil {
		return err
	}

	if m.HTTPServerAPI.UseTLS() {
		go func() {
			log.Fatal(http.ListenAndServeTLSRedirect(""))
		}()
	}

	m.HTTPServerAPI.LogService.ReportInfo(context.Background(), fmt.Sprintf("Starting application on: %s", m.HTTPServerAPI.URL()))

	return nil
}
