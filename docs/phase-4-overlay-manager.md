# Phase 4 — Implement the overlay manager

Part of the [implementation plan](../plan.md).

Build a separate overlay component that controls the glow window.

### Tasks

- Reuse the appropriate MyGo glow example implementation.
- Resolve the target display from the pointer location at recording start.
- Position and size the overlay using the display geometry and coordinate units expected by MyGo.
- Keep the overlay frameless, transparent and above ordinary application windows.
- Apply the native non-activation and click-through mechanisms established in Phase 1.
- Animate the glow in and out without blocking recording.
- Show a subtle transcribing state if processing takes longer than the initial transition.
- Handle display reconfiguration and overlay creation failures gracefully.

### Acceptance criteria

The original text field remains focused while the overlay is visible. Clicking through the overlay works where supported. The glow appears on the correct display in multi-monitor configurations and never leaves an unusable full-screen window behind after a failure.
