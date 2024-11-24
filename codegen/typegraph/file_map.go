package typegraph

import (
	"cmp"
	"go/ast"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

// ReplaceData is used to replace types in each file.
type ReplaceData struct {
	Filename    string
	Dec         *decorator.Decorator
	DstFile     *dst.File
	ImportMap   imports.ImportMap
	TargetNodes []*TypeNode
}

// GatherReplaceData converts g into a map of *ast.File to *ReplaceData.
// The data can be later used to modify ast.
func (g *TypeGraph) GatherReplaceData(
	parser *imports.ImportParser,
	seqFactory func(g *TypeGraph) iter.Seq2[TypeIdent, *TypeNode],
) (data map[*ast.File]*ReplaceData, err error) {

	type wrapped struct {
		e error
	}

	defer func() {
		rec := recover()
		if w, ok := rec.(*wrapped); ok {
			err = w.e
			return
		}
		if rec != nil {
			panic(rec)
		}
	}()

	data = hiter.ReduceGroup(
		func(accum *ReplaceData, current *TypeNode) *ReplaceData {
			if accum == nil {
				importMap, err := parser.Parse(current.File.Imports)
				if err != nil {
					panic(&wrapped{err})
				}
				dec := decorator.NewDecorator(current.Pkg.Fset)
				df, err := dec.DecorateFile(current.File)
				if err != nil {
					panic(&wrapped{err})
				}
				accum = &ReplaceData{
					Filename:  current.Pkg.Fset.Position(current.File.FileStart).Filename,
					Dec:       dec,
					DstFile:   df,
					ImportMap: importMap,
				}
			}
			accum.TargetNodes = append(accum.TargetNodes, current)
			return accum
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
				seqFactory(g),
			),
		),
	)

	for _, replacer := range data {
		slices.SortFunc(replacer.TargetNodes, func(i, j *TypeNode) int { return cmp.Compare(i.Pos, j.Pos) })
		replacer.TargetNodes = slices.CompactFunc(replacer.TargetNodes, func(i, j *TypeNode) bool { return i.Pos == j.Pos })
	}

	return data, nil
}
