# Phase 11 — Packaging and release readiness

Part of the [implementation plan](../plan.md).

### Tasks

- Define the initial supported operating system and architecture.
- Package the runtime executable and required native resources appropriately.
- Ensure model files remain outside the application bundle.
- Review third-party dependency licenses and model attribution.
- Document first-run permissions and setup.
- In that first-run documentation, state that operating-system clipboard history and cloud clipboard sync can retain or upload clipboard contents during the short paste window, before the previous clipboard contents are restored.
- Validate startup, single-instance behavior and clean shutdown.
- Test the packaged application on a clean environment.
- Document unsupported platforms or platform-specific limitations explicitly.

### Acceptance criteria

A clean installation can launch the app, install a model, configure the shortcut and complete the end-to-end workflow without development-only tools or manually installed inference dependencies.
