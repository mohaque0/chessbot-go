package game

import (
	"chessbot-go/board"
	"chessbot-go/controller"
	"fmt"
)

type GameResult uint

const (
	WhiteWins GameResult = iota
	BlackWins
	Draw
)

type Game struct {
	board board.BitBoard
	mover board.Player
	white func(board board.BitBoard) (board.Move, error)
	black func(board board.BitBoard) (board.Move, error)
}

func NewGame() Game {
	return Game{
		board: board.NewBitBoard(),
		mover: board.White,
		white: func(b board.BitBoard) (board.Move, error) { return controller.AlphaBeta(b, board.White, 5) },
		black: func(b board.BitBoard) (board.Move, error) { return controller.AlphaBeta(b, board.Black, 5) },
	}
}

func (g *Game) Run() GameResult {
	for {
		moves := g.board.GetMoves(g.mover)
		if len(moves) == 0 {
			if g.board.IsInCheck(g.mover) {
				return playerToWinner(g.mover.Other())
			}
			return Draw
		}
		e := g.onePlayerMoves()
		if e != nil {
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
		if e != nil {
			// In this case there was an error getting the move.
			// We will not retry. This is a forfeit.
			return e
		}

		b, e := g.board.MakeMove(m)
		if e != nil {
			// Illegal move.
			fmt.Printf("Illegal move: %s %s %s\n", player.String(), m.String(), e.Error())
		} else {
			// Success
			g.board = b
			g.mover = g.mover.Other()
			break
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
