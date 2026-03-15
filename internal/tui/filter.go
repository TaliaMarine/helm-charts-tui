package tui

import (
	"strings"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// FilterRepos filters repos by case-insensitive substring match on name or URL.
func FilterRepos(repos []helm.Repo, text string) []helm.Repo {
	if text == "" {
		return repos
	}
	lower := strings.ToLower(text)
	var result []helm.Repo
	for _, r := range repos {
		if strings.Contains(strings.ToLower(r.Name), lower) ||
			strings.Contains(strings.ToLower(r.URL), lower) {
			result = append(result, r)
		}
	}
	return result
}

// FilterCharts filters charts by case-insensitive substring match on name, version, app version, or description.
func FilterCharts(charts []helm.Chart, text string) []helm.Chart {
	if text == "" {
		return charts
	}
	lower := strings.ToLower(text)
	var result []helm.Chart
	for _, c := range charts {
		if strings.Contains(strings.ToLower(c.Name), lower) ||
			strings.Contains(strings.ToLower(c.Version), lower) ||
			strings.Contains(strings.ToLower(c.AppVersion), lower) ||
			strings.Contains(strings.ToLower(c.Description), lower) {
			result = append(result, c)
		}
	}
	return result
}
