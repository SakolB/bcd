package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakolb/bcd/src/entry"
	"github.com/sakolb/bcd/src/ranker"
)

const maxVisibleResults = 15

type EntryMsg *entry.PathEntry

type CrawlDoneMsg struct{}

type Model struct {
	textInput textinput.Model
	ranker    *ranker.Ranker
	results   []ranker.ScoredEntry
	cursor    int
	selected  string
	quitting  bool
	baseDir   string

	mu sync.Mutex
}

func InitModel(baseDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		textInput: ti,
		ranker:    ranker.NewRanker(),
		results:   []ranker.ScoredEntry{},
		cursor:    0,
		baseDir:   baseDir,
	}
}

func (m Model) Selected() string {
	return m.selected
}

func (m *Model) AddEntry(e *entry.PathEntry) {
	m.mu.Lock()
	m.ranker.AddEntry(e)
	m.mu.Unlock()
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			m.selected = ""
			return m, tea.Quit

		case "enter":
			if len(m.results) > 0 && m.cursor < len(m.results) {
				m.selected = m.results[m.cursor].Entry.AbsPath
			}
			m.quitting = true
			return m, tea.Quit

		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return m, nil
		}

	case EntryMsg:
		m.mu.Lock()
		m.ranker.AddEntry(msg)
		m.ranker.SetQuery(m.textInput.Value())
		m.results = m.ranker.Results()
		m.mu.Unlock()
		m.clampCursor()
		return m, nil

	case CrawlDoneMsg:
		m.mu.Lock()
		m.ranker.SetQuery(m.textInput.Value())
		m.results = m.ranker.Results()
		m.mu.Unlock()
		m.clampCursor()
		return m, nil
	}

	prevValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)

	if m.textInput.Value() != prevValue {
		m.mu.Lock()
		m.ranker.SetQuery(m.textInput.Value())
		m.results = m.ranker.Results()
		m.mu.Unlock()
		m.cursor = 0
	}
	return m, cmd
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(" ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	total := len(m.results)
	b.WriteString(fmt.Sprintf("	%d results\n", total))
	b.WriteString("	-------------------------------------------\n")

	visible := m.results
	if len(visible) > maxVisibleResults {
		visible = visible[:maxVisibleResults]
	}

	for i, res := range visible {
		cursor := " "
		if i == m.cursor {
			cursor = "> "
		}

		displayPath := res.Entry.AbsPath
		if strings.HasPrefix(displayPath, m.baseDir) {
			displayPath = "." + strings.TrimPrefix(displayPath, m.baseDir)
		}

		if len(displayPath) > 60 {
			displayPath = "..." + displayPath[len(displayPath)-57:]
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, displayPath))
	}

	if len(m.results) > maxVisibleResults {
		b.WriteString(fmt.Sprintf("\n	... and %d more\n", len(m.results)-maxVisibleResults))
	}

	b.WriteString("\n, ↑/↓: navigate • enter: select • esc: quit\n")

	return b.String()
}
