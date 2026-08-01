package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
