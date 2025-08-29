package tests

import (
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/generator/undgen/internal/testtargets/deeplynested"
	"github.com/ngicks/go-codegen/codegen/generator/undgen/internal/testtargets/implementor"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

var (
	compareOptionPointerImplementorString = gocmp.Comparer(func(i, j und.Und[*implementor.Implementor[string]]) bool {
		return i.EqualFunc(j, func(i, j *implementor.Implementor[string]) bool {
			if i == nil || j == nil {
				return i == nil && j == nil
			}
			return *i == *j
		})
	})
	compareUndImplementorString = gocmp.Comparer(func(i, j und.Und[implementor.Implementor[string]]) bool {
		return i.EqualFunc(j, func(i, j implementor.Implementor[string]) bool {
			return i == j
		})
	})
)

var (
	implementorValid = implementor.Implementor[string]{
		T:   "valid",
		Yay: "yay",
	}
	// implementorInvalid = implementor.Implementor[string]{
	// 	T:   "invalid",
	// 	Yay: "nay",
	// }
	dependantValid = deeplynested.Dependant{
		Opt: option.Some("foo"),
	}
	dependantInvalid = deeplynested.Dependant{
		Opt: option.None[string](),
	}
)

func Test_deeplynested_UndPlain(t *testing.T) {
	d := deeplynested.DeeplyNestedImplementor{
		A: make([]map[string][5]und.Und[implementor.Implementor[string]], 0),
		B: make([][][]map[int]implementor.Implementor[string], 0),
		C: make([]map[string][5]und.Und[*implementor.Implementor[string]], 0),
		D: make([][][]map[int]*implementor.Implementor[string], 0),
	}
	p := d.UndPlain()
	assert.DeepEqual(
		t,
		deeplynested.DeeplyNestedImplementorPlain{
			A: make([]map[string][5]implementor.ImplementorPlain[string], 0),
			B: make([][][]map[int]implementor.ImplementorPlain[string], 0),
			C: make([]map[string][5]*implementor.ImplementorPlain[string], 0),
			D: make([][][]map[int]*implementor.ImplementorPlain[string], 0),
		},
		p,
	)

	d = deeplynested.DeeplyNestedImplementor{
		A: []map[string][5]und.Und[implementor.Implementor[string]]{
			{
				"foo": {
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
				},
				"bar": {
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
					und.Defined(implementorValid),
				},
			},
		},
		B: [][][]map[int]implementor.Implementor[string]{
			{
				{
					{10: implementorValid},
					{20: implementorValid},
				},
			},
		},
		C: []map[string][5]und.Und[*implementor.Implementor[string]]{
			{
				"foo": {
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
				},
				"bar": {
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
					und.Defined(&implementorValid),
				},
			},
		},
		D: [][][]map[int]*implementor.Implementor[string]{
			{
				{
					{10: &implementorValid},
					{20: &implementorValid},
				},
			},
		},
	}
	implP := implementorValid.UndPlain()
	p = d.UndPlain()
	assert.DeepEqual(
		t,
		deeplynested.DeeplyNestedImplementorPlain{
			A: []map[string][5]implementor.ImplementorPlain[string]{
				{
					"foo": {
						implP,
						implP,
						implP,
						implP,
						implP,
					},
					"bar": {
						implP,
						implP,
						implP,
						implP,
						implP,
					},
				},
			},
			B: [][][]map[int]implementor.ImplementorPlain[string]{
				{
					{
						{10: implP},
						{20: implP},
					},
				},
			},
			C: []map[string][5]*implementor.ImplementorPlain[string]{
				{
					"foo": {
						&implP,
						&implP,
						&implP,
						&implP,
						&implP,
					},
					"bar": {
						&implP,
						&implP,
						&implP,
						&implP,
						&implP,
					},
				},
			},
			D: [][][]map[int]*implementor.ImplementorPlain[string]{
				{
					{
						{10: &implP},
						{20: &implP},
					},
				},
			},
		},
		p,
	)
	r := p.UndRaw()
	assert.DeepEqual(
		t,
		d,
		r,
		compareOptionPointerImplementorString,
		compareUndImplementorString,
	)
}

func Test_deeplynested_UndValidate(t *testing.T) {
	d := deeplynested.DeeplyNestedDependant{
		A: make([]map[string][5]und.Und[deeplynested.Dependant], 0),
		B: make([][][]map[int]deeplynested.Dependant, 0),
		C: make([]map[string][5]und.Und[*deeplynested.Dependant], 0),
		D: make([][][]map[int]*deeplynested.Dependant, 0),
	}
	assert.NilError(t, d.UndValidate())
	valid := func() deeplynested.DeeplyNestedDependant {
		return deeplynested.DeeplyNestedDependant{
			A: []map[string][5]und.Und[deeplynested.Dependant]{
				{
					"foo": {
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
					},
					"bar": {
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
						und.Defined(dependantValid),
					},
				},
			},
			B: [][][]map[int]deeplynested.Dependant{
				{
					{
						{10: dependantValid},
						{20: dependantValid},
					},
				},
			},
			C: []map[string][5]und.Und[*deeplynested.Dependant]{
				{
					"foo": {
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
					},
					"bar": {
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
						und.Defined(&dependantValid),
					},
				},
			},
			D: [][][]map[int]*deeplynested.Dependant{
				{
					{
						{10: &dependantValid},
						{20: &dependantValid},
					},
				},
			},
		}
	}
	assert.NilError(t, valid().UndValidate())

	d = valid()
	d.A = []map[string][5]und.Und[deeplynested.Dependant]{
		{
			"foo": {
				und.Defined(dependantValid),
				und.Defined(dependantValid),
				und.Defined(dependantValid),
				und.Defined(dependantValid),
				und.Defined(dependantValid),
			},
			"bar": {
				und.Defined(dependantValid),
				und.Defined(dependantValid),
				und.Defined(dependantInvalid),
				und.Defined(dependantValid),
				und.Defined(dependantValid),
			},
		},
	}
	assert.ErrorContains(t, d.UndValidate(), ".A[0][bar][2].Opt:")
	d = valid()
	d.B = append(
		d.B,
		[][]map[int]deeplynested.Dependant{
			{
				{10: dependantValid},
				{20: dependantInvalid},
			},
		},
	)
	assert.ErrorContains(t, d.UndValidate(), ".B[1][0][1][20].Opt:")
	d = valid()
	d.C = append(
		d.C,
		map[string][5]und.Und[*deeplynested.Dependant]{
			"foo": {
				und.Defined(&dependantValid),
				und.Defined(&dependantInvalid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
			},
			"bar": {
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
			},
		},
	)
	assert.ErrorContains(t, d.UndValidate(), ".C[1][foo][1].Opt:")
	d = valid()
	d.D = append(
		d.D,
		[][]map[int]*deeplynested.Dependant{
			{
				{10: &dependantValid},
				{40: &dependantInvalid},
			},
		},
	)
	assert.ErrorContains(t, d.UndValidate(), ".D[1][0][1][40].Opt:")

	d = valid()

	d.C = append(
		d.C,
		map[string][5]und.Und[*deeplynested.Dependant]{
			"foo": {
				und.Defined(&dependantValid),
				und.Defined[*deeplynested.Dependant](nil),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined[*deeplynested.Dependant](nil),
			},
			"bar": {
				und.Defined(&dependantValid),
				und.Defined[*deeplynested.Dependant](nil),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
				und.Defined(&dependantValid),
			},
		},
	)
	d.D = append(
		d.D,
		[][]map[int]*deeplynested.Dependant{
			{
				{10: nil},
				{40: nil},
			},
		},
	)
	assert.NilError(t, d.UndValidate())
}
