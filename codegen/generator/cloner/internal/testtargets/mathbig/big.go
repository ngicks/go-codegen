package mathbig

import "math/big"

type Big struct {
	Int   *big.Int
	Float *big.Float
	Rat   *big.Rat
}
