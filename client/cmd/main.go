package main

import (
	"bufio"
	"cliChat/client"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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

	// Get username
	fmt.Println("Qual o seu nome meu bom?")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err = scanner.Err()
	if err != nil {
		log.Fatal(err)
	}

	err = c.SetName(scanner.Text())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case err := <-c.Error():
				if err == io.EOF {
					fmt.Println("Connection closed connection from server.")
				} else {
					panic(err)
				}
			case msg := <-c.Incoming():
				fmt.Printf("%v: %v\n", msg.Name, msg.Message)
			}
		}
	}()

	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err = scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		err = c.SendMessage(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
	}
}
