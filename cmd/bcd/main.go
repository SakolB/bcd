package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakolb/bcd/internal/crawler"
	"github.com/sakolb/bcd/internal/entry"
	"github.com/sakolb/bcd/internal/tui"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		baseDir, err = filepath.Abs(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	model := tui.InitModel(baseDir)

	// Check if stdout is redirected (e.g., in shell function)
	// If so, use /dev/tty for both input and output to receive resize signals
	var p *tea.Program
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// stdout is redirected, use /dev/tty for TUI to receive signals
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			defer tty.Close()
			p = tea.NewProgram(model, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
		} else {
			// Fallback if /dev/tty unavailable
			p = tea.NewProgram(model, tea.WithAltScreen())
		}
	} else {
		// stdout is a terminal, use normally
		p = tea.NewProgram(model, tea.WithAltScreen())
	}

	c := crawler.NewCrawler()
	go func() {
		go c.Crawl(baseDir)

		for path := range c.Paths() {
			e, err := entry.NewPathEntry(path, baseDir)
			if err != nil {
				continue
			}
			p.Send(tui.EntryMsg(e))
		}
		p.Send(tui.CrawlDoneMsg{})
	}()

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		if selected := m.Selected(); selected != "" {
			fmt.Fprintf(os.Stderr, "BCD_SELECTED_PATH:%s\n", selected)
		}
	}
}
