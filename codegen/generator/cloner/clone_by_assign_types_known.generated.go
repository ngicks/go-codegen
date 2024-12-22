package cloner

import "github.com/ngicks/go-codegen/codegen/imports"

var knownCloneByAssign = map[imports.TargetType]struct{}{
	{ImportPath: "", Name: "error"}:        {},
	{ImportPath: "unique", Name: "Handle"}: {},
}
