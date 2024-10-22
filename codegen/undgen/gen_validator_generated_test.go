package undgen

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/undgen/testdata/validatortarget"
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"github.com/ngicks/und/validate"
	"gotest.tools/v3/assert"
)

// tests for generated code.

func Test_validator_All(t *testing.T) {
	tgt := validatortarget.All{
		Foo:    "foo",
		Bar:    option.None[string](),
		Baz:    option.Some("baz"),
		Qux:    und.Defined("qux"),
		Quux:   elastic.Null[string](),
		Corge:  sliceund.Null[string](),
		Grault: sliceelastic.Undefined[string](),
	}

	assert.NilError(t, tgt.UndValidate())

	for _, mod := range []func(a validatortarget.All) validatortarget.All{
		func(a validatortarget.All) validatortarget.All { a.Foo = ""; return a },
		func(a validatortarget.All) validatortarget.All { a.Bar = option.Some("bar"); return a },
		func(a validatortarget.All) validatortarget.All { a.Qux = und.Undefined[string](); return a },
		func(a validatortarget.All) validatortarget.All {
			a.Quux = elastic.FromValues("foo", "bar", "baz")
			return a
		},
		func(a validatortarget.All) validatortarget.All {
			a.Quux = elastic.FromPointers(ptr("foo"), nil, ptr("baz"))
			return a
		},
		func(a validatortarget.All) validatortarget.All { a.Corge = sliceund.Undefined[string](); return a },
		func(a validatortarget.All) validatortarget.All {
			a.Grault = sliceelastic.FromValues("foo", "bar")
			return a
		},
	} {
		modified := mod(tgt)
		assert.NilError(t, modified.UndValidate())
	}

	for _, mod := range []func(a validatortarget.All) (validatortarget.All, string){
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Qux = und.Null[string]()
			return a, "Qux"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Quux = elastic.Undefined[string]()
			return a, "Quux"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Quux = elastic.FromValues("foo", "bar")
			return a, "Quux"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Quux = elastic.FromValues("foo", "bar", "baz", "qux")
			return a, "Quux"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Corge = sliceund.Defined("corge")
			return a, "Corge"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Grault = sliceelastic.Null[string]()
			return a, "Grault"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Grault = sliceelastic.FromValue("grault")
			return a, "Grault"
		},
		func(a validatortarget.All) (validatortarget.All, string) {
			a.Grault = sliceelastic.FromPointers(ptr("foo"), nil, ptr("baz"))
			return a, "Grault"
		},
	} {
		modified, chain := mod(tgt)
		err := modified.UndValidate()
		assert.Assert(t, err != nil)
		vErr := err.(*validate.ValidationError)
		t.Logf("%v", vErr)
		assert.ErrorContains(t, vErr, "."+chain+":")
	}
}

func Test_validator_MapSliceArray(t *testing.T) {
	tgt := validatortarget.MapSliceArray{
		Foo: map[string]option.Option[string]{
			"foo": option.Some("foo"),
		},
		Bar: []und.Und[string]{und.Null[string](), und.Null[string]()},
		Baz: [5]elastic.Elastic[string]{elastic.FromValues("foo", "bar")},
	}

	assert.NilError(t, tgt.UndValidate())

	for _, mod := range []func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray{
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Foo = map[string]option.Option[string]{
				"foo": option.Some("foo"),
				"bar": option.Some("bar"),
			}
			return a
		},
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Foo = nil
			return a
		},
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Foo = map[string]option.Option[string]{}
			return a
		},
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Bar = nil
			return a
		},
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Bar = []und.Und[string]{}
			return a
		},
		func(a validatortarget.MapSliceArray) validatortarget.MapSliceArray {
			a.Baz = [5]elastic.Elastic[string]{elastic.FromValues("foo", "bar", "baz", "qux", "quux")}
			return a
		},
	} {
		modified := mod(tgt)
		assert.NilError(t, modified.UndValidate())
	}

	for _, mod := range []func(a validatortarget.MapSliceArray) (validatortarget.MapSliceArray, string){
		func(a validatortarget.MapSliceArray) (validatortarget.MapSliceArray, string) {
			a.Foo = map[string]option.Option[string]{
				"foo": option.Some("foo"),
				"bar": option.Some("bar"),
				"baz": option.None[string](),
			}
			return a, "foo[baz]"
		},
		func(a validatortarget.MapSliceArray) (validatortarget.MapSliceArray, string) {
			a.Bar = []und.Und[string]{und.Defined("foo")}
			return a, "bar[0]"
		},
		func(a validatortarget.MapSliceArray) (validatortarget.MapSliceArray, string) {
			a.Baz = [5]elastic.Elastic[string]{elastic.Undefined[string](), elastic.Undefined[string](), elastic.FromValue("1")}
			return a, "baz[2]"
		},
	} {
		modified, chain := mod(tgt)
		err := modified.UndValidate()
		assert.Assert(t, err != nil)
		vErr := err.(*validate.ValidationError)
		t.Logf("%v", vErr)
		assert.ErrorContains(t, vErr, "."+chain+":")
	}
}

func Test_validator_ContainsImplementor(t *testing.T) {
	tgt := validatortarget.ContainsImplementor{
		I: validatortarget.Implementor{
			Foo: "foo",
		},
		O: option.Some(validatortarget.Implementor{Foo: "bar"}),
	}

	assert.NilError(t, tgt.UndValidate())

	for _, mod := range []func(a validatortarget.ContainsImplementor) (validatortarget.ContainsImplementor, string){
		func(a validatortarget.ContainsImplementor) (validatortarget.ContainsImplementor, string) {
			a.I = validatortarget.Implementor{}
			return a, "I"
		},
		func(a validatortarget.ContainsImplementor) (validatortarget.ContainsImplementor, string) {
			a.O = option.None[validatortarget.Implementor]()
			return a, "O"
		},
		func(a validatortarget.ContainsImplementor) (validatortarget.ContainsImplementor, string) {
			a.O = option.Some(validatortarget.Implementor{})
			return a, "O"
		},
	} {
		modified, chain := mod(tgt)
		err := modified.UndValidate()
		assert.Assert(t, err != nil)
		vErr := err.(*validate.ValidationError)
		t.Logf("%v", vErr)
		assert.ErrorContains(t, vErr, "."+chain+":")
	}
}

func Test_validator_MapSliceArrayContainsImplementor(t *testing.T) {
	assert.NilError(t, validatortarget.MapSliceArrayContainsImplementor{}.UndValidate())
	tgt := validatortarget.MapSliceArrayContainsImplementor{
		Foo: map[string]option.Option[validatortarget.Implementor]{
			"foo": option.Some(validatortarget.Implementor{Foo: "yay"}),
		},
		Bar: []und.Und[validatortarget.Implementor]{und.Null[validatortarget.Implementor](), und.Null[validatortarget.Implementor]()},
		Baz: [5]elastic.Elastic[validatortarget.Implementor]{elastic.FromValues(validatortarget.Implementor{Foo: "bar"}, validatortarget.Implementor{Foo: "baz"})},
	}
	assert.NilError(t, tgt.UndValidate())
	for _, mod := range []func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string){
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Foo = map[string]option.Option[validatortarget.Implementor]{
				"foo": option.Some(validatortarget.Implementor{Foo: "yay"}),
				"bar": option.None[validatortarget.Implementor](),
			}
			return a, "Foo[bar]"
		},
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Foo = map[string]option.Option[validatortarget.Implementor]{
				"foo": option.Some(validatortarget.Implementor{Foo: "yay"}),
				"baz": option.Some(validatortarget.Implementor{}),
			}
			return a, "Foo[baz]"
		},
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Bar = []und.Und[validatortarget.Implementor]{und.Null[validatortarget.Implementor](), und.Undefined[validatortarget.Implementor]()}
			return a, "Bar[1]"
		},
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Bar = []und.Und[validatortarget.Implementor]{und.Null[validatortarget.Implementor](), und.Null[validatortarget.Implementor](), und.Defined(validatortarget.Implementor{})}
			return a, "Bar[2]"
		},
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Baz = [5]elastic.Elastic[validatortarget.Implementor]{}
			a.Baz[4] = elastic.FromValue(validatortarget.Implementor{Foo: "aa"})
			return a, "Baz[4]"
		},
		func(a validatortarget.MapSliceArrayContainsImplementor) (validatortarget.MapSliceArrayContainsImplementor, string) {
			a.Baz = [5]elastic.Elastic[validatortarget.Implementor]{}
			a.Baz[4] = elastic.FromValues(validatortarget.Implementor{Foo: "aa"}, validatortarget.Implementor{})
			return a, "Baz[4]"
		},
	} {
		modified, chain := mod(tgt)
		err := modified.UndValidate()
		assert.Assert(t, err != nil)
		vErr := err.(*validate.ValidationError)
		t.Logf("%v", vErr)
		assert.ErrorContains(t, vErr, "."+chain+":")
	}
}

func Test_validator_A(t *testing.T) {
	assert.NilError(t, validatortarget.A{}.UndValidate())
	assert.NilError(
		t,
		validatortarget.A{
			elastic.FromValue(
				validatortarget.Implementor{
					Foo: "foo",
				},
			),
		}.UndValidate(),
	)
	assert.Assert(
		t,
		validatortarget.A{
			elastic.FromValue(
				validatortarget.Implementor{},
			),
		}.UndValidate() != nil,
	)
}

func Test_validator_B(t *testing.T) {
	assert.NilError(t, validatortarget.B{}.UndValidate())
	assert.NilError(
		t,
		validatortarget.B{
			"foo": sliceelastic.FromValue(
				validatortarget.Implementor{
					Foo: "foo",
				},
			),
		}.UndValidate(),
	)
	assert.Assert(
		t,
		validatortarget.B{
			"bar": sliceelastic.FromValue(
				validatortarget.Implementor{},
			),
		}.UndValidate() != nil,
	)
}

var validAll = validatortarget.All{
	Foo:    "foo",
	Bar:    option.None[string](),
	Baz:    option.Some("baz"),
	Qux:    und.Defined("qux"),
	Quux:   elastic.Null[string](),
	Corge:  sliceund.Null[string](),
	Grault: sliceelastic.Undefined[string](),
}

func Test_validator_C(t *testing.T) {
	tgt := validatortarget.C{}
	assert.NilError(t, tgt.UndValidate())
	someAll := option.Some(validAll)
	tgt = validatortarget.C{someAll, someAll, someAll}
	assert.NilError(t, tgt.UndValidate())
	tgt[1] = option.Some(validatortarget.All{})
	assert.Assert(
		t,
		tgt.UndValidate() != nil,
	)
}

func Test_validator_D(t *testing.T) {
	tgt := validatortarget.D{Foo: validAll, Bar: option.Some(validAll)}
	assert.NilError(t, tgt.UndValidate())
	tgt.Bar = option.Some(validatortarget.All{})
	assert.Assert(t, tgt.UndValidate() != nil)
}
