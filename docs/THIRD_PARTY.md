# Third-party components

## Runtime and UI

- [MyGo](https://github.com/egoist/mygo) — desktop shell, windows, tray, clipboard.
- [malgo](https://github.com/gen2brain/malgo) — microphone capture.

## Speech recognition

- [whisper.cpp](https://github.com/ggml-org/whisper.cpp) — local inference (statically linked via CGO on macOS builds).
- Whisper model files (ggml format) from [ggerganov/whisper.cpp on Hugging Face](https://huggingface.co/ggerganov/whisper.cpp). Models are downloaded by speek into the user config directory, not bundled in the app.

## Models and attribution

OpenAI Whisper weights, distributed in ggml format by the whisper.cpp project. See model entries in `internal/models/catalog.go` for pinned URLs and SHA-256 checksums.

## Licenses

Review upstream LICENSE files in vendored or build-time dependencies before distribution. whisper.cpp and MyGo carry their own license terms.
