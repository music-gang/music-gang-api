package mgvm

import (
	"context"
	"sync"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

var InitializerCPUsPool func(p *CPUsPool)

// Core represents a virtual machine core.
// It not performs any operations itself, but it provides a way blocking the caller until the core is available.
type Core struct{}

// NewCore returns a new core.
func NewCore() *Core {
	return &Core{}
}

// CorePool is a pool of cores.
// It is used to execute VM operations.
// Every time a core is requested, it is removed from the pool.
// When the core is released, it is added to the pool.
type CorePool chan *Core

// Release releases the given core back to the pool.
func (cPool CorePool) Release(c *Core) {
	cPool <- c
}

// FuelCorePool is a pool of cores for a specific fuel.
// Every keys is a fuel limit and caller, based on the fuel cost of his operation, will be assigned to the appropriate core pool.
// This meccanism is used to balance the load of the cores and permits to mantain a few cores for quick operations and high availability for those operations that require more resources.
type FuelCorePool struct {
	Pools    map[entity.Fuel]CorePool
	Fallback CorePool
}

// CPUsPool is a pool of cores for a specific operation + fuel limit.
// Every keys is a vm operation, the VM will assign the appropriate core pool to the caller based on the operation and the fuel limit.
type CPUsPool struct {
	OpsCorePools map[entity.VmOperation]FuelCorePool
}

// NewCPUsPool returns a new CPUsPool.
func NewCPUsPool() *CPUsPool {
	p := &CPUsPool{
		OpsCorePools: make(map[entity.VmOperation]FuelCorePool),
	}
	if InitializerCPUsPool != nil {
		InitializerCPUsPool(p)

		for _, fPool := range p.OpsCorePools {
			for _, cPool := range fPool.Pools {
				for i := 0; i < cap(cPool); i++ {
					cPool <- NewCore()
				}
			}
			if fPool.Fallback != nil {
				for i := 0; i < cap(fPool.Fallback); i++ {
					fPool.Fallback <- NewCore()
				}
			}
		}
	}
	return p
}

// AcquireCore executes the given vmCall.
func (p *CPUsPool) AcquireCore(ctx context.Context, call service.VmCallable) (release func(), err error) {

	cPool, err := p.getCorePool(call)
	if err != nil {
		return nil, err
	}

	var oneReleasePermit sync.Once

	select {
	case <-ctx.Done():
		return nil, apperr.Errorf(apperr.EMGVM_CORE_POOL_TIMEOUT, "core pool timeout")
	case core := <-cPool:
		return func() {
			oneReleasePermit.Do(func() {
				cPool.Release(core)
			})
		}, nil
	}
}

// getCorePool returns the fuel core pool for the given operation.
func (p *CPUsPool) getCorePool(call service.VmCallable) (CorePool, error) {

	if fPool, ok := p.OpsCorePools[call.Operation()]; ok {

		for maxFuel, cPool := range fPool.Pools {

			if call.Fuel() <= maxFuel {
				return cPool, nil
			}
		}

		if fPool.Fallback != nil {
			return fPool.Fallback, nil
		}

		return nil, apperr.Errorf(apperr.EMGVM_CORE_POOL_NOT_FOUND, "no core pool found for the given fuel %d", call.Fuel())
	}

	return nil, apperr.Errorf(apperr.EMGVM_CORE_POOL_NOT_FOUND, "Fuel core pool not found for operation %s", call.Operation())
}
