# Phase 2 — Establish application structure

Part of the [implementation plan](../plan.md).

Create the basic application and its service boundaries.

Suggested package organization:

```text
cmd/
  app/
internal/
  app/
  config/
  shortcut/
  overlay/
  audio/
  models/
  transcription/
  paste/
  platform/
ui/
  settings/
  tray/
```

Adjust this structure to match the existing repository rather than enforcing it mechanically.

Define small interfaces for the major services:

```go
type AudioCapture interface {
    Start(ctx context.Context) error
    Stop() ([]float32, error)
}

type TranscriptionEngine interface {
    Transcribe(
        ctx context.Context,
        audio []float32,
        modelID string,
    ) (string, error)
}

type PasteService interface {
    Paste(ctx context.Context, text string) error
}
```

Add similarly narrow interfaces for push-to-talk events, overlay control and model management. Their concrete implementations should not leak operating-system types into application orchestration.

The audio and transcription signatures above are illustrative. Use the audio representation actually supported by the selected engine and add resampling explicitly where necessary.

### Acceptance criteria

- The application starts without showing a large main window.
- A tray icon and settings window are available.
- Configuration loads and saves successfully.
- Core services can be tested independently using fake implementations.
- The project remains buildable after each change.
