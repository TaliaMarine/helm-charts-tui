---
description: 'Instructions for writing Go CLI & TUI applications with idiomatic Go and usability best practices'
applyTo: '**/*.go,**/go.mod,**/go.sum'
---

# Go CLI & TUI Development Instructions

Follow idiomatic Go practices and community standards with a focus on **command‑line interfaces (CLI)** and **terminal user interfaces (TUI)**. These instructions adapt guidance from community norms such as Effective Go, Go Code Review Comments, and widely adopted CLI/TUI conventions.

> **Scope**: Go projects that produce terminal applications (CLI commands and interactive TUIs). General Go style guidance remains in force unless overridden by a CLI/TUI specific rule below.

---

## 1) Design Principles for CLI/TUI

- **Clarity first**: simple, predictable behavior; principle of least surprise.
- **Fast startup**: CLIs must start quickly. Defer heavy initialization until needed.
- **Consistent UX**: flags, outputs, colors, and key bindings are consistent across commands.
- **Discoverable**: great `--help`, examples, and autocompletion.
- **Accessible**: works without colors, respects terminal width, accessible key bindings.
- **Scriptable**: stable stdout for data, stderr for human messages; machine-readable modes (`--output json`).
- **Portable**: works across Linux/macOS/Windows; degrades gracefully on limited terminals/CI.
- **Reliable**: clear errors, actionable suggestions, proper exit codes.
- **Interruptible**: respect Ctrl‑C; support timeouts and cancellation via `context.Context`.

---

## 2) CLI Specific Guidance

### 2.1 Command & Subcommand Design

- Prefer a **noun-verb** hierarchy that scales: `tool repo clone`, `tool repo list`, `tool user add`.
- Keep top-level concise; group topics into logical subcommands.
- Provide **short** (single-letter) and **long** flags for frequent options: `-o, --output`.
- Avoid breaking changes; if unavoidable, guard behind feature flags and document migration.

### 2.2 Flags, Arguments, and Configuration

- **Flag naming**: kebab-case (`--access-token`), avoid underscores.
- **Defaults**: choose sensible defaults; make zero values useful.
- **Boolean flags**: provide `--no-<flag>` for disabling (`--color/--no-color`).
- **Mutually exclusive**: validate at parse-time; produce a clear error.
- **Repeatable**: allow repeated flags for lists (`--label a --label b`).
- **Required**: fail early with explicit messages.
- **Config precedence** (highest → lowest):
  1. CLI flags
  2. Environment variables
  3. Config file(s) (e.g., `$XDG_CONFIG_HOME/app/config.yaml`)
  4. Project-local config (e.g., `.app/config`) when applicable
  5. Built-in defaults
- Provide `--config` to point to a file and `--print-config` to show the effective configuration.
- Offer `--output/-o` with values like `text` (default), `json`, `yaml` and **guarantee stable schemas**.

### 2.3 I/O Conventions

- **stdout**: primary, parseable output; **stderr**: logs, progress, prompts.
- Detect **non‑TTY**: disable spinners, progress bars, and colors; avoid interactive prompts unless `--yes`.
- Respect **`NO_COLOR`** env var to disable color; provide `--no-color` and `--color=auto|always|never`.
- Provide `--quiet/-q` to minimize non-essential output.
- Respect **terminal width** when wrapping; avoid truncation of critical info.
- **Exit codes**:
  - `0` success
  - `1` generic error
  - `2` usage error (bad flags/args)
  - `3` unavailable/unsupported environment (e.g., missing dependency)
  - `130` interrupted by signal (Ctrl‑C)
  - Add domain-specific codes sparingly and document them.

### 2.4 Help, Usage, and Examples

- `--help` on any command prints usage, flags, defaults, env overrides, and **copy‑pasteable** examples.
- Provide `--version` including version, commit, build date, and OS/arch.
- Generate **shell completions** for bash, zsh, fish, PowerShell and document installation.

### 2.5 Networking & HTTP (Clients)

- Accept `context.Context` in all network operations; honor deadlines and cancellation.
- Configure `*http.Client` (timeouts, transport). Do not mutate transport after first use.
- Construct a **fresh** `*http.Request` per call; do not store per‑request state on a long‑lived client.
- Always close response bodies (`defer resp.Body.Close()`).
- Support proxies and standard envs (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`).

---

## 3) TUI Specific Guidance

### 3.1 When to Build a TUI

- Prefer TUI when the task benefits from live feedback, navigation, or multi-pane status. Prefer plain CLI for single-shot commands, scripting, and CI.

### 3.2 Event Loop & Rendering

- Use a framework with a clean event/update/view model (e.g., Bubble Tea) or `tcell` for lower-level control.
- Keep the render loop **pure**: compute view from state; mutate state in response to messages/events.
- Handle **window resize**; reflow content to the current width/height.
- Avoid excessive rendering; debounce updates when streaming logs.

### 3.3 Navigation & Key Bindings

- Keyboard-first; optional mouse support. Provide **discoverable** help overlay (`?`) listing keys.
- Use conventional bindings:
  - Navigation: `↑/↓`, `PgUp/PgDn`, `Home/End`
  - Quit: `q`, `Ctrl+C`
  - Confirm: `Enter`
  - Search: `/`, next `n`, prev `N`
  - Help: `?`
- Offer **vim-style** alternatives (`j/k`, `g/G`) when suitable.

### 3.4 Accessibility & Color

- High-contrast defaults; avoid color-only distinctions; use symbols and labels.
- Respect `NO_COLOR` and provide monochrome mode. Provide `--theme=dark|light|none`.
- Use ANSI safely; normalize on Windows; avoid 24-bit color assumptions—fallback to 16/256-color.
- Do not rely on emoji glyph availability.

### 3.5 Input, Forms, and Secrets

- Validate input inline with clear messages; keep focus in context.
- Mask secrets; support paste from clipboard; allow "show/hide" toggle.
- Provide default values and placeholders.

### 3.6 Logs, Progress & Long-Running Tasks

- Show determinate progress bars when the total is known; indeterminate spinners otherwise.
- Always allow cancel; reflect cancellation status immediately.
- Stream logs to a dedicated pane; support pause and filtering.

### 3.7 Error Surfaces

- Keep errors readable, one line summary with optional details (`d` to expand details/stack in TUI or use `--verbose` in CLI mode).
- Offer retry actions where feasible.

---

## 4) Cross-Platform Terminal Considerations

- Detect TTY with `isatty`; disable interactive features when not attached to a TTY.
- On Windows, enable virtual terminal processing or use colorable writers.
- Use `golang.org/x/term` for terminal size and raw mode; avoid OS-specific syscalls directly.
- Handle locales and Unicode width; avoid breaking wide characters when truncating.

---

## 5) Code Style & Naming (Go-specific)

- Write simple, idiomatic Go; keep the happy path left-aligned and return early.
- Prefer standard library; avoid unnecessary dependencies.
- **Package names**: lowercase, single word, no underscores; singular nouns.
- **Avoid stutter**: `cli.Command`, not `cli.CLICommand`.
- **Exported symbols** start with a capital; unexported with lowercase.
- **Comments**: explain *why*, not obvious *what*; English by default.
- Format with `gofmt`; manage imports with `goimports`.

### 5.1 Package Declaration Rules (CRITICAL)

- Each Go file must have **exactly one** `package` line.
- When editing, **preserve** the existing `package` declaration.
- New files must use the same package name as others in the directory (or the directory name for new packages).
- When replacing file content, ensure the new content has only one `package` declaration at the top.

---

## 6) Concurrency, Cancellation, and Signals

- Avoid background goroutines in libraries unless documented with lifecycle and cleanup.
- Use `context.Context` to propagate cancellation and timeouts.
- Treat Ctrl‑C (`SIGINT`) as cancellation and exit with code `130`.
- Ensure goroutines exit; use `sync.WaitGroup` or channels to coordinate.

**Example: signal-aware context in `main`**

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := run(ctx); err != nil {
        // human-friendly message to stderr
        fmt.Fprintln(os.Stderr, err)
        // use 130 when interrupted
        if errors.Is(err, context.Canceled) {
            os.Exit(130)
        }
        os.Exit(1)
    }
}

func run(ctx context.Context) error {
    // Simulate work that respects ctx
    ticker := time.NewTicker(200 * time.Millisecond)
    defer ticker.Stop()
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // progress to stderr
            fmt.Fprintf(os.Stderr, "step %d/10\n", i+1)
        }
    }
    fmt.Println("{"status":"ok"}") // stdout: machine‑readable
    return nil
}
```

---

## 7) Output, Colors, and Formatting

**Detect TTY and color**

```go
w := colorable.NewColorableStderr() // Windows-safe
isTTY := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
noColor := os.Getenv("NO_COLOR") != "" || !isTTY || forceNoColor

style := lipgloss.NewStyle()
if noColor {
    style = style.UnsetForeground().UnsetBackground()
}
fmt.Fprintln(w, style.Bold(true).Render("Processing..."))
```

**Width-aware wrapping**

```go
width, _, err := term.GetSize(int(os.Stdout.Fd()))
if err != nil || width <= 0 { width = 80 }
wrapped := wrapTo(width, text) // implement or use a library
```

---

## 8) Errors, Messages, and Suggestions

- Keep error messages lowercase, no trailing punctuation; add context with `%w`.
- Do **not** log and return the same error; choose one level to handle.
- Provide suggestions: `did you mean --output json?` when parsing fails.
- For CLI: short message to stderr by default; detailed traces under `--verbose`.

```go
if err != nil {
    return fmt.Errorf("reading config %q: %w", path, err)
}
```

---

## 9) Testing CLIs & TUIs

- **Unit tests**: table-driven, focus on pure logic.
- **Golden tests**: compare CLI stdout/stderr against golden files; normalize nondeterminism (timestamps, paths).
- **End‑to‑end**: run `exec.CommandContext` with temp dirs/env; assert exit codes.
- **PTY tests** for TUIs: exercise resize, key events; skip in CI when terminal not available.
- **Fuzzing**: for parsers and format readers.

```go
cmd := exec.CommandContext(ctx, binary, "--output", "json")
cmd.Env = append(os.Environ(), "NO_COLOR=1")
out, err := cmd.CombinedOutput()
if exitErr, ok := err.(*exec.ExitError); ok {
    if exitErr.ExitCode() != 0 { t.Fatalf("unexpected exit: %v", err) }
}
compareGolden(t, "list.json", out)
```

---

## 10) Security Best Practices

- **Never** build shell commands via concatenated strings; use `exec.Command(name, args...)`.
- Sanitize and validate any file paths; prevent path traversal.
- Mask secrets in logs and screens; provide `--redact` for dumps.
- Use `crypto/rand` for randomness; avoid custom crypto.
- Follow least-privilege for file permissions (e.g., `0600` for secrets).

---

## 11) Documentation & Help Generation

- Auto-generate help/usage from your flag/command framework.
- Include README sections: install, quickstart, examples, configuration, completions, troubleshooting.
- Offer `man` pages or `--help-man` output where appropriate.

---

## 12) Packaging, Releases, and Telemetry

- Use reproducible builds: `-trimpath`, avoid embedding volatile paths.
- Cross-compile and package with `goreleaser`; provide checksums and SBOM where feasible.
- Provide completion scripts and installers; avoid requiring admin permissions.
- Version with SemVer; include commit and build date in `--version`.
- Telemetry **must be opt‑in**; clearly document what is collected, how to disable, and respect env toggles.

---

## 13) Recommended Libraries (choose as needed)

- **CLI frameworks**: `spf13/cobra` (+ `pflag`) or `urfave/cli/v2`.
- **Config**: `spf13/viper` (be mindful of implicit env key mapping; document precedence) or simple bespoke loaders.
- **TUI**: `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`, `charmbracelet/glamour` (render Markdown).
- **Terminal**: `golang.org/x/term`, `github.com/mattn/go-isatty`, `github.com/mattn/go-colorable`.
- **Low-level TUI**: `tcell` and higher-level `tview` if not using Bubble Tea.

> Prefer small, well-maintained dependencies; audit licenses; pin versions in `go.mod` and run `go mod tidy`.

---

## 14) Example: Cobra skeleton with context, outputs, and completions

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
)

var (
    flagOutput  string
    flagColor   string // auto|always|never
    flagVerbose bool
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    root := &cobra.Command{
        Use:   "tool",
        Short: "A fast, accessible CLI",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // validate flags here, e.g., output format
            switch flagOutput {
            case "text", "json", "yaml":
            default:
                return fmt.Errorf("invalid --output %q (allowed: text,json,yaml)", flagOutput)
            }
            switch flagColor { case "auto", "always", "never": default: return fmt.Errorf("invalid --color %q", flagColor) }
            return nil
        },
    }

    root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "text", "Output format: text|json|yaml")
    root.PersistentFlags().StringVar(&flagColor, "color", "auto", "Color mode: auto|always|never")
    root.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose logs to stderr")

    root.AddCommand(newVersionCmd())
    root.AddCommand(newListCmd())

    // Completions
    root.AddCommand(&cobra.Command{
        Use:   "completion [bash|zsh|fish|powershell]",
        Short: "Generate shell completion scripts",
        Args:  cobra.ExactValidArgs(1),
        ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
        RunE: func(cmd *cobra.Command, args []string) error {
            // emit completion to stdout
            return genCompletion(cmd, args[0])
        },
    })

    if err := root.ExecuteContext(ctx); err != nil {
        fmt.Fprintln(os.Stderr, err)
        if errors.Is(err, context.Canceled) { os.Exit(130) }
        os.Exit(1)
    }
}
```

---

## 15) Tools & Workflow

- **Formatting & linting**: `gofmt`, `goimports`, `go vet`, `golangci-lint`, `staticcheck`.
- **Testing**: `go test -race`, golden tests; CI pipelines for multiple OSes.
- **Profiling**: `pprof` for hot paths; measure before optimizing.
- **Pre-commit**: enforce formatting and linting.
- **Dependencies**: minimal set, periodic updates, `go mod tidy`.

---

## 16) Common Pitfalls

- Coupling user-visible output with logs; always separate stdout/stderr.
- Interactive defaults in non-TTY contexts (breaks CI). Detect and disable or require `--yes`.
- Overuse of colors/emojis; fails on limited terminals and Windows without VT.
- Goroutine leaks from background progress/render loops; ensure exit on context cancel.
- Assuming ANSI/UTF‑8 everywhere; handle locales and width.
- Storing per-request state on long-lived clients (HTTP or DB); build per call instead.
- Forgetting to close files, responses; use `defer`.

---

### Appendix: Quick Checklist

- [ ] Clean `--help` with examples
- [ ] Stable `--output` formats and schemas
- [ ] Colors configurable; `NO_COLOR` respected
- [ ] TTY detection; CI-safe behavior
- [ ] Ctrl‑C cancels promptly; exit code 130
- [ ] Structured logs to stderr; quiet/verbose modes
- [ ] Shell completions and `--version` include metadata
- [ ] Cross-platform ANSI handling
- [ ] Tests: unit, golden, E2E (PTY for TUI)
- [ ] Reproducible builds, signed artifacts

