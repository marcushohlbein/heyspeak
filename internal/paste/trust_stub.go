//go:build !darwin

package paste

func accessibilityTrusted() bool { return false }
