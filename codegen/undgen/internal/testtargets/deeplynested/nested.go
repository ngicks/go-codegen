// tests dependant and implementor
package deeplynested

import (
	"github.com/ngicks/go-codegen/codegen/undgen/internal/testtargets/implementor"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
)

type DeeplyNestedImplementor struct {
	A []map[string][5]und.Und[implementor.Implementor[string]] `und:"required"`
	B [][][]map[int]implementor.Implementor[string]
	C []map[string][5]und.Und[*implementor.Implementor[string]] `und:"required"`
	D [][][]map[int]*implementor.Implementor[string]
}

type Dependant struct {
	Opt option.Option[string] `und:"required"`
}

type DeeplyNestedDependant struct {
	A []map[string][5]und.Und[Dependant] `und:"required"`
	B [][][]map[int]Dependant
	C []map[string][5]und.Und[*Dependant] `und:"required"`
	D [][][]map[int]*Dependant
}

type DeeplyNestedImplementorMap []map[string][5]und.Und[implementor.Implementor[string]]

type DeeplyNestedDependantMap []map[string][5]und.Und[Dependant]
