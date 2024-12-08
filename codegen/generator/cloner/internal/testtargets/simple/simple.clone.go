// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen cloner --help
package simple

func (v A) Clone() A {
	return A{
		A: v.A,
		B: v.B,
		C: func(v *int) *int {
			out := new(int)

			inner := out
			if v != nil {
				v := *v
				*inner = v
			}

			return out
		}(v.C),
	}
}

func (v B) Clone() B {
	return B{
		A: func(v []*[]string) []*[]string {
			out := make([]*[]string, len(v))

			inner := out
			for k, v := range v {
				outer := &inner
				inner := new([]string)
				if v != nil {
					v := *v
					outer := &inner
					inner := make([]string, len(v))
					for k, v := range v {
						inner[k] = v
					}
					(*outer) = &inner
				}
				(*outer)[k] = inner
			}

			return out
		}(v.A),
		B: func(v map[string]int) map[string]int {
			out := make(map[string]int, len(v))

			inner := out
			for k, v := range v {
				inner[k] = v
			}

			return out
		}(v.B),
	}
}
