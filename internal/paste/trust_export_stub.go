//go:build !darwin

package paste

// AccessibilityTrusted is false on unsupported platforms.
func AccessibilityTrusted() bool {
	return false
}
