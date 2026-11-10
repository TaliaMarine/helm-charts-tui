package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// renderYAML colors keys and comments and wraps long lines to width.
// Under NO_COLOR only wrapping is applied.
func renderYAML(src string, width int) string {
	lines := strings.Split(strings.TrimRight(src, "\n"), "\n")
	for i, l := range lines {
		if !noColor {
			l = highlightYAMLLine(l)
		}
		if width > 4 {
			l = ansi.Wrap(l, width-1, " ")
		}
		lines[i] = l
	}
	return strings.Join(lines, "\n")
}

func highlightYAMLLine(l string) string {
	trimmed := strings.TrimLeft(l, " ")
	indent := l[:len(l)-len(trimmed)]
	if strings.HasPrefix(trimmed, "#") {
		return dimStyle.Render(l)
	}
	dash := ""
	if strings.HasPrefix(trimmed, "- ") {
		dash, trimmed = "- ", trimmed[2:]
	}
	k, v, ok := strings.Cut(trimmed, ":")
	if !ok || strings.ContainsAny(k, " \"'{[") {
		return l
	}
	return indent + dash + accentStyle.Render(k) + ":" + v
}
