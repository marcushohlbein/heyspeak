package overlay

// Controller drives the recording and transcribing glow overlays.
type Controller interface {
	// PrepareSession locks the display under the pointer before capture starts.
	PrepareSession()
	ShowRecording()
	ShowTranscribing()
	// Pulse flares the glow once (0..1) — success acknowledgment.
	Pulse(strength float32)
	// DoublePulse signals a failed take: two soft blips.
	DoublePulse()
	Hide()
	Close()
}
