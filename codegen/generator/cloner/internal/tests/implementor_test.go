package tests

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
	
	undIntCmp := cmp.Comparer(func(a, b und.Und[int]) bool {
		return und.Equal(a, b)
	})
	undStringCmp := cmp.Comparer(func(a, b und.Und[string]) bool {
		return und.Equal(a, b)
	})
	wowCmp := cmp.Comparer(func(a, b implementor.Wow) bool {
		return und.Equal(und.Und[string](a), und.Und[string](b))
	})
	
	assert.DeepEqual(
		t,
		ci,
		ci.CloneFunc(func(i int) int { return i }), 
		undIntCmp,
		undStringCmp,
		wowCmp,
	)
	assert.Equal(
		t,
		30,
		ci.CloneFunc(func(i int) int { return i * 2 }).U.Value(),
	)
}
