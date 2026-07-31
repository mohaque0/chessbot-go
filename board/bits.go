package board

import (
	"fmt"
	"math/bits"
)

type Bits uint64

func NewBits(v uint64) Bits { return Bits(v) }

func CoordToIdx(x, y int) int { return y*8 + x }

func IdxToCoord(idx int) (int, int) { return idx % 8, idx / 8 }

func SingleBit(x, y int) Bits { return 1 << uint(y*8+x) }

func SingleBitIdx(idx int) Bits { return 1 << uint(idx) }

func (b Bits) IsEmpty() bool { return b == 0 }

func (b Bits) Get(x, y int) bool { return b&SingleBit(x, y) != 0 }

func (b *Bits) Set(x, y int) { *b |= SingleBit(x, y) }

func (b *Bits) Unset(x, y int) { *b &^= SingleBit(x, y) }

func (b *Bits) SetIdx(idx int) { *b |= SingleBitIdx(idx) }

func (b Bits) LSB() int { return bits.TrailingZeros64(uint64(b)) }

func (b Bits) MSB() int { return 63 - bits.LeadingZeros64(uint64(b)) }

func (b Bits) ClearLSB() Bits { return b & (b - 1) }

func ShiftUp(b Bits, n int) Bits { return b << uint(8*n) }

func ShiftDown(b Bits, n int) Bits { return b >> uint(8*n) }

func (b Bits) String() string { return fmt.Sprintf("0x%016X", uint64(b)) }
