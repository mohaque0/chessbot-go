package controller

import (
	"chessbot-go/board"
	"errors"
	"fmt"
	"math"
)

type alphaBetaKind uint

const (
	Minimizer alphaBetaKind = iota
	Maximizer
)

func AlphaBeta(b board.BitBoard, player board.Player, depth uint) (board.Move, error) {
	moves := b.GetMoves(player)
	if len(moves) == 0 {
		return board.Move{}, errors.New("no more moves")
	}

	alpha := math.MinInt
	beta := math.MaxInt
	value := math.MinInt
	bestMove := moves[0]
	for _, mv := range moves {
		candidate := alphaBetaRecursion(Minimizer, b.ApplyMove(mv), player, depth-1, alpha, beta)
		if candidate > value {
			value = candidate
			bestMove = mv
		}
		if value > alpha {
			alpha = value
		}
		if alpha > beta {
			break
		}
	}

	return bestMove, nil
}

func alphaBetaRecursion(kind alphaBetaKind, board board.BitBoard, player board.Player, depth uint, alpha int, beta int) int {
	if depth == 0 {
		return utility(board, player)
	}

	switch kind {
	case Maximizer:
		moves := board.GetMoves(player)
		if len(moves) == 0 {
			if board.IsInCheck(player) {
				// Checkmate
				return -10000
			} else {
				return utility(board, player)
			}
		}

		value := math.MinInt
		for _, mv := range moves {
			candidate := alphaBetaRecursion(Minimizer, board.ApplyMove(mv), player, depth-1, alpha, beta)
			if candidate > value {
				value = candidate
			}
			if value > alpha {
				alpha = value
			}
			if alpha > beta {
				break
			}
		}

		return value

	case Minimizer:
		mover := player.Other()
		moves := board.GetMoves(mover)
		if len(moves) == 0 {
			if board.IsInCheck(mover) {
				// Checkmate
				return 10000
			} else {
				return utility(board, player)
			}
		}

		value := math.MaxInt
		for _, mv := range moves {
			candidate := alphaBetaRecursion(Maximizer, board.ApplyMove(mv), player, depth-1, alpha, beta)
			if candidate < value {
				value = candidate
			}
			if value < beta {
				beta = value
			}
			if alpha > beta {
				break
			}
		}

		return value
	}

	fmt.Println("Unknown AlphaBeta behavior.")
	return 0
}

func utility(b board.BitBoard, player board.Player) int {
	score := 0
	for idx := range 64 {
		x, y := board.IdxToCoord(idx)
		p, found := b.GetPiece(x, y)
		if found {
			piece_score := 0
			piece_factor := 1
			if p.Player != player {
				piece_factor = -1
			}

			switch p.Type {
			case board.NoPiece:
			case board.King:
				piece_score += 10000
			case board.Queen:
				piece_score += 50
			case board.Bishop:
				piece_score += 30
			case board.Knight:
				piece_score += 30
			case board.Rook:
				piece_score += 50
			case board.Pawn:
				piece_score += 15
			}

			score += piece_score * piece_factor
		}
	}

	return score
}
