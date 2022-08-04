package handler

import (
	"context"

	"github.com/music-gang/music-gang-api/app/entity"
)

// StatsVM returns the VM stats.
func (s *ServiceHandler) StatsVM(ctx context.Context) (*entity.FuelStat, error) {

	stats, err := s.VmCallableService.Stats(ctx)
	if err != nil {
		s.LogService.ReportError(ctx, err)
		return nil, err
	}

	return stats, nil
}
