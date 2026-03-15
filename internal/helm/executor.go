package helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Executor abstracts Helm CLI operations for testability.
type Executor interface {
	// RepoList returns all configured helm repos.
	RepoList(ctx context.Context) ([]Repo, error)

	// RepoAdd adds a new repo with the given name and URL.
	RepoAdd(ctx context.Context, name, url string) error

	// RepoUpdate updates the local cache for the named repo.
	RepoUpdate(ctx context.Context, name string) error

	// SearchRepo searches for charts in a repo (pass "reponame/" for all charts in a repo).
	SearchRepo(ctx context.Context, keyword string) ([]Chart, error)

	// SearchRepoVersions returns all versions of charts matching keyword.
	SearchRepoVersions(ctx context.Context, keyword string) ([]Chart, error)

	// ShowChart returns the YAML chart metadata for a specific chart and version.
	ShowChart(ctx context.Context, chartName, version string) (string, error)
}

// RealExecutor calls the helm CLI binary.
type RealExecutor struct{}

func (e *RealExecutor) RepoList(ctx context.Context) ([]Repo, error) {
	out, err := runHelm(ctx, "repo", "list", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("running helm repo list: %w", err)
	}
	repos, err := ParseRepoList(out)
	if err != nil {
		return nil, fmt.Errorf("parsing helm repo list: %w", err)
	}
	return repos, nil
}

func (e *RealExecutor) RepoAdd(ctx context.Context, name, url string) error {
	_, err := runHelm(ctx, "repo", "add", name, url)
	if err != nil {
		return fmt.Errorf("running helm repo add: %w", err)
	}
	return nil
}

func (e *RealExecutor) RepoUpdate(ctx context.Context, name string) error {
	_, err := runHelm(ctx, "repo", "update", name)
	if err != nil {
		return fmt.Errorf("running helm repo update: %w", err)
	}
	return nil
}

func (e *RealExecutor) SearchRepo(ctx context.Context, keyword string) ([]Chart, error) {
	out, err := runHelm(ctx, "search", "repo", keyword, "--output", "json")
	if err != nil {
		// helm search returns error when no results found
		if strings.Contains(err.Error(), "exit status") {
			return nil, nil
		}
		return nil, fmt.Errorf("running helm search repo: %w", err)
	}
	charts, err := ParseChartList(out)
	if err != nil {
		return nil, fmt.Errorf("parsing helm search repo: %w", err)
	}
	return charts, nil
}

func (e *RealExecutor) SearchRepoVersions(ctx context.Context, keyword string) ([]Chart, error) {
	out, err := runHelm(ctx, "search", "repo", keyword, "--versions", "--output", "json")
	if err != nil {
		if strings.Contains(err.Error(), "exit status") {
			return nil, nil
		}
		return nil, fmt.Errorf("running helm search repo --versions: %w", err)
	}
	charts, err := ParseChartList(out)
	if err != nil {
		return nil, fmt.Errorf("parsing helm search repo --versions: %w", err)
	}
	return charts, nil
}

func (e *RealExecutor) ShowChart(ctx context.Context, chartName, version string) (string, error) {
	out, err := runHelm(ctx, "show", "chart", chartName, "--version", version)
	if err != nil {
		return "", fmt.Errorf("running helm show chart: %w", err)
	}
	return string(out), nil
}

func runHelm(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "helm", args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			return nil, fmt.Errorf("%w: %s", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return out, nil
}

// ParseRepoList parses the JSON output of `helm repo list --output json`.
func ParseRepoList(data []byte) ([]Repo, error) {
	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

// ParseChartList parses the JSON output of `helm search repo --output json`.
func ParseChartList(data []byte) ([]Chart, error) {
	var charts []Chart
	if err := json.Unmarshal(data, &charts); err != nil {
		return nil, err
	}
	return charts, nil
}
