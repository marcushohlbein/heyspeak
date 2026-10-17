package permissions

import (
	"runtime"

	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/paste"
)

// NeedsAttention reports whether a permission is still missing on this platform.
func NeedsAttention() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	return !audio.MicrophoneStatus().Granted() || !paste.AccessibilityTrusted()
}
