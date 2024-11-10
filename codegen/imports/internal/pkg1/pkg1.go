package pkg1

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
)

type Pkg1 struct {
	Und     und.Und[string]
	Elastic elastic.Elastic[string]
}
