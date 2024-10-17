package patchtarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//undgen:generated
type AllPatch struct {
	Foo sliceund.Und[string]
	Bar sliceund.Und[*int]
	Baz sliceund.Und[*struct{}]
	Qux sliceund.Und[[]string]

	Opt          sliceund.Und[string]
	Und          und.Und[string]
	Elastic      elastic.Elastic[string]
	SliceUnd     sliceund.Und[string]
	SliceElastic sliceelastic.Elastic[string]
}

//undgen:generated
func (p *AllPatch) FromValue(v All) {
	//nolint
	*p = AllPatch{
		Foo:          sliceund.Defined(v.Foo),
		Bar:          sliceund.Defined(v.Bar),
		Baz:          sliceund.Defined(v.Baz),
		Qux:          sliceund.Defined(v.Qux),
		Opt:          option.MapOrOption(v.Opt, sliceund.Null[string](), sliceund.Defined[string]),
		Und:          v.Und,
		Elastic:      v.Elastic,
		SliceUnd:     v.SliceUnd,
		SliceElastic: v.SliceElastic,
	}
}

//undgen:generated
func (p AllPatch) ToValue() All {
	//nolint
	return All{
		Foo:          p.Foo.Value(),
		Bar:          p.Bar.Value(),
		Baz:          p.Baz.Value(),
		Qux:          p.Qux.Value(),
		Opt:          option.FlattenOption(p.Opt.Unwrap()),
		Und:          p.Und,
		Elastic:      p.Elastic,
		SliceUnd:     p.SliceUnd,
		SliceElastic: p.SliceElastic,
	}
}

//undgen:generated
func (p AllPatch) Merge(r AllPatch) AllPatch {
	//nolint
	return AllPatch{
		Foo:          sliceund.FromOption(r.Foo.Unwrap().Or(p.Foo.Unwrap())),
		Bar:          sliceund.FromOption(r.Bar.Unwrap().Or(p.Bar.Unwrap())),
		Baz:          sliceund.FromOption(r.Baz.Unwrap().Or(p.Baz.Unwrap())),
		Qux:          sliceund.FromOption(r.Qux.Unwrap().Or(p.Qux.Unwrap())),
		Opt:          sliceund.FromOption(r.Opt.Unwrap().Or(p.Opt.Unwrap())),
		Und:          und.FromOption(r.Und.Unwrap().Or(p.Und.Unwrap())),
		Elastic:      elastic.FromUnd(und.FromOption(r.Elastic.Unwrap().Unwrap().Or(p.Elastic.Unwrap().Unwrap()))),
		SliceUnd:     sliceund.FromOption(r.SliceUnd.Unwrap().Or(p.SliceUnd.Unwrap())),
		SliceElastic: sliceelastic.FromUnd(sliceund.FromOption(r.SliceElastic.Unwrap().Unwrap().Or(p.SliceElastic.Unwrap().Unwrap()))),
	}
}

//undgen:generated
func (p AllPatch) ApplyPatch(v All) All {
	var orgP AllPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//undgen:generated
type HmmPatch struct {
	Ah sliceund.Und[Ignored]
}

//undgen:generated
func (p *HmmPatch) FromValue(v Hmm) {
	//nolint
	*p = HmmPatch{
		Ah: sliceund.Defined(v.Ah),
	}
}

//undgen:generated
func (p HmmPatch) ToValue() Hmm {
	//nolint
	return Hmm{
		Ah: p.Ah.Value(),
	}
}

//undgen:generated
func (p HmmPatch) Merge(r HmmPatch) HmmPatch {
	//nolint
	return HmmPatch{
		Ah: sliceund.FromOption(r.Ah.Unwrap().Or(p.Ah.Unwrap())),
	}
}

//undgen:generated
func (p HmmPatch) ApplyPatch(v Hmm) Hmm {
	var orgP HmmPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
