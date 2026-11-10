package tui

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// drain executes cmd and feeds every resulting message back into the model,
// recursing through batches. Commands that sleep (toast/animation ticks) are
// abandoned after a short timeout, and spinner ticks are skipped.
func drain(m Model, cmd tea.Cmd) Model {
	return drainDepth(m, cmd, 0)
}

func drainDepth(m Model, cmd tea.Cmd, depth int) Model {
	if cmd == nil || depth > 8 {
		return m
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(10 * time.Millisecond):
		return m
	}
	switch msg := msg.(type) {
	case nil:
		return m
	case spinner.TickMsg:
		return m
	case tea.BatchMsg:
		for _, c := range msg {
			m = drainDepth(m, c, depth+1)
		}
		return m
	default:
		next, c := m.Update(msg)
		return drainDepth(next.(Model), c, depth+1)
	}
}

func runeKey(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func send(m Model, msg tea.Msg) (Model, tea.Cmd) { return toModel(m.Update(msg)) }

// loadedRepoModel returns a model sitting on the repo list with data loaded.
func loadedRepoModel(mock *helm.MockExecutor) Model {
	m := New(context.Background(), mock)
	m, _ = send(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = send(m, reposLoadedMsg([]helm.Repo{
		{Name: "stable", URL: "https://charts.helm.sh/stable"},
		{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
	}))
	m, _ = send(m, chartCountsLoadedMsg(map[string]int{"stable": 2, "bitnami": 9}))
	return m
}

type updateRecorder struct {
	mu    sync.Mutex
	names []string
}

func (r *updateRecorder) fn(fail string) func(context.Context, string) error {
	return func(_ context.Context, name string) error {
		r.mu.Lock()
		r.names = append(r.names, name)
		r.mu.Unlock()
		if name == fail {
			return errors.New("boom\nstack")
		}
		return nil
	}
}

func TestReloadKeepsDataVisible(t *testing.T) {
	calls := 0
	m := loadedRepoModel(&helm.MockExecutor{
		RepoListFunc: func(context.Context) ([]helm.Repo, error) {
			calls++
			return []helm.Repo{{Name: "stable", URL: "u"}}, nil
		},
	})

	m, cmd := send(m, runeKey('r'))
	if m.IsLoading() || !m.IsRefreshing() {
		t.Fatalf("loading=%v refreshing=%v, want refreshing only", m.IsLoading(), m.IsRefreshing())
	}
	if len(m.Repos()) != 2 {
		t.Error("stale repos should stay while reloading")
	}
	m = drain(m, cmd)
	if calls != 1 || m.IsRefreshing() || len(m.Repos()) != 1 {
		t.Errorf("calls=%d refreshing=%v repos=%d", calls, m.IsRefreshing(), len(m.Repos()))
	}
	if m.StatusMsg() == "" {
		t.Error("expected a reload confirmation toast")
	}
}

func TestReloadOnEveryScreen(t *testing.T) {
	var charts, versions, details int
	mock := &helm.MockExecutor{
		SearchRepoFunc: func(context.Context, string) ([]helm.Chart, error) {
			charts++
			return []helm.Chart{{Name: "stable/nginx", Version: "1"}}, nil
		},
		SearchRepoVersionsFunc: func(context.Context, string) ([]helm.Chart, error) {
			versions++
			return []helm.Chart{{Name: "stable/nginx", Version: "1"}}, nil
		},
		ShowChartFunc: func(context.Context, string, string) (string, error) {
			details++
			return "name: nginx\n", nil
		},
	}
	m := loadedRepoModel(mock)
	for range 3 { // descend: repo -> charts -> versions -> detail
		if m.CurrentScreen() == ScreenRepoList {
			m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
		}
		var cmd tea.Cmd
		m, cmd = send(m, tea.KeyMsg{Type: tea.KeyEnter})
		m = drain(m, cmd)
	}
	if m.CurrentScreen() != ScreenChartDetail {
		t.Fatalf("screen = %d", m.CurrentScreen())
	}
	charts, versions, details = 0, 0, 0

	m, cmd := send(m, tea.KeyMsg{Type: tea.KeyF5})
	m = drain(m, cmd)
	if details != 1 || charts != 0 {
		t.Errorf("detail reload: details=%d charts=%d", details, charts)
	}
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	m, cmd = send(m, runeKey('r'))
	m = drain(m, cmd)
	if versions != 1 {
		t.Errorf("versions reloads = %d, want 1", versions)
	}
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyEsc})
	m, cmd = send(m, runeKey('r'))
	_ = drain(m, cmd)
	if charts != 1 {
		t.Errorf("charts reloads = %d, want 1", charts)
	}
}

func TestReloadErrorKeepsStaleData(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{
		RepoListFunc: func(context.Context) ([]helm.Repo, error) { return nil, errors.New("helm gone") },
	})
	m, cmd := send(m, runeKey('r'))
	m = drain(m, cmd)
	if m.Error() != nil {
		t.Error("a failed refresh must not replace the screen with an error")
	}
	if len(m.Repos()) != 2 || m.StatusMsg() == "" || !m.statusErr {
		t.Errorf("repos=%d status=%q err=%v", len(m.Repos()), m.StatusMsg(), m.statusErr)
	}
}

func TestUpdateAllRepos(t *testing.T) {
	rec := &updateRecorder{}
	m := loadedRepoModel(&helm.MockExecutor{RepoUpdateFunc: rec.fn("")})

	m, cmd := send(m, runeKey('R'))
	if !m.IsUpdating() {
		t.Fatal("R should start an update")
	}
	// A second R while running is rejected.
	m2, cmd2 := send(m, runeKey('R'))
	if !m2.statusErr || cmd2 == nil {
		t.Error("concurrent update should be refused with an error toast")
	}
	m = drain(m, cmd)
	if m.IsUpdating() {
		t.Error("update should have finished")
	}
	if len(rec.names) != 2 {
		t.Errorf("updated %v, want both repos", rec.names)
	}
	if m.StatusMsg() != "updated 2 repositories" {
		t.Errorf("toast = %q", m.StatusMsg())
	}
}

func TestUpdateReportsPartialFailure(t *testing.T) {
	rec := &updateRecorder{}
	m := loadedRepoModel(&helm.MockExecutor{RepoUpdateFunc: rec.fn("bitnami")})
	m, cmd := send(m, runeKey('R'))
	m = drain(m, cmd)
	if m.StatusMsg() != "updated 1/2 repositories; failed: bitnami" || !m.statusErr {
		t.Errorf("toast = %q (err=%v)", m.StatusMsg(), m.statusErr)
	}
}

func TestUpdateHighlightedRepoOnly(t *testing.T) {
	rec := &updateRecorder{}
	m := loadedRepoModel(&helm.MockExecutor{RepoUpdateFunc: rec.fn("")})

	// Cursor is on the "Add new repo" row: nothing to update.
	m, cmd := send(m, runeKey('u'))
	if m.IsUpdating() {
		t.Error("u on the add row must not update")
	}
	_ = cmd

	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown}) // second repo: bitnami
	m, cmd = send(m, runeKey('u'))
	m = drain(m, cmd)
	if len(rec.names) != 1 || rec.names[0] != "bitnami" {
		t.Errorf("updated %v, want [bitnami]", rec.names)
	}
	if m.StatusMsg() != "updated bitnami" {
		t.Errorf("toast = %q", m.StatusMsg())
	}
}

func TestUpdateFromChartScreenTargetsSelectedRepo(t *testing.T) {
	rec := &updateRecorder{}
	m := loadedRepoModel(&helm.MockExecutor{RepoUpdateFunc: rec.fn("")})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, cmd := send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = drain(m, cmd)

	m, cmd = send(m, runeKey('R'))
	_ = drain(m, cmd)
	if len(rec.names) != 1 || rec.names[0] != "stable" {
		t.Errorf("updated %v, want [stable]", rec.names)
	}
}

func TestRemoveRepoFlow(t *testing.T) {
	var removed string
	m := loadedRepoModel(&helm.MockExecutor{
		RepoRemoveFunc: func(_ context.Context, n string) error { removed = n; return nil },
	})

	// Not on a repo row: refused.
	m, _ = send(m, runeKey('x'))
	if m.CurrentMode() != ModeNormal {
		t.Fatal("x on the add row must not open the dialog")
	}

	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, runeKey('x'))
	if m.CurrentMode() != ModeConfirmRemove || m.pendingRemove != "stable" {
		t.Fatalf("mode=%d pending=%q", m.CurrentMode(), m.pendingRemove)
	}

	// q must not quit while the dialog is open; n cancels.
	m, cmd := send(m, runeKey('q'))
	if cmd != nil || m.CurrentMode() != ModeConfirmRemove {
		t.Error("q should be inert in the confirm dialog")
	}
	m, _ = send(m, runeKey('n'))
	if m.CurrentMode() != ModeNormal || removed != "" {
		t.Error("n should cancel without removing")
	}

	m, _ = send(m, runeKey('x'))
	m, cmd = send(m, runeKey('y'))
	m = drain(m, cmd)
	if removed != "stable" || m.StatusMsg() != `repo "stable" removed` {
		t.Errorf("removed=%q toast=%q", removed, m.StatusMsg())
	}
}

func TestCopyShortcuts(t *testing.T) {
	var got string
	old := writeClipboard
	writeClipboard = func(s string) error { got = s; return nil }
	defer func() { writeClipboard = old }()

	m := loadedRepoModel(&helm.MockExecutor{
		SearchRepoFunc: func(context.Context, string) ([]helm.Chart, error) {
			return []helm.Chart{{Name: "stable/nginx", Version: "1.2.3"}}, nil
		},
		SearchRepoVersionsFunc: func(context.Context, string) ([]helm.Chart, error) {
			return []helm.Chart{{Name: "stable/nginx", Version: "1.2.3"}}, nil
		},
		ShowChartFunc: func(context.Context, string, string) (string, error) { return "name: nginx\n", nil },
	})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, runeKey('y'))
	if got != "https://charts.helm.sh/stable" {
		t.Errorf("repo copy = %q", got)
	}
	m, _ = send(m, runeKey('Y'))
	if got != "helm repo add stable https://charts.helm.sh/stable" {
		t.Errorf("repo cmd copy = %q", got)
	}

	m, cmd := send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = drain(m, cmd)
	m, _ = send(m, runeKey('y'))
	if got != "stable/nginx" {
		t.Errorf("chart copy = %q", got)
	}
	m, _ = send(m, runeKey('Y'))
	if got != "helm install nginx stable/nginx --version 1.2.3" {
		t.Errorf("install cmd = %q", got)
	}

	for range 2 { // chart list -> versions -> detail
		m, cmd = send(m, tea.KeyMsg{Type: tea.KeyEnter})
		m = drain(m, cmd)
	}
	m, _ = send(m, runeKey('y'))
	if got != "name: nginx\n" {
		t.Errorf("yaml copy = %q", got)
	}
}

func TestCopyFailureShowsError(t *testing.T) {
	old := writeClipboard
	writeClipboard = func(string) error { return errors.New("no clipboard") }
	defer func() { writeClipboard = old }()

	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, runeKey('y'))
	if !m.statusErr {
		t.Errorf("expected error toast, got %q", m.StatusMsg())
	}
}

func TestSortCycle(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	names := func() []string {
		var out []string
		for _, r := range m.filteredRepos {
			out = append(out, r.Name)
		}
		return out
	}
	want := [][]string{
		{"bitnami", "stable"}, // name asc
		{"stable", "bitnami"}, // name desc
		{"bitnami", "stable"}, // charts desc (9 vs 2)
		{"stable", "bitnami"}, // back to helm order
	}
	for i, w := range want {
		m, _ = send(m, runeKey('s'))
		got := names()
		if got[0] != w[0] || got[1] != w[1] {
			t.Errorf("step %d: %v, want %v", i, got, w)
		}
	}
	if m.repos[0].Name != "stable" {
		t.Error("sorting must not reorder the source data")
	}
}

func TestSortKeepsEnterMappingCorrect(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, runeKey('s')) // name asc: bitnami first
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.SelectedRepo().Name != "bitnami" {
		t.Errorf("selected %q, want bitnami", m.SelectedRepo().Name)
	}
}

func TestVimBackAndOpen(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, runeKey('h'))
	if m.CurrentScreen() != ScreenRepoList || m.CurrentMode() != ModeNormal {
		t.Error("h on the top screen must be a no-op")
	}
	m, _ = send(m, runeKey('j'))
	m, _ = send(m, runeKey('l'))
	if m.CurrentScreen() != ScreenChartList {
		t.Fatalf("l should drill in, screen=%d", m.CurrentScreen())
	}
	m, _ = send(m, runeKey('h'))
	if m.CurrentScreen() != ScreenRepoList {
		t.Error("h should go back")
	}
}

func TestFilterModeTypesNavigationLetters(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, runeKey('/'))
	for _, r := range "lhrRxyus" {
		m, _ = send(m, runeKey(r))
	}
	if m.CurrentMode() != ModeFilter || m.IsUpdating() || m.IsRefreshing() {
		t.Error("shortcut letters must be typed into the filter, not executed")
	}
	if m.FilterText() != "lhrRxyus" {
		t.Errorf("filter = %q", m.FilterText())
	}
}

func TestStaleResponseIgnoredAfterBack(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // loading charts
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyEsc})   // user bails out
	if m.IsLoading() {
		t.Fatal("going back must stop the loading state")
	}
	m, _ = send(m, chartsLoadedMsg([]helm.Chart{{Name: "stable/late"}}))
	if m.CurrentScreen() != ScreenRepoList || m.IsLoading() {
		t.Error("late response must not disturb the repo list")
	}
	m, _ = send(m, loadErrMsg{err: errors.New("late error")})
	if m.Error() != nil {
		t.Error("late error must be ignored")
	}
}

func TestOldToastTimerDoesNotClearNewToast(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, runeKey('s'))
	first := m.statusSeq
	m, _ = send(m, runeKey('s'))
	m, _ = send(m, clearStatusMsg{seq: first})
	if m.StatusMsg() == "" {
		t.Error("stale timer cleared a newer toast")
	}
	m, _ = send(m, toastFadeMsg{seq: m.statusSeq})
	if !m.statusDim || m.StatusMsg() == "" {
		t.Error("toast should dim before clearing")
	}
	m, _ = send(m, clearStatusMsg{seq: m.statusSeq})
	if m.StatusMsg() != "" {
		t.Error("toast should clear")
	}
}

func TestSpinnerStopsWhenIdleAndRestarts(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m.syncing = false
	m.loading = false
	m, cmd := send(m, spinner.TickMsg{})
	if cmd != nil || m.spinnerActive {
		t.Error("idle spinner should stop ticking")
	}
	m, cmd = send(m, runeKey('r'))
	if !m.spinnerActive || cmd == nil {
		t.Error("starting work must restart the spinner")
	}
}

func TestChartCountsFailureKeepsOldCounts(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, chartCountsLoadedMsg(nil))
	if m.chartCounts["bitnami"] != 9 {
		t.Error("a failed count query must not wipe known counts")
	}
}

func TestMouseWheelMovesCursor(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{})
	m, _ = send(m, tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if m.repoTable.Cursor() != 2 { // clamped: add row + 2 repos
		t.Errorf("cursor = %d, want 2", m.repoTable.Cursor())
	}
	m, _ = send(m, tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	if m.repoTable.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.repoTable.Cursor())
	}
}

func TestViewRendersEverywhere(t *testing.T) {
	m := loadedRepoModel(&helm.MockExecutor{
		SearchRepoFunc: func(context.Context, string) ([]helm.Chart, error) {
			return []helm.Chart{{Name: "stable/nginx", Version: "1.0.0", Description: "Web"}}, nil
		},
		ShowChartFunc: func(context.Context, string, string) (string, error) {
			return "# c\nname: nginx\n- a: b\n", nil
		},
	})
	for _, mode := range []Mode{ModeNormal, ModeHelp, ModeConfirmExit, ModeAddRepo, ModeConfirmRemove} {
		m.mode = mode
		if v := m.View(); v == "" {
			t.Errorf("mode %d rendered empty", mode)
		}
	}
	m.mode = ModeNormal
	m, _ = send(m, tea.KeyMsg{Type: tea.KeyDown})
	m, cmd := send(m, tea.KeyMsg{Type: tea.KeyEnter})
	_ = m.View() // loading
	m = drain(m, cmd)
	m, _ = send(m, runeKey('/'))
	for _, r := range "zzz" {
		m, _ = send(m, runeKey(r))
	}
	_ = m.View() // empty-filter state
	for _, size := range []tea.WindowSizeMsg{{Width: 20, Height: 6}, {Width: 200, Height: 60}} {
		m, _ = send(m, size)
		_ = m.View()
	}
}
