// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help
package param

//codegen:generated
func (v Param[T, U]) CloneFunc(cloneT func(T) T, cloneU func(U) U) Param[T, U] {
	return Param[T, U]{
		U: cloneU(v.U),
		T: cloneT(v.T),
	}
}

//codegen:generated
func (v Param2[T, U]) CloneFunc(cloneT func(T) T, cloneU func(U) U) Param2[T, U] {
	return Param2[T, U]{
		U: func(v map[string]*U) map[string]*U {
			out := make(map[string]*U, len(v))

			inner := out
			for k, v := range v {
				outer := &inner
				inner := (*U)(nil)
				if v != nil {
					v := *v
					vv := cloneU(v)
					inner = &vv
				}
				(*outer)[k] = inner
			}
			out = inner

			return out
		}(v.U),
		T: func(v *T) *T {
			out := (*T)(nil)

			inner := out
			if v != nil {
				v := *v
				vv := cloneT(v)
				inner = &vv
			}
			out = inner

			return out
		}(v.T),
	}
}
