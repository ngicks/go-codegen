package importmap

import (
	"fmt"
	"go/ast"
	"go/types"
	"iter"
	"strconv"

	"github.com/dave/dst"
	"golang.org/x/tools/go/packages"
)

type Parser struct {
	// invariants: always sorted by path
	added        []NamedSpec
	dependencies Specs
}

func NewParser(pkg *types.Package) *Parser {
	parser := &Parser{
		dependencies: make(Specs),
	}
	for pkg := range enumeratePkg(pkg) {
		parser.dependencies.Insert(specFromTypesPackage(pkg))
	}
	return parser
}

func enumeratePkg(pkg *types.Package) iter.Seq[*types.Package] {
	return func(yield func(*types.Package) bool) {
		if !yield(pkg) {
			return
		}
		// Doing breadth-first search because recursive depth-first search is inefficient.
		// You don't need to worry about recursion check because Go does not allow cyclic imports.
		childrenNotVisited := []*types.Package{pkg}
		var nextLayer []*types.Package
		for len(childrenNotVisited) != 0 {
			for _, parent := range childrenNotVisited {
				if i := parent.Imports(); len(i) > 0 {
					nextLayer = append(nextLayer, i...)
					for _, ii := range i {
						if !yield(ii) {
							return
						}
					}
				}
			}
			childrenNotVisited, nextLayer = nextLayer, childrenNotVisited
			nextLayer = nextLayer[:0]
		}
	}
}

func NewParserFromPackages(pkgs []*packages.Package) *Parser {
	parser := &Parser{
		dependencies: make(Specs),
	}
	for pkg := range enumeratePkgs(pkgs) {
		parser.dependencies.Insert(specFromTypesPackage(pkg))
	}
	return parser
}

func enumeratePkgs(pkgs []*packages.Package) iter.Seq[*types.Package] {
	return func(yield func(*types.Package) bool) {
		packages.Visit(
			pkgs,
			func(p *packages.Package) bool {
				return yield(p.Types)
			},
			nil,
		)
	}
}

func (p *Parser) AddSpec(specs ...NamedSpec) {
	p.added = append(p.added, specs...)
	p.added = cleanNamedSpec(p.added)
}

func (p *Parser) parseLit(seq iter.Seq2[string, string]) (*Map, error) {
	m := &Map{
		names: map[string]struct{}{},
		p:     p,
	}

	for ident, rawPkgPath := range seq {
		pkgPath := unquoteBasicLitString(rawPkgPath)

		s, ok := p.dependencies[pkgPath]
		if !ok {
			return nil, fmt.Errorf("input import specs have unknown path: %q", pkgPath)
		}

		ns := NamedSpec{Ident: ident, Spec: s}

		if ident != "." && ident != "_" {
			name, _ := ns.Name()
			m.names[name] = struct{}{}
		}
		m.ident = append(m.ident, ns)
	}

	m.ident = cleanNamedSpec(m.ident)

	for _, spec := range m.p.added {
		_, ok := searchNamedSpecs(m.ident, spec)
		if !ok {
			m.appendFallingBack(spec)
		}
	}
	m.missing = cleanNamedSpec(m.missing)

	return m, nil
}

func (m *Map) appendFallingBack(spec NamedSpec) NamedSpec {
	orgIdent, needOverride := spec.Name()

	// find conflicting name.
	_, has := m.names[orgIdent]
	if !has {
		if needOverride {
			spec.Ident = orgIdent
		}
		m.names[orgIdent] = struct{}{}
		m.missing = append(m.missing, spec)
		return spec
	}

	var ident string
	for i := int64(1); ; i++ {
		ident = orgIdent + "_" + strconv.FormatInt(i, 10)

		_, has := m.names[ident]
		if !has {
			spec.Ident = ident
			m.names[ident] = struct{}{}
			m.missing = append(m.missing, spec)
			return spec
		}
	}
}

func (m *Map) appendFallingBackClean(spec NamedSpec) NamedSpec {
	ret := m.appendFallingBack(spec)
	m.missing = cleanNamedSpec(m.missing)
	return ret
}

func (p *Parser) ParseAst(specs []*ast.ImportSpec) (*Map, error) {
	return p.parseLit(func(yield func(string, string) bool) {
		for _, spec := range specs {
			var ident string
			if spec.Name != nil {
				ident = spec.Name.Name
			}
			if !yield(ident, spec.Path.Value) {
				return
			}
		}
	})
}

func (p *Parser) ParseDst(specs []*dst.ImportSpec) (*Map, error) {
	return p.parseLit(func(yield func(string, string) bool) {
		for _, spec := range specs {
			var ident string
			if spec.Name != nil {
				ident = spec.Name.Name
			}
			if !yield(ident, spec.Path.Value) {
				return
			}
		}
	})
}

// strips " or ` from basic lit string.
func unquoteBasicLitString(s string) string {
	if len(s) == 0 {
		// impossible. just avoiding panic. or should we panic?
		return s
	}
	if s[0] == '"' {
		pkgPath, err := strconv.Unquote(s)
		if err != nil {
			panic(fmt.Errorf("malformed import: %w", err))
		}
		return pkgPath
	} else {
		return s[1 : len(s)-1]
	}
}
