package typegraph

import (
	"cmp"
	"go/ast"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/ngicks/go-codegen/codegen/codegen"
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
	TargetNodes []*Node
}

// GatherReplaceData converts g into a map of *ast.File to *ReplaceData.
// The data can be later used to modify ast.
func (g *Graph) GatherReplaceData(
	parser *imports.ImportParser,
	seqFactory func(g *Graph) iter.Seq2[Ident, *Node],
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
		func(accum *ReplaceData, current *Node) *ReplaceData {
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
				codegen.TrimPackageComment(df)
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
			func(_ Ident, n *Node) (*ast.File, *Node) {
				return n.File, n
			},
			xiter.Filter2(
				func(_ Ident, n *Node) bool {
					return n != nil
				},
				seqFactory(g),
			),
		),
	)

	for _, replacer := range data {
		slices.SortFunc(replacer.TargetNodes, func(i, j *Node) int { return cmp.Compare(i.Pos, j.Pos) })
		replacer.TargetNodes = slices.CompactFunc(replacer.TargetNodes, func(i, j *Node) bool { return i.Pos == j.Pos })
	}

	return data, nil
}
