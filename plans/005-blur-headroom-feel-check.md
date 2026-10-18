# 005 — Blur headroom feel-check (conditional fix)

**Commit:** 648ffc0 · **Severity:** LOW · **Category:** Performance
**Files:** none by default; `internal/overlayglow/voice.go` only if the check fails

## Problem

At full voice the largest layer's blur reaches ~24px: `15 × 0.95 × BlurMult`, where `BlurMult = 1 + wave*0.65` peaks at 1.65 (`voice.go:89`, applied at `view.go:75`). That is over the ~20px comfort budget for transition-time blur. Mitigating context: this is a native GPU-composited painter, and the glow is skipped entirely when `p.Accelerated()` is false (`view.go:60-62`), so the CPU path never pays it.

Whether this costs frames on real hardware cannot be judged from code. This plan is a measurement first, fix second.

## Steps

1. Run the app on the target hardware. Prefer an Intel Mac if available (integrated GPU is the risk case); on Apple Silicon this check is a formality.
2. Generate sustained loud input during a recording (play music near the mic, or speak loudly) so `wave` sits near 1 and blur rides at its ~24px peak for several seconds.
3. Watch for dropped frames / stutter in the glow rotation. The rotation (20°/s) makes judder easy to see against the screen edge. If available, capture with a screen recorder at 60fps and scrub for duplicated frames.
4. **Only if stutter is observed:** in `voice.go:89`, cap the blur contribution — change `BlurMult: float32(1 + wave*0.65)` to `float32(1 + wave*0.4)` (peak ≈ 1.4, blur ≈ 20px). Leave `OpacityMult` and `WidthMult` untouched so the pulse remains clearly visible.

## Scope boundaries

- Do NOT change blur values without observed stutter.
- Do NOT touch the GPU gate, layer widths (`glowLayers`), or the base `0.95` factor.
- If stutter appears only on a non-GPU path, stop: that path draws no glow by design; report instead of fixing.

## Verification

- With no stutter observed: no code change; record the result in this file (one line under **Result**).
- With the cap applied: `go test ./internal/overlayglow/...` and repeat step 2-3 — stutter gone, pulse still visible.

**Result:** _(fill in after the check)_
