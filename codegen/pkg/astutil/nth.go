package astutil

import (
	"fmt"
	"go/types"
)

type ints interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Nth[I ints](d I) string {
	num := fmt.Sprintf("%d", d)
	switch num[len(num)-1] {
	case '1':
		return num + "st"
	case '2':
		return num + "nd"
	case '3':
		return num + "rd"
	default:
		return num + "th"
	}
}

func PrintFieldDesc(typeName string, i int, field *types.Var) string {
	field.Parent()
	var pkgPath string
	if pkg := field.Pkg(); pkg != nil {
		pkgPath = pkg.Path()
	} else {
		// There could be named builtin type...right?
		pkgPath = "builtin"
	}
	return fmt.Sprintf(
		"%q(%s field) of type %q defined in %q",
		field.Name(), Nth(i), typeName, pkgPath,
	)
}
