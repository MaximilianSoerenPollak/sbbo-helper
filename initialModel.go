package main

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

var t = &testing.T{}

type ListModel struct {
	listM  list.Model
	choice string
	// TODO: Make this better.
	modelNr int // 0 = default, 1 == detailed
	err     error
	dump    io.Writer
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func InitiallistModel(dump io.Writer) ListModel {
	items := []list.Item{
		item("New run"),
		item("View last run"),
	}
	l := list.New(items, itemDelegate{}, 10, 10)

	l.Title = "Select what to do"
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return ListModel{
		listM:   l,
		modelNr: 0,
		dump:    dump,
	}
}

func detailedListModel(dump io.Writer) ListModel {
	cmds, err := loadConfig("commands.yaml")
	commands := getAllCommands(cmds)
	items := []list.Item{item("Test All"), item("Build All")}
	for _, v := range commands {
		items = append(items, item(v.Name))
	}

	l := list.New(items, itemDelegate{}, 0, 20)
	l.Title = "What command do you want to run?"
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	w := lipgloss.Width(l.Title)
	l.SetWidth(w + 5)
	return ListModel{
		listM:   l,
		err:     err,
		modelNr: 1,
		dump:    dump,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

// TODO: Add a progress bar
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.dump != nil {
		spew.Fdump(m.dump, fmt.Sprintf("Message: %+v, Current List: %s, Current ModelNr: %d", msg, m.listM.Items(), m.modelNr))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := ListStyle.GetFrameSize()
		m.listM.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.listM.SelectedItem().(item)
			spew.Fdump(m.dump, fmt.Sprintf("SELECTED ITEM: %s, OK: %s", i, ok))
			if ok {
				m.choice = string(i)
			}
			if m.choice == "New run" && m.modelNr == 0 {
				return detailedListModel(m.dump), nil
			}
			if m.modelNr == 1 {
				return m, func() tea.Msg { return switchToConfirmModel{m.choice} }
			}

		}
	}
	var cmd tea.Cmd
	m.listM, cmd = m.listM.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	if m.choice != "View last run" && m.choice != "" {
		return "Sorry not implemented yet"
	}
	return m.listM.View()
}
