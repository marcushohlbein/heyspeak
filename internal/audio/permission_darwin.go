//go:build darwin

package audio

/*
#cgo darwin LDFLAGS: -framework AVFoundation -framework Foundation
#include "permission.h"
*/
import "C"

// MicrophoneStatus is the current macOS microphone authorization.
func MicrophoneStatus() MicStatus {
	return MicStatus(C.speek_mic_status())
}

// RequestMicrophone asks macOS for microphone access and returns the result.
func RequestMicrophone() MicStatus {
	C.speek_mic_request()
	return MicrophoneStatus()
}

func ensureMicPermission() error {
	status := MicrophoneStatus()
	if status == MicNotDetermined {
		status = RequestMicrophone()
	}
	if status.Granted() {
		return nil
	}
	return ErrPermissionDenied
}
