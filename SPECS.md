# Helm Charts TUI -- Specification

Module: `github.com/TaliaMarine/helm-charts-tui`
Go version: 1.26+

---

## 1. Navigation Flow

The app has four screens arranged as a drill-down hierarchy:

```
Repo List -> Chart List -> Chart Versions -> Chart Detail
```

`Enter` drills in. `Esc` drills out (with special behavior on the top screen and when a filter is active -- see sections 3 and 4).

### 1.1 Repo List (startup screen)

- Data source: `helm repo list --output json`
- Columns: **Name**, **URL**, **Charts** (count)
- Chart counts are loaded via a single `helm search repo "" --output json` call; results are grouped by repo name prefix (`reponame/chartname`).
- First row is a synthetic **"+ Add new repo"** entry (see section 5).

### 1.2 Chart List

- Data source: `helm search repo REPONAME/ --output json`
- Header: repo name and URL
- Columns: **Name**, **Version** (latest), **App Version**, **Description**

### 1.3 Chart Versions

- Data source: `helm search repo CHARTNAME --versions --output json`
- Header: repo name > chart name
- Columns: **Name**, **Version**, **App Version**, **Description**

### 1.4 Chart Detail

- Data source: `helm show chart CHARTNAME --version VERSION`
- Header: repo name > chart name > version
- Displayed as **scrollable raw text** (YAML output) in a viewport.

---

## 2. Key Bindings

### Global (always active)

| Key      | Action              | Notes                                    |
|----------|---------------------|------------------------------------------|
| `Ctrl+C` | Quit immediately   | Works in every mode, including text input |

### Normal mode (not typing in an input field)

| Key           | Action             |
|---------------|---------------------|
| `q`           | Quit immediately    |
| `Enter`       | Select / drill in   |
| `Esc`         | Back / clear filter / confirm-exit (see section 3) |
| `/`           | Activate filter (list screens only) |
| `?`           | Show help overlay   |
| `j` / `Down`  | Move cursor down   |
| `k` / `Up`    | Move cursor up     |
| `PgDn`        | Page down          |
| `PgUp`        | Page up            |
| `g` / `Home`  | Go to top          |
| `G` / `End`   | Go to bottom       |

### Filter mode (typing in the filter input)

| Key     | Action                                      |
|---------|----------------------------------------------|
| Any key | Types into filter; results update live        |
| `Enter` | Confirm filter, return to normal mode         |
| `Esc`   | Cancel filter, clear text, return to normal   |

`q` types the letter `q` (does **not** quit) while in filter mode.

### Add-repo dialog

| Key         | Action                        |
|-------------|-------------------------------|
| `Tab` / `Shift+Tab` | Switch between Name and URL fields |
| `Enter`     | Submit (validates non-empty)  |
| `Esc`       | Cancel, return to normal mode |

### Help overlay

Any key dismisses the overlay.

### Confirm-exit dialog (repo list `Esc` only)

| Key           | Action |
|---------------|--------|
| `y` / `Enter` | Quit  |
| `n` / `Esc`   | Cancel |

---

## 3. Esc Behavior (detailed)

`Esc` has layered behavior depending on context:

1. **Filter mode active** -- cancel filter, clear text, return to normal mode.
2. **Confirmed filter text exists** (normal mode) -- clear the filter (restore full list). Does **not** navigate back.
3. **No filter, on Chart List / Chart Versions / Chart Detail** -- navigate back one screen.
4. **No filter, on Repo List** -- open the confirm-exit dialog ("Exit? y/n").

This means going back from a filtered list always takes two `Esc` presses: first clears the filter, second navigates.

---

## 4. Filtering

- Available on the three list screens (Repo List, Chart List, Chart Versions). Not on Chart Detail (which scrolls instead).
- Activated by pressing `/` in normal mode. A text input appears at the bottom.
- **Live filtering**: the list updates as the user types (case-insensitive substring match across all visible columns).
- `Enter` confirms the filter and returns to normal mode. The filter text stays active and is shown in the status bar.
- `Esc` cancels and clears the filter.
- Filter text is cleared automatically when navigating to a different screen (both forward and back).

---

## 5. Add Repo Flow

- Triggered by pressing `Enter` on the "Add new repo" row (always row 0 in the repo list).
- A centered dialog appears with two text inputs: **Name** and **URL**.
- `Tab` / `Shift+Tab` switches focus between the two inputs.
- `Enter` submits. Validation: both fields must be non-empty.
- On submit, runs `helm repo add NAME URL` then `helm repo update NAME`.
- On success: dialog closes, a transient status message appears (clears after 3s), repo list and chart counts reload.
- On error: error message is shown in the dialog.
- `Esc` cancels and closes the dialog without making changes.

---

## 6. Helm CLI Commands Used

All commands are executed via `exec.CommandContext` with `context.Context` for cancellation. Arguments are always passed as separate args (never shell string concatenation).

| Operation       | Command                                            |
|-----------------|----------------------------------------------------|
| List repos      | `helm repo list --output json`                     |
| Add repo        | `helm repo add NAME URL`                           |
| Update repo     | `helm repo update NAME`                            |
| Search charts   | `helm search repo KEYWORD --output json`           |
| Search versions | `helm search repo KEYWORD --versions --output json`|
| Show chart      | `helm show chart CHARTNAME --version VERSION`      |

---

## 7. Build & Distribution

- Installable via `go install github.com/TaliaMarine/helm-charts-tui@latest`
- Buildable via `go build .` (or `just build` which outputs to `bin/`)
- `--version` / `-v` flag prints version, git commit (short), and build date
- Version info embedded via `-ldflags` at build time; falls back to `debug.ReadBuildInfo()` VCS data

---

## 8. Implementation Constraints

- **Go 1.26+**
- **Bubble Tea** (`charmbracelet/bubbletea`) with a single `Model` and `Screen` enum -- not nested models
- **`helm.Executor` interface** abstracts all CLI calls; `MockExecutor` (function fields) enables unit tests without a real helm binary
- **`NO_COLOR` env var** disables all ANSI color output
- **Justfile** with recipes: `build`, `build-release`, `test`, `test-verbose`, `test-coverage`, `vet`, `lint`, `tidy`, `fmt`, `clean`, `check` (all), `run`
- **Unit tests**: table-driven for JSON parsing and filter logic; message-driven state transition tests for TUI using `MockExecutor`
- **Exit codes**: `0` success, `1` error, `130` interrupted (SIGINT/SIGTERM)
