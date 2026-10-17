# Phase 7 — Implement local transcription

Part of the [implementation plan](../plan.md).

Integrate one transcription backend and prove the complete local inference path.

### Tasks

- Use `whisper.cpp` as the default backend unless Phase 1 establishes a concrete reason not to.
- Verify runtime availability and bundled executable compatibility.
- Start with a multilingual model suitable for everyday dictation.
- Allow explicit language selection and automatic language detection if supported.
- Load the selected model once per worker/session when the chosen runtime architecture permits.
- Avoid unnecessary model reloads between consecutive utterances.
- Capture standard output and error in a controlled way.
- Propagate cancellation and report engine failures.
- Report whether the engine detected speech. A no-speech result is silence even when the decoded text is non-empty. For `whisper.cpp`, use its no-speech or voice-activity signal before accepting a transcript.
- Keep all processing local.
- Do not expose a local HTTP inference service beyond loopback; preferably avoid opening a network listener unless the selected engine requires one.

If using a local worker executable, make its lifecycle explicit and terminate it cleanly when the application exits. Validate how model switching and concurrent requests work before choosing the worker architecture.

Measure latency with short and longer utterances. If starting with a command-line process that reloads the model for each utterance, measure the cost and document it before treating that as the final design.

### Acceptance criteria

A downloaded model successfully transcribes German and English test utterances with useful punctuation. A silence-only recording inserts no text, including when the engine returns non-empty text for audio with no speech. Transcription still works with networking disabled once the model and runtime are installed. Engine errors do not crash the desktop application.
