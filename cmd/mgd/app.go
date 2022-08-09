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
	}
}

// Close closes the main application
func (m *App) Close() error {

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
func (m *App) Run(ctx context.Context) error {

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

	logService := log15.New("app", "mgd")

	m.HTTPServerAPI.Addr = config.GetConfig().APP.HTTP.Addr
	m.HTTPServerAPI.Domain = config.GetConfig().APP.HTTP.Domain
	m.HTTPServerAPI.LogService = logService.New("module", "http")

	m.HTTPServerAPI.ServiceHandler = handler.NewServiceHandler()
	m.HTTPServerAPI.ServiceHandler.AuthSearchService = authService
	m.HTTPServerAPI.ServiceHandler.ContractSearchService = postgresContractService
	m.HTTPServerAPI.ServiceHandler.UserSearchService = postgresUserService
	m.HTTPServerAPI.ServiceHandler.JWTService = jwtService
	m.HTTPServerAPI.ServiceHandler.Logger = m.HTTPServerAPI.LogService

	fuelTankService := mgvm.NewFuelTank()
	fuelTankService.LockService = redis.NewLockService(m.Redis, "fuel-tank-lock")
	fuelTankService.FuelTankService = redis.NewFuelTankService(m.Redis)

	fuelStationService := mgvm.NewFuelStation()
	fuelStationService.FuelTankService = fuelTankService
	fuelStationService.LogService = logService.New("module", "fuel-station")
	fuelStationService.FuelRefillAmount = entity.FuelRefillAmount
	fuelStationService.FuelRefillRate = entity.FuelRefillRate

	anchorageExecutor := executor.NewAnchorageContractExecutor()
	engineService := mgvm.NewEngine()
	engineService.Executors[entity.AnchorageVersion] = anchorageExecutor

	InitializerCPUsPool()

	cpusPoolService := mgvm.NewCPUsPool()

	m.VM.LogService = logService.New("module", "vm")
	m.VM.FuelTank = fuelTankService
	m.VM.FuelStation = fuelStationService
	m.VM.EngineService = engineService
	m.VM.CPUsPoolService = cpusPoolService

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

	logService.Info("Starting application",
		"addr", m.HTTPServerAPI.Addr,
		"domain", m.HTTPServerAPI.Domain,
		"tls", m.HTTPServerAPI.UseTLS(),
		"vm_fuel_tank_capacity", entity.FuelTankCapacity,
		"vm_fuel_refill_amount", entity.FuelRefillAmount,
		"vm_fuel_refill_rate", entity.FuelRefillRate,
		"vm_max_execution_time", entity.MaxExecutionTime,
	)

	return nil
}
