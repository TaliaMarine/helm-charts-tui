# helm-charts-tui

A terminal user interface for browsing Helm chart repositories interactively.

## Features

- Browse configured Helm repos with chart counts
- Navigate charts, versions, and full chart metadata
- Filter any list with `/` (like `less`)
- Add new Helm repositories from within the TUI
- Keyboard-driven navigation with vim-style keybindings

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

| Key         | Action                                     |
|-------------|--------------------------------------------|
| `Enter`     | Select item / confirm                      |
| `Esc`       | Go back / clear filter / exit confirmation |
| `q`         | Quit immediately                           |
| `Ctrl+C`    | Quit immediately                           |
| `/`         | Filter current list                        |
| `?`         | Show help overlay                          |
| `j` / `k`   | Move down / up                             |
| `Up` / `Down`| Move up / down                            |
| `PgUp` / `PgDn` | Page up / down                        |
| `Home` / `g`| Go to top                                  |
| `End` / `G` | Go to bottom                               |
| `Tab`       | Next field (in add-repo dialog)            |

## Environment Variables

| Variable   | Effect                  |
|------------|------------------------|
| `NO_COLOR` | Disables color output  |

## Development

See [DEVELOPERS.md](DEVELOPERS.md) for architecture details, building, and testing.

## License

MIT
