package overlay

// Controller drives the recording and transcribing glow overlays.
type Controller interface {
	// PrepareSession locks the display under the pointer before capture starts.
	PrepareSession()
	ShowRecording()
	ShowTranscribing()
	Hide()
	Close()
}
