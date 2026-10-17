package diagnostics

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/speek/speek/internal/audio"
	"github.com/speek/speek/internal/paste"
)

// Snapshot describes permission and runtime health for the settings panel.
type Snapshot struct {
	Platform             string
	AccessibilityTrusted bool
	InputDevices         int
	InputDeviceList      string
	PasteSupported       bool
}

// Collect gathers diagnostics without reading clipboard or audio content.
func Collect() Snapshot {
	s := Snapshot{
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		PasteSupported: runtime.GOOS == "darwin",
	}
	if runtime.GOOS == "darwin" {
		s.AccessibilityTrusted = paste.AccessibilityTrusted()
	}
	devs, err := audio.ListInputDevices()
	if err == nil {
		s.InputDevices = len(devs)
		var names []string
		for _, d := range devs {
			names = append(names, fmt.Sprintf("%s (%s)", d.Name, d.ID))
		}
		s.InputDeviceList = strings.Join(names, "\n")
	}
	return s
}
