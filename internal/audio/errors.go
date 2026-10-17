package audio

import "errors"

// ErrNotRecording is returned when Stop is called without an active session.
var ErrNotRecording = errors.New("audio: not recording")

// ErrAlreadyRecording is returned when Start is called twice.
var ErrAlreadyRecording = errors.New("audio: capture already running")

// ErrNoInputDevice is returned when no capture devices are available.
var ErrNoInputDevice = errors.New("audio: no input device available")

// ErrDeviceNotFound is returned when the configured device ID is missing.
var ErrDeviceNotFound = errors.New("audio: input device not found")

// ErrPermissionDenied is returned when the OS blocks microphone access.
var ErrPermissionDenied = errors.New("audio: microphone permission denied")

// ErrSilentInput is returned when the device delivers a buffer of pure silence.
var ErrSilentInput = errors.New("audio: microphone produced silence")
