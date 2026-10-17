# speek

Local push-to-talk dictation for macOS (Apple Silicon and Intel). Hold a global shortcut, speak, release — text is transcribed with whisper.cpp and pasted into the app you were using.

## Supported platforms

| Platform | Status |
|----------|--------|
| macOS 13+ (arm64, amd64) | Supported — PTT, overlay, paste, in-process whisper |
| Linux / Windows | Not supported yet (build stubs only) |

## First run

1. **Build native whisper libraries** (one time per machine):

   ```bash
   make whisper
   ```

2. **Build and run**:

   ```bash
   make speek
   ./bin/speek
   ```

   Or: `CGO_ENABLED=1 go run ./cmd/app`

3. **Permissions (macOS)**  
   - **Microphone** — prompted on first recording.  
   - **Accessibility** — required for global push-to-talk and automated Cmd+V paste. Enable speek in **System Settings → Privacy & Security → Accessibility**.

4. **Model** — Click the **speek** icon in the menu bar (top right), choose **Settings…**, confirm model ID (default `ggml-tiny`), click **Download model**, then **Save**.

   If you do not see the icon, check the macOS menu bar overflow (›) or run with settings open:

   ```bash
   ./bin/speek -settings
   ```

5. **Clipboard** — After a successful paste, speek restores your previous plain-text clipboard when that option is enabled in Settings. macOS clipboard history and iCloud clipboard may still copy the transcript during the short window before restoration.

Only one speek instance runs at a time.

## Configuration

Settings and downloaded models live under:

`~/Library/Application Support/speek/`

Model binaries are stored in `models/` beside `config.json` and are **not** placed inside the application bundle.

## Packaging

```bash
make package
```

Produces `dist/Speek.app` with the built binary. You still need whisper.cpp built (`make whisper`) before compiling speek. Metal shader sources are resolved from `third_party/whisper.cpp` at runtime unless `SPEEK_WHISPER_ROOT` is set.

## Development

See [docs/MANUAL_TESTING.md](docs/MANUAL_TESTING.md) and phase docs in [docs/](docs/). Third-party notices: [docs/THIRD_PARTY.md](docs/THIRD_PARTY.md).

## Releases

Versioning is managed by [release-please](https://github.com/googleapis/release-please). Use [Conventional Commits](https://www.conventionalcommits.org/) on `main` (`feat:`, `fix:`, etc.); the workflow opens or updates a release PR with changelog and version bumps. Merge that PR to tag a GitHub release.

Current version is tracked in `internal/version/version.go` and `scripts/macos/Info.plist`.

## Spikes

- `go run ./cmd/spike-capture-paste` — mic + clipboard + paste chord  
- Overlay and PTT spikes under `cmd/`
