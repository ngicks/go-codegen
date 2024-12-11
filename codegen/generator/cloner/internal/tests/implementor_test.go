package tests

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ngicks/go-codegen/codegen/generator/cloner/internal/testtargets/implementor"
	"github.com/ngicks/und"
	"gotest.tools/v3/assert"
)

func TestImplementor(t *testing.T) {
	ci := implementor.ContainsImplementor[int]{
		U:  und.Defined(15),
		US: und.Defined("foo"),
		W:  implementor.Wow(und.Defined("bar")),
	}
	assert.DeepEqual(
		t,
		ci,
		ci.CloneFunc(func(i int) int { return i }), cmpopts.EquateComparable(implementor.Wow{}),
	)
	assert.Equal(
		t,
		30,
		ci.CloneFunc(func(i int) int { return i * 2 }).U.Value(),
	)
}
