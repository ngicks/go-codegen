package targettypes

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type Example struct {
	Foo   string                    `json:"foo"`
	Bar   option.Option[string]     `json:"bar" und:"required"`
	Baz   und.Und[string]           `json:"baz" und:"def"`
	Qux   und.Und[string]           `json:"qux" und:"def,null"`
	Quux  sliceelastic.Elastic[int] `json:"quux" und:"len==3"`
	Corge sliceelastic.Elastic[int] `json:"corge" und:"len>2,values:nonnull"`
}
