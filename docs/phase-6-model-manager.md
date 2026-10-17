# Phase 6 — Implement the model manager

Part of the [implementation plan](../plan.md).

Create a model catalog with explicit compatibility information.

A model definition should include:

- Stable model ID.
- Human-readable name.
- Transcription engine.
- Supported languages.
- Download URL.
- Expected file size, where known.
- Trusted expected SHA-256 digest.
- Model format and license/attribution information.
- Approximate resource requirements.

### Tasks

- Provide installed-model discovery.
- Show model descriptions, sizes and installation status.
- Download to a temporary filename.
- Display progress and support cancellation.
- Verify the downloaded file against a pinned, trusted digest.
- Atomically move verified files into their final locations.
- Recover cleanly from interrupted or corrupted downloads.
- Prevent simultaneous duplicate downloads of the same model.
- Allow removal of unused model files without deleting unrelated user data.
- Store models in the application's appropriate user-data directory.

Do not invent checksums or trust a checksum calculated from the first download. Use trusted upstream metadata or a pinned, verified artifact.

### Acceptance criteria

The model list distinguishes available models from installed models. Corrupt or incomplete files are never marked as installed. Download failures provide actionable errors, and existing installed models remain usable.
