package main

import (
	"io"
	"strconv"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

type ResultTableModel struct {
	table   table.Model
	results []CommandResult
	dump    io.Writer
}

func (m ResultTableModel) Init() tea.Cmd { return nil }

func InitResultsTable(dump io.Writer, results []CommandResult) ResultTableModel {
	columns := []table.Column{
		{Title: "Command Name", Width: 30},
		{Title: "Command Exec", Width: 30},
		{Title: "Result", Width: 30},
		{Title: "Errors", Width: 30},
		{Title: "Warnings", Width: 30},
	}
	rows := []table.Row{}
	for _, v := range results {
		rows = append(rows, table.Row{v.Name, v.Command, v.Passed, strconv.Itoa(len(v.Errors)), strconv.Itoa(len(v.Warnings))})
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(results)+4),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return ResultTableModel{
		table:   t,
		results: results,
		dump:    dump,
	}
}

func (m ResultTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.dump != nil {
		spew.Fdump(m.dump, fmt.Sprintf("Msg: %s, rows: %s", msg, m.table.Rows()))
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.table.SetHeight(msg.Height)
		m.table.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, func() tea.Msg { return switchToDetailedResultModel{m.results[m.table.Cursor()]} }
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ResultTableModel) View() string {
	return m.table.View()
}
