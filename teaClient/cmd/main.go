package main

import (
	"cliChat/client"
	"cliChat/teaClient"
	"flag"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"log"
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

	p := tea.NewProgram(teaClient.InitialModel(c))

	go func(program *tea.Program) {
		for {
			select {
			case err := <-c.Error():
				if err == io.EOF {
					p.Kill()
					log.Fatal("Connection closed by server")
				} else {
					p.Kill()
					log.Fatal(err)
				}
			case msg := <-c.Incoming():
				program.Send(msg)
			}
		}
	}(p)

	if _, err := p.Run(); err != nil {
		p.Kill()
		log.Fatal(err)
	}
}
