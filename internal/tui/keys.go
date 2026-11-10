package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	HalfUp    key.Binding
	HalfDown  key.Binding
	Home      key.Binding
	End       key.Binding
	Enter     key.Binding // dialogs and filter input
	Open      key.Binding // normal mode drill-in
	Back      key.Binding
	Escape    key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
	Filter    key.Binding
	Help      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding

	Reload    key.Binding
	UpdateAll key.Binding
	UpdateOne key.Binding
	AddRepo   key.Binding
	Remove    key.Binding
	Copy      key.Binding
	CopyCmd   key.Binding
	Sort      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		PageUp:    key.NewBinding(key.WithKeys("pgup", "ctrl+b"), key.WithHelp("pgup", "page up")),
		PageDown:  key.NewBinding(key.WithKeys("pgdown", "ctrl+f"), key.WithHelp("pgdn", "page down")),
		HalfUp:    key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "half page up")),
		HalfDown:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "half page down")),
		Home:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", "go to top")),
		End:       key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", "go to bottom")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Open:      key.NewBinding(key.WithKeys("enter", "l", "right"), key.WithHelp("enter/l/→", "open")),
		Back:      key.NewBinding(key.WithKeys("h", "left", "backspace"), key.WithHelp("h/←/bksp", "back")),
		Escape:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back / clear filter")),
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Filter:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		ShiftTab:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),

		Reload:    key.NewBinding(key.WithKeys("r", "f5", "ctrl+r"), key.WithHelp("r/F5", "reload this view")),
		UpdateAll: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "update repos from remote (all / current)")),
		UpdateOne: key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "update highlighted repo")),
		AddRepo:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add repo")),
		Remove:    key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x/del", "remove highlighted repo")),
		Copy:      key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "copy name / URL / YAML")),
		CopyCmd:   key.NewBinding(key.WithKeys("Y"), key.WithHelp("Y", "copy helm install command")),
		Sort:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle sort order")),
	}
}

type helpSection struct {
	title string
	binds []key.Binding
}

// helpSections returns the key bindings to display in the help overlay.
func helpSections() []helpSection {
	km := defaultKeyMap()
	return []helpSection{
		{"Navigate", []key.Binding{km.Up, km.Down, km.PageUp, km.PageDown, km.HalfUp, km.HalfDown, km.Home, km.End, km.Open, km.Back, km.Escape}},
		{"Refresh", []key.Binding{km.Reload, km.UpdateAll, km.UpdateOne}},
		{"Repos & lists", []key.Binding{km.Filter, km.Sort, km.AddRepo, km.Remove, km.Copy, km.CopyCmd}},
		{"General", []key.Binding{km.Help, km.Quit, km.ForceQuit}},
	}
}

func tableKey(keys ...string) key.Binding { return key.NewBinding(key.WithKeys(keys...)) }
