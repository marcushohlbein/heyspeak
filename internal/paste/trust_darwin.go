//go:build darwin

package paste

import (
	"sync"

	"github.com/ebitengine/purego"
)

var (
	trustOnce sync.Once
	trustFn   func() bool
	trustErr  error
)

func accessibilityTrusted() bool {
	trustOnce.Do(func() {
		ax, err := purego.Dlopen("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices", purego.RTLD_GLOBAL)
		if err != nil {
			trustErr = err
			return
		}
		purego.RegisterLibFunc(&trustFn, ax, "AXIsProcessTrusted")
	})
	if trustErr != nil || trustFn == nil {
		return false
	}
	return trustFn()
}
