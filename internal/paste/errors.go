package paste

import "errors"

// ErrUnsupported is returned when synthetic paste is unavailable.
var ErrUnsupported = errors.New("paste: unsupported on this platform")

// ErrPostFailed is returned when synthetic key events could not be created.
var ErrPostFailed = errors.New("paste: could not post keyboard events")

// ErrPermissionDenied is returned when Accessibility trust is required but missing.
var ErrPermissionDenied = errors.New("paste: accessibility permission required for automated paste")

// ErrFocusChanged is returned when another application is focused at paste time.
var ErrFocusChanged = errors.New("paste: another application is focused")

// ErrTargetLost is returned when the original target application is no longer running.
var ErrTargetLost = errors.New("paste: original target application is not available")
