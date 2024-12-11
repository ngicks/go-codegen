package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/generator/cloner/internal/testtargets/simple"
	"gotest.tools/v3/assert"
)

func ptr[T any](t T) *T {
	return &t
}

func TestSimple(t *testing.T) {
	b := simple.B{
		A: []*[]string{
			ptr([]string{"foo"}),
			ptr([]string{"bar", "baz"}),
		},
		B: map[string]int{
			"foo": 15,
			"bar": 24,
		},
		C: []*map[int][3]string{
			ptr(map[int][3]string{
				5: {},
			}),
			ptr(map[int][3]string{
				78: {},
			}),
			ptr(map[int][3]string{
				845: {},
			}),
		},
	}

	assert.DeepEqual(t, b, b.Clone())

	cloned := b.Clone()
	(*cloned.A[0])[0] = "foofoo"

	assert.Assert(t, (*b.A[0])[0] != (*cloned.A[0])[0])
}
