package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"glowing-agent/simulator"
)

const version = "0.3.0"

type commandConfig struct {
	json          bool
	preset        string
	seed          string
	thinkingDepth string
	version       bool
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	config, taskArgs, err := parseCommand(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if config.version {
		_, err := fmt.Fprintf(stdout, "glowing-agent %s\n", version)
		return err
	}
	if config.json {
		return runJSON(config, taskArgs, stdin, stdout)
	}
	if config.preset != "" || config.seed != "" || config.thinkingDepth != "none" || len(taskArgs) > 0 {
		return errors.New("task and simulation settings belong in the TUI; use --json for automation")
	}
	if !isTerminal(stdin) || !isTerminal(stdout) {
		return errors.New("the TUI requires an interactive terminal; use --json for automation")
	}
	_, err = tea.NewProgram(newTUIModel(), tea.WithInput(stdin), tea.WithOutput(stdout)).Run()
	return err
}

func parseCommand(args []string, output io.Writer) (commandConfig, []string, error) {
	config := commandConfig{thinkingDepth: "none"}
	flags := flag.NewFlagSet("glowing-agent", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.BoolVar(&config.json, "json", false, "write one simulation as JSON instead of starting the TUI")
	flags.StringVar(&config.preset, "preset", "", "JSON mode: run a named preset")
	flags.StringVar(&config.seed, "seed", "", "JSON mode: use a reproducible signed 64-bit seed")
	flags.StringVar(&config.thinkingDepth, "thinking-depth", "none", "JSON mode: reasoning depth")
	flags.BoolVar(&config.version, "version", false, "print the version")
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: glowing-agent [--json [JSON flags] [task...]]")
		fmt.Fprintln(output, "")
		fmt.Fprintln(output, "Start the full-screen glowing-agent TUI. Use --json only for scripting.")
		fmt.Fprintln(output, "")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return commandConfig{}, nil, err
	}
	if !simulator.ValidThinkingDepth(config.thinkingDepth) {
		return commandConfig{}, nil, fmt.Errorf("unknown thinking depth %q", config.thinkingDepth)
	}
	return config, flags.Args(), nil
}

func runJSON(config commandConfig, taskArgs []string, stdin io.Reader, stdout io.Writer) error {
	task, err := resolveJSONTask(taskArgs, config.preset, stdin)
	if err != nil {
		return err
	}
	if count := utf8.RuneCountInString(task); count == 0 || count > 1000 {
		return errors.New("task text must be between 1 and 1000 characters")
	}
	var seed *int64
	if config.seed != "" {
		value, err := strconv.ParseInt(config.seed, 10, 64)
		if err != nil {
			return fmt.Errorf("--seed must be a signed 64-bit integer: %w", err)
		}
		seed = &value
	}
	return json.NewEncoder(stdout).Encode(simulator.GenerateWithThinkingDepth(task, seed, config.thinkingDepth))
}

func resolveJSONTask(taskArgs []string, presetID string, stdin io.Reader) (string, error) {
	if len(taskArgs) > 0 {
		return strings.TrimSpace(strings.Join(taskArgs, " ")), nil
	}
	if presetID != "" {
		preset, ok := simulator.PresetByID(presetID)
		if !ok {
			return "", fmt.Errorf("unknown preset %q", presetID)
		}
		return preset.Task, nil
	}
	value, err := io.ReadAll(io.LimitReader(stdin, 4001))
	if err != nil {
		return "", fmt.Errorf("read task: %w", err)
	}
	return strings.TrimSpace(string(value)), nil
}

func isTerminal(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
