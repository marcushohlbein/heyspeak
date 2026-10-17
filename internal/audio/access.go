package audio

// MicStatus is the operating system's microphone authorization.
type MicStatus int

const (
	MicNotDetermined MicStatus = 0
	MicRestricted    MicStatus = 1
	MicDenied        MicStatus = 2
	MicAuthorized    MicStatus = 3
)

// Granted reports whether capture is allowed.
func (s MicStatus) Granted() bool { return s == MicAuthorized }

// Label is a short status for the permissions window.
func (s MicStatus) Label() string {
	switch s {
	case MicAuthorized:
		return "Granted"
	case MicNotDetermined:
		return "Not asked yet"
	case MicDenied:
		return "Denied"
	case MicRestricted:
		return "Restricted"
	default:
		return "Unknown"
	}
}
