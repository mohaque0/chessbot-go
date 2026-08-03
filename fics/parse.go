package fics

import (
	"chessbot-go/board"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var soughtPattern = regexp.MustCompile(`^\s*(\d+)\s+(\d+|\+{4})\s+(\w+)\S*\s+(\d+)\s+(\d+)\s+(unrated|rated)\s+(standard|blitz|lightning|suicide|wild|odds)\s*(\[white\]|\[black\])?\s`)

func ficsLetterToPiece(ch byte) (board.Piece, bool) {
	switch ch {
	case 'K':
		return board.Piece{Type: board.King, Player: board.White}, true
	case 'Q':
		return board.Piece{Type: board.Queen, Player: board.White}, true
	case 'B':
		return board.Piece{Type: board.Bishop, Player: board.White}, true
	case 'N':
		return board.Piece{Type: board.Knight, Player: board.White}, true
	case 'R':
		return board.Piece{Type: board.Rook, Player: board.White}, true
	case 'P':
		return board.Piece{Type: board.Pawn, Player: board.White}, true
	case 'k':
		return board.Piece{Type: board.King, Player: board.Black}, true
	case 'q':
		return board.Piece{Type: board.Queen, Player: board.Black}, true
	case 'b':
		return board.Piece{Type: board.Bishop, Player: board.Black}, true
	case 'n':
		return board.Piece{Type: board.Knight, Player: board.Black}, true
	case 'r':
		return board.Piece{Type: board.Rook, Player: board.Black}, true
	case 'p':
		return board.Piece{Type: board.Pawn, Player: board.Black}, true
	default:
		return board.Piece{}, false
	}
}

func parseGameType(s string) GameType {
	switch s {
	case "standard":
		return Standard
	case "blitz":
		return Blitz
	case "lightning":
		return Lightning
	case "suicide":
		return Suicide
	case "odds":
		return Odds
	default:
		return Unknown
	}
}

func parseSoughtLine(line string) (FicsReceivedSoughtGame, bool) {
	m := soughtPattern.FindStringSubmatch(line)
	if m == nil {
		return FicsReceivedSoughtGame{}, false
	}

	adIdx, _ := strconv.ParseUint(m[1], 10, 64)

	var rating *uint
	if m[2] != "++++" {
		r, _ := strconv.ParseUint(m[2], 10, 64)
		ru := uint(r)
		rating = &ru
	}

	var requestedPlayer *board.Player
	if m[8] == "[white]" {
		p := board.White
		requestedPlayer = &p
	} else if m[8] == "[black]" {
		p := board.Black
		requestedPlayer = &p
	}

	return FicsReceivedSoughtGame{
		AdIdx:           uint(adIdx),
		Rating:          rating,
		Username:        m[3],
		Rated:           m[6] == "rated",
		GameType:        parseGameType(m[7]),
		RequestedPlayer: requestedPlayer,
	}, true
}

// parseStyle12 parses a FICS style-12 board line.
//
// Format: <12> row8 row7 row6 row5 row4 row3 row2 row1 color ... move_string ...
// Tokens (0-indexed after splitting):
//
//	0: <12>
//	1-8: board rows (rank 8 down to rank 1)
//	9: color to move (W or B)
//	10: double pawn push column (-1 if none)
//	11: white can castle kingside (0/1)
//	12: white can castle queenside (0/1)
//	13: black can castle kingside (0/1)
//	14: black can castle queenside (0/1)
//	19: my relation to game (1=white, -1=black, 0=observing)
//	27: move string (e.g. "P/e2-e4", "o-o", "none")
func parseStyle12(line string) (FicsReceivedBoard, bool) {
	idx := strings.Index(line, "<12>")
	if idx < 0 {
		return FicsReceivedBoard{}, false
	}
	tokens := strings.Fields(line[idx:])
	if len(tokens) < 28 {
		return FicsReceivedBoard{}, false
	}

	b := board.NewBlankBitBoard()

	for row := 0; row < 8; row++ {
		rankStr := tokens[row+1]
		if len(rankStr) != 8 {
			return FicsReceivedBoard{}, false
		}
		y := 7 - row
		for x := 0; x < 8; x++ {
			p, ok := ficsLetterToPiece(rankStr[x])
			if ok {
				b.SetPiece(p, x, y)
			}
		}
	}

	var mover board.Player
	if tokens[9] == "W" {
		mover = board.White
	} else {
		mover = board.Black
	}

	dpCol, _ := strconv.Atoi(tokens[10])
	if dpCol >= 0 && dpCol < 8 {
		b.SetPawnDoubleStep(mover.Other(), dpCol)
	}

	whiteKS, _ := strconv.Atoi(tokens[11])
	whiteQS, _ := strconv.Atoi(tokens[12])
	blackKS, _ := strconv.Atoi(tokens[13])
	blackQS, _ := strconv.Atoi(tokens[14])
	b.SetCastlingRights(board.White, whiteKS == 1, whiteQS == 1)
	b.SetCastlingRights(board.Black, blackKS == 1, blackQS == 1)

	myRelation, _ := strconv.Atoi(tokens[19])

	moveStr := tokens[27]
	var lastMove *board.Move
	if moveStr != "none" {
		mv, err := parseStyle12Move(moveStr, mover.Other())
		if err == nil {
			lastMove = &mv
		}
	}

	return FicsReceivedBoard{
		Board:      b,
		Mover:      mover,
		LastMove:   lastMove,
		MyRelation: myRelation,
	}, true
}

// parseStyle12Move parses a FICS style-12 move string.
// Formats: "P/e2-e4", "N/g1-f3", "o-o", "o-o-o", "P/a7-a8=Q"
func parseStyle12Move(moveStr string, player board.Player) (board.Move, error) {
	lower := strings.ToLower(moveStr)
	if lower == "o-o" {
		return board.Castle(board.KingsideCastle, player), nil
	}
	if lower == "o-o-o" {
		return board.Castle(board.QueensideCastle, player), nil
	}

	// Format: P/e2-e4 or P/a7-a8=Q
	parts := strings.SplitN(moveStr, "/", 2)
	if len(parts) != 2 {
		return board.Move{}, fmt.Errorf("unexpected move format: %s", moveStr)
	}
	coords := parts[1]

	// Split on '-' to get src and dst
	srcDst := strings.SplitN(coords, "-", 2)
	if len(srcDst) != 2 || len(srcDst[0]) != 2 {
		return board.Move{}, fmt.Errorf("unexpected move coordinates: %s", coords)
	}

	srcX := int(srcDst[0][0] - 'a')
	srcY := int(srcDst[0][1] - '1')

	dstPart := srcDst[1]
	var promo board.PieceType

	if eqIdx := strings.Index(dstPart, "="); eqIdx >= 0 {
		promoStr := dstPart[eqIdx+1:]
		dstPart = dstPart[:eqIdx]
		switch promoStr {
		case "Q":
			promo = board.Queen
		case "R":
			promo = board.Rook
		case "B":
			promo = board.Bishop
		case "N":
			promo = board.Knight
		default:
			return board.Move{}, fmt.Errorf("unknown promotion piece: %s", promoStr)
		}
	}

	if len(dstPart) != 2 {
		return board.Move{}, fmt.Errorf("unexpected destination: %s", dstPart)
	}

	dstX := int(dstPart[0] - 'a')
	dstY := int(dstPart[1] - '1')

	if promo != board.NoPiece {
		return board.MvPromo(srcX, srcY, dstX, dstY, promo), nil
	}
	return board.Mv(srcX, srcY, dstX, dstY), nil
}
