package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var noColor = os.Getenv("NO_COLOR") != ""

func newStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}

var (
	headerStyle = func() lipgloss.Style {
		s := newStyle().Bold(true).MarginBottom(1)
		if !noColor {
			s = s.Foreground(lipgloss.Color("12"))
		}
		return s
	}()

	dimStyle = func() lipgloss.Style {
		s := newStyle()
		if !noColor {
			s = s.Foreground(lipgloss.Color("8"))
		}
		return s
	}()

	errorStyle = func() lipgloss.Style {
		s := newStyle().Bold(true)
		if !noColor {
			s = s.Foreground(lipgloss.Color("9"))
		}
		return s
	}()

	statusBarStyle = func() lipgloss.Style {
		s := newStyle()
		if !noColor {
			s = s.Foreground(lipgloss.Color("7"))
		}
		return s
	}()

	dialogStyle = func() lipgloss.Style {
		return newStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
	}()

	dialogTitleStyle = func() lipgloss.Style {
		s := newStyle().Bold(true).MarginBottom(1)
		if !noColor {
			s = s.Foreground(lipgloss.Color("12"))
		}
		return s
	}()

	addRepoMarker = func() lipgloss.Style {
		s := newStyle().Bold(true)
		if !noColor {
			s = s.Foreground(lipgloss.Color("10"))
		}
		return s
	}()

	filterPromptStyle = func() lipgloss.Style {
		s := newStyle().Bold(true)
		if !noColor {
			s = s.Foreground(lipgloss.Color("11"))
		}
		return s
	}()

	activeFilterStyle = func() lipgloss.Style {
		s := newStyle()
		if !noColor {
			s = s.Foreground(lipgloss.Color("11"))
		}
		return s
	}()

	loadingStyle = func() lipgloss.Style {
		s := newStyle()
		if !noColor {
			s = s.Foreground(lipgloss.Color("8"))
		}
		return s
	}()
)
