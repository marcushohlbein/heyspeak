//go:build !darwin

package paste

func frontmostPID() int { return -1 }

func ownPID() int { return -1 }

func activatePID(_ int) bool { return false }
