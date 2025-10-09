package filteredge

import "github.com/ngicks/go-codegen/codegen/pkg/typegraph/testdata/faketarget"

// direct
type A struct {
	D D
}

// slice
type B struct {
	D []D
}

// map
type C struct {
	D map[int]D
}

// direct, slice and map
type D struct {
	A MatchedStruct
	B []MatchedStruct
	C map[string]MatchedStruct
}

type MatchedStruct struct {
	Target faketarget.FakeTarget
}

// define types before and after matched types
// so that it can test if test target correctly being independent or order.

// direct to C
type E struct {
	C C
}

// slice to C
type F struct {
	C []C
}

// map to C
type G struct {
	C map[bool]C
}

type H []C

type I map[string]C
