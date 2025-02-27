package utils

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(0)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	focusedStyle      = lipgloss.NewStyle()
	blurredStyle      = lipgloss.NewStyle()
	cursorStyle       = focusedStyle.Copy()
	noStyle           = lipgloss.NewStyle()
)

type item string

type inputModel struct {
	textInput textinput.Model
	err       error
}

func initialInputModel() inputModel {
	ti := textinput.New()
	ti.Placeholder = "Pinata JWT"
	ti.Focus()
	ti.Width = 35
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	ti.Prompt = "> "
	ti.PromptStyle = itemStyle
	ti.TextStyle = itemStyle

	return inputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		"Enter your Pinata JWT",
		m.textInput.View(),
		"(press enter to submit)",
	) + "\n"
}

func GetInput(placeholder string) (string, error) {
	p := tea.NewProgram(initialInputModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	input := m.(inputModel).textInput.Value()
	return input, nil
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s set as the default gateway", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("No gateway selected")
	}
	return "\n" + m.list.View()
}

func MultiSelect(options []string) (string, error) {
	items := make([]list.Item, len(options))
	for i, option := range options {
		items[i] = item(option)
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a gateway"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if program, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		return "", err
	} else {
		return program.(model).choice, nil
	}
}
