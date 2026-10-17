# Phase 8 — Implement paste service

Part of the [implementation plan](../plan.md).

Create a platform adapter for inserting text into the previously focused application.

### Tasks

- Preserve the target application's focus while recording and transcribing.
- On key-down, capture the target application where the platform supports it.
- At paste time, compare the focused application with that target.
- If this app or its overlay took focus, restore the captured target and paste there.
- If a different application is focused, do not send a paste shortcut and do not activate any application. Keep the transcript on the clipboard, show the error, remove the overlay, and return the coordinator to `Idle`.
- Put the transcript onto the system clipboard.
- Send the appropriate native paste shortcut: typically `Ctrl+V` on Windows/Linux and `Cmd+V` on macOS.
- Use native input injection where the operating system requires it.
- Provide a clear permission error if automated input requires permission that has not been granted.
- After a successful paste, restore the previous plain-text clipboard contents immediately. This is the default. Restore only when the clipboard still contains the plain text this operation inserted.
- Preserve the transcript in the clipboard when pasting fails so the user can recover it.
- Avoid overwriting clipboard content that the user has changed in the meantime.

Do not assume that changing the clipboard automatically inserts text. Do not claim that all clipboard formats can be restored perfectly. Operating-system clipboard history and cloud clipboard sync can still retain or upload the transcript during the short window before restoration.

### Acceptance criteria

Dictated text is inserted at the cursor position in a standard text editor and at least one additional target application on the initial supported platform. The transcription application never steals focus during the workflow. If the user has switched to another application, the transcript stays on the clipboard and is not inserted there. After a successful paste, the previous plain-text clipboard contents are restored. Paste failures are recoverable from the clipboard.
