package mgvm_test

import (
	"context"
	"testing"
	"time"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
	"github.com/music-gang/music-gang-api/mgvm"
)

func TestCPUsPool_AcquireCore(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		ctx := context.Background()

		mgvm.InitializerCPUsPool = func(p *mgvm.CPUsPool) {
			p.OpsCorePools[entity.VmOperationGeneric] = mgvm.FuelCorePool{
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.FuelExtremeActionAmount: make(mgvm.CorePool, 5),
				},
				Fallback: make(mgvm.CorePool, 1),
			}
		}

		pool := mgvm.NewCPUsPool()

		fuel := entity.FuelExtremeActionAmount

		r1, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}
		r2, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}
		r3, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}
		r4, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}
		r5, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}

		ctxCancel, cancel := context.WithCancel(ctx)

		go func() {
			time.Sleep(time.Second)
			cancel()
		}()

		_, err = pool.AcquireCore(ctxCancel, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err == nil {
			t.Fatal("expected error")
		}

		r1()

		r7, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err != nil {
			t.Fatal(err)
		}

		r7()
		r2()
		r3()
		r4()
		r5()

		// try to acquire the fallback core

		fuelToFallbackCore := entity.FuelAbsoluteActionAmount

		r8, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuelToFallbackCore,
		}))
		if err != nil {
			t.Fatal(err)
		}
		r8()
	})

	t.Run("ErrNoFuelCorePool", func(t *testing.T) {

		ctx := context.Background()

		mgvm.InitializerCPUsPool = func(p *mgvm.CPUsPool) {
			p.OpsCorePools[entity.VmOperationGeneric] = mgvm.FuelCorePool{
				Pools: map[entity.Fuel]mgvm.CorePool{},
			}
		}

		pool := mgvm.NewCPUsPool()

		fuel := entity.FuelExtremeActionAmount

		_, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationAuthenticate,
			CustomMaxFuel: &fuel,
		}))
		if err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM_CORE_POOL_NOT_FOUND {
			t.Fatalf("expected error code, got %s, want %s", errCode, apperr.EMGVM_CORE_POOL_NOT_FOUND)
		}
	})

	t.Run("ErrNoCorePool", func(t *testing.T) {

		ctx := context.Background()

		mgvm.InitializerCPUsPool = func(p *mgvm.CPUsPool) {
			p.OpsCorePools[entity.VmOperationGeneric] = mgvm.FuelCorePool{
				Pools: map[entity.Fuel]mgvm.CorePool{},
			}
		}

		pool := mgvm.NewCPUsPool()

		fuel := entity.FuelExtremeActionAmount

		_, err := pool.AcquireCore(ctx, service.NewVmCallWithConfig(service.VmCallOpt{
			VmOperation:   entity.VmOperationGeneric,
			CustomMaxFuel: &fuel,
		}))
		if err == nil {
			t.Fatal("expected error")
		} else if errCode := apperr.ErrorCode(err); errCode != apperr.EMGVM_CORE_POOL_NOT_FOUND {
			t.Fatalf("expected error code, got %s, want %s", errCode, apperr.EMGVM_CORE_POOL_NOT_FOUND)
		}
	})
}
