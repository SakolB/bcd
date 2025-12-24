package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakolb/bcd/src/entry"
	"github.com/sakolb/bcd/src/ranker"
)

const maxVisibleResults = 50

type EntryMsg *entry.PathEntry

type CrawlDoneMsg struct{}

type QueryUpdateMsg struct {
	query string
}

type RankerCmd struct {
	AddEntry *entry.PathEntry
	SetQuery *string
}

type ResultsUpdateMsg struct {
	results []ranker.ScoredEntry
}

type Model struct {
	textInput      textinput.Model
	ranker         *ranker.Ranker
	results        []ranker.ScoredEntry
	cursor         int
	viewportOffset int
	selected       string
	quitting       bool
	baseDir        string

	pendingQuery string
	activeQuery  string

	rankerCmdChan    chan RankerCmd
	rankerResultChan chan ResultsUpdateMsg

	mu sync.Mutex
}

func InitModel(baseDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 100

	cmdChan := make(chan RankerCmd, 1000)
	resultChan := make(chan ResultsUpdateMsg, 1)

	return Model{
		textInput:        ti,
		ranker:           ranker.NewRanker(),
		results:          []ranker.ScoredEntry{},
		cursor:           0,
		viewportOffset:   0,
		baseDir:          baseDir,
		pendingQuery:     "",
		activeQuery:      "",
		rankerCmdChan:    cmdChan,
		rankerResultChan: resultChan,
	}
}

func debounceQueryCmd(query string, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return QueryUpdateMsg{query: query}
	})
}

func startRankerWorker(r *ranker.Ranker, cmdChan chan RankerCmd, resultChan chan ResultsUpdateMsg) {
	go func() {
		for cmd := range cmdChan {
			if cmd.AddEntry != nil {
				// Just add entry, don't send results yet
				r.AddEntry(cmd.AddEntry)
			}
			if cmd.SetQuery != nil {
				// Score everything, then send ONE complete result
				r.SetQuery(*cmd.SetQuery)
				resultChan <- ResultsUpdateMsg{results: r.Results()}
			}
		}
	}()
}

func waitForRankerResult(resultChan chan ResultsUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		return <-resultChan
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
	// Start the ranker worker goroutine
	startRankerWorker(m.ranker, m.rankerCmdChan, m.rankerResultChan)

	// Start listening for results and blinking cursor
	return tea.Batch(
		textinput.Blink,
		waitForRankerResult(m.rankerResultChan),
	)
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
				if m.cursor < m.viewportOffset {
					m.viewportOffset = m.cursor
				}
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor < len(m.results)-1 {
				m.cursor++
				if m.cursor >= m.viewportOffset+maxVisibleResults {
					m.viewportOffset = m.cursor - maxVisibleResults + 1
				}
			}
			return m, nil
		}

	case EntryMsg:
		// Send entry to ranker worker (non-blocking, just appends)
		select {
		case m.rankerCmdChan <- RankerCmd{AddEntry: msg}:
		default:
			// Channel full, skip this entry
		}
		return m, nil

	case CrawlDoneMsg:
		// Trigger a final query to score all collected entries
		query := m.textInput.Value()
		m.rankerCmdChan <- RankerCmd{SetQuery: &query}
		return m, nil

	case QueryUpdateMsg:
		// Only update if this query is still pending (user hasn't typed more)
		if msg.query == m.textInput.Value() {
			query := msg.query
			// Send query to ranker worker (non-blocking)
			// Worker will score and send back complete results
			select {
			case m.rankerCmdChan <- RankerCmd{SetQuery: &query}:
			default:
			}
			m.activeQuery = msg.query
			m.cursor = 0
			m.viewportOffset = 0
		}
		return m, nil

	case ResultsUpdateMsg:
		// Received complete results from ranker worker
		m.results = msg.results
		m.clampCursor()
		// Keep listening for more results
		return m, waitForRankerResult(m.rankerResultChan)
	}

	prevValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)

	if m.textInput.Value() != prevValue {
		// Don't update ranker immediately - schedule it
		m.pendingQuery = m.textInput.Value()
		return m, tea.Batch(cmd, debounceQueryCmd(m.textInput.Value(), 100*time.Millisecond))
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

	b.WriteString(fmt.Sprintf(" cwd: %s\n", m.baseDir))
	b.WriteString(" ")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	total := len(m.results)
	b.WriteString(fmt.Sprintf("	%d results\n", total))
	b.WriteString("	-------------------------------------------\n")

	end := m.viewportOffset + maxVisibleResults
	if end > len(m.results) {
		end = len(m.results)
	}
	visible := m.results[m.viewportOffset:end]

	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("22"))

	for i, res := range visible {
		cursor := "  "
		displayPath := res.Entry.AbsPath

		if len(displayPath) > 100 {
			displayPath = "..." + displayPath[len(displayPath)-97:]
		}

		line := displayPath
		if i+m.viewportOffset == m.cursor {
			cursor = "> "
			line = highlightStyle.Render(displayPath)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, line))
	}

	if len(m.results) > maxVisibleResults {
		b.WriteString(fmt.Sprintf("\n	... and %d more\n", len(m.results)-maxVisibleResults))
	}

	b.WriteString("\n, ↑/↓: navigate • enter: select • esc: quit\n")

	return b.String()
}
