package helm

import "context"

// MockExecutor implements Executor with configurable function fields for testing.
type MockExecutor struct {
	RepoListFunc           func(ctx context.Context) ([]Repo, error)
	RepoAddFunc            func(ctx context.Context, name, url string) error
	RepoUpdateFunc         func(ctx context.Context, name string) error
	SearchRepoFunc         func(ctx context.Context, keyword string) ([]Chart, error)
	SearchRepoVersionsFunc func(ctx context.Context, keyword string) ([]Chart, error)
	ShowChartFunc          func(ctx context.Context, chartName, version string) (string, error)
}

func (m *MockExecutor) RepoList(ctx context.Context) ([]Repo, error) {
	if m.RepoListFunc != nil {
		return m.RepoListFunc(ctx)
	}
	return nil, nil
}

func (m *MockExecutor) RepoAdd(ctx context.Context, name, url string) error {
	if m.RepoAddFunc != nil {
		return m.RepoAddFunc(ctx, name, url)
	}
	return nil
}

func (m *MockExecutor) RepoUpdate(ctx context.Context, name string) error {
	if m.RepoUpdateFunc != nil {
		return m.RepoUpdateFunc(ctx, name)
	}
	return nil
}

func (m *MockExecutor) SearchRepo(ctx context.Context, keyword string) ([]Chart, error) {
	if m.SearchRepoFunc != nil {
		return m.SearchRepoFunc(ctx, keyword)
	}
	return nil, nil
}

func (m *MockExecutor) SearchRepoVersions(ctx context.Context, keyword string) ([]Chart, error) {
	if m.SearchRepoVersionsFunc != nil {
		return m.SearchRepoVersionsFunc(ctx, keyword)
	}
	return nil, nil
}

func (m *MockExecutor) ShowChart(ctx context.Context, chartName, version string) (string, error) {
	if m.ShowChartFunc != nil {
		return m.ShowChartFunc(ctx, chartName, version)
	}
	return "", nil
}
