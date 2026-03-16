# AGENTS.md - npcs

## Project Overview

Go 1.24 multiplayer game competition system for pitting bots against each other. Module: `github.com/meschbach/npcs`.

Key packages:
- `t3/` - Tic-tac-toe game engine and gRPC networking
- `competition/` - Matchmaking, scoring, and game orchestration
- `cmd/` - CLI binaries: `competition`, `simple`, `simpled`, `t3`
- `junk/` - Shared utilities (in-process gRPC, process management)

## Build Commands

```bash
# Build individual binaries
go build -o bin/competition ./cmd/competition
go build -o bin/simple ./cmd/simple
go build -o bin/simpled ./cmd/simpled
go build -o bin/t3 ./cmd/t3

# Build all binaries
./dev.sh

# Cross-compile for release (linux/darwin, amd64/arm64)
./dev.sh release

# Generate gRPC/protobuf code (requires protoc)
./build-grpc.sh
```

## Test Commands

```bash
# Run all tests
go test -v -timeout 60s ./...

# Run a single test by name
go test -v -run TestGame ./t3

# Run a single subtest
go test -v -run "TestGame/When_asked_complete_4_turns" ./t3

# Run tests in a specific package
go test -v ./t3/...
go test -v ./competition/...

# Run unit tests only (pre-commit pattern)
go test -count 1 ./internal/...
```

## Lint and Format Commands

```bash
# Format code
gofmt -w .
goimports -w .

# Run linter (required before commit)
golangci-lint run --timeout=5m

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## Pre-commit Checklist

Run these in order (automated via `.pre-commit-config.yaml`):
1. `gofmt -w .`
2. `go mod tidy`
3. `go vet ./...`
4. `golangci-lint run --timeout=5m`
5. `go test -count 1 ./internal/...`
6. Build verification for all four binaries

## Code Style Guidelines

### Imports

Group imports in three blocks separated by blank lines:
1. Standard library
2. Third-party packages
3. Internal packages (`github.com/meschbach/npcs/...`)

```go
import (
    "context"
    "errors"
    "sync"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "google.golang.org/grpc"

    "github.com/meschbach/npcs/t3"
    "github.com/meschbach/npcs/competition/wire"
)
```

### Package Documentation

Document packages with a comment on the package line:
```go
// Package t3 implements a tic-tac-toe game engine.
package t3
```

### Naming Conventions

- Exported types use PascalCase: `Game`, `Board`, `Player`
- Unexported types use camelCase: `gamePhase`, `gameInstanceLobby`
- Constants use PascalCase with package prefix: `GameStatePreStart`, `SystemStarting`
- Interfaces are named by capability: `Player`, `Session`, `Auth`, `Network`
- Receiver names are short, typically single letter: `(t *Game)`, `(m *matcher)`, `(h *Hub)`

### Error Handling

- Return errors early; do not wrap in if/else blocks
- Use `errors.Join()` to combine multiple errors (handles nil automatically)
- Custom error types should implement `Error() string` and `Unwrap() error`
- Define sentinel errors as package-level vars: `var PlayerDisconnected = errors.New("player disconnected")`
- **Never** use `context.Background()` or `context.TODO()` (enforced by linter)
- In tests, use `t.Context()` for context

```go
// Good
func (p *PushClient) Serve(ctx context.Context) (problem error) {
    conn, err := grpc.NewClient(p.server, p.grpcOpts...)
    if err != nil {
        return err
    }
    defer func() {
        problem = errors.Join(problem, conn.Close())
    }()
    // ...
}
```

### Custom Error Types

```go
type PlayerError struct {
    WhichPlayer int
    Performing  string
    Underlying  error
}

func (p *PlayerError) Error() string {
    return fmt.Sprintf("encountered problem while player %d was %s: %s",
        p.WhichPlayer, p.Performing, p.Underlying)
}

func (p *PlayerError) Unwrap() error {
    return p.Underlying
}
```

### Testing

- All tests must call `t.Parallel()` at the top level
- Use `github.com/stretchr/testify/assert` for soft assertions
- Use `github.com/stretchr/testify/require` for hard assertions (fail fast)
- Use `t.Context()` for test context (never `context.Background()`)
- Use `t.Cleanup()` for teardown instead of defer where appropriate
- Subtests use descriptive names: `t.Run("Given player 1 has the first row", func(t *testing.T) { ... })`

```go
func TestGame(t *testing.T) {
    t.Parallel()
    ctx := t.Context()

    t.Run("When asked to step", func(t *testing.T) {
        t.Parallel()
        require.NoError(t, game.Step(ctx))
        assert.True(t, game.Concluded())
    })
}
```

### Concurrency Patterns

- Use `sync.Mutex` for simple state protection
- Use `sync.Cond` with a mutex for waiting on state changes
- Always pair Lock/Unlock, typically with defer
- Use channels for player input in game logic

### Logging

- Use structured logging via `log/slog`
- Pass context for trace correlation: `slog.InfoContext(ctx, "message", "key", value)`

### Interfaces

- Define interfaces where they are consumed, not where implemented
- Keep interfaces small and focused on a single capability
- Embed `Unimplemented*Server` in gRPC server implementations

### Linting Rules (enforced by golangci-lint)

- Cyclomatic complexity limit: 8 (gocyclo)
- Function length limit: 150 lines / 100 statements (funlen)
- Required linters: errcheck, govet, ineffassign, staticcheck, unused, gosec, gocritic, paralleltest, testifylint
- Forbidden: `context.Background`, `context.TODO`
- US English spelling (misspell)
- Check type assertions (errcheck)
- Duplicate code threshold: 80 tokens (dupl)
