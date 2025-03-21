package main

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

type DetailedResultModel struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	dump       io.Writer
	w          int
	h          int
}

func (m DetailedResultModel) Init() tea.Cmd {
	return nil
}

func InitDetailedResultModel(dump io.Writer, result CommandResult) DetailedResultModel {
	return DetailedResultModel{
		Tabs: []string{"Errors", "Warnings", "Info", "Raw Std. Out"},
		TabContent: []string{
			strings.Join(result.Errors, " "),
			strings.Join(result.Warnings, " "),
			strings.Join(result.Infos, " "),
			result.RawOutput,
		},
		activeTab: 0,
		dump:      dump,
	}

}

func (m DetailedResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.h = msg.Height
		m.h = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "esc":
			return m, func() tea.Msg { return switchToMainModel{} }
		}
	}
	return m, nil
}

func (m DetailedResultModel) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
	return lipgloss.Place(m.w, m.h-20,lipgloss.Center, lipgloss.Left, doc.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
