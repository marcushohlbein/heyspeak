# Phase 1 — Technical feasibility spikes

Part of the [implementation plan](../plan.md).

Implement small, isolated prototypes that answer the following questions.

### 1A. Global push-to-talk

Verify that the application can reliably detect both pressing and releasing a configurable global shortcut while another application has focus.

MyGo's documented `GlobalShortcut.Register` callback handles shortcut activation; do not assume it provides a paired key-up event.

Investigate a native platform-specific keyboard adapter if necessary. Keep its public interface independent from the underlying operating system.

Do not use a timer to guess when a shortcut was released. Do not silently replace hold-to-talk with toggle-to-record.

### 1B. Display glow overlay

Create a prototype that:

- Detects the display nearest the pointer.
- Covers the correct display bounds.
- Renders a subtle glow around the display edges.
- Handles display changes, mixed DPI scaling and negative display coordinates.
- Does not activate or focus the overlay.
- Allows mouse interactions to pass through it.

Investigate the native window handle and platform APIs if MyGo does not expose the required behavior directly.

### 1C. Microphone and paste integration

Verify that the application can:

- Capture microphone audio while a global shortcut is held.
- Stop recording and retrieve the captured audio.
- Put test text on the clipboard.
- Paste into another application without transferring focus to the transcription application.

Investigate operating-system permissions and limitations, especially for macOS accessibility features and Linux Wayland.

### Acceptance criteria

Each prototype runs independently on the initial target platform. Document any permission requirements or platform restrictions and use the results to finalize the adapters.

If a required platform capability cannot be implemented reliably, document the blocker and the smallest viable alternative instead of hiding the limitation.

---

## Spike report (completed)

Date: 2026-10-17. Platform: **macOS arm64**. Module: `github.com/speek/speek` (`go 1.27.1`, `mygo v0.5.5`).

### Artifacts

| Spike | Command | Package |
|-------|---------|---------|
| **1A** PTT press/release | `go run ./cmd/spike-ptt` | `internal/holdshortcut` — CGEvent tap (not MyGo `GlobalShortcut`) |
| **1B** Display glow overlay | `go run ./cmd/spike-overlay` | `internal/overlayglow` + MyGo window APIs |
| **1C** Mic + clipboard + paste | `go run ./cmd/spike-capture-paste` | `internal/audio` (malgo), `internal/paste` |

Default hold chord: **`Ctrl+Shift+Space`** (`SPEEK_PTT_SHORTCUT`). Overlay toggle: **`Alt+Shift+G`**. Paste is on by default; set `SPEEK_SPIKE_PASTE=0` to skip Cmd+V.

### Verified on this Mac

All three spikes were run manually and behave as intended: hold and release, full-display glow, and capture plus paste.

### 1A — Global push-to-talk

MyGo `GlobalShortcut.Register` is press-only. `holdshortcut.Monitor` uses a CGEvent tap for key down, key up, and modifier release. Chord events are swallowed so Ctrl+Shift+Space does not repeat as NUL (`^@`) into the focused app. **Accessibility** is required.

### 1B — Display glow overlay

A transparent, click-through, non-activating MyGo panel. Edge glow follows `examples/chat-glow`. The panel is placed on the screen under the pointer with `NSScreen.frame`, above the menu bar and Dock. `NSPanel` would otherwise shrink the frame to the visible area and clip the bottom edge. MyGo fullscreen mode is not used: it opens another Space, and panels cannot enter it.

### 1C — Microphone and paste

`malgo` captures mono S16 at 16 kHz until the chord is released. `mygo.Clipboard.WriteText` stores the text. `internal/paste` posts Cmd+V on the session event tap. Same Accessibility trust as 1A. Microphone permission is requested on first capture.

### Carry into Phase 2

- Keep the event tap on its own `LockOSThread` goroutine.
- Reuse `PlaceOverDisplay` for the recording overlay. Do not size it with MyGo's `(0,0)` centering or `SetBounds` alone.
- Linux and Windows adapters are still stubs.
- Accessibility must be granted to the app that actually runs the binary (Terminal or Cursor).
