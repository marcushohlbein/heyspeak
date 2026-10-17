# Phase 0 — Repository reconnaissance

Part of the [implementation plan](../plan.md).

### Tasks

1. Inspect the repository structure, `go.mod`, build configuration, application lifecycle and existing tests.
2. Read the current implementation of the MyGo glow example and native launcher example.
3. Verify the current APIs for:
   - Window creation and lifecycle.
   - Transparent windows and full-display positioning.
   - Non-activating window display.
   - Pointer/display detection and DPI scaling.
   - Clipboard operations.
   - Global shortcut registration and deregistration.
   - System tray and application shutdown.
4. Determine the project's actual target operating system and available development environment.
5. Establish that the repository builds before modifying it.
6. Identify the simplest viable audio capture and local inference integrations.

### Deliverables

- A brief architecture and dependency assessment.
- A baseline build and test result.
- A list of platform capabilities and unresolved risks.
- A proposed initial implementation sequence.

### Acceptance criteria

The agent can identify the relevant MyGo APIs and examples and can build the starting application.

Do not begin a large implementation before resolving the feasibility questions in Phase 1.

---

## Reconnaissance report (completed)

Date: 2026-10-17. Upstream reference: `github.com/egoist/mygo` @ shallow clone (mygo CLI **0.5.5**, `go 1.27.1` in upstream `go.mod`).

### Repository state (speek)

| Item | Finding |
|------|---------|
| Structure | Greenfield: `plan.md` + `docs/phase-*.md` only. No `go.mod`, `cmd/`, or application code yet. |
| Build | **Cannot build speek** until Phase 2 scaffolding adds a module and `main`. |
| Tests | None in this repo. |

Initial target platform matches the dev machine: **macOS 27 (Darwin) arm64**, **Go 1.26.0**. `cmake` present; `pkg-config` not on PATH. **`mygo` CLI not installed** globally; built successfully from upstream via `go build ./cmd/mygo`.

### Architecture and dependency assessment

- **Application shell:** [MyGo](https://github.com/egoist/mygo) — Go-native desktop runtime with `mygo.App` lifecycle (`WhenReady`, `Run`, `OnWillQuit`, `SetActivationPolicy`, single-instance helpers), `mygo.NewWindow`, system tray (`mygo.NewTray`), and modular APIs in `modules.go` / `window.go`.
- **Overlay UI:** Native `github.com/egoist/mygo/ui` (not webview) for the glow; adapt patterns from `examples/chat-glow` (edge glow) and window positioning from `examples/launcher`.
- **Audio (planned):** `github.com/gen2brain/malgo` v0.11.26 — CGO bindings to miniaudio; `go get` succeeds on this host.
- **Transcription (planned):** `github.com/ggerganov/whisper.cpp/bindings/go` — module resolves; expects CGO + built whisper.cpp artifacts (verify in Phase 1/7 spike).
- **Config / HTTP:** stdlib JSON + `net/http` per plan.

### Baseline build and test (upstream MyGo)

On the same machine, against a fresh clone of `egoist/mygo`:

| Check | Result |
|-------|--------|
| `go test ./... -count=0` | **Pass** (108 packages) |
| `go build ./examples/launcher` | **Pass** |
| `go build ./examples/chat-glow` | **Pass** |
| `go build ./cmd/mygo` | **Pass** → `mygo 0.5.5` |

### MyGo examples reviewed

1. **`examples/chat-glow`** — Edge glow for speek’s overlay visual:
   - Full-window pass-through layers: `ui.Box(...).Absolute().PassThrough()` with `BorderConicGradient` + `Blur` (`app.edge` in `app.go`).
   - GPU-gated (`p.Accelerated()`); degrades gracefully without GPU.
   - Default window is a normal sized chat window; speek must use **frameless, transparent, display-sized** window options (see launcher + `Window.ShowInactive` / `SetIgnoreMouseEvents`).
2. **`examples/launcher`** — Background / tray / global shortcut pattern:
   - `mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)`, `OnWindowAllClosed` no-op to stay alive.
   - `mygo.GlobalShortcut.Register` with fallback accelerators; **callback on press only** (no key-up).
   - `mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())` + `SetPosition` for multi-monitor.
   - Tray menu with `RoleQuit`, `OnWillQuit` → `tray.Destroy()`.
   - Uses `Panel: true` + vibrancy for a floating panel; overlay will differ (full display bounds).

### Platform capabilities (MyGo APIs)

| Capability | API / location | Notes for speek |
|------------|----------------|-----------------|
| Window lifecycle | `mygo.NewWindow`, `Show`/`Hide`/`Close`, handlers `OnClose`, `OnBlur` | Launcher hides instead of quitting. |
| Transparent / frameless | `WindowOptions`: `Frameless`, `Transparent`, `Vibrancy` | See `examples/frameless`. |
| Non-activating show | `Window.ShowInactive()` | `window.go` |
| Click-through | `Window.SetIgnoreMouseEvents(true)` | Pair with `PassThrough` in UI. |
| Display + DPI | `Screen.Displays()`, `Bounds`, `WorkArea`, `ScaleFactor`, `DisplayNearestPoint`, `OnDisplaysChanged` | Launcher uses scale for icons; overlay must use **Bounds** in DIPs (supports negative origins). |
| Clipboard | `Clipboard.ReadText` / `WriteText` (+ HTML/image) | Paste spike still needs **Cmd+V to foreground app** without focus — not a dedicated MyGo API; likely AX / CGEvent (Phase 1C). |
| Global shortcut | `GlobalShortcut.Register` / `Unregister` / `UnregisterAll` | macOS: Carbon `RegisterEventHotKey` — **press only** (`internal/darwin/hotkey.go`). Hold-to-talk needs Phase 1A adapter. |
| Tray / quit | `NewTray`, `SetMenu`, `App.Quit`, `OnWillQuit` | Documented in launcher. |

### Unresolved risks

1. **Push-to-talk key release** — `GlobalShortcut` fires on activation only; Phase 1A must add native key-down/up tracking (plan already flags this).
2. **Paste without focus** — Clipboard write is supported; synthesizing paste in another app may require macOS Accessibility permission and is fragile on Wayland later.
3. **Full-display overlay** — chat-glow is in-window edges; speek must combine display bounds + `ShowInactive` + ignore mouse events + display change listener; verify no focus steal on macOS.
4. **whisper.cpp packaging** — CGO, model paths, and binary size; spike before committing to binding layout.
5. **malgo + UI thread** — Capture must stay off main thread (plan requirement).
6. **Tooling** — Install or vendor `mygo` CLI for `mygo build` / bundling (Phase 11).

### Proposed initial implementation sequence

1. **Phase 1** — Spikes: key hold/release, display-sized glow overlay, mic + clipboard + paste.
2. **Phase 2** — `go mod`, `cmd/app`, service interfaces, MyGo app shell (tray-only, accessory policy).
3. **Phases 3–4** — State machine + overlay manager (port chat-glow edge to per-display window).
4. **Phases 5–7** — malgo capture, model download manager, whisper engine adapter.
5. **Phases 8–9** — Paste service + end-to-end PTT workflow.
6. **Phases 10–11** — Settings, tests, `mygo` packaging.

### Acceptance criteria (Phase 0)

- [x] Relevant MyGo APIs and examples identified (`chat-glow`, `launcher`, `frameless`, `modules.go`, `window.go`).
- [x] Upstream MyGo builds on dev OS.
- [ ] Speek application build — **blocked** until Phase 2 creates the module (expected).

**Next step:** Phase 1 technical feasibility spikes; do not start full app workflow until 1A–1C pass on macOS.
