# 008 — Idle breath when the mic is quiet

**Commit:** 648ffc0 · **Severity:** opportunity (aliveness, taste-gated) · **Category:** Missed opportunities
**Files:** `internal/overlayglow/voice.go`

## Problem

Below voice 0.02 the ring is perfectly static apart from rotation — `voice.go:77-79` returns neutral multipliers:

```go
	if voice <= 0.02 {
		return VoiceBreath{1, 1, 1}
	}
```

While recording, "dead still" reads as "not listening". A barely-there breathing motion keeps the ring alive between phrases. This is a taste call against a crisp personality — the values below are deliberately at the edge of perception.

## Target values (copy exactly)

- Idle breath: width only, amplitude **0.025** (±2.5%), frequency **0.2Hz** (5s period → `sin(now * 1.2566)`).
- Hard-gated off under Reduce Motion (breathing is motion).

## Fix

Replace the early return in `BreathForVoice` (`voice.go:76-79`):

```go
func BreathForVoice(now float64, voice float64, reduceMotion bool) VoiceBreath {
	if voice <= 0.02 {
		if reduceMotion {
			return VoiceBreath{1, 1, 1}
		}
		return VoiceBreath{WidthMult: float32(1 + 0.025*math.Sin(now*1.2566))}
	}
	...
```

(`OpacityMult`/`BlurMult` stay 1 at idle — only the stroke breathes.)

## Test updates

`TestBreathForVoiceReduceMotionIsTimeInvariant` must still pass for the idle range — add cases:

```go
	// idle breathing is motion: frozen under Reduce Motion
	if BreathForVoice(1, 0.01, true) != BreathForVoice(1000, 0.01, true) {
		t.Fatalf("idle breath must be time-invariant under Reduce Motion")
	}
	// and alive without it
	if BreathForVoice(1, 0.01, false) == BreathForVoice(1000, 0.01, false) {
		t.Fatalf("idle breath should vary over time without Reduce Motion")
	}
```

Existing voice-active tests are unaffected (they use voice > 0.02).

## Scope boundaries

- Do NOT add idle opacity or blur movement; width only.
- Do NOT change the 0.02 threshold or the voice-active branch.
- If the result reads as flicker rather than breath at 1×, the amplitude is wrong — halve it to 0.0125 once, and if it still reads wrong, revert the plan entirely (this opportunity is optional by nature).

## Verification

1. `go test ./internal/overlayglow/...`.
2. Feel-check: record and stay silent 10s — the ring's stroke should swell/shrink about once every 5 seconds, noticeable only when you look for it. Toggle Reduce Motion: breathing stops dead (frozen, per plan 003's freeze). Speak: voice pulse takes over immediately.
