package main

import (
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/mgvm"
)

// InitializerCPUsPool setups the CPUsPool for all possible vm operations.
func InitializerCPUsPool() {

	mgvm.InitializerCPUsPool = func(p *mgvm.CPUsPool) {

		p.OpsCorePools = map[entity.VmOperation]mgvm.FuelCorePool{
			entity.VmOperationExecuteContract: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.FuelInstantActionAmount:  make(mgvm.CorePool, 1),
					entity.FuelQuickActionAmount:    make(mgvm.CorePool, 3),
					entity.FuelFastestActionAmount:  make(mgvm.CorePool, 5),
					entity.FuelFastActionAmount:     make(mgvm.CorePool, 8),
					entity.FuelMidActionAmount:      make(mgvm.CorePool, 10),
					entity.FuelSlowActionAmount:     make(mgvm.CorePool, 12),
					entity.FuelLongActionAmount:     make(mgvm.CorePool, 15),
					entity.FuelExtremeActionAmount:  make(mgvm.CorePool, 17),
					entity.FuelAbsoluteActionAmount: make(mgvm.CorePool, 20),
				},
				Fallback: make(mgvm.CorePool, 10),
			},
			entity.VmOperationCreateContract: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationCreateContract): make(mgvm.CorePool, 10),
				},
			},
			entity.VmOperationUpdateContract: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationUpdateContract): make(mgvm.CorePool, 15),
				},
			},
			entity.VmOperationDeleteContract: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationDeleteContract): make(mgvm.CorePool, 5),
				},
			},
			entity.VmOperationMakeContractRevision: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationMakeContractRevision): make(mgvm.CorePool, 15),
				},
			},
			entity.VmOperationCreateUser: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationCreateUser): make(mgvm.CorePool, 5),
				},
			},
			entity.VmOperationUpdateUser: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationUpdateUser): make(mgvm.CorePool, 10),
				},
			},
			entity.VmOperationDeleteUser: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationDeleteUser): make(mgvm.CorePool, 5),
				},
			},
			entity.VmOperationAuthenticate: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationAuthenticate): make(mgvm.CorePool, 20),
				},
			},
			entity.VmOperationCreateAuth: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationCreateAuth): make(mgvm.CorePool, 5),
				},
			},
			entity.VmOperationDeleteAuth: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationDeleteAuth): make(mgvm.CorePool, 5),
				},
			},
			entity.VmOperationVmStats: {
				Pools: map[entity.Fuel]mgvm.CorePool{
					entity.VmOperationCost(entity.VmOperationVmStats): make(mgvm.CorePool, 5),
				},
			},
		}
	}
}
