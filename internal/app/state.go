package app

// State is the dictation coordinator state.
type State int

const (
	StateIdle State = iota
	StateStarting
	StateRecording
	StateTranscribing
	StatePasting
	StateError
)

var stateNames = [...]string{"Idle", "Starting", "Recording", "Transcribing", "Pasting", "Error"}

func (s State) String() string {
	if s < 0 || int(s) >= len(stateNames) {
		return "Unknown"
	}
	return stateNames[s]
}
