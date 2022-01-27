package mgvm

// MusicGangVM is a virtual machine for the Mg language.
type MusicGangVM struct {
	FuelTank *FuelTank
}

// MusicGangVM creates a new MusicGangVM.
// It should be called only once.
func NewMusicGangVM() *MusicGangVM {
	return &MusicGangVM{}
}
