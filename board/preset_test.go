package board

import "testing"

func TestShiftColumnsMask(t *testing.T) {
	expected := [9]uint64{
		0xFFFFFFFFFFFFFFFF, // shift 0: all bits
		0x7F7F7F7F7F7F7F7F, // shift 1
		0x3F3F3F3F3F3F3F3F, // shift 2
		0x1F1F1F1F1F1F1F1F, // shift 3
		0x0F0F0F0F0F0F0F0F, // shift 4
		0x0707070707070707, // shift 5
		0x0303030303030303, // shift 6
		0x0101010101010101, // shift 7
		0x0000000000000000, // shift 8
	}
	for i, want := range expected {
		if ShiftColumnsMask[i] != want {
			t.Errorf("ShiftColumnsMask[%d] = 0x%016X, want 0x%016X", i, ShiftColumnsMask[i], want)
		}
	}
}

func TestRayMaskNW(t *testing.T) {
	// NW 1 (1,0): only (0,1) set = bit 8
	if RayMask[NW][1] != 1<<8 {
		t.Errorf("RayMask[NW][1] = 0x%X, want 0x%X", RayMask[NW][1], uint64(1<<8))
	}
	// NW 7 (7,0): diagonal (6,1),(5,2),(4,3),(3,4),(2,5),(1,6),(0,7)
	var want uint64
	for i := 1; i <= 7; i++ {
		want |= sqBit(7-i, i)
	}
	if RayMask[NW][7] != want {
		t.Errorf("RayMask[NW][7] = 0x%X, want 0x%X", RayMask[NW][7], want)
	}
}

func TestRayMaskN(t *testing.T) {
	// N 0 (0,0): bits at (0,1)...(0,7) = bits 8,16,24,32,40,48,56
	var want uint64
	for r := 1; r < 8; r++ {
		want |= sqBit(0, r)
	}
	if RayMask[N][0] != want {
		t.Errorf("RayMask[N][0] = 0x%X, want 0x%X", RayMask[N][0], want)
	}
	// N 56 (0,7): no squares north
	if RayMask[N][56] != 0 {
		t.Errorf("RayMask[N][56] = 0x%X, want 0", RayMask[N][56])
	}
}

func TestRayMaskE(t *testing.T) {
	// E 0 (0,0): bits 1-7 = 0xFE
	if RayMask[E][0] != 0xFE {
		t.Errorf("RayMask[E][0] = 0x%X, want 0xFE", RayMask[E][0])
	}
	// E 7 (7,0): no squares east
	if RayMask[E][7] != 0 {
		t.Errorf("RayMask[E][7] = 0x%X, want 0", RayMask[E][7])
	}
}

func TestKingMoves(t *testing.T) {
	// King at 0 (0,0): can go (1,0), (0,1), (1,1) = bits 1, 8, 9
	want := sqBit(1, 0) | sqBit(0, 1) | sqBit(1, 1)
	if KingMoves[0] != want {
		t.Errorf("KingMoves[0] = 0x%X, want 0x%X", KingMoves[0], want)
	}
}

func TestKnightMoves(t *testing.T) {
	// Knight at 0 (0,0): can go (2,1), (1,2) = bits 10, 17
	want := sqBit(2, 1) | sqBit(1, 2)
	if KnightMoves[0] != want {
		t.Errorf("KnightMoves[0] = 0x%X, want 0x%X", KnightMoves[0], want)
	}
}

func TestWhitePawnMoves(t *testing.T) {
	// White pawn at 0 (0,0): moves to (0,1) = bit 8
	if WhitePawnMoves[0] != sqBit(0, 1) {
		t.Errorf("WhitePawnMoves[0] = 0x%X, want 0x%X", WhitePawnMoves[0], sqBit(0, 1))
	}
	// White pawn start at 0 (0,0): moves to (0,1) and (0,2)
	want := sqBit(0, 1) | sqBit(0, 2)
	if WhitePawnStartMoves[0] != want {
		t.Errorf("WhitePawnStartMoves[0] = 0x%X, want 0x%X", WhitePawnStartMoves[0], want)
	}
	// White pawn at 56 (0,7): no moves
	if WhitePawnMoves[56] != 0 {
		t.Errorf("WhitePawnMoves[56] = 0x%X, want 0", WhitePawnMoves[56])
	}
}

func TestWhitePawnAttacks(t *testing.T) {
	// White pawn at 0 (0,0): attacks (1,1) only (no col -1)
	if WhitePawnAttacks[0] != sqBit(1, 1) {
		t.Errorf("WhitePawnAttacks[0] = 0x%X, want 0x%X", WhitePawnAttacks[0], sqBit(1, 1))
	}
	// White pawn at 1 (1,0): attacks (0,1) and (2,1)
	want := sqBit(0, 1) | sqBit(2, 1)
	if WhitePawnAttacks[1] != want {
		t.Errorf("WhitePawnAttacks[1] = 0x%X, want 0x%X", WhitePawnAttacks[1], want)
	}
}

func TestBlackPawnMoves(t *testing.T) {
	// Black pawn at 8 (0,1): moves to (0,0) = bit 0
	if BlackPawnMoves[8] != sqBit(0, 0) {
		t.Errorf("BlackPawnMoves[8] = 0x%X, want 0x%X", BlackPawnMoves[8], sqBit(0, 0))
	}
	// Black pawn at 0 (0,0): no moves south
	if BlackPawnMoves[0] != 0 {
		t.Errorf("BlackPawnMoves[0] = 0x%X, want 0", BlackPawnMoves[0])
	}
}

func TestBlackPawnAttacks(t *testing.T) {
	// Black pawn at 8 (0,1): attacks (1,0) only
	if BlackPawnAttacks[8] != sqBit(1, 0) {
		t.Errorf("BlackPawnAttacks[8] = 0x%X, want 0x%X", BlackPawnAttacks[8], sqBit(1, 0))
	}
	// Black pawn at 9 (1,1): attacks (0,0) and (2,0)
	want := sqBit(0, 0) | sqBit(2, 0)
	if BlackPawnAttacks[9] != want {
		t.Errorf("BlackPawnAttacks[9] = 0x%X, want 0x%X", BlackPawnAttacks[9], want)
	}
}

func TestRayMoveIdx(t *testing.T) {
	moves := []RayMove{NW, N, NE, E, SE, S, SW, W}
	for i, m := range moves {
		if int(m) != i {
			t.Errorf("RayMove %d has value %d", i, int(m))
		}
	}
}

func TestExactRustValues(t *testing.T) {
	// Spot-check exact values from the Rust source

	// NW 7 (7,0) from Rust: 0b00000001_00000010_00000100_00001000_00010000_00100000_01000000_00000000
	rustNW7 := uint64(0x0102040810204000)
	if RayMask[NW][7] != rustNW7 {
		t.Errorf("RayMask[NW][7] = 0x%016X, want 0x%016X", RayMask[NW][7], rustNW7)
	}

	// NE 0 (0,0) from Rust: 0b10000000_01000000_00100000_00010000_00001000_00000100_00000010_00000000
	rustNE0 := uint64(0x8040201008040200)
	if RayMask[NE][0] != rustNE0 {
		t.Errorf("RayMask[NE][0] = 0x%016X, want 0x%016X", RayMask[NE][0], rustNE0)
	}

	// King at 9 (1,1) from Rust: 0b00000000_..._00000111_00000101_00000111
	rustKing9 := uint64(0x0000000000070507)
	if KingMoves[9] != rustKing9 {
		t.Errorf("KingMoves[9] = 0x%016X, want 0x%016X", KingMoves[9], rustKing9)
	}

	// Knight at 0 (0,0) from Rust: 0b00000000_..._00000010_00000100_00000000
	rustKnight0 := uint64(0x0000000000020400)
	if KnightMoves[0] != rustKnight0 {
		t.Errorf("KnightMoves[0] = 0x%016X, want 0x%016X", KnightMoves[0], rustKnight0)
	}

	// White pawn attacks at 7 (7,0): attacks (6,1) only = 0b..._01000000_00000000
	rustWPA7 := uint64(0x0000000000004000)
	if WhitePawnAttacks[7] != rustWPA7 {
		t.Errorf("WhitePawnAttacks[7] = 0x%016X, want 0x%016X", WhitePawnAttacks[7], rustWPA7)
	}
}
