# AGENTS.md — MailMole Coding Agent Guidelines

## Project Overview

MailMole (`github.com/kocdeniz/mailmole`) is a terminal-first IMAP-to-IMAP email migration tool written in pure Go.
It features a TUI built with Bubble Tea + Lip Gloss, parallel folder workers, smart retry with
exponential backoff, O(1) duplicate detection, and checkpoint persistence.

**Module:** `github.com/kocdeniz/mailmole` (go.mod)  
**Go version:** 1.24.2  
**No Node.js, TypeScript, Python, or frontend tooling — pure Go only.**

---

## Build & Run Commands

```bash
# Run directly from source
go run .

# Build a portable static binary
CGO_ENABLED=0 go build -o mailmole .

# Build with version injection via ldflags
go build \
  -ldflags "-X github.com/kocdeniz/mailmole/internal/buildinfo.Version=1.2.0 \
             -X github.com/kocdeniz/mailmole/internal/buildinfo.BuildDate=$(date -u +%Y-%m-%d)" \
  -o mailmole .

# Run the compiled binary
./mailmole
```

## Lint & Vet

There is no external linter config (no `.golangci.yml`). Use the standard Go toolchain:

```bash
# Vet all packages (must pass clean)
go vet ./...

# Format all source files (gofmt-compatible)
gofmt -w .

# Check formatting without writing (CI-style)
gofmt -l .
```

## Tests

**There are currently no test files (`*_test.go`) in this repository.**

When adding tests, follow these conventions:

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./internal/sync/...

# Run a single test by name
go test ./internal/sync/... -run TestFunctionName

# Run with verbose output
go test -v ./internal/sync/... -run TestFunctionName

# Run with race detector
go test -race ./...
```

New test files must be placed alongside the package they test (e.g., `internal/sync/engine_test.go`)
and use `package sync` (white-box) or `package sync_test` (black-box) naming.

---

## Package Layout

```
main.go                  — Entrypoint only; zero business logic
internal/
  buildinfo/             — Version and build date injected via -ldflags
  imap/                  — Network layer: dial, auth, IMAP operations
  sync/                  — Migration engine, queue parser, checkpoint state, mock, report
  ui/                    — Bubble Tea model, update loop, view rendering, styles
```

Rules:
- `main.go` must stay a thin entrypoint; all logic belongs in `internal/`.
- Sub-packages under `internal/` are not importable outside the module.
- Each file carries a package-level doc comment: `// Package imap wraps go-imap/v2 for...`

---

## Code Style Guidelines

### Formatting

- All code must pass `gofmt` without diffs.
- Line length is not enforced by a tool, but keep lines readable (prefer ≤100 chars where natural).
- Use tabs for indentation (Go standard).

### Imports

Group imports in three blocks separated by blank lines, in this order:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "sync"

    // 2. Third-party modules
    tea "github.com/charmbracelet/bubbletea"
    imaplib "github.com/emersion/go-imap/v2"

    // 3. Internal packages
    imapconn "github.com/kocdeniz/mailmole/internal/imap"
)
```

- Alias third-party packages to avoid name collisions (e.g., `imaplib`, `tea`, `imapconn`).
- Never use dot-imports (`. "pkg"`).
- Remove all unused imports; `go vet` will catch them.

### Naming Conventions

| Kind | Convention | Example |
|------|------------|---------|
| Exported types/funcs/consts | PascalCase | `ConnectedMsg`, `RunMigration` |
| Unexported funcs/vars | camelCase | `migrateFolder`, `retryIMAP` |
| Enum-style constants | TypePrefix + Name | `PhaseIntro`, `StateSyncing`, `ConnReady` |
| Error variables | `ErrXxx` | `ErrConnectionLost` |
| Interface types | Noun or `-er` suffix | `StatusKind`, `Dialer` |
| Test functions | `TestFunctionName_scenario` | `TestRunMigration_RetryOnTimeout` |

- Use typed `int` aliases with `iota` for enum-like constants:
  ```go
  type AppPhase int
  const (
      PhaseIntro AppPhase = iota
      PhaseSelect
      PhaseSyncing
  )
  ```

### Types & Structs

- Configuration structs (`Config`, `EngineConfig`, `ConnConfig`) are plain value types — do not use pointers unless mutation across goroutines is required.
- Stats/state structs (`AccountStats`, `AccountState`) use pointer receivers only when methods mutate the struct.
- The root Bubble Tea `Model` struct (in `ui/model.go`) owns all UI and runtime state; keep it the single source of truth for TUI state.

### Error Handling

- Always wrap errors with context using `%w` for unwrappable chains:
  ```go
  return fmt.Errorf("append message to %s: %w", folder, err)
  ```
- Classify network/IMAP errors using `errors.As` + type/string matching; see `imap.IsRetryableError` and `imap.IsConnectionLostError`.
- Do **not** use `panic` anywhere; propagate or log all errors.
- Non-fatal errors (e.g., `FolderStatus` failure) should continue with zero/default values and log to the TUI activity log.
- Never silently discard errors (`_ = err` is forbidden except in deferred `Close` calls).

### Concurrency

- Worker pools: use a `jobs := make(chan string)` channel drained by goroutines with `sync.WaitGroup`.
- Protect all shared mutable state (`stats`, `adaptiveGap`, etc.) with `sync.Mutex`.
- Bubble Tea message bus: `chan StatusUpdateMsg` — one-directional, buffered at 256.
- Re-schedule channel reads inside `Update` via `tea.Cmd` (never block the update loop).
- Keep-alive goroutines must accept a `stop <-chan struct{}` for clean shutdown.
- Use `context.Context` for cancellation in all network-bound operations.

### Comments

- Every file must have a package-level doc comment.
- Use `// ---- Section name ----` dividers for logical sections within long files.
- Use `// ============================================================` for major phase boundaries.
- Annotate non-obvious logic with inline comments (e.g., `// O(1) lookup`, `// Idempotent:`).
- `//nolint:gosec` is acceptable only for the intentional `InsecureSkipVerify` on bare-IP TLS connections; document why.

### File & State I/O

- Write state files atomically: write to a `.tmp` file, then `os.Rename` to the final path.
- Open log files with `os.OpenFile(..., os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)`.
- Use `0o644` for files and `0o755` for directories.

---

## Environment Variable Overrides

| Variable | Default | Purpose |
|---|---|---|
| `MAILMOLE_MESSAGE_DELAY_MS` | 0 | Per-message delay in milliseconds |
| `MAILMOLE_FOLDER_WORKERS` | 3 | Folder worker goroutine count (max 3) |
| `MAILMOLE_STATE_FILE` | `migration_state.json` | Override checkpoint file path |
| `MAILMOLE_NO_IMAGE` | unset | Disable iTerm2 inline image on intro screen |

---

## Dependencies

Add dependencies with `go get` and commit both `go.mod` and `go.sum`. Do not vendor.  
Core libraries in use:

| Package | Purpose |
|---|---|
| `github.com/charmbracelet/bubbletea` | TUI event loop (Elm architecture) |
| `github.com/charmbracelet/bubbles` | Reusable TUI components |
| `github.com/charmbracelet/lipgloss` | Terminal styling/layout |
| `github.com/emersion/go-imap/v2` | IMAP client (aliased as `imaplib`) |
| `github.com/atotto/clipboard` | Clipboard access for credentials |

---

## CI / GitHub Actions

The only workflow (`.github/workflows/opencode.yml`) triggers the OpenCode AI assistant on
issue/PR comments containing `/oc` or `/opencode`. There is no automated build, test, or release
pipeline. When adding CI, add a new workflow file rather than modifying `opencode.yml`.
