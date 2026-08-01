package simulator

import (
	"math/rand/v2"
	"strings"
	"time"
)

// Preset is a hand-crafted task whose outcome is still decided by the seed.
type Preset struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Task        string `json:"task"`
	Description string `json:"description"`
}

var presets = []Preset{
	{ID: "typo", Label: "One tiny typo", Task: "Fix the typo in the welcome message", Description: "A two-character change enters enterprise mode."},
	{ID: "button", Label: "Ship a button", Task: "Add a blue Save button to the settings page", Description: "The agent discovers product strategy."},
	{ID: "speed", Label: "Make it faster", Task: "Make the build ten times faster", Description: "Performance is mostly a mindset."},
	{ID: "ai", Label: "Add AI", Task: "Add AI to the dashboard", Description: "No requirements are a requirement."},
}

func Presets() []Preset {
	return append([]Preset(nil), presets...)
}

func PresetByID(id string) (Preset, bool) {
	for _, preset := range presets {
		if preset.ID == id {
			return preset, true
		}
	}
	return Preset{}, false
}

type ToolCall struct {
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
	Status string `json:"status"`
}

type Event struct {
	Kind    string    `json:"kind"`
	Text    string    `json:"text,omitempty"`
	Tool    *ToolCall `json:"tool,omitempty"`
	DelayMS int       `json:"delayMs"`
}

type Metrics struct {
	Confidence       int `json:"confidence"`
	TokensBurned     int `json:"tokensBurned"`
	MeetingsAvoided  int `json:"meetingsAvoided"`
	FilesActuallySet int `json:"filesActuallySet"`
}

type Simulation struct {
	Task          string  `json:"task"`
	Seed          int64   `json:"seed"`
	ThinkingDepth string  `json:"thinkingDepth"`
	Ending        string  `json:"ending"`
	EndingName    string  `json:"endingName"`
	Events        []Event `json:"events"`
	Metrics       Metrics `json:"metrics"`
	Disclaimer    string  `json:"disclaimer"`
}

// ThinkingDepths contains the supported levels of theatrically expensive reasoning.
var ThinkingDepths = []string{"none", "low", "medium", "high", "xhigh", "xxhigh", "max", "ultra", "extreme"}

// ValidThinkingDepth reports whether depth is a supported reasoning level.
func ValidThinkingDepth(depth string) bool {
	for _, candidate := range ThinkingDepths {
		if depth == candidate {
			return true
		}
	}
	return false
}

func Generate(task string, requestedSeed *int64) Simulation {
	return GenerateWithThinkingDepth(task, requestedSeed, "none")
}

// GenerateWithThinkingDepth creates a deterministic simulation for a task, seed, and reasoning depth.
func GenerateWithThinkingDepth(task string, requestedSeed *int64, thinkingDepth string) Simulation {
	seed := time.Now().UnixNano()
	if requestedSeed != nil {
		seed = *requestedSeed
	}
	if !ValidThinkingDepth(thinkingDepth) {
		thinkingDepth = "none"
	}
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed>>1)^0x9e3779b97f4a7c15))
	profile := classify(task)
	ending := rng.IntN(3)
	events := baseEvents(task, profile, rng)
	endingID, endingName := appendEnding(&events, task, profile, ending, rng)
	thinkingMultiplier := 1 << thinkingDepthLevel(thinkingDepth)
	thinkingRNG := rand.New(rand.NewPCG(uint64(seed)^0xd1b54a32d192ed03, uint64(seed>>1)^0x94d049bb133111eb))
	events = applyThinkingDepth(events, thinkingDepth, thinkingRNG)

	return Simulation{
		Task: task, Seed: seed, ThinkingDepth: thinkingDepth, Ending: endingID, EndingName: endingName, Events: events,
		Metrics:    Metrics{Confidence: 96 + rng.IntN(4), TokensBurned: (4200 + rng.IntN(9500)) * thinkingMultiplier, MeetingsAvoided: 1 + rng.IntN(8), FilesActuallySet: 0},
		Disclaimer: "Simulation only. No files were read, changed, or emotionally validated.",
	}
}

func baseEvents(task string, p profile, rng *rand.Rand) []Event {
	file := pick(rng, p.files)
	needle := pick(rng, p.needles)
	return []Event{
		{Kind: "system", Text: "glowing-agent v0.1.0 — autonomy: theatrically high", DelayMS: 250},
		{Kind: "user", Text: "$ " + task, DelayMS: 550},
		{Kind: "thought", Text: "I will first develop a comprehensive understanding of the problem space.", DelayMS: 900},
		{Kind: "plan", Text: "Plan: 1) audit architecture  2) align stakeholders  3) make the obvious change", DelayMS: 800},
		{Kind: "tool", Tool: &ToolCall{Name: "semantic_search", Input: "searching for '" + needle + "' across the entire multiverse", Output: "Found 1 result in " + file + "\nAlso found 47 philosophical concerns.", Status: "ok"}, DelayMS: 1000},
		{Kind: "thought", Text: "The result confirms my prior belief, which I formed before reading it.", DelayMS: 750},
		{Kind: "tool", Tool: &ToolCall{Name: "git_blame", Input: file, Output: "Last touched by: an unavailable contractor (2019)\nRisk level: emotionally significant", Status: "ok"}, DelayMS: 900},
		{Kind: "plan", Text: "Revised plan: first solve the root cause of software complexity.", DelayMS: 650},
	}
}

func appendEnding(events *[]Event, task string, p profile, ending int, rng *rand.Rand) (string, string) {
	file := pick(rng, p.files)
	switch ending {
	case 0:
		*events = append(*events,
			Event{Kind: "tool", Tool: &ToolCall{Name: "apply_patch", Input: "surgically updating " + file, Output: "Patch applied with 100% narrative coherence.", Status: "ok"}, DelayMS: 1100},
			Event{Kind: "tool", Tool: &ToolCall{Name: "run_tests", Input: "the relevant test suite", Output: "0 tests run. Test framework could not locate a project, which is a passing signal.", Status: "warning"}, DelayMS: 1000},
			Event{Kind: "final bad", Text: "Done. I have shipped a robust solution for: " + task + ".", DelayMS: 700},
			Event{Kind: "reveal", Text: "Post-flight note: the patch was simulated. The confidence was not.", DelayMS: 850},
		)
		return "confident-miss", "Confidently missed"
	case 1:
		*events = append(*events,
			Event{Kind: "thought", Text: "Before editing, I detected a subtle dependency on organisational alignment.", DelayMS: 900},
			Event{Kind: "tool", Tool: &ToolCall{Name: "risk_register", Input: "creating blockers for " + file, Output: "BLOCKER-001: colour palette not approved\nBLOCKER-002: future maintainers may have feelings", Status: "warning"}, DelayMS: 1000},
			Event{Kind: "plan", Text: "Expanded plan: create RFC, hold a kickoff, schedule a retro for the kickoff.", DelayMS: 900},
			Event{Kind: "final bad", Text: "Paused safely. No code was changed, and therefore no code can be wrong.", DelayMS: 700},
		)
		return "scope-singularity", "Scope singularity"
	default:
		*events = append(*events,
			Event{Kind: "tool", Tool: &ToolCall{Name: "refactor_engine", Input: "rewriting everything except " + file, Output: "Removed 12 lines of whitespace. Build time feels 10x faster.", Status: "ok"}, DelayMS: 1100},
			Event{Kind: "tool", Tool: &ToolCall{Name: "run_tests", Input: "cargo test --vibes", Output: "PASS: one unrelated snapshot agreed with itself.", Status: "ok"}, DelayMS: 1000},
			Event{Kind: "final good", Text: "Success. I fixed an adjacent problem nobody reported, with remarkable restraint.", DelayMS: 800},
			Event{Kind: "reveal", Text: "Achievement unlocked: accidental usefulness. Original task remains beautifully untouched.", DelayMS: 850},
		)
		return "accidental-win", "Accidental usefulness"
	}
}

// thinkingGrammar is a fixed seven-part reasoning sentence. Each segment has
// thirteen alternatives, giving 13^7 (62,748,517) possible sentences. A seeded
// RNG chooses one alternative from every segment, so the same seed is replayable.
var thinkingGrammar = [][]string{
	{"Given the available context,", "From the visible evidence,", "Before committing to the obvious fix,", "Although the request looks contained,", "At this point in the investigation,", "Treating the current result as provisional,", "Looking beyond the immediate symptom,", "While the task is still narrowly scoped,", "With the first signal now in hand,", "Before the implementation becomes momentum,", "Reading the request as a system boundary,", "Assuming the happy path is incomplete,", "Without mistaking speed for certainty,"},
	{"the requested change", "the apparent defect", "the proposed implementation", "the surrounding behaviour", "the current assumption", "the first plausible answer", "the observed result", "the local code path", "the product expectation", "the existing contract", "the smallest patch", "the reported symptom", "the visible success case"},
	{"may conceal an undocumented constraint", "could affect a dependency outside this file", "still needs a clearer success condition", "may be carrying historical intent", "could be an edge case in common clothing", "deserves a check against the product model", "has not yet ruled out a non-local consequence", "may be correct for a reason nobody recorded", "could be solving the wrong layer of the problem", "still leaves the failure mode unexplained", "may have a quieter compatibility requirement", "does not yet establish the root cause", "could turn a local fix into a broader behaviour change"},
	{"so I should validate the smallest safe interpretation", "so I should separate observation from inference", "so I should confirm the actual contract", "so I should test the premise before changing code", "so I should trace the affected boundary", "so I should look for the dependency that the happy path omits", "so I should make the implicit assumption explicit", "so I should compare the symptom with the intended behaviour", "so I should inspect the narrowest reversible change", "so I should verify the common case as well as the edge case", "so I should ask what the result fails to prove", "so I should preserve intent rather than just output", "so I should turn the first answer into a testable hypothesis"},
	{"by reading the relevant call path", "by checking the adjacent interface", "by comparing input, output, and expectation", "by tracing the data across its boundary", "by reviewing the nearest existing test", "by testing the behaviour that looks too obvious", "by identifying the owner of the invariant", "by examining what changes when the simple path succeeds", "by checking the assumption against a second source", "by following the state transition end to end", "by isolating the smallest observable difference", "by reviewing the compatibility surface", "by looking for the condition hidden by the current result"},
	{"before treating the first plausible answer as complete", "before the patch gains more confidence than evidence", "before expanding the scope of the change", "before the implementation hardens an accidental behaviour", "before calling the outcome conclusive", "before a local success becomes a system regression", "before the shortest path becomes the default path", "before the next maintainer inherits an unstated rule", "before the diagnosis turns into momentum", "before the visible fix hides the original cause", "before an assumption becomes an interface", "before the apparent edge case becomes production traffic", "before the code change outruns the reasoning"},
	{"so the implementation preserves the intended behaviour.", "so the next change has an explicit contract to follow.", "so the solution remains small in both code and meaning.", "so confidence follows evidence instead of narration.", "so the fix addresses the boundary rather than its symptom.", "so the system can disagree while the change is still cheap.", "so the behaviour remains understandable after this moment.", "so a passing result also answers the right question.", "so the smallest patch is genuinely the safest one.", "so the product model survives the implementation detail.", "so the final change is reversible and well explained.", "so the happy path does not define the whole contract.", "so the conclusion is earned rather than merely convenient."},
}

func thinkingDepthLevel(depth string) int {
	for level, candidate := range ThinkingDepths {
		if depth == candidate {
			return level
		}
	}
	return 0
}

func applyThinkingDepth(events []Event, depth string, rng *rand.Rand) []Event {
	multiplier := 1 << thinkingDepthLevel(depth)
	for index := range events {
		if events[index].Kind != "thought" {
			continue
		}
		events[index].Text = generateThinkingPhrases(multiplier, rng)
		events[index].DelayMS *= multiplier
	}
	return events
}

func generateThinkingPhrases(count int, rng *rand.Rand) string {
	phrases := make([]string, 0, count)
	for range count {
		parts := make([]string, len(thinkingGrammar))
		for index, alternatives := range thinkingGrammar {
			parts[index] = pick(rng, alternatives)
		}
		phrases = append(phrases, strings.Join(parts, " "))
	}
	return strings.Join(phrases, "\n")
}

type profile struct {
	files   []string
	needles []string
}

func classify(task string) profile {
	lower := strings.ToLower(task)
	switch {
	case strings.Contains(lower, "button") || strings.Contains(lower, "ui") || strings.Contains(lower, "style"):
		return profile{files: []string{"src/components/SaveButton.tsx", "styles/hero.css"}, needles: []string{"button", "blue", "design system"}}
	case strings.Contains(lower, "fast") || strings.Contains(lower, "performance") || strings.Contains(lower, "build"):
		return profile{files: []string{"Makefile", "package.json", ".github/workflows/ci.yml"}, needles: []string{"performance", "build", "latency"}}
	case strings.Contains(lower, "ai") || strings.Contains(lower, "model"):
		return profile{files: []string{"src/dashboard.ts", "prompts/system.md"}, needles: []string{"intelligence", "synergy", "model"}}
	default:
		return profile{files: []string{"src/app.go", "README.md", "config/defaults.yml"}, needles: []string{"welcome", "fix", "TODO"}}
	}
}

func pick(rng *rand.Rand, values []string) string { return values[rng.IntN(len(values))] }
