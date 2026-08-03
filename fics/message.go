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
func (FicsReceivedLoginPrompt) sealed()    {}
func (FicsReceivedLoggedInAs) sealed()     {}
func (FicsReceivedRequestLogin) sealed()   {}
func (FicsReceivedRequestUnknown) sealed() {}
func (FicsReceivedBoard) sealed()          {}
func (FicsReceivedGameEnd) sealed()        {}
func (FicsReceivedIllegalMove) sealed()    {}

var _ FicsMessageReceived = (*FicsReceivedSoughtGame)(nil)
var _ FicsMessageReceived = (*FicsReceivedLoginPrompt)(nil)
var _ FicsMessageReceived = (*FicsReceivedLoggedInAs)(nil)
var _ FicsMessageReceived = (*FicsReceivedRequestLogin)(nil)
var _ FicsMessageReceived = (*FicsReceivedRequestUnknown)(nil)
var _ FicsMessageReceived = (*FicsReceivedBoard)(nil)
var _ FicsMessageReceived = (*FicsReceivedGameEnd)(nil)
var _ FicsMessageReceived = (*FicsReceivedIllegalMove)(nil)

type FicsReceivedSoughtGame struct {
	AdIdx           uint
	Rating          *uint
	Username        string
	Rated           bool
	GameType        GameType
	RequestedPlayer *board.Player
}

type FicsReceivedLoginPrompt struct{}

type FicsReceivedLoggedInAs struct {
	Username string
}

type FicsReceivedRequestLogin struct{}

type FicsReceivedRequestUnknown struct {
	Text string
}

type FicsReceivedBoard struct {
	Board    board.BitBoard
	Mover    board.Player
	LastMove *board.Move
}

type FicsReceivedIllegalMove struct {
	Text string
}

type GameEndResult uint

const (
	WhiteWins GameEndResult = iota
	BlackWins
	DrawResult
)

type FicsReceivedGameEnd struct {
	Result  GameEndResult
	Reason  string
	Message string
}

//
// Send message types
//

type FicsMessageSend interface {
	sealed()
}

func (FicsSendSought) sealed()    {}
func (FicsSendText) sealed()      {}
func (FicsSendPlay) sealed()      {}
func (FicsSendMove) sealed()      {}
func (FicsSendSetStyle) sealed()  {}
func (FicsSendSetNoWrap) sealed() {}

var _ FicsMessageSend = (*FicsSendSought)(nil)
var _ FicsMessageSend = (*FicsSendText)(nil)
var _ FicsMessageSend = (*FicsSendPlay)(nil)
var _ FicsMessageSend = (*FicsSendMove)(nil)
var _ FicsMessageSend = (*FicsSendSetStyle)(nil)
var _ FicsMessageSend = (*FicsSendSetNoWrap)(nil)

type FicsSendSought struct{}
type FicsSendText string
type FicsSendPlay struct{ GameID uint }
type FicsSendMove struct{ Move board.Move }
type FicsSendSetStyle struct{}
type FicsSendSetNoWrap struct{}
