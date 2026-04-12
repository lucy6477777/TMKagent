# Testing Guide

## Goals

- keep fast feedback for local development
- cover pure logic aggressively with unit tests
- exercise API adapters through local mock servers
- keep a thin layer of real integration tests for end-to-end confidence

## Test Types

### Unit Tests

Run by default:

```bash
make test
```

These should cover:

- config loading and overrides
- audio file handling and VAD behavior
- websocket / upload helpers
- translation / ASR / TTS client request formatting using mock HTTP servers
- pure pipeline helpers

### Coverage Run

```bash
make test-cover
```

This writes `coverage.out` and prints per-package coverage via `go tool cover -func`.

Recommended workflow:

1. run `make test-cover`
2. look for the lowest-covered package with meaningful logic
3. add tests for pure helpers before touching integration coverage

### Integration Tests

Run only when intentionally enabled:

```bash
OPENAI_API_KEY=sk-... make test-integration
```

Integration tests live under `tests/integration` and use the `integration` build tag.

Use integration tests for:

- real Whisper transcription behavior
- API compatibility checks that cannot be trusted from mocks alone

Do not use integration tests for:

- simple branching logic
- parsing helpers
- config behavior
- deterministic handler behavior

## Coverage Expectations

The target is high confidence, not meaningless line-count inflation.

Preferred order for improving coverage:

1. pure functions
2. HTTP handlers and adapter code
3. constructor/default behavior
4. integration-only paths

If the project claims `90%+` coverage, that should come from:

- strong unit coverage in core packages
- not from counting generated assets or untestable external SDK internals

## Adding New Tests

When adding tests:

- keep them deterministic
- avoid real network calls unless the file is explicitly an integration test
- prefer table-driven tests for pure parsing and config logic
- prefer package-internal tests when unexported helpers need coverage
- keep test names behavior-based, not implementation-based

Examples of good targets:

- `publicAccessInfo`
- `.env` loading precedence
- chunk splitting
- translation response parsing
- upload handler request validation

## Current Commands

```bash
make test
make test-cover
make test-integration
```

## When Coverage Stalls

If a package remains low-coverage, ask:

- can the logic be extracted into a pure helper?
- can the external API be replaced by a local `httptest.Server`?
- is the remaining code mostly thin SDK glue that should be covered by one focused smoke test instead of many mocks?
