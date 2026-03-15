package tui

import (
	"time"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// reposLoadedMsg is sent when helm repos are loaded.
type reposLoadedMsg []helm.Repo

// chartCountsLoadedMsg carries per-repo chart counts.
type chartCountsLoadedMsg map[string]int

// chartsLoadedMsg is sent when charts for a repo are loaded.
type chartsLoadedMsg []helm.Chart

// versionsLoadedMsg is sent when chart versions are loaded.
type versionsLoadedMsg []helm.Chart

// chartDetailLoadedMsg carries the YAML output from helm show chart.
type chartDetailLoadedMsg string

// loadErrMsg reports an error from an async operation.
type loadErrMsg struct {
	context string
	err     error
}

// repoAddedMsg signals a repo was added successfully.
type repoAddedMsg string

// repoAddErrMsg reports an error when adding a repo.
type repoAddErrMsg struct {
	err error
}

// clearStatusMsg clears the transient status message.
type clearStatusMsg struct{}

// tickMsg is used for timed events.
type tickMsg time.Time
