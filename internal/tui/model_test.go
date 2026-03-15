package tui

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

func testModel() Model {
	mock := &helm.MockExecutor{
		RepoListFunc: func(ctx context.Context) ([]helm.Repo, error) {
			return []helm.Repo{
				{Name: "stable", URL: "https://charts.helm.sh/stable"},
				{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
			}, nil
		},
		SearchRepoFunc: func(ctx context.Context, keyword string) ([]helm.Chart, error) {
			return []helm.Chart{
				{Name: "stable/nginx", Version: "1.0.0", AppVersion: "1.19", Description: "Web server"},
				{Name: "stable/redis", Version: "2.0.0", AppVersion: "6.2", Description: "KV store"},
			}, nil
		},
		SearchRepoVersionsFunc: func(ctx context.Context, keyword string) ([]helm.Chart, error) {
			return []helm.Chart{
				{Name: "stable/nginx", Version: "1.0.0", AppVersion: "1.19", Description: "Web server"},
				{Name: "stable/nginx", Version: "0.9.0", AppVersion: "1.18", Description: "Web server"},
			}, nil
		},
		ShowChartFunc: func(ctx context.Context, chartName, version string) (string, error) {
			return "apiVersion: v2\nname: nginx\nversion: 1.0.0\n", nil
		},
	}

	m := New(context.Background(), mock)
	// Simulate window size
	m, _ = toModel(m.Update(tea.WindowSizeMsg{Width: 120, Height: 40}))
	return m
}

func toModel(model tea.Model, cmd tea.Cmd) (Model, tea.Cmd) {
	return model.(Model), cmd
}

// Simulate receiving a message from a command.
func runCmd(m Model, cmd tea.Cmd) (Model, tea.Cmd) {
	if cmd == nil {
		return m, nil
	}
	msg := cmd()
	return toModel(m.Update(msg))
}

func TestInitialState(t *testing.T) {
	m := testModel()
	if m.CurrentScreen() != ScreenRepoList {
		t.Errorf("initial screen = %d, want ScreenRepoList", m.CurrentScreen())
	}
	if m.CurrentMode() != ModeNormal {
		t.Errorf("initial mode = %d, want ModeNormal", m.CurrentMode())
	}
	if !m.IsLoading() {
		t.Error("initial loading should be true")
	}
}

func TestReposLoaded(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	if m.IsLoading() {
		t.Error("loading should be false after repos loaded")
	}
	if len(m.Repos()) != 1 {
		t.Errorf("repos count = %d, want 1", len(m.Repos()))
	}
}

func TestChartCountsLoaded(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))
	m, _ = toModel(m.Update(chartCountsLoadedMsg(map[string]int{"stable": 42})))

	if m.chartCounts["stable"] != 42 {
		t.Errorf("chart count for stable = %d, want 42", m.chartCounts["stable"])
	}
}

func TestNavigateToChartList(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	// Move cursor down to first repo (past "Add new repo" row)
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	// Press enter to select repo
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))

	if m.CurrentScreen() != ScreenChartList {
		t.Errorf("screen = %d, want ScreenChartList", m.CurrentScreen())
	}
	if !m.IsLoading() {
		t.Error("should be loading charts")
	}
	if m.SelectedRepo().Name != "stable" {
		t.Errorf("selected repo = %q, want %q", m.SelectedRepo().Name, "stable")
	}

	// Simulate charts loaded
	m, _ = runCmd(m, cmd)
	if m.IsLoading() {
		t.Error("should not be loading after charts loaded")
	}
}

func TestNavigateToChartVersions(t *testing.T) {
	m := testModel()
	// Load repos
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))
	// Select repo
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)

	// Select chart
	m, cmd = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	if m.CurrentScreen() != ScreenChartVersions {
		t.Errorf("screen = %d, want ScreenChartVersions", m.CurrentScreen())
	}

	// Load versions
	m, _ = runCmd(m, cmd)
	if len(m.Versions()) != 2 {
		t.Errorf("versions count = %d, want 2", len(m.Versions()))
	}
}

func TestNavigateToChartDetail(t *testing.T) {
	m := testModel()
	// Load repos -> select repo -> load charts -> select chart -> load versions
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)
	m, cmd = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)

	// Select version
	m, cmd = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	if m.CurrentScreen() != ScreenChartDetail {
		t.Errorf("screen = %d, want ScreenChartDetail", m.CurrentScreen())
	}

	// Load detail
	m, _ = runCmd(m, cmd)
	if m.Detail() == "" {
		t.Error("detail should not be empty after loading")
	}
}

func TestEscapeNavigation(t *testing.T) {
	m := testModel()
	// Navigate to chart list
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)

	if m.CurrentScreen() != ScreenChartList {
		t.Fatalf("expected ScreenChartList, got %d", m.CurrentScreen())
	}

	// Press ESC to go back
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentScreen() != ScreenRepoList {
		t.Errorf("screen = %d, want ScreenRepoList after ESC", m.CurrentScreen())
	}
}

func TestEscapeOnRepoListShowsConfirm(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentMode() != ModeConfirmExit {
		t.Errorf("mode = %d, want ModeConfirmExit", m.CurrentMode())
	}

	// Press 'n' to cancel
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}))
	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after cancel", m.CurrentMode())
	}
}

func TestQuitFromAnyScreen(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	// Navigate to chart list
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)

	// Press 'q' - should quit
	_, cmd = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	// tea.Quit returns a special cmd; we just verify it's not nil
	if cmd == nil {
		t.Error("q should produce a quit command")
	}
}

func TestCtrlCAlwaysQuits(t *testing.T) {
	m := testModel()
	// Even in filter mode, ctrl+c should quit
	m.mode = ModeFilter
	_, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}))
	if cmd == nil {
		t.Error("ctrl+c should produce a quit command")
	}
}

func TestQuitNotActiveInFilterMode(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	// Enter filter mode
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}))
	if m.CurrentMode() != ModeFilter {
		t.Fatalf("mode = %d, want ModeFilter", m.CurrentMode())
	}

	// Press 'q' - should NOT quit, should type 'q' in filter
	m2, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	// Should still be in filter mode
	if m2.CurrentMode() != ModeFilter {
		t.Errorf("mode = %d, want ModeFilter (q should type, not quit)", m2.CurrentMode())
	}
	_ = cmd
}

func TestFilterActivation(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	// Press '/' to enter filter mode
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}))
	if m.CurrentMode() != ModeFilter {
		t.Errorf("mode = %d, want ModeFilter", m.CurrentMode())
	}

	// Press Escape to cancel filter
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after ESC", m.CurrentMode())
	}
	if m.FilterText() != "" {
		t.Errorf("filter text = %q, want empty after cancel", m.FilterText())
	}
}

func TestFilterApply(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
		{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
	})))

	// Enter filter mode
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}))

	// Type "bit"
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}))
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}))
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}))

	// Confirm with Enter
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))

	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after enter", m.CurrentMode())
	}
	if m.FilterText() != "bit" {
		t.Errorf("filter text = %q, want %q", m.FilterText(), "bit")
	}
	if len(m.filteredRepos) != 1 {
		t.Errorf("filtered repos = %d, want 1", len(m.filteredRepos))
	}
}

func TestEscapeClearsFilterBeforeNavigating(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))
	// Navigate to chart list
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyDown}))
	m, cmd := toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	m, _ = runCmd(m, cmd)

	// Apply a filter
	m.filterText = "nginx"
	m.applyFilter()

	// First ESC should clear filter, not navigate
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentScreen() != ScreenChartList {
		t.Errorf("screen = %d, want ScreenChartList (first ESC should clear filter)", m.CurrentScreen())
	}
	if m.FilterText() != "" {
		t.Errorf("filter should be cleared after first ESC")
	}

	// Second ESC should navigate back
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentScreen() != ScreenRepoList {
		t.Errorf("screen = %d, want ScreenRepoList after second ESC", m.CurrentScreen())
	}
}

func TestAddRepoFlow(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
	})))

	// Cursor is at row 0 ("Add new repo"), press Enter
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	if m.CurrentMode() != ModeAddRepo {
		t.Errorf("mode = %d, want ModeAddRepo", m.CurrentMode())
	}

	// Press Escape to cancel
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEscape}))
	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after cancel", m.CurrentMode())
	}
}

func TestAddRepoValidation(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{})))

	// Enter add repo mode
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))

	// Try to submit empty
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	if m.Error() == nil {
		t.Error("should have validation error for empty inputs")
	}
}

func TestRepoAddedSuccess(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{})))

	m, _ = toModel(m.Update(repoAddedMsg("newrepo")))
	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after add success", m.CurrentMode())
	}
	if m.StatusMsg() == "" {
		t.Error("status message should be set after successful add")
	}
}

func TestLoadError(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(loadErrMsg{context: "loading repos", err: fmt.Errorf("connection refused")}))

	if m.Error() == nil {
		t.Error("error should be set")
	}
	if m.IsLoading() {
		t.Error("loading should be false after error")
	}
}

func TestHelpOverlay(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{})))

	// Press '?'
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}))
	if m.CurrentMode() != ModeHelp {
		t.Errorf("mode = %d, want ModeHelp", m.CurrentMode())
	}

	// Any key dismisses help
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	if m.CurrentMode() != ModeNormal {
		t.Errorf("mode = %d, want ModeNormal after dismissing help", m.CurrentMode())
	}
}

func TestWindowResize(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(tea.WindowSizeMsg{Width: 200, Height: 50}))

	if m.width != 200 {
		t.Errorf("width = %d, want 200", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestClearStatus(t *testing.T) {
	m := testModel()
	m.statusMsg = "some message"
	m, _ = toModel(m.Update(clearStatusMsg{}))
	if m.StatusMsg() != "" {
		t.Errorf("status = %q, want empty after clear", m.StatusMsg())
	}
}

func TestFilterNotAvailableOnDetailScreen(t *testing.T) {
	m := testModel()
	m.screen = ScreenChartDetail
	m.mode = ModeNormal

	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}))
	if m.CurrentMode() == ModeFilter {
		t.Error("filter should not activate on detail screen")
	}
}

func TestTabSwitchesInputInAddRepo(t *testing.T) {
	m := testModel()
	m, _ = toModel(m.Update(reposLoadedMsg([]helm.Repo{})))
	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))

	if m.addRepoFocus != 0 {
		t.Errorf("initial focus = %d, want 0 (name)", m.addRepoFocus)
	}

	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyTab}))
	if m.addRepoFocus != 1 {
		t.Errorf("focus after tab = %d, want 1 (url)", m.addRepoFocus)
	}

	m, _ = toModel(m.Update(tea.KeyMsg{Type: tea.KeyTab}))
	if m.addRepoFocus != 0 {
		t.Errorf("focus after second tab = %d, want 0 (name)", m.addRepoFocus)
	}
}
