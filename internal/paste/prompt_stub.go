//go:build !darwin

package paste

// PromptAccessibility is unavailable off macOS.
func PromptAccessibility() bool { return false }
