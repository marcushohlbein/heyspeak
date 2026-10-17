//go:build !darwin

package paste

func paste() error { return ErrUnsupported }
