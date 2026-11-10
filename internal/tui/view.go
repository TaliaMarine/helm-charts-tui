package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// View renders the current screen.
func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderContent())
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	view := b.String()

	switch m.mode {
	case ModeAddRepo:
		view = overlay(view, m.renderAddRepoDialog(), m.width, m.height)
	case ModeConfirmExit:
		view = overlay(view, m.renderConfirmExitDialog(), m.width, m.height)
	case ModeConfirmRemove:
		view = overlay(view, m.renderConfirmRemoveDialog(), m.width, m.height)
	case ModeHelp:
		view = overlay(view, m.renderHelpOverlay(), m.width, m.height)
	}
	return view
}

// renderContent draws the body: spinner, error, empty state or the screen's list.
func (m Model) renderContent() string {
	h := m.contentHeight()
	center := func(s string) string {
		return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, s)
	}

	switch {
	case m.loading && m.mode != ModeAddRepo:
		return center(m.spinner.View() + loadingStyle.Render(" "+m.loadingLabel()+"…"))
	case m.err != nil && m.mode != ModeAddRepo:
		return center(errorStyle.Render("Error: "+firstLine(m.err)) + "\n\n" +
			dimStyle.Render("r reload   esc back"))
	}

	body := ""
	switch m.screen {
	case ScreenRepoList:
		body = m.repoTable.View()
	case ScreenChartList:
		body = m.emptyOr(len(m.filteredCharts), len(m.charts), m.chartTable.View(), "No charts in this repository.", h)
	case ScreenChartVersions:
		body = m.emptyOr(len(m.filteredVersions), len(m.versions), m.versionTable.View(), "No versions found.", h)
	case ScreenChartDetail:
		body = m.viewport.View()
	}
	return body
}

// emptyOr swaps the table for a hint when there is nothing to list.
func (m Model) emptyOr(shown, total int, tableView, noData string, h int) string {
	if shown > 0 {
		return tableView
	}
	msg := noData + "\n" + dimStyle.Render("r reload   R update repo from remote")
	if total > 0 {
		msg = fmt.Sprintf("No matches for %q.\n%s", m.filterText, dimStyle.Render("esc clears the filter"))
	}
	return lipgloss.Place(m.width, h, lipgloss.Center, lipgloss.Center, msg)
}

func (m Model) loadingLabel() string {
	switch m.screen {
	case ScreenRepoList:
		return "Loading repositories"
	case ScreenChartList:
		return "Loading charts for " + m.selectedRepo.Name
	case ScreenChartVersions:
		return "Loading versions of " + m.selectedChart.Name
	default:
		return "Loading chart details"
	}
}

// crumb renders breadcrumb parts, emphasizing the last.
func crumbs(parts ...string) string {
	sep := faintStyle.Render(" › ")
	out := make([]string, len(parts))
	for i, p := range parts {
		if i == len(parts)-1 {
			out[i] = accentStyle.Render(p)
		} else {
			out[i] = dimStyle.Render(p)
		}
	}
	return strings.Join(out, sep)
}

func (m Model) renderHeader() string {
	var trail string
	switch m.screen {
	case ScreenRepoList:
		trail = crumbs("Repositories")
	case ScreenChartList:
		trail = crumbs(m.selectedRepo.Name) + "  " + faintStyle.Render(m.selectedRepo.URL)
	case ScreenChartVersions:
		trail = crumbs(m.selectedRepo.Name, m.selectedChart.Name)
	case ScreenChartDetail:
		trail = crumbs(m.selectedRepo.Name, m.selectedChart.Name, m.selectedVersion.Version)
	}

	left := badgeStyle.Render("helm charts") + " " + trail
	right := m.renderActivity()
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		// Drop the right-hand activity before clipping the breadcrumb.
		right, gap = "", m.width-lipgloss.Width(left)
		if gap < 0 {
			gap = 0
		}
	}
	line := ansi.Truncate(left+strings.Repeat(" ", gap)+right, m.width, "…")
	rule := faintStyle.Render(strings.Repeat("─", m.width))
	return line + "\n" + rule
}

// renderActivity is the animated indicator at the right of the header.
func (m Model) renderActivity() string {
	switch {
	case m.updating && m.updateTotal > 1:
		return fmt.Sprintf("%s updating %d/%d %s", m.spinner.View(), m.updateDone, m.updateTotal, m.progress.View())
	case m.updating:
		return m.spinner.View() + " updating repo"
	case m.refreshing:
		return m.spinner.View() + " refreshing"
	case m.syncing:
		return m.spinner.View() + dimStyle.Render(" syncing repos")
	case m.loading:
		return m.spinner.View()
	}
	return ""
}

func (m Model) renderStatusBar() string {
	if m.mode == ModeFilter {
		return filterPromptStyle.Render("/") + m.filterInput.View()
	}

	left := m.statusLeft()
	right := m.statusRight()
	room := m.width - lipgloss.Width(right) - 1
	if room < 0 {
		room = 0
	}
	left = ansi.Truncate(left, room, "…")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) statusLeft() string {
	if m.statusMsg != "" {
		switch {
		case m.statusDim:
			return dimStyle.Render(m.statusMsg)
		case m.statusErr:
			return errorStyle.Render("✗ " + m.statusMsg)
		default:
			return okStyle.Render("✓ " + m.statusMsg)
		}
	}

	var parts []string
	if m.filterText != "" {
		parts = append(parts, activeFilterStyle.Render("filter: /"+m.filterText))
	}
	if l := sortLabel(m.currentSort()); l != "" {
		parts = append(parts, activeFilterStyle.Render("sort: "+l))
	}
	parts = append(parts, hints(m.screenHints()))
	return strings.Join(parts, "  ")
}

func (m Model) currentSort() int {
	switch m.screen {
	case ScreenRepoList:
		return m.repoSort
	case ScreenChartList:
		return m.chartSort
	}
	return sortDefault
}

// screenHints lists the most useful keys for the current screen as key/desc pairs.
func (m Model) screenHints() [][2]string {
	switch m.screen {
	case ScreenRepoList:
		return [][2]string{{"enter", "open"}, {"a", "add"}, {"r", "reload"}, {"u/R", "update"}, {"/", "filter"}, {"?", "help"}, {"q", "quit"}}
	case ScreenChartList:
		return [][2]string{{"enter", "open"}, {"esc", "back"}, {"r", "reload"}, {"R", "update"}, {"/", "filter"}, {"y", "copy"}, {"?", "help"}}
	case ScreenChartVersions:
		return [][2]string{{"enter", "open"}, {"esc", "back"}, {"r", "reload"}, {"R", "update"}, {"/", "filter"}, {"Y", "install cmd"}, {"?", "help"}}
	}
	return [][2]string{{"esc", "back"}, {"r", "reload"}, {"y", "copy yaml"}, {"Y", "install cmd"}, {"g/G", "top/end"}, {"?", "help"}}
}

func hints(pairs [][2]string) string {
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = keyStyle.Render(p[0]) + " " + dimStyle.Render(p[1])
	}
	return strings.Join(out, "  ")
}

// statusRight shows the cursor position / scroll percentage.
func (m Model) statusRight() string {
	if m.loading || m.err != nil {
		return ""
	}
	switch m.screen {
	case ScreenRepoList:
		return dimStyle.Render(fmt.Sprintf("%d/%d ", maxInt(m.repoTable.Cursor(), 0), len(m.filteredRepos)))
	case ScreenChartList:
		return dimStyle.Render(fmt.Sprintf("%d/%d ", minInt(m.chartTable.Cursor()+1, len(m.filteredCharts)), len(m.filteredCharts)))
	case ScreenChartVersions:
		return dimStyle.Render(fmt.Sprintf("%d/%d ", minInt(m.versionTable.Cursor()+1, len(m.filteredVersions)), len(m.filteredVersions)))
	}
	return dimStyle.Render(fmt.Sprintf("%3.0f%% ", m.viewport.ScrollPercent()*100))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) renderAddRepoDialog() string {
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("Add Helm Repository"))
	b.WriteString("\n\n")
	b.WriteString(m.addRepoName.View())
	b.WriteString("\n")
	b.WriteString(m.addRepoURL.View())
	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(firstLine(m.err)))
	}
	if m.loading {
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View() + loadingStyle.Render(" Adding repository…"))
	}
	b.WriteString("\n\n")
	b.WriteString(hints([][2]string{{"enter", "submit"}, {"tab", "next field"}, {"esc", "cancel"}}))
	return dialogStyle.Render(b.String())
}

func (m Model) renderConfirmExitDialog() string {
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("Exit?"))
	b.WriteString("\n\n")
	b.WriteString("Are you sure you want to exit?")
	b.WriteString("\n\n")
	b.WriteString(hints([][2]string{{"y", "yes"}, {"n", "no"}}))
	return dialogStyle.Render(b.String())
}

func (m Model) renderConfirmRemoveDialog() string {
	var b strings.Builder
	b.WriteString(errorStyle.MarginBottom(1).Render("Remove repository?"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("helm repo remove %s", accentStyle.Render(m.pendingRemove)))
	b.WriteString("\n\n")
	b.WriteString(hints([][2]string{{"y", "remove"}, {"n", "cancel"}}))
	return dangerDialogStyle.Render(b.String())
}

func (m Model) renderHelpOverlay() string {
	var cols []string
	for _, sec := range helpSections() {
		var b strings.Builder
		b.WriteString(accentStyle.Render(sec.title) + "\n")
		for _, binding := range sec.binds {
			help := binding.Help()
			fmt.Fprintf(&b, "%s %s\n", keyStyle.Render(fmt.Sprintf("%-12s", help.Key)), help.Desc)
		}
		cols = append(cols, strings.TrimRight(b.String(), "\n"))
	}

	// Two columns on wide terminals keeps the overlay short enough for small windows.
	var body string
	if m.width >= 110 {
		gap := strings.Repeat(" ", 4)
		left := cols[0]
		right := strings.Join(cols[1:], "\n\n")
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	} else {
		body = strings.Join(cols, "\n\n")
	}

	return dialogStyle.Render(dialogTitleStyle.Render("Key Bindings") + "\n" + body +
		"\n\n" + dimStyle.Render("press any key to close"))
}

// overlay draws dialog centered on top of base, leaving the rest of base visible.
func overlay(base, dialog string, w, h int) string {
	lines := strings.Split(base, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	dl := strings.Split(dialog, "\n")
	dw := lipgloss.Width(dialog)
	x := maxInt((w-dw)/2, 0)
	y := maxInt((h-len(dl))/2, 0)

	for i, d := range dl {
		row := y + i
		if row >= len(lines) {
			break
		}
		left := ansi.Truncate(lines[row], x, "")
		if pad := x - ansi.StringWidth(left); pad > 0 {
			left += strings.Repeat(" ", pad)
		}
		right := ansi.TruncateLeft(lines[row], x+dw, "")
		lines[row] = left + "\x1b[0m" + d + "\x1b[0m" + right
	}
	return strings.Join(lines, "\n")
}
