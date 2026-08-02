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

const (
	minimumTerminalWidth  = 60
	minimumTerminalHeight = 24
	minimumLogHeight      = 5
	sidebarBreakpoint     = 96
	sidebarWidth          = 31
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
	colorCanvas       = lipgloss.Color("#1D1922")
	colorSurface      = lipgloss.Color("#292330")
	colorBorder       = lipgloss.Color("#45394F")
	colorText         = lipgloss.Color("#DED7E5")
	colorMuted        = lipgloss.Color("#81768B")
	colorSubtle       = lipgloss.Color("#A69AAD")
	colorViolet       = lipgloss.Color("#7657FF")
	colorPink         = lipgloss.Color("#ED5CFF")
	colorMint         = lipgloss.Color("#39D98A")
	colorWarning      = lipgloss.Color("#E9B86B")
	colorError        = lipgloss.Color("#FF6B81")
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(colorViolet)
	titleAccentStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorPink)
	diagonalStyle     = lipgloss.NewStyle().Foreground(colorViolet)
	mutedStyle        = lipgloss.NewStyle().Foreground(colorMuted)
	subtleStyle       = lipgloss.NewStyle().Foreground(colorSubtle)
	errorStyle        = lipgloss.NewStyle().Foreground(colorError)
	panelStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(colorBorder).PaddingLeft(1)
	panelLabelStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorPink)
	focusStyle        = lipgloss.NewStyle().Bold(true).Foreground(colorText).Background(colorViolet).Padding(0, 1)
	controlStyle      = lipgloss.NewStyle().Foreground(colorSubtle).Background(colorSurface).Padding(0, 1)
	toolStyle         = lipgloss.NewStyle().Foreground(colorPink)
	thoughtStyle      = lipgloss.NewStyle().Foreground(colorSubtle)
	planStyle         = lipgloss.NewStyle().Foreground(colorViolet)
	finalStyle        = lipgloss.NewStyle().Foreground(colorMint)
	warningStyle      = lipgloss.NewStyle().Foreground(colorWarning)
	statusReadyStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorViolet)
	statusActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(colorMint)
	statusPausedStyle = lipgloss.NewStyle().Bold(true).Foreground(colorWarning)
	statusDoneStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorPink)
	sidebarStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(colorBorder).PaddingLeft(2)
	promptFrameStyle  = lipgloss.NewStyle().Border(lipgloss.ThickBorder(), false, false, false, true).BorderForeground(colorViolet).PaddingLeft(1)
	sectionStyle      = lipgloss.NewStyle().Foreground(colorMuted)
)

func newTUIModel() tuiModel {
	task := textarea.New()
	task.Prompt = "> "
	task.Placeholder = "e.g. Fix the typo in the welcome message"
	task.CharLimit = 1000
	task.MaxHeight = 5
	task.ShowLineNumbers = false
	task.KeyMap.InsertNewline.SetKeys("shift+enter")
	task.SetHeight(4)
	taskStyles := textarea.DefaultDarkStyles()
	taskStyles.Focused.Base = taskStyles.Focused.Base.Background(colorCanvas)
	taskStyles.Focused.Text = taskStyles.Focused.Text.Foreground(colorText)
	taskStyles.Focused.Prompt = taskStyles.Focused.Prompt.Foreground(colorPink).Bold(true)
	taskStyles.Focused.Placeholder = taskStyles.Focused.Placeholder.Foreground(colorMuted)
	taskStyles.Blurred.Base = taskStyles.Blurred.Base.Background(colorCanvas)
	taskStyles.Blurred.Text = taskStyles.Blurred.Text.Foreground(colorSubtle)
	taskStyles.Blurred.Prompt = taskStyles.Blurred.Prompt.Foreground(colorViolet)
	taskStyles.Blurred.Placeholder = taskStyles.Blurred.Placeholder.Foreground(colorMuted)
	task.SetStyles(taskStyles)

	seed := textinput.New()
	seed.Prompt = ""
	seed.Placeholder = "chaos"
	seed.CharLimit = 20
	seed.Validate = validateSeedInput
	seedStyles := textinput.DefaultDarkStyles()
	seedStyles.Focused.Text = seedStyles.Focused.Text.Foreground(colorText)
	seedStyles.Focused.Prompt = seedStyles.Focused.Prompt.Foreground(colorPink).Bold(true)
	seedStyles.Focused.Placeholder = seedStyles.Focused.Placeholder.Foreground(colorMuted)
	seedStyles.Blurred.Text = seedStyles.Blurred.Text.Foreground(colorSubtle)
	seedStyles.Blurred.Prompt = seedStyles.Blurred.Prompt.Foreground(colorMuted)
	seedStyles.Blurred.Placeholder = seedStyles.Blurred.Placeholder.Foreground(colorMuted)
	seed.SetStyles(seedStyles)

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
			m.err = ""
			m.presetIndex = -1
			m.task.SetValue("")
			m.resize()
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
		m.resize()
		return nil
	}
	if m.seed.Err != nil {
		m.err = "Seed must be a signed 64-bit integer."
		m.resize()
		return nil
	}
	var seed *int64
	if value := m.seed.Value(); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			m.err = "Seed must be a signed 64-bit integer."
			m.resize()
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
	m.resize()
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
	if m.width < minimumTerminalWidth || m.height < minimumTerminalHeight {
		return
	}
	mainWidth := m.mainWidth()
	inner := max(20, mainWidth-3)
	m.task.SetWidth(inner)
	m.task.SetHeight(4)
	seedWidth := 18
	if !m.hasSidebar() && m.width < 80 {
		seedWidth = 10
	}
	m.seed.SetWidth(seedWidth)
	m.viewport.SetWidth(inner)
	m.viewport.SetHeight(max(minimumLogHeight, m.height-m.fixedHeight()-1))
	m.refreshLog(false)
}

func (m tuiModel) fixedHeight() int {
	content := []string{
		m.sessionHeaderView(),
		m.taskView(),
	}
	if !m.hasSidebar() {
		content = append(content, m.settingsView())
	}
	if m.err != "" {
		content = append(content, errorStyle.Render(m.err))
	}
	if m.simulation != nil && !m.hasSidebar() {
		content = append(content, m.metricsView())
	}
	content = append(content, mutedStyle.Render(m.helpText()))
	return lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m tuiModel) hasSidebar() bool {
	return m.width >= sidebarBreakpoint
}

func (m tuiModel) mainWidth() int {
	if m.hasSidebar() {
		return m.width - sidebarWidth
	}
	return m.width
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
		resultStyle := mutedStyle
		if event.Tool.Status == "warning" {
			resultStyle = warningStyle
		}
		return toolStyle.Render("▣ tool  "+event.Tool.Name) + " " + subtleStyle.Render(event.Tool.Input) + "\n" + resultStyle.Render(output)
	}
	switch event.Kind {
	case "thought":
		return thoughtStyle.Render("◌ reasoning  " + event.Text)
	case "plan":
		return planStyle.Render("▸ plan       " + event.Text)
	case "final good", "final bad", "reveal":
		if event.Kind == "reveal" {
			return warningStyle.Render("◇ reveal     " + event.Text)
		}
		return finalStyle.Render("◆ final      " + event.Text)
	default:
		return mutedStyle.Render("· " + event.Kind + "  " + event.Text)
	}
}

func (m tuiModel) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("Loading glowing-agent…")
	}
	if m.width < minimumTerminalWidth || m.height < minimumTerminalHeight {
		view := tea.NewView(fmt.Sprintf("glowing-agent\n\nPlease enlarge the terminal to at least %d columns × %d rows.\n\nctrl+c to quit", minimumTerminalWidth, minimumTerminalHeight))
		view.AltScreen = true
		view.BackgroundColor = colorCanvas
		view.ForegroundColor = colorText
		view.WindowTitle = "glowing-agent"
		return view
	}

	content := []string{
		m.sessionHeaderView(),
		m.viewport.View(),
		m.taskView(),
	}
	if !m.hasSidebar() {
		content = append(content, m.settingsView())
	}
	if m.err != "" {
		content = append(content, errorStyle.Render(m.err))
	}
	if m.simulation != nil && !m.hasSidebar() {
		content = append(content, m.metricsView())
	}
	content = append(content, mutedStyle.Render(m.helpText()))

	main := lipgloss.NewStyle().Width(m.mainWidth()).Render(lipgloss.JoinVertical(lipgloss.Left, content...))
	canvas := main
	if m.hasSidebar() {
		canvas = lipgloss.JoinHorizontal(lipgloss.Top, main, m.sidebarView())
	}
	view := tea.NewView(canvas)
	view.AltScreen = true
	view.BackgroundColor = colorCanvas
	view.ForegroundColor = colorText
	view.WindowTitle = "glowing-agent"
	return view
}

func (m tuiModel) sessionHeaderView() string {
	return mutedStyle.Render("  AGENT SESSION") + "  " + m.statusView()
}

func (m tuiModel) taskView() string {
	return promptFrameStyle.Width(max(20, m.mainWidth()-3)).Render(m.task.View())
}

func (m tuiModel) sidebarView() string {
	settings := []string{
		m.sidebarControl(focusPreset, "Preset", m.presetName()),
		m.sidebarControl(focusSeed, "Seed", m.seed.View()),
		m.sidebarControl(focusDepth, "Thinking", simulator.ThinkingDepths[m.depthIndex]),
		m.sidebarControl(focusSpeed, "Speed", fmt.Sprintf("%gx", playbackSpeeds[m.speedIndex])),
	}
	session := []string{
		sectionRule("SESSION"),
		mutedStyle.Render("Modified Files"),
		faintValue("None"),
	}
	if m.simulation != nil {
		metrics := m.simulation.Metrics
		session = []string{
			sectionRule("RESULT"),
			subtleStyle.Render(truncateText(m.simulation.EndingName, sidebarWidth-4)),
			mutedStyle.Render(fmt.Sprintf("Confidence  %d%%", metrics.Confidence)),
			mutedStyle.Render(fmt.Sprintf("Tokens      %d", metrics.TokensBurned)),
			mutedStyle.Render(fmt.Sprintf("Files       %d", metrics.FilesActuallySet)),
		}
	}
	content := []string{
		diagonalStyle.Render("╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱"),
		titleStyle.Render("glowing") + titleAccentStyle.Render("-agent"),
		mutedStyle.Render("SIMULATION WORKBENCH"),
		diagonalStyle.Render("╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱"),
		"",
		m.statusView() + "  " + subtleStyle.Render("New Session"),
		mutedStyle.Render("./glowing-agent"),
		"",
		sectionRule("RUN SETTINGS"),
	}
	content = append(content, settings...)
	content = append(content, "")
	content = append(content, session...)
	return sidebarStyle.Width(sidebarWidth - 3).Height(m.height).Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func sectionRule(label string) string {
	remaining := max(1, 24-len(label))
	return sectionStyle.Render(label + " " + strings.Repeat("─", remaining))
}

func faintValue(value string) string {
	return mutedStyle.Faint(true).Render(value)
}

func truncateText(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:max(1, width-1)]) + "…"
}

func (m tuiModel) presetName() string {
	if m.presetIndex >= 0 {
		return simulator.Presets()[m.presetIndex].Label
	}
	return "custom"
}

func (m tuiModel) sidebarControl(target focusTarget, label, value string) string {
	marker := mutedStyle.Render("◇")
	text := subtleStyle.Render(label + "  " + value)
	if m.state == editing && m.focus == target {
		marker = titleAccentStyle.Render("◆")
		text = titleStyle.Render(label + "  " + value)
	}
	return marker + " " + text
}

func (m tuiModel) settingsView() string {
	preset := m.presetName()
	labels := []string{"PRESET", "SEED", "THINKING", "SPEED"}
	speed := fmt.Sprintf("%gx", playbackSpeeds[m.speedIndex])
	if m.width < 80 {
		labels = []string{"P", "S", "T", "×"}
		speed = fmt.Sprintf("%g", playbackSpeeds[m.speedIndex])
	}
	controls := []string{
		m.control(focusPreset, labels[0], preset),
		m.control(focusSeed, labels[1], m.seed.View()),
		m.control(focusDepth, labels[2], simulator.ThinkingDepths[m.depthIndex]),
		m.control(focusSpeed, labels[3], speed),
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
	values := fmt.Sprintf("ENDING %s   CONFIDENCE %d%%   TOKENS %d   FILES %d",
		m.simulation.EndingName, metrics.Confidence, metrics.TokensBurned, metrics.FilesActuallySet)
	if m.width < 80 {
		values = fmt.Sprintf("ENDING %s · %d%% · %d TOKENS", m.simulation.EndingName, metrics.Confidence, metrics.TokensBurned)
	}
	return panelLabelStyle.Render("RESULT  ") + subtleStyle.Render(values)
}

func (m tuiModel) statusView() string {
	status := m.statusText()
	switch m.state {
	case running:
		return statusActiveStyle.Render(status)
	case paused:
		return statusPausedStyle.Render(status)
	case completed:
		return statusDoneStyle.Render(status)
	default:
		return statusReadyStyle.Render(status)
	}
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
	if m.mainWidth() < 80 {
		if m.state == editing {
			return "tab focus · enter run · ctrl+c quit"
		}
		return "space pause · r replay · n new · pg scroll · q quit"
	}
	if m.state == editing {
		return "tab/shift+tab focus · shift+enter newline · enter run · ctrl+c quit"
	}
	return "space pause/resume · r replay · n new task · pgup/pgdown scroll · q quit"
}
