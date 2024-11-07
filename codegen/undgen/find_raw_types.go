package undgen

import (
	"slices"
)

type TargetImport struct {
	ImportPath string
	Types      []string
}

func AppendTargetImports(tis []TargetImport, additive TargetImport) []TargetImport {
	idx := slices.IndexFunc(tis, func(t TargetImport) bool { return t.ImportPath == additive.ImportPath })
	if idx >= 0 {
		t := tis[idx]
		t.Types = append(t.Types, additive.Types...)
		slices.Sort(t.Types)
		t.Types = slices.Compact(t.Types)
		tis[idx] = t
	} else {
		tis = append(tis, additive)
	}
	return tis
}
