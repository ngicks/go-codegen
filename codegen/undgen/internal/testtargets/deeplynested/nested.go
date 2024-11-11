package deeplynested

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
)

type Implementor struct {
	Opt option.Option[string] `und:"required"`
}

type DeeplyNested struct {
	A []map[string][5]und.Und[Implementor] `und:"required"`
	B [][][]map[int]Implementor
	C []map[string][5]und.Und[*Implementor] `und:"required"`
	D [][][]map[int]*Implementor
}
