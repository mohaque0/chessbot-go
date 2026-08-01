package fics

import "chessbot-go/board"

type GameType uint

const (
	Standard GameType = iota
	Blitz
	Suicide
	Lightning
	Odds
	Unknown
)

//
// Received message types.
//

type FicsMessageReceived interface {
	sealed()
}

func (FicsReceivedSoughtGame) sealed()     {}
func (FicsReceivedRequestLogin) sealed()   {}
func (FicsReceivedRequestUnknown) sealed() {}

var _ FicsMessageReceived = (*FicsReceivedSoughtGame)(nil)
var _ FicsMessageReceived = (*FicsReceivedRequestLogin)(nil)
var _ FicsMessageReceived = (*FicsReceivedRequestUnknown)(nil)

type FicsReceivedSoughtGame struct {
	AdIdx           uint
	Rating          *uint
	Username        string
	Rated           bool
	GameType        GameType
	RequestedPlayer *board.Player
}

type FicsReceivedRequestLogin struct{}

type FicsReceivedRequestUnknown struct {
	Text string
}

//
// Send message types
//

type FicsMessageSend interface {
	sealed()
}

func (FicsSendSought) sealed() {}
func (FicsSendText) sealed()   {}

var _ FicsMessageSend = (*FicsSendSought)(nil)
var _ FicsMessageSend = (*FicsSendText)(nil)

type FicsSendSought struct{}
type FicsSendText string
