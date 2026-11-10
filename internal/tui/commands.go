package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) loadRepos() tea.Cmd {
	return func() tea.Msg {
		repos, err := m.helm.RepoList(m.ctx)
		if err != nil {
			return loadErrMsg{context: "loading repos", err: err}
		}
		return reposLoadedMsg(repos)
	}
}

func (m Model) loadChartCounts() tea.Cmd {
	return func() tea.Msg {
		charts, err := m.helm.SearchRepo(m.ctx, "")
		if err != nil {
			// Non-fatal: counts keep their previous value (or "..." initially)
			return chartCountsLoadedMsg(nil)
		}
		counts := make(map[string]int)
		for _, c := range charts {
			parts := strings.SplitN(c.Name, "/", 2)
			if len(parts) == 2 {
				counts[parts[0]]++
			}
		}
		return chartCountsLoadedMsg(counts)
	}
}

func (m Model) loadCharts(repoName string) tea.Cmd {
	return func() tea.Msg {
		charts, err := m.helm.SearchRepo(m.ctx, repoName+"/")
		if err != nil {
			return loadErrMsg{context: fmt.Sprintf("loading charts for %s", repoName), err: err}
		}
		return chartsLoadedMsg(charts)
	}
}

func (m Model) loadVersions(chartName string) tea.Cmd {
	return func() tea.Msg {
		charts, err := m.helm.SearchRepoVersions(m.ctx, chartName)
		if err != nil {
			return loadErrMsg{context: fmt.Sprintf("loading versions for %s", chartName), err: err}
		}
		return versionsLoadedMsg(charts)
	}
}

func (m Model) loadDetail(chartName, version string) tea.Cmd {
	return func() tea.Msg {
		detail, err := m.helm.ShowChart(m.ctx, chartName, version)
		if err != nil {
			return loadErrMsg{context: fmt.Sprintf("loading chart %s@%s", chartName, version), err: err}
		}
		return chartDetailLoadedMsg(detail)
	}
}

// repoUpdateAll is the silent startup sync.
func (m Model) repoUpdateAll() tea.Cmd {
	return func() tea.Msg {
		_ = m.helm.RepoUpdateAll(m.ctx)
		return repoUpdateAllDoneMsg{}
	}
}

// updateRepo refreshes a single repo from its remote. Concurrency is bounded
// so refreshing every repo doesn't spawn an unbounded number of helm processes.
func (m Model) updateRepo(name string) tea.Cmd {
	return func() tea.Msg {
		select {
		case m.updateSem <- struct{}{}:
			defer func() { <-m.updateSem }()
		case <-m.ctx.Done():
			return repoUpdatedMsg{name: name, err: m.ctx.Err()}
		}
		return repoUpdatedMsg{name: name, err: m.helm.RepoUpdate(m.ctx, name)}
	}
}

func (m Model) addRepo(name, url string) tea.Cmd {
	return func() tea.Msg {
		if err := m.helm.RepoAdd(m.ctx, name, url); err != nil {
			return repoAddErrMsg{err: fmt.Errorf("adding repo %q: %w", name, err)}
		}
		if err := m.helm.RepoUpdate(m.ctx, name); err != nil {
			return repoAddErrMsg{err: fmt.Errorf("updating repo %q: %w", name, err)}
		}
		return repoAddedMsg(name)
	}
}

func (m Model) removeRepo(name string) tea.Cmd {
	return func() tea.Msg {
		if err := m.helm.RepoRemove(m.ctx, name); err != nil {
			return repoRemoveErrMsg{name: name, err: err}
		}
		return repoRemovedMsg(name)
	}
}

func after(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

// clearStatusAfter schedules an unconditional status clear.
func clearStatusAfter(d time.Duration) tea.Cmd {
	return after(d, clearStatusMsg{})
}
