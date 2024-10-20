package targettypes

import (
	"fmt"

	"github.com/ngicks/und/option"
	undtag "github.com/ngicks/und/undtag"
)

//undgen:generated
func (v All) UndValidate() error {
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidOpt(v.OptRequired) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidOpt(v.OptNullish) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidOpt(v.OptDef) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
			}),
		}

		if !validator.ValidOpt(v.OptNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Und: true,
			}),
		}

		if !validator.ValidOpt(v.OptUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
				Und: true,
			}),
		}

		if !validator.ValidOpt(v.OptDefOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
		}

		if !validator.ValidOpt(v.OptDefOrNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidOpt(v.OptNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidOpt(v.OptDefOrNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidUnd(v.UndRequired) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidUnd(v.UndNullish) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidUnd(v.UndDef) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
			}),
		}

		if !validator.ValidUnd(v.UndNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Und: true,
			}),
		}

		if !validator.ValidUnd(v.UndUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
				Und: true,
			}),
		}

		if !validator.ValidUnd(v.UndDefOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
		}

		if !validator.ValidUnd(v.UndDefOrNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidUnd(v.UndNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidUnd(v.UndDefOrNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidElastic(v.ElaRequired) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidElastic(v.ElaNullish) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidElastic(v.ElaDef) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
			}),
		}

		if !validator.ValidElastic(v.ElaNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Und: true,
			}),
		}

		if !validator.ValidElastic(v.ElaUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
				Und: true,
			}),
		}

		if !validator.ValidElastic(v.ElaDefOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
		}

		if !validator.ValidElastic(v.ElaDefOrNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidElastic(v.ElaNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			}),
		}

		if !validator.ValidElastic(v.ElaDefOrNullOrUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEq) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpGr,
			}),
		}

		if !validator.ValidElastic(v.ElaGr) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpGrEq,
			}),
		}

		if !validator.ValidElastic(v.ElaGrEq) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpLe,
			}),
		}

		if !validator.ValidElastic(v.ElaLe) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpLeEq,
			}),
		}

		if !validator.ValidElastic(v.ElaLeEq) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEquRequired) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEquNullish) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEquDef) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEquNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
				Und: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEquUnd) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNullSlice) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Null: true,
			}),
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNullNullSlice) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			}),
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNullSingle) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			}),
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNullNullSingle) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 3,
				Op:  undtag.LenOpEqEq,
			}),
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNull) {
			return fmt.Errorf("yay")
		}
	}
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def:  true,
				Null: true,
			}),
			Len: option.Some(undtag.LenValidator{
				Len: 3,
				Op:  undtag.LenOpEqEq,
			}),
			Values: option.Some(undtag.ValuesValidator{
				Nonnull: true,
			}),
		}

		if !validator.ValidElastic(v.ElaEqEqNonNullNull) {
			return fmt.Errorf("yay")
		}
	}

	return nil
}

//undgen:generated
func (v WithTypeParam[T]) UndValidate() error {
	{
		validator := undtag.UndOpt{
			States: option.Some(undtag.StateValidator{
				Def: true,
			}),
		}

		if !validator.ValidOpt(v.Baz) {
			return fmt.Errorf("yay")
		}
	}

	return nil
}
