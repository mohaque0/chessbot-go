package board

import "fmt"

type Bits uint64

func NewEmpty() Bits {
	return 0
}

func NewSingleBit(idx uint) Bits {
	return 1 << idx
}

func (b Bits) Invert() Bits {
	return ^(b)
}

func (b Bits) Get(x uint8, y uint8) int {
	if b.Occupied(x, y) {
		return 1
	}
	return 0
}

func (b *Bits) Set(x uint8, y uint8) {
	*b = Union(*b, NewSingleBit(coord_to_idx(x, y)))
}

func (b *Bits) Unset(x uint8, y uint8) {
	*b = Intersect(*b, NewSingleBit(coord_to_idx(x, y)).Invert())
}

func (b Bits) Occupied(x uint8, y uint8) bool {
	return NewSingleBit(coord_to_idx(x, y))&b != 0
}
func (b Bits) String() string {
	return fmt.Sprintf("%064b", uint64(b))
}

func Union(others ...Bits) Bits {
	var ret = NewEmpty()
	for _, o := range others {
		ret |= o
	}
	return ret
}

func Intersect(others ...Bits) Bits {
	var ret = NewEmpty()
	for _, o := range others {
		ret &= o
	}
	return ret
}

func coord_to_idx(x uint8, y uint8) uint {
	return uint(x) + uint(y)*8
}
