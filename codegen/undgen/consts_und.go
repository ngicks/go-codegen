package undgen

type ConstSet struct {
	Imports          []TargetImport
	ConversionMethod ConversionMethodsSet
}

var ConstUnd = ConstSet{
	Imports: []TargetImport{
		{
			ImportPath: "github.com/ngicks/und/option",
			Types:      []string{"Option"},
		},
		{
			ImportPath: "github.com/ngicks/und",
			Types:      []string{"Und"},
		},
		{
			ImportPath: "github.com/ngicks/und/elastic",
			Types:      []string{"Elastic"},
		},
		{
			ImportPath: "github.com/ngicks/und/sliceund",
			Types:      []string{"Und"},
		},
		{
			ImportPath: "github.com/ngicks/und/sliceund/elastic",
			Types:      []string{"Elastic"},
		},
		{
			ImportPath: "github.com/ngicks/und/undtag",
			Types:      []string{},
		},
		{
			ImportPath: "github.com/ngicks/und/validate",
			Types:      []string{},
		},
	},
	ConversionMethod: ConversionMethodsSet{
		ToRaw:   "UndRaw",
		ToPlain: "UndPlain",
	},
}

var (
	UndTargetTypes = []TargetType{
		UndTargetTypeOption,
		UndTargetTypeUnd,
		UndTargetTypeElastic,
		UndTargetTypeSliceUnd,
		UndTargetTypeSliceElastic,
	}
	UndTargetTypeOption = TargetType{
		ImportPath: "github.com/ngicks/und/option",
		Name:       "Option",
	}
	UndTargetTypeUnd = TargetType{
		ImportPath: "github.com/ngicks/und",
		Name:       "Und",
	}
	UndTargetTypeElastic = TargetType{
		ImportPath: "github.com/ngicks/und/elastic",
		Name:       "Elastic",
	}
	UndTargetTypeSliceUnd = TargetType{
		ImportPath: "github.com/ngicks/und/sliceund",
		Name:       "Und",
	}
	UndTargetTypeSliceElastic = TargetType{
		ImportPath: "github.com/ngicks/und/sliceund/elastic",
		Name:       "Elastic",
	}
)

const (
	UndPathUndTag   = "github.com/ngicks/und/undtag"
	UndPathValidate = "github.com/ngicks/und/validate"
)
