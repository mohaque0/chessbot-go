package board

import "fmt"

type Player uint8

const (
	White Player = 0
	Black Player = 1
)

func (p Player) Other() Player { return 1 - p }

func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

type PieceType uint8

const (
	NoPiece PieceType = iota
	King
	Queen
	Bishop
	Knight
	Rook
	Pawn
)

func (pt PieceType) Letter() string {
	switch pt {
	case King:
		return "K"
	case Queen:
		return "Q"
	case Bishop:
		return "B"
	case Knight:
		return "N"
	case Rook:
		return "R"
	case Pawn:
		return "P"
	default:
		return ""
	}
}

type Piece struct {
	Type   PieceType
	Player Player
}

type MoveKind uint8

const (
	NormalMove MoveKind = iota
	KingsideCastle
	QueensideCastle
)

type Move struct {
	Kind    MoveKind
	SrcX    uint8
	SrcY    uint8
	DstX    uint8
	DstY    uint8
	Promote PieceType
	Player  Player
}

func Mv(srcX, srcY, dstX, dstY int) Move {
	return Move{
		Kind: NormalMove,
		SrcX: uint8(srcX), SrcY: uint8(srcY),
		DstX: uint8(dstX), DstY: uint8(dstY),
	}
}

func MvPromo(srcX, srcY, dstX, dstY int, promo PieceType) Move {
	return Move{
		Kind: NormalMove,
		SrcX: uint8(srcX), SrcY: uint8(srcY),
		DstX: uint8(dstX), DstY: uint8(dstY),
		Promote: promo,
	}
}

func Castle(kind MoveKind, player Player) Move {
	return Move{Kind: kind, Player: player}
}

func (m Move) String() string {
	switch m.Kind {
	case KingsideCastle:
		return "0-0"
	case QueensideCastle:
		return "0-0-0"
	default:
		s := fmt.Sprintf("%c%c %c%c",
			'a'+rune(m.SrcX), '1'+rune(m.SrcY),
			'a'+rune(m.DstX), '1'+rune(m.DstY))
		if m.Promote != NoPiece {
			s += m.Promote.Letter()
		}
		return s
	}
}
