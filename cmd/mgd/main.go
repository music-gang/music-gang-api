package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/config"
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

	app.Commit = os.Getenv("HEROKU_SLUG_COMMIT")
	app.ReleaseVersion = os.Getenv("HEROKU_RELEASE_VERSION")
	app.ReleaseCreatedAt = os.Getenv("HEROKU_RELEASE_CREATED_AT")

	// Create a context that is cancelled when the program is terminated
	ctx, cancel := context.WithCancel(context.Background())
	ctx = app.NewContextWithTags(ctx, []string{app.ContextTagCLI})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	m := NewApp()

	if err := m.Run(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()

	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
