package imports

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

var (
	cfg = &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes,
	}
	testdataPkgs []*packages.Package
)

func init() {
	var err error
	testdataPkgs, err = packages.Load(cfg, "./internal/...")
	if err != nil {
		panic(err)
	}
}

var undExtra = []TargetImport{
	{
		Import: Import{"github.com/ngicks/und/option", "option"},
		Types:  []string{"Option"},
	},
	{
		Import: Import{"github.com/ngicks/und", "und"},
		Types:  []string{"Und"},
	},
	{
		Import: Import{"github.com/ngicks/und/elastic", "elastic"},
		Types:  []string{"Elastic"},
	},
	{
		Import: Import{"github.com/ngicks/und/sliceund", "sliceund"},
		Types:  []string{"Und"},
	},
	{
		Import: Import{"github.com/ngicks/und/sliceund/elastic", "elastic"},
		Ident:  "sliceelastic",
		Types:  []string{"Elastic"},
	},
	{
		Import: Import{"github.com/ngicks/und/conversion", "conversion"},
		Types:  []string{"Empty"},
	},
}

func TestImports(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}
	p := NewParserPackages(testdataPkgs)
	p.AppendExtra(undExtra...)
	im, err := p.Parse(pkg1.Syntax[0].Imports)
	assert.NilError(t, err)

	assert.DeepEqual(
		t,
		map[string]TargetImport{
			"conversion": {
				Import: Import{Path: "github.com/ngicks/und/conversion", Name: "conversion"},
				Types:  []string{"Empty"},
			},
			"elastic": {
				Import: Import{Path: "github.com/ngicks/und/elastic", Name: "elastic"},
				Types:  []string{"Elastic"},
			},
			"option": {
				Import: Import{Path: "github.com/ngicks/und/option", Name: "option"},
				Types:  []string{"Option"},
			},
			"sliceelastic": {
				Import: Import{Path: "github.com/ngicks/und/sliceund/elastic", Name: "elastic"},
				Ident:  "sliceelastic",
				Types:  []string{"Elastic"},
			},
			"sliceund": {
				Import: Import{Path: "github.com/ngicks/und/sliceund", Name: "sliceund"},
				Types:  []string{"Und"},
			},
			"und": {
				Import: Import{Path: "github.com/ngicks/und", Name: "und"},
				Types:  []string{"Und"},
			},
		},
		im.ident,
	)
	assert.DeepEqual(
		t,
		map[string]TargetImport{
			"conversion": {
				Import: Import{"github.com/ngicks/und/conversion", "conversion"},
				Ident:  "",
				Types:  []string{"Empty"},
			},
			"option": {
				Import: Import{"github.com/ngicks/und/option", "option"},
				Ident:  "",
				Types:  []string{"Option"},
			},
			"sliceelastic": {
				Import: Import{"github.com/ngicks/und/sliceund/elastic", "elastic"},
				Ident:  "sliceelastic",
				Types:  []string{"Elastic"},
			},
			"sliceund": {
				Import: Import{"github.com/ngicks/und/sliceund", "sliceund"},
				Ident:  "",
				Types:  []string{"Und"},
			},
		},
		im.missing,
	)
	assert.DeepEqual(
		t,
		map[string]TargetImport{
			"github.com/ngicks/und": {
				Import: Import{Path: "github.com/ngicks/und", Name: "und"},
				Types:  []string{"Und"},
			},
			"github.com/ngicks/und/conversion": {
				Import: Import{Path: "github.com/ngicks/und/conversion", Name: "conversion"},
				Types:  []string{"Empty"},
			},
			"github.com/ngicks/und/elastic": {
				Import: Import{Path: "github.com/ngicks/und/elastic", Name: "elastic"},
				Types:  []string{"Elastic"},
			},
			"github.com/ngicks/und/option": {
				Import: Import{Path: "github.com/ngicks/und/option", Name: "option"},
				Types:  []string{"Option"},
			},
			"github.com/ngicks/und/sliceund": {
				Import: Import{Path: "github.com/ngicks/und/sliceund", Name: "sliceund"},
				Types:  []string{"Und"},
			},
			"github.com/ngicks/und/sliceund/elastic": {
				Import: Import{Path: "github.com/ngicks/und/sliceund/elastic", Name: "elastic"},
				Ident:  "sliceelastic",
				Types:  []string{"Elastic"},
			},
		},
		im.extra,
	)
	assert.DeepEqual(
		t,
		map[string]TargetImport{
			"github.com/ngicks/go-codegen/codegen/imports/internal/pkg1": {
				Import: Import{
					Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg1",
					Name: "pkg1",
				},
				Types: []string{"Pkg1"},
			},
			"github.com/ngicks/go-codegen/codegen/imports/internal/pkg2": {
				Import: Import{
					Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg2",
					Name: "pkg2",
				},
				Types: []string{"Pkg2"},
			},
			"github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2": {
				Import: Import{
					Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2",
					Name: "pkg22",
				},
				Types: []string{"Pkg2_2"},
			},
			"github.com/ngicks/go-codegen/codegen/imports/internal/pkg3": {
				Import: Import{
					Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg3",
					Name: "pkg3diff",
				},
				Types: []string{"Pkg3"},
			},
		},
		im.dependencies,
	)

	ident, ti, ok := im.getIdent("github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2", "Pkg2_2")
	assert.Assert(t, ok)
	assert.Equal(t, "pkg22", ident)
	assert.DeepEqual(
		t,
		TargetImport{
			Import: Import{
				Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2",
				Name: "pkg22",
			},
			Types: []string{"Pkg2_2"},
		},
		ti,
	)
	assert.DeepEqual(
		t,
		map[string]TargetImport{
			"conversion": {
				Import: Import{"github.com/ngicks/und/conversion", "conversion"},
				Ident:  "",
				Types:  []string{"Empty"},
			},
			"option": {
				Import: Import{"github.com/ngicks/und/option", "option"},
				Ident:  "",
				Types:  []string{"Option"},
			},
			"pkg22": {
				Import: Import{
					Path: "github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2",
					Name: "pkg22",
				},
				Types: []string{"Pkg2_2"},
			},
			"sliceelastic": {
				Import: Import{"github.com/ngicks/und/sliceund/elastic", "elastic"},
				Ident:  "sliceelastic",
				Types:  []string{"Elastic"},
			},
			"sliceund": {
				Import: Import{"github.com/ngicks/und/sliceund", "sliceund"},
				Ident:  "",
				Types:  []string{"Und"},
			},
		},
		im.missing,
	)

	dec := decorator.NewDecorator(pkg1.Fset)
	df, err := dec.DecorateFile(pkg1.Syntax[0])
	assert.NilError(t, err)
	im.AddMissingImports(df)
	assertAddMissingImports := func() {
		t.Helper()
		assert.DeepEqual(
			t,
			[]*dst.ImportSpec{
				{
					Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und"`},
					Decs: dst.ImportSpecDecorations{NodeDecs: dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}},
				},
				{
					Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/elastic"`},
					Decs: dst.ImportSpecDecorations{NodeDecs: dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}},
				},
				{
					Name: dst.NewIdent("pkg22"),
					Path: &dst.BasicLit{
						Kind:  token.STRING,
						Value: `"github.com/ngicks/go-codegen/codegen/imports/internal/pkg2/pkg2-2"`,
					},
				},
				{
					Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/conversion"`},
				},
				{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/option"`}},
				{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/sliceund"`}},
				{
					Name: dst.NewIdent("sliceelastic"),
					Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/sliceund/elastic"`},
				},
			},
			df.Imports,
		)
	}
	assertAddMissingImports()
	im.AddMissingImports(df)
	assertAddMissingImports()
}
