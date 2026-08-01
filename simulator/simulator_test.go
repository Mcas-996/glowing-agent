package simulator

import (
	"reflect"
	"strings"
	"testing"
)

func TestGenerateIsDeterministicForSeed(t *testing.T) {
	seed := int64(42)
	first := Generate("Fix the typo", &seed)
	second := Generate("Fix the typo", &seed)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("expected identical simulations for task and seed")
	}
}

func TestEveryEndingCanBeReached(t *testing.T) {
	found := map[string]bool{}
	for seed := int64(0); seed < 1000 && len(found) < 3; seed++ {
		result := Generate("Add a button", &seed)
		found[result.Ending] = true
	}
	for _, ending := range []string{"confident-miss", "scope-singularity", "accidental-win"} {
		if !found[ending] {
			t.Fatalf("ending %q was not generated", ending)
		}
	}
}

func TestPresetByID(t *testing.T) {
	preset, ok := PresetByID("button")
	if !ok || preset.Task == "" {
		t.Fatal("expected button preset")
	}
}

func TestThinkingDepthExpandsThoughtsAndTokens(t *testing.T) {
	seed := int64(42)
	none := GenerateWithThinkingDepth("Fix the typo", &seed, "none")
	high := GenerateWithThinkingDepth("Fix the typo", &seed, "high")
	thoughtAt := func(simulation Simulation, index int) Event {
		thoughts := make([]Event, 0)
		for _, event := range simulation.Events {
			if event.Kind == "thought" {
				thoughts = append(thoughts, event)
			}
		}
		return thoughts[index]
	}
	if got, want := thoughtAt(high, 0).DelayMS, thoughtAt(none, 0).DelayMS*8; got != want {
		t.Fatalf("high thought delay = %d, want %d", got, want)
	}
	if got := len(strings.Split(thoughtAt(high, 0).Text, "\n")); got != 8 {
		t.Fatalf("high thought has %d phrases, want 8", got)
	}
	if got, want := high.Metrics.TokensBurned, none.Metrics.TokensBurned*8; got != want {
		t.Fatalf("high tokens = %d, want %d", got, want)
	}
}

func TestThinkingDepthIsDeterministic(t *testing.T) {
	seed := int64(42)
	first := GenerateWithThinkingDepth("Fix the typo", &seed, "max")
	second := GenerateWithThinkingDepth("Fix the typo", &seed, "max")
	if !reflect.DeepEqual(first, second) {
		t.Fatal("expected identical simulations for task, seed, and thinking depth")
	}
}

func TestValidThinkingDepth(t *testing.T) {
	if !ValidThinkingDepth("ultra") || ValidThinkingDepth("deep") {
		t.Fatal("unexpected thinking depth validation")
	}
}
