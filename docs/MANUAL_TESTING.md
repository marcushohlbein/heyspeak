# Manual testing checklist

Run automated tests with `CGO_ENABLED=1 go test ./...` first. Use this list for platform behavior that unit tests cannot cover.

## Push-to-talk and focus

- [ ] Hold the shortcut while another app is focused; release and confirm text inserts at the caret.
- [ ] Tap the shortcut (release before recording starts); confirm the take still completes when audio was captured.
- [ ] Switch to another app during transcription; confirm no paste into the new app and transcript remains on the clipboard.
- [ ] Cancel from the tray during recording and during transcribing; confirm no paste.

## Overlay and displays

- [ ] Overlay appears without activating speek.
- [ ] Overlay tracks the display under the pointer at recording start.
- [ ] Unplug or rearrange monitors during a session (best effort).

## Audio and limits

- [ ] First capture prompts for microphone permission; denial shows a clear error.
- [ ] Hold past the max recording duration; confirm auto-stop and tray tooltip.

## Transcription and paste

- [ ] Silence-only recording shows “No speech detected” on overlay/tray without pasting.
- [ ] Paste into a text editor and a second app (browser, Notes, etc.).
- [ ] After successful paste, prior plain-text clipboard is restored when the preference is on.
- [ ] Deny Accessibility; confirm transcript stays on clipboard and error explains the setting.

## Settings and models

- [ ] Download, remove, and re-download a model.
- [ ] Invalid shortcut is rejected on save.
- [ ] Save applies shortcut reload.

## Lifecycle

- [ ] Launching a second instance exits with “already running”.
- [ ] Quit during recording or transcription shuts down cleanly.

Logs should describe steps and errors without printing dictated transcript text.
