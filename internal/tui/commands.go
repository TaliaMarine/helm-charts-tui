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
			// Non-fatal: counts will just show "..."
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

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}
