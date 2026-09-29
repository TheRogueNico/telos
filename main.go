package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Catppuccin Mocha
var (
	lavender = lipgloss.Color("#b4befe")
	overlay0 = lipgloss.Color("#6c7086")
)

// All Lip Gloss styles live here.
type paneStyle struct {
	edge  lipgloss.Style
	title lipgloss.Style
	body  lipgloss.Style
}

var (
	roundedBorder = lipgloss.RoundedBorder()

	activePane = paneStyle{
		edge:  lipgloss.NewStyle().Foreground(lavender),
		title: lipgloss.NewStyle().Foreground(lavender).Bold(true),
		body: lipgloss.NewStyle().
			Border(roundedBorder, false, true, true, true).
			BorderForeground(lavender),
	}

	inactivePane = paneStyle{
		edge:  lipgloss.NewStyle().Foreground(overlay0),
		title: lipgloss.NewStyle().Foreground(overlay0),
		body: lipgloss.NewStyle().
			Border(roundedBorder, false, true, true, true).
			BorderForeground(overlay0),
	}
)

type pane int

const (
	paneList pane = iota
	paneDetails
)

func (p pane) title() string {
	switch p {
	case paneList:
		return "[H] Task List"
	case paneDetails:
		return "[L] Details"
	}
	return ""
}

type keyMap struct {
	Quit         key.Binding
	FocusList    key.Binding
	FocusDetails key.Binding
}

var keys = keyMap{
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	FocusList:    key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "task list")),
	FocusDetails: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "details")),
}

type model struct {
	width, height int
	focus         pane
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.FocusList):
			m.focus = paneList
		case key.Matches(msg, keys.FocusDetails):
			m.focus = paneDetails
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m model) render() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	listW := m.width / 3
	detailsW := m.width - listW

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPane(paneList, listW, m.height),
		m.renderPane(paneDetails, detailsW, m.height),
	)
}

// renderPane draws an empty bordered box with the title embedded in the top border line.
func (m model) renderPane(p pane, w, h int) string {
	s := inactivePane
	if m.focus == p {
		s = activePane
	}
	b := roundedBorder

	// Top edge with title
	inner := w - lipgloss.Width(b.TopLeft) - lipgloss.Width(b.TopRight)
	text := " " + p.title() + " "
	lead := 1 // border characters before the title
	if lead+lipgloss.Width(text) > inner {
		text, lead = "", 0 // too narrow for the title
	}
	fill := max(0, inner-lead-lipgloss.Width(text))

	top := s.edge.Render(b.TopLeft+strings.Repeat(b.Top, lead)) +
		s.title.Render(text) +
		s.edge.Render(strings.Repeat(b.Top, fill)+b.TopRight)

	body := s.body.Width(w).Height(h - 1).Render("")

	return lipgloss.JoinVertical(lipgloss.Left, top, body)
}

func main() {
	if _, err := tea.NewProgram(model{focus: paneList}).Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
