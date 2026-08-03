package game

import (
	"chessbot-go/board"
	"fmt"
	"time"
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

func NewGame(white, black func(board.BitBoard) (board.Move, error)) Game {
	return Game{
		board: board.NewBitBoard(),
		mover: board.White,
		white: white,
		black: black,
	}
}

func (g *Game) Run() GameResult {

	idx := 0
	fmt.Printf("%d\n%s\n", idx, g.board.String())

	for {
		moves := g.board.GetMoves(g.mover)
		if len(moves) == 0 {
			if g.board.IsInCheck(g.mover) {
				return playerToWinner(g.mover.Other())
			}
			return Draw
		}
		start := time.Now()
		player, m, e := g.onePlayerMoves()
		elapsed := time.Since(start)
		idx++

		fmt.Printf("%d: %s played %s (%s)\n%s\n", idx, player, m, elapsed, g.board.String())
		if e != nil {
			return playerToWinner(g.mover.Other())
		}
	}
}

func (g *Game) onePlayerMoves() (board.Player, board.Move, error) {

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
			return player, board.Move{}, e
		}

		b, e := g.board.MakeMove(m)
		if e != nil {
			// Illegal move.
			fmt.Printf("Illegal move: %s %s %s\n", player.String(), m.String(), e.Error())
		} else {
			// Success
			g.board = b
			g.mover = g.mover.Other()
			return player, m, nil
		}
	}
}

func playerToWinner(player board.Player) GameResult {
	switch player {
	case board.White:
		return WhiteWins
	default:
		return BlackWins
	}
}
