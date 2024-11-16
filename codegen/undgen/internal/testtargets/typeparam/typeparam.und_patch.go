// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen patch --help
package typeparam

import (
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//undgen:generated
type WithTypeParamPatch[T any] struct {
	Foo sliceund.Und[string]                        `json:",omitempty"`
	Bar sliceund.Und[T]                             `json:",omitempty"`
	Baz sliceund.Und[T]                             `json:",omitempty" und:"required"`
	Qux sliceund.Und[map[string]elastic.Elastic[T]] `json:"qux,omitempty" und:"len==2,values:nonnull"`
}

//undgen:generated
func (p *WithTypeParamPatch[T]) FromValue(v WithTypeParam[T]) {
	//nolint
	*p = WithTypeParamPatch[T]{
		Foo: sliceund.Defined(v.Foo),
		Bar: sliceund.Defined(v.Bar),
		Baz: option.MapOr(v.Baz, sliceund.Null[T](), sliceund.Defined[T]),
		Qux: sliceund.Defined(v.Qux),
	}
}

//undgen:generated
func (p WithTypeParamPatch[T]) ToValue() WithTypeParam[T] {
	//nolint
	return WithTypeParam[T]{
		Foo: p.Foo.Value(),
		Bar: p.Bar.Value(),
		Baz: option.Flatten(p.Baz.Unwrap()),
		Qux: p.Qux.Value(),
	}
}

//undgen:generated
func (p WithTypeParamPatch[T]) Merge(r WithTypeParamPatch[T]) WithTypeParamPatch[T] {
	//nolint
	return WithTypeParamPatch[T]{
		Foo: sliceund.FromOption(r.Foo.Unwrap().Or(p.Foo.Unwrap())),
		Bar: sliceund.FromOption(r.Bar.Unwrap().Or(p.Bar.Unwrap())),
		Baz: sliceund.FromOption(r.Baz.Unwrap().Or(p.Baz.Unwrap())),
		Qux: sliceund.FromOption(r.Qux.Unwrap().Or(p.Qux.Unwrap())),
	}
}

//undgen:generated
func (p WithTypeParamPatch[T]) ApplyPatch(v WithTypeParam[T]) WithTypeParam[T] {
	var orgP WithTypeParamPatch[T]
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
