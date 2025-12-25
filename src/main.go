package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakolb/bcd/src/crawler"
	"github.com/sakolb/bcd/src/entry"
	"github.com/sakolb/bcd/src/tui"
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

	p := tea.NewProgram(&model, tea.WithAltScreen())

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
			fmt.Println(selected)
		}
	}
}
