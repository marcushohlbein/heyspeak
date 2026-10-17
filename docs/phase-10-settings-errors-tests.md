# Phase 10 — Settings, errors and tests

Part of the [implementation plan](../plan.md).

### Settings

Implement a compact settings view with:

- Push-to-talk shortcut.
- Selected transcription model.
- Model download and removal controls.
- Language selection.
- Microphone selection, if supported.
- Model installation status.
- Permission and runtime diagnostics.
- Clipboard restoration preference, on by default.

Validate shortcut conflicts and failed registrations. Do not overwrite settings with invalid values.

### Testing

Add unit tests for the state machine, configuration, model manifests, download verification and error handling.

Add integration tests where practical using fake audio, transcription and paste implementations.

Manually test the native behaviors that cannot reliably be exercised in unit tests:

- Global key-down and key-up while another app is focused.
- A key-up that arrives before recording has started.
- Overlay visibility and click-through.
- Multiple monitors and display scaling.
- Actual microphone capture.
- The recording duration limit while the shortcut is still held.
- Local inference.
- A silence-only recording, including one where the engine returns text.
- Paste into multiple applications.
- Switching to another application before paste.
- Clipboard restoration after a successful paste, and recovery when paste fails.
- Permission denial.
- Recording cancellation from the tray during recording and during transcription.
- Shutdown during recording or transcription.
- Download cancellation and corrupted model files.
- Repeated dictation sessions.

### Acceptance criteria

All automated tests pass. Manual platform checks are documented. Logs are useful without containing dictated content.
