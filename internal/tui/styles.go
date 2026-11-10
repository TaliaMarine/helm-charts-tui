package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var noColor = os.Getenv("NO_COLOR") != ""

// Palette. Adaptive so it stays legible on light and dark terminals.
var (
	colAccent = lipgloss.AdaptiveColor{Light: "25", Dark: "75"}
	colOK     = lipgloss.AdaptiveColor{Light: "28", Dark: "42"}
	colWarn   = lipgloss.AdaptiveColor{Light: "130", Dark: "221"}
	colErr    = lipgloss.AdaptiveColor{Light: "160", Dark: "203"}
	colDim    = lipgloss.AdaptiveColor{Light: "245", Dark: "242"}
	colFaint  = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
	colSelBG  = lipgloss.AdaptiveColor{Light: "153", Dark: "24"}
)

// style builds a style that carries no color when NO_COLOR is set.
func style(fg lipgloss.TerminalColor, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Bold(bold)
	if !noColor && fg != nil {
		s = s.Foreground(fg)
	}
	return s
}

var (
	dimStyle    = style(colDim, false)
	faintStyle  = style(colFaint, false)
	errorStyle  = style(colErr, true)
	okStyle     = style(colOK, true)
	warnStyle   = style(colWarn, true)
	accentStyle = style(colAccent, true)
	spinStyle   = style(colAccent, false)

	loadingStyle      = dimStyle
	filterPromptStyle = warnStyle
	activeFilterStyle = style(colWarn, false)
	dialogTitleStyle  = accentStyle.MarginBottom(1)

	// badge is the app tag at the far left of the header.
	badgeStyle = func() lipgloss.Style {
		s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
		if noColor {
			return s.Reverse(true)
		}
		return s.Foreground(lipgloss.Color("231")).Background(lipgloss.Color("63"))
	}()

	keyStyle = style(colAccent, true)

	dialogStyle = func() lipgloss.Style {
		s := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
		if !noColor {
			s = s.BorderForeground(colAccent)
		}
		return s
	}()

	dangerDialogStyle = func() lipgloss.Style {
		s := dialogStyle
		if !noColor {
			s = s.BorderForeground(colErr)
		}
		return s
	}()
)

// tableStyles returns table styling that follows the palette.
func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Bold(true).Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Cell = lipgloss.NewStyle().Padding(0, 1)
	if noColor {
		s.Selected = lipgloss.NewStyle().Bold(true).Reverse(true)
		return s
	}
	s.Header = s.Header.Foreground(colAccent).BorderForeground(colFaint)
	s.Selected = lipgloss.NewStyle().Bold(true).Background(colSelBG)
	return s
}
