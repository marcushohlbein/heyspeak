//go:build !darwin

package overlayglow

// PlaceOverDisplay is a no-op off macOS.
func PlaceOverDisplay(uintptr) {}

// PlaceOverFrame is a no-op off macOS.
func PlaceOverFrame(uintptr, float64, float64, float64, float64) {}
