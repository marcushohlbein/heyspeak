# Phase 9 — Integrate the end-to-end workflow

Part of the [implementation plan](../plan.md).

Connect all services through the coordinator.

### Key-down

1. Validate that the app is idle and a usable model is installed. On failure, show the error and stay `Idle`.
2. Leave `Idle` for `Starting` before microphone or overlay work.
3. Capture the current target and focus context where supported.
4. Determine the display nearest the pointer.
5. Start microphone capture.
6. Show the overlay without activation.
7. If key-up arrived during `Starting`, stop capture immediately and continue with the captured audio. Otherwise transition to `Recording`.

### Key-up

1. If the coordinator is `Starting`, remember the release and apply it when capture becomes active.
2. If the coordinator is `Recording`, stop microphone capture.
3. Ignore the key-up when no key-down is in progress.
4. Hide or transition the overlay to the transcribing state.
5. Transition to `Transcribing`.
6. Invoke the local transcription engine.
7. Treat an empty result, or a result the engine reports as no speech, as silence even when the returned text is non-empty. Do not paste. Show the empty-result error, remove the overlay, and return to `Idle`.
8. Before pasting, compare focus with the target captured on key-down. If this app or the overlay took focus, restore that target and paste there. If a different application is focused, do not send a paste shortcut and do not activate any app. Keep the transcript on the clipboard, show the error, remove the overlay, and return to `Idle`.
9. Paste into the original target application, then restore the previous plain-text clipboard contents when that restoration is still safe.
10. Transition to `Idle`.

When the maximum recording duration is reached while the shortcut is still held, follow this same path from the stop-capture step. Ignore further key-down events until the physical key-up.

Every step needs cancellation and failure handling. On a blocked start or a failed dictation step, show a short message on the non-activating overlay and set the tray tooltip to the same text. If the overlay cannot be created, use the tray tooltip alone. Cover these cases: the app is not idle, no usable model is installed, the microphone is missing or permission is denied, the overlay cannot be created, the result is empty or silence-only, transcription fails, or paste or input permission fails. Do not paste an empty or silence-only result. When paste or input permission fails, leave the transcript on the clipboard. Then remove the overlay, enter `Error`, and return to `Idle` once that message is showing. Preserve any recognized text for recovery.

Cancellation uses the tray item from Phase 3. It is available during `Recording` and `Transcribing`. It discards the audio and any partial transcript, does not paste, removes the overlay, and returns to `Idle`.

### Acceptance criteria

A user can hold the shortcut, dictate a sentence, release the shortcut and see the correct text inserted into the intended application without manually interacting with the transcription app.

A tap that releases the shortcut before capture is ready still ends that take. A silence-only recording inserts no text. Switching to another application during transcription leaves the transcript on the clipboard instead of pasting into the new application.

The app is ready for the next utterance after successful completion or recoverable failure.
