package teaClient

import (
	"cliChat/client"
	"cliChat/protocol"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type (
	errMsg error
)

type Model struct {
	c           client.ChatClient
	emptyName   bool
	err         error
	messages    []string
	name        string
	senderStyle lipgloss.Style
	textInput   textinput.Model
	textarea    textarea.Model
	viewport    viewport.Model
}

func InitialModel(cc client.ChatClient) Model {
	// TEXT AREA
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	// TODO: This can be prettier
	ta.Prompt = "┃ "
	ta.CharLimit = 144

	ta.SetWidth(30)
	ta.SetHeight(2)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	// VIEWPORT
	vp := viewport.New(80, 20)
	vp.SetContent(`Bem vindo ao chat!
Seja gentil e aperte Enter.`)

	// TEXT INPUT
	ti := textinput.New()
	ti.Placeholder = "Valter Branco"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		c:           cc,
		emptyName:   true,
		err:         nil,
		messages:    []string{},
		name:        "Out",
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		textarea:    ta,
		textInput:   ti,
		viewport:    vp,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnterAltScreen, textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "esc" || k == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if !m.emptyName {
		return updateChat(msg, m)
	}

	return updateName(msg, m)
}

func (m Model) View() string {
	if !m.emptyName {
		return chatView(m)
	}

	return nameView(m)
}

func chatView(m Model) string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func nameView(m Model) string {
	return fmt.Sprintf(
		"Say your name\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func updateChat(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			err := m.c.SendMessage(m.textarea.Value())
			if err != nil {
				m.err = err
				return m, nil
			}
			m.textarea.Reset()
		}
	case protocol.MessageCommand:
		m.messages = append(m.messages, m.senderStyle.Render(msg.Name+": ")+msg.Message)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	case errMsg:
		m.err = msg
		return m, tea.Quit
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func updateName(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.textInput.Value()) > 4 && len(m.textInput.Value()) < protocol.NameSize {
				m.emptyName = false
				m.name = m.textInput.Value()
				err := m.c.SetName(m.name)
				if err != nil {
					m.err = err
					return m, nil
				}
			}
			return m, cmd
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}
