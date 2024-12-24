package codegen

import (
	"fmt"
	"go/ast"
	"iter"

	"github.com/dave/dst"
)

type FieldDescAst struct {
	Pos   int
	Name  string
	Field *ast.Field
}

func FieldAst(st *ast.StructType) iter.Seq[FieldDescAst] {
	return func(yield func(FieldDescAst) bool) {
		if st.Fields == nil || len(st.Fields.List) == 0 {
			return
		}
		pos := 0
		for i := 0; i < len(st.Fields.List); i++ {
			f := st.Fields.List[i]
			names := f.Names
			if len(names) == 0 {
				// embedded field
				unwrapped := f.Type
				var name string
			UNWRAP:
				for {
					switch x := unwrapped.(type) {
					default:
						panic(fmt.Errorf("unknown type in an anonymous field: %T in %v", unwrapped, f.Type))
					case *ast.Ident:
						name = x.Name
						break UNWRAP
					case *ast.SelectorExpr:
						unwrapped = x.Sel
					case *ast.IndexExpr: // type param
						unwrapped = x.X
					case *ast.IndexListExpr:
						unwrapped = x.X
					}
				}
				if !yield(FieldDescAst{pos, name, f}) {
					return
				}
				pos++
			} else {
				for _, name := range names {
					if !yield(FieldDescAst{pos, name.Name, f}) {
						return
					}
					pos++
				}
			}
		}
	}
}

type FieldDescDst struct {
	Pos   int
	Name  string
	Field *dst.Field
}

func FieldDst(st *dst.StructType) iter.Seq[FieldDescDst] {
	return func(yield func(FieldDescDst) bool) {
		if st.Fields == nil || len(st.Fields.List) == 0 {
			return
		}
		pos := 0
		for i := 0; i < len(st.Fields.List); i++ {
			f := st.Fields.List[i]
			names := f.Names
			for _, name := range names {
				if !yield(FieldDescDst{pos, name.Name, f}) {
					return
				}
				pos++
			}
		}
	}
}
