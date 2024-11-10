package undgen

import (
	"cmp"
	"fmt"
	"go/ast"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

type replaceData struct {
	filename    string
	dec         *decorator.Decorator
	df          *dst.File
	importMap   imports.ImportMap
	targetNodes []*TypeNode
}

func gatherPlainUndTypes(
	pkgs []*packages.Package,
	parser *imports.ImportParser,
	edgeFilter func(edge TypeDependencyEdge) bool,
	seqFactory func(g *TypeGraph) iter.Seq2[TypeIdent, *TypeNode],
) (data map[*ast.File]*replaceData, err error) {
	graph, err := NewTypeGraph(
		pkgs,
		isUndPlainTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	return gatherUndTypes(graph, parser, edgeFilter, seqFactory)
}

func gatherValidatableUndTypes(
	pkgs []*packages.Package,
	parser *imports.ImportParser,
	edgeFilter func(edge TypeDependencyEdge) bool,
	seqFactory func(g *TypeGraph) iter.Seq2[TypeIdent, *TypeNode],
) (data map[*ast.File]*replaceData, err error) {
	graph, err := NewTypeGraph(
		pkgs,
		isUndValidatorTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	return gatherUndTypes(graph, parser, edgeFilter, seqFactory)
}

func gatherUndTypes(
	graph *TypeGraph,
	parser *imports.ImportParser,
	edgeFilter func(edge TypeDependencyEdge) bool,
	seqFactory func(g *TypeGraph) iter.Seq2[TypeIdent, *TypeNode],
) (data map[*ast.File]*replaceData, err error) {
	if edgeFilter != nil {
		graph.markTransitive(edgeFilter)
	}

	type wrapped struct {
		e error
	}

	defer func() {
		rec := recover()
		if w, ok := rec.(wrapped); ok {
			err = w.e
			return
		}
		if rec != nil {
			panic(rec)
		}
	}()

	return hiter.ReduceGroup(
		func(accumulator *replaceData, current *TypeNode) *replaceData {
			if accumulator == nil {
				importMap, err := parser.Parse(current.File.Imports)
				if err != nil {
					fmt.Printf("%#v\n\n", parser)
					panic(wrapped{err})
				}
				dec := decorator.NewDecorator(current.Pkg.Fset)
				df, err := dec.DecorateFile(current.File)
				if err != nil {
					panic(wrapped{err})
				}
				accumulator = &replaceData{
					filename:  current.Pkg.Fset.Position(current.File.FileStart).Filename,
					dec:       dec,
					df:        df,
					importMap: importMap,
				}
			}
			accumulator.targetNodes = append(accumulator.targetNodes, current)
			slices.SortFunc(accumulator.targetNodes, func(i, j *TypeNode) int { return cmp.Compare(i.Pos, j.Pos) })
			accumulator.targetNodes = slices.CompactFunc(accumulator.targetNodes, func(i, j *TypeNode) bool { return i.Pos == j.Pos })
			return accumulator
		},
		nil,
		xiter.Map2(
			func(_ TypeIdent, n *TypeNode) (*ast.File, *TypeNode) {
				return n.File, n
			},
			xiter.Filter2(
				func(_ TypeIdent, n *TypeNode) bool {
					return n != nil
				},
				seqFactory(graph),
			),
		),
	), nil
}

func enumerateFile(pkgs []*packages.Package) iter.Seq[*ast.File] {
	return func(yield func(*ast.File) bool) {
		for _, pkg := range pkgs {
			for _, f := range pkg.Syntax {
				if !yield(f) {
					return
				}
			}
		}
	}
}
