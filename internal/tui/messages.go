package tui

import (
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

// repoRemovedMsg signals a repo was removed successfully.
type repoRemovedMsg string

// repoRemoveErrMsg reports an error when removing a repo.
type repoRemoveErrMsg struct {
	name string
	err  error
}

// repoUpdatedMsg reports the outcome of updating one repo (user-triggered refresh).
type repoUpdatedMsg struct {
	name string
	err  error
}

// repoUpdateAllDoneMsg signals that the startup background repo update completed.
type repoUpdateAllDoneMsg struct{}

// toastFadeMsg dims the toast shortly before it disappears.
type toastFadeMsg struct{ seq int }

// clearStatusMsg clears the transient status message. A zero seq clears
// unconditionally; otherwise it only clears the toast that scheduled it.
type clearStatusMsg struct{ seq int }
