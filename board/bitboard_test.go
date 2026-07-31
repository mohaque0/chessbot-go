package board

import (
	"testing"
)

func TestNewBitBoardInitialPosition(t *testing.T) {
	b := NewBitBoard()

	tests := []struct {
		x, y int
		p    Piece
	}{
		{0, 0, Piece{Rook, White}},
		{1, 0, Piece{Knight, White}},
		{2, 0, Piece{Bishop, White}},
		{3, 0, Piece{Queen, White}},
		{4, 0, Piece{King, White}},
		{5, 0, Piece{Bishop, White}},
		{6, 0, Piece{Knight, White}},
		{7, 0, Piece{Rook, White}},
		{0, 7, Piece{Rook, Black}},
		{1, 7, Piece{Knight, Black}},
		{2, 7, Piece{Bishop, Black}},
		{3, 7, Piece{Queen, Black}},
		{4, 7, Piece{King, Black}},
		{5, 7, Piece{Bishop, Black}},
		{6, 7, Piece{Knight, Black}},
		{7, 7, Piece{Rook, Black}},
	}
	for _, tt := range tests {
		p, ok := b.GetPiece(tt.x, tt.y)
		if !ok || p != tt.p {
			t.Errorf("GetPiece(%d,%d) = %v %v, want %v true", tt.x, tt.y, p, ok, tt.p)
		}
	}

	for col := 0; col < 8; col++ {
		wp, ok := b.GetPiece(col, 1)
		if !ok || wp != (Piece{Pawn, White}) {
			t.Errorf("white pawn at (%d,1): got %v %v", col, wp, ok)
		}
		bp, ok := b.GetPiece(col, 6)
		if !ok || bp != (Piece{Pawn, Black}) {
			t.Errorf("black pawn at (%d,6): got %v %v", col, bp, ok)
		}
	}

	for y := 2; y <= 5; y++ {
		for x := 0; x < 8; x++ {
			_, ok := b.GetPiece(x, y)
			if ok {
				t.Errorf("expected empty at (%d,%d)", x, y)
			}
		}
	}
}

func TestNewBitBoardCastlingRights(t *testing.T) {
	b := NewBitBoard()
	if !b.CanCastleKingside(White) || !b.CanCastleQueenside(White) {
		t.Error("white should have both castling rights")
	}
	if !b.CanCastleKingside(Black) || !b.CanCastleQueenside(Black) {
		t.Error("black should have both castling rights")
	}
}

func TestNewBlankBitBoard(t *testing.T) {
	b := NewBlankBitBoard()
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if _, ok := b.GetPiece(x, y); ok {
				t.Errorf("blank board has piece at (%d,%d)", x, y)
			}
		}
	}
	if b.CanCastleKingside(White) || b.CanCastleQueenside(White) ||
		b.CanCastleKingside(Black) || b.CanCastleQueenside(Black) {
		t.Error("blank board should have no castling rights")
	}
}

func TestSetDelGetPiece(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Queen, White}, 3, 4)
	p, ok := b.GetPiece(3, 4)
	if !ok || p != (Piece{Queen, White}) {
		t.Errorf("after SetPiece: got %v %v", p, ok)
	}
	b.DelPiece(3, 4)
	_, ok = b.GetPiece(3, 4)
	if ok {
		t.Error("after DelPiece: piece still present")
	}
}

func TestInitialMoveCount(t *testing.T) {
	b := NewBitBoard()
	moves := b.GetMoves(White)
	if len(moves) != 20 {
		t.Errorf("white opening moves = %d, want 20", len(moves))
	}
}

func TestKnightMovesCenter(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Knight, White}, 3, 3)
	b.SetPiece(Piece{King, White}, 0, 0)
	moves := b.GetMoves(White)
	knightMoves := 0
	for _, m := range moves {
		if m.SrcX == 3 && m.SrcY == 3 {
			knightMoves++
		}
	}
	if knightMoves != 8 {
		t.Errorf("knight on d4 should have 8 moves, got %d", knightMoves)
	}
}

func TestPawnPromotion(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 0, 6)
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	moves := b.GetMoves(White)
	promoCount := 0
	for _, m := range moves {
		if m.Promote != NoPiece {
			promoCount++
		}
	}
	if promoCount != 4 {
		t.Errorf("expected 4 promotion moves, got %d", promoCount)
	}
}

func TestPawnPromotionCapture(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 3, 6)
	b.SetPiece(Piece{Rook, Black}, 2, 7)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)
	moves := b.GetMoves(White)
	promoCount := 0
	for _, m := range moves {
		if m.Promote != NoPiece {
			promoCount++
		}
	}
	// 4 from pushing forward + 4 from capturing left
	if promoCount != 8 {
		t.Errorf("expected 8 promotion moves (push + capture), got %d", promoCount)
	}
}

func TestEnPassantAvailable(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 4, 4)
	b.SetPiece(Piece{Pawn, Black}, 3, 4)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)
	b.SetPawnDoubleStep(Black, 3)

	moves := b.GetMoves(White)
	found := false
	for _, m := range moves {
		if m.SrcX == 4 && m.SrcY == 4 && m.DstX == 3 && m.DstY == 5 {
			found = true
		}
	}
	if !found {
		t.Error("en passant should be available")
	}
}

func TestEnPassantNotAvailableWithoutFlag(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 4, 4)
	b.SetPiece(Piece{Pawn, Black}, 3, 4)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.SrcX == 4 && m.SrcY == 4 && m.DstX == 3 && m.DstY == 5 {
			t.Error("en passant should NOT be available without double-step flag")
		}
	}
}

func TestEnPassantClearsAfterMove(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 4, 4)
	b.SetPiece(Piece{Pawn, Black}, 3, 4)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)
	b.SetPawnDoubleStep(Black, 3)

	// Make a different white move (king move); this should clear double-step flags.
	after := b.ApplyMove(Mv(0, 0, 1, 0))

	if after.PawnDoubleStep(Black, 3) {
		t.Error("pawn double-step flag should be cleared after a move")
	}
}

func TestEnPassantCapture(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 4, 4)
	b.SetPiece(Piece{Pawn, Black}, 3, 4)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)
	b.SetPawnDoubleStep(Black, 3)

	after := b.ApplyMove(Mv(4, 4, 3, 5))

	if _, ok := after.GetPiece(3, 4); ok {
		t.Error("en passant should remove the captured pawn")
	}
	p, ok := after.GetPiece(3, 5)
	if !ok || p != (Piece{Pawn, White}) {
		t.Errorf("white pawn should be at (3,5), got %v %v", p, ok)
	}
}

func TestCastleKingside(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	b.SetCastlingRights(White, true, false)

	moves := b.GetMoves(White)
	found := false
	for _, m := range moves {
		if m.Kind == KingsideCastle {
			found = true
		}
	}
	if !found {
		t.Error("kingside castle should be available")
	}

	after := b.ApplyMove(Castle(KingsideCastle, White))
	kp, ok := after.GetPiece(6, 0)
	if !ok || kp != (Piece{King, White}) {
		t.Error("king should be at g1 after kingside castle")
	}
	rp, ok := after.GetPiece(5, 0)
	if !ok || rp != (Piece{Rook, White}) {
		t.Error("rook should be at f1 after kingside castle")
	}
}

func TestCastleQueenside(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	b.SetCastlingRights(White, false, true)

	moves := b.GetMoves(White)
	found := false
	for _, m := range moves {
		if m.Kind == QueensideCastle {
			found = true
		}
	}
	if !found {
		t.Error("queenside castle should be available")
	}

	after := b.ApplyMove(Castle(QueensideCastle, White))
	kp, ok := after.GetPiece(2, 0)
	if !ok || kp != (Piece{King, White}) {
		t.Error("king should be at c1 after queenside castle")
	}
	rp, ok := after.GetPiece(3, 0)
	if !ok || rp != (Piece{Rook, White}) {
		t.Error("rook should be at d1 after queenside castle")
	}
}

func TestCastleBlockedByPiece(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{Bishop, White}, 5, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	b.SetCastlingRights(White, true, false)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.Kind == KingsideCastle {
			t.Error("kingside castle should be blocked by bishop on f1")
		}
	}
}

func TestCastleThroughCheck(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{Rook, Black}, 5, 7)
	b.SetPiece(Piece{King, Black}, 0, 7)
	b.SetCastlingRights(White, true, false)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.Kind == KingsideCastle {
			t.Error("cannot castle through check (f1 attacked)")
		}
	}
}

func TestCastleWhileInCheck(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{Rook, Black}, 4, 7)
	b.SetPiece(Piece{King, Black}, 0, 7)
	b.SetCastlingRights(White, true, false)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.Kind == KingsideCastle {
			t.Error("cannot castle while in check")
		}
	}
}

func TestCastleRightsRevokedOnKingMove(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 0, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	b.SetCastlingRights(White, true, true)

	after := b.ApplyMove(Mv(4, 0, 5, 0))
	if after.CanCastleKingside(White) || after.CanCastleQueenside(White) {
		t.Error("castling rights should be revoked after king moves")
	}
}

func TestCastleRightsRevokedOnRookMove(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, White}, 7, 0)
	b.SetPiece(Piece{King, Black}, 4, 7)
	b.SetCastlingRights(White, true, false)

	after := b.ApplyMove(Mv(7, 0, 7, 3))
	if after.CanCastleKingside(White) {
		t.Error("kingside castle right should be revoked after h1 rook moves")
	}
}

func TestIsInCheck(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, Black}, 4, 7)
	b.SetPiece(Piece{King, Black}, 0, 7)

	if !b.IsInCheck(White) {
		t.Error("white king should be in check from black rook on e8")
	}
	if b.IsInCheck(Black) {
		t.Error("black king should not be in check")
	}
}

func TestIsCheckmated(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{Rook, Black}, 0, 7)
	b.SetPiece(Piece{Rook, Black}, 1, 6)
	b.SetPiece(Piece{King, Black}, 7, 7)

	if !b.IsCheckmated(White) {
		t.Error("white should be checkmated")
	}
}

func TestNotCheckmatedCanEscape(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{Rook, Black}, 0, 7)
	b.SetPiece(Piece{King, Black}, 7, 7)

	if b.IsCheckmated(White) {
		t.Error("white should not be checkmated (king can move to b1 or b2)")
	}
}

func TestPinnedPieceCannotMove(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Bishop, White}, 4, 1)
	b.SetPiece(Piece{Rook, Black}, 4, 7)
	b.SetPiece(Piece{King, Black}, 0, 7)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.SrcX == 4 && m.SrcY == 1 {
			after := b.ApplyMove(m)
			if after.IsInCheck(White) {
				t.Errorf("pinned bishop move %s leaves king in check", m)
			}
		}
	}

	// The bishop on e2 is absolutely pinned and should have no legal moves
	// except along the file.
	for _, m := range moves {
		if m.SrcX == 4 && m.SrcY == 1 && m.DstX != 4 {
			t.Errorf("pinned bishop should not move off file: %s", m)
		}
	}
}

func TestRookMoveGeneration(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Rook, White}, 3, 3)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)

	moves := b.GetMoves(White)
	rookMoves := 0
	for _, m := range moves {
		if m.SrcX == 3 && m.SrcY == 3 {
			rookMoves++
		}
	}
	if rookMoves != 14 {
		t.Errorf("rook on d4 should have 14 moves, got %d", rookMoves)
	}
}

func TestBishopMoveGeneration(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Bishop, White}, 3, 3)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)

	moves := b.GetMoves(White)
	bishopMoves := 0
	for _, m := range moves {
		if m.SrcX == 3 && m.SrcY == 3 {
			bishopMoves++
		}
	}
	// d4 bishop: 4 diags, but a1 blocked by own king = 12
	if bishopMoves != 12 {
		t.Errorf("bishop on d4 should have 12 moves, got %d", bishopMoves)
	}
}

func TestMoveString(t *testing.T) {
	tests := []struct {
		m    Move
		want string
	}{
		{Mv(0, 1, 0, 3), "a2 a4"},
		{Mv(4, 6, 4, 7), "e7 e8"},
		{MvPromo(0, 6, 0, 7, Queen), "a7 a8Q"},
		{MvPromo(0, 6, 0, 7, Knight), "a7 a8N"},
		{Castle(KingsideCastle, White), "0-0"},
		{Castle(QueensideCastle, Black), "0-0-0"},
	}
	for _, tt := range tests {
		if got := tt.m.String(); got != tt.want {
			t.Errorf("Move.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestPawnDoubleStepFromStart(t *testing.T) {
	b := NewBitBoard()
	moves := b.GetMoves(White)
	doubleSteps := 0
	for _, m := range moves {
		if m.SrcY == 1 && m.DstY == 3 {
			doubleSteps++
		}
	}
	if doubleSteps != 8 {
		t.Errorf("expected 8 pawn double-step moves, got %d", doubleSteps)
	}
}

func TestPawnCannotJumpOverPiece(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, White}, 3, 1)
	b.SetPiece(Piece{Knight, Black}, 3, 2)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.SrcX == 3 && m.SrcY == 1 {
			t.Errorf("pawn blocked by piece should have no moves, got %s", m)
		}
	}
}

func TestBlackEnPassant(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{Pawn, Black}, 4, 3)
	b.SetPiece(Piece{Pawn, White}, 5, 3)
	b.SetPiece(Piece{King, White}, 0, 0)
	b.SetPiece(Piece{King, Black}, 7, 7)
	b.SetPawnDoubleStep(White, 5)

	moves := b.GetMoves(Black)
	found := false
	for _, m := range moves {
		if m.SrcX == 4 && m.SrcY == 3 && m.DstX == 5 && m.DstY == 2 {
			found = true
		}
	}
	if !found {
		t.Error("black en passant should be available")
	}

	after := b.ApplyMove(Mv(4, 3, 5, 2))
	if _, ok := after.GetPiece(5, 3); ok {
		t.Error("en passant should remove the captured pawn")
	}
}

func TestKingCannotMoveIntoCheck(t *testing.T) {
	b := NewBlankBitBoard()
	b.SetPiece(Piece{King, White}, 4, 0)
	b.SetPiece(Piece{Rook, Black}, 5, 7)
	b.SetPiece(Piece{King, Black}, 0, 7)

	moves := b.GetMoves(White)
	for _, m := range moves {
		if m.DstX == 5 {
			t.Errorf("king should not be able to move to f-file (attacked by rook): %s", m)
		}
	}
}

func TestApplyMoveReturnsNewBoard(t *testing.T) {
	b := NewBitBoard()
	after := b.ApplyMove(Mv(4, 1, 4, 3))

	// Original board unchanged.
	p, ok := b.GetPiece(4, 1)
	if !ok || p != (Piece{Pawn, White}) {
		t.Error("original board should be unchanged")
	}

	// New board has the move applied.
	_, ok = after.GetPiece(4, 1)
	if ok {
		t.Error("pawn should have moved from e2")
	}
	p, ok = after.GetPiece(4, 3)
	if !ok || p != (Piece{Pawn, White}) {
		t.Error("pawn should be at e4")
	}
}

func TestMakeMoveRejectsIllegal(t *testing.T) {
	b := NewBitBoard()
	_, err := b.MakeMove(Mv(4, 0, 4, 5))
	if err == nil {
		t.Error("moving king to e6 on opening should be illegal")
	}
}

func TestMakeMoveAcceptsLegal(t *testing.T) {
	b := NewBitBoard()
	after, err := b.MakeMove(Mv(4, 1, 4, 3))
	if err != nil {
		t.Errorf("e2e4 should be legal: %v", err)
	}
	p, ok := after.GetPiece(4, 3)
	if !ok || p != (Piece{Pawn, White}) {
		t.Error("pawn should be at e4 after e2e4")
	}
}

func TestScholarsMate(t *testing.T) {
	b := NewBitBoard()
	moves := []Move{
		Mv(4, 1, 4, 3), // e4
		Mv(4, 6, 4, 4), // e5
		Mv(5, 0, 2, 3), // Bc4
		Mv(1, 7, 2, 5), // Nc6
		Mv(3, 0, 7, 4), // Qh5
		Mv(6, 7, 5, 5), // Nf6
		Mv(7, 4, 5, 6), // Qxf7#
	}
	for _, m := range moves {
		var err error
		b, err = b.MakeMove(m)
		if err != nil {
			t.Fatalf("move %s should be legal: %v", m, err)
		}
	}
	if !b.IsCheckmated(Black) {
		t.Error("black should be checkmated (Scholar's Mate)")
	}
}
