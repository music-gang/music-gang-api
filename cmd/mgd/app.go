package main

import (
	"context"
	"fmt"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/auth"
	"github.com/music-gang/music-gang-api/auth/jwt"
	"github.com/music-gang/music-gang-api/config"
	"github.com/music-gang/music-gang-api/event"
	"github.com/music-gang/music-gang-api/executor"
	"github.com/music-gang/music-gang-api/handler"
	"github.com/music-gang/music-gang-api/http"
	"github.com/music-gang/music-gang-api/mgvm"
	"github.com/music-gang/music-gang-api/postgres"
	"github.com/music-gang/music-gang-api/redis"
)

type App struct {
	ctx context.Context

	VM *mgvm.MusicGangVM

	Postgres *postgres.DB

	Redis *redis.DB

	HTTPServerAPI *http.ServerAPI

	EventService *event.EventService
}

// NewApp returns a new instance of Main
func NewApp() *App {

	redisHost := config.GetConfig().APP.Redis.Host
	redisPort := config.GetConfig().APP.Redis.Port
	redisPassword := config.GetConfig().APP.Redis.Password

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	return &App{
		Postgres:      postgres.NewDB(config.BuildDSNFromDatabaseConfigForPostgres(config.GetConfig().APP.Databases.Postgres)),
		Redis:         redis.NewDB(redisAddr, redisPassword),
		HTTPServerAPI: http.NewServerAPI(),
		VM:            mgvm.NewMusicGangVM(),
		EventService:  event.NewEventService(),
	}
}

// Close closes the main application
func (a *App) Close() error {

	if a.VM != nil {
		if err := a.VM.Close(); err != nil {
			return err
		}
	}

	if a.HTTPServerAPI != nil {
		if err := a.HTTPServerAPI.Close(); err != nil {
			return err
		}
	}

	if a.Postgres != nil {
		if err := a.Postgres.Close(); err != nil {
			return err
		}
	}

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Run starts the main application
func (a *App) Run(ctx context.Context) error {

	a.ctx = ctx

	if err := a.Postgres.Open(); err != nil {
		return err
	}

	if err := a.Redis.Open(); err != nil {
		return err
	}

	cacheStateService := redis.NewStateService(a.Redis)

	postgresAuthService := postgres.NewAuthService(a.Postgres)
	postgresUserService := postgres.NewUserService(a.Postgres)
	postgresContractService := postgres.NewContractService(a.Postgres)
	postgresStateService := postgres.NewStateService(a.Postgres)

	postgresStateService.CreateLockService = func(ctx context.Context, revisionID int64) (service.LockService, error) {
		userID := app.UserIDFromContext(ctx)
		if userID == 0 {
			return nil, apperr.Errorf(apperr.EUNAUTHORIZED, "user not authorized")
		}
		if revisionID == 0 {
			return nil, apperr.Errorf(apperr.EINVALID, "revisionID is 0")
		}
		return redis.NewLockService(a.Redis, fmt.Sprintf(redis.StateLockKeyTemplate, userID, revisionID)), nil
	}
	postgresStateService.CacheStateSearchService = cacheStateService

	authService := auth.NewAuth(postgresAuthService, postgresUserService, config.GetConfig().APP.Auths)

	jwtService := jwt.NewJWTService()
	jwtService.Secret = config.GetConfig().APP.JWT.Secret
	jwtService.JWTBlacklistService = redis.NewJWTBlacklistService(a.Redis)

	logService := log15.New("app", "mgd")

	a.HTTPServerAPI.Addr = config.GetConfig().APP.HTTP.Addr
	a.HTTPServerAPI.Domain = config.GetConfig().APP.HTTP.Domain
	a.HTTPServerAPI.LogService = logService.New("module", "http")

	a.HTTPServerAPI.ServiceHandler = handler.NewServiceHandler()
	a.HTTPServerAPI.ServiceHandler.AuthSearchService = authService
	a.HTTPServerAPI.ServiceHandler.ContractSearchService = postgresContractService
	a.HTTPServerAPI.ServiceHandler.UserSearchService = postgresUserService
	a.HTTPServerAPI.ServiceHandler.JWTService = jwtService
	a.HTTPServerAPI.ServiceHandler.Logger = a.HTTPServerAPI.LogService

	fuelTankService := mgvm.NewFuelTank()
	fuelTankService.LockService = redis.NewLockService(a.Redis, "fuel-tank-lock")
	fuelTankService.FuelTankService = redis.NewFuelTankService(a.Redis)

	fuelStationService := mgvm.NewFuelStation()
	fuelStationService.FuelTankService = fuelTankService
	fuelStationService.LogService = logService.New("module", "fuel-station")
	fuelStationService.FuelRefillAmount = entity.FuelRefillAmount
	fuelStationService.FuelRefillRate = entity.FuelRefillRate

	anchorageExecutor := executor.NewAnchorageContractExecutor()
	engineService := mgvm.NewEngine()
	engineService.Executors[entity.AnchorageVersion] = anchorageExecutor

	fuelMonitorService := mgvm.NewFuelMonitor()
	fuelMonitorService.EngineStateService = engineService
	fuelMonitorService.EventService = a.EventService
	fuelMonitorService.FuelService = fuelTankService
	fuelMonitorService.LogService = logService.New("module", "fuel-monitor")

	InitializerCPUsPool()

	cpusPoolService := mgvm.NewCPUsPool()

	a.VM.LogService = logService.New("module", "vm")

	a.VM.EventService = a.EventService

	a.VM.FuelTank = fuelTankService
	a.VM.FuelStation = fuelStationService
	a.VM.FuelMonitor = fuelMonitorService
	a.VM.EngineService = engineService
	a.VM.CPUsPoolService = cpusPoolService

	a.VM.ContractManagmentService = postgresContractService
	a.VM.UserManagmentService = postgresUserService
	a.VM.AuthManagmentService = authService
	a.VM.StateService = postgresStateService
	a.VM.CacheStateService = cacheStateService

	if err := a.VM.Run(); err != nil {
		return err
	}

	a.HTTPServerAPI.ServiceHandler.VmCallableService = a.VM

	if err := a.HTTPServerAPI.Open(); err != nil {
		return err
	}

	if a.HTTPServerAPI.UseTLS() {
		go func() {
			log.Fatal(http.ListenAndServeTLSRedirect(""))
		}()
	}

	logService.Info("Starting application",
		"addr", a.HTTPServerAPI.Addr,
		"domain", a.HTTPServerAPI.Domain,
		"tls", a.HTTPServerAPI.UseTLS(),
		"vm_fuel_tank_capacity", entity.FuelTankCapacity,
		"vm_fuel_refill_amount", entity.FuelRefillAmount,
		"vm_fuel_refill_rate", entity.FuelRefillRate,
		"vm_max_execution_time", entity.MaxExecutionTime,
	)

	return nil
}
