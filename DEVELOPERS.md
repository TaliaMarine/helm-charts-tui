# Developer Guide

## Architecture

The application follows the [Bubble Tea](https://github.com/charmbracelet/bubbletea) model-update-view architecture with a single top-level model managing all state.

### Package Layout

```
main.go                        Entry point, --version flag, signal handling
internal/
  helm/
    types.go                   Data types: Repo, Chart
    executor.go                Executor interface + RealExecutor (calls helm CLI)
    mock.go                    MockExecutor for tests
    executor_test.go           JSON parsing unit tests
  tui/
    model.go                   Model struct, screen/mode enums, New(), Init()
    update.go                  Update() -- input handling, state transitions
    view.go                    View() -- rendering per screen
    commands.go                tea.Cmd factories (async helm calls)
    messages.go                Custom tea.Msg types
    keys.go                    Key binding definitions
    styles.go                  Lipgloss styles, NO_COLOR support
    filter.go                  List filtering logic
    filter_test.go             Filter unit tests
    model_test.go              State transition tests
```

### Design Decisions

- **Single model with screen enum**: Rather than nested Bubble Tea models, there's one `Model` with a `Screen` enum (`ScreenRepoList`, `ScreenChartList`, `ScreenChartVersions`, `ScreenChartDetail`) and a `Mode` enum (`ModeNormal`, `ModeFilter`, `ModeAddRepo`, `ModeConfirmExit`, `ModeHelp`).
- **Executor interface**: All helm CLI interactions go through `helm.Executor`. The real implementation uses `exec.CommandContext`; tests use `MockExecutor` with configurable function fields.
- **Bubbles components**: `table.Model` for the 3 list screens, `viewport.Model` for the detail screen, `textinput.Model` for filter input and add-repo form.
- **Chart counts**: Loaded via a single `helm search repo ""` call at startup, then grouped by repo prefix. This avoids N separate helm calls.
- **Filter behavior**: Active filter clears on screen navigation. First ESC clears filter; second ESC navigates back.

### State Machine

```
ScreenRepoList <-> ScreenChartList <-> ScreenChartVersions <-> ScreenChartDetail
     |
     +-- ModeAddRepo (Enter on "Add new repo" row)
     +-- ModeConfirmExit (ESC on repo list)

Any screen:
     +-- ModeFilter (/ key on list screens)
     +-- ModeHelp (? key)
```

### Update Flow

The `Update` method dispatches in layers:

1. **Window resize** -- always handled first
2. **Global quit keys** -- `Ctrl+C` always quits; `q` quits in normal/help mode
3. **Async data messages** -- repos loaded, charts loaded, errors, etc.
4. **Mode-specific key dispatch** -- filter input, add-repo form, confirm-exit dialog, help overlay, or normal navigation

## Building

```sh
just build              # Development build -> bin/helm-charts-tui
just build-release      # Build with embedded version, commit, date
just clean              # Remove bin/
```

Linker flags for version info:

```sh
go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc1234 -X main.date=2025-01-01"
```

## Testing

```sh
just test               # go test -race ./...
just test-verbose       # go test -race -v ./...
just test-coverage      # Coverage report
```

### Test Structure

- **`internal/helm/executor_test.go`** -- Table-driven tests for JSON parsing of helm output.
- **`internal/tui/filter_test.go`** -- Unit tests for `FilterRepos()` and `FilterCharts()`.
- **`internal/tui/model_test.go`** -- State transition tests: create a model with `MockExecutor`, send messages, assert screen/mode changes. Covers navigation, filtering, add-repo flow, quit behavior, error handling.

### Adding a New Screen

1. Add a constant to `Screen` enum in `model.go`
2. Add data fields to `Model` struct
3. Add case to the Update switch for key handling in `update.go`
4. Add rendering in `view.go`
5. Add a loading command in `commands.go`
6. Add message types in `messages.go`
7. Add tests in `model_test.go`

## Code Quality

```sh
just check              # Runs: tidy, fmt, vet, test, build
just vet                # go vet ./...
just lint               # golangci-lint run ./...
just fmt                # gofmt -w .
just tidy               # go mod tidy
```

## Dependencies

Direct dependencies (minimal set):

| Package | Purpose |
|---------|---------|
| `charmbracelet/bubbletea` | TUI framework (model-update-view) |
| `charmbracelet/bubbles` | UI components: table, viewport, textinput, key |
| `charmbracelet/lipgloss` | Terminal styling |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generic error |
| 130 | Interrupted (SIGINT/SIGTERM) |
