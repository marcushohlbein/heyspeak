package audio

import (
	"strings"
)

func mapMalgoError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "permission"),
		strings.Contains(msg, "not authorized"),
		strings.Contains(msg, "denied"):
		return ErrPermissionDenied
	case strings.Contains(msg, "no device"),
		strings.Contains(msg, "device not found"):
		return ErrNoInputDevice
	default:
		return err
	}
}
