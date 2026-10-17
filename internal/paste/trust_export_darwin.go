//go:build darwin

package paste

// AccessibilityTrusted reports whether the process may post synthetic input.
func AccessibilityTrusted() bool {
	return accessibilityTrusted()
}
