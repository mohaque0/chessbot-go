package board

import "fmt"

// flags bitfield layout:
//
//	bits 0-3:   castling rights
//	bits 8-15:  white pawn double-step column flags
//	bits 16-23: black pawn double-step column flags
const (
	flagWhiteKingside  uint32 = 1 << 0
	flagWhiteQueenside uint32 = 1 << 1
	flagBlackKingside  uint32 = 1 << 2
	flagBlackQueenside uint32 = 1 << 3
	whitePawnDSShift          = 8
	blackPawnDSShift          = 16
	pawnDSMask         uint32 = 0xFF<<whitePawnDSShift | 0xFF<<blackPawnDSShift
)

type BitBoard struct {
	kings   [2]Bits
	queens  [2]Bits
	bishops [2]Bits
	knights [2]Bits
	rooks   [2]Bits
	pawns   [2]Bits
	flags   uint32
}

// NewBitBoard returns the standard starting position.
func NewBitBoard() BitBoard {
	var b BitBoard

	for col := 0; col < 8; col++ {
		b.pawns[White].Set(col, 1)
		b.pawns[Black].Set(col, 6)
	}

	b.rooks[White].Set(0, 0)
	b.knights[White].Set(1, 0)
	b.bishops[White].Set(2, 0)
	b.queens[White].Set(3, 0)
	b.kings[White].Set(4, 0)
	b.bishops[White].Set(5, 0)
	b.knights[White].Set(6, 0)
	b.rooks[White].Set(7, 0)

	b.rooks[Black].Set(0, 7)
	b.knights[Black].Set(1, 7)
	b.bishops[Black].Set(2, 7)
	b.queens[Black].Set(3, 7)
	b.kings[Black].Set(4, 7)
	b.bishops[Black].Set(5, 7)
	b.knights[Black].Set(6, 7)
	b.rooks[Black].Set(7, 7)

	b.flags = flagWhiteKingside | flagWhiteQueenside | flagBlackKingside | flagBlackQueenside
	return b
}

// NewBlankBitBoard returns an empty board with no pieces or castling rights.
func NewBlankBitBoard() BitBoard { return BitBoard{} }

// ---------------------------------------------------------------------------
// Flag accessors
// ---------------------------------------------------------------------------

func (b *BitBoard) CanCastleKingside(p Player) bool {
	if p == White {
		return b.flags&flagWhiteKingside != 0
	}
	return b.flags&flagBlackKingside != 0
}

func (b *BitBoard) CanCastleQueenside(p Player) bool {
	if p == White {
		return b.flags&flagWhiteQueenside != 0
	}
	return b.flags&flagBlackQueenside != 0
}

func (b *BitBoard) SetCastlingRights(p Player, kingside, queenside bool) {
	ks, qs := flagWhiteKingside, flagWhiteQueenside
	if p == Black {
		ks, qs = flagBlackKingside, flagBlackQueenside
	}
	b.flags &^= ks | qs
	if kingside {
		b.flags |= ks
	}
	if queenside {
		b.flags |= qs
	}
}

func (b *BitBoard) PawnDoubleStep(p Player, col int) bool {
	shift := whitePawnDSShift
	if p == Black {
		shift = blackPawnDSShift
	}
	return b.flags&(1<<uint(shift+col)) != 0
}

func (b *BitBoard) SetPawnDoubleStep(p Player, col int) {
	shift := whitePawnDSShift
	if p == Black {
		shift = blackPawnDSShift
	}
	b.flags |= 1 << uint(shift+col)
}

func (b *BitBoard) clearAllPawnDoubleSteps() {
	b.flags &^= pawnDSMask
}

func (b *BitBoard) revokeCastleRightsForSquare(x, y int) {
	switch {
	case y == 0 && x == 0:
		b.flags &^= flagWhiteQueenside
	case y == 0 && x == 4:
		b.flags &^= flagWhiteKingside | flagWhiteQueenside
	case y == 0 && x == 7:
		b.flags &^= flagWhiteKingside
	case y == 7 && x == 0:
		b.flags &^= flagBlackQueenside
	case y == 7 && x == 4:
		b.flags &^= flagBlackKingside | flagBlackQueenside
	case y == 7 && x == 7:
		b.flags &^= flagBlackKingside
	}
}

// ---------------------------------------------------------------------------
// Piece placement
// ---------------------------------------------------------------------------

func (b *BitBoard) pieceBits(pt PieceType, p Player) *Bits {
	switch pt {
	case King:
		return &b.kings[p]
	case Queen:
		return &b.queens[p]
	case Bishop:
		return &b.bishops[p]
	case Knight:
		return &b.knights[p]
	case Rook:
		return &b.rooks[p]
	case Pawn:
		return &b.pawns[p]
	}
	return nil
}

func (b *BitBoard) clearSquare(p Player, x, y int) {
	if x > 7 || y > 7 {
		return
	}
	b.kings[p].Unset(x, y)
	b.queens[p].Unset(x, y)
	b.bishops[p].Unset(x, y)
	b.knights[p].Unset(x, y)
	b.rooks[p].Unset(x, y)
	b.pawns[p].Unset(x, y)
}

func (b *BitBoard) clearAllAt(x, y int) {
	b.clearSquare(White, x, y)
	b.clearSquare(Black, x, y)
}

func (b *BitBoard) SetPiece(piece Piece, x, y int) {
	b.clearSquare(piece.Player, x, y)
	b.pieceBits(piece.Type, piece.Player).Set(x, y)
}

func (b *BitBoard) DelPiece(x, y int) {
	b.clearAllAt(x, y)
}

func (b *BitBoard) GetPiece(x, y int) (Piece, bool) {
	for p := White; p <= Black; p++ {
		if b.kings[p].Get(x, y) {
			return Piece{King, p}, true
		}
		if b.queens[p].Get(x, y) {
			return Piece{Queen, p}, true
		}
		if b.bishops[p].Get(x, y) {
			return Piece{Bishop, p}, true
		}
		if b.knights[p].Get(x, y) {
			return Piece{Knight, p}, true
		}
		if b.rooks[p].Get(x, y) {
			return Piece{Rook, p}, true
		}
		if b.pawns[p].Get(x, y) {
			return Piece{Pawn, p}, true
		}
	}
	return Piece{}, false
}

func (b *BitBoard) occupied(p Player) Bits {
	return b.kings[p] | b.queens[p] | b.bishops[p] | b.knights[p] | b.rooks[p] | b.pawns[p]
}

func (b *BitBoard) allOccupied() Bits {
	return b.occupied(White) | b.occupied(Black)
}

// ---------------------------------------------------------------------------
// Move application
// ---------------------------------------------------------------------------

// ApplyMove returns a new board with the move applied. No legality check.
func (b *BitBoard) ApplyMove(m Move) BitBoard {
	result := *b
	result.clearAllPawnDoubleSteps()

	switch m.Kind {
	case NormalMove:
		srcX, srcY := int(m.SrcX), int(m.SrcY)
		dstX, dstY := int(m.DstX), int(m.DstY)

		piece, ok := b.GetPiece(srcX, srcY)
		if !ok {
			return result
		}
		if m.Promote != NoPiece {
			piece = Piece{m.Promote, piece.Player}
		}

		result.revokeCastleRightsForSquare(srcX, srcY)
		result.revokeCastleRightsForSquare(dstX, dstY)
		result.clearAllAt(srcX, srcY)
		result.clearAllAt(dstX, dstY)

		if piece.Type == Pawn {
			dy := dstY - srcY
			if dy < 0 {
				dy = -dy
			}
			if dy == 2 {
				result.SetPawnDoubleStep(piece.Player, dstX)
			}
			// En passant: the captured pawn is on the same rank as the source,
			// not on the destination square.
			if piece.Player == White && srcY == 4 && dstX != srcX && b.PawnDoubleStep(Black, dstX) {
				result.clearSquare(Black, dstX, srcY)
			}
			if piece.Player == Black && srcY == 3 && dstX != srcX && b.PawnDoubleStep(White, dstX) {
				result.clearSquare(White, dstX, srcY)
			}
		}

		result.pieceBits(piece.Type, piece.Player).Set(dstX, dstY)

	case KingsideCastle:
		y := 0
		if m.Player == Black {
			y = 7
		}
		result.clearSquare(m.Player, 4, y)
		result.clearSquare(m.Player, 7, y)
		result.kings[m.Player].Set(6, y)
		result.rooks[m.Player].Set(5, y)
		result.revokeCastleRightsForSquare(4, y)
		result.revokeCastleRightsForSquare(7, y)

	case QueensideCastle:
		y := 0
		if m.Player == Black {
			y = 7
		}
		result.clearSquare(m.Player, 0, y)
		result.clearSquare(m.Player, 1, y)
		result.clearSquare(m.Player, 4, y)
		result.kings[m.Player].Set(2, y)
		result.rooks[m.Player].Set(3, y)
		result.revokeCastleRightsForSquare(0, y)
		result.revokeCastleRightsForSquare(4, y)
	}

	return result
}

// MakeMove validates and applies a move, returning an error for illegal moves.
func (b *BitBoard) MakeMove(m Move) (BitBoard, error) {
	var player Player
	switch m.Kind {
	case NormalMove:
		p, ok := b.GetPiece(int(m.SrcX), int(m.SrcY))
		if !ok {
			return BitBoard{}, fmt.Errorf("no piece at source")
		}
		player = p.Player
	default:
		player = m.Player
	}
	for _, legal := range b.GetMoves(player) {
		if legal == m {
			return b.ApplyMove(m), nil
		}
	}
	return BitBoard{}, fmt.Errorf("invalid move")
}

// ---------------------------------------------------------------------------
// Attack detection
// ---------------------------------------------------------------------------

// allAttacked returns all squares controlled by player (including empty ones).
func (b *BitBoard) allAttacked(player Player) Bits {
	occupied := b.allOccupied()
	var result Bits

	for _, info := range [3]struct {
		bb Bits
		pt PieceType
	}{
		{b.queens[player], Queen},
		{b.bishops[player], Bishop},
		{b.rooks[player], Rook},
	} {
		for remaining := info.bb; remaining != 0; remaining = remaining.ClearLSB() {
			pieceIdx := remaining.LSB()
			for _, dir := range pieceRayMoves(info.pt) {
				ray := Bits(RayMask[dir][pieceIdx])
				collisions := ray & occupied
				if collisions.IsEmpty() {
					result |= ray
				} else {
					firstIdx := findFirstCollisionIdx(dir, collisions)
					result |= ray &^ Bits(RayMask[dir][firstIdx])
				}
			}
		}
	}

	for remaining := b.kings[player]; remaining != 0; remaining = remaining.ClearLSB() {
		result |= Bits(KingMoves[remaining.LSB()])
	}
	for remaining := b.knights[player]; remaining != 0; remaining = remaining.ClearLSB() {
		result |= Bits(KnightMoves[remaining.LSB()])
	}
	pawnTbl := &WhitePawnAttacks
	if player == Black {
		pawnTbl = &BlackPawnAttacks
	}
	for remaining := b.pawns[player]; remaining != 0; remaining = remaining.ClearLSB() {
		result |= Bits((*pawnTbl)[remaining.LSB()])
	}

	return result
}

// IsAttacking reports whether player attacks square (x, y).
func (b *BitBoard) IsAttacking(player Player, x, y int) bool {
	return b.allAttacked(player).Get(x, y)
}

// IsInCheck reports whether the given player's king is in check.
func (b *BitBoard) IsInCheck(player Player) bool {
	if b.kings[player].IsEmpty() {
		return false
	}
	kingIdx := b.kings[player].LSB()
	x, y := IdxToCoord(kingIdx)
	return b.IsAttacking(player.Other(), x, y)
}

// IsCheckmated reports whether the given player is in checkmate.
func (b *BitBoard) IsCheckmated(player Player) bool {
	return b.IsInCheck(player) && len(b.GetMoves(player)) == 0
}

// ---------------------------------------------------------------------------
// Move generation
// ---------------------------------------------------------------------------

func pieceRayMoves(pt PieceType) []RayMove {
	switch pt {
	case Queen:
		return QueenMoves[:]
	case Bishop:
		return BishopMoves[:]
	case Rook:
		return RookMoves[:]
	default:
		return nil
	}
}

func findFirstCollisionIdx(dir RayMove, collisions Bits) int {
	if dir <= E {
		return collisions.LSB()
	}
	return collisions.MSB()
}

func (b *BitBoard) getPieceMoves(pt PieceType, player Player, coordIdx int) []Move {
	var moves []Move
	srcX, srcY := IdxToCoord(coordIdx)

	occupied := b.allOccupied()
	selfOccupied := b.occupied(player)
	enemyOccupied := b.occupied(player.Other())

	// Ray moves (queen, bishop, rook).
	for _, dir := range pieceRayMoves(pt) {
		ray := Bits(RayMask[dir][coordIdx])
		collisions := ray & occupied
		valid := ray

		if !collisions.IsEmpty() {
			firstIdx := findFirstCollisionIdx(dir, collisions)
			firstBit := SingleBitIdx(firstIdx)
			rayFromCollision := firstBit | Bits(RayMask[dir][firstIdx])
			valid = ray &^ rayFromCollision

			if firstBit&enemyOccupied != 0 {
				valid |= firstBit
			}
		}

		for remaining := valid; remaining != 0; remaining = remaining.ClearLSB() {
			dstX, dstY := IdxToCoord(remaining.LSB())
			moves = append(moves, Mv(srcX, srcY, dstX, dstY))
		}
	}

	// Space moves (king, knight, pawn).
	var spaceMoves Bits
	switch pt {
	case King:
		spaceMoves = Bits(KingMoves[coordIdx])
	case Knight:
		spaceMoves = Bits(KnightMoves[coordIdx])
	case Pawn:
		if player == White {
			if srcY == 1 {
				spaceMoves = Bits(WhitePawnStartMoves[coordIdx])
			} else {
				spaceMoves = Bits(WhitePawnMoves[coordIdx])
			}
		} else {
			if srcY == 6 {
				spaceMoves = Bits(BlackPawnStartMoves[coordIdx])
			} else {
				spaceMoves = Bits(BlackPawnMoves[coordIdx])
			}
		}
	}

	spaceMoves &^= selfOccupied

	if pt == Pawn {
		// Pawn cannot capture by moving forward.
		spaceMoves &^= enemyOccupied

		// Pawn cannot jump over a piece.
		selfBit := SingleBitIdx(coordIdx)
		others := occupied &^ selfBit
		if player == White {
			spaceMoves &^= ShiftUp(others, 1)
		} else {
			spaceMoves &^= ShiftDown(others, 1)
		}

		// Pawn diagonal captures.
		var pawnAtk Bits
		if player == White {
			pawnAtk = Bits(WhitePawnAttacks[coordIdx])
		} else {
			pawnAtk = Bits(BlackPawnAttacks[coordIdx])
		}
		spaceMoves |= pawnAtk & enemyOccupied

		// En passant.
		b.generateEnPassant(&moves, player, srcX, srcY)
	}

	// Prevent king from stepping into an attacked square.
	if pt == King {
		attacked := b.allAttacked(player.Other())
		for remaining := spaceMoves; remaining != 0; remaining = remaining.ClearLSB() {
			x, y := IdxToCoord(remaining.LSB())
			if attacked.Get(x, y) {
				spaceMoves.Unset(x, y)
			}
		}
	}

	for remaining := spaceMoves; remaining != 0; remaining = remaining.ClearLSB() {
		dstX, dstY := IdxToCoord(remaining.LSB())
		addMoveWithPromotions(&moves, pt, player, srcX, srcY, dstX, dstY)
	}

	return moves
}

func (b *BitBoard) generateEnPassant(moves *[]Move, player Player, srcX, srcY int) {
	opponent := player.Other()

	if player == White && srcY == 4 {
		if srcX > 0 && b.pawns[opponent].Get(srcX-1, srcY) && b.PawnDoubleStep(opponent, srcX-1) {
			*moves = append(*moves, Mv(srcX, srcY, srcX-1, srcY+1))
		}
		if srcX < 7 && b.pawns[opponent].Get(srcX+1, srcY) && b.PawnDoubleStep(opponent, srcX+1) {
			*moves = append(*moves, Mv(srcX, srcY, srcX+1, srcY+1))
		}
	}
	if player == Black && srcY == 3 {
		if srcX > 0 && b.pawns[opponent].Get(srcX-1, srcY) && b.PawnDoubleStep(opponent, srcX-1) {
			*moves = append(*moves, Mv(srcX, srcY, srcX-1, srcY-1))
		}
		if srcX < 7 && b.pawns[opponent].Get(srcX+1, srcY) && b.PawnDoubleStep(opponent, srcX+1) {
			*moves = append(*moves, Mv(srcX, srcY, srcX+1, srcY-1))
		}
	}
}

func addMoveWithPromotions(moves *[]Move, pt PieceType, player Player, srcX, srcY, dstX, dstY int) {
	if pt != Pawn || (dstY != 0 && dstY != 7) {
		*moves = append(*moves, Mv(srcX, srcY, dstX, dstY))
		return
	}
	for _, promo := range [4]PieceType{Queen, Bishop, Knight, Rook} {
		*moves = append(*moves, MvPromo(srcX, srcY, dstX, dstY, promo))
	}
}

func (b *BitBoard) getPseudoMovesNoCastling(player Player) []Move {
	var moves []Move
	for _, info := range [6]struct {
		bb Bits
		pt PieceType
	}{
		{b.kings[player], King},
		{b.queens[player], Queen},
		{b.bishops[player], Bishop},
		{b.knights[player], Knight},
		{b.rooks[player], Rook},
		{b.pawns[player], Pawn},
	} {
		for remaining := info.bb; remaining != 0; remaining = remaining.ClearLSB() {
			moves = append(moves, b.getPieceMoves(info.pt, player, remaining.LSB())...)
		}
	}
	return moves
}

func (b *BitBoard) getPseudoMoves(player Player) []Move {
	moves := b.getPseudoMovesNoCastling(player)

	y := 0
	if player == Black {
		y = 7
	}
	opponent := player.Other()

	// Kingside castle.
	if b.CanCastleKingside(player) {
		kp, kok := b.GetPiece(4, y)
		rp, rok := b.GetPiece(7, y)
		occ := b.allOccupied()
		if kok && kp == (Piece{King, player}) && rok && rp == (Piece{Rook, player}) &&
			!occ.Get(5, y) && !occ.Get(6, y) &&
			!b.IsAttacking(opponent, 4, y) &&
			!b.IsAttacking(opponent, 5, y) &&
			!b.IsAttacking(opponent, 6, y) {
			moves = append(moves, Castle(KingsideCastle, player))
		}
	}

	// Queenside castle.
	if b.CanCastleQueenside(player) {
		kp, kok := b.GetPiece(4, y)
		rp, rok := b.GetPiece(0, y)
		occ := b.allOccupied()
		if kok && kp == (Piece{King, player}) && rok && rp == (Piece{Rook, player}) &&
			!occ.Get(1, y) && !occ.Get(2, y) && !occ.Get(3, y) &&
			!b.IsAttacking(opponent, 4, y) &&
			!b.IsAttacking(opponent, 3, y) &&
			!b.IsAttacking(opponent, 2, y) {
			moves = append(moves, Castle(QueensideCastle, player))
		}
	}

	return moves
}

// GetMoves returns all legal moves for the given player.
func (b *BitBoard) GetMoves(player Player) []Move {
	pseudo := b.getPseudoMoves(player)
	var legal []Move
	for _, m := range pseudo {
		after := b.ApplyMove(m)
		if !after.IsInCheck(player) {
			legal = append(legal, m)
		}
	}
	return legal
}
