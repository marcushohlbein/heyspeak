# Animation Plans — overlay & gradient motion

Audited at commit `648ffc0` (branch `feat/voice-reactive-overlay-glow`), focus: overlay + gradient animation surface.

## Execution order

| # | Plan | Status | Depends on | Note |
|---|------|--------|------------|------|
| 001 | [Entrance ease-out + 240ms](001-entrance-ease-out.md) | DONE (code) | — | Highest leverage; introduces the `easeOutShow`/`cubicBezier` helpers |
| 003 | [RM freeze rotation](003-rm-freeze-rotation.md) | DONE (code) | — | One line, independent |
| 004 | [Ripple phase continuity](004-ripple-phase-continuity.md) | DONE (code) | — | One line; coordinate with 002 on the same code block |
| 002 | [Transcribe dim ease](002-transcribe-dim-ease.md) | DONE (code) | 001 | Uses `easeOutShow`; touches the same flip block as 004 |
| 008 | [Idle breath](008-idle-breath.md) | DONE (code) | 003 (for its RM check) | Taste-gated; revert if it reads as flicker |
| 006 | [Success pulse](006-success-pulse.md) | DONE (code) | — | Reuses peak pipeline; touch coordinator |
| 007 | [No-speech double pulse](007-no-speech-cue.md) | DONE (code) | 006 | Guard passed: `userFailure` hides via the 520ms fade, so the blips read |
| 005 | [Blur headroom feel-check](005-blur-headroom-feel-check.md) | **NEEDS HUMAN** | 008 (so blur load is final) | Requires real hardware + loud audio; measurement first |

## Status legend

- **DONE (code)** — implemented and covered by `go test`; the plan's human **feel-check is still outstanding** (watch the motion at 1× and in slow motion before calling it done).
- **NEEDS HUMAN** — cannot be automated; see the plan's steps.

## Notes

- 001 before everything that eases: it lands the shared bezier helper.
- 002 and 004 edit the same transcribe-flip block — do them back to back.
- 005 last: the idle breath (008) adds constant motion, so measure blur cost only after the motion surface is final.
- Every plan ends with a feel-check; none are done until someone has watched the motion at 1× and in slow motion.
- Re-run `improve-animations reconcile` after executing to refresh this table.
