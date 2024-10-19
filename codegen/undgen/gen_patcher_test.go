package undgen

import (
	"maps"
	"slices"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/undgen/testdata/patchtarget"
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func Test_generatePatcher(t *testing.T) {
	var pkg *packages.Package
	for _, p := range testdataPackages {
		if p.PkgPath == "github.com/ngicks/go-codegen/codegen/undgen/testdata/targettypes" {
			pkg = p
			break
		}
	}

	testPrinter := suffixwriter.NewTestWriter("yay")
	err := GeneratePatcher(
		testPrinter.Writer,
		pkg,
		ConstUnd.Imports,
		"All", "WithTypeParam", "A", "B", "IncludesSubTarget",
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}

func ptr[T any](t T) *T {
	return &t
}

func Test_patcher_ApplyPatch(t *testing.T) {
	all := patchtarget.All{
		Foo:          "foo",
		Bar:          ptr(6),
		Baz:          &struct{}{},
		Qux:          []string{"foo", "bar"},
		Opt:          option.Some("yay"),
		Und:          und.Defined("nay"),
		Elastic:      elastic.FromValue("wow"),
		SliceUnd:     sliceund.Defined("mah"),
		SliceElastic: sliceelastic.FromValue("hahahh"),
	}

	assert.DeepEqual(
		t,
		patchtarget.All{
			Foo:          "",
			Bar:          ptr(6),
			Baz:          &struct{}{},
			Qux:          []string{"foo", "bar"},
			Opt:          option.Some("yay"),
			Und:          und.Defined("nay"),
			Elastic:      elastic.FromValue("wow"),
			SliceUnd:     sliceund.Defined("mah"),
			SliceElastic: sliceelastic.FromValue("hahahh"),
		},
		patchtarget.AllPatch{Foo: sliceund.Null[string]()}.ApplyPatch(all),
	)
	assert.DeepEqual(
		t,
		patchtarget.All{
			Foo:          "foo",
			Bar:          ptr(6),
			Baz:          nil,
			Qux:          []string{"foo", "bar"},
			Opt:          option.Some("yay"),
			Und:          und.Defined("nay"),
			Elastic:      elastic.FromValue("wow"),
			SliceUnd:     sliceund.Defined("mah"),
			SliceElastic: sliceelastic.FromValue("hahahh"),
		},
		patchtarget.AllPatch{Baz: sliceund.Defined[*struct{}](nil)}.ApplyPatch(all),
	)
	assert.DeepEqual(
		t,
		patchtarget.All{
			Foo:          "foo",
			Bar:          ptr(6),
			Baz:          &struct{}{},
			Qux:          []string{"foo", "bar"},
			Opt:          option.None[string](),
			Und:          und.Defined("nay"),
			Elastic:      elastic.FromValues([]string{"foo", "bar"}),
			SliceUnd:     sliceund.Null[string](),
			SliceElastic: sliceelastic.Null[string](),
		},
		patchtarget.AllPatch{
			Opt:          sliceund.Null[string](),
			Elastic:      elastic.FromValues([]string{"foo", "bar"}),
			SliceUnd:     sliceund.Null[string](),
			SliceElastic: sliceelastic.Null[string](),
		}.ApplyPatch(all),
	)
}

func Test_patcher_Merge(t *testing.T) {
	var p patchtarget.AllPatch
	p.FromValue(patchtarget.All{
		Foo:          "foo",
		Bar:          ptr(6),
		Baz:          &struct{}{},
		Qux:          []string{"foo", "bar"},
		Opt:          option.Some("yay"),
		Und:          und.Defined("nay"),
		Elastic:      elastic.FromValue("wow"),
		SliceUnd:     sliceund.Defined("mah"),
		SliceElastic: sliceelastic.FromValue("hahahh"),
	})
	assert.DeepEqual(
		t,
		patchtarget.AllPatch{
			Foo:          sliceund.Defined("foo"),
			Bar:          sliceund.Defined(ptr(6)),
			Baz:          sliceund.Defined(&struct{}{}),
			Qux:          sliceund.Defined([]string{"foo", "bar"}),
			Opt:          sliceund.Defined("yay"),
			Und:          und.Defined("nay"),
			Elastic:      elastic.FromValue("wow"),
			SliceUnd:     sliceund.Defined("mah"),
			SliceElastic: sliceelastic.FromValue("hahahh"),
		},
		p,
		gocmp.Comparer(
			func(i, j sliceund.Und[[]string]) bool {
				return i.EqualFunc(j, slices.Equal[[]string])
			},
		),
		gocmp.Comparer(
			func(i, j sliceund.Und[*int]) bool {
				return i.EqualFunc(
					j,
					func(i, j *int) bool {
						if i == nil || j == nil {
							return i == nil && j == nil
						}
						return *i == *j
					},
				)
			},
		),
	)
}
