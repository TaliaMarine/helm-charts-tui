package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current screen.
func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Main content
	if m.loading {
		b.WriteString("\n")
		b.WriteString(loadingStyle.Render("  Loading..."))
		b.WriteString("\n")
	} else if m.err != nil && m.mode != ModeAddRepo {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
		b.WriteString("\n")
	} else {
		switch m.screen {
		case ScreenRepoList:
			b.WriteString(m.repoTable.View())
		case ScreenChartList:
			b.WriteString(m.chartTable.View())
		case ScreenChartVersions:
			b.WriteString(m.versionTable.View())
		case ScreenChartDetail:
			b.WriteString(m.viewport.View())
		}
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	// Filter input
	if m.mode == ModeFilter {
		b.WriteString("\n")
		b.WriteString(filterPromptStyle.Render("/") + m.filterInput.View())
	}

	view := b.String()

	// Overlays
	switch m.mode {
	case ModeAddRepo:
		view = m.placeOverlay(view, m.renderAddRepoDialog())
	case ModeConfirmExit:
		view = m.placeOverlay(view, m.renderConfirmExitDialog())
	case ModeHelp:
		view = m.placeOverlay(view, m.renderHelpOverlay())
	}

	return view
}

func (m Model) renderHeader() string {
	switch m.screen {
	case ScreenRepoList:
		return headerStyle.Render("Helm Chart Repositories")
	case ScreenChartList:
		return headerStyle.Render(
			m.selectedRepo.Name + "  " + dimStyle.Render(m.selectedRepo.URL),
		)
	case ScreenChartVersions:
		return headerStyle.Render(
			m.selectedRepo.Name + " > " + m.selectedChart.Name,
		)
	case ScreenChartDetail:
		return headerStyle.Render(
			m.selectedRepo.Name + " > " + m.selectedChart.Name + " > " + m.selectedVersion.Version,
		)
	}
	return ""
}

func (m Model) renderStatusBar() string {
	var parts []string

	if m.statusMsg != "" {
		parts = append(parts, m.statusMsg)
	}

	if m.filterText != "" && m.mode != ModeFilter {
		parts = append(parts, activeFilterStyle.Render(fmt.Sprintf("filter: /%s", m.filterText)))
	}

	// Navigation hints
	switch m.screen {
	case ScreenRepoList:
		parts = append(parts, dimStyle.Render("enter:select  /:filter  ?:help  q:quit"))
	case ScreenChartList, ScreenChartVersions:
		parts = append(parts, dimStyle.Render("enter:select  esc:back  /:filter  ?:help  q:quit"))
	case ScreenChartDetail:
		parts = append(parts, dimStyle.Render("esc:back  ?:help  q:quit"))
	}

	return statusBarStyle.Render(strings.Join(parts, "  "))
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
		b.WriteString(errorStyle.Render(m.err.Error()))
	}
	if m.loading {
		b.WriteString("\n\n")
		b.WriteString(loadingStyle.Render("Adding repository..."))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter:submit  tab:next field  esc:cancel"))
	return dialogStyle.Render(b.String())
}

func (m Model) renderConfirmExitDialog() string {
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("Exit?"))
	b.WriteString("\n\n")
	b.WriteString("Are you sure you want to exit?")
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("y:yes  n:no"))
	return dialogStyle.Render(b.String())
}

func (m Model) renderHelpOverlay() string {
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("Key Bindings"))
	b.WriteString("\n\n")

	bindings := helpBindings()
	for _, binding := range bindings {
		help := binding.Help()
		fmt.Fprintf(&b, "  %-12s %s\n", help.Key, help.Desc)
	}

	return dialogStyle.Render(b.String())
}

// placeOverlay centers a dialog over the base view.
func (m Model) placeOverlay(base, overlay string) string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
