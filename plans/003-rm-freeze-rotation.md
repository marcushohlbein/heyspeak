# 003 — Reduce Motion: freeze rotation in place

**Commit:** 648ffc0 · **Severity:** MEDIUM · **Category:** Accessibility
**Files:** `internal/overlayglow/view.go`

## Problem

`stepGradient` (`view.go:82-90`) returns constant 90° under Reduce Motion:

```go
func stepGradient(phase *float64, dt float64, reduceMotion bool) float64 {
	if reduceMotion {
		return 90
	}
	...
}
```

Toggling the OS Reduce Motion preference mid-recording teleports the ring to 90° in one frame. Reduced motion means *stop moving*, not *snap to a canonical position*.

## Fix

Preserve the accumulated phase and stop advancing:

```go
func stepGradient(phase *float64, dt float64, reduceMotion bool) float64 {
	if reduceMotion {
		return float32(90 + *phase)
	}
	if dt > 0 {
		*phase += dt * gradientDegPerSec
	}
	return float32(90 + *phase)
}
```

Then check for tests pinning the old behavior: `grep -rn "stepGradient" internal/overlayglow/*_test.go`. If a test asserts `90` under reduce motion, update it to assert the phase is preserved and NOT advanced across calls (two calls with different `dt` return the same angle under RM).

## Scope boundaries

- Do NOT touch rotation speed, the phase integration, or the voice pulse (its reduce-motion handling is separate and already correct).
- Do NOT animate the freeze — freezing in place is the point.

## Verification

1. `go test ./internal/overlayglow/...`.
2. Manual: start a dictation, let the ring rotate a few seconds, toggle Reduce Motion in System Settings mid-recording. The ring must hold its current angle — no jump. Toggle back off: rotation resumes from the held angle.
