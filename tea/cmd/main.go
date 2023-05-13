package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log"
	"simpleChat/client"
	"simpleChat/protocol"
	"strings"
)

func main() {
	var err error

	addr := flag.String("server", "localhost:3333", "Which server to connect to")

	flag.Parse()

	c := client.NewClient()
	err = c.Dial(*addr)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	go c.Start()

	p := tea.NewProgram(initialModel(c))

	go func(program *tea.Program) {
		for {
			select {
			case err := <-c.Error():
				if err == io.EOF {
					program.Send(err)
					log.Println("Connection closed connection from server.")
					// TODO: Handle error with tea
				} else {
					// TODO: This always happen when hit ctrl + c
					//  think in another way to exit this, or fix the panic
					//  maybe we should finish this routine before exit the main
					//  program, dont know
					panic(err)
				}
			case msg := <-c.Incoming():
				program.Send(msg)
			}
		}
	}(p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
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

func initialModel(cc client.ChatClient) model {
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

	return model{
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

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnterAltScreen, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	if !m.emptyName {
		return chatView(m)
	}

	return nameView(m)
}

func chatView(m model) string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func nameView(m model) string {
	return fmt.Sprintf(
		"Say your name\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func updateChat(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
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

func updateName(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.textInput.Value()) > 4 {
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
