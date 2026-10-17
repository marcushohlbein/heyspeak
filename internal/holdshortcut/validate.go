package holdshortcut

import "fmt"

// ValidateAccelerator checks whether an accelerator string can be registered.
func ValidateAccelerator(accelerator string) error {
	if accelerator == "" {
		return fmt.Errorf("shortcut is empty")
	}
	if _, err := parseChord(accelerator); err != nil {
		return fmt.Errorf("invalid shortcut: %w", err)
	}
	return nil
}
