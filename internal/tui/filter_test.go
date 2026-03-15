package tui

import (
	"testing"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

func TestFilterRepos(t *testing.T) {
	repos := []helm.Repo{
		{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
		{Name: "jetstack", URL: "https://charts.jetstack.io"},
	}

	tests := []struct {
		filter string
		want   int
	}{
		{"", 3},
		{"bit", 1},
		{"helm.sh", 1},
		{"xyz", 0},
		{"BIT", 1},
		{"charts", 3},
		{"jet", 1},
		{"STABLE", 1},
	}

	for _, tt := range tests {
		t.Run("filter="+tt.filter, func(t *testing.T) {
			got := FilterRepos(repos, tt.filter)
			if len(got) != tt.want {
				t.Errorf("FilterRepos(%q) returned %d results, want %d", tt.filter, len(got), tt.want)
			}
		})
	}
}

func TestFilterReposEmpty(t *testing.T) {
	got := FilterRepos(nil, "test")
	if len(got) != 0 {
		t.Errorf("FilterRepos(nil) returned %d results, want 0", len(got))
	}
}

func TestFilterCharts(t *testing.T) {
	charts := []helm.Chart{
		{Name: "stable/nginx", Version: "1.0.0", AppVersion: "1.19", Description: "An nginx web server"},
		{Name: "stable/redis", Version: "2.0.0", AppVersion: "6.2", Description: "A key-value store"},
		{Name: "stable/postgresql", Version: "3.0.0", AppVersion: "14.0", Description: "A relational database"},
	}

	tests := []struct {
		filter string
		want   int
	}{
		{"", 3},
		{"nginx", 1},
		{"stable", 3},
		{"key-value", 1},
		{"1.0.0", 1},
		{"14.0", 1},
		{"database", 1},
		{"REDIS", 1},
		{"xyz", 0},
	}

	for _, tt := range tests {
		t.Run("filter="+tt.filter, func(t *testing.T) {
			got := FilterCharts(charts, tt.filter)
			if len(got) != tt.want {
				t.Errorf("FilterCharts(%q) returned %d results, want %d", tt.filter, len(got), tt.want)
			}
		})
	}
}

func TestFilterChartsEmpty(t *testing.T) {
	got := FilterCharts(nil, "test")
	if len(got) != 0 {
		t.Errorf("FilterCharts(nil) returned %d results, want 0", len(got))
	}
}
