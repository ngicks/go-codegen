package msg

import (
	"go/types"
	"strconv"

	"github.com/ngicks/go-codegen/codegen/matcher"
)

// PkgPathPrefixedName prints ty's name in "pkgPath".ObjectName style.
// The pkgPath prefix is present only when ty is defined under a package (= the error built-in type prints only "error".)
func PkgPathPrefixedName(ty types.Type) string {
	pkgPath, name := matcher.Name(ty)
	if pkgPath != "" {
		pkgPath = strconv.Quote(pkgPath) + "."
	}
	return pkgPath + name
}
