package main

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/davecgh/go-spew/spew"
)

var clearCache bool

type ConfirmForm struct {
	form             *huh.Confirm
	choice           string
	dump             io.Writer
	runCommands      bool
	commandsFinished bool
	err              error
}

func InitialConfirmModel(dump io.Writer, choice string) ConfirmForm {
	return ConfirmForm{
		form:   huh.NewConfirm().Title("Clear bazel cache & delete '_build'?").Value(&clearCache),
		choice: choice,
		dump:   dump,
	}
}

func (m ConfirmForm) Init() tea.Cmd {
	m.form.Run()
	return m.form.Init()
}

func (m ConfirmForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	if m.dump != nil {
		spew.Fdump(m.dump, fmt.Sprintf("Msg: %s, Value: %s", msg, m.form.GetValue()))
	}
	// We only check global keys here.
	// Ctrl+c for example
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.form.WithHeight(msg.Height)
		m.form.WithWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
			// This seems to never execute?
		case "esc":
			return m, func() tea.Msg { return switchToListModel{} }
		case "enter":
			m.runCommands = true
			cmds, err := loadConfig("commands.yaml")
			m.err = err
			runResults := runCommands(m.choice, cmds, m.form.GetValue().(bool))
			return m, func() tea.Msg { return switchToTableModel{runResults} }
		}
	}
	return m, nil
}

func (m ConfirmForm) View() string {
	if m.runCommands {
		return "RUNNING COMMANDS NOW"
	}
	return m.form.View()
}
