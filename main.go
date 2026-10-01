package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Catppuccin Mocha
var (
	lavender = lipgloss.Color("#b4befe")
	blue     = lipgloss.Color("#89b4fa")
	green    = lipgloss.Color("#a6e3a1")
	yellow   = lipgloss.Color("#f9e2af")
	red      = lipgloss.Color("#f38ba8")
	text     = lipgloss.Color("#cdd6f4")
	subtext1 = lipgloss.Color("#bac2de")
	subtext0 = lipgloss.Color("#a6adc8")
	overlay2 = lipgloss.Color("#9399b2")
	overlay1 = lipgloss.Color("#7f849c")
	overlay0 = lipgloss.Color("#6c7086")
)

// Custom colors
var selection = lipgloss.Color("#3d4f7c")

// Lipgloss styles
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
			BorderForeground(lavender).
			Padding(0, 1),
	}

	inactivePane = paneStyle{
		edge:  lipgloss.NewStyle().Foreground(overlay0),
		title: lipgloss.NewStyle().Foreground(overlay0),
		body: lipgloss.NewStyle().
			Border(roundedBorder, false, true, true, true).
			BorderForeground(overlay0).
			Padding(0, 1),
	}

	// Task list window
	groupStyle            = lipgloss.NewStyle().Foreground(subtext1).Bold(true)
	taskStyle             = lipgloss.NewStyle().Foreground(text)
	canceledStyle         = lipgloss.NewStyle().Foreground(overlay1).Strikethrough(true)
	selectedStyle         = lipgloss.NewStyle().Background(selection)
	selectedCanceledStyle = canceledStyle.Foreground(overlay2)

	statusStyles = map[taskStatus]lipgloss.Style{
		statusActive:    lipgloss.NewStyle().Foreground(blue),
		statusPaused:    lipgloss.NewStyle().Foreground(yellow),
		statusCanceled:  lipgloss.NewStyle().Foreground(red),
		statusCompleted: lipgloss.NewStyle().Foreground(green),
	}

	// Details
	detailTitleStyle = lipgloss.NewStyle().Foreground(text).Bold(true)
	detailLabelStyle = lipgloss.NewStyle().Foreground(subtext0).Bold(true)
	detailMetaStyle  = lipgloss.NewStyle().Foreground(subtext0)
	detailTextStyle  = lipgloss.NewStyle().Foreground(text)
	keyDoneStyle     = lipgloss.NewStyle().Foreground(green)
	keyDoneTextStyle = lipgloss.NewStyle().Foreground(overlay1)
	keyTodoStyle     = lipgloss.NewStyle().Foreground(subtext0)
)

type taskStatus string

const (
	statusActive    taskStatus = "active"
	statusPaused    taskStatus = "paused"
	statusCanceled  taskStatus = "canceled"
	statusCompleted taskStatus = "completed"
)

type keyTask struct {
	description string
	completed   bool
}

type task struct {
	id        int
	title     string
	intent    string
	keyTasks  []keyTask
	endState  string
	status    taskStatus
	createdAt time.Time
}

type group struct {
	id    int
	name  string
	tasks []task
}

// Nerd Font icons
const (
	iconGroupOpen   = "\uf07c"
	iconGroupClosed = "\uf07b"
	iconGroupEmpty  = "\uf114"
	iconKeyTodo     = "\uf10c"
	iconKeyDone     = "\uf111"
)

var statusIcons = map[taskStatus]string{
	statusActive:    "\uf04b",
	statusPaused:    "\uf28b",
	statusCanceled:  "\uf057",
	statusCompleted: "\uf058",
}

// Task list indentation
const (
	groupIndent = " "
	taskIndent  = "  "
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
	Up           key.Binding
	Down         key.Binding
	Toggle       key.Binding
}

var keys = keyMap{
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	FocusList:    key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "task list")),
	FocusDetails: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "details")),
	Up:           key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "up")),
	Down:         key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "down")),
	Toggle:       key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter", "toggle group")),
}

// row is one visible line in the task list, a group header or a task.
type row struct {
	group int // index into model.groups
	task  int // index into group.tasks. -1 for the group header
}

type model struct {
	width, height int
	focus         pane
	groups        []group
	collapsed     map[int]bool // group ID -> collapsed
	cursor        int          // index into rows()
}

// rows flattens the groups into the lines currently visible in the list.
func (m model) rows() []row {
	var rows []row
	for gi, g := range m.groups {
		rows = append(rows, row{group: gi, task: -1})
		if m.collapsed[g.id] {
			continue
		}
		for ti := range g.tasks {
			rows = append(rows, row{group: gi, task: ti})
		}
	}
	return rows
}

// selected returns the row under the cursor, if there is one.
func (m model) selected() (row, bool) {
	rows := m.rows()
	if m.cursor >= len(rows) {
		return row{}, false
	}
	return rows[m.cursor], true
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
		case m.focus == paneList:
			m = m.updateList(msg)
		}
	}
	return m, nil
}

// updateList handles the keys that only apply while the task list is focused.
func (m model) updateList(msg tea.KeyPressMsg) model {
	switch {
	case key.Matches(msg, keys.Up):
		m.cursor = max(m.cursor-1, 0)
	case key.Matches(msg, keys.Down):
		m.cursor = max(min(m.cursor+1, len(m.rows())-1), 0)
	case key.Matches(msg, keys.Toggle):
		if r, ok := m.selected(); ok && r.task == -1 {
			id := m.groups[r.group].id
			m.collapsed[id] = !m.collapsed[id]
		}
	}
	return m
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
	listInner := listW - activePane.body.GetHorizontalFrameSize()

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPane(paneList, listW, m.height, m.renderList(listInner)),
		m.renderPane(paneDetails, detailsW, m.height, m.renderDetails()),
	)
}

func (m model) groupIcon(g group) string {
	switch {
	case len(g.tasks) == 0:
		return iconGroupEmpty
	case m.collapsed[g.id]:
		return iconGroupClosed
	default:
		return iconGroupOpen
	}
}

// renderList draws the visible rows. inner is the pane's usable width.
func (m model) renderList(inner int) string {
	rows := m.rows()
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = m.renderRow(r, i == m.cursor, inner)
	}
	return strings.Join(lines, "\n")
}

// renderRow draws one row of the task list. A selected row gets a background
// bar spanning inner cells.
func (m model) renderRow(r row, selected bool, inner int) string {
	g := m.groups[r.group]

	var icon, label, indent string
	var iconStyle, labelStyle lipgloss.Style
	canceled := false

	if r.task == -1 {
		icon, label, indent = m.groupIcon(g), g.name, groupIndent
		iconStyle, labelStyle = groupStyle, groupStyle
	} else {
		t := g.tasks[r.task]
		icon, label, indent = statusIcons[t.status], t.title, taskIndent
		iconStyle, labelStyle = statusStyles[t.status], taskStyle
		if t.status == statusCanceled {
			labelStyle = canceledStyle
			canceled = true
		}
	}

	if selected {
		// The bar only adds a background, so each item keeps its own
		// colors. Canceled titles switch to a lighter gray to stay readable.
		if canceled {
			labelStyle = selectedCanceledStyle
		}
		iconStyle = iconStyle.Inherit(selectedStyle)
		labelStyle = labelStyle.Inherit(selectedStyle).Bold(true)
	}

	line := iconStyle.Render(indent+icon+" ") + labelStyle.Render(label)
	if selected {
		// Extend the bar to the full width of the pane
		pad := max(0, inner-lipgloss.Width(line))
		line += selectedStyle.Render(strings.Repeat(" ", pad))
	}
	return line
}

func (m model) renderDetails() string {
	r, ok := m.selected()
	if !ok {
		return ""
	}

	g := m.groups[r.group]
	if r.task == -1 {
		return renderGroupDetails(g)
	}
	return renderTaskDetails(g.tasks[r.task])
}

func renderGroupDetails(g group) string {
	count := fmt.Sprintf("%d tasks", len(g.tasks))
	if len(g.tasks) == 1 {
		count = "1 task"
	}
	return detailTitleStyle.Render(g.name) + "\n" + detailMetaStyle.Render(count)
}

func renderTaskDetails(t task) string {
	lines := []string{
		detailTitleStyle.Render(t.title),
		statusStyles[t.status].Render(statusIcons[t.status]) + " " +
			detailMetaStyle.Render(string(t.status)+" · created "+t.createdAt.Format(time.DateOnly)),
		"",
		detailLabelStyle.Render("Intent"),
		detailTextStyle.Render(t.intent),
		"",
		detailLabelStyle.Render("Key Tasks"),
	}

	for _, k := range t.keyTasks {
		icon := keyTodoStyle.Render(iconKeyTodo)
		desc := detailTextStyle.Render(k.description)
		if k.completed {
			icon = keyDoneStyle.Render(iconKeyDone)
			desc = keyDoneTextStyle.Render(k.description)
		}
		lines = append(lines, icon+" "+desc)
	}

	lines = append(lines,
		"",
		detailLabelStyle.Render("End State"),
		detailTextStyle.Render(t.endState),
	)
	return strings.Join(lines, "\n")
}

// renderPane draws a bordered box with the title embedded in the top border line.
func (m model) renderPane(p pane, w, h int, content string) string {
	s := inactivePane
	if m.focus == p {
		s = activePane
	}
	b := roundedBorder

	// Top edge with title
	inner := w - lipgloss.Width(b.TopLeft) - lipgloss.Width(b.TopRight)
	label := " " + p.title() + " "
	lead := 1 // border characters before the title
	if lead+lipgloss.Width(label) > inner {
		label, lead = "", 0 // too narrow for the title
	}
	fill := max(0, inner-lead-lipgloss.Width(label))

	top := s.edge.Render(b.TopLeft+strings.Repeat(b.Top, lead)) +
		s.title.Render(label) +
		s.edge.Render(strings.Repeat(b.Top, fill)+b.TopRight)

	body := s.body.Width(w).Height(h - 1).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, top, body)
}

func main() {
	m := model{
		focus:     paneList,
		groups:    sampleGroups(),
		collapsed: map[int]bool{3: true},
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error in main: %v\n", err)
		os.Exit(1)
	}
}

func date(y int, mo time.Month, d int) time.Time {
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}

// sampleGroups is placeholder data until tasks come from SQLite.
// Everything here is AI generated (for now)
func sampleGroups() []group {
	return []group{
		{id: 1, name: "Work", tasks: []task{
			{
				id:     1,
				title:  "Ship telos v0.1",
				intent: "Release a first usable version of the task manager.",
				keyTasks: []keyTask{
					{"Design the data model", true},
					{"Build the TUI layout", true},
					{"Store tasks in SQLite", false},
				},
				endState:  "A tagged v0.1 release that I use every day.",
				status:    statusActive,
				createdAt: date(2026, time.September, 29),
			},
			{
				id:     2,
				title:  "Write the README",
				intent: "Document how to install and use telos.",
				keyTasks: []keyTask{
					{"Draft the installation steps", false},
					{"Add a screenshot", false},
				},
				endState:  "A newcomer can run telos from the README alone.",
				status:    statusPaused,
				createdAt: date(2026, time.September, 30),
			},
			{
				id:     3,
				title:  "Set up CI",
				intent: "Catch build failures before they reach main.",
				keyTasks: []keyTask{
					{"Add a build workflow", true},
					{"Run go vet on every push", true},
				},
				endState:  "Every push is built and vetted automatically.",
				status:    statusCompleted,
				createdAt: date(2026, time.September, 25),
			},
		}},
		{id: 2, name: "Personal", tasks: []task{
			{
				id:     4,
				title:  "Plan a weekend trip",
				intent: "Get away for a relaxing weekend.",
				keyTasks: []keyTask{
					{"Pick a destination", true},
					{"Book accommodation", false},
					{"Pack", false},
				},
				endState:  "Two relaxing days away with everything booked in advance.",
				status:    statusActive,
				createdAt: date(2026, time.September, 27),
			},
			{
				id:     5,
				title:  "Bake sourdough",
				intent: "Bake a decent loaf at home.",
				keyTasks: []keyTask{
					{"Start a sourdough starter", false},
				},
				endState:  "A loaf I'm happy to share.",
				status:    statusCanceled,
				createdAt: date(2026, time.September, 20),
			},
		}},
		{id: 3, name: "Learning", tasks: []task{
			{
				id:     6,
				title:  "Finish a Go course",
				intent: "Get comfortable writing idiomatic Go.",
				keyTasks: []keyTask{
					{"Complete the concurrency module", false},
				},
				endState:  "I can build small Go tools without looking things up.",
				status:    statusActive,
				createdAt: date(2026, time.September, 22),
			},
		}},
		{id: 4, name: "Someday"},
	}
}
