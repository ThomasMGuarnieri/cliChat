package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
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
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel(cc client.ChatClient) model {
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

	vp := viewport.New(80, 20)
	// TODO: this also can be prettier
	vp.SetContent(`Bem vindo ao chat!
Seja gentil e aperte Enter.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		c:           cc,
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	// TODO: May ask for the user name here?
	//  or other screen just to get the name
	return tea.Batch(textarea.Blink, tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			err := m.c.SendMessage(m.textarea.Value())
			if err != nil {
				m.err = err
				return m, nil
			}
			m.textarea.Reset()
		}
	case protocol.MessageCommand:
		// TODO: Set the user name here
		m.messages = append(m.messages, m.senderStyle.Render("Out: ")+msg.Message)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, tea.Quit
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
