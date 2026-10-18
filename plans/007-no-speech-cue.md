# 007 — No-speech cue: double soft pulse

**Commit:** 648ffc0 · **Severity:** opportunity (clarity) · **Category:** Missed opportunities
**Files:** `internal/overlay/manager.go`, `internal/app/coordinator.go`
**Depends on:** plan 006 (introduces `Manager.Pulse`; this plan builds on it)

## Problem

`coordinator.go:287-290` routes `ErrNoSpeech` / `ErrEmptyAudio` to a text failure ("No speech detected.") while the overlay fades out identically to success. The user must read a banner to learn the take failed; the peripheral channel says nothing. A distinct overlay cue — two quick, softer pulses (vs. success's single strong flare) — communicates failure pre-attentively.

## Steps

1. `internal/overlay/manager.go` — add a double pulse below `Pulse`:

```go
// DoublePulse signals a failed take: two soft blips, distinct from the
// single success flare.
func (m *Manager) DoublePulse() {
	m.Pulse(0.45)
	time.AfterFunc(180*time.Millisecond, func() {
		m.Pulse(0.45)
	})
}
```

2. `internal/app/coordinator.go` — in the `ErrNoSpeech`/`ErrEmptyAudio` branch (lines 287-290) and the empty-text branch below it, call `c.svc.Overlay.DoublePulse()` before `c.userFailure(...)`.
3. Recon obligation (do this first): read `userFailure` and the overlay hide path it triggers. The cue only reads if the glow is still on screen for ≥300ms. If the failure path hides the overlay instantly (opacity slam, not the 520ms fade), the pulses must fire *before* the hide begins — with the 180ms gap they need roughly 450ms of visible glow. If the hide cannot be delayed, stop and report: the plan is wrong for this code path, do not force it.

## Scope boundaries

- Do NOT add sound, color, text, or a third pulse.
- Do NOT touch the success pulse from plan 006 (single, 0.9, different rhythm — that contrast is the signal).
- `context.Canceled` (user aborted) stays cue-less by design.

## Verification

1. `go test ./internal/...`.
2. Feel-check: dictate silence twice, then a real phrase. The two blips must read as a distinct "nothing there" rhythm vs. the single success flare — testable with eyes on the screen edge, not the banner. If you cannot tell them apart without reading the banner, increase the second blip to 0.6 and re-test before changing anything else.
