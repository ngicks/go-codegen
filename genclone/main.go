// genclone generates deep cloner methods
//
// genclone parses a package using "go/ast",
// then collects type information it needs.

package genclone

// TypeInfo records result earned by parsing a target package.
type TypeInfo struct {
	// Name of type.
	Name string
	// Module qualifier of type.
	Qual string
	// Types that refers
	RefFrom []TypeInfo
	Cloner  *ClonerInfo
	RefTo   []TypeInfo
}

type ClonerInfo struct {
	// Name of method.
	// If empty, it will fall back to default name, Clone.
	Name string
	// Already implemented
	Implemented bool
}
