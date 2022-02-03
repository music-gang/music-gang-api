package mgvm

// State represents the state of the MusicGangVM.
type State int32

const (
	// StateRunning is the state of the MusicGangVM when it is running.
	StateInitializing State = iota
	StateRunning
	StatePaused
	StateClosed
)

// String returns a string representation of the State.
func (s State) String() string {
	switch s {
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}
