package models

import "errors"

var (
	ErrUnknownModel     = errors.New("models: unknown model id")
	ErrAlreadyInstalled = errors.New("models: already installed")
	ErrDownloadInFlight = errors.New("models: download already in progress")
	ErrNotInstalled     = errors.New("models: not installed")
	ErrChecksumMismatch = errors.New("models: checksum verification failed")
	// ErrNoSelection is returned when multiple models are installed but none is selected.
	ErrNoSelection = errors.New("models: no model selected")
)
