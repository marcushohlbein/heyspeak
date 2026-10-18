# 002 — Transcribe dim: ease instead of one-frame pop

**Commit:** 648ffc0 · **Severity:** MEDIUM · **Category:** Interruptibility
**Files:** `internal/overlay/manager.go`, `internal/overlay/animate.go`, `internal/overlayglow/view.go`
**Depends on:** plan 001 (provides `easeOutShow`)

## Problem

On every longer dictation, when processing exceeds 300ms the glow brightness pops to 45% in a single frame — no transition. The flip happens in `manager.go:236-240` (`m.glow.Visual = overlayglow.VisualTranscribing`) and the multiplier is applied unconditionally in `view.go:63-66`:

```go
			strength := s.Opacity
			if s.Visual == VisualTranscribing {
				strength *= 0.45
			}
```

A jarring state change must be transitioned (the opacity fade already retargets exactly this way).

## Target values (copy exactly)

- Dim target: **0.45** (unchanged), full: **1**.
- Dim transition: **200ms**, easing `easeOutShow` (from plan 001).

## Steps

1. `internal/overlayglow/view.go` — add a field to `State`:

```go
	TranscribeDim float64 // 1 full brightness, eased toward 0.45 while transcribing
```

Replace the hard multiplier in `Draw`:

```go
			strength := s.Opacity * s.TranscribeDim
```

(Draw must not return early when `TranscribeDim` is animating — it doesn't: visibility/fade gating lives in the manager.)

2. `internal/overlay/animate.go` — add the constant:

```go
	transcribeDimDuration = 200 * time.Millisecond
```

3. `internal/overlay/manager.go` — add fields mirroring the opacity animation (next to `opacityFrom`/`animStart`):

```go
	dimFrom   float64
	dimStart  time.Time
	dimTarget float64
```

4. In `ShowRecording` (the block that resets `m.glow`), reset the dim: `m.dimFrom, m.dimStart, m.dimTarget = 1, time.Time{}, 1` and `m.glow.TranscribeDim = 1`. Do the same in `finishHide` and `teardown` (set `m.glow.TranscribeDim = 1`).
5. In `view()`, right after the existing transcribe flip block (`manager.go:236-240`), start the dim animation when it flips:

```go
	if m.showTranscribe {
		if m.dimTarget != 0.45 {
			m.dimFrom = m.glow.TranscribeDim
			m.dimStart = now
			m.dimTarget = 0.45
		}
	} else if m.dimTarget != 1 {
		m.dimFrom = m.glow.TranscribeDim
		m.dimStart = now
		m.dimTarget = 1
	}
	m.glow.TranscribeDim = advanceOpacity(now, m.dimStart, m.dimFrom, m.dimTarget, transcribeDimDuration, easeOutShow)
```

Keep the existing `m.glow.Start` reset line inside the flip block only if plan 004 has not landed; if it has, respect its removal.

## Scope boundaries

- Do NOT change the 300ms `TranscribeVisualDelay`, the 0.45 target, or the rotation/voice logic.
- Do NOT add `TranscribeDim` to the volatile `frameState` copy-back — the manager computes it every frame; Draw never mutates it.

## Verification

1. `go test ./internal/overlay/... ./internal/overlayglow/...`.
2. Feel-check: dictate continuously past 5 seconds. At the 300ms-processing mark the glow must glide to half brightness over ~200ms — no visible step. Screen-record at 1/4 speed and scrub the transition frame-by-frame: brightness should change monotonically, never jump.
