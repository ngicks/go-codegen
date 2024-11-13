package undgen

import (
	"go/ast"
	"go/parser"
	"go/token"
)

func ParseFileIgnoringUndgenGeneratedFiles(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	f, err := parser.ParseFile(fset, filename, src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return f, err
	}
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		direction, _, err := ParseUndComment(genDecl.Doc)
		if err != nil {
			continue
			// no error at this moment
		}
		if direction.generated {
			// leave nothing
			f.Decls = nil
			f.Comments = nil
			break
		}
	}
	return f, nil
}
