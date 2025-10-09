package directive

import (
	"fmt"
	"go/ast"
	"go/types"
)

func ExcludeIgnoredGenDecl(genDecl *ast.GenDecl) (bool, error) {
	parsed, found := ParseAst(genDecl.Doc, DirectivePrefix)
	if !found {
		return true, nil
	}
	
	var dir Direction
	if err := parsed.Unmarshal(&dir); err != nil {
		return false, err
	}
	
	// Check for unknown directives
	for key := range parsed {
		if key != DirectiveCommentIgnore && key != DirectiveCommentGenerated {
			return false, fmt.Errorf("unknown: %v", key)
		}
	}
	
	return !dir.MustIgnore(), nil
}

func ExcludeIgnoredTypeSpec(ts *ast.TypeSpec, _ types.Object) (bool, error) {
	parsed, found := ParseAst(ts.Doc, DirectivePrefix)
	if !found {
		return true, nil
	}
	
	var dir Direction
	if err := parsed.Unmarshal(&dir); err != nil {
		return false, err
	}
	
	// Check for unknown directives
	for key := range parsed {
		if key != DirectiveCommentIgnore && key != DirectiveCommentGenerated {
			return false, fmt.Errorf("unknown: %v", key)
		}
	}
	
	return !dir.MustIgnore(), nil
}