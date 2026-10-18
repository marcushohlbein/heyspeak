# 001 — Entrance lag: ease-out show curve + 240ms fade-in

**Commit:** 648ffc0 · **Severity:** HIGH · **Category:** Easing & duration
**Files:** `internal/overlay/animate.go`, `internal/overlay/manager.go`, `internal/overlay/animate_test.go`

## Problem

The edge glow is invisible for the first ~200ms of every dictation — the app's most frequent animation. Three compounding causes:

1. `animate.go:9` — `fadeInDuration = 420 * time.Millisecond`.
2. `animate.go:42,46-50` — `advanceOpacity` applies cubic `easeInOut` to **both** directions. Ease-in-out starts at its slowest, delaying exactly the moment the user is watching ("did my keypress register?"). Entrances must ease-**out** (fast start, gentle settle).
3. `internal/overlayglow/view.go:97-99` — `appearBloom = opacity²` squares the early opacity: at 100ms of a 420ms ease-in-out, opacity ≈ 0.055 → visible strength ≈ 0.3%.

## Target values (copy exactly)

- Fade-in duration: **240ms** (was 420ms). Keep `fadeOutDuration = 520ms`, keep RM values `fadeInReduced = 120ms`, `fadeOutReduced = 150ms`.
- Show easing: **cubic-bezier(0.23, 1, 0.32, 1)** — implemented as a Go function via the solver below. Hide easing: keep the existing `easeInOut`.

## Current code

`animate.go:32-50`:

```go
// advanceOpacity moves current toward target over the given duration.
func advanceOpacity(now, animStart time.Time, from, target float64, duration time.Duration) float64 {
	if duration <= 0 {
		return target
	}
	elapsed := now.Sub(animStart)
	if elapsed >= duration {
		return target
	}
	t := float64(elapsed) / float64(duration)
	t = easeInOut(t)
	return from + (target-from)*t
}
```

Call site `manager.go:233`:

```go
m.opacity = advanceOpacity(now, m.animStart, m.opacityFrom, m.opacityTarget, fadeDur)
```

## Steps

1. In `animate.go`, add the exact show curve (paste verbatim):

```go
// easeOutShow is cubic-bezier(0.23, 1, 0.32, 1): strong ease-out for UI
// entrances — fast start, gentle settle.
func easeOutShow(t float64) float64 { return cubicBezier(0.23, 1, 0.32, 1, t) }

// cubicBezier evaluates a CSS-style cubic-bezier(x1,y1,x2,y2) at x by
// bisecting the parametric curve. 24 iterations give <1e-6 precision.
func cubicBezier(x1, y1, x2, y2, x float64) float64 {
	lo, hi, t := 0.0, 1.0, x
	for i := 0; i < 24; i++ {
		v := bez(t, x1, x2) - x
		if math.Abs(v) < 1e-6 {
			break
		}
		if v < 0 {
			lo = t
		} else {
			hi = t
		}
		t = (lo + hi) / 2
	}
	return bez(t, y1, y2)
}

func bez(t, p1, p2 float64) float64 {
	u := 1 - t
	return 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t
}
```

2. Change `fadeInDuration` to `240 * time.Millisecond`.
3. Add an easing parameter to `advanceOpacity`:

```go
func advanceOpacity(now, animStart time.Time, from, target float64, duration time.Duration, ease func(float64) float64) float64 {
	...
	t = ease(t)
	...
}
```

4. In `manager.go:233`, pick the curve by direction:

```go
	ease := easeInOut
	if m.opacityTarget > m.opacityFrom {
		ease = easeOutShow
	}
	m.opacity = advanceOpacity(now, m.animStart, m.opacityFrom, m.opacityTarget, fadeDur, ease)
```

5. Update `animate_test.go`: `TestAdvanceOpacityReachesTarget` gains the `easeOutShow` argument (unchanged assertion). Replace `TestAdvanceOpacityMidpoint`'s loose bound with a fast-start assertion — the new curve reaches ~0.77 opacity at 25% elapsed:

```go
func TestAdvanceOpacityFastStart(t *testing.T) {
	start := time.Now()
	got := advanceOpacity(start.Add(fadeInDuration/4), start, 0, 1, fadeInDuration, easeOutShow)
	if got < 0.7 {
		t.Fatalf("expected fast entrance (~0.77 at 25%%), got %v", got)
	}
}
```

## Scope boundaries

- Do NOT change `fadeOutDuration`, the RM durations, `appearBloom`, or anything in `overlayglow`.
- Do NOT touch the transcribe-dim logic (plan 002) or add tokens/config.

## Verification

1. `go test ./internal/overlay/...` — all pass, including the new fast-start test.
2. Feel-check (required): run the app, dictate ~10 times. The ring must be clearly visible within ~50ms of the hotkey — no slow fade-up. Then reduce: screen-record at 1/4 speed; the entrance should read as an immediate appear-with-settle, not a swell.
