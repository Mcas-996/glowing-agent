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

var thinkingPhrases = []string{
	"The request appears small, but its assumptions may extend beyond the visible change.",
	"Before acting, I should distinguish the stated requirement from the requirement it quietly implies.",
	"The surrounding code may be correct for reasons that are no longer documented.",
	"A quick fix is still a hypothesis until the system has had an opportunity to disagree.",
	"The most relevant constraint may be the one that has not yet been named.",
	"This behaviour deserves a second look from the perspective of the next maintainer.",
	"The evidence is encouraging, though it has not yet earned the right to be conclusive.",
	"A local change can carry a surprisingly non-local interpretation.",
	"The implementation path is clear enough to be suspicious.",
	"I should verify whether the apparent edge case is actually the common case in disguise.",
	"The task may be asking for a code change while revealing a boundary in the product model.",
	"It is worth separating what is observable from what merely feels architecturally significant.",
	"A stable solution needs to preserve the intent that led to the current behaviour.",
	"The safest next step is to make the implicit contract explicit in my reasoning.",
	"The available context supports a direction, not yet a conclusion.",
	"There may be a dependency here that only becomes visible when the simple path succeeds.",
	"The shortest implementation is not always the smallest conceptual change.",
	"I should account for the consequences that the happy path has politely omitted.",
	"This is probably straightforward, which makes it an excellent place to inspect the premise.",
	"The system is offering an answer; the remaining question is whether it is answering the right problem.",
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
	previous := -1
	for range count {
		index := rng.IntN(len(thinkingPhrases) - boolToInt(previous >= 0))
		if previous >= 0 && index >= previous {
			index++
		}
		phrases = append(phrases, thinkingPhrases[index])
		previous = index
	}
	return strings.Join(phrases, "\n")
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
