package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	Home      key.Binding
	End       key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
	Filter    key.Binding
	Help      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "move up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "move down")),
		PageUp:    key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown:  key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		Home:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home/g", "go to top")),
		End:       key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end/G", "go to bottom")),
		Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Escape:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Filter:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		ShiftTab:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	}
}

// helpBindings returns the key bindings to display in the help overlay.
func helpBindings() []key.Binding {
	km := defaultKeyMap()
	return []key.Binding{
		km.Up, km.Down, km.PageUp, km.PageDown,
		km.Home, km.End, km.Enter, km.Escape,
		km.Filter, km.Help, km.Quit, km.ForceQuit,
	}
}
