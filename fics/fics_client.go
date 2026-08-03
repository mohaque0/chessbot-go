package fics

import (
	"chessbot-go/board"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	enterAsGuestPattern = regexp.MustCompile(`Press return to enter the server as`)
	loggedInAsPattern   = regexp.MustCompile(`Logging you in as "(\w+)"`)
	gameEndPattern      = regexp.MustCompile(`\{Game \d+ \(\S+ vs\. \S+\) (.+)\} (1-0|0-1|1/2-1/2)`)
)

type FicsClient struct {
	telnet *TelnetClient
	Recv   chan FicsMessageReceived
	Send   chan FicsMessageSend
}

func NewFicsClient(debugWr io.Writer) (*FicsClient, error) {
	telnet, err := NewTelnetClient("freechess.org:23", debugWr)
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
		// Strip the "fics% " prompt if it appears at the start of the line.
		// The server sends "fics% " as a command-line prompt; it may be
		// concatenated with real data in the same read buffer.
		cleaned := strings.TrimLeft(input, " ")
		if strings.HasPrefix(cleaned, "fics%") {
			cleaned = strings.TrimLeft(cleaned[5:], " ")
		}

		switch {
		case strings.Contains(input, "login:"):
			dst <- FicsReceivedLoginPrompt{}
		case loggedInAsPattern.MatchString(input):
			m := loggedInAsPattern.FindStringSubmatch(input)
			dst <- FicsReceivedLoggedInAs{Username: m[1]}
		case enterAsGuestPattern.MatchString(input):
			dst <- FicsReceivedRequestLogin{}
		case strings.HasPrefix(cleaned, "Illegal move"):
			dst <- FicsReceivedIllegalMove{Text: cleaned}
		case gameEndPattern.MatchString(input):
			m := gameEndPattern.FindStringSubmatch(input)
			reason := m[1]
			var result GameEndResult
			switch m[2] {
			case "1-0":
				result = WhiteWins
			case "0-1":
				result = BlackWins
			default:
				result = DrawResult
			}
			dst <- FicsReceivedGameEnd{Result: result, Reason: reason, Message: input}
		case strings.Contains(cleaned, "<12>"):
			if b, ok := parseStyle12(cleaned); ok {
				dst <- b
			} else {
				dst <- FicsReceivedRequestUnknown{Text: input}
			}
		default:
			if sg, ok := parseSoughtLine(cleaned); ok {
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
			switch m.Kind {
			case board.KingsideCastle:
				dst <- "o-o"
			case board.QueensideCastle:
				dst <- "o-o-o"
			default:
				s := fmt.Sprintf("%c%c%c%c", 'a'+rune(m.SrcX), '1'+rune(m.SrcY), 'a'+rune(m.DstX), '1'+rune(m.DstY))
				if m.Promote != board.NoPiece {
					s += "=" + m.Promote.Letter()
				}
				dst <- s
			}
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
