package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"glowing-agent/simulator"
)

type tuiState uint8

const (
	editing tuiState = iota
	running
	paused
	completed
)

type focusTarget uint8

const (
	focusTask focusTarget = iota
	focusPreset
	focusSeed
	focusDepth
	focusSpeed
	focusCount
)

type playbackTickMsg struct {
	runID uint64
}

type tuiModel struct {
	task     textarea.Model
	seed     textinput.Model
	viewport viewport.Model

	focus       focusTarget
	presetIndex int
	depthIndex  int
	speedIndex  int
	state       tuiState
	simulation  *simulator.Simulation
	shown       int
	runID       uint64
	err         string
	width       int
	height      int
}

var playbackSpeeds = []float64{0.5, 1, 2, 5}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB000"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7E8795"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	panelStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4D6078")).Padding(0, 1)
	focusStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#101820")).Background(lipgloss.Color("#56D4DD")).Padding(0, 1)
	controlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C8D2DC")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4D6078")).Padding(0, 1)
	toolStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#63D8FF"))
	thoughtStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C5A3FF"))
	planStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166"))
	finalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8CE99A"))
)

func newTUIModel() tuiModel {
	task := textarea.New()
	task.Prompt = "› "
	task.Placeholder = "e.g. Fix the typo in the welcome message"
	task.CharLimit = 1000
	task.MaxHeight = 5
	task.KeyMap.InsertNewline.SetKeys("shift+enter")
	task.SetHeight(4)

	seed := textinput.New()
	seed.Prompt = "seed "
	seed.Placeholder = "chaos"
	seed.CharLimit = 20
	seed.Validate = validateSeedInput

	logs := viewport.New(viewport.WithWidth(80), viewport.WithHeight(8))
	logs.SoftWrap = true
	logs.FillHeight = false

	return tuiModel{
		task: task, seed: seed, viewport: logs,
		presetIndex: -1, depthIndex: 0, speedIndex: 1,
		state: editing,
	}
}

func validateSeedInput(value string) error {
	if value == "" || value == "-" {
		return nil
	}
	_, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.New("signed 64-bit integer required")
	}
	return nil
}

func (m tuiModel) Init() tea.Cmd {
	return m.task.Focus()
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil
	case playbackTickMsg:
		if msg.runID != m.runID || m.state != running {
			return m, nil
		}
		return m, m.advance()
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}

	if m.state != editing {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m tuiModel) updateKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	pressed := key.String()
	if pressed == "ctrl+c" {
		return m, tea.Quit
	}
	if m.state != editing {
		switch pressed {
		case "q":
			return m, tea.Quit
		case "space":
			if m.state == running {
				m.state = paused
			} else if m.state == paused {
				m.state = running
				return m, m.scheduleNext()
			}
			return m, nil
		case "r":
			if m.simulation != nil {
				m.state = running
				m.shown = 0
				m.runID++
				m.refreshLog(true)
				return m, m.advance()
			}
		case "n":
			m.state = editing
			m.shown = 0
			m.runID++
			m.simulation = nil
			m.presetIndex = -1
			m.task.SetValue("")
			m.refreshLog(false)
			return m, m.setFocus(focusTask)
		case "pgup", "up":
			m.viewport.PageUp()
			return m, nil
		case "pgdown", "down":
			m.viewport.PageDown()
			return m, nil
		}
		return m, nil
	}

	switch pressed {
	case "enter":
		return m, m.startSimulation()
	case "tab":
		return m, m.setFocus((m.focus + 1) % focusCount)
	case "shift+tab":
		return m, m.setFocus((m.focus + focusCount - 1) % focusCount)
	}

	switch m.focus {
	case focusTask:
		var cmd tea.Cmd
		m.task, cmd = m.task.Update(key)
		return m, cmd
	case focusSeed:
		var cmd tea.Cmd
		m.seed, cmd = m.seed.Update(key)
		return m, cmd
	case focusPreset:
		if pressed == "left" || pressed == "up" {
			m.selectPreset(-1)
		} else if pressed == "right" || pressed == "down" {
			m.selectPreset(1)
		}
	case focusDepth:
		if pressed == "left" || pressed == "up" {
			m.depthIndex = (m.depthIndex + len(simulator.ThinkingDepths) - 1) % len(simulator.ThinkingDepths)
		} else if pressed == "right" || pressed == "down" {
			m.depthIndex = (m.depthIndex + 1) % len(simulator.ThinkingDepths)
		}
	case focusSpeed:
		if pressed == "left" || pressed == "up" {
			m.speedIndex = (m.speedIndex + len(playbackSpeeds) - 1) % len(playbackSpeeds)
		} else if pressed == "right" || pressed == "down" {
			m.speedIndex = (m.speedIndex + 1) % len(playbackSpeeds)
		}
	}
	return m, nil
}

func (m *tuiModel) selectPreset(delta int) {
	presets := simulator.Presets()
	option := m.presetIndex + 1 // zero is the custom-task option.
	option = (option + delta + len(presets) + 1) % (len(presets) + 1)
	m.presetIndex = option - 1
	if m.presetIndex >= 0 {
		m.task.SetValue(presets[m.presetIndex].Task)
	}
}

func (m *tuiModel) setFocus(target focusTarget) tea.Cmd {
	m.focus = target
	m.task.Blur()
	m.seed.Blur()
	switch target {
	case focusTask:
		return m.task.Focus()
	case focusSeed:
		return m.seed.Focus()
	default:
		return nil
	}
}

func (m *tuiModel) startSimulation() tea.Cmd {
	task := strings.TrimSpace(m.task.Value())
	if count := utf8.RuneCountInString(task); count == 0 || count > 1000 {
		m.err = "Task text must be between 1 and 1000 characters."
		return nil
	}
	if m.seed.Err != nil {
		m.err = "Seed must be a signed 64-bit integer."
		return nil
	}
	var seed *int64
	if value := m.seed.Value(); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			m.err = "Seed must be a signed 64-bit integer."
			return nil
		}
		seed = &parsed
	}
	result := simulator.GenerateWithThinkingDepth(task, seed, simulator.ThinkingDepths[m.depthIndex])
	m.simulation = &result
	m.seed.SetValue(strconv.FormatInt(result.Seed, 10))
	m.state = running
	m.shown = 0
	m.runID++
	m.err = ""
	m.refreshLog(true)
	return m.advance()
}

func (m *tuiModel) advance() tea.Cmd {
	if m.simulation == nil || m.shown >= len(m.simulation.Events) {
		m.state = completed
		return nil
	}
	m.shown++
	m.refreshLog(true)
	if m.shown == len(m.simulation.Events) {
		m.state = completed
		return nil
	}
	return m.scheduleNext()
}

func (m tuiModel) scheduleNext() tea.Cmd {
	if m.simulation == nil || m.shown >= len(m.simulation.Events) {
		return nil
	}
	delay := time.Duration(float64(m.simulation.Events[m.shown].DelayMS) * float64(time.Millisecond) / playbackSpeeds[m.speedIndex])
	runID := m.runID
	return tea.Tick(delay, func(time.Time) tea.Msg { return playbackTickMsg{runID: runID} })
}

func (m *tuiModel) resize() {
	if m.width < 60 || m.height < 16 {
		return
	}
	inner := max(20, m.width-6)
	m.task.SetWidth(inner)
	m.task.SetHeight(4)
	m.seed.SetWidth(22)
	m.viewport.SetWidth(inner)
	m.viewport.SetHeight(max(5, m.height-18))
	m.refreshLog(false)
}

func (m *tuiModel) refreshLog(follow bool) {
	if m.simulation == nil || m.shown == 0 {
		m.viewport.SetContent("Waiting for an opportunity to overthink.")
		return
	}
	lines := make([]string, 0, m.shown*2)
	for _, event := range m.simulation.Events[:m.shown] {
		lines = append(lines, formatEvent(event))
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	if follow {
		m.viewport.GotoBottom()
	}
}

func formatEvent(event simulator.Event) string {
	if event.Tool != nil {
		output := "  " + strings.ReplaceAll(event.Tool.Output, "\n", "\n  ")
		return toolStyle.Render("[tool] "+event.Tool.Name+" "+event.Tool.Input) + "\n" + mutedStyle.Render(output)
	}
	prefix := "[" + event.Kind + "] "
	switch event.Kind {
	case "thought":
		return thoughtStyle.Render(prefix + event.Text)
	case "plan":
		return planStyle.Render(prefix + event.Text)
	case "final good", "final bad", "reveal":
		return finalStyle.Render(prefix + event.Text)
	default:
		return mutedStyle.Render(prefix + event.Text)
	}
}

func (m tuiModel) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("Loading glowing-agent…")
	}
	if m.width < 60 || m.height < 16 {
		view := tea.NewView("glowing-agent\n\nPlease enlarge the terminal to at least 60 columns × 16 rows.\n\nctrl+c to quit")
		view.AltScreen = true
		view.WindowTitle = "glowing-agent"
		return view
	}

	content := []string{
		titleStyle.Render("glowing-agent") + "  " + mutedStyle.Render("SIMULATION MODE · "+m.statusText()),
		panelStyle.Width(m.width - 4).Render("TASK\n" + m.task.View()),
		m.settingsView(),
		panelStyle.Width(m.width - 4).Render("AGENT SESSION LOG\n" + m.viewport.View()),
	}
	if m.err != "" {
		content = append(content, errorStyle.Render(m.err))
	}
	if m.simulation != nil {
		content = append(content, m.metricsView())
	}
	content = append(content, mutedStyle.Render(m.helpText()))

	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content...))
	view.AltScreen = true
	view.WindowTitle = "glowing-agent"
	return view
}

func (m tuiModel) settingsView() string {
	presets := simulator.Presets()
	preset := "custom"
	if m.presetIndex >= 0 {
		preset = presets[m.presetIndex].Label
	}
	controls := []string{
		m.control(focusPreset, "PRESET", preset),
		m.control(focusSeed, "SEED", m.seed.View()),
		m.control(focusDepth, "THINKING", simulator.ThinkingDepths[m.depthIndex]),
		m.control(focusSpeed, "SPEED", fmt.Sprintf("%gx", playbackSpeeds[m.speedIndex])),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, controls...)
}

func (m tuiModel) control(target focusTarget, label, value string) string {
	content := label + " " + value
	if m.state == editing && m.focus == target {
		return focusStyle.Render(content)
	}
	return controlStyle.Render(content)
}

func (m tuiModel) metricsView() string {
	metrics := m.simulation.Metrics
	return panelStyle.Width(m.width - 4).Render(fmt.Sprintf("ENDING %s   CONFIDENCE %d%%   TOKENS %d   MEETINGS %d   FILES CHANGED %d   SEED %d",
		m.simulation.EndingName, metrics.Confidence, metrics.TokensBurned, metrics.MeetingsAvoided, metrics.FilesActuallySet, m.simulation.Seed))
}

func (m tuiModel) statusText() string {
	switch m.state {
	case running:
		return "RUNNING"
	case paused:
		return "PAUSED"
	case completed:
		return "COMPLETE"
	default:
		return "READY"
	}
}

func (m tuiModel) helpText() string {
	if m.state == editing {
		return "tab/shift+tab focus · shift+enter newline · enter run · ctrl+c quit"
	}
	return "space pause/resume · r replay · n new task · pgup/pgdown scroll · q quit"
}
