# helm-charts-tui

A terminal user interface for browsing Helm chart repositories interactively.

## Features

- Browse configured Helm repos with chart counts
- Navigate charts, versions, and full chart metadata
- Filter any list with `/` (like `less`)
- Add new Helm repositories from within the TUI
- Keyboard-driven navigation with vim-style keybindings
- One-key reload (`r`) and remote update (`R` / `u`) of one or all repositories, with live progress
- Animated spinners, a progress bar for multi-repo updates, and fading toast notifications
- Sort, copy to clipboard (name, URL, YAML, `helm install` command), remove repos, mouse-wheel scrolling

## Prerequisites

- [Helm CLI](https://helm.sh/docs/intro/install/) v3+ installed and on `$PATH`
- At least one Helm repo configured (e.g., `helm repo add bitnami https://charts.bitnami.com/bitnami`)

## Installation

### go install

```sh
go install github.com/TaliaMarine/helm-charts-tui@latest
```

### From source

```sh
git clone https://github.com/TaliaMarine/helm-charts-tui.git
cd helm-charts-tui
just build
./bin/helm-charts-tui
```

## Usage

```sh
helm-charts-tui           # Launch the TUI
helm-charts-tui --version # Print version information
```

## Screens

1. **Repository List** -- All configured Helm repos with chart counts. Select a repo to browse its charts, or select "Add new repo" to add one interactively.
2. **Chart List** -- Charts within a selected repo with name, version, app version, and description.
3. **Chart Versions** -- All published versions of a selected chart.
4. **Chart Detail** -- Full chart metadata (YAML) in a scrollable viewport.

## Key Bindings

Press `?` in the app for the full overlay.

| Key | Action |
|-----|--------|
| `Enter` / `l` / `→` | Open selected item |
| `Esc` | Clear filter, then go back (confirm-exit on the repo list) |
| `h` / `←` / `Backspace` | Go back |
| `j` `k` / `↑` `↓` | Move down / up |
| `PgUp` `PgDn` / `Ctrl+B` `Ctrl+F` | Page up / down |
| `Ctrl+U` `Ctrl+D` | Half page up / down |
| `g` `G` / `Home` `End` | Top / bottom |
| `/` | Filter the current list |
| `s` | Cycle sort order (repo list, chart list) |
| `r` / `F5` | Reload the current view from the local cache |
| `R` | Update from remote: all repos (repo list) or the current repo (chart screens) |
| `u` | Update the highlighted repo from remote (repo list) |
| `a` | Add a repository |
| `x` / `Delete` | Remove the highlighted repository (asks first) |
| `y` | Copy repo URL / chart name / chart YAML (detail screen) |
| `Y` | Copy `helm repo add` / `helm install` command |
| `?` | Help overlay |
| `q` / `Ctrl+C` | Quit |

Mouse wheel scrolls lists and the detail view. (Hold `Shift` to select text in most terminals.)

## Refreshing

- `r` re-reads the current screen. Existing rows stay visible while a spinner runs in the header; a failed reload leaves the old data in place and reports the error in a toast.
- `R` / `u` run `helm repo update` against the remote. Updating several repos runs up to four in parallel and shows an `updating 3/6` progress bar; the view reloads when done and any failures are listed.
- Repos are also synced once in the background at startup.

## Environment Variables

| Variable   | Effect                  |
|------------|------------------------|
| `NO_COLOR` | Disables color output  |

## Development

See [DEVELOPERS.md](DEVELOPERS.md) for architecture details, building, and testing.

## License

MIT
