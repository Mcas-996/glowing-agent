package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"glowing-agent/simulator"
)

func TestJSONModeUsesTaskAndSeed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--json", "--seed", "7", "--thinking-depth", "high", "Fix a typo"}, strings.NewReader("ignored"), &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	var result simulator.Simulation
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	if result.Task != "Fix a typo" || result.Seed != 7 || result.ThinkingDepth != "high" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestJSONModeReadsTaskFromStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--json", "--seed", "5"}, strings.NewReader("Add a button\n"), &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	var result simulator.Simulation
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Task != "Add a button" {
		t.Fatalf("got task %q", result.Task)
	}
}

func TestCommandRejectsTUISettings(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--seed", "7"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "TUI") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJSONModeRejectsInvalidInput(t *testing.T) {
	for _, args := range [][]string{{"--json", "--thinking-depth", "deep", "task"}, {"--json", "--seed", "not-a-number", "task"}, {"--json", "--preset", "missing"}} {
		var stdout, stderr bytes.Buffer
		if err := run(args, strings.NewReader(""), &stdout, &stderr); err == nil {
			t.Fatalf("expected an error for %v", args)
		}
	}
}
