package undgen

import (
	"bytes"
	"slices"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/undgen/internal/plaintarget"
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes"
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"gotest.tools/v3/assert"
)

var (
	compareOptionStringSlice = gocmp.Comparer(func(i, j []option.Option[string]) bool {
		return option.Options[string](i).Equal(option.Options[string](j))
	})
	compareOptionOptionStringSlice = gocmp.Comparer(func(i, j option.Option[[]option.Option[string]]) bool {
		return i.EqualFunc(j, func(i, j []option.Option[string]) bool {
			return option.Options[string](i).Equal(option.Options[string](j))
		})
	})
	compareUndStringSlice = gocmp.Comparer(func(i, j und.Und[[]string]) bool {
		return i.EqualFunc(j, func(i, j []string) bool { return slices.Equal(i, j) })
	})
)

// tests for generated code.

func Test_plain_All(t *testing.T) {
	a := targettypes.All{
		Foo:                      "foo",
		Bar:                      ptr("bar"),
		Baz:                      nil,
		Qux:                      []string{"foo", "bar", "baz"},
		UntouchedOpt:             option.Some(5),
		UntouchedUnd:             und.Defined(25),
		UntouchedSliceUnd:        sliceund.Defined(44),
		OptRequired:              option.Some("OptRequired"),
		OptNullish:               option.Some("OptNullish"),
		OptDef:                   option.Some("OptDef"),
		OptNull:                  option.Some("OptNull"),
		OptUnd:                   option.Some("OptUnd"),
		OptDefOrUnd:              option.Some("OptDefOrUnd"),
		OptDefOrNull:             option.Some("OptDefOrNull"),
		OptNullOrUnd:             option.Some("OptNullOrUnd"),
		OptDefOrNullOrUnd:        option.Some("OptDefOrNullOrUnd"),
		UndRequired:              und.Defined("UndRequired"),
		UndNullish:               und.Defined("UndNullish"),
		UndDef:                   und.Defined("UndDef"),
		UndNull:                  und.Defined("UndNull"),
		UndUnd:                   und.Defined("UndUnd"),
		UndDefOrUnd:              und.Defined("UndDefOrUnd"),
		UndDefOrNull:             und.Defined("UndDefOrNull"),
		UndNullOrUnd:             und.Defined("UndNullOrUnd"),
		UndDefOrNullOrUnd:        und.Defined("UndDefOrNullOrUnd"),
		ElaRequired:              elastic.FromOptions(option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired")),
		ElaNullish:               elastic.FromOptions(option.Some("ElaNullish"), option.None[string](), option.Some("ElaNullish"), option.None[string](), option.Some("ElaNullish")),
		ElaDef:                   elastic.FromOptions(option.Some("ElaDef"), option.None[string](), option.Some("ElaDef"), option.None[string](), option.Some("ElaDef")),
		ElaNull:                  elastic.FromOptions(option.Some("ElaNull"), option.None[string](), option.Some("ElaNull"), option.None[string](), option.Some("ElaNull")),
		ElaUnd:                   elastic.FromOptions(option.Some("ElaUnd"), option.None[string](), option.Some("ElaUnd"), option.None[string](), option.Some("ElaUnd")),
		ElaDefOrUnd:              elastic.FromOptions(option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd")),
		ElaDefOrNull:             elastic.FromOptions(option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull")),
		ElaNullOrUnd:             elastic.FromOptions(option.Some("ElaNullOrUnd"), option.None[string](), option.Some("ElaNullOrUnd"), option.None[string](), option.Some("ElaNullOrUnd")),
		ElaDefOrNullOrUnd:        elastic.FromOptions(option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd")),
		ElaEqEq:                  elastic.FromOptions(option.Some("ElaEqEq"), option.None[string](), option.Some("ElaEqEq"), option.None[string](), option.Some("ElaEqEq")),
		ElaGr:                    elastic.FromOptions(option.Some("ElaGr"), option.None[string](), option.Some("ElaGr"), option.None[string](), option.Some("ElaGr")),
		ElaGrEq:                  elastic.FromOptions(option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq")),
		ElaLe:                    elastic.FromOptions(option.Some("ElaLe"), option.None[string](), option.Some("ElaLe"), option.None[string](), option.Some("ElaLe")),
		ElaLeEq:                  elastic.FromOptions(option.Some("ElaLeEq"), option.None[string](), option.Some("ElaLeEq"), option.None[string](), option.Some("ElaLeEq")),
		ElaEqEquRequired:         elastic.FromOptions(option.Some("ElaEqEquRequired"), option.None[string](), option.Some("ElaEqEquRequired"), option.None[string](), option.Some("ElaEqEquRequired")),
		ElaEqEquNullish:          elastic.FromOptions(option.Some("ElaEqEquNullish"), option.None[string](), option.Some("ElaEqEquNullish"), option.None[string](), option.Some("ElaEqEquNullish")),
		ElaEqEquDef:              elastic.FromOptions(option.Some("ElaEqEquDef"), option.None[string](), option.Some("ElaEqEquDef"), option.None[string](), option.Some("ElaEqEquDef")),
		ElaEqEquNull:             elastic.FromOptions(option.Some("ElaEqEquNull"), option.None[string](), option.Some("ElaEqEquNull"), option.None[string](), option.Some("ElaEqEquNull")),
		ElaEqEquUnd:              elastic.FromOptions(option.Some("ElaEqEquUnd"), option.None[string](), option.Some("ElaEqEquUnd"), option.None[string](), option.Some("ElaEqEquUnd")),
		ElaEqEqNonNullSlice:      elastic.FromOptions(option.Some("ElaEqEqNonNullSlice"), option.None[string](), option.Some("ElaEqEqNonNullSlice"), option.None[string](), option.Some("ElaEqEqNonNullSlice")),
		ElaEqEqNonNullNullSlice:  elastic.FromOptions(option.Some("ElaEqEqNonNullNullSlice"), option.None[string](), option.Some("ElaEqEqNonNullNullSlice"), option.None[string](), option.Some("ElaEqEqNonNullNullSlice")),
		ElaEqEqNonNullSingle:     elastic.FromOptions(option.Some("ElaEqEqNonNullSingle"), option.None[string](), option.Some("ElaEqEqNonNullSingle"), option.None[string](), option.Some("ElaEqEqNonNullSingle")),
		ElaEqEqNonNullNullSingle: elastic.FromOptions(option.Some("ElaEqEqNonNullNullSingle"), option.None[string](), option.Some("ElaEqEqNonNullNullSingle"), option.None[string](), option.Some("ElaEqEqNonNullNullSingle")),
		ElaEqEqNonNull:           elastic.FromOptions(option.Some("ElaEqEqNonNull"), option.None[string](), option.Some("ElaEqEqNonNull"), option.None[string](), option.Some("ElaEqEqNonNull")),
		ElaEqEqNonNullNull:       elastic.FromOptions(option.Some("ElaEqEqNonNullNull"), option.None[string](), option.Some("ElaEqEqNonNullNull"), option.None[string](), option.Some("ElaEqEqNonNullNull")),
	}

	p := a.UndPlain()

	assertDeepEqualAllPlain := func(t *testing.T, i, j targettypes.AllPlain) {
		t.Helper()
		assert.DeepEqual(
			t,
			i, j,
			compareOptionStringSlice, compareOptionOptionStringSlice, compareUndStringSlice,
		)
	}

	assertDeepEqualAllPlain(
		t,
		targettypes.AllPlain{
			Foo:                      "foo",
			Bar:                      ptr("bar"),
			Baz:                      nil,
			Qux:                      []string{"foo", "bar", "baz"},
			UntouchedOpt:             option.Some(5),
			UntouchedUnd:             und.Defined(25),
			UntouchedSliceUnd:        sliceund.Defined(44),
			OptRequired:              "OptRequired",
			OptNullish:               nil,
			OptDef:                   "OptDef",
			OptNull:                  nil,
			OptUnd:                   nil,
			OptDefOrUnd:              option.Some("OptDefOrUnd"),
			OptDefOrNull:             option.Some("OptDefOrNull"),
			OptNullOrUnd:             nil,
			OptDefOrNullOrUnd:        option.Some("OptDefOrNullOrUnd"),
			UndRequired:              "UndRequired",
			UndNullish:               option.None[*struct{}](),
			UndDef:                   "UndDef",
			UndNull:                  nil,
			UndUnd:                   nil,
			UndDefOrUnd:              option.Some("UndDefOrUnd"),
			UndDefOrNull:             option.Some("UndDefOrNull"),
			UndNullOrUnd:             option.None[*struct{}](),
			UndDefOrNullOrUnd:        und.Defined("UndDefOrNullOrUnd"),
			ElaRequired:              []option.Option[string]{option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired")},
			ElaNullish:               option.None[*struct{}](),
			ElaDef:                   []option.Option[string]{option.Some("ElaDef"), option.None[string](), option.Some("ElaDef"), option.None[string](), option.Some("ElaDef")},
			ElaNull:                  nil,
			ElaUnd:                   nil,
			ElaDefOrUnd:              option.Some([]option.Option[string]{option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd")}),
			ElaDefOrNull:             option.Some([]option.Option[string]{option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull")}),
			ElaNullOrUnd:             option.None[*struct{}](),
			ElaDefOrNullOrUnd:        elastic.FromOptions(option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd")),
			ElaEqEq:                  option.Some("ElaEqEq"),
			ElaGr:                    []option.Option[string]{option.Some("ElaGr"), option.None[string](), option.Some("ElaGr"), option.None[string](), option.Some("ElaGr")},
			ElaGrEq:                  []option.Option[string]{option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq")},
			ElaLe:                    []option.Option[string]{},
			ElaLeEq:                  []option.Option[string]{option.Some("ElaLeEq")},
			ElaEqEquRequired:         [2]option.Option[string]{option.Some("ElaEqEquRequired"), option.None[string]()},
			ElaEqEquNullish:          und.Defined([2]option.Option[string]{option.Some("ElaEqEquNullish"), option.None[string]()}),
			ElaEqEquDef:              [2]option.Option[string]{option.Some("ElaEqEquDef"), option.None[string]()},
			ElaEqEquNull:             option.Some([2]option.Option[string]{option.Some("ElaEqEquNull"), option.None[string]()}),
			ElaEqEquUnd:              option.Some([2]option.Option[string]{option.Some("ElaEqEquUnd"), option.None[string]()}),
			ElaEqEqNonNullSlice:      und.Defined([]string{"ElaEqEqNonNullSlice", "", "ElaEqEqNonNullSlice", "", "ElaEqEqNonNullSlice"}),
			ElaEqEqNonNullNullSlice:  nil,
			ElaEqEqNonNullSingle:     "ElaEqEqNonNullSingle",
			ElaEqEqNonNullNullSingle: option.Some("ElaEqEqNonNullNullSingle"),
			ElaEqEqNonNull:           [3]string{"ElaEqEqNonNull", "", "ElaEqEqNonNull"},
			ElaEqEqNonNullNull:       option.Some([3]string{"ElaEqEqNonNullNull", "", "ElaEqEqNonNullNull"}),
		},
		p,
	)

	assert.DeepEqual(
		t,
		targettypes.All{
			Foo:                      "foo",
			Bar:                      ptr("bar"),
			Baz:                      nil,
			Qux:                      []string{"foo", "bar", "baz"},
			UntouchedOpt:             option.Some(5),
			UntouchedUnd:             und.Defined(25),
			UntouchedSliceUnd:        sliceund.Defined(44),
			OptRequired:              option.Some("OptRequired"),
			OptNullish:               option.None[string](),
			OptDef:                   option.Some("OptDef"),
			OptNull:                  option.None[string](),
			OptUnd:                   option.None[string](),
			OptDefOrUnd:              option.Some("OptDefOrUnd"),
			OptDefOrNull:             option.Some("OptDefOrNull"),
			OptNullOrUnd:             option.None[string](),
			OptDefOrNullOrUnd:        option.Some("OptDefOrNullOrUnd"),
			UndRequired:              und.Defined("UndRequired"),
			UndNullish:               und.Undefined[string](),
			UndDef:                   und.Defined("UndDef"),
			UndNull:                  und.Null[string](),
			UndUnd:                   und.Undefined[string](),
			UndDefOrUnd:              und.Defined("UndDefOrUnd"),
			UndDefOrNull:             und.Defined("UndDefOrNull"),
			UndNullOrUnd:             und.Undefined[string](),
			UndDefOrNullOrUnd:        und.Defined("UndDefOrNullOrUnd"),
			ElaRequired:              elastic.FromOptions(option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired")),
			ElaNullish:               elastic.Undefined[string](),
			ElaDef:                   elastic.FromOptions(option.Some("ElaDef"), option.None[string](), option.Some("ElaDef"), option.None[string](), option.Some("ElaDef")),
			ElaNull:                  elastic.Null[string](),
			ElaUnd:                   elastic.Undefined[string](),
			ElaDefOrUnd:              elastic.FromOptions(option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd")),
			ElaDefOrNull:             elastic.FromOptions(option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull")),
			ElaNullOrUnd:             elastic.Undefined[string](),
			ElaDefOrNullOrUnd:        elastic.FromOptions(option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd"), option.None[string](), option.Some("ElaDefOrNullOrUnd")),
			ElaEqEq:                  elastic.FromOptions(option.Some("ElaEqEq")),
			ElaGr:                    elastic.FromOptions(option.Some("ElaGr"), option.None[string](), option.Some("ElaGr"), option.None[string](), option.Some("ElaGr")),
			ElaGrEq:                  elastic.FromOptions(option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq"), option.None[string](), option.Some("ElaGrEq")),
			ElaLe:                    elastic.FromOptions[string](),
			ElaLeEq:                  elastic.FromOptions(option.Some("ElaLeEq")),
			ElaEqEquRequired:         elastic.FromOptions(option.Some("ElaEqEquRequired"), option.None[string]()),
			ElaEqEquNullish:          elastic.FromOptions(option.Some("ElaEqEquNullish"), option.None[string]()),
			ElaEqEquDef:              elastic.FromOptions(option.Some("ElaEqEquDef"), option.None[string]()),
			ElaEqEquNull:             elastic.FromOptions(option.Some("ElaEqEquNull"), option.None[string]()),
			ElaEqEquUnd:              elastic.FromOptions(option.Some("ElaEqEquUnd"), option.None[string]()),
			ElaEqEqNonNullSlice:      elastic.FromOptions(option.Some("ElaEqEqNonNullSlice"), option.Some(""), option.Some("ElaEqEqNonNullSlice"), option.Some(""), option.Some("ElaEqEqNonNullSlice")),
			ElaEqEqNonNullNullSlice:  elastic.Null[string](),
			ElaEqEqNonNullSingle:     elastic.FromOptions(option.Some("ElaEqEqNonNullSingle")),
			ElaEqEqNonNullNullSingle: elastic.FromOptions(option.Some("ElaEqEqNonNullNullSingle")),
			ElaEqEqNonNull:           elastic.FromOptions(option.Some("ElaEqEqNonNull"), option.Some(""), option.Some("ElaEqEqNonNull")),
			ElaEqEqNonNullNull:       elastic.FromOptions(option.Some("ElaEqEqNonNullNull"), option.Some(""), option.Some("ElaEqEqNonNullNull")),
		},
		p.UndRaw(),
	)

	a = targettypes.All{
		ElaGr: elastic.FromValue("foo"),
	}
	assert.DeepEqual(
		t,
		[]option.Option[string]{option.Some("foo"), option.None[string]()},
		a.UndPlain().ElaGr,
		compareOptionStringSlice,
	)
	assert.DeepEqual(
		t,
		elastic.FromOptions(option.Some("foo"), option.None[string]()),
		a.UndPlain().UndRaw().ElaGr,
	)
}

func Test_plain_IncludesImplementor(t *testing.T) {
	now := time.Now()
	r := plaintarget.IncludesImplementor{
		Impl:         sub.Foo[time.Time]{T: now, Yay: "yay"},
		Opt:          option.Some(sub.Foo[time.Time]{T: now, Yay: "yay"}),
		Und:          und.Defined(sub.Foo[*bytes.Buffer]{T: nil, Yay: "yay"}),
		Elastic:      elastic.FromValue(sub.Foo[string]{T: "hm", Yay: "yay"}),
		SliceUnd:     sliceund.Defined(sub.Foo[int]{T: 15, Yay: "yay"}),
		SliceElastic: sliceelastic.FromValue(sub.Foo[bool]{T: true, Yay: "yaya"}),
	}

	p := r.UndPlain()

	assert.DeepEqual(
		t,
		plaintarget.IncludesImplementorPlain{
			Impl:         sub.FooPlain[time.Time]{T: now, Nay: "nay"},
			Opt:          sub.FooPlain[time.Time]{T: now, Nay: "nay"},
			Und:          sub.FooPlain[*bytes.Buffer]{T: nil, Nay: "nay"},
			Elastic:      []option.Option[sub.FooPlain[string]]{option.Some(sub.FooPlain[string]{T: "hm", Nay: "nay"})},
			SliceUnd:     sub.FooPlain[int]{T: 15, Nay: "nay"},
			SliceElastic: [2]option.Option[sub.FooPlain[bool]]{option.Some(sub.FooPlain[bool]{T: true, Nay: "yaya"})},
		},
		p,
	)

	assert.DeepEqual(
		t,
		plaintarget.IncludesImplementor{
			Impl:     sub.Foo[time.Time]{T: now, Yay: "yay"},
			Opt:      option.Some(sub.Foo[time.Time]{T: now, Yay: "yay"}),
			Und:      und.Defined(sub.Foo[*bytes.Buffer]{T: nil, Yay: "yay"}),
			Elastic:  elastic.FromValue(sub.Foo[string]{T: "hm", Yay: "yay"}),
			SliceUnd: sliceund.Defined(sub.Foo[int]{T: 15, Yay: "yay"}),
			SliceElastic: sliceelastic.FromOptions(
				option.Some(sub.Foo[bool]{T: true, Yay: "yaya"}),
				option.None[sub.Foo[bool]](),
			),
		},
		p.UndRaw(),
	)
}

func Test_plain_IncludesImplementorArraySliceMap(t *testing.T) {
	now := time.Now()
	r := plaintarget.IncludesImplementorArraySliceMap{
		A1: [3]sub.Foo[time.Time]{{T: now, Yay: "0"}, {T: now, Yay: "1"}, {T: now, Yay: "2"}},
		A2: [5]option.Option[sub.Foo[time.Time]]{
			option.Some(sub.Foo[time.Time]{T: now, Yay: "0"}),
			option.Some(sub.Foo[time.Time]{T: now, Yay: "1"}),
			option.None[sub.Foo[time.Time]](),
			option.Some(sub.Foo[time.Time]{T: now, Yay: "3"}),
			option.None[sub.Foo[time.Time]](),
		},
		S1: []und.Und[sub.Foo[*bytes.Buffer]]{
			und.Undefined[sub.Foo[*bytes.Buffer]](),
			und.Null[sub.Foo[*bytes.Buffer]](),
			und.Defined(sub.Foo[*bytes.Buffer]{T: nil, Yay: "0"}),
		},
		S2: []elastic.Elastic[sub.Foo[string]]{
			elastic.Undefined[sub.Foo[string]](),
			elastic.Null[sub.Foo[string]](),
			elastic.FromValue(sub.Foo[string]{T: "hm", Yay: "0"}),
		},
		M1: map[string]sliceund.Und[sub.Foo[int]]{
			"foo": sliceund.Defined(sub.Foo[int]{T: 15, Yay: "yay"}),
			"bar": sliceund.Null[sub.Foo[int]](),
		},
		M2: map[string]sliceelastic.Elastic[sub.Foo[bool]]{
			"foo": sliceelastic.FromValue(sub.Foo[bool]{T: true, Yay: "yay"}),
			"bar": sliceelastic.Null[sub.Foo[bool]](),
		},
	}

	p := r.UndPlain()

	assert.DeepEqual(
		t,
		plaintarget.IncludesImplementorArraySliceMapPlain{
			A1: [3]sub.FooPlain[time.Time]{{T: now, Nay: "0"}, {T: now, Nay: "1"}, {T: now, Nay: "2"}},
			A2: [5]sub.FooPlain[time.Time]{
				{T: now, Nay: "0"},
				{T: now, Nay: "1"},
				{},
				{T: now, Nay: "3"},
				{},
			},
			S1: []sub.FooPlain[*bytes.Buffer]{
				{},
				{},
				{T: nil, Nay: "0"},
			},
			S2: [][]option.Option[sub.FooPlain[string]]{
				nil,
				nil,
				{option.Some(sub.FooPlain[string]{T: "hm", Nay: "0"})},
			},
			M1: map[string]sub.FooPlain[int]{
				"foo": {T: 15, Nay: "nay"},
				"bar": {},
			},
			M2: map[string][2]option.Option[sub.FooPlain[bool]]{
				"foo": {option.Some(sub.FooPlain[bool]{T: true, Nay: "nay"})},
				"bar": {},
			},
		},
		p,
	)

	assert.DeepEqual(
		t,
		plaintarget.IncludesImplementorArraySliceMap{
			A1: [3]sub.Foo[time.Time]{{T: now, Yay: "0"}, {T: now, Yay: "1"}, {T: now, Yay: "2"}},
			A2: [5]option.Option[sub.Foo[time.Time]]{
				option.Some(sub.Foo[time.Time]{T: now, Yay: "0"}),
				option.Some(sub.Foo[time.Time]{T: now, Yay: "1"}),
				option.Some(sub.Foo[time.Time]{}),
				option.Some(sub.Foo[time.Time]{T: now, Yay: "3"}),
				option.Some(sub.Foo[time.Time]{}),
			},
			S1: []und.Und[sub.Foo[*bytes.Buffer]]{
				und.Defined(sub.Foo[*bytes.Buffer]{}),
				und.Defined(sub.Foo[*bytes.Buffer]{}),
				und.Defined(sub.Foo[*bytes.Buffer]{T: nil, Yay: "0"}),
			},
			S2: []elastic.Elastic[sub.Foo[string]]{
				elastic.FromValues[sub.Foo[string]](),
				elastic.FromValues[sub.Foo[string]](),
				elastic.FromValue(sub.Foo[string]{T: "hm", Yay: "0"}),
			},
			M1: map[string]sliceund.Und[sub.Foo[int]]{
				"foo": sliceund.Defined(sub.Foo[int]{T: 15, Yay: "yay"}),
				"bar": sliceund.Defined(sub.Foo[int]{}),
			},
			M2: map[string]sliceelastic.Elastic[sub.Foo[bool]]{
				"foo": sliceelastic.FromOptions(option.Some(sub.Foo[bool]{T: true, Yay: "yay"}), option.None[sub.Foo[bool]]()),
				"bar": sliceelastic.FromOptions(option.None[sub.Foo[bool]](), option.None[sub.Foo[bool]]()),
			},
		},
		p.UndRaw(),
	)
}
