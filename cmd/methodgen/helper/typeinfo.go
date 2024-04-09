package helper

import "go/ast"

type TypeInfo struct {
	TypeName        string
	TypeParams      []string
	InnerMemberName string
	StructType      *ast.StructType
}
