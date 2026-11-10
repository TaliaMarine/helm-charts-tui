package tui

import (
	"context"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// Screen identifies which screen is currently displayed.
type Screen int

const (
	ScreenRepoList Screen = iota
	ScreenChartList
	ScreenChartVersions
	ScreenChartDetail
)

// Mode identifies the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeFilter
	ModeAddRepo
	ModeConfirmExit
	ModeHelp
	ModeConfirmRemove
)

// maxParallelUpdates bounds concurrent `helm repo update NAME` processes.
const maxParallelUpdates = 4

// Toast timing: visible, then dimmed, then gone (3s total, matching the spec).
const (
	toastVisible = 2500 * time.Millisecond
	toastFade    = 500 * time.Millisecond
)

// Model is the Bubble Tea model for the application.
type Model struct {
	helm helm.Executor
	ctx  context.Context
	keys keyMap

	screen Screen
	mode   Mode

	width  int
	height int

	// Data
	repos       []helm.Repo
	chartCounts map[string]int
	charts      []helm.Chart
	versions    []helm.Chart
	detail      string

	// Filtered (and sorted) views; table row i maps to filtered*[i] (repos: i-1).
	filteredRepos    []helm.Repo
	filteredCharts   []helm.Chart
	filteredVersions []helm.Chart

	// Selections for breadcrumb
	selectedRepo    helm.Repo
	selectedChart   helm.Chart
	selectedVersion helm.Chart

	// Bubbles components
	repoTable    table.Model
	chartTable   table.Model
	versionTable table.Model
	viewport     viewport.Model

	// Filter
	filterInput textinput.Model
	filterText  string

	// Sort (per list screen)
	repoSort  int
	chartSort int

	// Add repo
	addRepoName  textinput.Model
	addRepoURL   textinput.Model
	addRepoFocus int // 0=name, 1=url

	// Remove repo confirmation
	pendingRemove string

	// Status
	loading      bool // nothing to show yet: full-screen spinner
	refreshing   bool // reloading while stale data stays visible
	silentReload bool // next finished reload shouldn't toast (a more specific toast is showing)
	syncing      bool // startup `helm repo update` running in the background
	err          error

	// Toast
	statusMsg string
	statusErr bool
	statusDim bool
	statusSeq int

	// User-triggered repo updates
	updating     bool
	updateTotal  int
	updateDone   int
	updateFailed []string
	updateErr    error
	updateSem    chan struct{}

	// Animation
	spinner       spinner.Model
	spinnerActive bool
	progress      progress.Model
}

// New creates a new TUI model.
func New(ctx context.Context, executor helm.Executor) Model {
	fi := textinput.New()
	fi.Prompt = "/"
	fi.CharLimit = 100

	nameInput := textinput.New()
	nameInput.Prompt = "Name: "
	nameInput.CharLimit = 100
	nameInput.Placeholder = "my-repo"

	urlInput := textinput.New()
	urlInput.Prompt = "URL:  "
	urlInput.CharLimit = 200
	urlInput.Placeholder = "https://charts.example.com"

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinStyle
	if noColor {
		sp.Spinner = spinner.Line // plain ASCII
	}

	fill := progress.WithScaledGradient("#5f5fff", "#00d787")
	if noColor {
		fill = progress.WithSolidFill("#ffffff")
	}
	pr := progress.New(progress.WithWidth(16), progress.WithoutPercentage(), fill)

	return Model{
		helm: executor,
		ctx:  ctx,
		keys: defaultKeyMap(),

		screen: ScreenRepoList,
		mode:   ModeNormal,

		repoTable:    newTable(),
		chartTable:   newTable(),
		versionTable: newTable(),
		viewport:     viewport.New(80, 20),

		filterInput: fi,

		addRepoName: nameInput,
		addRepoURL:  urlInput,

		loading: true,
		syncing: true, // Init fires the background repo update

		updateSem: make(chan struct{}, maxParallelUpdates),

		spinner:       sp,
		spinnerActive: true, // Init starts the tick chain
		progress:      pr,
	}
}

func newTable() table.Model {
	t := table.New(table.WithFocused(true))
	t.SetStyles(tableStyles())
	// Free u/d/b/f for app shortcuts; keep ctrl-based paging.
	km := table.DefaultKeyMap()
	km.PageUp = tableKey("pgup", "ctrl+b")
	km.PageDown = tableKey("pgdown", "ctrl+f")
	km.HalfPageUp = tableKey("ctrl+u")
	km.HalfPageDown = tableKey("ctrl+d")
	t.KeyMap = km
	return t
}

// Init starts loading repos and kicks off a background repo update.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadRepos(),
		m.loadChartCounts(),
		m.repoUpdateAll(),
		m.spinner.Tick,
	)
}

// busy reports whether anything is in flight that the spinner should reflect.
func (m Model) busy() bool {
	return m.loading || m.refreshing || m.syncing || m.updating
}

// wake (re)starts the spinner tick chain if it has stopped.
func (m *Model) wake() tea.Cmd {
	if m.spinnerActive {
		return nil
	}
	m.spinnerActive = true
	return m.spinner.Tick
}

// contentHeight returns the available height for content after header and status bar.
func (m Model) contentHeight() int {
	// header (2 lines) + status bar (1 line)
	h := m.height - 3
	if h < 5 {
		h = 5
	}
	return h
}

func (m *Model) syncSizes() {
	h := m.contentHeight()
	w := m.width

	m.repoTable.SetWidth(w)
	m.repoTable.SetHeight(h)
	m.chartTable.SetWidth(w)
	m.chartTable.SetHeight(h)
	m.versionTable.SetWidth(w)
	m.versionTable.SetHeight(h)
	m.viewport.Width = w
	m.viewport.Height = h
}

func (m *Model) setRows(t *table.Model, cols []table.Column, rows []table.Row) {
	cursor := t.Cursor()
	t.SetColumns(cols)
	t.SetRows(rows)
	t.SetCursor(cursor) // clamps when the list shrank
}

func (m *Model) rebuildRepoTable() {
	m.filteredRepos = sortRepos(FilterRepos(m.repos, m.filterText), m.repoSort, m.chartCounts)

	nameW := 24
	chartsW := 8
	urlW := m.width - nameW - chartsW - 8
	if urlW < 20 {
		urlW = 20
	}

	cols := []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "URL", Width: urlW},
		{Title: "Charts", Width: chartsW},
	}

	rows := make([]table.Row, 0, len(m.filteredRepos)+1)
	rows = append(rows, table.Row{"+ Add new repo", "", ""})

	for _, r := range m.filteredRepos {
		count := "..."
		if m.chartCounts != nil {
			if c, ok := m.chartCounts[r.Name]; ok {
				count = strconv.Itoa(c)
			}
		}
		rows = append(rows, table.Row{r.Name, r.URL, count})
	}

	m.setRows(&m.repoTable, cols, rows)
}

func chartColumns(width int) []table.Column {
	nameW := 30
	verW := 12
	appW := 12
	descW := width - nameW - verW - appW - 10
	if descW < 20 {
		descW = 20
	}
	return []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Version", Width: verW},
		{Title: "App Version", Width: appW},
		{Title: "Description", Width: descW},
	}
}

func chartRows(charts []helm.Chart) []table.Row {
	rows := make([]table.Row, 0, len(charts))
	for _, c := range charts {
		rows = append(rows, table.Row{c.Name, c.Version, c.AppVersion, c.Description})
	}
	return rows
}

func (m *Model) rebuildChartTable() {
	m.filteredCharts = sortCharts(FilterCharts(m.charts, m.filterText), m.chartSort)
	m.setRows(&m.chartTable, chartColumns(m.width), chartRows(m.filteredCharts))
}

func (m *Model) rebuildVersionTable() {
	m.filteredVersions = FilterCharts(m.versions, m.filterText)
	m.setRows(&m.versionTable, chartColumns(m.width), chartRows(m.filteredVersions))
}

func (m *Model) renderDetail() {
	m.viewport.SetContent(renderYAML(m.detail, m.width))
}

func (m *Model) applyFilter() {
	switch m.screen {
	case ScreenRepoList:
		m.rebuildRepoTable()
	case ScreenChartList:
		m.rebuildChartTable()
	case ScreenChartVersions:
		m.rebuildVersionTable()
	}
}

// Accessors for testing.

func (m Model) CurrentScreen() Screen     { return m.screen }
func (m Model) CurrentMode() Mode         { return m.mode }
func (m Model) IsLoading() bool           { return m.loading }
func (m Model) IsRefreshing() bool        { return m.refreshing }
func (m Model) IsUpdating() bool          { return m.updating }
func (m Model) Error() error              { return m.err }
func (m Model) StatusMsg() string         { return m.statusMsg }
func (m Model) FilterText() string        { return m.filterText }
func (m Model) SelectedRepo() helm.Repo   { return m.selectedRepo }
func (m Model) SelectedChart() helm.Chart { return m.selectedChart }
func (m Model) Repos() []helm.Repo        { return m.repos }
func (m Model) Charts() []helm.Chart      { return m.charts }
func (m Model) Versions() []helm.Chart    { return m.versions }
func (m Model) Detail() string            { return m.detail }
