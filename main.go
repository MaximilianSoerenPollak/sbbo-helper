package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
	"sync"
	"os/exec"
	"os"


	"github.com/charmbracelet/bubbles/list"
	//"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandResult struct {
	Name    string
	Command string
	Status  int
	Stdout  string
	Stderr  string
	Errors  string
}

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

type model struct {
	screen         int // 0: Select Option, 1: Confirm Cache, 2: Show Results, 3: Show Details
	list           list.Model
	results        []CommandResult
	selectedResult *CommandResult
	clearCache     bool
	buildOption    string // "Test" or "Build"
}

type CommandConfig struct {
        Commands []struct {
                Name    string `yaml:"name"`
                Command string `yaml:"command"`
        } `yaml:"commands"`
}

func loadConfig(filename string) (CommandConfig, error) {
        data, err := os.ReadFile(filename)
        if err != nil {
                return CommandConfig{}, err
        }

        var config CommandConfig
        err = yaml.Unmarshal(data, &config)
        return config, err
}

func (m model) Init() tea.Cmd {
	return nil
}

func initialModel() model {
	items := []list.Item{
		item{title: "Test"},
		item{title: "Build"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Build Option"
	return model{
		screen: 0,
		list:   l,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.screen {
		case 0: // Select "Test" or "Build"
			if msg.String() == "enter" {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.buildOption = i.title
					m.screen = 1 // Go to cache confirmation
					m.list.Title = "Clear Cache? (y/n)"
					m.list.SetItems([]list.Item{item{title: "yes"}, item{title: "no"}})
					return m, nil
				}
			}

		case 1: // Confirm Cache
			if msg.String() == "enter" {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.clearCache = i.title == "yes"
					m.screen = 2                           // Go to running
					m.list.Title = "Running Commands..."   //set a title
					m.list.SetItems([]list.Item{})         //clear the list
					return m, m.runCommands(m.buildOption) // Start commands
				}
			}
		case 2: //show results
			if msg.String() == "enter" {
				if i, ok := m.list.SelectedItem().(item); ok {
					for _, res := range m.results {
						if res.Name == i.Title() { //find the command
							m.selectedResult = &res
							m.screen = 3
							return m, nil
						}
					}

				}
			}

		case 3: //show details.
			if msg.String() == "esc" {
				m.selectedResult = nil
				m.screen = 2 //go back to results
				return m, nil
			}
		}

	case []CommandResult: // Results from command execution
		m.results = msg
		items := make([]list.Item, len(msg))
		for i, result := range msg {
			items[i] = item{title: result.Name} // Use command names in list
		}
		m.list.SetItems(items)
		m.screen = 2 // Go to results screen
		m.list.Title = "Command Results"
		return m, nil
	case error:
		//handle errors.
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.screen {
	case 0, 1, 2:
		return m.list.View()
	case 3: // Show command details
		if m.selectedResult != nil {
			return fmt.Sprintf("Command: %s\nStatus: %d\nErrors: %s\n\nParsed Stdout:\n%s\n\nRaw Stdout:\n%s\n\nPress esc to return.",
				m.selectedResult.Command, m.selectedResult.Status, m.selectedResult.Errors, m.selectedResult.Stdout, m.selectedResult.Stderr)
		} else {
			return "Error: No result selected." // Should not happen, but handle it
		}
	default:
		return "Unknown Screen"
	}
}

func (m model) runCommands(option string) tea.Cmd {
	return func() tea.Msg {
		config, err := loadConfig("commands.yaml") // Load from file
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err) // Return error
		}

		results := make([]CommandResult, 0)
		var wg sync.WaitGroup

		for _, cmdConfig := range config.Commands {
			// Filter based on selected option
			if (option == "Test" && strings.Contains(cmdConfig.Command, "test")) ||
				(option == "Build" && strings.Contains(cmdConfig.Command, "build")) {
				wg.Add(1)
				go func(cmdConfig struct {
					Name    string `yaml:"name"`
					Command string `yaml:"command"`
				}) {
					defer wg.Done()
					if m.clearCache {
						exec.Command("bazel", "clean").Run() // Ignore error for clean
					}
					cmd := exec.Command("bash", "-c", cmdConfig.Command)
					output, err := cmd.CombinedOutput()
					result := CommandResult{
						Name:    cmdConfig.Name,
						Command: cmdConfig.Command,
						Status:  cmd.ProcessState.ExitCode(),
						Stdout:  string(output),
						Stderr:  "",
						Errors:  "",
					}
					if err != nil {
						result.Stderr = string(output)
						errorRegex := regexp.MustCompile(`ERROR: .*`)
						result.Errors = errorRegex.FindString(string(output))
					}
					results = append(results, result)
				}(cmdConfig)
			}
		}

		wg.Wait()
		return results // Return the results
	}
}

func main() {
        p := tea.NewProgram(initialModel(), tea.WithAltScreen())
        if _, err := p.Run(); err != nil {
                fmt.Printf("Alas, there's been an error: %v", err)
                panic(err)
        }
}
