package undgen

import (
	"cmp"
	"go/ast"
	"iter"
	"slices"

	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

func gatherPlainUndTypes(
	pkgs []*packages.Package,
	imports []TargetImport,
	edgeFilter func(edge typeDependencyEdge) bool,
	seqFactory func(g *typeGraph) iter.Seq2[typeIdent, *typeNode],
) (data map[*ast.File]*replaceData, err error) {
	graph, err := newTypeGraph(
		pkgs,
		isUndPlainTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	return gatherUndTypes(graph, imports, edgeFilter, seqFactory)
}

func gatherValidatableUndTypes(
	pkgs []*packages.Package,
	imports []TargetImport,
	edgeFilter func(edge typeDependencyEdge) bool,
	seqFactory func(g *typeGraph) iter.Seq2[typeIdent, *typeNode],
) (data map[*ast.File]*replaceData, err error) {
	graph, err := newTypeGraph(
		pkgs,
		isUndValidatorTarget,
		excludeUndIgnoredCommentedGenDecl,
		excludeUndIgnoredCommentedTypeSpec,
	)
	if err != nil {
		return nil, err
	}
	return gatherUndTypes(graph, imports, edgeFilter, seqFactory)
}

func gatherUndTypes(
	graph *typeGraph,
	imports []TargetImport,
	edgeFilter func(edge typeDependencyEdge) bool,
	seqFactory func(g *typeGraph) iter.Seq2[typeIdent, *typeNode],
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
		func(accumulator *replaceData, current *typeNode) *replaceData {
			if accumulator == nil {
				importMap := parseImports(current.file.Imports, imports)
				dec := decorator.NewDecorator(current.pkg.Fset)
				df, err := dec.DecorateFile(current.file)
				if err != nil {
					panic(wrapped{err})
				}
				accumulator = &replaceData{
					filename:  current.pkg.Fset.Position(current.file.FileStart).Filename,
					dec:       dec,
					df:        df,
					importMap: importMap,
				}
			}
			accumulator.targetNodes = append(accumulator.targetNodes, current)
			slices.SortFunc(accumulator.targetNodes, func(i, j *typeNode) int { return cmp.Compare(i.pos, j.pos) })
			accumulator.targetNodes = slices.CompactFunc(accumulator.targetNodes, func(i, j *typeNode) bool { return i.pos == j.pos })
			return accumulator
		},
		nil,
		xiter.Map2(
			func(_ typeIdent, n *typeNode) (*ast.File, *typeNode) {
				return n.file, n
			},
			xiter.Filter2(
				func(_ typeIdent, n *typeNode) bool {
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
