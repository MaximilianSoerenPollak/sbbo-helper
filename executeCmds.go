package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

type CommandResult struct {
	Name       string
	Command    string
	Status     int
	Passed     string
	RawOutput  string       // Raw output from command
	Warnings   []string     // Lines starting with [WARNING]
	Debugs     []string     // Lines starting with [DEBUG]
	Infos      []string     // Lines starting with [INFO]
	Errors     []string     // ERROR lines from output
	ParsedLogs ParsedOutput // Structured output
}

type CommandConfig struct {
	Commands []map[string][]Command `yaml:"commands"`
}

type Command struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
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

func runCommands(choice string, commands CommandConfig, clearCache bool) []CommandResult {

	var results []CommandResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	filteredCommands := make([]Command, 0)

	// Filter commands based on the selected option
	if choice == "Test All" {
		filteredCommands = append(filteredCommands, getCommandsByCategory(commands, "Test")...)
	} else if choice == "Build All" {
		filteredCommands = append(filteredCommands, getCommandsByCategory(commands, "Build")...)
	} else {
		filteredCommands = append(filteredCommands, findCommandByName(commands, choice))
	}

	results = make([]CommandResult, 0, len(filteredCommands))

	for _, cmdConfig := range filteredCommands {
		wg.Add(1)
		go func(cmdConfig Command) {
			defer wg.Done()

			// Clear cache if requested
			if clearCache {
				cleanCmd := exec.Command("bazel", "clean", "&&", "rm", "-r", "_build")
				cleanCmd.Run() // Ignore errors from clean
			}

			// Run the command
			cmd := exec.Command("bash", "-c", cmdConfig.Command)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stdout
			err := cmd.Run()

			output := stdout.String()
			passed, parsed, warnings, debugs, infos, errors := parseCommandOutput(output)

			result := CommandResult{
				Name:       cmdConfig.Name,
				Command:    cmdConfig.Command,
				RawOutput:  output,
				Passed:     passed,
				Status:     0,
				Warnings:   warnings,
				Debugs:     debugs,
				Infos:      infos,
				Errors:     errors,
				ParsedLogs: parsed,
			}

			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					result.Status = exitErr.ExitCode()
				} else {
					result.Status = 999 // Generic error code
				}
			}

			// Safely append to results
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(cmdConfig)
	}

	wg.Wait()

	// Sort results by name for consistency
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	// Return the results
	return results
}

func parseCommandOutput(output string) (string, ParsedOutput, []string, []string, []string, []string) {
	var warnings, debugs, infos, errors []string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[WARNING]") {
			warnings = append(warnings, line)
		} else if strings.HasPrefix(line, "[DEBUG]") {
			debugs = append(debugs, line)
		} else if strings.HasPrefix(line, "[INFO]") {
			infos = append(infos, line)
		}
	}

	// Extract ERROR lines using regex
	errorRegex := regexp.MustCompile(`(?m)^ERROR: .*$`)
	errorMatches := errorRegex.FindAllString(output, -1)
	if errorMatches != nil {
		errors = errorMatches
	}

	// Create formatted strings for each log type
	parsed := ParsedOutput{
		Warnings: strings.Join(warnings, "\n"),
		Debugs:   strings.Join(debugs, "\n"),
		Infos:    strings.Join(infos, "\n"),
		Errors:   strings.Join(errors, "\n"),
	}
	passed := "\u2705" // ✅
	if len(warnings) > 0 {
		passed = "\u274C" // ❌
	}
	if len(errors) > 0 {
		passed = "\u26D4" // ⛔
	}

	return passed, parsed, warnings, debugs, infos, errors
}

func getCommandsByCategory(conf CommandConfig, category string) []Command {
	for _, categoryMap := range conf.Commands {
		for cat, commands := range categoryMap {
			if cat == category {
				return commands
			}
		}
	}
	return nil
}

func getAllCommands(conf CommandConfig) []Command {
	var allCommands []Command
	for _, categoryMap := range conf.Commands {
		for _, commands := range categoryMap {
			allCommands = append(allCommands, commands...)
		}
	}
	return allCommands
}

func findCommandByName(conf CommandConfig, name string) Command {
	allCommands := getAllCommands(conf)
	idx := slices.IndexFunc(allCommands, func(c Command) bool { return c.Name == name })
	// This should never return -1
	assert.Assert(t, idx != -1, fmt.Sprintf("Somehow name %s was not found in commands %v", name, conf))
	return allCommands[idx]
}
