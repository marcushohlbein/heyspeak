package holdshortcut

import (
	"runtime"
	"strings"
)

// FormatAccelerator builds an accelerator string from modifier flags and a key name.
func FormatAccelerator(ctrl, shift, alt, meta bool, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	var parts []string
	if ctrl {
		parts = append(parts, "Ctrl")
	}
	if shift {
		parts = append(parts, "Shift")
	}
	if alt {
		parts = append(parts, "Alt")
	}
	if meta {
		if runtime.GOOS == "darwin" {
			parts = append(parts, "Cmd")
		} else {
			parts = append(parts, "Super")
		}
	}
	parts = append(parts, key)
	return strings.Join(parts, "+")
}
