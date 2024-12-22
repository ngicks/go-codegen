package tests

import (
	"math/big"
	"testing"

	"github.com/ngicks/go-codegen/codegen/generator/cloner/internal/testtargets/mathbig"
	"gotest.tools/v3/assert"
)

func TestMathBig(t *testing.T) {
	b := mathbig.Big{
		Int:   big.NewInt(12),
		Float: big.NewFloat(6.809),
		Rat:   big.NewRat(5, 5),
	}

	assertNums := func(t *testing.T, v mathbig.Big, i int64, f1, f2 float64) {
		t.Helper()

		assert.Equal(t, i, v.Int.Int64())

		f, _ := v.Float.Float64()
		assert.Equal(t, f1, f)

		f, _ = v.Rat.Float64()
		assert.Equal(t, f2, f)
	}

	assertNums(t, b, 12, 6.809, 1)

	cloned := b.Clone()

	assertNums(t, cloned, 12, 6.809, 1)

	cloned.Int.Abs(big.NewInt(-14))
	cloned.Float.Abs(big.NewFloat(13.78))
	cloned.Rat.Abs(big.NewRat(12, 4))

	assertNums(t, cloned, 14, 13.78, 3)

	assertNums(t, b, 12, 6.809, 1)
}
