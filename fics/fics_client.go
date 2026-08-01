package fics

import (
	"fmt"
	"regexp"
)

var enterAsGuestPattern = regexp.MustCompile(`"Press return to enter the server as \"(?<username>.*)\"`)

type FicsClient struct {
	telnet *TelnetClient
	Recv   chan FicsMessageReceived
	Send   chan FicsMessageSend
}

func NewFicsClient() (*FicsClient, error) {
	telnet, err := NewTelnetClient("freechess.org:23")
	if err != nil {
		return nil, err
	}

	recv := make(chan FicsMessageReceived)
	send := make(chan FicsMessageSend)

	go readFromFics(telnet.Recv, recv)
	go sendToFics(send, telnet.Send)

	return &FicsClient{
		telnet,
		recv,
		send,
	}, nil
}

func (c *FicsClient) Close() {
	c.telnet.Close()
}

func readFromFics(src chan string, dst chan FicsMessageReceived) {
	defer close(dst)
	for input := range src {
		switch {
		case enterAsGuestPattern.MatchString(input):
			dst <- FicsReceivedRequestLogin{}
		default:
			dst <- FicsReceivedRequestUnknown{Text: input}
		}
	}
}

func sendToFics(src chan FicsMessageSend, dst chan string) {
	defer close(dst)
	for msg := range src {
		switch msg := msg.(type) {
		case FicsSendSought:
			dst <- "sought"
		case FicsSendText:
			dst <- string(msg)
		default:
			fmt.Print("Error: Unknown")
			return
		}
	}
}
