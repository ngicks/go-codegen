package importmap

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/graft/internal/loader"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

var testdataPkgs = loader.LoadPackagesPanicking("./internal/...")

var undExtra = []NamedSpec{
	{
		Spec: Spec{Package: Package{Path: "github.com/ngicks/und/option", Name: "option"}},
	},
	{
		Spec: Spec{Package: Package{Path: "github.com/ngicks/und", Name: "und"}},
	},
	{
		Spec: Spec{Package: Package{Path: "github.com/ngicks/und/elastic"}},
	},
	{
		Spec: Spec{Package: Package{Path: "github.com/ngicks/und/sliceund", Name: "sliceund"}},
	},
	{
		Ident: "sliceelastic",
		Spec:  Spec{Package: Package{Path: "github.com/ngicks/und/sliceund/elastic", Name: "elastic"}},
	},
	{
		Spec: Spec{Package: Package{Path: "github.com/ngicks/und/conversion", Name: "conversion"}},
	},
}

func TestParser(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}

	var (
		m   *Map
		err error
	)
	t.Run("parse ast", func(t *testing.T) {
		p := NewParserFromPackages(testdataPkgs)
		p.AddSpec(undExtra...)
		m, err = p.ParseAst(pkg1.Syntax[0].Imports)
		assert.NilError(t, err)
	})

	t.Run("inspecting internal(ident)", func(t *testing.T) {
		existingIdents := make(map[string]string) // ident -> path
		for _, ns := range m.ident {
			name, _ := ns.Name()
			existingIdents[name] = ns.Spec.Package.Path
		}
		assert.DeepEqual(
			t,
			map[string]string{
				"und":     "github.com/ngicks/und",
				"elastic": "github.com/ngicks/und/elastic",
			},
			existingIdents,
		)
	})

	t.Run("Ident for import spec defined inside the file", func(t *testing.T) {
		ident, ok := m.Ident("github.com/ngicks/und")
		assert.Assert(t, ok)
		assert.Equal(t, "und", ident)
	})

	t.Run("Ident for dependency", func(t *testing.T) {
		missingSpecs := []NamedSpec{
			{
				Spec: Spec{
					Package: Package{Path: "github.com/ngicks/und/conversion", Name: "conversion"},
				},
			},
			{
				Spec: Spec{Package: Package{Path: "github.com/ngicks/und/option", Name: "option"}},
			},
			{
				Spec: Spec{
					Package: Package{Path: "github.com/ngicks/und/sliceund", Name: "sliceund"},
				},
			},
			{
				Ident: "sliceelastic",
				Spec: Spec{
					Package: Package{Path: "github.com/ngicks/und/sliceund/elastic", Name: "elastic"},
				},
			},
		}

		assert.DeepEqual(
			t,
			missingSpecs,
			m.missing,
		)

		ident, ok := m.Ident("github.com/ngicks/go-codegen/graft/pkg/importmap/internal/pkg2/pkg2-2")
		assert.Assert(t, ok)
		assert.Equal(t, "pkg22", ident)

		assert.DeepEqual(
			t,
			append(
				[]NamedSpec{{
					Spec: Spec{
						Package: Package{Path: "github.com/ngicks/go-codegen/graft/pkg/importmap/internal/pkg2/pkg2-2", Name: "pkg22"},
					},
				}},
				missingSpecs...,
			),
			m.missing,
		)
	})

	t.Run("AddMissingImports", func(t *testing.T) {
		dec := decorator.NewDecorator(pkg1.Fset)
		df, err := dec.DecorateFile(pkg1.Syntax[0])
		assert.NilError(t, err)

		assertAddMissingImports := func() {
			t.Helper()
			assert.DeepEqual(
				t,
				[]*dst.ImportSpec{
					// Original imports from pkg1.go
					{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und"`},
						Decs: dst.ImportSpecDecorations{NodeDecs: dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}},
					},
					{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/elastic"`},
						Decs: dst.ImportSpecDecorations{NodeDecs: dst.NodeDecs{Before: dst.NewLine, After: dst.NewLine}},
					},
					// Added missing imports (sorted by path)
					{
						Name: dst.NewIdent("pkg22"),
						Path: &dst.BasicLit{
							Kind:  token.STRING,
							Value: `"github.com/ngicks/go-codegen/graft/pkg/importmap/internal/pkg2/pkg2-2"`,
						},
					},
					{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/conversion"`},
					},
					{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/option"`},
					},
					{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/sliceund"`},
					},
					{
						Name: dst.NewIdent("sliceelastic"),
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/ngicks/und/sliceund/elastic"`},
					},
				},
				df.Imports,
			)
		}

		m.AddMissingImports(df)
		assertAddMissingImports()

		// safe for multiple runs
		m.AddMissingImports(df)
		assertAddMissingImports()
	})
}

func TestParser_Fallback(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}
	p := NewParserFromPackages(testdataPkgs)
	p.AddSpec(
		NamedSpec{
			Spec: Spec{Package: Package{Path: "foo0", Name: "foo"}},
		},
		NamedSpec{
			Spec: Spec{Package: Package{Path: "foo1", Name: "foo"}},
		},
		NamedSpec{
			Spec: Spec{Package: Package{Path: "foo2", Name: "foo"}},
		},
		NamedSpec{
			Spec: Spec{Package: Package{Path: "foo3", Name: "foo"}},
		},
		NamedSpec{
			Spec: Spec{Package: Package{Path: "foo4", Name: "foo"}},
		},
	)
	m, err := p.ParseAst(pkg1.Syntax[0].Imports)
	assert.NilError(t, err)

	assert.DeepEqual(
		t,
		[]NamedSpec{
			{
				Ident: "",
				Spec:  Spec{Package: Package{Path: "foo0", Name: "foo"}},
			},
			{
				Ident: "foo_1",
				Spec:  Spec{Package: Package{Path: "foo1", Name: "foo"}},
			},
			{
				Ident: "foo_2",
				Spec:  Spec{Package: Package{Path: "foo2", Name: "foo"}},
			},
			{
				Ident: "foo_3",
				Spec:  Spec{Package: Package{Path: "foo3", Name: "foo"}},
			},
			{
				Ident: "foo_4",
				Spec:  Spec{Package: Package{Path: "foo4", Name: "foo"}},
			},
		},
		m.missing,
	)
}

func TestParser_AstExpr(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}
	p := NewParserFromPackages(testdataPkgs)
	p.AddSpec(undExtra...)
	m, err := p.ParseAst(pkg1.Syntax[0].Imports)
	assert.NilError(t, err)

	expr := m.AstExpr(QualifiedType{PackagePath: "github.com/ngicks/und", TypeName: "Und"})
	assert.Assert(t, expr != nil)
	assert.Equal(t, "und", expr.X.(*ast.Ident).Name)
	assert.Equal(t, "Und", expr.Sel.Name)

	expr = m.AstExpr(QualifiedType{PackagePath: "unknown/package", TypeName: "Type"})
	assert.Assert(t, expr == nil)
}

func TestParser_DstExpr(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}
	p := NewParserFromPackages(testdataPkgs)
	p.AddSpec(undExtra...)
	m, err := p.ParseAst(pkg1.Syntax[0].Imports)
	assert.NilError(t, err)

	expr := m.DstExpr(QualifiedType{PackagePath: "github.com/ngicks/und", TypeName: "Und"})
	assert.Assert(t, expr != nil)
	assert.Equal(t, "und", expr.X.(*dst.Ident).Name)
	assert.Equal(t, "Und", expr.Sel.Name)

	expr = m.DstExpr(QualifiedType{PackagePath: "unknown/package", TypeName: "Type"})
	assert.Assert(t, expr == nil)
}

func TestParser_Qualifier(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}
	p := NewParserFromPackages(testdataPkgs)
	p.AddSpec(undExtra...)
	m, err := p.ParseAst(pkg1.Syntax[0].Imports)
	assert.NilError(t, err)

	qualifier := m.Qualifier(pkg1.Types.Path())

	// Current package should return empty string
	qual := qualifier(pkg1.Types)
	assert.Equal(t, "", qual)

	// Other packages should return their ident
	for _, imported := range pkg1.Types.Imports() {
		qual := qualifier(imported)
		expectedIdent, _ := m.Ident(imported.Path())
		assert.Equal(t, expectedIdent, qual)
	}
}

func TestParser_ParseDst(t *testing.T) {
	var pkg1 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg1" {
			pkg1 = pkg
		}
	}

	dec := decorator.NewDecorator(pkg1.Fset)
	df, err := dec.DecorateFile(pkg1.Syntax[0])
	assert.NilError(t, err)

	p := NewParserFromPackages(testdataPkgs)
	p.AddSpec(undExtra...)
	m, err := p.ParseDst(df.Imports)
	assert.NilError(t, err)

	// Should have the same imports as when parsing AST
	ident, ok := m.Ident("github.com/ngicks/und")
	assert.Assert(t, ok)
	assert.Equal(t, "und", ident)
}

// Ensure that we handle imports with explicit ident correctly
func TestParser_ExplicitIdent(t *testing.T) {
	var pkg2 *packages.Package
	for _, pkg := range testdataPkgs {
		if pkg.Name == "pkg2" {
			pkg2 = pkg
		}
	}
	p := NewParserFromPackages(testdataPkgs)
	m, err := p.ParseAst(pkg2.Syntax[0].Imports)
	assert.NilError(t, err)

	// pkg2 imports pkg2-2 with explicit ident "pkg22"
	ident, ok := m.Ident("github.com/ngicks/go-codegen/graft/pkg/importmap/internal/pkg2/pkg2-2")
	assert.Assert(t, ok)
	assert.Equal(t, "pkg22222", ident)
}
