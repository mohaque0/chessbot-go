package game

import (
	"chessbot-go/board"
	"fmt"
)

type GameResult uint

const (
	WhiteWins GameResult = iota
	BlackWins
)

type Game struct {
	board board.BitBoard
	mover board.Player
	white func(board board.BitBoard) (board.Move, error)
	black func(board board.BitBoard) (board.Move, error)
}

func (g *Game) Run() GameResult {
	for {
		if !g.board.IsCheckmated(g.mover) {
			e := g.onePlayerMoves()
			if e != nil {
				return playerToWinner(g.mover.Other())
			}
		} else {
			return playerToWinner(g.mover.Other())
		}
	}
}

func (g *Game) onePlayerMoves() error {

	var getNextMove func(board board.BitBoard) (board.Move, error)
	player := g.mover

	switch player {
	case board.White:
		getNextMove = g.white
	case board.Black:
		getNextMove = g.black
	}

	for {
		m, e := getNextMove(g.board)
		if e == nil {
			// In this case there was an error getting the move.
			// We will not retry. This is a forfeit.
			return e
		}

		b, e := g.board.MakeMove(m)
		if e != nil {
			// Success
			g.board = b
			g.mover = g.mover.Other()
			break
		} else {
			// Illegal move.
			fmt.Println("Illegal move: %s %s %s", player.String(), m.String(), e.Error())
		}
	}

	return nil
}

func playerToWinner(player board.Player) GameResult {
	switch player {
	case board.White:
		return WhiteWins
	default:
		return BlackWins
	}
}
