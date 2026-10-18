# 006 — Success pulse on completed dictation

**Commit:** 648ffc0 · **Severity:** opportunity (delight) · **Category:** Missed opportunities
**Files:** `internal/overlay/manager.go`, `internal/app/coordinator.go`

## Problem

A successful dictation — the app's core "it worked" moment — ends with the glow fading out exactly like every other exit. No confirmation, no closure. The overlay already has a full pulse pipeline (mic peak → envelope → breath); a success flare needs one public method and one call.

Deliberate scope cut: the pulse is brightness-only. The breath multipliers scale the whole gradient, so a "warm-colored" pulse would require palette surgery — not worth it. The flare reads because it coincides with the paste.

## Steps

1. `internal/overlay/manager.go` — add a public pulse that reuses the existing peak path:

```go
// Pulse flares the glow once (0..1). Safe from any goroutine; the peak
// handoff and envelope do the rest.
func (m *Manager) Pulse(strength float32) {
	mygo.RunOnMain(func() {
		m.mu.Lock()
		visible := m.mode == modeRecording || m.mode == modeTranscribing
		m.mu.Unlock()
		if !visible {
			return
		}
		m.OnMicPeak(int(strength * 3200))
	})
}
```

(`OnMicPeak` maps 3200 → full scale via `PeakToPulseLevel`; the envelope's 28ms attack / 110ms release shapes the flare.)

2. `internal/app/coordinator.go` — find the success path (after transcription returns text and before the overlay is hidden / `returnIdle`-equivalent runs, around line 293+). Call the pulse immediately before the hide so the flare peaks during the first ~100ms of the 520ms fade-out, while opacity is still near 1:

```go
	c.svc.Overlay.Pulse(0.9)
```

Check whether `svc.Overlay` is typed as `*overlay.Manager` or an interface in this file; if an interface, add `Pulse(float32)` to it (the concrete manager already satisfies it after step 1).

3. Recon obligation: read the success branch end-to-end and confirm where hide is triggered. If hide is triggered *before* this point in the flow, move the pulse call to just before that trigger instead.

## Scope boundaries

- Do NOT touch the palette, add color, or change fade durations.
- Do NOT pulse on error/no-speech paths (plan 007 owns those) or on `context.Canceled`.
- Keep the pulse subtle: 0.9 max, one flare, no repeat.

## Verification

1. `go test ./internal/...`.
2. Feel-check: dictate 10 short phrases. Each successful paste should be followed by a single quick swell-and-settle of the ring as it fades — readable as acknowledgment, not confetti. Cancel mid-dictation: no pulse.
