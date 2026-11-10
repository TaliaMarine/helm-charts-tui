package tui

import (
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// wheelStep is how many rows one mouse-wheel notch scrolls.
const wheelStep = 3

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncSizes()
		m.applyFilter()
		if m.detail != "" {
			m.renderDetail()
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

	case tea.MouseMsg:
		return m.updateMouse(msg)

	case spinner.TickMsg:
		if !m.busy() {
			m.spinnerActive = false // let the tick chain die; wake() restarts it
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case reposLoadedMsg:
		m.repos = []helm.Repo(msg)
		m.rebuildRepoTable()
		return m, m.finishLoad(ScreenRepoList, fmt.Sprintf("%d repositories", len(m.repos)))

	case chartCountsLoadedMsg:
		if msg != nil { // nil means the count query failed; keep what we had
			m.chartCounts = map[string]int(msg)
		}
		m.rebuildRepoTable()
		return m, nil

	case chartsLoadedMsg:
		m.charts = []helm.Chart(msg)
		m.rebuildChartTable()
		return m, m.finishLoad(ScreenChartList, fmt.Sprintf("%d charts", len(m.charts)))

	case versionsLoadedMsg:
		m.versions = []helm.Chart(msg)
		m.rebuildVersionTable()
		return m, m.finishLoad(ScreenChartVersions, fmt.Sprintf("%d versions", len(m.versions)))

	case chartDetailLoadedMsg:
		m.detail = string(msg)
		m.renderDetail()
		if m.screen == ScreenChartDetail && !m.refreshing {
			m.viewport.GotoTop()
		}
		return m, m.finishLoad(ScreenChartDetail, "chart details")

	case loadErrMsg:
		if !m.loading && !m.refreshing {
			return m, nil // stale: user already navigated away
		}
		wasRefreshing := m.refreshing
		m.loading = false
		m.refreshing = false
		m.silentReload = false
		if wasRefreshing { // keep the stale data on screen, report via toast
			return m, m.toast(firstLine(msg.err), true)
		}
		m.err = msg.err
		return m, nil

	case repoAddedMsg:
		m.mode = ModeNormal
		m.loading = false
		m.err = nil
		m.silentReload = true
		cmd := m.toast(fmt.Sprintf("repo %q added successfully", string(msg)), false)
		return m, tea.Batch(m.loadRepos(), m.loadChartCounts(), cmd)

	case repoAddErrMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case repoRemovedMsg:
		m.silentReload = true
		cmd := m.toast(fmt.Sprintf("repo %q removed", string(msg)), false)
		return m, tea.Batch(m.loadRepos(), m.loadChartCounts(), cmd)

	case repoRemoveErrMsg:
		m.refreshing = false
		return m, m.toast(fmt.Sprintf("removing %q: %s", msg.name, firstLine(msg.err)), true)

	case repoUpdatedMsg:
		return m.handleRepoUpdated(msg)

	case repoUpdateAllDoneMsg:
		m.syncing = false
		return m, tea.Batch(m.loadRepos(), m.loadChartCounts())

	case toastFadeMsg:
		if msg.seq == m.statusSeq && m.statusMsg != "" {
			m.statusDim = true
			return m, after(toastFade, clearStatusMsg{seq: msg.seq})
		}
		return m, nil

	case clearStatusMsg:
		if msg.seq == 0 || msg.seq == m.statusSeq {
			m.statusMsg = ""
			m.statusDim = false
			m.statusErr = false
		}
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
		case ModeConfirmRemove:
			return m.updateConfirmRemove(keyMsg)
		case ModeHelp:
			return m.updateHelp(keyMsg)
		case ModeNormal:
			return m.updateNormal(keyMsg)
		}
	}

	return m, nil
}

// finishLoad clears busy flags once data for screen s arrived. It returns a
// confirmation toast when the load was a user-requested reload.
func (m *Model) finishLoad(s Screen, what string) tea.Cmd {
	if m.screen != s {
		return nil // response for a screen the user already left
	}
	wasRefreshing := m.refreshing
	m.loading = false
	m.refreshing = false
	m.err = nil
	if m.silentReload {
		m.silentReload = false
		return nil
	}
	if wasRefreshing {
		return m.toast("reloaded "+what, false)
	}
	return nil
}

// toast shows a transient message that dims, then clears (3s total).
func (m *Model) toast(s string, isErr bool) tea.Cmd {
	m.statusSeq++
	m.statusMsg, m.statusErr, m.statusDim = s, isErr, false
	return after(toastVisible, toastFadeMsg{seq: m.statusSeq})
}

func firstLine(err error) string {
	s := err.Error()
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}

// markReload flags the current screen as reloading. Existing data stays
// visible (refreshing); otherwise the full-screen spinner is shown (loading).
func (m *Model) markReload(hasData bool) {
	if hasData && m.err == nil {
		m.refreshing = true
	} else {
		m.loading = true
	}
	m.err = nil
}

// reloadCurrent re-fetches the data behind the current screen.
func (m *Model) reloadCurrent() tea.Cmd {
	var cmds []tea.Cmd
	switch m.screen {
	case ScreenRepoList:
		m.markReload(len(m.repos) > 0)
		cmds = append(cmds, m.loadRepos(), m.loadChartCounts())
	case ScreenChartList:
		m.markReload(len(m.charts) > 0)
		cmds = append(cmds, m.loadCharts(m.selectedRepo.Name))
	case ScreenChartVersions:
		m.markReload(len(m.versions) > 0)
		cmds = append(cmds, m.loadVersions(m.selectedChart.Name))
	case ScreenChartDetail:
		m.markReload(m.detail != "")
		cmds = append(cmds, m.loadDetail(m.selectedVersion.Name, m.selectedVersion.Version))
	}
	cmds = append(cmds, m.wake())
	return tea.Batch(cmds...)
}

// startUpdate refreshes the named repos from their remotes, in parallel,
// reporting progress as each finishes.
func (m *Model) startUpdate(names []string) tea.Cmd {
	if m.updating {
		return m.toast("an update is already running", true)
	}
	if len(names) == 0 {
		return m.toast("no repositories to update", true)
	}
	m.updating = true
	m.updateTotal = len(names)
	m.updateDone = 0
	m.updateFailed = nil
	m.updateErr = nil
	cmds := make([]tea.Cmd, 0, len(names)+2)
	cmds = append(cmds, m.progress.SetPercent(0), m.wake())
	for _, n := range names {
		cmds = append(cmds, m.updateRepo(n))
	}
	return tea.Batch(cmds...)
}

func (m Model) handleRepoUpdated(msg repoUpdatedMsg) (tea.Model, tea.Cmd) {
	if !m.updating {
		return m, nil
	}
	m.updateDone++
	if msg.err != nil {
		m.updateFailed = append(m.updateFailed, msg.name)
		m.updateErr = msg.err
	}
	bar := m.progress.SetPercent(float64(m.updateDone) / float64(m.updateTotal))
	if m.updateDone < m.updateTotal {
		return m, bar
	}

	m.updating = false
	ok := m.updateTotal - len(m.updateFailed)
	var summary string
	switch {
	case len(m.updateFailed) == 0 && m.updateTotal == 1:
		summary = "updated " + msg.name
	case len(m.updateFailed) == 0:
		summary = fmt.Sprintf("updated %d repositories", ok)
	case m.updateTotal == 1:
		summary = fmt.Sprintf("updating %s: %s", msg.name, firstLine(m.updateErr))
	default:
		summary = fmt.Sprintf("updated %d/%d repositories; failed: %s", ok, m.updateTotal, strings.Join(m.updateFailed, ", "))
	}

	m.silentReload = true
	cmds := []tea.Cmd{bar, m.toast(summary, len(m.updateFailed) > 0), m.reloadCurrent()}
	if m.screen != ScreenRepoList { // keep repo list + counts fresh for when we go back
		cmds = append(cmds, m.loadRepos(), m.loadChartCounts())
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.mode != ModeNormal {
		return m, nil
	}
	up := msg.Button == tea.MouseButtonWheelUp
	if !up && msg.Button != tea.MouseButtonWheelDown {
		return m, nil
	}
	switch m.screen {
	case ScreenRepoList:
		scrollTable(&m.repoTable, up)
	case ScreenChartList:
		scrollTable(&m.chartTable, up)
	case ScreenChartVersions:
		scrollTable(&m.versionTable, up)
	case ScreenChartDetail:
		if up {
			m.viewport.ScrollUp(wheelStep)
		} else {
			m.viewport.ScrollDown(wheelStep)
		}
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	onRepos := m.screen == ScreenRepoList
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

	case key.Matches(msg, m.keys.Back):
		if !onRepos {
			return m.handleEscape()
		}
		return m, nil

	case key.Matches(msg, m.keys.Open):
		return m.handleEnter()

	case key.Matches(msg, m.keys.Reload):
		return m, m.reloadCurrent()

	case key.Matches(msg, m.keys.UpdateAll):
		if onRepos {
			names := make([]string, 0, len(m.repos))
			for _, r := range m.repos {
				names = append(names, r.Name)
			}
			return m, m.startUpdate(names)
		}
		return m, m.startUpdate([]string{m.selectedRepo.Name})

	case onRepos && key.Matches(msg, m.keys.UpdateOne):
		if r, ok := m.highlightedRepo(); ok {
			return m, m.startUpdate([]string{r.Name})
		}
		return m, m.toast("highlight a repository first", true)

	case onRepos && key.Matches(msg, m.keys.AddRepo):
		return m.openAddRepo()

	case onRepos && key.Matches(msg, m.keys.Remove):
		if r, ok := m.highlightedRepo(); ok {
			m.pendingRemove = r.Name
			m.mode = ModeConfirmRemove
			return m, nil
		}
		return m, m.toast("highlight a repository first", true)

	case key.Matches(msg, m.keys.Copy):
		return m.copySelection(false)

	case key.Matches(msg, m.keys.CopyCmd):
		return m.copySelection(true)

	case key.Matches(msg, m.keys.Sort):
		return m.cycleSort()

	case m.screen == ScreenChartDetail && key.Matches(msg, m.keys.Home):
		m.viewport.GotoTop()
		return m, nil

	case m.screen == ScreenChartDetail && key.Matches(msg, m.keys.End):
		m.viewport.GotoBottom()
		return m, nil
	}

	// Delegate navigation to active component
	var cmd tea.Cmd
	switch m.screen {
	case ScreenRepoList:
		m.repoTable, cmd = m.repoTable.Update(msg)
	case ScreenChartList:
		m.chartTable, cmd = m.chartTable.Update(msg)
	case ScreenChartVersions:
		m.versionTable, cmd = m.versionTable.Update(msg)
	case ScreenChartDetail:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// highlightedRepo returns the repo under the cursor (not the "Add new repo" row).
func (m Model) highlightedRepo() (helm.Repo, bool) {
	idx := m.repoTable.Cursor() - 1 // row 0 is the synthetic add row
	if idx < 0 || idx >= len(m.filteredRepos) {
		return helm.Repo{}, false
	}
	return m.filteredRepos[idx], true
}

func (m Model) openAddRepo() (tea.Model, tea.Cmd) {
	m.mode = ModeAddRepo
	m.addRepoName.SetValue("")
	m.addRepoURL.SetValue("")
	m.addRepoFocus = 0
	m.addRepoName.Focus()
	m.addRepoURL.Blur()
	m.err = nil
	return m, textinput.Blink
}

// highlightedChart returns the chart/version row under the cursor.
func (m Model) highlightedChart() (helm.Chart, bool) {
	var rows []helm.Chart
	var idx int
	switch m.screen {
	case ScreenChartList:
		rows, idx = m.filteredCharts, m.chartTable.Cursor()
	case ScreenChartVersions:
		rows, idx = m.filteredVersions, m.versionTable.Cursor()
	case ScreenChartDetail:
		return m.selectedVersion, m.selectedVersion.Name != ""
	}
	if idx < 0 || idx >= len(rows) {
		return helm.Chart{}, false
	}
	return rows[idx], true
}

// copySelection copies a name/URL/YAML, or with asCmd a ready-to-run helm command.
func (m Model) copySelection(asCmd bool) (tea.Model, tea.Cmd) {
	var text, what string
	if m.screen == ScreenRepoList {
		r, ok := m.highlightedRepo()
		if !ok {
			return m, m.toast("nothing to copy", true)
		}
		text, what = r.URL, "repo URL"
		if asCmd {
			text, what = fmt.Sprintf("helm repo add %s %s", r.Name, r.URL), "helm repo add command"
		}
	} else {
		c, ok := m.highlightedChart()
		if !ok {
			return m, m.toast("nothing to copy", true)
		}
		switch {
		case asCmd:
			text = fmt.Sprintf("helm install %s %s --version %s", path.Base(c.Name), c.Name, c.Version)
			what = "helm install command"
		case m.screen == ScreenChartDetail:
			text, what = m.detail, "chart YAML"
		default:
			text, what = c.Name, "chart name"
		}
	}
	if err := writeClipboard(text); err != nil {
		return m, m.toast("clipboard unavailable: "+firstLine(err), true)
	}
	return m, m.toast("copied "+what, false)
}

func (m Model) cycleSort() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenRepoList:
		m.repoSort = (m.repoSort + 1) % 4
		m.rebuildRepoTable()
		m.repoTable.GotoTop()
		return m, m.toast(sortToast(m.repoSort), false)
	case ScreenChartList:
		m.chartSort = (m.chartSort + 1) % 3
		m.rebuildChartTable()
		m.chartTable.GotoTop()
		return m, m.toast(sortToast(m.chartSort), false)
	}
	return m, nil
}

func sortToast(mode int) string {
	if mode == sortDefault {
		return "sort: default order"
	}
	return "sort: " + sortLabel(mode)
}

func (m Model) handleEscape() (tea.Model, tea.Cmd) {
	// If filter is active, clear it first
	if m.filterText != "" {
		m.filterText = ""
		m.applyFilter()
		return m, nil
	}

	// Leaving a screen abandons whatever it was waiting for.
	switch m.screen {
	case ScreenRepoList:
		m.mode = ModeConfirmExit
		return m, nil
	case ScreenChartList:
		m.screen = ScreenRepoList
		m.charts = nil
		m.filteredCharts = nil
		m.chartSort = sortDefault
		m.leave()
		m.rebuildRepoTable()
	case ScreenChartVersions:
		m.screen = ScreenChartList
		m.versions = nil
		m.filteredVersions = nil
		m.leave()
		m.rebuildChartTable()
	case ScreenChartDetail:
		m.screen = ScreenChartVersions
		m.detail = ""
		m.leave()
		m.rebuildVersionTable()
	}
	return m, nil
}

// leave resets transient state shared by every back navigation.
func (m *Model) leave() {
	m.filterText = ""
	m.err = nil
	m.loading = false
	m.refreshing = false
	m.silentReload = false
}

// enter resets state shared by every forward navigation and starts the spinner.
func (m *Model) enter(s Screen) tea.Cmd {
	m.screen = s
	m.loading = true
	m.refreshing = false
	m.filterText = ""
	m.err = nil
	return m.wake()
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenRepoList:
		idx := m.repoTable.Cursor()
		if idx == 0 {
			return m.openAddRepo() // "Add new repo" row
		}
		// offset by 1 for the "Add new repo" row
		dataIdx := idx - 1
		if dataIdx >= 0 && dataIdx < len(m.filteredRepos) {
			repo := m.filteredRepos[dataIdx]
			m.selectedRepo = repo
			wake := m.enter(ScreenChartList)
			return m, tea.Batch(m.loadCharts(repo.Name), wake)
		}

	case ScreenChartList:
		idx := m.chartTable.Cursor()
		if idx >= 0 && idx < len(m.filteredCharts) {
			m.selectedChart = m.filteredCharts[idx]
			wake := m.enter(ScreenChartVersions)
			return m, tea.Batch(m.loadVersions(m.selectedChart.Name), wake)
		}

	case ScreenChartVersions:
		idx := m.versionTable.Cursor()
		if idx >= 0 && idx < len(m.filteredVersions) {
			ver := m.filteredVersions[idx]
			m.selectedVersion = ver
			wake := m.enter(ScreenChartDetail)
			return m, tea.Batch(m.loadDetail(ver.Name, ver.Version), wake)
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
	// Keep the cursor on a real row as the list shrinks.
	m.clampCursor()
	return m, cmd
}

func (m *Model) clampCursor() {
	switch m.screen {
	case ScreenRepoList:
		m.repoTable.SetCursor(m.repoTable.Cursor())
	case ScreenChartList:
		m.chartTable.SetCursor(m.chartTable.Cursor())
	case ScreenChartVersions:
		m.versionTable.SetCursor(m.versionTable.Cursor())
	}
}

func (m Model) updateAddRepo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeNormal
		m.err = nil
		m.loading = false
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
		return m, tea.Batch(m.addRepo(name, url), m.wake())
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

func (m Model) updateConfirmRemove(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.mode = ModeNormal
		m.refreshing = true
		return m, tea.Batch(m.removeRepo(m.pendingRemove), m.wake())
	case "n", "N", "esc":
		m.mode = ModeNormal
		m.pendingRemove = ""
		return m, nil
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key dismisses help
	m.mode = ModeNormal
	return m, nil
}

func scrollTable(t interface {
	MoveUp(int)
	MoveDown(int)
}, up bool) {
	if up {
		t.MoveUp(wheelStep)
	} else {
		t.MoveDown(wheelStep)
	}
}
