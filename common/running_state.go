package common

import "sync/atomic"

// RunningState is a struct that can be used as a safe running state flag for a goroutine.
type RunningState struct {
	running int32
}

// IsRunning returns true if the SafeRunningState is running
func (s *RunningState) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// SetRunningState sets the running state of the FuelStation.
func (s *RunningState) SetRunningState(val int32) {
	atomic.StoreInt32(&s.running, val)
}
