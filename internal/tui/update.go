package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncSizes()
		if m.screen == ScreenChartDetail && m.detail != "" {
			m.viewport.SetContent(m.detail)
		}
		return m, nil

	case tea.KeyMsg:
		// ctrl+c always quits regardless of mode
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}
		// q quits only in normal/help mode (not when typing)
		if (m.mode == ModeNormal || m.mode == ModeHelp) && key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

	case reposLoadedMsg:
		m.repos = []helm.Repo(msg)
		m.loading = false
		m.err = nil
		m.rebuildRepoTable()
		return m, nil

	case chartCountsLoadedMsg:
		m.chartCounts = map[string]int(msg)
		m.rebuildRepoTable()
		return m, nil

	case chartsLoadedMsg:
		m.charts = []helm.Chart(msg)
		m.loading = false
		m.err = nil
		m.rebuildChartTable()
		return m, nil

	case versionsLoadedMsg:
		m.versions = []helm.Chart(msg)
		m.loading = false
		m.err = nil
		m.rebuildVersionTable()
		return m, nil

	case chartDetailLoadedMsg:
		m.detail = string(msg)
		m.loading = false
		m.err = nil
		m.viewport.SetContent(m.detail)
		m.viewport.GotoTop()
		return m, nil

	case loadErrMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case repoAddedMsg:
		m.mode = ModeNormal
		m.loading = false
		m.err = nil
		m.statusMsg = fmt.Sprintf("repo %q added successfully", string(msg))
		return m, tea.Batch(m.loadRepos(), m.loadChartCounts(), clearStatusAfter(3*time.Second))

	case repoAddErrMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil
	}

	// Key message dispatch by mode
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch m.mode {
		case ModeFilter:
			return m.updateFilter(keyMsg)
		case ModeAddRepo:
			return m.updateAddRepo(keyMsg)
		case ModeConfirmExit:
			return m.updateConfirmExit(keyMsg)
		case ModeHelp:
			return m.updateHelp(keyMsg)
		case ModeNormal:
			return m.updateNormal(keyMsg)
		}
	}

	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Filter):
		if m.screen != ScreenChartDetail {
			m.mode = ModeFilter
			m.filterInput.SetValue("")
			m.filterInput.Focus()
			return m, textinput.Blink
		}

	case key.Matches(msg, m.keys.Help):
		m.mode = ModeHelp
		return m, nil

	case key.Matches(msg, m.keys.Escape):
		return m.handleEscape()

	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()
	}

	// Delegate navigation to active component
	switch m.screen {
	case ScreenRepoList:
		var cmd tea.Cmd
		m.repoTable, cmd = m.repoTable.Update(msg)
		return m, cmd
	case ScreenChartList:
		var cmd tea.Cmd
		m.chartTable, cmd = m.chartTable.Update(msg)
		return m, cmd
	case ScreenChartVersions:
		var cmd tea.Cmd
		m.versionTable, cmd = m.versionTable.Update(msg)
		return m, cmd
	case ScreenChartDetail:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleEscape() (tea.Model, tea.Cmd) {
	// If filter is active, clear it first
	if m.filterText != "" {
		m.filterText = ""
		m.applyFilter()
		return m, nil
	}

	switch m.screen {
	case ScreenRepoList:
		m.mode = ModeConfirmExit
		return m, nil
	case ScreenChartList:
		m.screen = ScreenRepoList
		m.charts = nil
		m.filteredCharts = nil
		m.filterText = ""
		m.err = nil
		return m, nil
	case ScreenChartVersions:
		m.screen = ScreenChartList
		m.versions = nil
		m.filteredVersions = nil
		m.filterText = ""
		m.err = nil
		return m, nil
	case ScreenChartDetail:
		m.screen = ScreenChartVersions
		m.detail = ""
		m.err = nil
		return m, nil
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenRepoList:
		idx := m.repoTable.Cursor()
		if idx == 0 {
			// "Add new repo" row
			m.mode = ModeAddRepo
			m.addRepoName.SetValue("")
			m.addRepoURL.SetValue("")
			m.addRepoFocus = 0
			m.addRepoName.Focus()
			m.addRepoURL.Blur()
			m.err = nil
			return m, textinput.Blink
		}
		// offset by 1 for the "Add new repo" row
		dataIdx := idx - 1
		if dataIdx >= 0 && dataIdx < len(m.filteredRepos) {
			repo := m.filteredRepos[dataIdx]
			m.selectedRepo = repo
			m.screen = ScreenChartList
			m.loading = true
			m.filterText = ""
			m.err = nil
			return m, m.loadCharts(repo.Name)
		}

	case ScreenChartList:
		idx := m.chartTable.Cursor()
		if idx >= 0 && idx < len(m.filteredCharts) {
			chart := m.filteredCharts[idx]
			m.selectedChart = chart
			m.screen = ScreenChartVersions
			m.loading = true
			m.filterText = ""
			m.err = nil
			return m, m.loadVersions(chart.Name)
		}

	case ScreenChartVersions:
		idx := m.versionTable.Cursor()
		if idx >= 0 && idx < len(m.filteredVersions) {
			ver := m.filteredVersions[idx]
			m.selectedVersion = ver
			m.screen = ScreenChartDetail
			m.loading = true
			m.err = nil
			return m, m.loadDetail(ver.Name, ver.Version)
		}
	}

	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		m.filterInput.Blur()
		// Cancel filter: restore unfiltered state
		m.filterText = ""
		m.applyFilter()
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		m.mode = ModeNormal
		m.filterInput.Blur()
		m.filterText = m.filterInput.Value()
		m.applyFilter()
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	// Live filtering as user types
	m.filterText = m.filterInput.Value()
	m.applyFilter()
	return m, cmd
}

func (m Model) updateAddRepo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		m.err = nil
		return m, nil

	case key.Matches(msg, m.keys.Tab), key.Matches(msg, m.keys.ShiftTab):
		if m.addRepoFocus == 0 {
			m.addRepoName.Blur()
			m.addRepoURL.Focus()
			m.addRepoFocus = 1
		} else {
			m.addRepoURL.Blur()
			m.addRepoName.Focus()
			m.addRepoFocus = 0
		}
		return m, textinput.Blink

	case key.Matches(msg, m.keys.Enter):
		name := strings.TrimSpace(m.addRepoName.Value())
		url := strings.TrimSpace(m.addRepoURL.Value())
		if name == "" || url == "" {
			m.err = fmt.Errorf("both name and URL are required")
			return m, nil
		}
		m.loading = true
		m.err = nil
		return m, m.addRepo(name, url)
	}

	// Delegate to focused input
	var cmd tea.Cmd
	if m.addRepoFocus == 0 {
		m.addRepoName, cmd = m.addRepoName.Update(msg)
	} else {
		m.addRepoURL, cmd = m.addRepoURL.Update(msg)
	}
	return m, cmd
}

func (m Model) updateConfirmExit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		return m, tea.Quit
	case "n", "N", "esc":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key dismisses help
	m.mode = ModeNormal
	return m, nil
}
