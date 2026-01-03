package importmap

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_searchNamedSpecsByPath(t *testing.T) {
	ns := []NamedSpec{
		{
			Ident: ".",
			Spec: Spec{
				Package: Package{
					Path: "foo",
					Name: "foo",
				},
			},
		},
		{
			Ident: "_",
			Spec: Spec{
				Package: Package{
					Path: "foo",
					Name: "foo",
				},
			},
		},
	}
	ns = cleanNamedSpec(ns)

	_, ok := searchNamedSpecsByPath(ns, "foo")
	assert.Assert(t, !ok)

	ns = append(
		ns,
		NamedSpec{
			Spec: Spec{
				Package: Package{
					Path: "foo",
					Name: "foo",
				},
			},
		},
	)
	ns = cleanNamedSpec(ns)

	t.Log(ns)

	found, ok := searchNamedSpecsByPath(ns, "foo")
	assert.Assert(t, ok)
	assert.DeepEqual(
		t,
		NamedSpec{
			Spec: Spec{
				Package: Package{
					Path: "foo",
					Name: "foo",
				},
			},
		},
		found,
	)
}
