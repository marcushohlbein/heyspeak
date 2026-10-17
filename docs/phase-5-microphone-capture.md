# Phase 5 — Implement microphone capture

Part of the [implementation plan](../plan.md).

Integrate the chosen audio capture library.

### Tasks

- Enumerate and select audio input devices if supported.
- Use an appropriate capture format and convert it to the transcription engine's required input format.
- Capture audio only while the coordinator is `Starting` or `Recording`.
- Provide a bounded buffer and a configurable maximum recording duration.
- When the maximum recording duration is reached, stop capture even if the shortcut is still held, show the transcribing overlay, and follow the normal transcription and paste path. Set the tray tooltip to "Recording reached the time limit" until the next successful paste. Ignore push-to-talk key-down events until the physical key-up.
- Release audio devices and resources after failures or cancellation.
- Handle missing devices, permission denial and device disconnection.
- Keep captured audio in memory where practical.

For `whisper.cpp`, validate the expected input requirements and resample as needed rather than assuming microphone sample rates match the model's expected rate.

### Acceptance criteria

A recorded utterance can be inspected using a test-only diagnostic path. Normal application logging never records the audio itself. Repeated recording sessions do not leak resources or accumulate unbounded memory. A recording that hits the duration limit produces one transcript, and the still-held shortcut does not start a second recording.
