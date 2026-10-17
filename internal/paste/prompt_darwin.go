//go:build darwin

package paste

/*
#cgo darwin LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "prompt.h"
*/
import "C"

// PromptAccessibility asks macOS to show the Accessibility permission dialog.
func PromptAccessibility() bool {
	return C.speek_prompt_accessibility() != 0
}
