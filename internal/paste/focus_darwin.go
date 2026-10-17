//go:build darwin

package paste

/*
#cgo darwin LDFLAGS: -framework AppKit
#include "focus.h"
*/
import "C"

func frontmostPID() int {
	return int(C.speek_frontmost_pid())
}

func ownPID() int {
	return int(C.speek_current_pid())
}

func activatePID(pid int) bool {
	return C.speek_activate_pid(C.int(pid)) != 0
}
