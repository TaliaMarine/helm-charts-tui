package tui

import (
	"sort"
	"strings"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// Sort orders cycled with `s`. Zero is always the order helm returned.
const (
	sortDefault = iota
	sortNameAsc
	sortNameDesc
	sortCountDesc // repos only
)

func sortLabel(mode int) string {
	switch mode {
	case sortNameAsc:
		return "name ↑"
	case sortNameDesc:
		return "name ↓"
	case sortCountDesc:
		return "charts ↓"
	}
	return ""
}

// sortRepos returns a sorted copy of repos (the input is never mutated).
func sortRepos(repos []helm.Repo, mode int, counts map[string]int) []helm.Repo {
	if mode == sortDefault {
		return repos
	}
	out := append([]helm.Repo(nil), repos...)
	sort.SliceStable(out, func(i, j int) bool {
		switch mode {
		case sortNameDesc:
			return strings.ToLower(out[i].Name) > strings.ToLower(out[j].Name)
		case sortCountDesc:
			return counts[out[i].Name] > counts[out[j].Name]
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

// sortCharts returns a sorted copy of charts (the input is never mutated).
func sortCharts(charts []helm.Chart, mode int) []helm.Chart {
	if mode == sortDefault || mode == sortCountDesc {
		return charts
	}
	out := append([]helm.Chart(nil), charts...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		if mode == sortNameDesc {
			return a > b
		}
		return a < b
	})
	return out
}
