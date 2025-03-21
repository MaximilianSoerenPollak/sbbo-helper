package main

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

type switchToMainModel struct{}
type switchToListModel struct{}
type switchToConfirmModel struct{ choice string }
type switchToTableModel struct{ results []CommandResult }
type switchToDetailedResultModel struct{ result CommandResult }

type ParsedOutput struct {
	Warnings string
	Debugs   string
	Infos    string
	Errors   string
}

type MainModel struct {
	activeModel tea.Model
	dump        io.Writer
}

func initialMainModel(dump io.Writer) MainModel {
	return MainModel{
		activeModel: InitiallistModel(dump),
		dump:        dump,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) View() string {
	// fmt.Printf("MODEL RENDERING: %s", m.activeModel)
	return m.activeModel.View()
}

// parseCommandOutput extracts different log types from command output

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case switchToListModel:
		m.activeModel = InitiallistModel(m.dump)
		return m, m.activeModel.Init()
	case switchToTableModel:
		m.activeModel = InitResultsTable(m.dump, msg.results)
		return m, m.activeModel.Init()
	case switchToDetailedResultModel:
		m.activeModel = InitDetailedResultModel(m.dump, msg.result)
		return m, m.activeModel.Init()
	case switchToConfirmModel:
		// Cause it's a huh.Confirm not a form we have to go this route
		m.activeModel = InitialConfirmModel(m.dump, msg.choice)
		return m, m.activeModel.Init()
	}
	m.activeModel, cmd = m.activeModel.Update(msg)
	return m, cmd
}
