package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	ACTION uint8 = iota
	RUNNING_SERVER
	SELECT_URL
	SELECT_CONNECTIONS
	SELECT_MESSAGES
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list   list.Model
	cursor int
	state  uint8
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		// check which key was pressed
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter", " ":
			fmt.Printf("enter pressed: %d", m.list.Index())
			switch m.list.Index() {
			case 0:
				m.state = RUNNING_SERVER
			case 1:
				m.state = SELECT_URL
			default:
				m.state = ACTION
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case ACTION:
		return docStyle.Render(m.list.View())

	case RUNNING_SERVER:
		return docStyle.Render("RUNNING_SERVER")

	case SELECT_URL:
		return docStyle.Render("SELECT_URL")

	case SELECT_CONNECTIONS:
		return docStyle.Render("SELECT_CONNECTIONS")

	case SELECT_MESSAGES:
		return docStyle.Render("SELECT_MESSAGES")

	default:
		return docStyle.Render("DEFAULT")
	}
}

func StartCli() {
	items := []list.Item{
		item{title: "Start Server", desc: "Start running the server normally (default)"},
		item{title: "Start Benchmark", desc: "Start running a benchmark of a server"},
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Choose an Action"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
