package simulator

import (
	"reflect"
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
