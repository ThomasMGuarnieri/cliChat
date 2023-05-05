package protocol

import "errors"

var (
	UnknownCommand = errors.New("unknown command")
)

type SendCommand struct {
	Message string
}

type NameCommand struct {
	Name string
}

type MessageCommand struct {
	Name    string
	Message string
}
