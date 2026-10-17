//go:build !darwin

package audio

func ensureMicPermission() error { return nil }

// MicrophoneStatus reports access as granted on platforms without a prompt.
func MicrophoneStatus() MicStatus { return MicAuthorized }

// RequestMicrophone is a no-op off macOS.
func RequestMicrophone() MicStatus { return MicAuthorized }
