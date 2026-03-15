package tui

import (
	"context"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TaliaMarine/helm-charts-tui/internal/helm"
)

// Screen identifies which screen is currently displayed.
type Screen int

const (
	ScreenRepoList     Screen = iota
	ScreenChartList    Screen = iota
	ScreenChartVersions Screen = iota
	ScreenChartDetail  Screen = iota
)

// Mode identifies the current input mode.
type Mode int

const (
	ModeNormal      Mode = iota
	ModeFilter      Mode = iota
	ModeAddRepo     Mode = iota
	ModeConfirmExit Mode = iota
	ModeHelp        Mode = iota
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

	// Filtered views
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

	// Add repo
	addRepoName textinput.Model
	addRepoURL  textinput.Model
	addRepoFocus int // 0=name, 1=url

	// Status
	loading   bool
	err       error
	statusMsg string
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
	}
}

func newTable() table.Model {
	t := table.New(
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	t.SetStyles(s)
	return t
}

// Init starts loading repos.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadRepos(),
		m.loadChartCounts(),
	)
}

// contentHeight returns the available height for content after header and status bar.
func (m Model) contentHeight() int {
	// header (2 lines) + status bar (1 line) + margins
	h := m.height - 4
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

func (m *Model) rebuildRepoTable() {
	m.filteredRepos = FilterRepos(m.repos, m.filterText)

	nameW := 20
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
	rows = append(rows, table.Row{addRepoMarker.Render("+") + " Add new repo", "", ""})

	for _, r := range m.filteredRepos {
		count := "..."
		if m.chartCounts != nil {
			if c, ok := m.chartCounts[r.Name]; ok {
				count = strconv.Itoa(c)
			}
		}
		rows = append(rows, table.Row{r.Name, r.URL, count})
	}

	m.repoTable.SetColumns(cols)
	m.repoTable.SetRows(rows)
}

func (m *Model) rebuildChartTable() {
	m.filteredCharts = FilterCharts(m.charts, m.filterText)

	nameW := 30
	verW := 12
	appW := 12
	descW := m.width - nameW - verW - appW - 10
	if descW < 20 {
		descW = 20
	}

	cols := []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Version", Width: verW},
		{Title: "App Version", Width: appW},
		{Title: "Description", Width: descW},
	}

	rows := make([]table.Row, 0, len(m.filteredCharts))
	for _, c := range m.filteredCharts {
		rows = append(rows, table.Row{c.Name, c.Version, c.AppVersion, c.Description})
	}

	m.chartTable.SetColumns(cols)
	m.chartTable.SetRows(rows)
}

func (m *Model) rebuildVersionTable() {
	m.filteredVersions = FilterCharts(m.versions, m.filterText)

	nameW := 30
	verW := 12
	appW := 12
	descW := m.width - nameW - verW - appW - 10
	if descW < 20 {
		descW = 20
	}

	cols := []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Version", Width: verW},
		{Title: "App Version", Width: appW},
		{Title: "Description", Width: descW},
	}

	rows := make([]table.Row, 0, len(m.filteredVersions))
	for _, c := range m.filteredVersions {
		rows = append(rows, table.Row{c.Name, c.Version, c.AppVersion, c.Description})
	}

	m.versionTable.SetColumns(cols)
	m.versionTable.SetRows(rows)
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

func (m Model) CurrentScreen() Screen { return m.screen }
func (m Model) CurrentMode() Mode     { return m.mode }
func (m Model) IsLoading() bool       { return m.loading }
func (m Model) Error() error          { return m.err }
func (m Model) StatusMsg() string     { return m.statusMsg }
func (m Model) FilterText() string    { return m.filterText }
func (m Model) SelectedRepo() helm.Repo   { return m.selectedRepo }
func (m Model) SelectedChart() helm.Chart  { return m.selectedChart }
func (m Model) Repos() []helm.Repo        { return m.repos }
func (m Model) Charts() []helm.Chart       { return m.charts }
func (m Model) Versions() []helm.Chart     { return m.versions }
func (m Model) Detail() string             { return m.detail }
