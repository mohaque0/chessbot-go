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

// Flags encapsulates castling rights and pawn double-step state in a bitfield.
//
//	bits 0-3:   castling rights
//	bits 8-15:  white pawn double-step column flags
//	bits 16-23: black pawn double-step column flags
type Flags uint32

const (
	flagWhiteKingside  Flags = 1 << 0
	flagWhiteQueenside Flags = 1 << 1
	flagBlackKingside  Flags = 1 << 2
	flagBlackQueenside Flags = 1 << 3
	whitePawnDSShift         = 8
	blackPawnDSShift         = 16
	pawnDSMask         Flags = 0xFF<<whitePawnDSShift | 0xFF<<blackPawnDSShift
	allCastlingRights  Flags = flagWhiteKingside | flagWhiteQueenside | flagBlackKingside | flagBlackQueenside
)

func NewFlags() Flags { return allCastlingRights }

func (f Flags) CanCastleKingside(p Player) bool {
	if p == White {
		return f&flagWhiteKingside != 0
	}
	return f&flagBlackKingside != 0
}

func (f Flags) CanCastleQueenside(p Player) bool {
	if p == White {
		return f&flagWhiteQueenside != 0
	}
	return f&flagBlackQueenside != 0
}

func (f Flags) WithCastlingRights(p Player, kingside, queenside bool) Flags {
	ks, qs := flagWhiteKingside, flagWhiteQueenside
	if p == Black {
		ks, qs = flagBlackKingside, flagBlackQueenside
	}
	f &^= ks | qs
	if kingside {
		f |= ks
	}
	if queenside {
		f |= qs
	}
	return f
}

func (f Flags) PawnDoubleStep(p Player, col int) bool {
	shift := whitePawnDSShift
	if p == Black {
		shift = blackPawnDSShift
	}
	return f&(1<<uint(shift+col)) != 0
}

func (f Flags) WithPawnDoubleStep(p Player, col int) Flags {
	shift := whitePawnDSShift
	if p == Black {
		shift = blackPawnDSShift
	}
	return f | 1<<uint(shift+col)
}

func (f Flags) ClearedPawnDoubleSteps() Flags {
	return f &^ pawnDSMask
}

func (f Flags) RevokedForSquare(x, y int) Flags {
	switch {
	case y == 0 && x == 0:
		f &^= flagWhiteQueenside
	case y == 0 && x == 4:
		f &^= flagWhiteKingside | flagWhiteQueenside
	case y == 0 && x == 7:
		f &^= flagWhiteKingside
	case y == 7 && x == 0:
		f &^= flagBlackQueenside
	case y == 7 && x == 4:
		f &^= flagBlackKingside | flagBlackQueenside
	case y == 7 && x == 7:
		f &^= flagBlackKingside
	}
	return f
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
