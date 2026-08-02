package main

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"glowing-agent/simulator"
)

func updateTUI(t *testing.T, model tuiModel, msg tea.Msg) tuiModel {
	t.Helper()
	updated, _ := model.Update(msg)
	return updated.(tuiModel)
}

func key(code rune, modifier tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Mod: modifier})
}

func TestTUIStartsSimulationAndReplays(t *testing.T) {
	model := newTUIModel()
	model.task.SetValue("Fix a typo")
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: 100, Height: 30})
	model = updateTUI(t, model, key(tea.KeyEnter, 0))
	if model.state != running || model.simulation == nil || model.shown != 1 {
		t.Fatalf("simulation did not start: state=%d shown=%d", model.state, model.shown)
	}
	firstRun := model.runID
	model = updateTUI(t, model, key(' ', 0))
	if model.state != paused {
		t.Fatalf("state = %d, want paused", model.state)
	}
	shown := model.shown
	model = updateTUI(t, model, playbackTickMsg{runID: firstRun})
	if model.shown != shown {
		t.Fatal("paused model consumed a playback tick")
	}
	model = updateTUI(t, model, key('r', 0))
	if model.state != running || model.runID == firstRun || model.shown != 1 {
		t.Fatalf("replay did not reset correctly: state=%d run=%d shown=%d", model.state, model.runID, model.shown)
	}
	model = updateTUI(t, model, playbackTickMsg{runID: firstRun})
	if model.shown != 1 {
		t.Fatal("stale tick advanced replayed simulation")
	}
}

func TestTUINewTaskReturnsToEditing(t *testing.T) {
	model := newTUIModel()
	model.task.SetValue("Fix a typo")
	model.startSimulation()
	model = updateTUI(t, model, key('n', 0))
	if model.state != editing || model.shown != 0 || model.focus != focusTask || model.simulation != nil || model.task.Value() != "" {
		t.Fatalf("new task state = %+v", model)
	}
}

func TestTUIRejectsEmptyTask(t *testing.T) {
	model := newTUIModel()
	model.startSimulation()
	if model.err == "" || model.simulation != nil {
		t.Fatalf("expected task validation error: %+v", model)
	}
}

func TestTUIUsesShiftEnterForNewline(t *testing.T) {
	model := newTUIModel()
	model.task.SetValue("First line")
	model.task.Focus()
	model = updateTUI(t, model, key(tea.KeyEnter, tea.ModShift))
	if model.task.Value() != "First line\n" {
		t.Fatalf("shift+enter value = %q", model.task.Value())
	}
}

func TestTUISelectsPresetAndResizes(t *testing.T) {
	model := newTUIModel()
	model.selectPreset(1)
	if model.presetIndex != 0 || model.task.Value() == "" {
		t.Fatalf("preset selection failed: %+v", model)
	}
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: 100, Height: 30})
	if model.viewport.Width() == 0 || model.viewport.Height() < 5 {
		t.Fatalf("resize did not size viewport: %dx%d", model.viewport.Width(), model.viewport.Height())
	}
}

func TestTUICompletedViewKeepsSummaryAndHelpVisible(t *testing.T) {
	model := newTUIModel()
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: 100, Height: minimumTerminalHeight})
	model.task.SetValue("Fix a typo")
	model.startSimulation()
	model.shown = len(model.simulation.Events)
	model.state = completed
	model.refreshLog(true)

	view := model.View().Content
	if height := lipgloss.Height(view); height > model.height {
		t.Fatalf("rendered view is %d rows high in a %d-row terminal", height, model.height)
	}
	for _, expected := range []string{"RESULT", "Confidence", "space pause", "q quit"} {
		if !strings.Contains(view, expected) {
			t.Errorf("completed view does not contain %q", expected)
		}
	}
}

func TestTUILayoutMakesRoomForResultsAndRestoresLogHeight(t *testing.T) {
	model := newTUIModel()
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: 80, Height: 30})
	editingHeight := model.viewport.Height()
	model.task.SetValue("Fix a typo")
	model.startSimulation()
	runningHeight := model.viewport.Height()
	if runningHeight >= editingHeight {
		t.Fatalf("running viewport height = %d, want less than editing height %d", runningHeight, editingHeight)
	}

	model = updateTUI(t, model, key('n', 0))
	if model.viewport.Height() != editingHeight {
		t.Fatalf("new-task viewport height = %d, want restored height %d", model.viewport.Height(), editingHeight)
	}
}

func TestTUIRejectsTerminalBelowMinimumHeight(t *testing.T) {
	model := newTUIModel()
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: minimumTerminalWidth, Height: minimumTerminalHeight - 1})
	view := model.View().Content
	if !strings.Contains(view, "Please enlarge the terminal") || !strings.Contains(view, "24 rows") {
		t.Fatalf("small-terminal view did not report the minimum size: %q", view)
	}
}

func TestTUIThemeAndEventHierarchy(t *testing.T) {
	model := newTUIModel()
	model = updateTUI(t, model, tea.WindowSizeMsg{Width: 80, Height: 24})
	view := model.View()
	if view.BackgroundColor == nil || view.ForegroundColor == nil {
		t.Fatal("the themed TUI must set terminal foreground and background colors")
	}
	if got := color.RGBAModel.Convert(view.BackgroundColor); got != color.RGBAModel.Convert(colorCanvas) {
		t.Fatalf("background = %v, want %v", got, color.RGBAModel.Convert(colorCanvas))
	}
	for _, expected := range []string{"AGENT SESSION", "Waiting for an opportunity", "PRESET", "THINKING"} {
		if !strings.Contains(view.Content, expected) {
			t.Errorf("themed view does not contain %q", expected)
		}
	}

	events := []struct {
		name  string
		event simulator.Event
		want  string
	}{
		{name: "reasoning", event: simulator.Event{Kind: "thought", Text: "thinking"}, want: "◌ reasoning"},
		{name: "plan", event: simulator.Event{Kind: "plan", Text: "planning"}, want: "▸ plan"},
		{name: "tool", event: simulator.Event{Kind: "tool", Tool: &simulator.ToolCall{Name: "search", Status: "ok"}}, want: "▣ tool"},
		{name: "final", event: simulator.Event{Kind: "final good", Text: "done"}, want: "◆ final"},
		{name: "reveal", event: simulator.Event{Kind: "reveal", Text: "surprise"}, want: "◇ reveal"},
	}
	for _, test := range events {
		t.Run(test.name, func(t *testing.T) {
			if got := formatEvent(test.event); !strings.Contains(got, test.want) {
				t.Fatalf("formatEvent() = %q, want marker %q", got, test.want)
			}
		})
	}
}

func TestTUIUsesResponsiveCrushStyleSidebar(t *testing.T) {
	wide := newTUIModel()
	wide = updateTUI(t, wide, tea.WindowSizeMsg{Width: 100, Height: 30})
	for _, expected := range []string{"╱╱╱", "glowing", "-agent", "SIMULATION WORKBENCH", "RUN SETTINGS", "SESSION"} {
		if !strings.Contains(wide.View().Content, expected) {
			t.Errorf("wide view does not contain %q", expected)
		}
	}

	compact := newTUIModel()
	compact = updateTUI(t, compact, tea.WindowSizeMsg{Width: 80, Height: 24})
	if strings.Contains(compact.View().Content, "RUN SETTINGS") {
		t.Fatal("compact view should fold the sidebar into inline controls")
	}
	if !strings.Contains(compact.View().Content, "PRESET") {
		t.Fatal("compact view should keep run settings accessible")
	}
}

func TestTUIViewFitsSupportedTerminalSizes(t *testing.T) {
	for _, size := range []struct{ width, height int }{{60, 24}, {80, 24}, {100, 30}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			model := newTUIModel()
			model = updateTUI(t, model, tea.WindowSizeMsg{Width: size.width, Height: size.height})
			model.task.SetValue("Fix a typo")
			model.startSimulation()
			model.shown = len(model.simulation.Events)
			model.state = completed
			model.refreshLog(true)

			view := model.View().Content
			if got := lipgloss.Height(view); got > size.height {
				t.Fatalf("view height = %d, terminal height = %d", got, size.height)
			}
			if got := lipgloss.Width(view); got > size.width {
				t.Fatalf("view width = %d, terminal width = %d", got, size.width)
			}
		})
	}
}
