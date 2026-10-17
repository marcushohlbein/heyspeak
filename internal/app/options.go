package app

import "time"

// DefaultMaxRecording is the cap on a single hold-to-talk session.
const DefaultMaxRecording = 5 * time.Minute

// Options wires coordinator callbacks.
type Options struct {
	// OnStateChange is invoked after each state transition (may run on any goroutine).
	OnStateChange func(State)
	// OnError reports a user-visible failure; recognized text may be non-empty for recovery.
	OnError func(message string, recognized string)
}
