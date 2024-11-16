package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/undgen/internal/testtargets/all"
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	"gotest.tools/v3/assert"
)

func Test_all_UndValidate(t *testing.T) {
	allValid := all.All{
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
		UndNullish:               und.Null[string](),
		UndDef:                   und.Defined("UndDef"),
		UndNull:                  und.Null[string](),
		UndUnd:                   und.Undefined[string](),
		UndDefOrUnd:              und.Defined("UndDefOrUnd"),
		UndDefOrNull:             und.Defined("UndDefOrNull"),
		UndNullOrUnd:             und.Null[string](),
		UndDefOrNullOrUnd:        und.Defined("UndDefOrNullOrUnd"),
		ElaRequired:              elastic.FromOptions(option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired")),
		ElaNullish:               elastic.Null[string](),
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
		ElaEqEqNonNullSlice:      elastic.FromOptions(option.Some("ElaEqEqNonNullSlice"), option.Some("ElaEqEqNonNullSlice"), option.Some("ElaEqEqNonNullSlice")),
		ElaEqEqNonNullNullSlice:  elastic.Null[string](),
		ElaEqEqNonNullSingle:     elastic.FromOptions(option.Some("ElaEqEqNonNullSingle")),
		ElaEqEqNonNullNullSingle: elastic.FromOptions(option.Some("ElaEqEqNonNullNullSingle")),
		ElaEqEqNonNull:           elastic.FromOptions(option.Some("ElaEqEqNonNull"), option.Some("ElaEqEqNonNull"), option.Some("ElaEqEqNonNull")),
		ElaEqEqNonNullNull:       elastic.FromOptions(option.Some("ElaEqEqNonNullNull"), option.Some("ElaEqEqNonNullNull"), option.Some("ElaEqEqNonNullNull")),
	}

	assert.NilError(t, allValid.UndValidate())

	type testCase struct {
		name          string
		patch         func(a all.All) all.All
		errorContains string
	}

	for _, tc := range []testCase{
		{
			name: "defined",
			patch: func(a all.All) all.All {
				a.OptRequired = option.None[string]()
				return a
			},
			errorContains: "opt_required", // name in json tag is used.
		},
		{
			name: "null",
			patch: func(a all.All) all.All {
				a.UndNull = und.Defined("aaa")
				return a
			},
			errorContains: "UndNull",
		},
		{
			name: "undefined",
			patch: func(a all.All) all.All {
				a.UndDefOrUnd = und.Null[string]()
				return a
			},
			errorContains: "UndDefOrUnd",
		},
		{
			name: "lenGr",
			patch: func(a all.All) all.All {
				a.ElaGr = elastic.FromValues[string]()
				return a
			},
			errorContains: "ElaGr",
		},
		{
			name: "lenLt",
			patch: func(a all.All) all.All {
				a.ElaLeEq = elastic.FromOptions(append(a.ElaLeEq.Unwrap().Value(), option.None[string]())...)
				return a
			},
			errorContains: "ElaLeEq",
		},
		{
			name: "lenEq",
			patch: func(a all.All) all.All {
				a.ElaEqEquRequired = elastic.FromOptions(option.Some("ElaEqEquRequired"), option.None[string](), option.None[string]())
				return a
			},
			errorContains: "ElaEqEquRequired",
		},
		{
			name: "nonnull",
			patch: func(a all.All) all.All {
				a.ElaEqEqNonNullSlice = elastic.FromOptions(append(a.ElaEqEqNonNullSlice.Unwrap().Value(), option.None[string]())...)
				return a
			},
			errorContains: "ElaEqEqNonNullSlice",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patched := tc.patch(allValid)
			assert.ErrorContains(t, patched.UndValidate(), "validation failed at ."+tc.errorContains+":")
		})
	}

}

func Test_all_UndPlain(t *testing.T) {
	allSample := all.All{
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

	p := allSample.UndPlain()

	assertDeepEqualAllPlain := func(t *testing.T, i, j all.AllPlain) {
		t.Helper()
		assert.DeepEqual(
			t,
			i, j,
			compareOptionStringSlice, compareOptionOptionStringSlice, compareUndStringSlice,
		)
	}

	assertDeepEqualAllPlain(
		t,
		all.AllPlain{
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
			UndNullish:               option.None[conversion.Empty](),
			UndDef:                   "UndDef",
			UndNull:                  nil,
			UndUnd:                   nil,
			UndDefOrUnd:              option.Some("UndDefOrUnd"),
			UndDefOrNull:             option.Some("UndDefOrNull"),
			UndNullOrUnd:             option.None[conversion.Empty](),
			UndDefOrNullOrUnd:        und.Defined("UndDefOrNullOrUnd"),
			ElaRequired:              []option.Option[string]{option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired"), option.None[string](), option.Some("ElaRequired")},
			ElaNullish:               option.None[conversion.Empty](),
			ElaDef:                   []option.Option[string]{option.Some("ElaDef"), option.None[string](), option.Some("ElaDef"), option.None[string](), option.Some("ElaDef")},
			ElaNull:                  nil,
			ElaUnd:                   nil,
			ElaDefOrUnd:              option.Some([]option.Option[string]{option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd"), option.None[string](), option.Some("ElaDefOrUnd")}),
			ElaDefOrNull:             option.Some([]option.Option[string]{option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull"), option.None[string](), option.Some("ElaDefOrNull")}),
			ElaNullOrUnd:             option.None[conversion.Empty](),
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
		all.All{
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

	a := all.All{
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
