package board

type RayMove int

const (
	NW RayMove = iota
	N
	NE
	E
	SE
	S
	SW
	W
)

var RookMoves = [4]RayMove{N, S, E, W}
var BishopMoves = [4]RayMove{NW, NE, SE, SW}
var QueenMoves = [8]RayMove{N, S, E, W, NW, NE, SE, SW}

var ShiftColumnsMask [9]uint64

// This is a mask: given a point (the index of the bit representing (x,y))
// all the bits in a given direction from (x,y) are set to 1. The rest are 0.
//
// Note:
//
//	That coordinates are in the order of Bits indexes.
//
//	That coordinates are (column, row) and columns start indexing from right to left.
//	  Cols: 0b76543210_76543210_76543210_76543210_76543210_76543210_76543210_76543210
//
//	The rows are listed in reverse order top to bottom:
//	  Rows: 0brow7...._row6...._row5...._ ... _row0....
//
//	In the bit representation below the directions are:
//	     N
//	  E <^> W
//	     S
//
//	The order of the directions follows: RayMove
var RayMask [8][64]uint64
var KingMoves [64]uint64
var KnightMoves [64]uint64
var BlackPawnStartMoves [64]uint64
var BlackPawnMoves [64]uint64
var BlackPawnAttacks [64]uint64
var WhitePawnStartMoves [64]uint64
var WhitePawnMoves [64]uint64
var WhitePawnAttacks [64]uint64

func init() {
	initShiftColumnsMask()
	initRayMask()
	initKingMoves()
	initKnightMoves()
	initPawnTables()
}

func sqBit(col, row int) uint64 {
	return 1 << uint(row*8+col)
}

func onBoard(col, row int) bool {
	return col >= 0 && col < 8 && row >= 0 && row < 8
}

func initShiftColumnsMask() {
	for shift := 0; shift < 9; shift++ {
		var mask uint64
		for row := 0; row < 8; row++ {
			for col := 0; col < 8-shift; col++ {
				mask |= sqBit(col, row)
			}
		}
		ShiftColumnsMask[shift] = mask
	}
}

func initRayMask() {
	deltas := [8][2]int{
		{-1, 1},  // NW
		{0, 1},   // N
		{1, 1},   // NE
		{1, 0},   // E
		{1, -1},  // SE
		{0, -1},  // S
		{-1, -1}, // SW
		{-1, 0},  // W
	}

	for dir := 0; dir < 8; dir++ {
		dc, dr := deltas[dir][0], deltas[dir][1]
		for sq := 0; sq < 64; sq++ {
			col, row := sq%8, sq/8
			var mask uint64
			c, r := col+dc, row+dr
			for onBoard(c, r) {
				mask |= sqBit(c, r)
				c += dc
				r += dr
			}
			RayMask[dir][sq] = mask
		}
	}
}

func initKingMoves() {
	for sq := 0; sq < 64; sq++ {
		col, row := sq%8, sq/8
		var mask uint64
		for dc := -1; dc <= 1; dc++ {
			for dr := -1; dr <= 1; dr++ {
				if dc == 0 && dr == 0 {
					continue
				}
				c, r := col+dc, row+dr
				if onBoard(c, r) {
					mask |= sqBit(c, r)
				}
			}
		}
		KingMoves[sq] = mask
	}
}

func initKnightMoves() {
	jumps := [8][2]int{
		{-2, 1}, {-1, 2}, {1, 2}, {2, 1},
		{2, -1}, {1, -2}, {-1, -2}, {-2, -1},
	}
	for sq := 0; sq < 64; sq++ {
		col, row := sq%8, sq/8
		var mask uint64
		for _, j := range jumps {
			c, r := col+j[0], row+j[1]
			if onBoard(c, r) {
				mask |= sqBit(c, r)
			}
		}
		KnightMoves[sq] = mask
	}
}

func initPawnTables() {
	for sq := 0; sq < 64; sq++ {
		col, row := sq%8, sq/8

		// White pawn moves (north = +row)
		var wMove, wStart, wAtk uint64
		if onBoard(col, row+1) {
			wMove |= sqBit(col, row+1)
			wStart = wMove
			if onBoard(col, row+2) {
				wStart |= sqBit(col, row+2)
			}
		}
		if onBoard(col-1, row+1) {
			wAtk |= sqBit(col-1, row+1)
		}
		if onBoard(col+1, row+1) {
			wAtk |= sqBit(col+1, row+1)
		}
		WhitePawnMoves[sq] = wMove
		WhitePawnStartMoves[sq] = wStart
		WhitePawnAttacks[sq] = wAtk

		// Black pawn moves (south = -row)
		var bMove, bStart, bAtk uint64
		if onBoard(col, row-1) {
			bMove |= sqBit(col, row-1)
			bStart = bMove
			if onBoard(col, row-2) {
				bStart |= sqBit(col, row-2)
			}
		}
		if onBoard(col-1, row-1) {
			bAtk |= sqBit(col-1, row-1)
		}
		if onBoard(col+1, row-1) {
			bAtk |= sqBit(col+1, row-1)
		}
		BlackPawnMoves[sq] = bMove
		BlackPawnStartMoves[sq] = bStart
		BlackPawnAttacks[sq] = bAtk
	}
}
