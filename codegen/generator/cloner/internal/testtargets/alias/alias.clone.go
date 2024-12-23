// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help

package alias

//codegen:generated
func (v A) Clone() A {
	return A{
		U: v.U.CloneFunc(
			func(v string) string {
				return v
			},
		),
		UM: func(v map[string]U) map[string]U {
			var out map[string]U

			if v != nil {
				out = make(map[string]U, len(v))
			}

			inner := out
			for k, v := range v {
				inner[k] = v.CloneFunc(
					func(v string) string {
						return v
					},
				)
			}
			out = inner

			return out
		}(v.UM),
	}
}

//codegen:generated
func (v B) Clone() B {
	return B{
		U2: func(v U2) U2 {
			var out U2

			if v != nil {
				out = make(U2, len(v), cap(v))
			}

			inner := out
			for k, v := range v {
				inner[k] = v.CloneFunc(
					func(v string) string {
						return v
					},
				)
			}
			out = inner

			return out
		}(v.U2),
	}
}

//codegen:generated
func (v C) Clone() C {
	return C{
		U3: func(v U3) U3 {
			var out U3

			if v != nil {
				out = make(U3, len(v))
			}

			inner := out
			for k, v := range v {
				outer := &inner
				var inner U2
				if v != nil {
					inner = make(U2, len(v), cap(v))
				}
				for k, v := range v {
					inner[k] = v.CloneFunc(
						func(v string) string {
							return v
						},
					)
				}
				(*outer)[k] = inner
			}
			out = inner

			return out
		}(v.U3),
	}
}

//codegen:generated
func (v D) Clone() D {
	return D{
		U5: v.U5.CloneFunc(
			func(v string) string {
				return v
			},
		),
	}
}
