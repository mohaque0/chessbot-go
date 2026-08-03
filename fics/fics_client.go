package fics

import (
	"fmt"
	"regexp"
	"strings"
)

var enterAsGuestPattern = regexp.MustCompile(`Press return to enter the server as`)

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
		case strings.Contains(input, "<12>"):
			if b, ok := parseStyle12(input); ok {
				dst <- b
			} else {
				dst <- FicsReceivedRequestUnknown{Text: input}
			}
		default:
			if sg, ok := parseSoughtLine(input); ok {
				dst <- sg
			} else {
				dst <- FicsReceivedRequestUnknown{Text: input}
			}
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
		case FicsSendPlay:
			dst <- fmt.Sprintf("play %d", msg.GameID)
		case FicsSendMove:
			m := msg.Move
			dst <- fmt.Sprintf("%c%c%c%c", 'a'+rune(m.SrcX), '1'+rune(m.SrcY), 'a'+rune(m.DstX), '1'+rune(m.DstY))
		case FicsSendSetStyle:
			dst <- "set style 12"
		case FicsSendSetNoWrap:
			dst <- "iset nowrap 1"
		default:
			fmt.Print("Error: Unknown send message type")
			return
		}
	}
}
