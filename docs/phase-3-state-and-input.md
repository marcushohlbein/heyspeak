# Phase 3 — Implement application state and input handling

Part of the [implementation plan](../plan.md).

Implement an explicit state machine:

```text
Idle
  -> Starting

Starting
  -> Recording     when capture is active and the shortcut is still held
  -> Transcribing  when key-up arrived before capture became active
  -> Idle          on cancel, or when capture fails before any audio is available

Recording
  -> Idle          on cancel
  -> Transcribing  on key-up, or when the maximum recording duration is reached

Transcribing
  -> Pasting
  -> Idle          on cancel
  -> Error         on failure

Pasting
  -> Idle
  -> Error         on failure

Error
  -> Idle          after the failure is shown, keeping any recognized text for recovery
```

### Tasks

- Route push-to-talk key-down and key-up events through a single coordinator.
- On key-down, leave `Idle` for `Starting` immediately, before microphone or overlay work.
- Remember a key-up that arrives during `Starting`. When capture becomes active, stop immediately and continue with the captured audio if that release was already seen. Otherwise enter `Recording`.
- Ignore a key-up only when no key-down is in progress.
- Ignore duplicate key-down events while already recording.
- After cancel, or after the recording duration limit, ignore further push-to-talk key-down events until the physical key-up.
- Prevent overlapping transcription requests.
- Do not register Escape, or any key other than the push-to-talk shortcut, as a cancel shortcut.
- Add a tray menu item, Cancel dictation, enabled only while the coordinator is `Recording` or `Transcribing`. Activating it stops the microphone, discards the captured audio and any partial transcript, does not paste, removes the overlay, and returns to `Idle`.
- Stop recording safely when the application is shutting down.
- Define behavior for missing models, missing microphones and revoked permissions.
- Keep all heavy processing off the UI/main thread.
- Ensure shortcut callbacks return quickly.

### Acceptance criteria

Unit tests cover valid transitions, duplicate events, a key-up that arrives during `Starting`, cancellation from `Recording` and `Transcribing`, the duration limit, failures and application shutdown.

An unexpected event must not leave recording active indefinitely or cause multiple recordings to run simultaneously.
