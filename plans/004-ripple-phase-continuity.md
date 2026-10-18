# 004 — Ripple phase continuity across the transcribe switch

**Commit:** 648ffc0 · **Severity:** LOW · **Category:** Physicality
**Files:** `internal/overlay/manager.go`, `internal/overlayglow/view.go`

## Problem

When the transcribe dim flips, `manager.go:239` resets the glow clock: `m.glow.Start = now`. That clock is the only input to the 2.5Hz voice ripple — `view.go:46` (`now := time.Since(s.Start).Seconds()`) feeds `BreathForVoice`, whose beat is `0.55 + 0.45*math.Sin(now*15.8)` (`voice.go:83`). Resetting `Start` mid-oscillation jumps the sine's phase → a one-frame twitch in width/blur.

`Start` serves nothing else: `Stops.Gradient` ignores its argument (`stops.go:35`), and gradient rotation uses `GradientPhase`, not `Start`.

## Fix

1. Delete `m.glow.Start = now` from the transcribe flip block in `manager.go` (line ~239).
2. Verify `Start` is still set once per session at show time — `ShowRecording` sets it (`manager.go`, `m.glow.Start = time.Now()`); check `ShowTranscribing` (manager.go:157) does the same, and if it does not, add `m.glow.Start = time.Now()` there so the clock is never zero-time when a session starts as transcribe-only.
3. `State.Start` and its reset sites stay; only the mid-session reset goes.

## Scope boundaries

- Do NOT remove `State.Start` itself or its initialization in `NewManager`/`ShowRecording`.
- Do NOT touch `BreathForVoice`, the ripple frequency, or plan 002's dim fields.

## Verification

1. `go test ./internal/overlay/...`.
2. Feel-check: dictate a long take (~6s) with the mic close so voice stays active across the 300ms transcribe mark. The pulse must continue smoothly through the dim — no tick or twitch at the transition. Hard to catch at 1× — screen-record and scrub the switch frame-by-frame.
