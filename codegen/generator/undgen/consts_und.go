package undgen

import (
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/pkg/typematcher"
)

type ConstSet struct {
	Imports          []imports.TargetImport
	ConversionMethod typematcher.CyclicConversionMethods
	ValidatorMethod  typematcher.ErrorMethod
}

var ConstUnd = ConstSet{
	Imports: []imports.TargetImport{
		{
			Import: imports.Import{Path: "github.com/ngicks/und/option", Name: "option"},
			Types:  []string{"Option"},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und", Name: "und"},
			Types:  []string{"Und"},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/elastic", Name: "elastic"},
			Types:  []string{"Elastic"},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/sliceund", Name: "sliceund"},
			Types:  []string{"Und"},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/sliceund/elastic", Name: "elastic"},
			Ident:  "sliceelastic",
			Types:  []string{"Elastic"},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/undtag", Name: "undtag"},
			Types:  []string{},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/validate", Name: "validate"},
			Types:  []string{},
		},
		{
			Import: imports.Import{Path: "github.com/ngicks/und/conversion", Name: "conversion"},
			Types:  []string{"Empty"},
		},
	},
	ConversionMethod: typematcher.CyclicConversionMethods{
		Reverse: "UndRaw",
		Convert: "UndPlain",
	},
	ValidatorMethod: typematcher.ErrorMethod{
		Name: "UndValidate",
	},
}

var (
	UndTargetTypes = []imports.TargetType{
		UndTargetTypeOption,
		UndTargetTypeUnd,
		UndTargetTypeElastic,
		UndTargetTypeSliceUnd,
		UndTargetTypeSliceElastic,
	}
	UndTargetTypeOption = imports.TargetType{
		ImportPath: "github.com/ngicks/und/option",
		Name:       "Option",
	}
	UndTargetTypeUnd = imports.TargetType{
		ImportPath: "github.com/ngicks/und",
		Name:       "Und",
	}
	UndTargetTypeElastic = imports.TargetType{
		ImportPath: "github.com/ngicks/und/elastic",
		Name:       "Elastic",
	}
	UndTargetTypeSliceUnd = imports.TargetType{
		ImportPath: "github.com/ngicks/und/sliceund",
		Name:       "Und",
	}
	UndTargetTypeSliceElastic = imports.TargetType{
		ImportPath: "github.com/ngicks/und/sliceund/elastic",
		Name:       "Elastic",
	}
	UndTargetTypeConversionEmpty = imports.TargetType{
		ImportPath: "github.com/ngicks/und/conversion",
		Name:       "Empty",
	}
)

const (
	UndPathConversion = "github.com/ngicks/und/conversion"
	UndPathUndTag     = "github.com/ngicks/und/undtag"
	UndPathValidate   = "github.com/ngicks/und/validate"
)
