package tui

import "github.com/atotto/clipboard"

// writeClipboard is a variable so tests can avoid touching the real clipboard.
var writeClipboard = clipboard.WriteAll
