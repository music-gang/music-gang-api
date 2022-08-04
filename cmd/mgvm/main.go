package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/auth"
	"github.com/music-gang/music-gang-api/auth/jwt"
	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/executor"
	"github.com/music-gang/music-gang-api/handler"
	"github.com/music-gang/music-gang-api/http"
	applog "github.com/music-gang/music-gang-api/log"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/postgres"
	"github.com/music-gang/music-gang-api/redis"
)

var (
	Commit string
)

func init() {

	// During the init func are loaded the vm configs from env variables

	fuelRefillAmountFromConfig := config.GetConfig().APP.Vm.RefuelAmount
	if f, err := entity.ParseFuel(fuelRefillAmountFromConfig); err == nil {
		entity.FuelRefillAmount = f
	}

	fuelRefillRateFromConfig := config.GetConfig().APP.Vm.RefuelRate
	if r, err := time.ParseDuration(fuelRefillRateFromConfig); err == nil {
		entity.FuelRefillRate = r
	}

	FuelTankCapacityFromConfig := config.GetConfig().APP.Vm.MaxFuelTank
	if f, err := entity.ParseFuel(FuelTankCapacityFromConfig); err == nil {
		entity.FuelTankCapacity = f
	}

	maxExecutionTimeFromConfig := config.GetConfig().APP.Vm.MaxExecutionTime
	if t, err := time.ParseDuration(maxExecutionTimeFromConfig); err == nil {
		entity.MaxExecutionTime = t
	}
}

func main() {

	app.Commit = Commit

	// Create a context that is cancelled when the program is terminated
	ctx, cancel := context.WithCancel(context.Background())
	ctx = app.NewContextWithTags(ctx, []string{app.ContextTagCLI})

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
	ctx context.Context

	VM *mgvm.MusicGangVM

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
		VM:            mgvm.NewMusicGangVM(),
	}
}

// Close closes the main application
func (m *Main) Close() error {

	if m.VM != nil {
		if err := m.VM.Close(); err != nil {
			return err
		}
	}

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

	m.ctx = ctx

	if err := m.Postgres.Open(); err != nil {
		return err
	}

	if err := m.Redis.Open(); err != nil {
		return err
	}

	cacheStateService := redis.NewStateService(m.Redis)

	postgresAuthService := postgres.NewAuthService(m.Postgres)
	postgresUserService := postgres.NewUserService(m.Postgres)
	postgresContractService := postgres.NewContractService(m.Postgres)
	postgresStateService := postgres.NewStateService(m.Postgres)

	postgresStateService.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
		userID := app.UserIDFromContext(ctx)
		if userID == 0 {
			return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user not authorized")
		}
		if revisionID == 0 {
			return nil, apperr.Errorf(apperr.EINVALID, "revisionID is 0")
		}
		return redis.NewLockService(m.Redis, fmt.Sprintf(redis.StateLockKeyTemplate, userID, revisionID)), nil
	}
	postgresStateService.CacheStateSearchService = cacheStateService

	authService := auth.NewAuth(postgresAuthService, postgresUserService, config.GetConfig().APP.Auths)

	jwtService := jwt.NewJWTService()
	jwtService.Secret = config.GetConfig().APP.JWT.Secret
	jwtService.JWTBlacklistService = redis.NewJWTBlacklistService(m.Redis)

	logService := &applog.Logger{}

	stdOutLogger := applog.NewStdOutputLogger()
	logService.AddBackend(stdOutLogger)

	// Add slack logger if provided webhook urlLogs
	if config.GetConfig().APP.Slack.Webhook != "" {
		slackLogger := applog.NewSlackLogger(config.GetConfig().APP.Slack.Webhook)
		slackLogger.Fallback = stdOutLogger
		logService.AddBackend(slackLogger)
	}

	serviceHandler := handler.NewServiceHandler()
	serviceHandler.LogService = logService

	m.HTTPServerAPI.Addr = config.GetConfig().APP.HTTP.Addr
	m.HTTPServerAPI.Domain = config.GetConfig().APP.HTTP.Domain
	m.HTTPServerAPI.LogService = logService

	m.HTTPServerAPI.ServiceHandler = serviceHandler
	m.HTTPServerAPI.ServiceHandler.AuthSearchService = authService
	m.HTTPServerAPI.ServiceHandler.ContractSearchService = postgresContractService
	m.HTTPServerAPI.ServiceHandler.UserSearchService = postgresUserService
	m.HTTPServerAPI.ServiceHandler.JWTService = jwtService

	fuelTankService := mgvm.NewFuelTank()
	fuelTankService.LockService = redis.NewLockService(m.Redis, "fuel-tank-lock")
	fuelTankService.FuelTankService = redis.NewFuelTankService(m.Redis)

	fuelStationService := mgvm.NewFuelStation()
	fuelStationService.FuelTankService = fuelTankService
	fuelStationService.LogService = logService
	fuelStationService.FuelRefillAmount = entity.FuelRefillAmount
	fuelStationService.FuelRefillRate = entity.FuelRefillRate

	anchorageExecutor := executor.NewAnchorageContractExecutor()
	engineService := mgvm.NewEngine()
	engineService.Executors[entity.AnchorageVersion] = anchorageExecutor

	m.VM.LogService = logService
	m.VM.FuelTank = fuelTankService
	m.VM.FuelStation = fuelStationService
	m.VM.EngineService = engineService
	m.VM.ContractManagmentService = postgresContractService
	m.VM.UserManagmentService = postgresUserService
	m.VM.AuthManagmentService = authService
	m.VM.StateService = postgresStateService
	m.VM.CacheStateService = cacheStateService

	if err := m.VM.Run(); err != nil {
		return err
	}

	m.HTTPServerAPI.ServiceHandler.VmCallableService = m.VM

	if err := m.HTTPServerAPI.Open(); err != nil {
		return err
	}

	if m.HTTPServerAPI.UseTLS() {
		go func() {
			log.Fatal(http.ListenAndServeTLSRedirect(""))
		}()
	}

	m.HTTPServerAPI.LogService.ReportInfo(context.Background(), fmt.Sprintf("Starting application %s", m.HTTPServerAPI.URL()))

	return nil
}
